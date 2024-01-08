//go:build !ignore_autogenerated

/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertStatus) DeepCopyInto(out *AlertStatus) {
	*out = *in
	out.AssociatedAlert = in.AssociatedAlert
	out.AssociatedAlertsConfig = in.AssociatedAlertsConfig
	in.LastUpdatedTimestamp.DeepCopyInto(&out.LastUpdatedTimestamp)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertStatus.
func (in *AlertStatus) DeepCopy() *AlertStatus {
	if in == nil {
		return nil
	}
	out := new(AlertStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertsConfig) DeepCopyInto(out *AlertsConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertsConfig.
func (in *AlertsConfig) DeepCopy() *AlertsConfig {
	if in == nil {
		return nil
	}
	out := new(AlertsConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AlertsConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertsConfigList) DeepCopyInto(out *AlertsConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AlertsConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertsConfigList.
func (in *AlertsConfigList) DeepCopy() *AlertsConfigList {
	if in == nil {
		return nil
	}
	out := new(AlertsConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AlertsConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertsConfigSpec) DeepCopyInto(out *AlertsConfigSpec) {
	*out = *in
	out.GlobalGVK = in.GlobalGVK
	if in.Alerts != nil {
		in, out := &in.Alerts, &out.Alerts
		*out = make(map[string]Config, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.GlobalParams != nil {
		in, out := &in.GlobalParams, &out.GlobalParams
		*out = make(OrderedMap, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertsConfigSpec.
func (in *AlertsConfigSpec) DeepCopy() *AlertsConfigSpec {
	if in == nil {
		return nil
	}
	out := new(AlertsConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AlertsConfigStatus) DeepCopyInto(out *AlertsConfigStatus) {
	*out = *in
	if in.AlertsStatus != nil {
		in, out := &in.AlertsStatus, &out.AlertsStatus
		*out = make(map[string]AlertStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AlertsConfigStatus.
func (in *AlertsConfigStatus) DeepCopy() *AlertsConfigStatus {
	if in == nil {
		return nil
	}
	out := new(AlertsConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AssociatedAlert) DeepCopyInto(out *AssociatedAlert) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AssociatedAlert.
func (in *AssociatedAlert) DeepCopy() *AssociatedAlert {
	if in == nil {
		return nil
	}
	out := new(AssociatedAlert)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AssociatedAlertsConfig) DeepCopyInto(out *AssociatedAlertsConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AssociatedAlertsConfig.
func (in *AssociatedAlertsConfig) DeepCopy() *AssociatedAlertsConfig {
	if in == nil {
		return nil
	}
	out := new(AssociatedAlertsConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
	out.GVK = in.GVK
	if in.Params != nil {
		in, out := &in.Params, &out.Params
		*out = make(OrderedMap, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Config.
func (in *Config) DeepCopy() *Config {
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GVK) DeepCopyInto(out *GVK) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GVK.
func (in *GVK) DeepCopy() *GVK {
	if in == nil {
		return nil
	}
	out := new(GVK)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WavefrontAlert) DeepCopyInto(out *WavefrontAlert) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WavefrontAlert.
func (in *WavefrontAlert) DeepCopy() *WavefrontAlert {
	if in == nil {
		return nil
	}
	out := new(WavefrontAlert)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WavefrontAlert) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WavefrontAlertList) DeepCopyInto(out *WavefrontAlertList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WavefrontAlert, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WavefrontAlertList.
func (in *WavefrontAlertList) DeepCopy() *WavefrontAlertList {
	if in == nil {
		return nil
	}
	out := new(WavefrontAlertList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WavefrontAlertList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WavefrontAlertSpec) DeepCopyInto(out *WavefrontAlertSpec) {
	*out = *in
	if in.Minutes != nil {
		in, out := &in.Minutes, &out.Minutes
		*out = new(int32)
		**out = **in
	}
	if in.ResolveAfter != nil {
		in, out := &in.ResolveAfter, &out.ResolveAfter
		*out = new(int32)
		**out = **in
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExportedParams != nil {
		in, out := &in.ExportedParams, &out.ExportedParams
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExportedParamsDefaultValues != nil {
		in, out := &in.ExportedParamsDefaultValues, &out.ExportedParamsDefaultValues
		*out = make(OrderedMap, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WavefrontAlertSpec.
func (in *WavefrontAlertSpec) DeepCopy() *WavefrontAlertSpec {
	if in == nil {
		return nil
	}
	out := new(WavefrontAlertSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WavefrontAlertStatus) DeepCopyInto(out *WavefrontAlertStatus) {
	*out = *in
	if in.AlertsStatus != nil {
		in, out := &in.AlertsStatus, &out.AlertsStatus
		*out = make(map[string]AlertStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WavefrontAlertStatus.
func (in *WavefrontAlertStatus) DeepCopy() *WavefrontAlertStatus {
	if in == nil {
		return nil
	}
	out := new(WavefrontAlertStatus)
	in.DeepCopyInto(out)
	return out
}
