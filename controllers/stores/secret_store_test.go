package stores_test

import (
	"context"
	"testing"

	"github.com/appuio/emergency-credentials-controller/api/v1beta1"
	"github.com/appuio/emergency-credentials-controller/controllers/stores"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
)

func Test_SecretStore_E2E(t *testing.T) {
	c := fakeClient(t)
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzk0MDB9.Sld3n7qeWVG4a-n5U0e5a5igcuQr0i1OeKncckhBHHE"

	ea := emcv1beta1.EmergencyAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	ss := stores.NewSecretStore(v1beta1.SecretStoreSpec{})
	ss.InjectClient(c)

	ref, err := ss.StoreToken(context.Background(), ea, testToken)
	require.NoError(t, err)
	require.NotEmpty(t, ref)

	token, err := ss.RetrieveToken(context.Background(), ea, ref)
	require.NoError(t, err)
	require.Equal(t, testToken, token)
}

func fakeClient(t *testing.T, initObjs ...client.Object) client.WithWatch {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	require.NoError(t, emcv1beta1.AddToScheme(scheme))

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		Build()
}
