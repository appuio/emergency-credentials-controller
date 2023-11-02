package stores

import (
	"context"

	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LogStore is a store that logs the token but does not store it anywhere
// Retreiving the token is thus not implemented
type LogStore struct {
	LogStoreSpec emcv1beta1.LogStoreSpec
	Client       client.Client
}

var _ TokenStorer = &LogStore{}

func NewLogStore(sts emcv1beta1.LogStoreSpec) *LogStore {
	return &LogStore{
		LogStoreSpec: sts,
	}
}

func (ss *LogStore) StoreToken(ctx context.Context, ea emcv1beta1.EmergencyAccount, token string) (string, error) {
	log.FromContext(ctx).Info("new token created", "token", token)
	return "", nil
}
