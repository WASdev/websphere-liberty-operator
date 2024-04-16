/*
  Copyright contributors to the WASdev project.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebSphereLibertyDumpSpec defines the desired state of WebSphereLibertyDump
type WebSphereLibertyDumpSpec struct {

	// License information is required.
	// +operator-sdk:csv:customresourcedefinitions:order=1,type=spec,displayName="License",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	License LicenseSimple `json:"license"`

	// The name of the Pod, which must be in the same namespace as the WebSphereLibertyDump CR.
	PodName string `json:"podName"`
	// Optional. List of memory dump types to request: thread, heap, system.
	// +listType=set
	Include []WebSphereLibertyDumpInclude `json:"include,omitempty"`
}

// Defines the possible values for dump types
// +kubebuilder:validation:Enum=thread;heap;system
type WebSphereLibertyDumpInclude string

const (
	//WebSphereLibertyDumpIncludeHeap heap dump
	WebSphereLibertyDumpIncludeHeap WebSphereLibertyDumpInclude = "heap"
	//WebSphereLibertyDumpIncludeThread thread dump
	WebSphereLibertyDumpIncludeThread WebSphereLibertyDumpInclude = "thread"
	//WebSphereLibertyDumpIncludeSystem system (core) dump
	WebSphereLibertyDumpIncludeSystem WebSphereLibertyDumpInclude = "system"
)

// Defines the observed state of WebSphereLibertyDump
type WebSphereLibertyDumpStatus struct {
	// +listType=atomic
	Conditions []OperationStatusCondition `json:"conditions,omitempty"`
	Versions   DumpStatusVersions         `json:"versions,omitempty"`
	// Location of the generated dump file
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Dump File Path",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	DumpFile string `json:"dumpFile,omitempty"`
	// The last generation of this WebSphereLibertyDump instance observed by the operator.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

type DumpStatusVersions struct {
	Reconciled string `json:"reconciled,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=webspherelibertydumps,scope=Namespaced,shortName=wldump;wldumps
// +kubebuilder:printcolumn:name="Started",type="string",JSONPath=".status.conditions[?(@.type=='Started')].status",priority=0,description="Indicates if dump operation has started"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type=='Started')].reason",priority=1,description="Reason for dump operation failing to start"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type=='Started')].message",priority=1,description="Message for dump operation failing to start"
// +kubebuilder:printcolumn:name="Completed",type="string",JSONPath=".status.conditions[?(@.type=='Completed')].status",priority=0,description="Indicates if dump operation has completed"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type=='Completed')].reason",priority=1,description="Reason for dump operation failing to complete"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type=='Completed')].message",priority=1,description="Message for dump operation failing to complete"
// +kubebuilder:printcolumn:name="Dump file",type="string",JSONPath=".status.dumpFile",priority=0,description="Indicates filename of the server dump"
// +operator-sdk:csv:customresourcedefinitions:displayName="WebSphereLibertyDump"
// Day-2 operation for generating server dumps. Documentation: For more information about installation parameters, see https://ibm.biz/wlo-crs. License: By installing this product, you accept the license terms at https://ibm.biz/was-license.
type WebSphereLibertyDump struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebSphereLibertyDumpSpec   `json:"spec,omitempty"`
	Status WebSphereLibertyDumpStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// WebSphereLibertyDumpList contains a list of WebSphereLibertyDump
type WebSphereLibertyDumpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebSphereLibertyDump `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WebSphereLibertyDump{}, &WebSphereLibertyDumpList{})
}
