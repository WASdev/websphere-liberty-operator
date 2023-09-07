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

package utils

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	wlv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	rcoutils "github.com/application-stacks/runtime-component-operator/utils"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Utility methods specific to WebSphere Liberty and its configuration

var log = logf.Log.WithName("websphereliberty_utils")

// Constant Values
const serviceabilityMountPath = "/serviceability"
const ltpaTokenEmptyDirMountPath = "/config/resources/security"
const ltpaTokenMountPath = "/output/ltpa"
const ltpaServerXMLMountPath = "/config/configDropins/overrides/"
const ssoEnvVarPrefix = "SEC_SSO_"
const OperandVersion = "1.2.2"

var editionProductID = map[wlv1.LicenseEdition]string{
	wlv1.LicenseEditionBase: "e7daacc46bbe4e2dacd2af49145a4723",
	wlv1.LicenseEditionCore: "87f3487c22f34742a799164f3f3ffa78",
	wlv1.LicenseEditionND:   "c6a988d93b0f4d1388200d40ddc84e5b",
}

var entitlementCloudPakID = map[wlv1.LicenseEntitlement]string{
	wlv1.LicenseEntitlementCP4Apps:       "4df52d2cdc374ba09f631a650ad2b5bf",
	wlv1.LicenseEntitlementFamilyEdition: "be8ae84b3dd04d81b90af0d846849182",
	wlv1.LicenseEntitlementWSHE:          "6358611af04743f99f42dadcd6e39d52",
}

// Validate if the WebSpherLibertyApplication is valid
func Validate(wlapp *wlv1.WebSphereLibertyApplication) (bool, error) {
	// Serviceability validation
	if wlapp.GetServiceability() != nil {
		if wlapp.GetServiceability().GetVolumeClaimName() == "" && wlapp.GetServiceability().GetSize() == "" {
			return false, fmt.Errorf("Invalid input for Serviceability. Specify one of the following: spec.serviceability.size, spec.serviceability.volumeClaimName")
		}
		if wlapp.GetServiceability().GetVolumeClaimName() == "" {
			if _, err := resource.ParseQuantity(wlapp.GetServiceability().GetSize()); err != nil {
				return false, fmt.Errorf("validation failed: cannot parse '%v': %v", wlapp.GetServiceability().GetSize(), err)
			}
		}
	}

	return true, nil
}

func requiredFieldMessage(fieldPaths ...string) string {
	return "must set the field(s): " + strings.Join(fieldPaths, ",")
}

// ExecuteCommandInContainer Execute command inside a container in a pod through API
func ExecuteCommandInContainer(config *rest.Config, podName, podNamespace, containerName string, command []string) (string, error) {

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed to create Clientset")
		return "", fmt.Errorf("Failed to create Clientset: %v", err.Error())
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(podNamespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Command:   command,
		Container: containerName,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("Encountered error while creating Executor: %v", err.Error())
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	if err != nil {
		return stderr.String(), fmt.Errorf("Encountered error while running command: %v ; Stderr: %v ; Error: %v", command, stderr.String(), err.Error())
	}

	return stderr.String(), nil
}

// CustomizeLibertyEnv adds configured env variables appending configured liberty settings
func CustomizeLibertyEnv(pts *corev1.PodTemplateSpec, la *wlv1.WebSphereLibertyApplication, client client.Client) error {
	// ENV variables have already been set, check if they exist before setting defaults
	targetEnv := []corev1.EnvVar{
		{Name: "WLP_LOGGING_CONSOLE_LOGLEVEL", Value: "info"},
		{Name: "WLP_LOGGING_CONSOLE_SOURCE", Value: "message,accessLog,ffdc,audit"},
		{Name: "WLP_LOGGING_CONSOLE_FORMAT", Value: "json"},
	}

	if la.GetServiceability() != nil {
		targetEnv = append(targetEnv,
			corev1.EnvVar{Name: "IBM_HEAPDUMPDIR", Value: serviceabilityMountPath},
			corev1.EnvVar{Name: "IBM_COREDIR", Value: serviceabilityMountPath},
			corev1.EnvVar{Name: "IBM_JAVACOREDIR", Value: serviceabilityMountPath},
		)
	}

	// If manageTLS is true or not set, and SEC_IMPORT_K8S_CERTS is not set then default it to "true"
	if la.GetManageTLS() == nil || *la.GetManageTLS() {
		targetEnv = append(targetEnv, corev1.EnvVar{Name: "SEC_IMPORT_K8S_CERTS", Value: "true"})
	}

	envList := pts.Spec.Containers[0].Env
	for _, v := range targetEnv {
		if _, found := findEnvVar(v.Name, envList); !found {
			pts.Spec.Containers[0].Env = append(pts.Spec.Containers[0].Env, v)
		}
	}

	/*
		if la.GetService() != nil && la.GetService().GetCertificateSecretRef() != nil {
			if err := addSecretResourceVersionAsEnvVar(pts, la, client, *la.GetService().GetCertificateSecretRef(), "SERVICE_CERT"); err != nil {
				return err
			}
		}

		if la.GetRoute() != nil && la.GetRoute().GetCertificateSecretRef() != nil {
			if err := addSecretResourceVersionAsEnvVar(pts, la, client, *la.GetRoute().GetCertificateSecretRef(), "ROUTE_CERT"); err != nil {
				return err
			}
		}
	*/

	return nil
}

func AddSecretResourceVersionAsEnvVar(pts *corev1.PodTemplateSpec, la *wlv1.WebSphereLibertyApplication, client client.Client, secretName string, envNamePrefix string) error {
	secret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: la.GetNamespace()}, secret)
	if err != nil {
		return errors.Wrapf(err, "Secret %q was not found in namespace %q", secretName, la.GetNamespace())
	}
	pts.Spec.Containers[0].Env = append(pts.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  envNamePrefix + "_SECRET_RESOURCE_VERSION",
		Value: secret.ResourceVersion})
	return nil
}

