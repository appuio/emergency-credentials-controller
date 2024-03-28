package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
)

func Test_EmergencyAccountReconciler_Reconcile(t *testing.T) {
	ctx := log.IntoContext(context.Background(), testr.New(t))
	clock := &mockClock{now: time.Date(2022, 12, 4, 22, 45, 0, 0, time.UTC)}

	ea := &emcv1beta1.EmergencyAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: emcv1beta1.EmergencyAccountSpec{
			ValidityDuration:        metav1.Duration{Duration: 24 * time.Hour},
			MinValidityDurationLeft: metav1.Duration{Duration: 12 * time.Hour},
			MinRecreateInterval:     metav1.Duration{Duration: 5 * time.Minute},
			TokenStores: []emcv1beta1.TokenStoreSpec{
				{
					Name: "testsecret",
					Type: "secret",
				},
				{
					Name: "testlog",
					Type: "log",
				},
			},
		},
	}

	c, control := fakeClient(t, clock, ea)

	subject := &EmergencyAccountReconciler{
		Client: c,
		Scheme: c.Scheme(),
		Clock:  clock,
	}

	// Create finalizer
	_, err := subject.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(ea)})
	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(ea), ea))
	t.Logf("status %+v", ea.Status)
	require.Len(t, ea.Finalizers, 1, "finalizer should be created")

	// Create token
	_, err = subject.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(ea)})
	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(ea), ea))
	t.Logf("status %+v", ea.Status)
	require.Len(t, ea.Status.Tokens, 1, "token should be created")
	require.WithinDuration(t, clock.Now(), ea.Status.LastTokenCreationTimestamp.Time, 0, "last token creation timestamp should be set")
	lastTimestamp := ea.Status.LastTokenCreationTimestamp.Time

	// Check token
	clock.Advance(1 * time.Hour)
	_, err = subject.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(ea)})
	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(ea), ea))
	t.Logf("status %+v", ea.Status)
	require.Len(t, ea.Status.Tokens, 1, "should not have created a new token")
	require.WithinDuration(t, lastTimestamp, ea.Status.LastTokenCreationTimestamp.Time, 0, "last created timestamp should not have changed")

	// Modify token store
	ea.Spec.TokenStores[1].LogSpec = emcv1beta1.LogStoreSpec{
		AdditionalFields: map[string]string{"test": "test"},
	}
	require.NoError(t, c.Update(ctx, ea))
	_, err = subject.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(ea)})
	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(ea), ea))
	t.Logf("status %+v", ea.Status)
	require.Len(t, ea.Status.Tokens, 2, "should add a new token")

	// Check token - too old and renew
	clock.Advance(12 * time.Hour)
	_, err = subject.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(ea)})
	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(ea), ea))
	t.Logf("status %+v", ea.Status)
	require.Len(t, ea.Status.Tokens, 3, "should add a new token")

	// Check token - verification fails
	clock.Advance(time.Minute)
	control.authenticationErr = fmt.Errorf("test error")
	_, err = subject.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(ea)})
	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(ea), ea))
	t.Logf("status %+v", ea.Status)
	require.Len(t, ea.Status.Tokens, 3, "still in MinRecreateInterval, should not have created a new token")

	// Check token - verification fails, but MinRecreateInterval is over
	clock.Advance(5 * time.Minute)
	_, err = subject.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(ea)})
	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(ea), ea))
	t.Logf("status %+v", ea.Status)
	require.Len(t, ea.Status.Tokens, 4, "should add a new token")

	// Finalizer should be removed and no metric left
	require.NoError(t, c.Delete(ctx, ea))
	deleted := &emcv1beta1.EmergencyAccount{}
	_, err = subject.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(ea)})
	require.NoError(t, err)
	require.Error(t, c.Get(ctx, client.ObjectKeyFromObject(ea), deleted))
	require.Len(t, deleted.Finalizers, 0, "finalizer should be removed")
	ml, err := testutil.GatherAndCount(metrics.Registry, MetricsNamespace+"_verified_tokens_valid_until_seconds")
	require.NoError(t, err)
	require.Equal(t, 0, ml, "metric should be removed")
}

type fakeClientControl struct {
	authenticationErr error
}

func fakeClient(t *testing.T, clock Clock, initObjs ...client.Object) (client.WithWatch, *fakeClientControl) {
	t.Helper()

	fcc := &fakeClientControl{}

	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	require.NoError(t, emcv1beta1.AddToScheme(scheme))
	require.NoError(t, authenticationv1.AddToScheme(scheme))

	icf := interceptor.Funcs{
		Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			// Intercept token review requests and return an error if configured
			tr, ok := obj.(*authenticationv1.TokenReview)
			if ok {
				if fcc.authenticationErr != nil {
					tr.Status.Authenticated = false
					tr.Status.Error = fcc.authenticationErr.Error()
					return nil
				}
				tr.Status.Authenticated = true
				return nil
			}

			return client.Create(ctx, obj, opts...)
		},
		SubResourceCreate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
			rq, ok := subResource.(*authenticationv1.TokenRequest)
			if !ok {
				return fmt.Errorf("unexpected subresource type: %T", subResource)
			}
			rq.Status = authenticationv1.TokenRequestStatus{
				ExpirationTimestamp: metav1.Time{Time: clock.Now().Add(time.Duration(time.Second * time.Duration(*rq.Spec.ExpirationSeconds)))},
				Token:               "eyJhbGciOiJSUzI1NiIsImtpZCI6IkZqSlFfNTZqQ2l5M3JKSS1IRi1vU2czR2Y4cmtDQlN5OTh6QzdNRXV6Vm8ifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiXSwiZXhwIjoxNjk4Njg4NjIwLCJpYXQiOjE2OTg2ODgwMjAsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImVtZXJnZW5jeWFjY291bnQtc2FtcGxlIiwidWlkIjoiM2U1Njg1YWEtYmE2ZC00MGRjLTk4M2MtYzU3MWU5MDQ1YjM4In19LCJuYmYiOjE2OTg2ODgwMjAsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmVtZXJnZW5jeWFjY291bnQtc2FtcGxlIn0.fE3w1IVhm6UZ7Il4ElNzuECChnyZVZ5Vug1DwaTHEJbDtT3Zf2vb-xYxC-PrE8oUrrdIfKEMsqC1xFszTeeQ-i_exFWkikN9_NRg38PuNg23segXC4I6lbBmO4fat2vyEWGQafI8lfnhyhli1ye2rslu1g86itLrDMO8ypykqkRsncoDJW22BVL-GoefjfQVQj-QmmFRzQSVdDYb3AVZAOmSzu_P_N-SjbYHbjrNOeWx-_8ftUS6SlbcVh3laL9MULpoz-gIRmDqs4MI1-CyJd4z9IDZJQI9tcyMsQYPgTYNW2TWxJHy2rxdoxQTR7LJ033tlVgh54jZRyybAaaD3w",
			}
			return nil
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		WithInterceptorFuncs(icf).
		WithStatusSubresource(
			&emcv1beta1.EmergencyAccount{},
		).
		Build()

	return cl, fcc
}

type mockClock struct {
	now time.Time
}

func (m mockClock) Now() time.Time {
	return m.now
}

func (m *mockClock) Advance(d time.Duration) {
	m.now = m.now.Add(d)
}
