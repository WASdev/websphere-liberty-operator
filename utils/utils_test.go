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
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	webspherelibertyv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	oputils "github.com/application-stacks/runtime-component-operator/utils"
	routev1 "github.com/openshift/api/route/v1"
	v1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	coretesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	name                = "app"
	namespace           = "webspherelibertyv1"
	appImage            = "my-image"
	consoleFormat       = "json"
	replicas      int32 = 3
	clusterType         = corev1.ServiceTypeClusterIP
)

type Test struct {
	test     string
	expected interface{}
	actual   interface{}
}

func TestCustomizeLibertyEnv(t *testing.T) {
	logger := zap.New()
	logf.SetLogger(logger)
	os.Setenv("WATCH_NAMESPACE", namespace)

	// Test default values no config
	svc := &webspherelibertyv1.WebSphereLibertyApplicationService{Port: 8080, Type: &clusterType}
	spec := webspherelibertyv1.WebSphereLibertyApplicationSpec{Service: svc}
	pts := &corev1.PodTemplateSpec{}

	targetEnv := []corev1.EnvVar{
		{Name: "TLS_DIR", Value: "/etc/x509/certs"},
		{Name: "WLP_LOGGING_CONSOLE_LOGLEVEL", Value: "info"},
		{Name: "WLP_LOGGING_CONSOLE_SOURCE", Value: "message,accessLog,ffdc,audit"},
		{Name: "WLP_LOGGING_CONSOLE_FORMAT", Value: "json"},
		{Name: "SEC_IMPORT_K8S_CERTS", Value: "true"},
	}
	// Always call CustomizePodSpec to populate Containers & simulate real behaviour
	wl := createWebSphereLibertyApp(name, namespace, spec)
	objs, s := []runtime.Object{wl}, scheme.Scheme
	s.AddKnownTypes(webspherelibertyv1.GroupVersion, wl)

	cl := fakeclient.NewFakeClient(objs...)
	rcl := fakeclient.NewFakeClient(objs...)

	rb := oputils.NewReconcilerBase(rcl, cl, s, &rest.Config{}, record.NewFakeRecorder(10))

	oputils.CustomizePodSpec(pts, wl)
	CustomizeLibertyEnv(pts, wl, rb.GetClient())

	testEnv := []Test{
		{"Test environment defaults", targetEnv, pts.Spec.Containers[0].Env},
	}

	if err := verifyTests(testEnv); err != nil {
		t.Fatalf("%v", err)
	}

	// test with env variables set by user
	userEnv := []corev1.EnvVar{
		{Name: "WLP_LOGGING_CONSOLE_LOGLEVEL", Value: "error"},
		{Name: "WLP_LOGGING_CONSOLE_SOURCE", Value: "trace,accessLog,ffdc"},
		{Name: "WLP_LOGGING_CONSOLE_FORMAT", Value: "basic"},
		{Name: "SEC_IMPORT_K8S_CERTS", Value: "false"},
	}

	spec = webspherelibertyv1.WebSphereLibertyApplicationSpec{
		Env:     userEnv,
		Service: svc,
	}
	pts = &corev1.PodTemplateSpec{}

	wl = createWebSphereLibertyApp(name, namespace, spec)
	oputils.CustomizePodSpec(pts, wl)
	CustomizeLibertyEnv(pts, wl, rb.GetClient())

	expectedEnv := append(userEnv, corev1.EnvVar{Name: "TLS_DIR", Value: "/etc/x509/certs"})
	testEnv = []Test{
		{"Test environment config", expectedEnv, pts.Spec.Containers[0].Env},
	}
	if err := verifyTests(testEnv); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestCustomizeEnvSSO(t *testing.T) {
	logger := zap.New()
	logf.SetLogger(logger)
	os.Setenv("WATCH_NAMESPACE", namespace)
	svc := &webspherelibertyv1.WebSphereLibertyApplicationService{Port: 8080, Type: &clusterType}
	spec := webspherelibertyv1.WebSphereLibertyApplicationSpec{Service: svc}
	liberty := createWebSphereLibertyApp(name, namespace, spec)
	objs, s := []runtime.Object{liberty}, scheme.Scheme
	s.AddKnownTypes(webspherelibertyv1.GroupVersion, liberty)

	cl := fakeclient.NewFakeClient(objs...)
	rcl := fakeclient.NewFakeClient(objs...)

	rb := oputils.NewReconcilerBase(rcl, cl, s, &rest.Config{}, record.NewFakeRecorder(10))

	terminationPolicy := v1.TLSTerminationReencrypt
	expose := true
	spec.Env = []corev1.EnvVar{
		{Name: "SEC_TLS_TRUSTDEFAULTCERTS", Value: "true"},
		{Name: "SEC_IMPORT_K8S_CERTS", Value: "true"},
	}
	spec.Expose = &expose
	spec.Route = &webspherelibertyv1.WebSphereLibertyApplicationRoute{
		Host:        "myapp.mycompany.com",
		Termination: &terminationPolicy,
	}
	spec.SSO = &webspherelibertyv1.WebSphereLibertyApplicationSSO{
		RedirectToRPHostAndPort: "redirectvalue",
		MapToUserRegistry:       &expose,
		Github:                  &webspherelibertyv1.GithubLogin{Hostname: "github.com"},
		OIDC: []webspherelibertyv1.OidcClient{
			{
				DiscoveryEndpoint:           "myapp.mycompany.com",
				ID:                          "custom3",
				GroupNameAttribute:          "specify-required-value1",
				UserNameAttribute:           "specify-required-value2",
				DisplayName:                 "specify-required-value3",
				UserInfoEndpointEnabled:     &expose,
				RealmNameAttribute:          "specify-required-value4",
				Scope:                       "specify-required-value5",
				TokenEndpointAuthMethod:     "specify-required-value6",
				HostNameVerificationEnabled: &expose,
			},
		},
		Oauth2: []webspherelibertyv1.OAuth2Client{
			{
				ID:                    "custom1",
				AuthorizationEndpoint: "specify-required-value",
				TokenEndpoint:         "specify-required-value",
			},
			{
				ID:                      "custom2",
				AuthorizationEndpoint:   "specify-required-value1",
				TokenEndpoint:           "specify-required-value2",
				GroupNameAttribute:      "specify-required-value3",
				UserNameAttribute:       "specify-required-value4",
				DisplayName:             "specify-required-value5",
				RealmNameAttribute:      "specify-required-value6",
				RealmName:               "specify-required-value7",
				Scope:                   "specify-required-value8",
				TokenEndpointAuthMethod: "specify-required-value9",
				AccessTokenHeaderName:   "specify-required-value10",
				AccessTokenRequired:     &expose,
				AccessTokenSupported:    &expose,
				UserApiType:             "specify-required-value11",
				UserApi:                 "specify-required-value12",
			},
		},
	}
	data := map[string][]byte{
		"github-clientId":     []byte("bW9vb29vb28="),
		"github-clientSecret": []byte("dGhlbGF1Z2hpbmdjb3c="),
		"oidc-clientId":       []byte("bW9vb29vb28="),
		"oidc-clientSecret":   []byte("dGhlbGF1Z2hpbmdjb3c="),
	}
	pts := &corev1.PodTemplateSpec{}
	ssoSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-wlapp-sso",
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	err := rb.GetClient().Create(context.TODO(), ssoSecret)
	if err != nil {
		t.Fatalf(err.Error())
	}
	wl := createWebSphereLibertyApp(name, namespace, spec)
	oputils.CustomizePodSpec(pts, wl)
	CustomizeEnvSSO(pts, wl, rb.GetClient(), false)

	podEnv := envSliceToMap(pts.Spec.Containers[0].Env, data, t)
	tests := []Test{
		{"Github clientid set", string(data["github-clientId"]), podEnv["SEC_SSO_GITHUB_CLIENTID"]},
		{"Github clientSecret set", string(data["github-clientSecret"]), podEnv["SEC_SSO_GITHUB_CLIENTSECRET"]},
		{"OIDC clientId set", string(data["oidc-clientId"]), podEnv["SEC_SSO_OIDC_CLIENTID"]},
		{"OIDC clientSecret set", string(data["oidc-clientSecret"]), podEnv["SEC_SSO_OIDC_CLIENTSECRET"]},
		{"redirect to rp host and port", "redirectvalue", podEnv["SEC_SSO_REDIRECTTORPHOSTANDPORT"]},
		{"map to user registry", "true", podEnv["SEC_SSO_MAPTOUSERREGISTRY"]},
		{"Github hostname set", "github.com", podEnv["SEC_SSO_GITHUB_HOSTNAME"]},
		{"oidc-custom3 discovery endpoint", "myapp.mycompany.com", podEnv["SEC_SSO_CUSTOM3_DISCOVERYENDPOINT"]},
		{"oidc-custom3 group name attribute", "specify-required-value1", podEnv["SEC_SSO_CUSTOM3_GROUPNAMEATTRIBUTE"]},
		{"oidc-custom3 user name attribute", "specify-required-value2", podEnv["SEC_SSO_CUSTOM3_USERNAMEATTRIBUTE"]},
		{"oidc-custom3 display name", "specify-required-value3", podEnv["SEC_SSO_CUSTOM3_DISPLAYNAME"]},
		{"oidc-custom3 user info endpoint enabled", "true", podEnv["SEC_SSO_CUSTOM3_USERINFOENDPOINTENABLED"]},
		{"oidc-custom3 realm name attribute", "specify-required-value4", podEnv["SEC_SSO_CUSTOM3_REALMNAMEATTRIBUTE"]},
		{"oidc-custom3 scope", "specify-required-value5", podEnv["SEC_SSO_CUSTOM3_SCOPE"]},
		{"oidc-custom3 token endpoint auth method", "specify-required-value6", podEnv["SEC_SSO_CUSTOM3_TOKENENDPOINTAUTHMETHOD"]},
		{"oidc-custom3 host name verification enabled", "true", podEnv["SEC_SSO_CUSTOM3_HOSTNAMEVERIFICATIONENABLED"]},
		{"oauth2-custom1 authorization endpoint", "specify-required-value", podEnv["SEC_SSO_CUSTOM1_AUTHORIZATIONENDPOINT"]},
		{"oauth2-custom1 token endpoint", "specify-required-value", podEnv["SEC_SSO_CUSTOM1_TOKENENDPOINT"]},
		{"oauth2-custom2 authorization endpoint", "specify-required-value1", podEnv["SEC_SSO_CUSTOM2_AUTHORIZATIONENDPOINT"]},
		{"oauth2-custom2 token endpoint", "specify-required-value2", podEnv["SEC_SSO_CUSTOM2_TOKENENDPOINT"]},
		{"oauth2-custom2 group name attribute", "specify-required-value3", podEnv["SEC_SSO_CUSTOM2_GROUPNAMEATTRIBUTE"]},
		{"oauth2-custom2 user name attribute", "specify-required-value4", podEnv["SEC_SSO_CUSTOM2_USERNAMEATTRIBUTE"]},
		{"oauth2-custom2 display name", "specify-required-value5", podEnv["SEC_SSO_CUSTOM2_DISPLAYNAME"]},
		{"oauth2-custom2 realm name attribute", "specify-required-value6", podEnv["SEC_SSO_CUSTOM2_REALMNAMEATTRIBUTE"]},
		{"oauth2-custom2 realm name", "specify-required-value7", podEnv["SEC_SSO_CUSTOM2_REALMNAME"]},
		{"oauth2-custom2 scope", "specify-required-value8", podEnv["SEC_SSO_CUSTOM2_SCOPE"]},
		{"oauth2-custom2 token endpoint auth method", "specify-required-value9", podEnv["SEC_SSO_CUSTOM2_TOKENENDPOINTAUTHMETHOD"]},
		{"oauth2-custom2 access token header name", "specify-required-value10", podEnv["SEC_SSO_CUSTOM2_ACCESSTOKENHEADERNAME"]},
		{"oauth2-custom2 access token required", "true", podEnv["SEC_SSO_CUSTOM2_ACCESSTOKENREQUIRED"]},
		{"oauth2-custom2 access token supported", "true", podEnv["SEC_SSO_CUSTOM2_ACCESSTOKENSUPPORTED"]},
		{"oauth2-custom2 user api type", "specify-required-value11", podEnv["SEC_SSO_CUSTOM2_USERAPITYPE"]},
		{"oauth2-custom2 user api", "specify-required-value12", podEnv["SEC_SSO_CUSTOM2_USERAPI"]},
	}

	if err := verifyTests(tests); err != nil {
		t.Fatalf("%v", err)
	}
}

type licenseTestData struct {
	// input from the spec
	metric  webspherelibertyv1.LicenseMetric
	edition webspherelibertyv1.LicenseEdition
	pes     webspherelibertyv1.LicenseEntitlement
	// whether CustomizeLicenseAnnotations is expected to return an err or not
	pass bool
}

func TestCustomizeLicenseAnnotations(t *testing.T) {
	t.Log("Starting license test")

	td := []licenseTestData{}
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionBase, pes: webspherelibertyv1.LicenseEntitlementStandalone})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionBase, pes: webspherelibertyv1.LicenseEntitlementCP4Apps})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionBase, pes: webspherelibertyv1.LicenseEntitlementFamilyEdition})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionBase, pes: webspherelibertyv1.LicenseEntitlementWSHE})

	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionCore, pes: webspherelibertyv1.LicenseEntitlementStandalone})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionCore, pes: webspherelibertyv1.LicenseEntitlementCP4Apps})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionCore, pes: webspherelibertyv1.LicenseEntitlementFamilyEdition})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionCore, pes: webspherelibertyv1.LicenseEntitlementWSHE})

	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionND, pes: webspherelibertyv1.LicenseEntitlementStandalone})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionND, pes: webspherelibertyv1.LicenseEntitlementCP4Apps})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionND, pes: webspherelibertyv1.LicenseEntitlementFamilyEdition})
	td = append(td, licenseTestData{edition: webspherelibertyv1.LicenseEditionND, pes: webspherelibertyv1.LicenseEntitlementWSHE})

	for _, s := range td {
		t.Logf("Testing %#v\n", s)
		spec := webspherelibertyv1.WebSphereLibertyApplicationSpec{}
		spec.License.Metric = s.metric
		spec.License.ProductEntitlementSource = s.pes
		spec.License.Edition = s.edition
		app := createWebSphereLibertyApp("myapp", "myns", spec)
		pts := &corev1.PodTemplateSpec{}
		pts.Annotations = make(map[string]string)

		CustomizeLicenseAnnotations(pts, app)

		// "productChargedContainers" should always be set to 'app'
		if pts.Annotations["productChargedContainers"] != "app" {
			t.Logf("productChargedContainers: expected 'app' but was %s\n", pts.Annotations["productChargedContainers"])
			t.Error("pcc was wrong")
		} else {
			t.Log("pcc was correct")
		}
		// 'productMetric' is PVU for Standalone and Family Edition. VPU otherwise.
		if s.pes == webspherelibertyv1.LicenseEntitlementStandalone || s.pes == webspherelibertyv1.LicenseEntitlementFamilyEdition {
			if pts.Annotations["productMetric"] != "PROCESSOR_VALUE_UNIT" {
				t.Errorf("product metric: expected 'PROCESSOR_VALUE_UNIT' but was %s\n", pts.Annotations["productMetric"])
			}
		} else {
			if pts.Annotations["productMetric"] != "VIRTUAL_PROCESSOR_CORE" {
				t.Errorf("product metric: expected 'VIRTUAL_PROCESSOR_CORE' but was %s\n", pts.Annotations["productMetric"])
			}
		}
		// 'productID' and 'productName' should always be direct mappings from spec.License.Edition
		switch s.edition {
		case webspherelibertyv1.LicenseEditionCore:
			if pts.Annotations["productID"] != "87f3487c22f34742a799164f3f3ffa78" {
				t.Errorf("Incorrect productID for edition core: %s\n", pts.Annotations["productID"])
			}
			if pts.Annotations["productName"] != "IBM WebSphere Application Server Liberty Core" {
				t.Errorf("Incorrect productName for edition core: %s\n", pts.Annotations["productName"])
			}
		case webspherelibertyv1.LicenseEditionBase:
			if pts.Annotations["productID"] != "e7daacc46bbe4e2dacd2af49145a4723" {
				t.Errorf("Incorrect productID for edition base: %s\n", pts.Annotations["productID"])
			}
			if pts.Annotations["productName"] != "IBM WebSphere Application Server" {
				t.Errorf("Incorrect productName for edition base: %s\n", pts.Annotations["productName"])
			}
		case webspherelibertyv1.LicenseEditionND:
			if pts.Annotations["productID"] != "c6a988d93b0f4d1388200d40ddc84e5b" {
				t.Errorf("Incorrect productID for edition ND: %s\n", pts.Annotations["productID"])
			}
			if pts.Annotations["productName"] != "IBM WebSphere Application Server Network Deployment" {
				t.Errorf("Incorrect productName for edition ND: %s\n", pts.Annotations["productName"])
			}
		default:
			t.Errorf("Unexpected test data for edition %s\n", s.edition)
		}
		// 'cloudPak*' annotations should _not_ be present if the entitlement is standalone
		if s.pes == webspherelibertyv1.LicenseEntitlementStandalone {
			if _, exists := pts.Annotations["productCloudpakRatio"]; exists {
				t.Errorf("productCloudpakRatio should not exist but was %s\n", pts.Annotations["productCloudpakRatio"])
			}
			if _, exists := pts.Annotations["cloudpakName"]; exists {
				t.Errorf("cloudpakName should not exist but was %s\n", pts.Annotations["cloudpakName"])
			}
			if _, exists := pts.Annotations["cloudpakId"]; exists {
				t.Errorf("cloudpakId should not exist but was %s\n", pts.Annotations["cloudpakId"])
			}
		} else {
			if !checkRatio(pts.Annotations["productCloudpakRatio"], s.edition) {
				t.Errorf("Unexpected productCloudpakRatio %s for edition %s\n", pts.Annotations["productCloudpakRatio"], s.edition)
			}
			if !checkID(pts.Annotations["cloudpakId"], s.pes) {
				t.Errorf("Unexpected cloudpakId %s for entitlement %s\n", pts.Annotations["cloudpakId"], s.pes)
			}
			if !checkName(pts.Annotations["cloudpakName"], s.pes) {
				t.Errorf("Unexpected cloudpakName %s for entitlement %s\n", pts.Annotations["cloudpakName"], s.pes)
			}
		}

	}

	// Test where annotations should be skipped
	s := licenseTestData{edition: webspherelibertyv1.LicenseEditionND, pes: webspherelibertyv1.LicenseEntitlementWSHE}
	t.Logf("Testing %#v\n", s)
	spec := webspherelibertyv1.WebSphereLibertyApplicationSpec{}
	spec.License.Metric = s.metric
	spec.License.ProductEntitlementSource = s.pes
	spec.License.Edition = s.edition
	app := createWebSphereLibertyApp("myapp", "myns", spec)
	app.ObjectMeta.Annotations = map[string]string{excludeLicenseAnnotationsKey: "true"}
	pts := &corev1.PodTemplateSpec{}
	pts.Annotations = make(map[string]string)
	CustomizeLicenseAnnotations(pts, app)
	if (pts.Annotations[productIDKey] != "") || (pts.Annotations[productChargedContainersKey] != "") || (pts.Annotations[productMetricKey] != "") || (pts.Annotations[productNameKey] != "") {
		t.Errorf("License annotations should not be set when %s is set: '%s', '%s', '%s', '%s'", excludeLicenseAnnotationsKey,
			pts.Annotations[productIDKey], pts.Annotations[productChargedContainersKey], pts.Annotations[productMetricKey], pts.Annotations[productNameKey])
	}

	// Test that annotations are skipped, even where they where previously set
	pts.Annotations[productIDKey] = editionProductID[webspherelibertyv1.LicenseEditionBase]
	pts.Annotations[productChargedContainersKey] = "app"
	pts.Annotations[productMetricKey] = "PROCESSOR_VALUE_UNIT"
	pts.Annotations[productNameKey] = "random-product-name" // This doesn't need to be correct, just checking it gets removed
	pts.Annotations[cloudPakNameKey] = "random-name"
	pts.Annotations[cloudPakRatioKey] = "4:1"
	pts.Annotations[cloudPakIdKey] = "random-pak-id"
	CustomizeLicenseAnnotations(pts, app)
	if (pts.Annotations[productIDKey] != "") || (pts.Annotations[productChargedContainersKey] != "") || (pts.Annotations[productMetricKey] != "") || (pts.Annotations[productNameKey] != "") {
		t.Errorf("License annotations should not be set when %s is set: '%s', '%s', '%s', '%s'", excludeLicenseAnnotationsKey,
			pts.Annotations[productIDKey], pts.Annotations[productChargedContainersKey], pts.Annotations[productMetricKey], pts.Annotations[productNameKey])
	}
	if (pts.Annotations[cloudPakNameKey] != "") || (pts.Annotations[cloudPakRatioKey] != "") || (pts.Annotations[cloudPakIdKey] != "") {
		t.Errorf("Cloud pak annotations should not be set when %s is set: '%s', '%s', '%s'", excludeLicenseAnnotationsKey,
			pts.Annotations[cloudPakNameKey], pts.Annotations[cloudPakRatioKey], pts.Annotations[cloudPakIdKey])
	}

}

