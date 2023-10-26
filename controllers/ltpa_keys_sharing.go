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
	"fmt"
	"io/ioutil"

	v1 "k8s.io/api/batch/v1"

	lutils "github.com/WASdev/websphere-liberty-operator/utils"

	wlv1 "github.com/WASdev/websphere-liberty-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Create the Deployment and Service objects for a Semeru Compiler used by a Websphere Liberty Application
func (r *ReconcileWebSphereLiberty) reconcileLTPAKeysSharing(instance *wlv1.WebSphereLibertyApplication, defaultMeta metav1.ObjectMeta) (error, string, string) {
	var ltpaSecretName string
	var err error
	if r.isLTPAKeySharingEnabled(instance) {
		err, ltpaSecretName = r.generateLTPAKeys(instance, defaultMeta)
		if err != nil {
			return err, "Failed to generate the shared LTPA Keys file", ltpaSecretName
		}
	} else {
		err := r.deleteLTPAKeysResources(instance, defaultMeta)
		if err != nil {
			return err, "Failed to delete LTPA Keys Resource", ltpaSecretName
		}
	}
	return nil, "", ltpaSecretName
}

// Returns true if the WebsphereLibertyApplication instance initiated the LTPA keys sharing process or sets the instance as the leader if the LTPA keys are not yet shared
func (r *ReconcileWebSphereLiberty) getOrSetLTPAKeysSharingLeader(instance *wlv1.WebSphereLibertyApplication) (error, string, bool, string) {
	ltpaServiceAccount := &corev1.ServiceAccount{}
	ltpaServiceAccount.Name = OperatorShortName + "-ltpa"
	ltpaServiceAccount.Namespace = instance.GetNamespace()
	ltpaServiceAccount.Labels = instance.GetLabels()
	err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaServiceAccount.Name, Namespace: ltpaServiceAccount.Namespace}, ltpaServiceAccount)
	if err != nil {
		if kerrors.IsNotFound(err) {
			r.CreateOrUpdate(ltpaServiceAccount, instance, func() error {
				return nil
			})
			return nil, instance.Name, true, ltpaServiceAccount.Name
		}
		return err, "", false, ltpaServiceAccount.Name
	}
	ltpaKeySharingLeaderName := ""
	for _, ownerReference := range ltpaServiceAccount.OwnerReferences {
		if ownerReference.Name == instance.Name {
			return nil, instance.Name, true, ltpaServiceAccount.Name
		}
		ltpaKeySharingLeaderName = ownerReference.Name
	}
	return nil, ltpaKeySharingLeaderName, false, ltpaServiceAccount.Name
}

