package stores_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr/funcr"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/appuio/emergency-credentials-controller/api/v1beta1"
	"github.com/appuio/emergency-credentials-controller/controllers/stores"
)

func Test_LogStore(t *testing.T) {
	testToken := "cooltoken123"
	called := false

	ctx := log.IntoContext(context.Background(), funcr.New(func(prefix, args string) {
		called = true
		require.Contains(t, args, testToken)
	}, funcr.Options{}))

	stores.NewLogStore(v1beta1.LogStoreSpec{}).StoreToken(ctx, v1beta1.EmergencyAccount{}, testToken)
	require.True(t, called, "log should be called")
}
