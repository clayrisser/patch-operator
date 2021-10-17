/**
 * File: /job.go
 * Project: util
 * File Created: 17-10-2021 16:35:30
 * Author: Clay Risser
 * -----
 * Last Modified: 17-10-2021 18:54:18
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

	patchv1alpha1 "gitlab.com/bitspur/community/patch-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type JobUtil struct {
	ctx       *context.Context
	cfg       *rest.Config
	clientset *kubernetes.Clientset
	patch     *patchv1alpha1.Patch
}

func NewJobUtil(patch *patchv1alpha1.Patch, ctx *context.Context) *JobUtil {
	cfg := ctrl.GetConfigOrDie()
	return &JobUtil{
		cfg:       cfg,
		clientset: kubernetes.NewForConfigOrDie(cfg),
		ctx:       ctx,
		patch:     patch,
	}
}

func (j *JobUtil) Create(command string, env *[]v1.EnvVar) (*batchv1.Job, error) {
	if command == "" {
		command = "true"
	}
	jobs := j.clientset.BatchV1().Jobs(j.patch.Namespace)
	var backoffLimit int32 = 0
	image := j.patch.Spec.Image
	if image == "" {
		image = "codejamninja/kube-commands:0.0.2"
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      j.patch.Name,
			Namespace: j.patch.Namespace,
			Labels:    j.patch.Labels,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: j.patch.Labels,
				},
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyNever,
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
	return jobs.Create(*j.ctx, job, metav1.CreateOptions{})
}

func (j *JobUtil) Get() (*batchv1.Job, error) {
	jobs := j.clientset.BatchV1().Jobs(j.patch.Namespace)
	return jobs.Get(*j.ctx, j.patch.Name, metav1.GetOptions{})
}

func (j *JobUtil) Delete() error {
	jobs := j.clientset.BatchV1().Jobs(j.patch.Namespace)
	if err := jobs.Delete(*j.ctx, j.patch.Name, metav1.DeleteOptions{}); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (j *JobUtil) Completed() (bool, error) {
	job, err := j.Get()
	if err != nil {
		if errors.IsNotFound(err) { // if job cleaned up, assume completed
			return true, nil
		}
		return false, err
	}
	jobComplete := FindJobStatusCondition(job.Status.Conditions, batchv1.JobComplete)
	if jobComplete == nil {
		return false, nil
	}
	if jobComplete.Status != "True" {
		return false, nil
	}
	return true, nil
}

func FindJobStatusCondition(conditions []batchv1.JobCondition, conditionType batchv1.JobConditionType) *batchv1.JobCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
