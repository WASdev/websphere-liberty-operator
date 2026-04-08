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
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

const PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME = "password-encryption"
const PasswordEncryptionKey = "passwordEncryptionKey"

const AES_ENCRYPTION_RESOURCE_SHARING_FILE_NAME = "aes-encryption"
const AESEncryptionKey = "aesEncryptionKey"

func init() {
	lutils.LeaderTrackerMutexes.Store(PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME, &sync.Mutex{})
}
func (r *ReconcileWebSphereLiberty) reconcileEncryptionKey(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (string, string, string, error) {
	if r.isPasswordEncryptionKeySharingEnabled(instance) {
		leaderName, thisInstanceIsLeader, _, err := r.reconcileLeader(instance, passwordEncryptionMetadata, PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME, true)
		if err != nil && !kerrors.IsNotFound(err) {
			return "", "", "", err
		}
		if thisInstanceIsLeader {
			// Is there a password encryption key to duplicate for internal use?
			if err := r.mirrorEncryptionKeySecretState(instance, passwordEncryptionMetadata, r.hasUserEncryptionKeySecret, r.hasInternalEncryptionKeySecret, PasswordEncryptionKey); err != nil {
				// Mirror the aes encryption key if exists
				if aesErr := r.mirrorEncryptionKeySecretState(instance, passwordEncryptionMetadata, r.hasUserAESEncryptionKeySecret, r.hasInternalAESEncryptionKeySecret, AESEncryptionKey); aesErr != nil {
					// This is when both the password encrpytion key and aes encryption key could not be found
					return "Failed to process the password encryption key Secret", "", "", err
				}
				// Here the aes encryption key was found and we can passthrough to reconciling the keys
				return r.reconcileAESEncryptionKey(instance, passwordEncryptionMetadata, thisInstanceIsLeader, leaderName)
			} else {
				// Here the password encryption key was found, but we still want to mirror the aes encryption key (allowing failures)
				// before reconciling the password encryption key
				if err := r.mirrorEncryptionKeySecretState(instance, passwordEncryptionMetadata, r.hasUserAESEncryptionKeySecret, r.hasInternalAESEncryptionKeySecret, AESEncryptionKey); err == nil {
					return r.reconcileAESEncryptionKey(instance, passwordEncryptionMetadata, thisInstanceIsLeader, leaderName)
				}
				return r.reconcilePasswordEncryptionKey(instance, passwordEncryptionMetadata, thisInstanceIsLeader, leaderName)
			}
		}
		// Give internal AES key the higher precedence (allowing failures)
		aesEncryptionErrMessage, aesEncryptionSecretName, aesEncryptionLastRotation, err := r.reconcileAESEncryptionKey(instance, passwordEncryptionMetadata, thisInstanceIsLeader, leaderName)
		if err == nil {
			// no error so return the aes encryption key
			return aesEncryptionErrMessage, aesEncryptionSecretName, aesEncryptionLastRotation, err
		}
		// Otherwise fallback to the password encryption key
		return r.reconcilePasswordEncryptionKey(instance, passwordEncryptionMetadata, thisInstanceIsLeader, leaderName)
	} else {
		err := r.RemoveLeaderTrackerReference(instance, PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME)
		if err != nil {
			return "Failed to remove leader tracking reference to the encryption key", "", "", err
		}
	}
	return "", "", "", nil
}

func (r *ReconcileWebSphereLiberty) reconcileAESEncryptionKey(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata, thisInstanceIsLeader bool, leaderName string) (string, string, string, error) {
	// Does the namespace already have a password encryption key sharing Secret?
	encryptionSecret, err := r.hasInternalAESEncryptionKeySecret(instance, passwordEncryptionMetadata)
	defer encryptionSecret.Destroy()
	if err == nil {
		// Return err if the password encryption key does not exist
		if _, found := encryptionSecret.LockedData.Get(AESEncryptionKey); !found {
			return "Failed to get the password encryption key Secret because " + AESEncryptionKey + " key is missing", "", "", err
		}
		// Is the password encryption key field in the Secret valid?
		if encryptionKey, _ := encryptionSecret.LockedData.Get(AESEncryptionKey); subtle.ConstantTimeCompare(encryptionKey, []byte{}) != 1 {
			// non-leaders should still be able to pass this process to return the encryption secret name
			if thisInstanceIsLeader {
				// Create the Liberty config that will mount into the pods
				err := r.createAESEncryptionKeyLibertyConfig(instance, passwordEncryptionMetadata, encryptionKey)
				if err != nil {
					return "Failed to create Liberty resources to share the AES encryption key", "", "", err
				}
			} else {
				// non-leaders should yield for the password encryption leader to mirror the encryption key's state
				if !r.isSecretMirrored(instance, passwordEncryptionMetadata, r.hasUserAESEncryptionKeySecret, r.hasInternalAESEncryptionKeySecret, AESEncryptionKey) {
					return "", "", "", fmt.Errorf("Waiting for WebSphereLibertyApplication instance '%s' to mirror the shared Password Encryption Key (aes) Secret for the namespace '%s'.", leaderName, instance.Namespace)
				}
			}
			lastRotation, _ := encryptionSecret.LockedData.Get("lastRotation")
			return "", encryptionSecret.Name, string(lastRotation), nil
		}
	}
	return "Failed to get the AES encryption key Secret", "", "", err
}

func (r *ReconcileWebSphereLiberty) reconcilePasswordEncryptionKey(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata, thisInstanceIsLeader bool, leaderName string) (string, string, string, error) {
	// Does the namespace already have a password encryption key sharing Secret?
	encryptionSecret, err := r.hasInternalEncryptionKeySecret(instance, passwordEncryptionMetadata)
	defer encryptionSecret.Destroy()
	if err == nil {
		// Return err if the password encryption key does not exist
		if _, found := encryptionSecret.LockedData[PasswordEncryptionKey]; !found {
			return "Failed to get the password encryption key Secret because " + PasswordEncryptionKey + " key is missing", "", "", err
		}
		// Is the password encryption key field in the Secret valid?
		if encryptionKey, _ := encryptionSecret.LockedData.Get(PasswordEncryptionKey); subtle.ConstantTimeCompare(encryptionKey, []byte{}) != 1 {
			// non-leaders should still be able to pass this process to return the encryption secret name
			if thisInstanceIsLeader {
				// Create the Liberty config that will mount into the pods
				err := r.createPasswordEncryptionKeyLibertyConfig(instance, passwordEncryptionMetadata, encryptionKey)
				if err != nil {
					return "Failed to create Liberty resources to share the password encryption key", "", "", err
				}
			} else {
				// non-leaders should yield for the password encryption leader to mirror the encryption key's state
				if !r.isSecretMirrored(instance, passwordEncryptionMetadata, r.hasUserEncryptionKeySecret, r.hasInternalEncryptionKeySecret, PasswordEncryptionKey) {
					return "", "", "", fmt.Errorf("Waiting for WebSphereLibertyApplication instance '%s' to mirror the shared Password Encryption Key (password) Secret for the namespace '%s'.", leaderName, instance.Namespace)
				}
			}
			lastRotation, _ := encryptionSecret.LockedData.Get("lastRotation")
			return "", encryptionSecret.Name, string(lastRotation), nil
		}
	}
	return "Failed to get the password encryption key Secret", "", "", err
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

func (r *ReconcileWebSphereLiberty) isUsingPasswordEncryptionKeySharing(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) bool {
	if r.isPasswordEncryptionKeySharingEnabled(instance) {
		_, passwordErr := r.hasUserEncryptionKeySecret(instance, passwordEncryptionMetadata)
		_, aesErr := r.hasUserAESEncryptionKeySecret(instance, passwordEncryptionMetadata)
		return passwordErr == nil || aesErr == nil
	}
	return false
}

func (r *ReconcileWebSphereLiberty) isUsingAESPasswordEncryptionKeySharing(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) bool {
	if r.isPasswordEncryptionKeySharingEnabled(instance) {
		_, aesErr := r.hasUserAESEncryptionKeySecret(instance, passwordEncryptionMetadata)
		return aesErr == nil
	}
	return false
}

func (r *ReconcileWebSphereLiberty) isUsingPlainPasswordEncryptionKeySharing(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) bool {
	if r.isPasswordEncryptionKeySharingEnabled(instance) {
		_, passwordErr := r.hasUserEncryptionKeySecret(instance, passwordEncryptionMetadata)
		return passwordErr == nil
	}
	return false
}

func (r *ReconcileWebSphereLiberty) getEncryptionKeyData(encryptionSecret *common.LockedBufferSecret, matchedKey string) ([]byte, []byte, bool) {
	encryptionKey := []byte{}
	encryptionKeyLastRotation := []byte{}
	if key, found := encryptionSecret.LockedData.Get(matchedKey); found {
		encryptionKey = key
	}
	if lastRotation, found := encryptionSecret.LockedData.Get("lastRotation"); found {
		encryptionKeyLastRotation = lastRotation
	}
	if subtle.ConstantTimeCompare(encryptionKey, []byte{}) == 1 || subtle.ConstantTimeCompare(encryptionKeyLastRotation, []byte{}) == 1 {
		// don't need to delete this misconfigured Secret because mirrorEncryptionKeySecretState will create/update it later
		return []byte{}, []byte{}, false
	}
	return encryptionKey, encryptionKeyLastRotation, true
}

func (r *ReconcileWebSphereLiberty) getValidInternalEncryptionKey(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, bool, bool, error) {
	sharingEnabled := r.isPasswordEncryptionKeySharingEnabled(instance)
	if !sharingEnabled {
		return nil, sharingEnabled, false, nil
	}

	aesSecret, err := r.hasInternalAESEncryptionKeySecret(instance, passwordEncryptionMetadata)
	aesFound := !kerrors.IsNotFound(err)
	if aesFound && err != nil {
		aesSecret.Destroy()
		return nil, sharingEnabled, aesFound, err
	}
	passwordSecret, err := r.hasInternalEncryptionKeySecret(instance, passwordEncryptionMetadata)
	passwordFound := !kerrors.IsNotFound(err)
	if passwordFound && err != nil {
		aesSecret.Destroy()
		passwordSecret.Destroy()
		return nil, sharingEnabled, aesFound, err
	}

	_, _, aesValid := r.getEncryptionKeyData(aesSecret, AESEncryptionKey)
	_, _, passwordValid := r.getEncryptionKeyData(passwordSecret, PasswordEncryptionKey)

	aesFoundAndValid := aesFound && aesValid
	passwordFoundAndValid := passwordFound && passwordValid

	if aesFoundAndValid {
		passwordSecret.Destroy()
		return aesSecret, sharingEnabled, aesFound, nil
	} else if passwordFoundAndValid {
		aesSecret.Destroy()
		return passwordSecret, sharingEnabled, aesFound, nil
	}

	// if aes/password were found but not valid then return a warning
	if aesFound {
		aesSecret.Destroy()
		passwordSecret.Destroy()
		return nil, sharingEnabled, aesFound, fmt.Errorf("the wlp-aes-encryption-key Secret was found but contained an invalid field")
	} else if passwordFound {
		aesSecret.Destroy()
		passwordSecret.Destroy()
		return nil, sharingEnabled, aesFound, fmt.Errorf("the wlp-password-encryption-key Secret was found but contained an invalid field")
	}

	aesSecret.Destroy()
	passwordSecret.Destroy()
	return nil, sharingEnabled, aesFound, nil
}

func (r *ReconcileWebSphereLiberty) getInternalEncryptionKeyState(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, bool, bool, error) {
	encryptionSecret, sharingEnabled, usingAES, err := r.getValidInternalEncryptionKey(instance, passwordEncryptionMetadata)

	if !sharingEnabled {
		if encryptionSecret != nil {
			encryptionSecret.Destroy()
		}
		return nil, sharingEnabled, false, nil
	}

	if err != nil {
		if encryptionSecret != nil {
			encryptionSecret.Destroy()
		}
		return nil, sharingEnabled, false, err
	}

	if encryptionSecret == nil {
		return nil, sharingEnabled, false, fmt.Errorf("a password encryption key Secret was either not found or misconfigured")
	}

	matchedKey := ""
	if usingAES {
		matchedKey = AESEncryptionKey
	} else {
		matchedKey = PasswordEncryptionKey
	}

	_, _, valid := r.getEncryptionKeyData(encryptionSecret, matchedKey)

	if !valid {
		encryptionSecret.Destroy()
		return nil, sharingEnabled, false, fmt.Errorf("a password encryption key Secret was either not found or misconfigured")
	}

	// Return the secret - caller is responsible for destroying it
	return encryptionSecret, sharingEnabled, usingAES, err
}

func (r *ReconcileWebSphereLiberty) getEncryptionKeyNameFromRef(secretNameRef string, passwordEncryptionMetadataName string) (string, string) {
	keySecretName := ""
	dataFieldName := ""
	if strings.HasPrefix(secretNameRef, lutils.LocalAESEncryptionKeyRootName) {
		keySecretName = lutils.AESEncryptionKeyRootName
		dataFieldName = AESEncryptionKey
	} else {
		keySecretName = lutils.PasswordEncryptionKeyRootName
		dataFieldName = PasswordEncryptionKey
	}
	keySecretName += passwordEncryptionMetadataName
	return keySecretName, dataFieldName
}

// Returns the Secret that contains the aes encryption key used internally by the operator
func (r *ReconcileWebSphereLiberty) hasInternalAESEncryptionKeySecret(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, error) {
	metaName := ""
	if passwordEncryptionMetadata != nil {
		metaName = passwordEncryptionMetadata.Name
	}
	return common.GetSecret(r.GetClient(), lutils.LocalAESEncryptionKeyRootName+metaName+"-internal", instance.GetNamespace())
}

// Returns the Secret that contains the aes encryption key provided by the user
func (r *ReconcileWebSphereLiberty) hasUserAESEncryptionKeySecret(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, error) {
	metaName := ""
	if passwordEncryptionMetadata != nil {
		metaName = passwordEncryptionMetadata.Name
	}
	return common.GetSecret(r.GetClient(), lutils.AESEncryptionKeyRootName+metaName, instance.GetNamespace())
}

// Returns the Secret that contains the password encryption key used internally by the operator
func (r *ReconcileWebSphereLiberty) hasInternalEncryptionKeySecret(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, error) {
	metaName := ""
	if passwordEncryptionMetadata != nil {
		metaName = passwordEncryptionMetadata.Name
	}
	return common.GetSecret(r.GetClient(), lutils.LocalPasswordEncryptionKeyRootName+metaName+"-internal", instance.GetNamespace())
}

// Returns the Secret that contains the password encryption key provided by the user
func (r *ReconcileWebSphereLiberty) hasUserEncryptionKeySecret(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, error) {
	metaName := ""
	if passwordEncryptionMetadata != nil {
		metaName = passwordEncryptionMetadata.Name
	}
	return common.GetSecret(r.GetClient(), lutils.PasswordEncryptionKeyRootName+metaName, instance.GetNamespace())
}

// Returns true if a user secret is mirrored to a corresponding "<user>-internal" secret
func (r *ReconcileWebSphereLiberty) isSecretMirrored(instance *wlv1.WebSphereLibertyApplication,
	passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata,
	hasUserSecretFunc func(*wlv1.WebSphereLibertyApplication, *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, error),
	hasInternalSecretFunc func(*wlv1.WebSphereLibertyApplication, *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, error),
	matchedKey string) bool {
	userSecret, err := hasUserSecretFunc(instance, passwordEncryptionMetadata)
	if err != nil {
		return false
	}
	internalSecret, err := hasInternalSecretFunc(instance, passwordEncryptionMetadata)
	if err != nil {
		return false
	}
	internalKey, _ := internalSecret.LockedData.Get(matchedKey)
	userKey, _ := userSecret.LockedData.Get(matchedKey)
	return subtle.ConstantTimeCompare(userKey, []byte{}) != 1 && subtle.ConstantTimeCompare(internalKey, userKey) == 1
}

// Mirrors an internal and user secret that syncs the value of syncedKey
func (r *ReconcileWebSphereLiberty) mirrorEncryptionKeySecretState(instance *wlv1.WebSphereLibertyApplication,
	passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata,
	hasUserSecretFunc func(*wlv1.WebSphereLibertyApplication, *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, error),
	hasInternalSecretFunc func(*wlv1.WebSphereLibertyApplication, *lutils.PasswordEncryptionMetadata) (*common.LockedBufferSecret, error),
	syncedKey string) error {
	userEncryptionSecret, userEncryptionSecretErr := hasUserSecretFunc(instance, passwordEncryptionMetadata)
	defer userEncryptionSecret.Destroy()
	userEncryptionFound := !kerrors.IsNotFound(userEncryptionSecretErr)
	// Error if there was an issue getting the userEncryptionSecret
	if userEncryptionFound && userEncryptionSecretErr != nil {
		return userEncryptionSecretErr
	}
	internalEncryptionSecret, internalEncryptionSecretErr := hasInternalSecretFunc(instance, passwordEncryptionMetadata)
	defer internalEncryptionSecret.Destroy()
	internalEncryptionFound := !kerrors.IsNotFound(internalEncryptionSecretErr)
	// Error if there was an issue getting the internalEncryptionSecret
	if internalEncryptionFound && internalEncryptionSecretErr != nil {
		return internalEncryptionSecretErr
	}
	// Case 0: no user encryption secret, no internal encryption secret: secrets already mirrored
	// Case 1: no user encryption secret, internal encryption secret exists
	if !userEncryptionFound {
		if !internalEncryptionFound {
			return fmt.Errorf("failed to get internal encryption key secret")
		} else {
			if err := r.DeleteSecretResource(internalEncryptionSecret); err != nil {
				return err
			}
		}
		return fmt.Errorf("failed to get encryption key secret")
	}

	// Case 2: user encryption secret exists, no internal secret: Create internalEncryptionSecret
	// Case 3: user encryption secret exists, internal secret exists: Update internalEncryptionSecret
	if internalEncryptionSecret.LockedData == nil {
		internalEncryptionSecret.LockedData = make(common.SecretMap)
	}
	if userEncryptionSecret.LockedData == nil {
		userEncryptionSecret.LockedData = make(common.SecretMap)
	}
	internalPasswordEncryptionKey, _ := internalEncryptionSecret.LockedData.Get(syncedKey) // it's possible the internal secret doesn't exist
	userPasswordEncryptionKey, found := userEncryptionSecret.LockedData.Get(syncedKey)     // however, the user encryption must exist
	if !found {
		return fmt.Errorf("could not get user encryption key data for %s", syncedKey)
	}
	if subtle.ConstantTimeCompare(internalPasswordEncryptionKey, userPasswordEncryptionKey) != 1 {
		internalEncryptionSecret.LockedData.SetCopy(syncedKey, userPasswordEncryptionKey)
		internalEncryptionSecret.LockedData.Set("lastRotation", []byte(fmt.Sprint(time.Now().Unix())))
	}
	objCleanup, err := r.CreateOrUpdateSecret(internalEncryptionSecret, nil, func() error { return nil })
	defer objCleanup()
	return err
}

// Creates the Liberty XML to mount the password encryption keys Secret into the application pods
func (r *ReconcileWebSphereLiberty) createPasswordEncryptionKeyLibertyConfig(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata, encryptionKey []byte) error {
	if len(encryptionKey) == 0 {
		return fmt.Errorf("a password encryption key was not specified")
	}

	// The Secret to hold the server.xml that will override the password encryption key for the Liberty server
	// This server.xml will be mounted in /output/liberty-operator/encryptionKey.xml
	encryptionXMLSecretName := OperatorShortName + lutils.ManagedEncryptionServerXML + passwordEncryptionMetadata.Name
	encryptionXMLSecret := &common.LockedBufferSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      encryptionXMLSecretName,
			Namespace: instance.GetNamespace(),
			Labels:    lutils.GetRequiredLabels(encryptionXMLSecretName, ""),
		},
	}
	if err := lutils.CustomizePasswordEncryptionKeyXML(encryptionXMLSecret, encryptionKey); err != nil {
		return err
	}
	objCleanup, err := r.CreateOrUpdateSecret(encryptionXMLSecret, nil, func() error { return nil })
	defer objCleanup()
	if err != nil {
		return err
	}

	// The Secret to hold the server.xml that will import the password encryption key into the Liberty server
	// This server.xml will be mounted in /config/configDropins/overrides/encryptionKeyMount.xml
	mountingXMLSecretName := OperatorShortName + lutils.ManagedEncryptionMountServerXML + passwordEncryptionMetadata.Name
	mountingXMLSecret := &common.LockedBufferSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mountingXMLSecretName,
			Namespace: instance.GetNamespace(),
			Labels:    lutils.GetRequiredLabels(mountingXMLSecretName, ""),
		},
	}

	mountDir := strings.Replace(lutils.SecureMountPath+"/"+lutils.EncryptionKeyXMLFileName, "/output", "${server.output.dir}", 1)
	lutils.CustomizeLibertyFileMountXML(mountingXMLSecret, lutils.EncryptionKeyMountXMLFileName, mountDir)
	objCleanup, err = r.CreateOrUpdateSecret(mountingXMLSecret, nil, func() error { return nil })
	defer objCleanup()
	if err != nil {
		return err
	}
	return nil
}

