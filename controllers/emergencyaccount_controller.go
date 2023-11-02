package controllers

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/slices"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/integer"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
	"github.com/appuio/emergency-credentials-controller/controllers/stores"
)

type Clock interface {
	Now() time.Time
}

// EmergencyAccountReconciler reconciles a EmergencyAccount object
type EmergencyAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Clock Clock
}

//+kubebuilder:rbac:groups=cluster.appuio.io,resources=emergencyaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.appuio.io,resources=emergencyaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cluster.appuio.io,resources=emergencyaccounts/finalizers,verbs=update

// Reconcile reconciles the EmergencyAccount resource.
// It creates a service account with the same name and namespace as the EmergencyAccount and requests a token for it.
// The token is then stored in the configured stores.
func (r *EmergencyAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithName("EmergencyAccountReconciler.Reconcile")

	instance := &emcv1beta1.EmergencyAccount{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			l.Info("EmergencyAccount resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("unable to get EmergencyAccount resource: %w", err)
	}

	sa, err := r.reconcileSA(ctx, instance)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to reconcile ServiceAccount: %w", err)
	}

	verified, failedVerification := r.verifyTokens(ctx, instance)
	if len(failedVerification) > 0 {
		us := make([]string, len(failedVerification))
		for i, tv := range failedVerification {
			us[i] = tv.String()
		}
		l.Info("unverified tokens found", "tokens", us)
	}
	l.Info("verified tokens found", "ntokens", len(verified))

	// Update metrics
	validUntilUnix := int64(0)
	for _, tv := range verified {
		validUntilUnix = integer.Int64Max(validUntilUnix, tv.tokenRef.ExpirationTimestamp.Unix())
	}
	verifiedTokensValidUntil.WithLabelValues(instance.Name).Set(float64(validUntilUnix))

	nValidityLeft := 0
	for _, tv := range verified {
		if tv.tokenRef.ExpirationTimestamp.Time.After(r.Clock.Now().Add(instance.Spec.MinValidityDurationLeft.Duration)) {
			nValidityLeft++
		}
	}
	if nValidityLeft > 0 {
		l.Info("enough tokens have validity left, not creating new one", "ntokens", nValidityLeft)
		return ctrl.Result{RequeueAfter: instance.Spec.CheckInterval.Duration}, nil
	}
	l.Info("not enough tokens have validity left, creating new one")

	if instance.Status.LastTokenCreationTimestamp.Add(instance.Spec.MinRecreateInterval.Duration).After(r.Clock.Now()) {
		l.Info("last token creation too recent, not creating a new one")
		requeueIn := instance.Status.LastTokenCreationTimestamp.Add(instance.Spec.MinRecreateInterval.Duration).Sub(r.Clock.Now())
		return ctrl.Result{RequeueAfter: requeueIn}, nil
	}

	if err := r.createAndStoreToken(ctx, instance, sa); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create and store token: %w", err)
	}

	return ctrl.Result{RequeueAfter: instance.Spec.CheckInterval.Duration}, nil
}

func (r *EmergencyAccountReconciler) reconcileSA(ctx context.Context, instance *emcv1beta1.EmergencyAccount) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, sa, func() error {
		err := controllerutil.SetControllerReference(instance, sa, r.Scheme)
		if err != nil {
			return fmt.Errorf("unable to set controller reference: %w", err)
		}
		return nil
	})
	if err != nil {
		return sa, fmt.Errorf("unable to create or update ServiceAccount: %w (op: %s)", err, op)
	}
	return sa, nil
}

type tokenVerification struct {
	tokenRef emcv1beta1.TokenStatus
	errs     []error
}

func (tv *tokenVerification) AddError(err error) {
	tv.errs = append(tv.errs, err)
}

func (tv *tokenVerification) Verified() bool {
	return len(tv.errs) == 0
}

func (tv *tokenVerification) String() string {
	if tv.Verified() {
		return fmt.Sprintf("%s: verified", tv.tokenRef.UID)
	}
	return fmt.Sprintf("%s: %v", tv.tokenRef.UID, tv.errs)
}

