/**
 * File: /main.go
 * Project: config
 * File Created: 16-10-2021 22:44:20
 * Author: Clay Risser
 * -----
 * Last Modified: 17-10-2021 17:43:22
 * Modified By: Clay Risser
 * -----
 * BitSpur Inc (c) Copyright 2021
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MaxRequeueDuration time.Duration = time.Duration(float64(time.Hour.Nanoseconds() * 6))

var StartTime metav1.Time = metav1.Now()

var DebugPatchEndpoint = os.Getenv("DEBUG_PATCH_ENDPOINT")

var DefaultRequeueAfter = time.Duration(time.Second * 15)
