package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EmergencyAccountSpec defines the desired state of EmergencyAccount
type EmergencyAccountSpec struct {
	// ValidityDuration is the duration for which the tokens are valid.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:default:="720h"
	ValidityDuration metav1.Duration `json:"validityDuration"`

	// MinValidityDurationLeft is the minimum duration the token must be valid.
	// A new token is created if the current token is not valid for this duration anymore.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:default:="168h"
	// +kubebuilder:validation:Optional
	MinValidityDurationLeft metav1.Duration `json:"minValidityDurationLeft,omitempty"`

	// CheckInterval is the interval in which the tokens are checked for validity.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:default:="5m"
	CheckInterval metav1.Duration `json:"checkInterval,omitempty"`
	// MinRecreateInterval is the minimum interval in which a new token is created.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:default:="5m"
	MinRecreateInterval metav1.Duration `json:"minRecreateInterval,omitempty"`

	// TokenStore defines the stores the created tokens are stored in.
	// +kubebuilder:validation:MinItems=1
	TokenStores []TokenStoreSpec `json:"tokenStores,omitempty"`
}

// EmergencyAccountStatus defines the observed state of EmergencyAccount
type EmergencyAccountStatus struct {
	// LastTokenCreationTimestamp is the timestamp when the last token was created.
	LastTokenCreationTimestamp metav1.Time `json:"lastTokenCreationTimestamp,omitempty"`
	// Tokens is a list of tokens that have been created
	Tokens []TokenStatus `json:"tokens,omitempty"`
	// LastTokenStoreConfigurationHashes is the hash of the last token store configuration.
	// It is used to detect changes in the token store configuration.
	// A change in the configuration triggers the creation of a new token.
	LastTokenStoreHashes []TokenStoreHash `json:"lastTokenStoreConfigurationHashes,omitempty"`
}

// TokenStore defines the store the created tokens are stored in
type TokenStoreSpec struct {
	// Name is the name of the store.
	// Must be unique within the EmergencyAccount
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Type defines the type of the store to use.
	// Currently `secret`, `s3`, and `log` stores are supported.
	// The stores can be further configured in the corresponding storeSpec.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=secret;log;s3
	Type string `json:"type"`

	// SecretSpec configures the secret store.
	// The secret store saves the tokens in a secret in the same namespace as the EmergencyAccount.
	SecretSpec SecretStoreSpec `json:"secretStore,omitempty"`
	// LogSpec configures the log store.
	// The log store outputs the token to the log but does not store it anywhere.
	LogSpec LogStoreSpec `json:"logStore,omitempty"`
	// S3Spec configures the S3 store.
	// The S3 store saves the tokens in an S3 bucket.
	S3Spec S3StoreSpec `json:"s3Store,omitempty"`
}

// S3StoreSpec configures the S3 store.
// The S3 store saves the tokens in an S3 bucket with optional encryption using PGP public keys.
type S3StoreSpec struct {
	// ObjectNameTemplate is the template for the object name to use.
	// Sprig functions can be used to generate the object name.
	// If not set, the object name is the name of the EmergencyAccount.
	// The name of the EmergencyAccount can be accessed with `{{ .Name }}`.
	// The namespace of the EmergencyAccount can be accessed with `{{ .Namespace }}`.
	// The full EmergencyAccount object can be accessed with `{{ .EmergencyAccount }}`.
	// Additional context can be passed with the `objectNameTemplateContext` field and is accessible with `{{ .Context.<key> }}`.
	// +kubebuilder:validation:Optional
	ObjectNameTemplate string `json:"objectNameTemplate,omitempty"`
	// ObjectNameTemplateContext is the additional context to use for the object name template.
	// +kubebuilder:validation:Optional
	ObjectNameTemplateContext map[string]string `json:"objectNameTemplateContext,omitempty"`

	S3 S3Spec `json:"s3"`
	// Encryption defines the encryption settings for the S3 store.
	// If not set, the tokens are stored unencrypted.
	// +kubebuilder:validation:Optional
	Encryption S3EncryptionSpec `json:"encryption,omitempty"`
}

type S3Spec struct {
	// Endpoint is the S3 endpoint to use.
	Endpoint string `json:"endpoint"`
	// Bucket is the S3 bucket to use.
	Bucket string `json:"bucket"`

	// AccessKeyId and SecretAccessKey are the S3 credentials to use.
	AccessKeyId string `json:"accessKeyId"`
	// SecretAccessKey is the S3 secret access key to use.
	SecretAccessKey string `json:"secretAccessKey"`

	// Region is the AWS region to use.
	Region string `json:"region,omitempty"`
	// Insecure allows to use an insecure connection to the S3 endpoint.
	Insecure bool `json:"insecure,omitempty"`
}

type S3EncryptionSpec struct {
	// Encrypt defines if the tokens should be encrypted.
	// If not set, the tokens are stored unencrypted.
	Encrypt bool `json:"encrypt,omitempty"`
	// PGPKeys is a list of PGP public keys to encrypt the tokens with.
	// At least one key must be given if encryption is enabled.
	PGPKeys []string `json:"pgpKeys,omitempty"`
}

// SecretStoreSpec configures the secret store.
// The secret store saves the tokens in a secret in the same namespace as the EmergencyAccount.
type SecretStoreSpec struct{}

// LogStoreSpec configures the log store.
// The log store outputs the token to the log but does not store it anywhere.
type LogStoreSpec struct {
	// AdditionalFields is a map of additional fields to log.
	AdditionalFields map[string]string `json:"additionalFields,omitempty"`
}

// TokenStatus defines the observed state of the managed token
type TokenStatus struct {
	// UID is the unique identifier of the token.
	// Currently only used for error messages.
	UID types.UID `json:"uid,omitempty"`
	// Refs holds references to the token in the configured stores.
	Refs []TokenStatusRef `json:"refs,omitempty"`
	// ExpirationTimestamp is the timestamp when the token expires
	ExpirationTimestamp metav1.Time `json:"expirationTimestamp"`
}

type TokenStatusRef struct {
	// Ref is a reference to the token. The used storage should be able to uniquely identify the token.
	// If no ref is given, the token is not checked for validity.
	// +kubebuilder:validation:Optional
	Ref string `json:"ref"`

	// Store is the name of the store the token is stored in.
	Store string `json:"store"`
}

type TokenStoreHash struct {
	// Name is the name of the store.
	Name string `json:"name"`
	// Sha256 is the hash of the store configuration.
	Sha256 string `json:"hash"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EmergencyAccount is the Schema for the emergencyaccounts API
type EmergencyAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EmergencyAccountSpec   `json:"spec,omitempty"`
	Status EmergencyAccountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EmergencyAccountList contains a list of EmergencyAccount
type EmergencyAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EmergencyAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EmergencyAccount{}, &EmergencyAccountList{})
}
