/**
 * File: /patch_types.go
 * Project: v1alpha1
 * File Created: 16-10-2021 12:21:20
 * Author: Clay Risser
 * -----
 * Last Modified: 16-10-2021 14:12:55
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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PatchType string

const (
	JsonPatchType      PatchType = "json"
	MergePatchType     PatchType = "merge"
	StrategicPatchType PatchType = "strategic"
)

// the desired state of the patch
type PatchSpec struct {
	// a list of patches to be applied in order
	Patches []PatchSpecPatch `json:"patches,omitempty"`
}

// PatchStatus defines the observed state of Patch
type PatchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Patch is the Schema for the patches API
type Patch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PatchSpec   `json:"spec,omitempty"`
	Status PatchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PatchList contains a list of Patch
type PatchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Patch `json:"items"`
}

// you can read more about kubernetes patches at the following link
// https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch
type PatchSpecPatch struct {
	// the patch to apply
	Patch string `json:"patch"`

	// you can read more about the patch types at the following link
	// https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#use-a-json-merge-patch-to-update-a-deployment
	Type PatchType `json:"type,omitempty"`

	// wait for time in milliseconds before applying patch
	WaitForTimeout int `json:"waitForTimeout,omitempty"`

	// wait for the resource to exist
	WaitForResource bool `json:"waitForResource,omitempty"`

	// apply patch if criteria met
	ApplyIf []PatchSpecPatchApplyIf `json:"applyIf,omitempty"`

	// the resource to patch
	Target Target `json:"target"`
}

type PatchSpecPatchApplyIf struct {
	// the target to check criteria against. if no target specified, the target
	// being patched will be used
	Target Target `json:"target,omitempty"`

	// the json patch to check the criteria against
	JsonPath string `json:"jsonPath,omitempty"`

	// an extended grep compatible regular expression
	Regex string `json:"regex,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Patch{}, &PatchList{})
}