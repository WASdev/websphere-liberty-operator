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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperationStatusCondition ...
type OperationStatusCondition struct {
	LastTransitionTime *metav1.Time                 `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     metav1.Time                  `json:"lastUpdateTime,omitempty"`
	Reason             string                       `json:"reason,omitempty"`
	Message            string                       `json:"message,omitempty"`
	Status             corev1.ConditionStatus       `json:"status,omitempty"`
	Type               OperationStatusConditionType `json:"type,omitempty"`
}

// OperatedResource ...
type OperatedResource struct {
	ResourceType string `json:"resourceType,omitempty"`
	ResourceName string `json:"resourceName,omitempty"`
}

// GetOperatedResourceName get the last operated resource name
func (or *OperatedResource) GetOperatedResourceName() string {
	return or.ResourceName
}

// SetOperatedResourceName sets the last operated resource name
func (or *OperatedResource) SetOperatedResourceName(n string) {
	or.ResourceName = n
}

// GetOperatedResourceType get the last operated resource type
func (or *OperatedResource) GetOperatedResourceType() string {
	return or.ResourceType
}

// SetOperatedResourceType sets the last operated resource type
func (or *OperatedResource) SetOperatedResourceType(t string) {
	or.ResourceType = t
}

// OperationStatusConditionType ...
type OperationStatusConditionType string

const (
	// OperationStatusConditionTypeEnabled indicates whether operation is enabled
	OperationStatusConditionTypeEnabled OperationStatusConditionType = "Enabled"
	// OperationStatusConditionTypeStarted indicates whether operation has been started
	OperationStatusConditionTypeStarted OperationStatusConditionType = "Started"
	// OperationStatusConditionTypeCompleted indicates whether operation has been completed
	OperationStatusConditionTypeCompleted OperationStatusConditionType = "Completed"
)

// GetOperationCondtion returns condition of specific type
func GetOperationCondtion(c []OperationStatusCondition, t OperationStatusConditionType) *OperationStatusCondition {
	for i := range c {
		if c[i].Type == t {
			return &c[i]
		}
	}
	return nil
}

// SetOperationCondtion set condition of specific type or appends if not present
func SetOperationCondtion(c []OperationStatusCondition, oc OperationStatusCondition) []OperationStatusCondition {
	conditon := GetOperationCondtion(c, oc.Type)

	if conditon != nil {
		if conditon.Status != oc.Status {
			conditon.LastTransitionTime = &metav1.Time{Time: time.Now()}
		}
		conditon.Status = oc.Status
		conditon.LastUpdateTime = metav1.Time{Time: time.Now()}
		conditon.Reason = oc.Reason
		conditon.Message = oc.Message
		return c
	}
	oc.LastUpdateTime = metav1.Time{Time: time.Now()}
	c = append(c, oc)
	return c
}
