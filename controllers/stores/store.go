package stores

import (
	"context"
	"fmt"

	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TokenStorer interface {
	StoreToken(ctx context.Context, ea emcv1beta1.EmergencyAccount, token string) (ref string, err error)
}

type TokenRetriever interface {
	RetrieveToken(ctx context.Context, ea emcv1beta1.EmergencyAccount, ref string) (token string, err error)
}

type ClientInjector interface {
	InjectClient(client.Client)
}

func FromSpec(sts emcv1beta1.TokenStoreSpec) (TokenStorer, error) {
	if sts.Type == "secret" {
		return NewSecretStore(sts.SecretSpec), nil
	}
	if sts.Type == "log" {
		return NewLogStore(sts.LogSpec), nil
	}
	return nil, fmt.Errorf("unknown token store type %s", sts.Type)
}
