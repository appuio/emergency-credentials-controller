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
}

// TokenStore defines the store the created tokens are stored in
type TokenStoreSpec struct {
	// Name is the name of the store.
	// Must be unique within the EmergencyAccount
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Type defines the type of the store to use.
	// Currently `secret`` and `log` stores are supported.
	// The stores can be further configured in the corresponding storeSpec.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=secret;log
	Type string `json:"type"`

	// SecretSpec configures the secret store.
	// The secret store saves the tokens in a secret in the same namespace as the EmergencyAccount.
	SecretSpec SecretStoreSpec `json:"secretStore,omitempty"`
	// LogSpec configures the log store.
	// The log store outputs the token to the log but does not store it anywhere.
	LogSpec LogStoreSpec `json:"logStore,omitempty"`
}

// SecretStoreSpec configures the secret store.
// The secret store saves the tokens in a secret in the same namespace as the EmergencyAccount.
type SecretStoreSpec struct{}

// LogStoreSpec configures the log store.
// The log store outputs the token to the log but does not store it anywhere.
type LogStoreSpec struct{}

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
