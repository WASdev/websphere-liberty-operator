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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	gherrors "github.com/pkg/errors"
)

type RegisterData struct {
	DiscoveryURL            string
	RouteURL                string
	RedirectToRPHostAndPort string
	ProviderId              string
	Scopes                  string
	GrantTypes              string
	InitialAccessToken      []byte
	InitialClientId         []byte
	InitialClientSecret     []byte
	RegistrationURL         string
	InsecureTLS             bool
}

func RegisterWithOidcProvider(regData RegisterData) ([]byte, []byte, error) {
	return doRegister(regData)
}

// register with oidc provider and create a new client.  return the new client id and client secret, or an error.
func doRegister(rdata RegisterData) ([]byte, []byte, error) {
	// process:
	//  1) call the provider's discovery endpoint to find the token and registration urls.
	//  2) If we do not have an initial access token,
	//  2.5) Use supplied clientId and secret in a Client Credentials grant to obtain an access token.
	//  3) Use the access token to register and obtain a new client id and secret.

	registrationURL, tokenURL, err := getURLs(rdata.DiscoveryURL, rdata.InsecureTLS, rdata.ProviderId)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	if tokenURL == "" {
		return []byte{}, []byte{}, gherrors.New("Provider " + rdata.ProviderId + ": failed to obtain token endpoint from discovery endpoint.")
	}

	// ICI: if we don't have initial token, use client and secret to go get one.
	var token = rdata.InitialAccessToken
	if len(token) == 0 && (len(rdata.InitialClientId) == 0 || len(rdata.InitialClientSecret) == 0) {
		id := rdata.ProviderId
		return []byte{}, []byte{}, gherrors.New("Provider " + id + ": registration data for Single sign-on (SSO) is missing required fields," +
			" one or more of " + id + "-autoreg-initialAccessToken, " + id + "-autoreg-initialClientId, or " + id + "-autoreg-initialClientSecret.")
	}
	if len(token) == 0 {
		rtoken, err := requestAccessToken(rdata, tokenURL)
		if err != nil {
			return []byte{}, []byte{}, err
		}
		if len(rtoken) == 0 {
			return []byte{}, []byte{}, gherrors.New("Provider " + rdata.ProviderId + ": failed to obtain access token for registration.")
		}
		rdata.InitialAccessToken = rtoken
	}

	if rdata.RegistrationURL != "" {
		registrationURL = rdata.RegistrationURL
	}
	// registrationURL should be in discovery data but allow it to be supplied manually if not.
	if registrationURL == "" {
		return []byte{}, []byte{}, gherrors.New("Provider " + rdata.ProviderId + ": failed to obtain registration URL - specify registrationURL in registration data secret.")
	}

	registrationRequestJson := buildRegistrationRequestJson(rdata)

	registrationResponse, err := sendHTTPRequest(registrationRequestJson, registrationURL, "POST", []byte{}, rdata.InitialAccessToken, rdata.InsecureTLS, rdata.ProviderId)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	// extract id and secret from body
	id, secret, err := parseRegistrationResponseJson(registrationResponse, rdata.ProviderId)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return id, secret, nil
}

func requestAccessToken(rdata RegisterData, tokenURL string) ([]byte, error) {
	tokenRequestContent := "grant_type=client_credentials&scope=" + getScopes(rdata)
	tokenResponse, err := sendHTTPRequest(tokenRequestContent, tokenURL, "POST", rdata.InitialClientId, rdata.InitialClientSecret, rdata.InsecureTLS, rdata.ProviderId)
	if err != nil {
		return []byte{}, err
	}
	token, err := parseTokenResponse(tokenResponse, rdata.ProviderId)
	if err != nil {
		return []byte{}, err
	}
	return token, nil
}

// parse token response and return token
func parseTokenResponse(respJson []byte, providerId string) ([]byte, error) {
	type token struct {
		Access_token json.RawMessage
	}
	var cdata token
	err := json.Unmarshal(respJson, &cdata)
	if err != nil {
		return []byte{}, errors.New("Provider " + string(providerId) + ": error parsing token response: " + err.Error() + " Data: " + string(respJson))
	}
	return cdata.Access_token, nil
}

// parse the response and return the client id and client secret
func parseRegistrationResponseJson(respJson []byte, providerId string) ([]byte, []byte, error) {
	type idsecret struct {
		Client_id     json.RawMessage
		Client_secret json.RawMessage
	}

	var cdata idsecret
	err := json.Unmarshal([]byte(respJson), &cdata)
	if err != nil {
		return []byte{}, []byte{}, errors.New("Provider " + providerId + ": error parsing registration response: " + err.Error() + " Data: " + string(respJson))
	}
	return cdata.Client_id, cdata.Client_secret, nil
}

// build the JSON for the client registration request. Form the redirectURL from the route URL.
func buildRegistrationRequestJson(rdata RegisterData) string {
	now := time.Now()
	sysClockMillisec := now.UnixNano() / 1000000
	//	rhsso will not accept a supplied value for client_id, so leave a comment in the name
	clientName := "LibertyOperator-" + strings.Replace(rdata.RouteURL, "https://", "", 1) + "-" +
		strconv.FormatInt(sysClockMillisec, 10)

	// IBM Security Verify needs some special things in the request.
	isvAttribs := ""
	if len(rdata.InitialClientId) > 0 {
		isvAttribs = "\"enforce_pkce\":false," +
			"\"all_users_entitled\":true," +
			"\"consent_action\":\"never_prompt\","
	}

	return "{" + isvAttribs +
		"\"client_name\":\"" + clientName + "\"," +
		"\"grant_types\":[" + getGrantTypes(rdata) + "]," +
		"\"scope\":\"" + getScopes(rdata) + "\"," +
		"\"redirect_uris\":[\"" + getRedirectUri(rdata) + "\"]}"
}

