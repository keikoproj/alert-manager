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

// WavefrontAlertSpec defines the desired state of WavefrontAlert
type WavefrontAlertSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// AlertType represents the type of the Alert in Wavefront. Defaults to CLASSIC alert
	// +optional
	AlertType AlertType `json:"alertType,omitempty"`

	//Name of the alert to be created in Wavefront
	// +required
	AlertName string `json:"alertName"`

	//A conditional expression that defines the threshold for the Classic alert. For CLASSIC (or default alerts) condition must be provided
	// +required
	Condition string `json:"condition"`

	//For classic alert type, mention the severity of the incident. This will be ignored for threshold type of alerts
	// +required
	Severity string `json:"severity"`

	//Minutes where alert is in "true" state continuously to trigger an alert
	// +required
	Minutes *int32 `json:"minutes"`

	//Minutes after the alert got back to "false" state to resolve the incident
	// +required
	ResolveAfter *int32 `json:"resolveAfterMinutes"`

	//Target (Optional) A comma-separated list of the email address or integration endpoint (such as PagerDuty or web hook)
	// to notify when the alert status changes.
	// Multiple target types can be in the list. Alert target format: ({email}|pd:{pd_key}
	// +optional
	Target string `json:"target,omitempty"`

	//Any additional information, such as a link to a run book.
	// +optional
	AdditionalInformation string `json:"additionalInformation,omitempty"`

	//Tags assigned to the alert.
	// +optional
	Tags []string `json:"tags,omitempty"`

	//Describe the functionality of the alert in simple words. This is just for CR and not used it to send it to wavefront
	Description string `json:"description,omitempty"`

	//Specify a display expression to get more details when the alert changes state
	// +required
	DisplayExpression string `json:"displayExpression"`

	//exportedParams can be used when AlertsConfig CRD used to provide config to WavefrontAlert CRD at the runtime for multiple alerts
	//when the exportedParams length is not empty, Alert will not be created when Alert CR is created but rather alerts will be created when AlertsConfig CR created.
	// +optional
	ExportedParams []string `json:"exportedParams,omitempty"`
	//exportedParamsDefaultValues can be used to provide the default values and will be used if alerts config doesn't provide any values. This could be useful if user
	// wants to use go lang template for a field but majority of the alerts can use the default values instead of providing in each and every alert config files.
	// +optional
	ExportedParamsDefaultValues OrderedMap `json:"exportedParamsDefaultValues,omitempty"`
}

// AlertType represents the type of the Alert in Wavefront. Defaults to CLASSIC alert
// +kubebuilder:default=CLASSIC
// +kubebuilder:validation:Enum=CLASSIC;THRESHOLD
type AlertType string

const (
	// Wavefront Classic Alert. Defaults to Classic Alert if none specified. For more info about CLASSIC Alert type:https://docs.wavefront.com/alerts.html
	ClassicAlert AlertType = "CLASSIC"

	// Wavefront Threshold Alert. For more info about THRESHOLD alert type: https://docs.wavefront.com/alerts.html
	ThresholdAlert AlertType = "THRESHOLD"
)

type State string

const (
	Ready               State = "Ready"
	Error               State = "Error"
	MalformedSpec       State = "MalformedSpec"
	ReadyToBeUsed       State = "ReadyToBeUsed"
	ClientExceededLimit State = "ClientExceededLimit"
	Creating            State = "Creating"
	Updating            State = "Updating"
	Deleting            State = "Deleting"
)

// WavefrontAlertStatus defines the observed state of WavefrontAlert
type WavefrontAlertStatus struct {
	//State of the resource
	State State `json:"state,omitempty"`
	//RetryCount in case of error
	RetryCount int `json:"retryCount,omitempty"`
	//ErrorDescription in case of error
	ErrorDescription string `json:"errorDescription,omitempty"`
	//Checksum of the exportedParams if exists
	ExportParamsChecksum string `json:"exportParamsChecksum,omitempty"`
	//This represents the checksum of the spec
	LastChangeChecksum string `json:"lastChangeChecksum,omitempty"`
	//ObservedGeneration will have the last generation from spec metadata
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	//AlertsStatus details includes individual alert details
	AlertsStatus map[string]AlertStatus `json:"alertsStatus,omitempty"`
}

//AlertStatus consists of individual alert details
type AlertStatus struct {
	ID                     string                 `json:"id"`
	Name                   string                 `json:"alertName"`
	Link                   string                 `json:"link,omitempty"`
	State                  State                  `json:"state,omitempty"`
	LastChangeChecksum     string                 `json:"lastChangeChecksum,omitempty"`
	AssociatedAlert        AssociatedAlert        `json:"associatedAlert,omitempty"`
	AssociatedAlertsConfig AssociatedAlertsConfig `json:"associatedAlertsConfig,omitempty"`
}

type AssociatedAlertsConfig struct {
	CR string `json:"CR,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WavefrontAlert is the Schema for the wavefrontalerts API
type WavefrontAlert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WavefrontAlertSpec   `json:"spec,omitempty"`
	Status WavefrontAlertStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WavefrontAlertList contains a list of WavefrontAlert
type WavefrontAlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WavefrontAlert `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WavefrontAlert{}, &WavefrontAlertList{})
}
