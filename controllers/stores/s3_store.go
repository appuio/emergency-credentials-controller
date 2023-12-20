package stores

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/appuio/emergency-credentials-controller/pkg/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/multierr"

	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
)

// MinioClient partially implements the minio.Client interface.
type MinioClient interface {
	PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (info minio.UploadInfo, err error)
}

type S3Store struct {
	minioClientFactory func(emcv1beta1.S3StoreSpec) (MinioClient, error)
	spec               emcv1beta1.S3StoreSpec
}

var _ TokenStorer = &S3Store{}

// NewS3Store creates a new S3Store
func NewS3Store(spec emcv1beta1.S3StoreSpec) *S3Store {
	return NewS3StoreWithClientFactory(spec, DefaultClientFactory)
}

// NewS3StoreWithClientFactory creates a new S3Store with the given client factory.
func NewS3StoreWithClientFactory(spec emcv1beta1.S3StoreSpec, minioClientFactory func(emcv1beta1.S3StoreSpec) (MinioClient, error)) *S3Store {
	return &S3Store{spec: spec, minioClientFactory: minioClientFactory}
}

// DefaultClientFactory is the default factory for creating a MinioClient.
func DefaultClientFactory(spec emcv1beta1.S3StoreSpec) (MinioClient, error) {
	return minio.New(spec.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(spec.S3.AccessKeyId, spec.S3.SecretAccessKey, ""),
		Secure: !spec.S3.Insecure,
		Region: spec.S3.Region,
	})
}

// StoreToken stores the token in the S3 bucket.
// If encryption is enabled, the token is encrypted with the given PGP public keys.
func (ss *S3Store) StoreToken(ctx context.Context, ea emcv1beta1.EmergencyAccount, token string) (string, error) {
	objectname := ea.Name
	if ss.spec.ObjectNameTemplate != "" {
		t, err := template.New("fileName").Funcs(sprig.TxtFuncMap()).Parse(ss.spec.ObjectNameTemplate)
		if err != nil {
			return "", fmt.Errorf("unable to parse file name template: %w", err)
		}
		buf := new(strings.Builder)
		if err := t.Execute(buf, struct {
			Name             string
			Namespace        string
			EmergencyAccount emcv1beta1.EmergencyAccount
			Context          map[string]string
		}{
			Name:             ea.Name,
			Namespace:        ea.Namespace,
			EmergencyAccount: ea,
			Context:          ss.spec.ObjectNameTemplateContext,
		}); err != nil {
			return "", fmt.Errorf("unable to execute file name template: %w", err)
		}
		objectname = buf.String()
	}

	cli, err := ss.minioClientFactory(ss.spec)
	if err != nil {
		return "", fmt.Errorf("unable to create S3 client: %w", err)
	}

	if ss.spec.Encryption.Encrypt {
		token, err = encrypt(token, ss.spec.Encryption.PGPKeys)
		if err != nil {
			return "", fmt.Errorf("unable to encrypt token: %w", err)
		}
	}

	tr := strings.NewReader(token)
	info, err := cli.PutObject(ctx, ss.spec.S3.Bucket, objectname, tr, int64(tr.Len()), minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to store token: %w", err)
	}

	return info.Key, nil
}

// EncryptedToken is the JSON structure of an encrypted token.
type EncryptedToken struct {
	Secrets []EncryptedTokenSecret `json:"secrets"`
}

// EncryptedTokenSecret is the JSON structure of an encrypted token secret.
type EncryptedTokenSecret struct {
	Data string `json:"data"`
}

// encrypt encrypts the token with the given PGP public keys.
// The token is encrypted with each key and the resulting encrypted tokens are returned as a JSON array.
func encrypt(token string, pgpKeys []string) (string, error) {
	if len(pgpKeys) == 0 {
		return "", fmt.Errorf("no PGP public keys given")
	}
	keys := []string{}
	for _, key := range pgpKeys {
		sk, err := utils.SplitPublicKeyBlocks(key)
		if err != nil {
			return "", fmt.Errorf("unable to parse PGP public key: %w", err)
		}
		keys = append(keys, sk...)
	}

	encrypted := make([]EncryptedTokenSecret, 0, len(keys))
	errs := []error{}
	for _, key := range keys {
		enc, err := helper.EncryptMessageArmored((key), token)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		encrypted = append(encrypted, EncryptedTokenSecret{Data: enc})
	}
	if multierr.Combine(errs...) != nil {
		return "", fmt.Errorf("unable to fully encrypt token: %w", multierr.Combine(errs...))
	}

	s, err := json.Marshal(EncryptedToken{
		Secrets: encrypted,
	})
	if err != nil {
		return "", fmt.Errorf("unable to marshal encrypted token: %w", err)
	}

	return string(s), nil
}
