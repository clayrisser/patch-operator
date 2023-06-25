/**
 * File: /main.go
 * Project: config
 * File Created: 25-06-2023 13:11:39
 * Author: Clay Risser
 * -----
 * Last Modified: 25-06-2023 14:03:33
 * Modified By: Clay Risser
 * -----
 * BitSpur Inc (c) Copyright 2021 - 2023
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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MaxRequeueDuration time.Duration = time.Duration(float64(time.Hour.Nanoseconds() * 6))

var StartTime metav1.Time = metav1.Now()

const DefaultRequeueAfter = time.Duration(time.Second * 15)

const PatchGroup = "patch"

const Domain = "rock8s.com"