func checkRatio(ratio string, edition webspherelibertyv1.LicenseEdition) bool {
	if ratio == "4:1" && edition == webspherelibertyv1.LicenseEditionBase {
		return true
	}
	if ratio == "8:1" && edition == webspherelibertyv1.LicenseEditionCore {
		return true
	}
	if ratio == "1:1" && edition == webspherelibertyv1.LicenseEditionND {
		return true
	}
	return false
}
func checkID(id string, pes webspherelibertyv1.LicenseEntitlement) bool {
	if id == "4df52d2cdc374ba09f631a650ad2b5bf" && pes == webspherelibertyv1.LicenseEntitlementCP4Apps {
		return true
	}
	if id == "be8ae84b3dd04d81b90af0d846849182" && pes == webspherelibertyv1.LicenseEntitlementFamilyEdition {
		return true
	}
	if id == "6358611af04743f99f42dadcd6e39d52" && pes == webspherelibertyv1.LicenseEntitlementWSHE {
		return true
	}
	return false
}
func checkName(name string, pes webspherelibertyv1.LicenseEntitlement) bool {
	if name == string(pes) {
		return true
	}
	return false
}

// Helper Functions
func envSliceToMap(env []corev1.EnvVar, data map[string][]byte, t *testing.T) map[string]string {
	out := map[string]string{}
	for _, el := range env {
		if el.ValueFrom != nil {
			val := data[el.ValueFrom.SecretKeyRef.Key]
			out[el.Name] = "" + string(val)
		} else {
			out[el.Name] = string(el.Value)
		}
	}
	return out
}
func createWebSphereLibertyApp(n, ns string, spec webspherelibertyv1.WebSphereLibertyApplicationSpec) *webspherelibertyv1.WebSphereLibertyApplication {
	app := &webspherelibertyv1.WebSphereLibertyApplication{
		ObjectMeta: metav1.ObjectMeta{Name: n, Namespace: ns},
		Spec:       spec,
	}
	return app
}

func createFakeDiscoveryClient() discovery.DiscoveryInterface {
	fakeDiscoveryClient := &fakediscovery.FakeDiscovery{Fake: &coretesting.Fake{}}
	fakeDiscoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: routev1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "routes", Namespaced: true, Kind: "Route"},
			},
		},
		{
			GroupVersion: servingv1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "services", Namespaced: true, Kind: "Service", SingularName: "service"},
			},
		},
	}

	return fakeDiscoveryClient
}

func createReconcileRequest(n, ns string) reconcile.Request {
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{Name: n, Namespace: ns},
	}
	return req
}

// verifyReconcile checks that there was no error and that the reconcile is valid
func verifyReconcile(res reconcile.Result, err error) error {
	if err != nil {
		return fmt.Errorf("reconcile: (%v)", err)
	}

	if res != (reconcile.Result{}) {
		return fmt.Errorf("reconcile did not return an empty result (%v)", res)
	}

	return nil
}

func verifyTests(tests []Test) error {
	for _, tt := range tests {
		if !reflect.DeepEqual(tt.actual, tt.expected) {
			return fmt.Errorf("%s test expected: (%v) actual: (%v)", tt.test, tt.expected, tt.actual)
		}
	}
	return nil
}