func CustomizeLibertyAnnotations(pts *corev1.PodTemplateSpec, la *wlv1.WebSphereLibertyApplication) {
	libertyAnnotations := map[string]string{
		"libertyOperator": "WebSphere Liberty",
	}
	pts.Annotations = rcoutils.MergeMaps(pts.Annotations, libertyAnnotations)
}

func CustomizeLicenseAnnotations(pts *corev1.PodTemplateSpec, la *wlv1.WebSphereLibertyApplication) {
	pid := ""
	if val, ok := editionProductID[la.Spec.License.Edition]; ok {
		pid = val
	}
	pts.Annotations["productID"] = pid
	pts.Annotations["productChargedContainers"] = "app"

	entitlement := la.Spec.License.ProductEntitlementSource

	metricValue := "PROCESSOR_VALUE_UNIT"
	if entitlement == wlv1.LicenseEntitlementWSHE || entitlement == wlv1.LicenseEntitlementCP4Apps {
		metricValue = "VIRTUAL_PROCESSOR_CORE"
	}
	pts.Annotations["productMetric"] = metricValue

	ratio := ""
	switch la.Spec.License.Edition {
	case wlv1.LicenseEditionBase:
		ratio = "4:1"
	case wlv1.LicenseEditionCore:
		ratio = "8:1"
	case wlv1.LicenseEditionND:
		ratio = "1:1"
	default:
		ratio = "4:1"
	}
	pts.Annotations["productName"] = string(la.Spec.License.Edition)

	if entitlement == wlv1.LicenseEntitlementStandalone {
		delete(pts.Annotations, "cloudpakName")
		delete(pts.Annotations, "cloudpakId")
		delete(pts.Annotations, "productCloudpakRatio")
	} else {
		pts.Annotations["cloudpakName"] = string(entitlement)
		pts.Annotations["productCloudpakRatio"] = ratio
		cloudpakId := ""
		if val, ok := entitlementCloudPakID[entitlement]; ok {
			cloudpakId = val
		}
		pts.Annotations["cloudpakId"] = cloudpakId
	}
}

// findEnvVars checks if the environment variable is already present
func findEnvVar(name string, envList []corev1.EnvVar) (*corev1.EnvVar, bool) {
	for i, val := range envList {
		if val.Name == name {
			return &envList[i], true
		}
	}
	return nil, false
}

