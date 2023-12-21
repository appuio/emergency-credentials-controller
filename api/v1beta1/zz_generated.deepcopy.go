//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmergencyAccount) DeepCopyInto(out *EmergencyAccount) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmergencyAccount.
func (in *EmergencyAccount) DeepCopy() *EmergencyAccount {
	if in == nil {
		return nil
	}
	out := new(EmergencyAccount)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EmergencyAccount) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmergencyAccountList) DeepCopyInto(out *EmergencyAccountList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]EmergencyAccount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmergencyAccountList.
func (in *EmergencyAccountList) DeepCopy() *EmergencyAccountList {
	if in == nil {
		return nil
	}
	out := new(EmergencyAccountList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EmergencyAccountList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmergencyAccountSpec) DeepCopyInto(out *EmergencyAccountSpec) {
	*out = *in
	out.ValidityDuration = in.ValidityDuration
	out.MinValidityDurationLeft = in.MinValidityDurationLeft
	out.CheckInterval = in.CheckInterval
	out.MinRecreateInterval = in.MinRecreateInterval
	if in.TokenStores != nil {
		in, out := &in.TokenStores, &out.TokenStores
		*out = make([]TokenStoreSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmergencyAccountSpec.
func (in *EmergencyAccountSpec) DeepCopy() *EmergencyAccountSpec {
	if in == nil {
		return nil
	}
	out := new(EmergencyAccountSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmergencyAccountStatus) DeepCopyInto(out *EmergencyAccountStatus) {
	*out = *in
	in.LastTokenCreationTimestamp.DeepCopyInto(&out.LastTokenCreationTimestamp)
	if in.Tokens != nil {
		in, out := &in.Tokens, &out.Tokens
		*out = make([]TokenStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmergencyAccountStatus.
func (in *EmergencyAccountStatus) DeepCopy() *EmergencyAccountStatus {
	if in == nil {
		return nil
	}
	out := new(EmergencyAccountStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogStoreSpec) DeepCopyInto(out *LogStoreSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogStoreSpec.
func (in *LogStoreSpec) DeepCopy() *LogStoreSpec {
	if in == nil {
		return nil
	}
	out := new(LogStoreSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3EncryptionSpec) DeepCopyInto(out *S3EncryptionSpec) {
	*out = *in
	if in.PGPKeys != nil {
		in, out := &in.PGPKeys, &out.PGPKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3EncryptionSpec.
func (in *S3EncryptionSpec) DeepCopy() *S3EncryptionSpec {
	if in == nil {
		return nil
	}
	out := new(S3EncryptionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3Spec) DeepCopyInto(out *S3Spec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3Spec.
func (in *S3Spec) DeepCopy() *S3Spec {
	if in == nil {
		return nil
	}
	out := new(S3Spec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3StoreSpec) DeepCopyInto(out *S3StoreSpec) {
	*out = *in
	if in.ObjectNameTemplateContext != nil {
		in, out := &in.ObjectNameTemplateContext, &out.ObjectNameTemplateContext
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.S3 = in.S3
	in.Encryption.DeepCopyInto(&out.Encryption)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3StoreSpec.
func (in *S3StoreSpec) DeepCopy() *S3StoreSpec {
	if in == nil {
		return nil
	}
	out := new(S3StoreSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretStoreSpec) DeepCopyInto(out *SecretStoreSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretStoreSpec.
func (in *SecretStoreSpec) DeepCopy() *SecretStoreSpec {
	if in == nil {
		return nil
	}
	out := new(SecretStoreSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TokenStatus) DeepCopyInto(out *TokenStatus) {
	*out = *in
	if in.Refs != nil {
		in, out := &in.Refs, &out.Refs
		*out = make([]TokenStatusRef, len(*in))
		copy(*out, *in)
	}
	in.ExpirationTimestamp.DeepCopyInto(&out.ExpirationTimestamp)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TokenStatus.
func (in *TokenStatus) DeepCopy() *TokenStatus {
	if in == nil {
		return nil
	}
	out := new(TokenStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TokenStatusRef) DeepCopyInto(out *TokenStatusRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TokenStatusRef.
func (in *TokenStatusRef) DeepCopy() *TokenStatusRef {
	if in == nil {
		return nil
	}
	out := new(TokenStatusRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TokenStoreSpec) DeepCopyInto(out *TokenStoreSpec) {
	*out = *in
	out.SecretSpec = in.SecretSpec
	out.LogSpec = in.LogSpec
	in.S3Spec.DeepCopyInto(&out.S3Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TokenStoreSpec.
func (in *TokenStoreSpec) DeepCopy() *TokenStoreSpec {
	if in == nil {
		return nil
	}
	out := new(TokenStoreSpec)
	in.DeepCopyInto(out)
	return out
}
