// +build !ignore_autogenerated

/*
LICENSE
*/
// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseHistory) DeepCopyInto(out *ReleaseHistory) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseHistory.
func (in *ReleaseHistory) DeepCopy() *ReleaseHistory {
	if in == nil {
		return nil
	}
	out := new(ReleaseHistory)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReleaseHistory) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseHistoryList) DeepCopyInto(out *ReleaseHistoryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ReleaseHistory, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseHistoryList.
func (in *ReleaseHistoryList) DeepCopy() *ReleaseHistoryList {
	if in == nil {
		return nil
	}
	out := new(ReleaseHistoryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReleaseHistoryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseHistoryRevision) DeepCopyInto(out *ReleaseHistoryRevision) {
	*out = *in
	in.DeployedAt.DeepCopyInto(&out.DeployedAt)
	if in.GVKs != nil {
		in, out := &in.GVKs, &out.GVKs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseHistoryRevision.
func (in *ReleaseHistoryRevision) DeepCopy() *ReleaseHistoryRevision {
	if in == nil {
		return nil
	}
	out := new(ReleaseHistoryRevision)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseHistorySpec) DeepCopyInto(out *ReleaseHistorySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseHistorySpec.
func (in *ReleaseHistorySpec) DeepCopy() *ReleaseHistorySpec {
	if in == nil {
		return nil
	}
	out := new(ReleaseHistorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseHistoryStatus) DeepCopyInto(out *ReleaseHistoryStatus) {
	*out = *in
	in.DeployedAt.DeepCopyInto(&out.DeployedAt)
	if in.Revisions != nil {
		in, out := &in.Revisions, &out.Revisions
		*out = make([]ReleaseHistoryRevision, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseHistoryStatus.
func (in *ReleaseHistoryStatus) DeepCopy() *ReleaseHistoryStatus {
	if in == nil {
		return nil
	}
	out := new(ReleaseHistoryStatus)
	in.DeepCopyInto(out)
	return out
}
