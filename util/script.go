/**
 * File: /script.go
 * Project: util
 * File Created: 26-11-2023 06:42:14
 * Author: Clay Risser
 * -----
 * BitSpur (c) Copyright 2021 - 2023
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
	"strings"

	"gitlab.com/bitspur/rock8s/patch-operator/api/v1alpha1"
	patchv1alpha1 "gitlab.com/bitspur/rock8s/patch-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ScriptUtil struct {
	script string
	patch  *patchv1alpha1.Patch
}

func NewScriptUtil(patch *patchv1alpha1.Patch) *ScriptUtil {
	return &ScriptUtil{
		script: fmt.Sprintf(`##### initialization #####
echo ===== initializing =====
echo ----- command -----
echo 'kubectl get pods -n %s \'
echo '    -l job-name=%s \'
echo '    -l %s=%s \'
echo '    --field-selector status.phase=Succeeded \'
echo '    -o yaml | kubectl delete -f -'
echo mkdir -p /tmp/patches
echo ----- output -----
kubectl get pods -n %s \
    -l job-name=%s \
    -l %s=%s \
    --field-selector status.phase=Succeeded \
    -o yaml | kubectl delete -f -
mkdir -p /tmp/patches
echo -e "===== done initializing =====\n\n\n"



`, patch.GetNamespace(), patch.GetName(), PatchLabel, patch.GetName(),
			patch.GetNamespace(), patch.GetName(), PatchLabel, patch.GetName(),
		),
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
	commandPreview := "echo ----- command -----\n"
	commandExecute := "echo ----- output -----\n"
	if patchItem.WaitForTimeout > 0 {
		commandPreview += fmt.Sprintf("echo sleep %d\n", patchItem.WaitForTimeout)
		commandExecute += fmt.Sprintf("sleep %d\n", patchItem.WaitForTimeout)
	}
	if patchItem.WaitForResource {
		commandPreview += fmt.Sprintf(`echo '    unset STATUS'
echo 'while [ "$STATUS" != "0" ]; do'
echo 'cat <<EOF | kubectl get -f - >/dev/null'
cat <<EOF
apiVersion: %s
kind: %s
metadata:
  name: %s
  namespace: %s
EOF
echo EOF
echo '    export STATUS=$?'
echo '    [ "$STATUS" = "0" ] || sleep 5'
echo 'done'
`, resource.GetAPIVersion(), resource.GetKind(), resource.GetName(), resource.GetNamespace())
		commandExecute += fmt.Sprintf(`unset STATUS
while [ "$STATUS" != "0" ]; do
    cat <<EOF | kubectl get -f - >/dev/null
apiVersion: %s
kind: %s
metadata:
  name: %s
  namespace: %s
EOF
    export STATUS=$?
    [ "$STATUS" = "0" ] || sleep 5
done
`, resource.GetAPIVersion(), resource.GetKind(), resource.GetName(), resource.GetNamespace())
	}
	if patchItem.SkipIf != nil {
		for _, skipIf := range patchItem.SkipIf {
			target := skipIf.Target
			if target == nil {
				target = &patchItem.Target
			}
			skipIfResource, err := s.targetToResource(patchId, s.patch, target)
			if err != nil {
				return err
			}
			jsonPath := ".items[0]"
			if skipIf.JsonPath != "" && skipIf.JsonPath != "." {
				if !strings.HasPrefix(skipIf.JsonPath, ".") {
					jsonPath += "."
				}
				jsonPath += skipIf.JsonPath
			}
			jsonPath = fmt.Sprintf(" -o jsonpath='{%s}'", jsonPath)
			regex := ".*"
			if skipIf.Regex != "" {
				regex = skipIf.Regex
			}
			commandPreview += fmt.Sprintf(`echo export SKIP_PATCH=true
echo 'cat <<EOF | kubectl get -f -'"%s"' | grep -q -E "%s" || export SKIP_PATCH=false'
cat <<EOF
apiVersion: %s
kind: %s
metadata:
  name: %s
  namespace: %s
EOF
echo EOF
`, jsonPath, regex, skipIfResource.GetAPIVersion(), skipIfResource.GetKind(), skipIfResource.GetName(), skipIfResource.GetNamespace())
			commandExecute += fmt.Sprintf(`export SKIP_PATCH=true
cat <<EOF | kubectl get -f -%s | grep -q -E "%s" || export SKIP_PATCH=false
apiVersion: %s
kind: %s
metadata:
  name: %s
  namespace: %s
EOF
`, jsonPath, regex, skipIfResource.GetAPIVersion(), skipIfResource.GetKind(), skipIfResource.GetName(), skipIfResource.GetNamespace())
		}
	}
	if patchItem.Type == patchv1alpha1.ScriptPatchType {
		commandPreview += fmt.Sprintf(`echo 'if [ "$SKIP_PATCH" != "true" ]; then'
    cat <<EOF
%s
EOF
echo fi`, patchItem.Patch)
		commandExecute += fmt.Sprintf(`if [ "$SKIP_PATCH" != "true" ]; then
%s
else
    echo skipping patch %s
fi`, patchItem.Patch, patchId)
	} else {
		patchType := ""
		if patchItem.Type != "" {
			patchType = " --type " + string(patchItem.Type)
		}
		commandPreview += fmt.Sprintf(`echo 'if [ "$SKIP_PATCH" != "true" ]; then'
		echo '    cat <<EOF | kubectl patch%s --patch-file /tmp/patches/%s.yaml'
cat <<EOF
%s
EOF
echo EOF
echo fi`, patchType, patchId, patchItem.Patch)
		commandExecute += fmt.Sprintf(`if [ "$SKIP_PATCH" != "true" ]; then
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
else
    echo skipping patch %s
fi`, patchId, patchItem.Patch, patchType, patchId,
			resource.GetAPIVersion(),
			resource.GetKind(),
			resource.GetName(),
			resource.GetNamespace(),
			patchId,
		)
	}
	s.script = s.script + script + fmt.Sprintf(`%s
%s
echo -e "===== done applying patch %s =====\n\n\n"


`, commandPreview, commandExecute, patchId)

	return nil
}

func (s *ScriptUtil) Get() string {
	return s.script + fmt.Sprintf(`##### finalization #####
echo ===== finalizing =====
echo ----- command -----
echo 'kubectl get pods -n %s \'
echo '    -l job-name=%s \'
echo '    -l %s=%s \'
echo '    --field-selector status.phase=Failed \'
echo '    -o yaml | kubectl delete -f -'
echo ----- output -----
kubectl get pods -n %s \
    -l job-name=%s \
    -l %s=%s \
    --field-selector status.phase=Failed \
    -o yaml | kubectl delete -f -
echo -e "===== done finalizing ====="
`, s.patch.GetNamespace(), s.patch.GetName(), PatchLabel, s.patch.GetName(),
		s.patch.GetNamespace(), s.patch.GetName(), PatchLabel, s.patch.GetName())
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
		return nil, errors.New(fmt.Sprintf("apiVersion missing in patch %s", patchId))
	}
	kind := target.Kind
	name := target.Name
	namespace := target.Namespace
	if namespace == "" {
		namespace = patch.GetNamespace()
	}
	resource.SetAPIVersion(apiVersion)
	resource.SetKind(kind)
	resource.SetName(name)
	resource.SetNamespace(namespace)
	return &resource, nil
}
