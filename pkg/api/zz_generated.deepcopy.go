//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package api

import (
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeconfigReference) DeepCopyInto(out *KubeconfigReference) {
	*out = *in
	out.SecretReference = in.SecretReference
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeconfigReference.
func (in *KubeconfigReference) DeepCopy() *KubeconfigReference {
	if in == nil {
		return nil
	}
	out := new(KubeconfigReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceAccountConfig) DeepCopyInto(out *ServiceAccountConfig) {
	*out = *in
	if in.CAFile != nil {
		in, out := &in.CAFile, &out.CAFile
		*out = new(string)
		**out = **in
	}
	if in.CAData != nil {
		in, out := &in.CAData, &out.CAData
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceAccountConfig.
func (in *ServiceAccountConfig) DeepCopy() *ServiceAccountConfig {
	if in == nil {
		return nil
	}
	out := new(ServiceAccountConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Target) DeepCopyInto(out *Target) {
	*out = *in
	if in.Kubeconfig != nil {
		in, out := &in.Kubeconfig, &out.Kubeconfig
		*out = new(v1.JSON)
		(*in).DeepCopyInto(*out)
	}
	if in.KubeconfigFile != nil {
		in, out := &in.KubeconfigFile, &out.KubeconfigFile
		*out = new(string)
		**out = **in
	}
	if in.KubeconfigRef != nil {
		in, out := &in.KubeconfigRef, &out.KubeconfigRef
		*out = new(KubeconfigReference)
		**out = **in
	}
	if in.ServiceAccount != nil {
		in, out := &in.ServiceAccount, &out.ServiceAccount
		*out = new(ServiceAccountConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Target.
func (in *Target) DeepCopy() *Target {
	if in == nil {
		return nil
	}
	out := new(Target)
	in.DeepCopyInto(out)
	return out
}
