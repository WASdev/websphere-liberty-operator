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
	"sync"
)

var LeaderTrackerMutexes *sync.Map

func init() {
	LeaderTrackerMutexes = &sync.Map{}
}

// Leader tracking constants
const ResourcesKey = "names"
const ResourceOwnersKey = "owners"
const ResourcePathsKey = "paths"
const ResourcePathIndicesKey = "pathIndices"

// const ResourceSubleasesKey = "subleases"

const LibertyURI = "webspherelibertyapplications.liberty.websphere.ibm.com"
const LeaderVersionLabel = LibertyURI + "/leader-version"
const ResourcePathIndexLabel = LibertyURI + "/resource-path-index"

func GetLastRotationLabelKey(sharedResourceName string) string {
	return LibertyURI + "/" + sharedResourceName + "-last-rotation"
}

func GetTrackedResourceName(sharedResourceName string) string {
	return kebabToCamelCase(sharedResourceName) + "TrackedResourceName"
}