func getScopes(rdata RegisterData) string {
	if rdata.Scopes == "" {
		return "openid profile"
	}

	var result = ""
	gts := strings.Split(rdata.Scopes, ",")
	for _, gt := range gts {
		result += strings.Trim(gt, " ") + " "
	}
	return strings.TrimSuffix(result, " ")
}

func getGrantTypes(rdata RegisterData) string {
	if rdata.GrantTypes == "" {
		return "\"authorization_code\",\"refresh_token\""
	}

	var result = ""
	gts := strings.Split(rdata.GrantTypes, ",")
	for _, gt := range gts {
		result += "\"" + strings.Trim(gt, " ") + "\"" + ","
	}
	return strings.TrimSuffix(result, ",")
}

func getRedirectUri(rdata RegisterData) string {
	providerId := rdata.ProviderId
	if providerId == "" {
		providerId = "oidc"
	}
	suffix := "/ibm/api/social-login/redirect/" + providerId
	if rdata.RedirectToRPHostAndPort != "" {
		return rdata.RedirectToRPHostAndPort + suffix
	}
	return rdata.RouteURL + suffix
}

// retrieve the registration and token URLs from the provider's discovery URL.
// return an error if we don't get back two valid url's.
// todo: more error checking needed to make that true?
func getURLs(discoveryURL string, insecureTLS bool, providerId string) (string, string, error) {
	discoveryResult, err := sendHTTPRequest("", discoveryURL, "GET", []byte{}, []byte{}, insecureTLS, providerId)
	if err != nil {
		return "", "", err
	}

	type regEp struct {
		Registration_endpoint string
	}

	type tokenEp struct {
		Token_endpoint string
	}

	var regdata regEp
	var tokendata tokenEp
	err = json.Unmarshal(discoveryResult, &regdata)
	if err != nil {
		return "", "", errors.New("Provider " + string(providerId) + ": error unmarshalling data from discovery endpoint: " + err.Error() + " Data: " + string(discoveryResult))
	}
	err = json.Unmarshal(discoveryResult, &tokendata)
	if err != nil {
		return "", "", errors.New("Provider " + string(providerId) + ": error unmarshalling data from discovery endpoint: " + err.Error() + " Data: " + string(discoveryResult))
	}

	return regdata.Registration_endpoint, tokendata.Token_endpoint, nil
}

// Send an http(s)  request.  return response body and error.
// content to send can be an empty string. Json will be detected. Method should be GET or POST.
// if id is set, send id and passwordOrToken as basic auth header, otherwise send token as bearer auth header.
// If error occurs, body will be "error".
func sendHTTPRequest(content string, URL string, method string, id []byte, passwordOrToken []byte, insecureTLS bool, providerId string) ([]byte, error) {
	rootCAPool, _ := x509.SystemCertPool()
	if rootCAPool == nil {
		rootCAPool = x509.NewCertPool()
	}

	if !insecureTLS {
		cert, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt")
		if err != nil {
			return []byte{}, errors.New("Error reading TLS certificates: " + err.Error())
		}
		rootCAPool.AppendCertsFromPEM(cert)
	}

	client := &http.Client{
		Timeout: time.Second * 20,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            rootCAPool,
				InsecureSkipVerify: insecureTLS,
			},
		},
	}

	var requestBody = []byte(content)

	request, err := http.NewRequest(method, URL, bytes.NewBuffer(requestBody))
	if strings.HasPrefix(content, "{") {
		request.Header.Set("Content-type", "application/json")
		request.Header.Set("Accept", "application/json")
	} else {
		request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	}

	if len(id) > 0 {
		request.SetBasicAuth(string(id), string(passwordOrToken))
	} else {
		if len(passwordOrToken) > 0 {
			request.Header.Set("Authorization", "Bearer "+string(passwordOrToken))
		}
	}

	errorStr := []byte("error")
	var errorMsgPreamble = "Provider " + providerId + ": error occurred communicating with OIDC provider.  URL: " + URL + ": "
	if err != nil {
		return errorStr, errors.New(errorMsgPreamble + err.Error())
	}

	response, err := client.Do(request)
	if response == nil {
		return errorStr, errors.New(errorMsgPreamble + err.Error()) // bad hostname, can't connect, etc.
	}
	defer response.Body.Close()

	if err != nil {
		return errorStr, errors.New(errorMsgPreamble + err.Error()) // timeout, conn reset, etc.
	}

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return errorStr, errors.New(errorMsgPreamble + err.Error())
	}

	// a successful registration usually has a 201 response code.
	if response.StatusCode != 200 && response.StatusCode != 201 {
		return errorStr, errors.New(errorMsgPreamble + response.Status + ". " + string(respBytes) + ". Data sent was: " + content)
	}
	return respBytes, nil
}
