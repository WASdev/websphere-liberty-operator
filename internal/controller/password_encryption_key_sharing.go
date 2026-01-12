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

package controller

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"
	"sync"
	"time"

	tree "github.com/OpenLiberty/open-liberty-operator/utils/tree"
	wlv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	lutils "github.com/WASdev/websphere-liberty-operator/utils"
	"github.com/application-stacks/runtime-component-operator/common"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

const PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME = "password-encryption"

func init() {
	lutils.LeaderTrackerMutexes.Store(PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME, &sync.Mutex{})
}

func (r *ReconcileWebSphereLiberty) reconcilePasswordEncryptionKey(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (string, string, string, error) {
	if r.isPasswordEncryptionKeySharingEnabled(instance) {
		leaderName, thisInstanceIsLeader, _, err := r.reconcileLeader(instance, passwordEncryptionMetadata, PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME, true)
		if err != nil && !kerrors.IsNotFound(err) {
			return "", "", "", err
		}
		if thisInstanceIsLeader {
			// Is there a password encryption key to duplicate for internal use?
			if err := r.mirrorEncryptionKeySecretState(recCtx, instance, passwordEncryptionMetadata); err != nil {
				return "Failed to process the password encryption key Secret", "", "", err
			}
		}

		// Does the namespace already have a password encryption key sharing Secret?
		encryptionSecret, _, err := r.hasInternalEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
		if err == nil {
			// Is the password encryption key field in the Secret valid?
			if encryptionKey := encryptionSecret.Data["passwordEncryptionKey"]; len(encryptionKey) > 0 {
				// non-leaders should still be able to pass this process to return the encryption secret name
				if thisInstanceIsLeader {
					// Create the Liberty config that will mount into the pods
					err := r.createPasswordEncryptionKeyLibertyConfig(recCtx, instance, passwordEncryptionMetadata, encryptionKey)
					if err != nil {
						return "Failed to create Liberty resources to share the password encryption key", "", "", err
					}
				} else {
					// non-leaders should yield for the password encryption leader to mirror the encryption key's state
					if !r.encryptionKeySecretMirrored(recCtx, instance, passwordEncryptionMetadata) {
						return "", "", "", fmt.Errorf("Waiting for WebSphereLibertyApplication instance '%s' to mirror the shared Password Encryption Key Secret for the namespace '%s'.", leaderName, instance.Namespace)
					}
				}
				return "", encryptionSecret.Name, string(encryptionSecret.Data["lastRotation"]), nil
			}
		} else if !kerrors.IsNotFound(err) {
			return "Failed to get the password encryption key Secret", "", "", err
		}
	} else {
		err := r.RemoveLeaderTrackerReference(instance, PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME)
		if err != nil {
			return "Failed to remove leader tracking reference to the password encryption key", "", "", err
		}
	}
	return "", "", "", nil
}

func (r *ReconcileWebSphereLiberty) reconcilePasswordEncryptionMetadata(treeMap map[string]interface{}, latestOperandVersion string) (lutils.LeaderTrackerMetadataList, error) {
	metadataList := &lutils.PasswordEncryptionMetadataList{}
	metadataList.Items = []lutils.LeaderTrackerMetadata{}

	pathOptionsList, pathChoicesList := r.getPasswordEncryptionPathOptionsAndChoices(latestOperandVersion)
	for i := range pathOptionsList {
		metadata := &lutils.PasswordEncryptionMetadata{}
		pathOptions := pathOptionsList[i]
		pathChoices := pathChoicesList[i]

		// convert the path options and choices into a labelString, for a path of length n, the labelString is
		// constructed as a weaved array in format "<pathOptions[0]>.<pathChoices[0]>.<pathOptions[1]>.<pathChoices[1]>...<pathOptions[n-1]>.<pathChoices[n-1]>"
		labelString, err := tree.GetLabelFromDecisionPath(latestOperandVersion, pathOptions, pathChoices)
		if err != nil {
			return metadataList, err
		}
		// validate that the decision path such as "v1_4_0.managePasswordEncryption.true" is a valid subpath in treeMap
		// an error here indicates a build time error created by the operator developer or pollution of the password-encryption-decision-tree.yaml
		// NOTE: validSubPath is a substring of labelString and a valid path within treeMap; it will always hold that len(validSubPath) <= len(labelString)
		validSubPath, err := tree.CanTraverseTree(treeMap, labelString, true)
		if err != nil {
			return metadataList, err
		}
		// NOTE: Checking the leaderTracker can be skipped assuming there is only one password encryption key per namespace
		// Leader tracker reconcile is only required to prevent overriding other shared resources (i.e. password encryption keys) in the same namespace
		// Uncomment code below to extend to multiple password encryption keys per namespace. See ltpa_keys_sharing.go for an example.

		// // retrieve the password encryption leader tracker to re-use an existing name or to create a new metadata.Name
		// leaderTracker, _, err := lutils.GetLeaderTracker(instance, OperatorShortName, PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME, r.GetClient())
		// if err != nil {
		// 	return metadataList, err
		// }
		// // if the leaderTracker is on a mismatched version, wait for a subsequent reconcile loop to re-create the leader tracker
		// if leaderTracker.Labels[lutils.LeaderVersionLabel] != latestOperandVersion {
		// 	return metadataList, fmt.Errorf("waiting for the Leader Tracker to be updated")
		// }

		// to avoid limitation with Kubernetes label values having a max length of 63, translate validSubPath into a path index
		pathIndex := tree.GetLeafIndex(treeMap, validSubPath)
		versionedPathIndex := fmt.Sprintf("%s.%d", latestOperandVersion, pathIndex)

		metadata.Path = validSubPath
		metadata.PathIndex = versionedPathIndex
		metadata.Name = r.getPasswordEncryptionMetadataName() // You could augment this function to extend to multiple password encryption keys per namespace. See ltpa_keys_sharing.go for an example.
		metadataList.Items = append(metadataList.Items, metadata)
	}
	return metadataList, nil
}

func (r *ReconcileWebSphereLiberty) getPasswordEncryptionPathOptionsAndChoices(latestOperandVersion string) ([][]string, [][]string) {
	var pathOptionsList, pathChoicesList [][]string
	if latestOperandVersion == "v1_4_0" {
		pathOptions := []string{"managePasswordEncryption"}
		pathChoices := []string{"true"} // there is only one possible password encryption key per namespace which corresponds to one path only
		pathOptionsList = append(pathOptionsList, pathOptions)
		pathChoicesList = append(pathChoicesList, pathChoices)
	}
	return pathOptionsList, pathChoicesList
}

func (r *ReconcileWebSphereLiberty) getPasswordEncryptionMetadataName() string {
	// NOTE: there is only one possible password encryption key per namespace which corresponds to one shared resource name from password-encryption-signature.yaml
	// If you would like to have more than one password encryption key in a single namespace, use ltpa-signature.yaml as a template
	//
	// _, sharedResourceName, err := lutils.CreateUnstructuredResourceFromSignature(PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME, OperatorShortName, "")
	// if err != nil {
	// 	return "", err
	// }
	// return sharedResourceName, nil
	return "" // there is only one password encryption key per namespace which is represented by the empty string suffix
}

func (r *ReconcileWebSphereLiberty) isPasswordEncryptionKeySharingEnabled(instance *wlv1.WebSphereLibertyApplication) bool {
	return instance.GetManagePasswordEncryption() != nil && *instance.GetManagePasswordEncryption()
}

func (r *ReconcileWebSphereLiberty) isUsingPasswordEncryptionKeySharing(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) bool {
	if r.isPasswordEncryptionKeySharingEnabled(instance) {
		_, err := r.hasUserEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
		return err == nil
	}
	return false
}

func (r *ReconcileWebSphereLiberty) getInternalPasswordEncryptionKeyState(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) ([]byte, []byte, bool, error) {
	if !r.isPasswordEncryptionKeySharingEnabled(instance) {
		return []byte{}, []byte{}, false, nil
	}
	internalEncryptionKey, _, err := r.hasInternalEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
	if err != nil {
		return []byte{}, []byte{}, true, err
	}
	passwordEncryptionKey, passwordEncryptionKeyFound := internalEncryptionKey.Data["passwordEncryptionKey"]
	if !passwordEncryptionKeyFound {
		return []byte{}, []byte{}, true, fmt.Errorf("The internal password encryption key is missing field 'passwordEncryptionKey'")
	}
	encryptionSecretLastRotation, encryptionSecretLastRotationFound := internalEncryptionKey.Data["lastRotation"]
	if !encryptionSecretLastRotationFound {
		return []byte{}, []byte{}, true, fmt.Errorf("The internal password encryption key is missing field 'lastRotation'")
	}
	return passwordEncryptionKey, encryptionSecretLastRotation, true, nil
}

// Returns the Secret that contains the password encryption key used internally by the operator
func (r *ReconcileWebSphereLiberty) hasInternalEncryptionKeySecret(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (*corev1.Secret, *sync.WaitGroup, error) {
	return r.getWaitableSecret(recCtx, instance, lutils.LocalPasswordEncryptionKeyRootName+passwordEncryptionMetadata.Name+"-internal")
}

// Returns the Secret that contains the password encryption key provided by the user
func (r *ReconcileWebSphereLiberty) hasUserEncryptionKeySecret(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (*corev1.Secret, error) {
	return r.getSecret(recCtx, instance, lutils.PasswordEncryptionKeyRootName+passwordEncryptionMetadata.Name)
}

func (r *ReconcileWebSphereLiberty) encryptionKeySecretMirrored(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) bool {
	userEncryptionSecret, err := r.hasUserEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
	if err != nil {
		return false
	}
	internalEncryptionSecret, _, err := r.hasInternalEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
	if err != nil {
		return false
	}
	internalPasswordEncryptionKey := internalEncryptionSecret.Data["passwordEncryptionKey"]
	userPasswordEncryptionKey := userEncryptionSecret.Data["passwordEncryptionKey"]
	return len(userPasswordEncryptionKey) > 0 && subtle.ConstantTimeCompare(internalPasswordEncryptionKey, userPasswordEncryptionKey) == 1
}

func (r *ReconcileWebSphereLiberty) mirrorEncryptionKeySecretState(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) error {
	userEncryptionSecret, userEncryptionSecretErr := r.hasUserEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
	// Error if there was an issue getting the userEncryptionSecret
	if userEncryptionSecretErr != nil && !kerrors.IsNotFound(userEncryptionSecretErr) {
		return userEncryptionSecretErr
	}
	internalEncryptionSecret, internalEncryptionSecretWaitGroup, internalEncryptionSecretErr := r.hasInternalEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
	// Error if there was an issue getting the internalEncryptionSecret
	if internalEncryptionSecretErr != nil && !kerrors.IsNotFound(internalEncryptionSecretErr) {
		return internalEncryptionSecretErr
	}
	// Case 0: no user encryption secret, no internal encryption secret: secrets already mirrored
	// Case 1: no user encryption secret, internal encryption secret exists: so delete internalEncryptionSecret
	if kerrors.IsNotFound(userEncryptionSecretErr) {
		if kerrors.IsNotFound(internalEncryptionSecretErr) {
			return nil
		} else {
			if err := r.DeleteResource(internalEncryptionSecret); err != nil {
				return err
			}
		}
	}

	// Case 2: user encryption secret exists, no internal secret: Create internalEncryptionSecret
	// Case 3: user encryption secret exists, internal secret exists: Update internalEncryptionSecret
	return r.TrackedCreateOrUpdate(internalEncryptionSecret, nil, func() error {
		if internalEncryptionSecret.Data == nil {
			internalEncryptionSecret.Data = make(map[string][]byte)
		}
		if userEncryptionSecret.Data == nil {
			userEncryptionSecret.Data = make(map[string][]byte)
		}
		internalPasswordEncryptionKey := internalEncryptionSecret.Data["passwordEncryptionKey"]
		userPasswordEncryptionKey := userEncryptionSecret.Data["passwordEncryptionKey"]
		if subtle.ConstantTimeCompare(internalPasswordEncryptionKey, userPasswordEncryptionKey) != 1 {
			if len(internalPasswordEncryptionKey) > 0 {
				clear(internalEncryptionSecret.Data["passwordEncryptionKey"])
			}
			if len(internalEncryptionSecret.Data["lastRotation"]) > 0 {
				clear(internalEncryptionSecret.Data["lastRotation"])
			}
			passwordEncryptionKey := make([]byte, len(userEncryptionSecret.Data["passwordEncryptionKey"]))
			copy(passwordEncryptionKey, userEncryptionSecret.Data["passwordEncryptionKey"])
			internalEncryptionSecret.Data["passwordEncryptionKey"] = passwordEncryptionKey
			internalEncryptionSecret.Data["lastRotation"] = []byte(fmt.Sprint(time.Now().Unix()))
		}
		return nil
	}, internalEncryptionSecretWaitGroup)
}

// Deletes the mirrored encryption key secret if the initial encryption key secret no longer exists
func (r *ReconcileWebSphereLiberty) deleteMirroredEncryptionKeySecret(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) error {
	_, userEncryptionSecretErr := r.hasUserEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
	// Error if there was an issue getting the userEncryptionSecret
	if userEncryptionSecretErr != nil && !kerrors.IsNotFound(userEncryptionSecretErr) {
		return userEncryptionSecretErr
	}
	internalEncryptionSecret, _, internalEncryptionSecretErr := r.hasInternalEncryptionKeySecret(recCtx, instance, passwordEncryptionMetadata)
	// Error if there was an issue getting the internalEncryptionSecret
	if internalEncryptionSecretErr != nil && !kerrors.IsNotFound(internalEncryptionSecretErr) {
		return internalEncryptionSecretErr
	}
	// Case 1: no user encryption secret, internal encryption secret exists: so delete internalEncryptionSecret
	if kerrors.IsNotFound(userEncryptionSecretErr) && !kerrors.IsNotFound(internalEncryptionSecretErr) {
		if err := r.DeleteResource(internalEncryptionSecret); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileWebSphereLiberty) getWaitableSecret(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, secretName string) (*corev1.Secret, *sync.WaitGroup, error) {
	secret, wg := common.NewWaitableSecret(recCtx, secretName, instance.GetNamespace())
	secret.Labels = lutils.GetRequiredLabels(secret.Name, "")
	err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret)
	return secret, wg, err
}

func (r *ReconcileWebSphereLiberty) getSecret(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, secretName string) (*corev1.Secret, error) {
	secret := common.NewSecret(recCtx, secretName, instance.GetNamespace())
	secret.Labels = lutils.GetRequiredLabels(secret.Name, "")
	err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret)
	return secret, err
}

// Creates the Liberty XML to mount the password encryption keys Secret into the application pods
func (r *ReconcileWebSphereLiberty) createPasswordEncryptionKeyLibertyConfig(recCtx context.Context, instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata, encryptionKey []byte) error {
	if len(encryptionKey) == 0 {
		return fmt.Errorf("a password encryption key was not specified")
	}

	// The Secret to hold the server.xml that will override the password encryption key for the Liberty server
	// This server.xml will be mounted in /output/liberty-operator/encryptionKey.xml
	encryptionKeyXMLName := OperatorShortName + lutils.ManagedEncryptionServerXML + passwordEncryptionMetadata.Name
	encryptionKeyXML, encryptionKeyXMLWaitGroup := common.NewWaitableSecret(recCtx, encryptionKeyXMLName, instance.GetNamespace())
	encryptionKeyXML.Labels = lutils.GetRequiredLabels(encryptionKeyXMLName, "")

	if err := r.TrackedCreateOrUpdate(encryptionKeyXML, nil, func() error {
		return lutils.CustomizeEncryptionKeyXML(encryptionKeyXML, encryptionKey)
	}, encryptionKeyXMLWaitGroup); err != nil {
		return err
	}

	// The Secret to hold the server.xml that will import the password encryption key into the Liberty server
	// This server.xml will be mounted in /config/configDropins/overrides/encryptionKeyMount.xml
	mountingXMLSecretName := OperatorShortName + lutils.ManagedEncryptionMountServerXML + passwordEncryptionMetadata.Name
	mountingXML, mountingXMLWaitGroup := common.NewWaitableSecret(recCtx, mountingXMLSecretName, instance.GetNamespace())
	mountingXML.Labels = lutils.GetRequiredLabels(mountingXMLSecretName, "")

	if err := r.TrackedCreateOrUpdate(mountingXML, nil, func() error {
		mountDir := strings.Replace(lutils.SecureMountPath+"/"+lutils.EncryptionKeyXMLFileName, "/output", "${server.output.dir}", 1)
		return lutils.CustomizeLibertyFileMountXML(mountingXML, lutils.EncryptionKeyMountXMLFileName, mountDir)
	}, mountingXMLWaitGroup); err != nil {
		return err
	}

	return nil
}

// Tracks existing password encryption resources by populating a LeaderTracker array used to initialize the LeaderTracker
func (r *ReconcileWebSphereLiberty) GetPasswordEncryptionResources(instance *wlv1.WebSphereLibertyApplication, treeMap map[string]interface{}, replaceMap map[string]map[string]string, latestOperandVersion string, assetsFolder *string) (*unstructured.UnstructuredList, string, error) {
	passwordEncryptionResources, _, err := lutils.CreateUnstructuredResourceListFromSignature(PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME, assetsFolder, "") // TODO: replace prefix "" to specify operator precedence such as with prefix "wlo-"
	if err != nil {
		return nil, "", err
	}
	passwordEncryptionResource, passwordEncryptionResourceName, err := lutils.CreateUnstructuredResourceFromSignature(PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME, assetsFolder, "", "") // TODO: replace prefix "" to specify operator precedence such as with prefix "wlo-"
	if err != nil {
		return nil, "", err
	}
	if err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: passwordEncryptionResourceName, Namespace: instance.GetNamespace()}, passwordEncryptionResource); err == nil {
		passwordEncryptionResources.Items = append(passwordEncryptionResources.Items, *passwordEncryptionResource)
	} else if !kerrors.IsNotFound(err) {
		return nil, "", err
	}
	return passwordEncryptionResources, passwordEncryptionResourceName, nil
}
