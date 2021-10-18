/**
 * File: /script.go
 * Project: util
 * File Created: 17-10-2021 19:01:54
 * Author: Clay Risser
 * -----
 * Last Modified: 17-10-2021 20:59:00
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

package util

import (
	"errors"
	"fmt"

	"gitlab.com/bitspur/community/patch-operator/api/v1alpha1"
	patchv1alpha1 "gitlab.com/bitspur/community/patch-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ScriptUtil struct {
	script string
	patch  *patchv1alpha1.Patch
}

func NewScriptUtil(patch *patchv1alpha1.Patch) *ScriptUtil {
	return &ScriptUtil{
		script: `##### initialization #####
echo ===== initializing =====
echo ----- command -----
echo mkdir -p /tmp/patches
echo ----- output -----
mkdir -p /tmp/patches
echo -e "===== done initializing =====\n\n\n"



`,
		patch: patch,
	}
}

func (s *ScriptUtil) AppendPatch(patchId string, patchItem *patchv1alpha1.PatchSpecPatch) error {
	resource, err := s.targetToResource(patchId, s.patch, &patchItem.Target)
	if err != nil {
		return err
	}
	script := fmt.Sprintf(`##### patch %s #####
echo ===== applying patch %s =====
`, patchId, patchId)
	if patchItem.Type == patchv1alpha1.ScriptPatchType {
		script += "\n" + patchItem.Patch
	} else {
		patchType := ""
		if patchItem.Type != "" {
			patchType = " --type " + string(patchItem.Type)
		}
		script += fmt.Sprintf(`echo ----- command -----
echo kubectl cat \<\<EOF \| patch%s --patch-file /tmp/patches/%s.yaml
cat <<EOF
%s
EOF
echo EOF
echo ----- output -----
cat <<EOF > /tmp/patches/%s.yaml
%s
EOF
cat <<EOF | kubectl patch -f -%s --patch-file /tmp/patches/%s.yaml
apiVersion: %s
kind: %s
metadata:
  name: %s
  namespace: %s
EOF
[ "$(echo $?)" = "0" ] || exit $?
`, patchType, patchId, patchItem.Patch, patchId, patchItem.Patch, patchType, patchId,
			resource.GetAPIVersion(),
			resource.GetKind(),
			resource.GetName(),
			resource.GetNamespace(),
		)
	}
	s.script = s.script + script + fmt.Sprintf(`echo -e "===== done applying patch %s =====\n\n\n"


`, patchId)
	return nil
}

func (s *ScriptUtil) Get() string {
	return s.script
}

func (s *ScriptUtil) targetToResource(patchId string, patch *v1alpha1.Patch, target *v1alpha1.Target) (*unstructured.Unstructured, error) {
	resource := unstructured.Unstructured{}
	apiVersion := target.ApiVersion
	if apiVersion == "" && target.Version != "" {
		if target.Group != "" {
			apiVersion = target.Group + "/"
		}
		apiVersion += target.Version
	}
	if apiVersion == "" {
		return nil, errors.New(fmt.Sprintf("target.apiVersion must be set for patch %s", patchId))
	}
	kind := target.Kind
	name := target.Name
	namespace := target.Namespace
	if namespace == "" {
		namespace = patch.Namespace
	}
	resource.SetAPIVersion(apiVersion)
	resource.SetKind(kind)
	resource.SetName(name)
	resource.SetNamespace(namespace)
	return &resource, nil
}
