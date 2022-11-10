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

package controllers

import (
	"context"
	"errors"

	wlv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileWebSphereLiberty) areSemeruResourcesReady(wlva *wlv1.WebSphereLibertyApplication) error {

	var replicas, readyReplicas, updatedReplicas int32
	namespacedName := types.NamespacedName{Name: wlva.GetName() + SemeruLabelNameSuffix, Namespace: wlva.GetNamespace()}

	// TODO - Pull this from WLA CR
	er := int32(1)
	expectedReplicas := &er

	// Check if deployment exists
	deployment := &appsv1.Deployment{}
	err := r.GetClient().Get(context.TODO(), namespacedName, deployment)
	if err != nil {
		// reason := "Semeru Cloud Compiler Deployment is not ready.", "NotCreated"
		// return c.SetConditionFields(msg, reason, corev1.ConditionFalse)
		return errors.New("semeru cloud compiler deployment is not ready")
	}
	// Get replicas
	ds := deployment.Status
	replicas, readyReplicas, updatedReplicas = ds.Replicas, ds.ReadyReplicas, ds.UpdatedReplicas

	// Check if all replicas are equal to the expected replicas
	if replicas == *expectedReplicas && readyReplicas == *expectedReplicas && updatedReplicas == *expectedReplicas {
		return nil // Semeru ready
	} else if replicas > *expectedReplicas {
		// reason = "ReplicaSetUpdating"
		// msg = "Replica set is progressing"
		// return c.SetConditionFields(msg, reason, corev1.ConditionFalse)
		return errors.New("semeru cloud compiler replica set is progressing")
	}
	return errors.New("semeru cloud compiler replica set is not ready")
}