// CreateServiceabilityPVC creates PersistentVolumeClaim for Serviceability
func CreateServiceabilityPVC(instance *wlv1.WebSphereLibertyApplication) *corev1.PersistentVolumeClaim {
	persistentVolume := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name + "-serviceability",
			Namespace:   instance.Namespace,
			Labels:      instance.GetLabels(),
			Annotations: instance.GetAnnotations(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(instance.GetServiceability().GetSize()),
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
		},
	}
	if instance.Spec.Serviceability.StorageClassName != "" {
		persistentVolume.Spec.StorageClassName = &instance.Spec.Serviceability.StorageClassName
	}
	return persistentVolume
}

// ConfigureServiceability setups the shared-storage for serviceability
func ConfigureServiceability(pts *corev1.PodTemplateSpec, la *wlv1.WebSphereLibertyApplication) {
	if la.GetServiceability() != nil {
		name := "serviceability"

		foundVolumeMount := false
		for _, v := range pts.Spec.Containers[0].VolumeMounts {
			if v.Name == name {
				foundVolumeMount = true
			}
		}

		if !foundVolumeMount {
			vm := corev1.VolumeMount{
				Name:      name,
				MountPath: serviceabilityMountPath,
			}
			pts.Spec.Containers[0].VolumeMounts = append(pts.Spec.Containers[0].VolumeMounts, vm)
		}

		foundVolume := false
		for _, v := range pts.Spec.Volumes {
			if v.Name == name {
				foundVolume = true
			}
		}

		if !foundVolume {
			claimName := la.Name + "-serviceability"
			if la.Spec.Serviceability.VolumeClaimName != "" {
				claimName = la.Spec.Serviceability.VolumeClaimName
			}
			vol := corev1.Volume{
				Name: name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: claimName,
					},
				},
			}
			pts.Spec.Volumes = append(pts.Spec.Volumes, vol)
		}
	}
}

func normalizeEnvVariableName(name string) string {
	return strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToUpper(name))
}

// getValue returns value for string
func getValue(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case bool:
		return strconv.FormatBool(v.(bool))
	default:
		return ""
	}
}

// createEnvVarSSO creates an environment variable for SSO
func createEnvVarSSO(loginID string, envSuffix string, value interface{}) *corev1.EnvVar {
	return &corev1.EnvVar{
		Name:  ssoEnvVarPrefix + loginID + envSuffix,
		Value: getValue(value),
	}
}

func writeSSOSecretIfNeeded(client client.Client, ssoSecret *corev1.Secret, ssoSecretUpdates map[string][]byte) error {
	var err error = nil
	if len(ssoSecretUpdates) > 0 {
		_, err = controllerutil.CreateOrUpdate(context.TODO(), client, ssoSecret, func() error {
			for key, value := range ssoSecretUpdates {
				ssoSecret.Data[key] = value
			}
			return nil
		})
	}
	return err
}

