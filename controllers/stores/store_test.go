package stores_test

import (
	"testing"

	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
	"github.com/appuio/emergency-credentials-controller/controllers/stores"
	"github.com/stretchr/testify/require"
)

func Test_FromSpec_SecretStore(t *testing.T) {
	s, err := stores.FromSpec(emcv1beta1.TokenStoreSpec{
		Type: "secret",
	})
	require.NoError(t, err)
	require.IsType(t, &stores.SecretStore{}, s)
}

func Test_FromSpec_LogStore(t *testing.T) {
	s, err := stores.FromSpec(emcv1beta1.TokenStoreSpec{
		Type: "log",
	})
	require.NoError(t, err)
	require.IsType(t, &stores.LogStore{}, s)
}

func Test_FromSpec_Unknown(t *testing.T) {
	_, err := stores.FromSpec(emcv1beta1.TokenStoreSpec{
		Type: "unknown",
	})
	require.Error(t, err)
}
