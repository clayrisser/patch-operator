/**
 * File: /shared_types.go
 * Project: v1alpha1
 * File Created: 16-10-2021 13:47:31
 * Author: Clay Risser
 * -----
 *Last Modified: Su-11-2023 06:44:40
 *Modified By: Clay Risser
 * -----
 * BitSpur (c) Copyright 2021
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

package v1alpha1

// Target locates a resource
type Target struct {
	Kind string `json:"kind"`

	Name string `json:"name"`

	Namespace string `json:"namespace,omitempty"`

	Group string `json:"group,omitempty"`

	Version string `json:"version,omitempty"`

	ApiVersion string `json:"apiVersion,omitempty"`
}

type NamespacedName struct {
	// name
	Name string `json:"name"`

	// namespace
	Namespace string `json:"namespace,omitempty"`
}

type Phase string

const (
	FailedPhase    Phase = "Failed"
	PendingPhase   Phase = "Pending"
	ReadyPhase     Phase = "Ready"
	SucceededPhase Phase = "Succeeded"
	UnknownPhase   Phase = "Unknown"
)
