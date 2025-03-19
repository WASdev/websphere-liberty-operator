package controller

import (
	"fmt"

	olutils "github.com/OpenLiberty/open-liberty-operator/utils"
	tree "github.com/OpenLiberty/open-liberty-operator/utils/tree"
	wlv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebSphereLibertyApplicationResourceSharingFactory struct {
	resourcesFunc              func() (olutils.LeaderTrackerMetadataList, error)
	leaderTrackersFunc         func(assetsFolder *string) ([]*unstructured.UnstructuredList, []string, error)
	createOrUpdateFunc         func(obj client.Object, owner metav1.Object, cb func() error) error
	deleteResourcesFunc        func(obj client.Object) error
	leaderTrackerNameFunc      func(map[string]interface{}) (string, error)
	cleanupUnusedResourcesFunc func() bool
	clientFunc                 func() client.Client
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) Resources() func() (olutils.LeaderTrackerMetadataList, error) {
	return rsf.resourcesFunc
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) SetResources(fn func() (olutils.LeaderTrackerMetadataList, error)) {
	rsf.resourcesFunc = fn
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) LeaderTrackers() func(*string) ([]*unstructured.UnstructuredList, []string, error) {
	return rsf.leaderTrackersFunc
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) SetLeaderTrackers(fn func(*string) ([]*unstructured.UnstructuredList, []string, error)) {
	rsf.leaderTrackersFunc = fn
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) CreateOrUpdate() func(obj client.Object, owner metav1.Object, cb func() error) error {
	return rsf.createOrUpdateFunc
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) SetCreateOrUpdate(fn func(obj client.Object, owner metav1.Object, cb func() error) error) {
	rsf.createOrUpdateFunc = fn
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) DeleteResources() func(obj client.Object) error {
	return rsf.deleteResourcesFunc
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) SetDeleteResources(fn func(obj client.Object) error) {
	rsf.deleteResourcesFunc = fn
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) LeaderTrackerName() func(map[string]interface{}) (string, error) {
	return rsf.leaderTrackerNameFunc
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) SetLeaderTrackerName(fn func(map[string]interface{}) (string, error)) {
	rsf.leaderTrackerNameFunc = fn
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) CleanupUnusedResources() func() bool {
	return rsf.cleanupUnusedResourcesFunc
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) SetCleanupUnusedResources(fn func() bool) {
	rsf.cleanupUnusedResourcesFunc = fn
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) Client() func() client.Client {
	return rsf.clientFunc
}

func (rsf *WebSphereLibertyApplicationResourceSharingFactory) SetClient(fn func() client.Client) {
	rsf.clientFunc = fn
}

func (r *ReconcileWebSphereLiberty) createResourceSharingFactoryBase() tree.ResourceSharingFactoryBase {
	rsf := &WebSphereLibertyApplicationResourceSharingFactory{}
	rsf.SetCreateOrUpdate(func(obj client.Object, owner metav1.Object, cb func() error) error {
		return r.CreateOrUpdate(obj, owner, cb)
	})
	rsf.SetDeleteResources(func(obj client.Object) error {
		return r.DeleteResource(obj)
	})
	rsf.SetCleanupUnusedResources(func() bool {
		return false
	})
	rsf.SetClient(func() client.Client {
		return r.GetClient()
	})
	return rsf
}

func (r *ReconcileWebSphereLiberty) createResourceSharingFactory(instance *wlv1.WebSphereLibertyApplication, treeMap map[string]interface{}, replaceMap map[string]map[string]string, latestOperandVersion string, leaderTrackerType string) tree.ResourceSharingFactory {
	var rsf *WebSphereLibertyApplicationResourceSharingFactory
	rsfb := r.createResourceSharingFactoryBase()
	rsf = rsfb.(*WebSphereLibertyApplicationResourceSharingFactory)
	rsf.SetLeaderTrackers(func(assetsFolder *string) ([]*unstructured.UnstructuredList, []string, error) {
		return r.WebSphereLibertyApplicationLeaderTrackerGenerator(instance, treeMap, replaceMap, latestOperandVersion, leaderTrackerType, assetsFolder)
	})
	rsf.SetLeaderTrackerName(func(obj map[string]interface{}) (string, error) {
		nameString, _, err := unstructured.NestedString(obj, "metadata", "name") // the LTPA and Password Encryption Secret will both use their .metadata.name as the leaderTracker key identifier
		return nameString, err
	})
	rsf.SetResources(func() (olutils.LeaderTrackerMetadataList, error) {
		return r.WebSphereLibertyApplicationSharedResourceGenerator(instance, treeMap, latestOperandVersion, leaderTrackerType)
	})
	return rsf

}

func (r *ReconcileWebSphereLiberty) reconcileResourceTrackingState(instance *wlv1.WebSphereLibertyApplication, leaderTrackerType string) (tree.ResourceSharingFactory, olutils.LeaderTrackerMetadataList, error) {
	treeMap, replaceMap, err := tree.ParseDecisionTree(leaderTrackerType, nil)
	if err != nil {
		return nil, nil, err
	}
	latestOperandVersion, err := tree.GetLatestOperandVersion(treeMap, "")
	if err != nil {
		return nil, nil, err
	}
	rsf := r.createResourceSharingFactory(instance, treeMap, replaceMap, latestOperandVersion, leaderTrackerType)
	trackerMetadataList, err := tree.ReconcileResourceTrackingState(instance.GetNamespace(), OperatorShortName, leaderTrackerType, rsf, treeMap, replaceMap, latestOperandVersion)
	return rsf, trackerMetadataList, err
}

func (r *ReconcileWebSphereLiberty) WebSphereLibertyApplicationSharedResourceGenerator(instance *wlv1.WebSphereLibertyApplication, treeMap map[string]interface{}, latestOperandVersion, leaderTrackerType string) (olutils.LeaderTrackerMetadataList, error) {
	// return the metadata specific to the operator version, instance configuration, and shared resource being reconciled
	if leaderTrackerType == LTPA_RESOURCE_SHARING_FILE_NAME {
		ltpaMetadataList, err := r.reconcileLTPAMetadata(instance, treeMap, latestOperandVersion, nil)
		if err != nil {
			return nil, err
		}
		return ltpaMetadataList, nil
	}
	if leaderTrackerType == PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME {
		passwordEncryptionMetadataList, err := r.reconcilePasswordEncryptionMetadata(treeMap, latestOperandVersion)
		if err != nil {
			return nil, err
		}
		return passwordEncryptionMetadataList, nil
	}
	return nil, fmt.Errorf("a leaderTrackerType was not provided when running reconcileResourceTrackingState")
}

func (r *ReconcileWebSphereLiberty) WebSphereLibertyApplicationLeaderTrackerGenerator(instance *wlv1.WebSphereLibertyApplication, treeMap map[string]interface{}, replaceMap map[string]map[string]string, latestOperandVersion string, leaderTrackerType string, assetsFolder *string) ([]*unstructured.UnstructuredList, []string, error) {
	var resourcesMatrix []*unstructured.UnstructuredList
	var resourcesRootNameList []string
	if leaderTrackerType == LTPA_RESOURCE_SHARING_FILE_NAME {
		// 1. Add LTPA key Secret
		resourcesList, resourceRootName, keyErr := r.GetLTPAKeyResources(instance, treeMap, replaceMap, latestOperandVersion, assetsFolder)
		if keyErr != nil {
			return nil, nil, keyErr
		}
		resourcesMatrix = append(resourcesMatrix, resourcesList)
		resourcesRootNameList = append(resourcesRootNameList, resourceRootName)
		// 2. Add LTPA password Secret (config 1)
		resourcesList, resourceRootName, keyErr = r.GetLTPAConfigResources(instance, treeMap, replaceMap, latestOperandVersion, assetsFolder, LTPA_CONFIG_1_RESOURCE_SHARING_FILE_NAME)
		if keyErr != nil {
			return nil, nil, keyErr
		}
		resourcesMatrix = append(resourcesMatrix, resourcesList)
		resourcesRootNameList = append(resourcesRootNameList, resourceRootName)
		// 3. Add LTPA password Secret (config 2)
		resourcesList, resourceRootName, keyErr = r.GetLTPAConfigResources(instance, treeMap, replaceMap, latestOperandVersion, assetsFolder, LTPA_CONFIG_2_RESOURCE_SHARING_FILE_NAME)
		if keyErr != nil {
			return nil, nil, keyErr
		}
		resourcesMatrix = append(resourcesMatrix, resourcesList)
		resourcesRootNameList = append(resourcesRootNameList, resourceRootName)
	} else if leaderTrackerType == PASSWORD_ENCRYPTION_RESOURCE_SHARING_FILE_NAME {
		resourcesList, resourceRootName, passwordErr := r.GetPasswordEncryptionResources(instance, treeMap, replaceMap, latestOperandVersion, assetsFolder)
		if passwordErr != nil {
			return nil, nil, passwordErr
		}
		resourcesMatrix = append(resourcesMatrix, resourcesList)
		resourcesRootNameList = append(resourcesRootNameList, resourceRootName)
	} else {
		return nil, nil, fmt.Errorf("a valid leaderTrackerType was not specified for createNewLeaderTrackerList")
	}
	return resourcesMatrix, resourcesRootNameList, nil
}

func hasLTPAKeyResourceSuffixesEnv(instance *wlv1.WebSphereLibertyApplication) (string, bool) {
	return hasResourceSuffixesEnv(instance, "LTPA_KEY_RESOURCE_SUFFIXES")
}

func hasLTPAConfigResourceSuffixesEnv(instance *wlv1.WebSphereLibertyApplication) (string, bool) {
	return hasResourceSuffixesEnv(instance, "LTPA_CONFIG_RESOURCE_SUFFIXES")
}

func hasResourceSuffixesEnv(instance *wlv1.WebSphereLibertyApplication, envName string) (string, bool) {
	for _, env := range instance.GetEnv() {
		if env.Name == envName {
			return env.Value, true
		}
	}
	return "", false
}