// CustomizeEnvSSO Process the configuration for SSO login providers
func CustomizeEnvSSO(pts *corev1.PodTemplateSpec, instance *wlv1.WebSphereLibertyApplication, client client.Client, isOpenShift bool) error {
	const ssoSecretNameSuffix = "-wlapp-sso"
	const autoregFragment = "-autoreg-"
	secretName := instance.GetName() + ssoSecretNameSuffix
	ssoSecret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.GetNamespace()}, ssoSecret)
	if err != nil {
		return errors.Wrapf(err, "Secret for Single sign-on (SSO) was not found. Create a secret named %q in namespace %q with the credentials for the login providers you selected in application image.", secretName, instance.GetNamespace())
	}

	ssoEnv := []corev1.EnvVar{}

	var secretKeys []string
	for k := range ssoSecret.Data { //ranging over a map returns it's keys.
		if strings.Contains(k, autoregFragment) { // skip -autoreg-
			continue
		}
		secretKeys = append(secretKeys, k)
	}
	sort.Strings(secretKeys)

	// append all the values in the secret into the env vars.
	for _, k := range secretKeys {
		ssoEnv = append(ssoEnv, corev1.EnvVar{
			Name: ssoEnvVarPrefix + normalizeEnvVariableName(k),
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ssoSecret.GetName(),
					},
					Key: k,
				},
			},
		})
	}

	// append all the values in the spec into the env vars.
	sso := instance.Spec.SSO
	if sso.MapToUserRegistry != nil {
		ssoEnv = append(ssoEnv, *createEnvVarSSO("", "MAPTOUSERREGISTRY", *sso.MapToUserRegistry))
	}

	if sso.RedirectToRPHostAndPort != "" {
		ssoEnv = append(ssoEnv, *createEnvVarSSO("", "REDIRECTTORPHOSTANDPORT", sso.RedirectToRPHostAndPort))
	}

	if sso.Github != nil && sso.Github.Hostname != "" {
		ssoEnv = append(ssoEnv, *createEnvVarSSO("", "GITHUB_HOSTNAME", sso.Github.Hostname))
	}

	ssoSecretUpdates := make(map[string][]byte)
	for _, oidcClient := range sso.OIDC {
		id := strings.ToUpper(oidcClient.ID)
		if id == "" {
			id = "OIDC"
		}
		if oidcClient.DiscoveryEndpoint != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_DISCOVERYENDPOINT", oidcClient.DiscoveryEndpoint))
		}
		if oidcClient.GroupNameAttribute != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_GROUPNAMEATTRIBUTE", oidcClient.GroupNameAttribute))
		}
		if oidcClient.UserNameAttribute != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_USERNAMEATTRIBUTE", oidcClient.UserNameAttribute))
		}
		if oidcClient.DisplayName != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_DISPLAYNAME", oidcClient.DisplayName))
		}
		if oidcClient.UserInfoEndpointEnabled != nil {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_USERINFOENDPOINTENABLED", *oidcClient.UserInfoEndpointEnabled))
		}
		if oidcClient.RealmNameAttribute != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_REALMNAMEATTRIBUTE", oidcClient.RealmNameAttribute))
		}
		if oidcClient.Scope != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_SCOPE", oidcClient.Scope))
		}
		if oidcClient.TokenEndpointAuthMethod != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_TOKENENDPOINTAUTHMETHOD", oidcClient.TokenEndpointAuthMethod))
		}
		if oidcClient.HostNameVerificationEnabled != nil {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_HOSTNAMEVERIFICATIONENABLED", *oidcClient.HostNameVerificationEnabled))
		}

		clientName := oidcClient.ID
		if clientName == "" {
			clientName = "oidc"
		}
		// if no clientId specified for this provider, try auto-registration
		clientId := string(ssoSecret.Data[clientName+"-clientId"])
		clientSecret := string(ssoSecret.Data[clientName+"-clientSecret"])

		if isOpenShift && clientId == "" {
			logf.Log.WithName("utils").Info("Processing OIDC registration for id :" + clientName)
			theRoute := &routev1.Route{}
			err = client.Get(context.TODO(), types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, theRoute)
			if err != nil {
				// if route is unavailable, we want to let reconciliation proceed so it will be created.
				// Update status of the instance so reconcilation will be triggered again.
				b := false
				instance.Status.RouteAvailable = &b
				logf.Log.WithName("utils").Info("CustomizeEnvSSO waiting for route to become available for provider " + clientName + ", requeue")
				return nil
			}

			// route available, we don't have a client id and secret yet, go get one
			prefix := clientName + autoregFragment
			buf := string(ssoSecret.Data[prefix+"insecureTLS"])
			insecure := strings.ToUpper(buf) == "TRUE"
			regData := RegisterData{
				DiscoveryURL:            oidcClient.DiscoveryEndpoint,
				RouteURL:                "https://" + theRoute.Spec.Host,
				RedirectToRPHostAndPort: sso.RedirectToRPHostAndPort,
				InitialAccessToken:      string(ssoSecret.Data[prefix+"initialAccessToken"]),
				InitialClientId:         string(ssoSecret.Data[prefix+"initialClientId"]),
				InitialClientSecret:     string(ssoSecret.Data[prefix+"initialClientSecret"]),
				GrantTypes:              string(ssoSecret.Data[prefix+"grantTypes"]),
				Scopes:                  string(ssoSecret.Data[prefix+"scopes"]),
				InsecureTLS:             insecure,
				ProviderId:              clientName,
			}

			clientId, clientSecret, err = RegisterWithOidcProvider(regData)
			if err != nil {
				writeSSOSecretIfNeeded(client, ssoSecret, ssoSecretUpdates) // preserve any registrations that succeeded
				return errors.Wrapf(err, "Error occured during registration with OIDC for provider "+clientName)
			}
			logf.Log.WithName("utils").Info("OIDC registration for id: " + clientName + " successful, obtained clientId: " + clientId)
			ssoSecretUpdates[clientName+autoregFragment+"RegisteredOidcClientId"] = []byte(clientId)
			ssoSecretUpdates[clientName+autoregFragment+"RegisteredOidcSecret"] = []byte(clientSecret)
			ssoSecretUpdates[clientName+"-clientId"] = []byte(clientId)
			ssoSecretUpdates[clientName+"-clientSecret"] = []byte(clientSecret)

			b := true
			instance.Status.RouteAvailable = &b
		} // end auto-reg
	} // end for
	err = writeSSOSecretIfNeeded(client, ssoSecret, ssoSecretUpdates)

	if err != nil {
		return errors.Wrapf(err, "Error occured when updating SSO secret")
	}

	for _, oauth2Client := range sso.Oauth2 {
		id := strings.ToUpper(oauth2Client.ID)
		if id == "" {
			id = "OAUTH2"
		}
		if oauth2Client.TokenEndpoint != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_TOKENENDPOINT", oauth2Client.TokenEndpoint))
		}
		if oauth2Client.AuthorizationEndpoint != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_AUTHORIZATIONENDPOINT", oauth2Client.AuthorizationEndpoint))
		}
		if oauth2Client.GroupNameAttribute != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_GROUPNAMEATTRIBUTE", oauth2Client.GroupNameAttribute))
		}
		if oauth2Client.UserNameAttribute != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_USERNAMEATTRIBUTE", oauth2Client.UserNameAttribute))
		}
		if oauth2Client.DisplayName != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_DISPLAYNAME", oauth2Client.DisplayName))
		}
		if oauth2Client.RealmNameAttribute != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_REALMNAMEATTRIBUTE", oauth2Client.RealmNameAttribute))
		}
		if oauth2Client.RealmName != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_REALMNAME", oauth2Client.RealmName))
		}
		if oauth2Client.Scope != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_SCOPE", oauth2Client.Scope))
		}
		if oauth2Client.TokenEndpointAuthMethod != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_TOKENENDPOINTAUTHMETHOD", oauth2Client.TokenEndpointAuthMethod))
		}
		if oauth2Client.AccessTokenHeaderName != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_ACCESSTOKENHEADERNAME", oauth2Client.AccessTokenHeaderName))
		}
		if oauth2Client.AccessTokenRequired != nil {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_ACCESSTOKENREQUIRED", *oauth2Client.AccessTokenRequired))
		}
		if oauth2Client.AccessTokenSupported != nil {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_ACCESSTOKENSUPPORTED", *oauth2Client.AccessTokenSupported))
		}
		if oauth2Client.UserApiType != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_USERAPITYPE", oauth2Client.UserApiType))
		}
		if oauth2Client.UserApi != "" {
			ssoEnv = append(ssoEnv, *createEnvVarSSO(id, "_USERAPI", oauth2Client.UserApi))
		}
	}

	secretRev := corev1.EnvVar{
		Name:  "SSO_SECRET_REV",
		Value: ssoSecret.ResourceVersion}
	ssoEnv = append(ssoEnv, secretRev)

	envList := pts.Spec.Containers[0].Env
	for _, v := range ssoEnv {
		if _, found := findEnvVar(v.Name, envList); !found {
			pts.Spec.Containers[0].Env = append(pts.Spec.Containers[0].Env, v)
		}
	}
	return nil
}