// Generates the LTPA keys file and returns the name of the Secret storing its metadata
func (r *ReconcileWebSphereLiberty) generateLTPAKeys(instance *wlv1.WebSphereLibertyApplication, defaultMeta metav1.ObjectMeta) (error, string) {
	// Don't generate LTPA keys if this instance is not the leader
	err, ltpaKeySharingLeaderName, isLTPAKeySharingLeader, ltpaServiceAccountName := r.getOrSetLTPAKeysSharingLeader(instance)
	if err != nil {
		return err, ""
	}

	// Initialize LTPA resources
	ltpaXMLSecret := &corev1.Secret{}
	ltpaXMLSecret.Name = OperatorShortName + lutils.LTPAServerXMLSuffix
	ltpaXMLSecret.Namespace = instance.GetNamespace()
	ltpaXMLSecret.Labels = instance.GetLabels()

	generateLTPAKeysJob := &v1.Job{}
	generateLTPAKeysJob.Name = OperatorShortName + "-managed-ltpa-keys-generation"
	generateLTPAKeysJob.Namespace = instance.GetNamespace()
	generateLTPAKeysJob.Labels = instance.GetLabels()

	deletePropagationBackground := metav1.DeletePropagationBackground

	ltpaJobRequest := &corev1.ConfigMap{}
	ltpaJobRequest.Name = OperatorShortName + "-ltpa-job-request"
	ltpaJobRequest.Namespace = instance.GetNamespace()
	ltpaJobRequest.Labels = instance.GetLabels()

	ltpaSecret := &corev1.Secret{}
	ltpaSecret.Name = OperatorShortName + "-managed-ltpa"
	ltpaSecret.Namespace = instance.GetNamespace()
	ltpaSecret.Labels = instance.GetLabels()
	// If the LTPA Secret does not exist, run the Kubernetes Job to generate the shared ltpa.keys file and Secret
	err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaSecret.Name, Namespace: ltpaSecret.Namespace}, ltpaSecret)
	if err != nil && kerrors.IsNotFound(err) {
		// If this instance is not the leader, exit the reconcile loop
		if !isLTPAKeySharingLeader {
			return fmt.Errorf("Waiting for " + ltpaKeySharingLeaderName + " to generate the shared LTPA keys file for the " + instance.Namespace + " namespace."), ""
		}

		err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaJobRequest.Name, Namespace: ltpaJobRequest.Namespace}, ltpaJobRequest)
		if err != nil {
			// Create the Job Request if it doesn't exist
			if kerrors.IsNotFound(err) {
				// Clear all LTPA-related resources from a prior reconcile
				err = r.DeleteResource(ltpaXMLSecret)
				if err != nil {
					return err, ""
				}
				err = r.GetClient().Delete(context.TODO(), generateLTPAKeysJob, &client.DeleteOptions{PropagationPolicy: &deletePropagationBackground})
				if err != nil && !kerrors.IsNotFound(err) {
					return err, ""
				}
				err := r.CreateOrUpdate(ltpaJobRequest, instance, func() error {
					return nil
				})
				if err != nil {
					return fmt.Errorf("Failed to create ConfigMap " + ltpaJobRequest.Name), ""
				}
			} else {
				return fmt.Errorf("Failed to get ConfigMap " + ltpaJobRequest.Name), ""
			}
		} else {
			// Create the Role/RoleBinding
			ltpaRole := &rbacv1.Role{}
			ltpaRole.Name = OperatorShortName + "-managed-ltpa-role"
			ltpaRole.Namespace = instance.GetNamespace()
			ltpaRole.Rules = []rbacv1.PolicyRule{
				{
					Verbs:     []string{"create", "get"},
					APIGroups: []string{""},
					Resources: []string{"secrets"},
				},
			}
			ltpaRole.Labels = instance.GetLabels()
			r.CreateOrUpdate(ltpaRole, instance, func() error {
				return nil
			})

			ltpaRoleBinding := &rbacv1.RoleBinding{}
			ltpaRoleBinding.Name = OperatorShortName + "-managed-ltpa-rolebinding"
			ltpaRoleBinding.Namespace = instance.GetNamespace()
			ltpaRoleBinding.Subjects = []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      ltpaServiceAccountName,
					Namespace: instance.GetNamespace(),
				},
			}
			ltpaRoleBinding.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     ltpaRole.Name,
			}
			ltpaRoleBinding.Labels = instance.GetLabels()
			r.CreateOrUpdate(ltpaRoleBinding, instance, func() error {
				return nil
			})

			// Create a ConfigMap to store the controllers/assets/create_ltpa_keys.sh script
			ltpaKeysCreationScriptConfigMap := &corev1.ConfigMap{}
			ltpaKeysCreationScriptConfigMap.Name = OperatorShortName + "-managed-ltpa-script"
			ltpaKeysCreationScriptConfigMap.Namespace = instance.GetNamespace()
			ltpaKeysCreationScriptConfigMap.Labels = instance.GetLabels()
			err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaKeysCreationScriptConfigMap.Name, Namespace: ltpaKeysCreationScriptConfigMap.Namespace}, ltpaKeysCreationScriptConfigMap)
			if err == nil {
				r.DeleteResource(ltpaKeysCreationScriptConfigMap)
			}
			if err != nil && kerrors.IsNotFound(err) {
				ltpaKeysCreationScriptConfigMap.Data = make(map[string]string)
				script, err := ioutil.ReadFile("controllers/assets/create_ltpa_keys.sh")
				if err != nil {
					return err, ""
				}
				ltpaKeysCreationScriptConfigMap.Data["create_ltpa_keys.sh"] = string(script)
				r.CreateOrUpdate(ltpaKeysCreationScriptConfigMap, instance, func() error {
					return nil
				})
			}

			// Verify the controllers/assets/create_ltpa_keys.sh script has been loaded before starting the LTPA Job
			err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaKeysCreationScriptConfigMap.Name, Namespace: ltpaKeysCreationScriptConfigMap.Namespace}, ltpaKeysCreationScriptConfigMap)
			if err == nil {
				// Run the Kubernetes Job to generate the shared ltpa.keys file and LTPA Secret
				err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: generateLTPAKeysJob.Name, Namespace: generateLTPAKeysJob.Namespace}, generateLTPAKeysJob)
				if err != nil && kerrors.IsNotFound(err) {
					err = r.CreateOrUpdate(generateLTPAKeysJob, instance, func() error {
						lutils.CustomizeLTPAJob(generateLTPAKeysJob, instance, ltpaSecret.Name, ltpaServiceAccountName, ltpaKeysCreationScriptConfigMap.Name)
						return nil
					})
					if err != nil {
						return fmt.Errorf("Failed to create Job " + generateLTPAKeysJob.Name), ""
					}
				} else if err != nil {
					return fmt.Errorf("Failed to get Job " + generateLTPAKeysJob.Name), ""
				}
			}
		}

		// Reconcile the Job
		err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: generateLTPAKeysJob.Name, Namespace: generateLTPAKeysJob.Namespace}, generateLTPAKeysJob)
		if err != nil && kerrors.IsNotFound(err) {
			return fmt.Errorf("Waiting for Job " + generateLTPAKeysJob.Name + " to be created."), ""
		} else if err != nil {
			return fmt.Errorf("Failed to get Job " + generateLTPAKeysJob.Name), ""
		}
		if len(generateLTPAKeysJob.Status.Conditions) > 0 && generateLTPAKeysJob.Status.Conditions[0].Type == v1.JobFailed {
			return fmt.Errorf("Job " + generateLTPAKeysJob.Name + " has failed. Manually clean up hung resources by setting .spec.manageLTPA to false in the " + ltpaKeySharingLeaderName + " instance."), ""
		}
		return fmt.Errorf("Waiting for Job " + generateLTPAKeysJob.Name + " to be completed."), ""
	} else if err != nil {
		return err, ""
	} else {
		if !isLTPAKeySharingLeader {
			return nil, ltpaSecret.Name
		}
	}

	// The LTPA Secret is created (in other words, the LTPA Job has completed), so delete the Job request
	err = r.DeleteResource(ltpaJobRequest)
	if err != nil {
		return err, ltpaSecret.Name
	}

	// Create the Liberty Server XML Secret if it doesn't exist
	serverXMLSecretErr := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaXMLSecret.Name, Namespace: ltpaXMLSecret.Namespace}, ltpaXMLSecret)
	if serverXMLSecretErr != nil && kerrors.IsNotFound(serverXMLSecretErr) {
		r.CreateOrUpdate(ltpaXMLSecret, nil, func() error {
			lutils.CustomizeLTPAServerXML(ltpaXMLSecret, instance, string(ltpaSecret.Data["password"]))
			return nil
		})
	}
	return nil, ltpaSecret.Name
}

