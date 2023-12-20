package stores_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	emcv1beta1 "github.com/appuio/emergency-credentials-controller/api/v1beta1"
	"github.com/appuio/emergency-credentials-controller/controllers/stores"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_S3Store_StoreToken(t *testing.T) {
	const (
		token      = "token"
		bucket     = "bucket"
		object     = "object"
		passphrase = "passphrase"
	)

	t.Run("without encryption", func(t *testing.T) {
		mm := &MinioMock{}
		st := stores.NewS3StoreWithClientFactory(emcv1beta1.S3StoreSpec{
			S3: emcv1beta1.S3Spec{
				Bucket: bucket,
			},
			ObjectNameTemplate: "em-{{ .Name | sha256sum }}",
		}, mm.ClientFactory)

		_, err := st.StoreToken(context.Background(), emcv1beta1.EmergencyAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: object,
			},
		}, token)
		require.NoError(t, err)
		stored := mm.get(bucket, fmt.Sprintf("em-%x", sha256.Sum256([]byte(object))))
		require.NotEmpty(t, stored)
		require.Equal(t, []byte(token), stored)
	})

	t.Run("encrypted", func(t *testing.T) {
		privk1, pubk1, err := generateKeyPair("test1", "test1@test.ch", passphrase, "rsa", 2048)
		require.NoError(t, err)
		privk2, pubk2, err := generateKeyPair("test2", "test2@test.ch", passphrase, "rsa", 2048)
		require.NoError(t, err)
		privk3, pubk3, err := generateKeyPair("test3", "test3@test.ch", passphrase, "rsa", 2048)
		require.NoError(t, err)

		mm := &MinioMock{}
		st := stores.NewS3StoreWithClientFactory(emcv1beta1.S3StoreSpec{
			S3: emcv1beta1.S3Spec{
				Bucket: bucket,
			},
			Encryption: emcv1beta1.S3EncryptionSpec{
				Encrypt: true,
				PGPKeys: []string{strings.Join([]string{pubk1, pubk2}, "\n"), pubk3},
			},
		}, mm.ClientFactory)

		_, err = st.StoreToken(context.Background(), emcv1beta1.EmergencyAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: object,
			},
		}, token)
		require.NoError(t, err)
		requireDecryptAll(t, string(mm.get(bucket, object)), token, passphrase, []string{privk1, privk2, privk3})
	})
}

func requireDecryptAll(t *testing.T, token, expectedMsg, passphrase string, keys []string) {
	t.Helper()

	var data stores.EncryptedToken
	err := json.Unmarshal([]byte(token), &data)
	require.NoError(t, err)

	for _, secret := range data.Secrets {
		requireDecrypt(t, secret.Data, expectedMsg, passphrase, keys)
	}
}

func requireDecrypt(t *testing.T, encrypted, expectedMsg, passphrase string, keys []string) {
	t.Helper()

	for _, key := range keys {
		msg, err := helper.DecryptMessageArmored(key, []byte(passphrase), encrypted)
		if err == nil {
			require.Equal(t, expectedMsg, string(msg))
			return
		}
	}
	require.Fail(t, "expected to decrypt token with one of the given private keys")
}

type MinioMock struct {
	files map[string]map[string][]byte
}

// ClientFactory returns itself.
func (mm *MinioMock) ClientFactory(emcv1beta1.S3StoreSpec) (stores.MinioClient, error) {
	return mm, nil
}

// PutObject implements the MinioClient interface.
func (mm *MinioMock) PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (info minio.UploadInfo, err error) {
	info.Bucket = bucketName
	info.Key = objectName

	if mm.files == nil {
		mm.files = make(map[string]map[string][]byte)
	}
	if mm.files[bucketName] == nil {
		mm.files[bucketName] = make(map[string][]byte)
	}
	buf := make([]byte, objectSize)
	_, err = reader.Read(buf)
	if err != nil {
		return info, err
	}
	mm.files[bucketName][objectName] = buf
	return info, nil
}

func (mm *MinioMock) get(bucketName string, objectName string) []byte {
	if mm.files == nil {
		return nil
	}

	if mm.files[bucketName] == nil {
		return nil
	}

	return mm.files[bucketName][objectName]
}

// generateKeyPair generates a key pair and returns the private and public key.
func generateKeyPair(name, email, passphrase string, keyType string, bits int) (privateKey string, publicKey string, err error) {
	privateKey, err = helper.GenerateKey(name, email, []byte(passphrase), keyType, bits)
	if err != nil {
		return "", "", err
	}

	ring, err := crypto.NewKeyFromArmoredReader(strings.NewReader(privateKey))
	if err != nil {
		return "", "", err
	}

	publicKey, err = ring.GetArmoredPublicKey()
	if err != nil {
		return "", "", err
	}

	return privateKey, publicKey, nil
}