// Creates the Liberty XML to mount the aes encryption keys Secret into the application pods
func (r *ReconcileWebSphereLiberty) createAESEncryptionKeyLibertyConfig(instance *wlv1.WebSphereLibertyApplication, passwordEncryptionMetadata *lutils.PasswordEncryptionMetadata, encryptionKey []byte) error {
	if len(encryptionKey) == 0 {
		return fmt.Errorf("an AES encryption key was not specified")
	}

	// The Secret to hold the server.xml that will override the password encryption key for the Liberty server
	// This server.xml will be mounted in /output/liberty-operator/encryptionKey.xml
	encryptionXMLSecretName := OperatorShortName + lutils.ManagedEncryptionServerXML + passwordEncryptionMetadata.Name
	encryptionXMLSecret := &common.LockedBufferSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      encryptionXMLSecretName,
			Namespace: instance.GetNamespace(),
			Labels:    lutils.GetRequiredLabels(encryptionXMLSecretName, ""),
		},
	}

	if err := lutils.CustomizeAESEncryptionKeyXML(encryptionXMLSecret, encryptionKey); err != nil {
		return err
	}
	objCleanup, err := r.CreateOrUpdateSecret(encryptionXMLSecret, nil, func() error { return nil })
	defer objCleanup()
	if err != nil {
		return err
	}

	// The Secret to hold the server.xml that will import the password encryption key into the Liberty server
	// This server.xml will be mounted in /config/configDropins/overrides/encryptionKeyMount.xml
	mountingXMLSecretName := OperatorShortName + lutils.ManagedEncryptionMountServerXML + passwordEncryptionMetadata.Name
	mountingXMLSecret := &common.LockedBufferSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mountingXMLSecretName,
			Namespace: instance.GetNamespace(),
			Labels:    lutils.GetRequiredLabels(mountingXMLSecretName, ""),
		},
	}
	mountDir := strings.Replace(lutils.SecureMountPath+"/"+lutils.EncryptionKeyXMLFileName, "/output", "${server.output.dir}", 1)
	if err := lutils.CustomizeLibertyFileMountXML(mountingXMLSecret, lutils.EncryptionKeyMountXMLFileName, mountDir); err != nil {
		return err
	}
	objCleanup, err = r.CreateOrUpdateSecret(mountingXMLSecret, nil, func() error { return nil })
	defer objCleanup()
	if err != nil {
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
