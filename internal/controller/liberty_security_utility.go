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
	"fmt"
	"os/exec"

	"github.com/WASdev/websphere-liberty-operator/utils"
)

const SECURITY_UTILITY_BINARY = "liberty/bin/securityUtility"
const SECURITY_UTILITY_ENCODE = "encode"
const SECURITY_UTILITY_CREATE_LTPA_KEYS = "createLTPAKeys"
const SECURITY_UTILITY_OUTPUT_FOLDER = "liberty/output"

func encode(password string, passwordKey *string) ([]byte, error) {
	params := []string{}
	params = append(params, SECURITY_UTILITY_ENCODE)
	params = append(params, fmt.Sprintf("--encoding=%s", "aes-128"))
	if passwordKey != nil && len(*passwordKey) > 0 {
		params = append(params, fmt.Sprintf("--key=%s", *passwordKey))
	}
	params = append(params, password)
	return callSecurityUtility(params)
}

func createLTPAKeys(password string, passwordKey *string) ([]byte, error) {
	tmpFileName := fmt.Sprintf("ltpa-keys-%s.keys", utils.GetRandomAlphanumeric(15))
	tmpFilePath := fmt.Sprintf("%s/%s", SECURITY_UTILITY_OUTPUT_FOLDER, tmpFileName)

	// delete possible colliding file
	callDeleteFile(tmpFilePath)

	// mkdir if not exists
	// callMkdir(SECURITY_UTILITY_OUTPUT_FOLDER)

	// create the key
	params := []string{}
	params = append(params, SECURITY_UTILITY_CREATE_LTPA_KEYS)
	params = append(params, fmt.Sprintf("--file=%s", tmpFilePath))
	params = append(params, fmt.Sprintf("--passwordEncoding=%s", "aes-128")) // use aes encoding
	if passwordKey != nil && len(*passwordKey) > 0 {
		params = append(params, fmt.Sprintf("--passwordKey=%s", *passwordKey))
	}
	params = append(params, fmt.Sprintf("--password=%s", password))
	callSecurityUtility(params)

	// read the key
	params = []string{}
	params = append(params, "-c")
	params = append(params, fmt.Sprintf("cat %s | base64", tmpFilePath))
	bytesOut, err := callCommand("/bin/bash", params)

	// delete the key
	callDeleteFile(tmpFilePath)
	return bytesOut, err
}

// func callMkdir(folderPath string) {
// 	params := []string{}
// 	params = append(params, "-c")
// 	params = append(params, fmt.Sprintf("mkdir -p %s", folderPath))
// 	callCommand("/bin/bash", params)
// }

func callDeleteFile(filePath string) {
	params := []string{}
	params = append(params, "-c")
	params = append(params, fmt.Sprintf("rm -f %s", filePath))
	callCommand("/bin/bash", params)
}

func callSecurityUtility(params []string) ([]byte, error) {
	return callCommand(SECURITY_UTILITY_BINARY, params)
}

func callCommand(binary string, params []string) ([]byte, error) {
	cmd := exec.Command(binary, params...)
	stdout, err := cmd.Output()
	if err != nil {
		return []byte{}, err
	}
	return stdout, nil
}
