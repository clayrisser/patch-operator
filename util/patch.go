/**
 * File: /patch.go
 * Project: util
 * File Created: 16-10-2021 22:37:55
 * Author: Clay Risser
 * -----
 * Last Modified: 17-10-2021 19:59:39
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
	"context"
	"fmt"
	"strings"
	"sync"

	patchv1alpha1 "gitlab.com/bitspur/community/patch-operator/api/v1alpha1"
	"gitlab.com/bitspur/community/patch-operator/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PatchUtil struct {
	client         *client.Client
	ctx            *context.Context
	kubectlUtil    *KubectlUtil
	log            *log.DelegatingLogger
	mutex          *sync.Mutex
	namespacedName types.NamespacedName
	req            *ctrl.Request
}

func NewPatchUtil(
	client *client.Client,
	ctx *context.Context,
	req *ctrl.Request,
	log *log.DelegatingLogger,
	namespacedName *patchv1alpha1.NamespacedName,
	mutex *sync.Mutex,
) *PatchUtil {
	operatorNamespace := GetOperatorNamespace()
	if mutex == nil {
		mutex = &sync.Mutex{}
	}
	return &PatchUtil{
		client:         client,
		ctx:            ctx,
		kubectlUtil:    NewKubectlUtil(ctx),
		log:            log,
		mutex:          mutex,
		namespacedName: EnsureNamespacedName(namespacedName, operatorNamespace),
		req:            req,
	}
}

func (u *PatchUtil) InitializeFinalizerProbe(patch *patchv1alpha1.Patch) bool {
	return !controllerutil.ContainsFinalizer(patch, patchv1alpha1.PatchFinalizer)
}

func (u *PatchUtil) InitializeFinalizer(patch *patchv1alpha1.Patch) (ctrl.Result, error) {
	controllerutil.AddFinalizer(patch, patchv1alpha1.PatchFinalizer)
	if err := u.update(patch); err != nil {
		return u.Error(err)
	}
	return ctrl.Result{}, nil
}

func (u *PatchUtil) PatchingProbe(patch *patchv1alpha1.Patch) bool {
	return (!u.getConditionStatus(patch, PatchPatching) && !u.getConditionStatus(patch, PatchPatched))
}

func (u *PatchUtil) Patching(patch *patchv1alpha1.Patch) (ctrl.Result, error) {
	scriptUtil := NewScriptUtil(patch)
	for i, patchItem := range patch.Spec.Patches {
		patchId := patchItem.Id
		if patchItem.Id == "" {
			patchId = fmt.Sprint(i)
		}
		scriptUtil.AppendPatch(patchId, &patchItem)
	}
	jobUtil := NewJobUtil(patch, u.ctx)
	jobUtil.Create(scriptUtil.Get(), &[]v1.EnvVar{})
	return u.UpdateStatusPatching()
}

func (u *PatchUtil) PatchedProbe(patch *patchv1alpha1.Patch) bool {
	return !u.getConditionStatus(patch, PatchPatched)
}

func (u *PatchUtil) Patched(patch *patchv1alpha1.Patch) (ctrl.Result, error) {
	jobUtil := NewJobUtil(patch, u.ctx)
	completed, err := jobUtil.Completed()
	if err != nil {
		return u.Error(err)
	}
	if !completed {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: config.DefaultRequeueAfter,
		}, nil
	}
	return u.UpdateStatusPatched()
}

func (u *PatchUtil) FinalizeProbe(patch *patchv1alpha1.Patch) bool {
	return patch.GetDeletionTimestamp() != nil
}

func (u *PatchUtil) Finalize(patch *patchv1alpha1.Patch) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(patch, patchv1alpha1.PatchFinalizer) {
		jobUtil := NewJobUtil(patch, u.ctx)
		if err := jobUtil.Delete(); err != nil {
			return u.Error(err)
		}
		controllerutil.RemoveFinalizer(patch, patchv1alpha1.PatchFinalizer)
		if err := u.update(patch); err != nil {
			return u.Error(err)
		}
	}
	return ctrl.Result{}, nil
}

func (u *PatchUtil) Get() (*patchv1alpha1.Patch, error) {
	client := *u.client
	ctx := *u.ctx
	patch := &patchv1alpha1.Patch{}
	if err := client.Get(ctx, u.namespacedName, patch); err != nil {
		return nil, err
	}
	return patch.DeepCopy(), nil
}

func (u *PatchUtil) Error(err error) (ctrl.Result, error) {
	patch, _err := u.Get()
	if _err != nil {
		u.log.Error(nil, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: config.MaxRequeueDuration,
		}, _err
	}
	requeueAfter := CalculateExponentialRequireAfter(
		patch.Status.LastUpdate,
		2,
	)
	u.log.Error(nil, err.Error())
	if strings.Index(err.Error(), registry.OptimisticLockErrorMsg) <= -1 {
		if _err := u.updateErrorStatus(err); _err != nil {
			if strings.Contains(_err.Error(), registry.OptimisticLockErrorMsg) {
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: requeueAfter,
				}, nil
			}
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: requeueAfter,
			}, _err
		}
	}
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: requeueAfter,
	}, nil
}

func (u *PatchUtil) UpdateStatus(
	phase patchv1alpha1.Phase,
	patchConditionType *PatchConditionType,
) (ctrl.Result, error) {
	patch, err := u.Get()
	if err != nil {
		return u.Error(err)
	}
	if phase != "" {
		u.setPhaseStatus(patch, phase)
	}
	if patchConditionType != nil {
		u.setCondition(patch, *patchConditionType, true, "")
	}
	if err := u.updateStatus(patch, false); err != nil {
		return u.Error(err)
	}
	return ctrl.Result{Requeue: true}, nil
}

func (u *PatchUtil) UpdateStatusPatching() (ctrl.Result, error) {
	patchConditionType := PatchPatching
	return u.UpdateStatus(patchv1alpha1.PendingPhase, &patchConditionType)
}

func (u *PatchUtil) UpdateStatusPatched() (ctrl.Result, error) {
	patchConditionType := PatchPatched
	return u.UpdateStatus(patchv1alpha1.SucceededPhase, &patchConditionType)
}

func (u *PatchUtil) updateErrorStatus(err error) error {
	patch, _err := u.Get()
	if _err != nil {
		return _err
	}
	u.setErrorStatus(patch, err)
	if _err := u.updateStatus(patch, true); _err != nil {
		return _err
	}
	return nil
}

func (u *PatchUtil) update(patch *patchv1alpha1.Patch) error {
	client := *u.client
	ctx := *u.ctx
	u.mutex.Lock()
	if err := client.Update(ctx, patch); err != nil {
		u.mutex.Unlock()
		return err
	}
	u.mutex.Unlock()
	return nil
}

func (u *PatchUtil) updateStatus(
	patch *patchv1alpha1.Patch,
	exponentialBackoff bool,
) error {
	client := *u.client
	ctx := *u.ctx
	if !exponentialBackoff ||
		patch.Status.LastUpdate.IsZero() ||
		config.StartTime.Unix() > patch.Status.LastUpdate.Unix() {
		patch.Status.LastUpdate = metav1.Now()
	}
	u.mutex.Lock()
	if err := client.Status().Update(ctx, patch); err != nil {
		u.mutex.Unlock()
		return err
	}
	u.mutex.Unlock()
	return nil
}

func (u *PatchUtil) getConditionStatus(patch *patchv1alpha1.Patch, patchConditionType PatchConditionType) bool {
	condition := u.getCondition(patch, patchConditionType)
	if condition == nil {
		return false
	}
	if condition.Status == "True" {
		return true
	}
	return false
}

func (u *PatchUtil) getCondition(patch *patchv1alpha1.Patch, patchConditionType PatchConditionType) *metav1.Condition {
	return meta.FindStatusCondition(patch.Status.Conditions, string(patchConditionType))
}

func (u *PatchUtil) setCondition(
	patch *patchv1alpha1.Patch,
	patchConditionType PatchConditionType,
	status bool,
	message string,
) {
	if message == "" {
		if patchConditionType == PatchPatched {
			message = "patch patched"
		} else if patchConditionType == PatchPatching {
			message = "patch patching"
		} else {
			message = "patch failed"
		}
	}
	condition := metav1.Condition{
		Message:            message,
		ObservedGeneration: patch.Generation,
		Status:             "False",
		Reason:             string(patchConditionType),
		Type:               string(patchConditionType),
	}
	if status {
		condition.Status = "True"
	}
	u.removeExceptCondition(patch, patchConditionType)
	meta.SetStatusCondition(&patch.Status.Conditions, condition)
}

func (u *PatchUtil) removeExceptCondition(
	patch *patchv1alpha1.Patch,
	patchConditionType PatchConditionType,
) {
	for _, conditionType := range patchConditionTypes {
		if conditionType != patchConditionType {
			meta.RemoveStatusCondition(&patch.Status.Conditions, string(conditionType))
		}
	}
}

func (u *PatchUtil) setPhaseStatus(
	patch *patchv1alpha1.Patch,
	phase patchv1alpha1.Phase,
) {
	if phase != patchv1alpha1.FailedPhase {
		patch.Status.Message = ""
	}
	patch.Status.Phase = phase
}

func (u *PatchUtil) setErrorStatus(patch *patchv1alpha1.Patch, err error) {
	message := err.Error()
	u.setCondition(patch, PatchFailed, true, message)
	patch.Status.Phase = patchv1alpha1.FailedPhase
	patch.Status.Message = message
}

var GlobalPatchMutex *sync.Mutex = &sync.Mutex{}

type PatchConditionType string

const (
	PatchFailed   PatchConditionType = "Failed"
	PatchPatched  PatchConditionType = "Patched"
	PatchPatching PatchConditionType = "Patching"
)

var patchConditionTypes []PatchConditionType = []PatchConditionType{PatchFailed, PatchPatched, PatchPatching}
