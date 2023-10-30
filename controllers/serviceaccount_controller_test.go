package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Test_ServiceAccountReconciler_Reconcile(t *testing.T) {
	ctx := context.Background()

	client := controllerClient(t)

	subject := &ServiceAccountReconciler{
		Client: client,
		Scheme: client.Scheme(),
	}

	_, err := subject.Reconcile(ctx, reconcile.Request{})
	require.NoError(t, err)
}

func controllerClient(t *testing.T, initObjs ...client.Object) client.WithWatch {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		WithStatusSubresource(
		// status subresources
		).
		Build()
}
