/**
 * File: /patch_types.go
 * Project: v1alpha1
 * File Created: 16-10-2021 12:21:20
 * Author: Clay Risser
 * -----
 * Last Modified: 18-10-2021 21:17:29
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
	"gitlab.com/bitspur/community/patch-operator/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PatchFinalizer = config.PatchGroup + "." + config.Domain + "/finalizer"

type PatchType string

const (
	JsonPatchType      PatchType = "json"
	MergePatchType     PatchType = "merge"
	ScriptPatchType    PatchType = "script"
	StrategicPatchType PatchType = "strategic"
)

// the desired state of the patch
type PatchSpec struct {
	// a list of patches to be applied in order
	Patches []PatchSpecPatch `json:"patches,omitempty"`

	// image used in the job
	Image string `json:"image,omitempty"`

	// change epoch to force recalibration
	Epoch string `json:"epoch,omitempty"`
}

// PatchStatus defines the observed state of Patch
type PatchStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// integration plug phase (Pending, Succeeded, Failed, Unknown)
	Phase Phase `json:"phase,omitempty"`

	// last update time
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`

	// status message
	Message string `json:"message,omitempty"`

	// spec hash
	SpecHash string `json:"specHash,omitempty"`

	// pause until update
	PauseUntilUpdate bool `json:"pauseUntilUpdate,omitempty"`
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

	// the resource to patch
	Target Target `json:"target"`

	// you can read more about the patch types at the following link
	// https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#use-a-json-merge-patch-to-update-a-deployment
	Type PatchType `json:"type,omitempty"`

	// wait for time in milliseconds before applying patch
	WaitForTimeout int `json:"waitForTimeout,omitempty"`

	// wait for the resource to exist
	WaitForResource bool `json:"waitForResource,omitempty"`

	// skip patch if criteria met
	SkipIf []PatchSpecPatchSkipIf `json:"skipIf,omitempty"`

	// optional patch id for reference
	Id string `json:"id,omitempty"`
}

type PatchSpecPatchSkipIf struct {
	// the target to check criteria against. if no target specified, the target
	// being patched will be used
	Target *Target `json:"target,omitempty"`

	// the json patch to check the criteria against
	JsonPath string `json:"jsonPath,omitempty"`

	// an extended grep compatible regular expression
	Regex string `json:"regex,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Patch{}, &PatchList{})
}
