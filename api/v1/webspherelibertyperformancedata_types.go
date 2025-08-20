package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebSphereLibertyPerformanceDataSpec defines the desired state of WebSphereLibertyPerformanceData
type WebSphereLibertyPerformanceDataSpec struct {

	// License information is required.
	// +operator-sdk:csv:customresourcedefinitions:order=1,type=spec,displayName="License",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	License LicenseSimple `json:"license"`

	// The name of the Pod, which must be in the same namespace as the WebSphereLibertyPerformanceData CR.
	PodName string `json:"podName"`

	// The total time, in seconds, for gathering performance data. The minimum value is 10 seconds. The maximum value is 600 seconds (10 minutes). Defaults to 240 seconds (4 minutes).
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=600
	Timespan *int `json:"timespan,omitempty"`

	// The time, in seconds, between executions. The minimum value is 1 second. Defaults to 30 seconds.
	// +kubebuilder:validation:Minimum=1
	Interval *int `json:"interval,omitempty"`
}

// Defines the observed state of WebSphereLibertyPerformanceData
type WebSphereLibertyPerformanceDataStatus struct {
	// +listType=atomic
	Conditions []OperationStatusCondition    `json:"conditions,omitempty"`
	Versions   PerformanceDataStatusVersions `json:"versions,omitempty"`
	// Location of the generated performance data file
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Performance Data File Path",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	PerformanceDataFile string `json:"performanceDataFile,omitempty"`
	// The generation identifier of this WebSphereLibertyPerformanceData instance completely reconciled by the Operator.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

type PerformanceDataStatusVersions struct {
	Reconciled string `json:"reconciled,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:path=webspherelibertyperformancedata,scope=Namespaced,shortName=olperfdata
// +kubebuilder:printcolumn:name="Started",type="string",JSONPath=".status.conditions[?(@.type=='Started')].status",priority=0,description="Indicates if performance data operation has started"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type=='Started')].reason",priority=1,description="Reason for performance data operation failing to start"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type=='Started')].message",priority=1,description="Message for performance data operation failing to start"
// +kubebuilder:printcolumn:name="Completed",type="string",JSONPath=".status.conditions[?(@.type=='Completed')].status",priority=0,description="Indicates if performance data operation has completed"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type=='Completed')].reason",priority=1,description="Reason for performance data operation failing to complete"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type=='Completed')].message",priority=1,description="Message for performance data operation failing to complete"
// +kubebuilder:printcolumn:name="Performance Data file",type="string",JSONPath=".status.performanceDataFile",priority=0,description="Indicates filename of the server performance data"
// +operator-sdk:csv:customresourcedefinitions:displayName="WebSphereLibertyPerformanceData"
// Day-2 operation for generating server performance data
type WebSphereLibertyPerformanceData struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebSphereLibertyPerformanceDataSpec   `json:"spec,omitempty"`
	Status WebSphereLibertyPerformanceDataStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// WebSphereLibertyPerformanceDataList contains a list of WebSphereLibertyPerformanceData
type WebSphereLibertyPerformanceDataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebSphereLibertyPerformanceData `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebSphereLibertyPerformanceData{}, &WebSphereLibertyPerformanceDataList{})
}

func getIntValueOrDefault(value *int, defaultValue int) int {
	if value == nil {
		return defaultValue
	}
	return *value
}

// GetTimespan returns the timespan in seconds for running a performance data operation. Defaults to 240.
func (cr *WebSphereLibertyPerformanceData) GetTimespan() int {
	defaultTimespan := 240
	return getIntValueOrDefault(cr.Spec.Timespan, defaultTimespan)
}

// GetInterval returns the time interval in seconds between performance data operations. Defaults to 30.
func (cr *WebSphereLibertyPerformanceData) GetInterval() int {
	defaultInterval := 30
	return getIntValueOrDefault(cr.Spec.Interval, defaultInterval)
}