func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func Remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func GetWLOLicenseAnnotations() map[string]string {
	annotations := make(map[string]string)
	annotations["productID"] = "cb1747ecb831410f88006195f024183f"
	annotations["productName"] = "WebSphere Liberty Operator"
	annotations["productMetric"] = "FREE"
	annotations["productChargedContainers"] = "ALL"
	return annotations
}

func isVolumeMountFound(pts *corev1.PodTemplateSpec, name string) bool {
	for _, v := range pts.Spec.Containers[0].VolumeMounts {
		if v.Name == name {
			return true
		}
	}
	return false
}

func isVolumeFound(pts *corev1.PodTemplateSpec, name string) bool {
	for _, v := range pts.Spec.Volumes {
		if v.Name == name {
			return true
		}
	}
	return false
}

// ConfigureLTPA setups the shared-storage for LTPA token generation
func ConfigureLTPA(pts *corev1.PodTemplateSpec, la *wlv1.WebSphereLibertyApplication) {
	// Create emptyDir at /output/resources/security
	emptyDirLtpaVolumeMount := GetEmptyDirLTPAVolumeMount(la, false)
	if !isVolumeMountFound(pts, emptyDirLtpaVolumeMount.Name) {
		pts.Spec.Containers[0].VolumeMounts = append(pts.Spec.Containers[0].VolumeMounts, emptyDirLtpaVolumeMount)
	}
	if !isVolumeFound(pts, emptyDirLtpaVolumeMount.Name) {
		vol := corev1.Volume{
			Name: emptyDirLtpaVolumeMount.Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
		pts.Spec.Volumes = append(pts.Spec.Volumes, vol)
	}
	// Create emptyDir at /config/configDropins/overrides
	emptyDirLtpaServerXMLVolumeMount := GetEmptyDirLTPAServerXMLVolumeMount(la)
	if !isVolumeMountFound(pts, emptyDirLtpaServerXMLVolumeMount.Name) {
		pts.Spec.Containers[0].VolumeMounts = append(pts.Spec.Containers[0].VolumeMounts, emptyDirLtpaServerXMLVolumeMount)
	}
	if !isVolumeFound(pts, emptyDirLtpaServerXMLVolumeMount.Name) {
		vol := corev1.Volume{
			Name: emptyDirLtpaServerXMLVolumeMount.Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
		pts.Spec.Volumes = append(pts.Spec.Volumes, vol)
	}

	// Add LTPA key into the ltpa folder
	ltpaKeyVolumeMount := GetLTPAVolumeMount(la, "keys", true)
	if !isVolumeMountFound(pts, ltpaKeyVolumeMount.Name) {
		pts.Spec.Containers[0].VolumeMounts = append(pts.Spec.Containers[0].VolumeMounts, ltpaKeyVolumeMount)
	}
	if !isVolumeFound(pts, ltpaKeyVolumeMount.Name) {
		claimName := la.GetName() + "-ltpa-token-pvc"
		vol := corev1.Volume{
			Name: ltpaKeyVolumeMount.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
				},
			},
		}
		pts.Spec.Volumes = append(pts.Spec.Volumes, vol)
	}

	// Add LTPA server.xml into the ltpa folder
	ltpaXMLVolumeMount := GetLTPAVolumeMount(la, "xml", true)
	if !isVolumeMountFound(pts, ltpaXMLVolumeMount.Name) {
		pts.Spec.Containers[0].VolumeMounts = append(pts.Spec.Containers[0].VolumeMounts, ltpaXMLVolumeMount)
	}
	if !isVolumeFound(pts, ltpaXMLVolumeMount.Name) {
		vol := corev1.Volume{
			Name: ltpaXMLVolumeMount.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: la.GetName() + "-ltpa-server-xml",
					},
				},
			},
		}
		pts.Spec.Volumes = append(pts.Spec.Volumes, vol)
	}

	// Create initContainer. Copy the LTPA key from the volume containing server.xml into the emptyDir volume
	ltpaKeyInitContainer := corev1.Container{
		Name:         "copy-ltpa-keys",
		Image:        "registry.access.redhat.com/ubi9/ubi",
		Command:      []string{"sh", "-c", "cp -f " + ltpaTokenMountPath + "/keys/ltpa.keys " + ltpaTokenEmptyDirMountPath + "; cp -f " + ltpaTokenMountPath + "/xml/ltpa.xml " + ltpaServerXMLMountPath},
		VolumeMounts: []corev1.VolumeMount{},
	}
	ltpaKeyInitContainer.VolumeMounts = append(ltpaKeyInitContainer.VolumeMounts, emptyDirLtpaVolumeMount)
	ltpaKeyInitContainer.VolumeMounts = append(ltpaKeyInitContainer.VolumeMounts, emptyDirLtpaServerXMLVolumeMount)
	ltpaKeyInitContainer.VolumeMounts = append(ltpaKeyInitContainer.VolumeMounts, ltpaKeyVolumeMount)
	ltpaKeyInitContainer.VolumeMounts = append(ltpaKeyInitContainer.VolumeMounts, ltpaXMLVolumeMount)
	pts.Spec.InitContainers = append(pts.Spec.InitContainers, ltpaKeyInitContainer)
}

func CustomizeLTPAPersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, la *wlv1.WebSphereLibertyApplication) {
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("500Mi"),
			},
		},
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteMany,
		},
	}
}

func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphabetLength := len(alphabet)
	buffer := make([]byte, length)
	for i := range buffer {
		buffer[i] = alphabet[rand.Intn(alphabetLength)]
	}
	return string(buffer)

}

func CustomizeLTPAServerXML(configMap *corev1.ConfigMap, la *wlv1.WebSphereLibertyApplication, encryptedPassword string) {
	configMap.Data = make(map[string]string)
	configMap.Data["ltpa.xml"] = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<server>\n    <ltpa keysFileName=\"" + ltpaTokenEmptyDirMountPath + "/ltpa.keys\" keysPassword=\"" + encryptedPassword + "\" />\n</server>"
}

func CustomizeLTPAJob(job *v1.Job, la *wlv1.WebSphereLibertyApplication, password string) {
	keyDirectory := ltpaTokenMountPath + "/keys"
	keyFile := keyDirectory + "/ltpa.keys"
	job.Spec.Template.ObjectMeta.Name = "liberty"
	job.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:    job.Spec.Template.ObjectMeta.Name,
			Image:   la.GetApplicationImage(),
			Command: []string{"/bin/bash", "-c"},
			Args:    []string{"mkdir -p " + keyDirectory + " && rm -f " + keyFile + " && securityUtility createLTPAKeys --file=" + keyFile + " --password=" + password + " --passwordEncoding=aes"},
			VolumeMounts: []corev1.VolumeMount{
				// Set the LTPA Job's volume mount as read/writeable
				GetLTPAVolumeMount(la, "keys", false),
			},
		},
	}
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
	job.Spec.Template.Spec.Volumes = GetLTPAVolume(la, "keys")
}

