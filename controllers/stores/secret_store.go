package stores

import (
	"context"
	"fmt"
	"strconv"
	"time"

	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
	"github.com/appuio/emergency-credentials-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch

type SecretStore struct {
	SecretStoreSpec emcv1beta1.SecretStoreSpec
	Client          client.Client
}

var _ TokenStorer = &SecretStore{}
var _ ClientInjector = &SecretStore{}
var _ TokenRetriever = &SecretStore{}

func NewSecretStore(sts emcv1beta1.SecretStoreSpec) *SecretStore {
	return &SecretStore{
		SecretStoreSpec: sts,
	}
}

// InjectClient injects the client into the SecretStore
func (ss *SecretStore) InjectClient(c client.Client) {
	ss.Client = c
}

func (ss *SecretStore) StoreToken(ctx context.Context, ea emcv1beta1.EmergencyAccount, token string) (string, error) {
	t, err := utils.ParseJWTWithoutVerify(token)
	if err != nil {
		return "", fmt.Errorf("unable to parse token: %w", err)
	}

	exp, err := t.Claims.GetExpirationTime()
	if err != nil {
		return "", fmt.Errorf("unable to get expiration time from token: %w", err)
	}

	s := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ea.Name + "-" + strconv.Itoa(int(exp.Unix())),
			Namespace: ea.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, ss.Client, &s, func() error {
		if s.Data == nil {
			s.Data = map[string][]byte{}
		}
		s.Data["token"] = []byte(token)

		if s.Annotations == nil {
			s.Annotations = map[string]string{}
		}
		s.Annotations["emergency-credentials-controller.appuio.ch/valid-until"] = exp.Format(time.RFC3339)

		return controllerutil.SetControllerReference(&ea, &s, ss.Client.Scheme())
	})
	if err != nil {
		return "", fmt.Errorf("unable to create or update secret: %w (op: %s, secret: %s)", err, op, s.Name)
	}
	log.FromContext(ctx).Info("stored token", "secret", s.Name, "op", op)

	return s.Name, nil
}

func (ss *SecretStore) RetrieveToken(ctx context.Context, ea emcv1beta1.EmergencyAccount, ref string) (string, error) {
	var s corev1.Secret
	err := ss.Client.Get(ctx, types.NamespacedName{Name: ref, Namespace: ea.Namespace}, &s)
	if err != nil {
		return "", fmt.Errorf("unable to get secret: %w", err)
	}
	token, ok := s.Data["token"]
	if !ok {
		return "", fmt.Errorf("secret does not contain token")
	}
	return string(token), nil
}
