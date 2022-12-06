//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutoscalingStrategy) DeepCopyInto(out *AutoscalingStrategy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutoscalingStrategy.
func (in *AutoscalingStrategy) DeepCopy() *AutoscalingStrategy {
	if in == nil {
		return nil
	}
	out := new(AutoscalingStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AutoscalingStrategy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutoscalingStrategyList) DeepCopyInto(out *AutoscalingStrategyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AutoscalingStrategy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutoscalingStrategyList.
func (in *AutoscalingStrategyList) DeepCopy() *AutoscalingStrategyList {
	if in == nil {
		return nil
	}
	out := new(AutoscalingStrategyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AutoscalingStrategyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutoscalingStrategySpec) DeepCopyInto(out *AutoscalingStrategySpec) {
	*out = *in
	if in.AutoscalerExtraArgs != nil {
		in, out := &in.AutoscalerExtraArgs, &out.AutoscalerExtraArgs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutoscalingStrategySpec.
func (in *AutoscalingStrategySpec) DeepCopy() *AutoscalingStrategySpec {
	if in == nil {
		return nil
	}
	out := new(AutoscalingStrategySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutoscalingStrategyStatus) DeepCopyInto(out *AutoscalingStrategyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutoscalingStrategyStatus.
func (in *AutoscalingStrategyStatus) DeepCopy() *AutoscalingStrategyStatus {
	if in == nil {
		return nil
	}
	out := new(AutoscalingStrategyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JoiningNode) DeepCopyInto(out *JoiningNode) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JoiningNode.
func (in *JoiningNode) DeepCopy() *JoiningNode {
	if in == nil {
		return nil
	}
	out := new(JoiningNode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JoiningNode) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JoiningNodeList) DeepCopyInto(out *JoiningNodeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]JoiningNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JoiningNodeList.
func (in *JoiningNodeList) DeepCopy() *JoiningNodeList {
	if in == nil {
		return nil
	}
	out := new(JoiningNodeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JoiningNodeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JoiningNodeSpec) DeepCopyInto(out *JoiningNodeSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JoiningNodeSpec.
func (in *JoiningNodeSpec) DeepCopy() *JoiningNodeSpec {
	if in == nil {
		return nil
	}
	out := new(JoiningNodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JoiningNodeStatus) DeepCopyInto(out *JoiningNodeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JoiningNodeStatus.
func (in *JoiningNodeStatus) DeepCopy() *JoiningNodeStatus {
	if in == nil {
		return nil
	}
	out := new(JoiningNodeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeImage) DeepCopyInto(out *NodeImage) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeImage.
func (in *NodeImage) DeepCopy() *NodeImage {
	if in == nil {
		return nil
	}
	out := new(NodeImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeImage) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeImageList) DeepCopyInto(out *NodeImageList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeImage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeImageList.
func (in *NodeImageList) DeepCopy() *NodeImageList {
	if in == nil {
		return nil
	}
	out := new(NodeImageList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeImageList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeImageSpec) DeepCopyInto(out *NodeImageSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeImageSpec.
func (in *NodeImageSpec) DeepCopy() *NodeImageSpec {
	if in == nil {
		return nil
	}
	out := new(NodeImageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeImageStatus) DeepCopyInto(out *NodeImageStatus) {
	*out = *in
	if in.Outdated != nil {
		in, out := &in.Outdated, &out.Outdated
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.UpToDate != nil {
		in, out := &in.UpToDate, &out.UpToDate
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Donors != nil {
		in, out := &in.Donors, &out.Donors
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Heirs != nil {
		in, out := &in.Heirs, &out.Heirs
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Mints != nil {
		in, out := &in.Mints, &out.Mints
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Pending != nil {
		in, out := &in.Pending, &out.Pending
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Obsolete != nil {
		in, out := &in.Obsolete, &out.Obsolete
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Invalid != nil {
		in, out := &in.Invalid, &out.Invalid
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeImageStatus.
func (in *NodeImageStatus) DeepCopy() *NodeImageStatus {
	if in == nil {
		return nil
	}
	out := new(NodeImageStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PendingNode) DeepCopyInto(out *PendingNode) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PendingNode.
func (in *PendingNode) DeepCopy() *PendingNode {
	if in == nil {
		return nil
	}
	out := new(PendingNode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PendingNode) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PendingNodeList) DeepCopyInto(out *PendingNodeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PendingNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PendingNodeList.
func (in *PendingNodeList) DeepCopy() *PendingNodeList {
	if in == nil {
		return nil
	}
	out := new(PendingNodeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PendingNodeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PendingNodeSpec) DeepCopyInto(out *PendingNodeSpec) {
	*out = *in
	if in.Deadline != nil {
		in, out := &in.Deadline, &out.Deadline
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PendingNodeSpec.
func (in *PendingNodeSpec) DeepCopy() *PendingNodeSpec {
	if in == nil {
		return nil
	}
	out := new(PendingNodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PendingNodeStatus) DeepCopyInto(out *PendingNodeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PendingNodeStatus.
func (in *PendingNodeStatus) DeepCopy() *PendingNodeStatus {
	if in == nil {
		return nil
	}
	out := new(PendingNodeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScalingGroup) DeepCopyInto(out *ScalingGroup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScalingGroup.
func (in *ScalingGroup) DeepCopy() *ScalingGroup {
	if in == nil {
		return nil
	}
	out := new(ScalingGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ScalingGroup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScalingGroupList) DeepCopyInto(out *ScalingGroupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ScalingGroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScalingGroupList.
func (in *ScalingGroupList) DeepCopy() *ScalingGroupList {
	if in == nil {
		return nil
	}
	out := new(ScalingGroupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ScalingGroupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScalingGroupSpec) DeepCopyInto(out *ScalingGroupSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScalingGroupSpec.
func (in *ScalingGroupSpec) DeepCopy() *ScalingGroupSpec {
	if in == nil {
		return nil
	}
	out := new(ScalingGroupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScalingGroupStatus) DeepCopyInto(out *ScalingGroupStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScalingGroupStatus.
func (in *ScalingGroupStatus) DeepCopy() *ScalingGroupStatus {
	if in == nil {
		return nil
	}
	out := new(ScalingGroupStatus)
	in.DeepCopyInto(out)
	return out
}
