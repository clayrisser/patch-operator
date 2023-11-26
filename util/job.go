/**
 * File: /job.go
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
	"context"

	patchv1alpha1 "gitlab.com/bitspur/rock8s/patch-operator/api/v1alpha1"
	"gitlab.com/bitspur/rock8s/patch-operator/config"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

const PatchLabel = config.PatchGroup + "." + config.Domain + "/patch"

type JobUtil struct {
	cfg       *rest.Config
	clientset *kubernetes.Clientset
	ctx       *context.Context
	patch     *patchv1alpha1.Patch
	scheme    *runtime.Scheme
}

func NewJobUtil(patch *patchv1alpha1.Patch, ctx *context.Context, scheme *runtime.Scheme) *JobUtil {
	cfg := ctrl.GetConfigOrDie()
	return &JobUtil{
		cfg:       cfg,
		clientset: kubernetes.NewForConfigOrDie(cfg),
		ctx:       ctx,
		patch:     patch,
		scheme:    scheme,
	}
}

func (j *JobUtil) Create(command string, env *[]v1.EnvVar) (*batchv1.Job, error) {
	if command == "" {
		command = "true"
	}
	jobs := j.clientset.BatchV1().Jobs(j.patch.GetNamespace())
	var backoffLimit int32 = 0
	serviceAccountName := j.patch.Spec.ServiceAccountName
	if serviceAccountName == "" {
		serviceAccountName = "default"
	}
	image := j.patch.Spec.Image
	if image == "" {
		image = "registry.gitlab.com/bitspur/rock8s/images/kube-commands:3.18.0"
	}
	labels := j.patch.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labels[PatchLabel] = j.patch.GetName()
	automountServiceAccountToken := true
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      j.patch.GetName() + "-patch",
			Namespace: j.patch.GetNamespace(),
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						PatchLabel: j.patch.GetName(),
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: v1.PodSpec{
					AutomountServiceAccountToken: &automountServiceAccountToken,
					RestartPolicy:                v1.RestartPolicyNever,
					ServiceAccountName:           serviceAccountName,
					Containers: []v1.Container{
						{
							Name:            "kubectl",
							Image:           image,
							ImagePullPolicy: v1.PullIfNotPresent,
							Command: []string{
								"/bin/sh",
								"-c",
								command,
							},
							Args: []string{},
							Env:  *env,
						},
					},
				},
			},
			BackoffLimit: &backoffLimit,
		},
	}
	ctrl.SetControllerReference(j.patch, job, j.scheme)
	return jobs.Create(*j.ctx, job, metav1.CreateOptions{})
}

func (j *JobUtil) Get() (*batchv1.Job, error) {
	jobs := j.clientset.BatchV1().Jobs(j.patch.GetNamespace())
	return jobs.Get(*j.ctx, j.patch.GetName()+"-patch", metav1.GetOptions{})
}

func (j *JobUtil) Owned() (bool, error) {
	job, err := j.Get()
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}
	for _, ownerReference := range job.OwnerReferences {
		if ownerReference.UID == j.patch.GetUID() {
			return true, nil
		}
	}
	return false, nil
}

func (j *JobUtil) Delete() error {
	owned, err := j.Owned()
	if err != nil {
		return err
	}
	if !owned {
		return nil
	}
	jobs := j.clientset.BatchV1().Jobs(j.patch.GetNamespace())
	if err := jobs.Delete(*j.ctx, j.patch.Name, metav1.DeleteOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (j *JobUtil) Completed() (bool, string, error) {
	job, err := j.Get()
	if err != nil {
		if k8sErrors.IsNotFound(err) { // if job cleaned up, assume completed
			return true, "", nil
		}
		return false, "", err
	}
	jobFailed := j.findJobStatusCondition(job.Status.Conditions, batchv1.JobFailed)
	if jobFailed != nil && jobFailed.Status == "True" {
		return true, jobFailed.Message, nil
	}
	jobComplete := j.findJobStatusCondition(job.Status.Conditions, batchv1.JobComplete)
	if jobComplete == nil {
		return false, "", nil
	}
	if jobComplete.Status != "True" {
		return false, "", nil
	}
	return true, "", nil
}

func (j *JobUtil) findJobStatusCondition(conditions []batchv1.JobCondition, conditionType batchv1.JobConditionType) *batchv1.JobCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