func (r *ReconcileWebSphereLiberty) isLTPAKeySharingEnabled(instance *wlv1.WebSphereLibertyApplication) bool {
	if instance.GetManageLTPA() != nil && *instance.GetManageLTPA() {
		return true
	}
	return false
}

// Deletes resources used to create the LTPA keys file
func (r *ReconcileWebSphereLiberty) deleteLTPAKeysResources(instance *wlv1.WebSphereLibertyApplication, defaultMeta metav1.ObjectMeta) error {
	// Don't delete LTPA keys resources if this instance is not the leader
	err, _, isLTPAKeySharingLeader, ltpaServiceAccountName := r.getOrSetLTPAKeysSharingLeader(instance)
	if err != nil {
		return err
	}
	if !isLTPAKeySharingLeader {
		return nil
	}

	generateLTPAKeysJob := &v1.Job{}
	generateLTPAKeysJob.Name = OperatorShortName + "-managed-ltpa-keys-generation"
	generateLTPAKeysJob.Namespace = instance.GetNamespace()
	deletePropagationBackground := metav1.DeletePropagationBackground
	err = r.GetClient().Delete(context.TODO(), generateLTPAKeysJob, &client.DeleteOptions{PropagationPolicy: &deletePropagationBackground})
	if err != nil && !kerrors.IsNotFound(err) {
		return err
	}

	ltpaKeysCreationScriptConfigMap := &corev1.ConfigMap{}
	ltpaKeysCreationScriptConfigMap.Name = OperatorShortName + "-managed-ltpa-script"
	ltpaKeysCreationScriptConfigMap.Namespace = instance.GetNamespace()
	err = r.DeleteResource(ltpaKeysCreationScriptConfigMap)
	if err != nil {
		return err
	}

	ltpaRoleBinding := &rbacv1.RoleBinding{}
	ltpaRoleBinding.Name = OperatorShortName + "-managed-ltpa-rolebinding"
	ltpaRoleBinding.Namespace = instance.GetNamespace()
	err = r.DeleteResource(ltpaRoleBinding)
	if err != nil {
		return err
	}

	ltpaRole := &rbacv1.Role{}
	ltpaRole.Name = OperatorShortName + "-managed-ltpa-role"
	ltpaRole.Namespace = instance.GetNamespace()
	err = r.DeleteResource(ltpaRole)
	if err != nil {
		return err
	}

	ltpaServiceAccount := &corev1.ServiceAccount{}
	ltpaServiceAccount.Name = ltpaServiceAccountName
	ltpaServiceAccount.Namespace = instance.GetNamespace()
	err = r.DeleteResource(ltpaServiceAccount)
	if err != nil {
		return err
	}

	return nil
}