func CustomizeLTPAEncryptedPasswordJob(job *v1.Job, la *wlv1.WebSphereLibertyApplication, password string) {
	job.Spec.Template.ObjectMeta.Name = "liberty"
	job.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:    job.Spec.Template.ObjectMeta.Name,
			Image:   la.GetApplicationImage(),
			Command: []string{"/bin/bash", "-c"},
			Args:    []string{"securityUtility encode --encoding=aes " + password},
		},
	}
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
}

func CustomizeLTPASecret(secret *corev1.Secret, la *wlv1.WebSphereLibertyApplication, password string) {
	secret.Type = corev1.SecretTypeOpaque
	secretStringData := make(map[string]string)
	secretStringData["password"] = password
	secretStringData["lastRotation"] = time.Now().UTC().String()
	secret.StringData = secretStringData
}

func GetEmptyDirLTPAVolumeMount(la *wlv1.WebSphereLibertyApplication, isReadOnly bool) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      la.GetName() + "-shared-empty-dir-ltpa-volume",
		MountPath: ltpaTokenEmptyDirMountPath,
		ReadOnly:  isReadOnly,
	}
}

func GetLTPAVolumeMount(la *wlv1.WebSphereLibertyApplication, subFolder string, isReadOnly bool) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      la.GetName() + "-shared-ltpa-volume-" + subFolder,
		MountPath: ltpaTokenMountPath + "/" + subFolder,
		ReadOnly:  isReadOnly,
	}
}

func GetEmptyDirLTPAServerXMLVolumeMount(la *wlv1.WebSphereLibertyApplication) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      la.GetName() + "-shared-empty-dir-ltpa-server-xml-volume",
		MountPath: ltpaServerXMLMountPath,
	}
}

func GetLTPAVolume(la *wlv1.WebSphereLibertyApplication, subFolder string) []corev1.Volume {
	return []corev1.Volume{{
		Name: la.GetName() + "-shared-ltpa-volume-" + subFolder,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: la.GetName() + "-ltpa-token-pvc",
			},
		},
	}}
}