func (r *EmergencyAccountReconciler) verifyTokens(ctx context.Context, instance *emcv1beta1.EmergencyAccount) (verified []tokenVerification, failed []tokenVerification) {
	l := log.FromContext(ctx).WithName("EmergencyAccountReconciler.verifyTokens")

	tvs := make([]tokenVerification, len(instance.Status.Tokens))
	for i, ts := range instance.Status.Tokens {
		tvs[i].tokenRef = ts
		tv := &tvs[i]
		l := l.WithValues("token", ts.UID)

		if ts.ExpirationTimestamp.Time.Before(r.Clock.Now()) {
			tv.AddError(fmt.Errorf("token expired"))
			continue
		}
		for _, store := range instance.Spec.TokenStores {
			refI := slices.IndexFunc(ts.Refs, func(ref emcv1beta1.TokenStatusRef) bool {
				return store.Name == ref.Store
			})
			if refI == -1 {
				tv.AddError(fmt.Errorf("reference not found for %q", store.Name))
				continue
			}
			ref := ts.Refs[refI]

			st, err := stores.FromSpec(store)
			if err != nil {
				tv.AddError(fmt.Errorf("unable to create store %q: %w", store.Name, err))
				continue
			}
			str, ok := st.(stores.TokenRetriever)
			if !ok {
				l.Info("store does not support token retrieval, not verifying token integrity", "store", store.Name)
				continue
			}
			if ij, ok := st.(stores.ClientInjector); ok {
				ij.InjectClient(r.Client)
			}
			token, err := str.RetrieveToken(ctx, *instance, ref.Ref)
			if err != nil {
				tv.AddError(fmt.Errorf("store %q unable to retrieve token: %w", store.Name, err))
				continue
			}
			rv := authenticationv1.TokenReview{
				Spec: authenticationv1.TokenReviewSpec{
					Token: token,
				},
			}
			err = r.Client.Create(ctx, &rv)
			if err != nil {
				tv.AddError(fmt.Errorf("unable to create TokenReview: %w", err))
				continue
			}
			if !rv.Status.Authenticated {
				tv.AddError(fmt.Errorf("token not authenticated: %s", rv.Status.Error))
			}
		}
	}

	verifiedTokens := make([]tokenVerification, 0, len(tvs))
	failedVerification := make([]tokenVerification, 0, len(tvs))
	for _, tv := range tvs {
		if tv.Verified() {
			verifiedTokens = append(verifiedTokens, tv)
		} else {
			failedVerification = append(failedVerification, tv)
		}
	}

	return verifiedTokens, failedVerification
}

func (r *EmergencyAccountReconciler) createAndStoreToken(ctx context.Context, instance *emcv1beta1.EmergencyAccount, sa *corev1.ServiceAccount) error {
	l := log.FromContext(ctx).WithName("EmergencyAccountReconciler.createAndStoreToken")

	tr := authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: ptr.To(int64(instance.Spec.ValidityDuration.Seconds())),
		},
	}

	err := r.Client.SubResource("token").Create(ctx, sa, &tr)
	if err != nil {
		return fmt.Errorf("unable to create TokenRequest: %w", err)
	}

	l.Info("token created", "expirationTimestamp", tr.Status.ExpirationTimestamp)

	status := emcv1beta1.TokenStatus{
		UID:                 uuid.NewUUID(),
		ExpirationTimestamp: tr.Status.ExpirationTimestamp,
		Refs:                make([]emcv1beta1.TokenStatusRef, 0, len(instance.Spec.TokenStores)),
	}
	for _, s := range instance.Spec.TokenStores {
		st, err := stores.FromSpec(s)
		if err != nil {
			return fmt.Errorf("unable to create store: %w", err)
		}
		if ij, ok := st.(stores.ClientInjector); ok {
			ij.InjectClient(r.Client)
		}
		ref, err := st.StoreToken(ctx, *instance, tr.Status.Token)
		if err != nil {
			return fmt.Errorf("unable to store token: %w", err)
		}
		status.Refs = append(status.Refs, emcv1beta1.TokenStatusRef{
			Ref:   ref,
			Store: s.Name,
		})
	}

	instance.Status.LastTokenCreationTimestamp = metav1.Time{Time: r.Clock.Now()}
	instance.Status.Tokens = append(instance.Status.Tokens, status)

	return r.Client.Status().Update(ctx, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *EmergencyAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&emcv1beta1.EmergencyAccount{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}
