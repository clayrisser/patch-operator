/**
 * File: /patch_controller.go
 * Project: controllers
 * File Created: 16-10-2021 12:21:20
 * Author: Clay Risser
 * -----
 * Last Modified: 25-06-2023 14:03:21
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

package controllers

import (
	"context"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	patchv1alpha1 "gitlab.com/bitspur/rock8s/patch-operator/api/v1alpha1"
	"gitlab.com/bitspur/rock8s/patch-operator/util"
)

// PatchReconciler reconciles a Patch object
type PatchReconciler struct {
	Scheme *runtime.Scheme
	client.Client
}

//+kubebuilder:rbac:groups=patch.rock8s.com,resources=patches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=patch.rock8s.com,resources=patches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=patch.rock8s.com,resources=patches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Patch object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *PatchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx, "patch", req.NamespacedName)
	log.Log.Info("RECONCILING PATCH")
	patchUtil := util.NewPatchUtil(&r.Client, &ctx, &req, r.Scheme, log.Log,
		&patchv1alpha1.NamespacedName{
			Name:      req.NamespacedName.Name,
			Namespace: req.NamespacedName.Namespace,
		}, util.GlobalPatchMutex,
	)
	patch, err := patchUtil.Get()
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if patchUtil.FinalizeProbe(patch) {
		return patchUtil.Finalize(patch)
	}

	if patchUtil.InitializeFinalizerProbe(patch) {
		return patchUtil.InitializeFinalizer(patch)
	}

	pause, err := patchUtil.PauseProbe(patch)
	if err != nil {
		return patchUtil.Error((err))
	}
	if pause {
		return patchUtil.Pause(patch)
	}

	if patchUtil.PatchingProbe(patch) {
		return patchUtil.Patching(patch)
	}

	if patchUtil.PatchedProbe(patch) {
		return patchUtil.Patched(patch)
	}

	recalibrate, err := patchUtil.RecalibrateProbe(patch)
	if err != nil {
		return patchUtil.Error(err)
	}
	if recalibrate {
		return patchUtil.Recalibrate(patch)
	}

	return ctrl.Result{}, nil
}

func filterPatchPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// return e.ObjectNew.GetGeneration() > e.ObjectOld.GetGeneration()
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return !e.DeleteStateUnknown
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PatchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	maxConcurrentReconciles := 3
	if value := os.Getenv("MAX_CONCURRENT_RECONCILES"); value != "" {
		if val, err := strconv.Atoi(value); err == nil {
			maxConcurrentReconciles = val
		}
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&patchv1alpha1.Patch{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: maxConcurrentReconciles}).
		WithEventFilter(filterPatchPredicate()).
		Complete(r)
}
