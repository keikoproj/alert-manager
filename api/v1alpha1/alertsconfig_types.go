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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlertsConfigSpec defines the desired state of AlertsConfig
type AlertsConfigSpec struct {
	//GlobalGVK- This is a global GVK config but user can overwrite it if an AlertsConfig supports multiple type of Alerts in future.
	//This CRD must be installed in the cluster otherwise AlertsConfig will go into error state
	GlobalGVK GVK `json:"globalGVK,omitempty"`
	//Alerts- Provide each individual alert config
	Alerts map[string]Config `json:"alerts,omitempty"`
	//GlobalParams is the place holder to provide any global param values which can be used in individual config sections.
	//Please note that if a param is mentioned in both global param section and individual config params section,
	//later will be taken into consideration and NOT the value from global param section
	// +optional
	GlobalParams OrderedMap `json:"globalParams,omitempty"`
}

//GVK struct represents the alert type and can be used as a global as well as in individual alert section
type GVK struct {
	//Group - CRD Group name which this config/s is related to
	Group string `json:"group,omitempty"`
	//Version - CRD Version name which this config/s is related to
	Version string `json:"version,omitempty"`
	//Kind - CRD Kind name which this config/s is related to
	Kind string `json:"kind,omitempty"`
}

//Config section provides the AlertsConfig for each individual alert
// +optional
type Config struct {
	//GVK can be used to provide CRD group, version and kind- If there is a global GVK already provided this will overwrite it
	// +optional
	GVK GVK `json:"gvk,omitempty"`
	//Params section can be used to provide exportParams key values
	// +optional
	Params OrderedMap `json:"params,omitempty"`
}

// AlertsConfigStatus defines the observed state of AlertsConfig
type AlertsConfigStatus struct {
	//State of the resource
	State State `json:"state,omitempty"`
	//RetryCount in case of error
	RetryCount int `json:"retryCount"`
	//AlertsCount provides total number of alerts configured
	AlertsCount int `json:"alertsCount,omitempty"`
	//ErrorDescription in case of error
	ErrorDescription string `json:"errorDescription,omitempty"`
	//AlertsStatus details includes individual alert details
	AlertsStatus map[string]AlertStatus `json:"alertsStatus,omitempty"`
}

type AssociatedAlert struct {
	CR         string `json:"CR,omitempty"`
	Generation int64  `json:"generation,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="current state of the alerts config"
// +kubebuilder:printcolumn:name="RetryCount",type="integer",JSONPath=".status.retryCount",description="Retry count"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="time passed since alerts config creation"
// AlertsConfig is the Schema for the alertsconfigs API
type AlertsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertsConfigSpec   `json:"spec,omitempty"`
	Status AlertsConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AlertsConfigList contains a list of AlertsConfig
type AlertsConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlertsConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlertsConfig{}, &AlertsConfigList{})
}
