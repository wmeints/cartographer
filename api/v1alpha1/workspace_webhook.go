/*
Copyright 2023 Willem Meints.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var workspacelog = logf.Log.WithName("workspace-resource")

func (r *Workspace) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-mlops-aigency-com-v1alpha1-workspace,mutating=true,failurePolicy=fail,sideEffects=None,groups=mlops.aigency.com,resources=workspaces,verbs=create;update,versions=v1alpha1,name=mworkspace.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Workspace{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Workspace) Default() {
	workspacelog.Info("default", "name", r.Name)

	defaultWorkflowsSpec(r)
	defaultExperimentTrackingSpec(r)
	defaultStorageSpec(r)
}

func defaultWorkflowsSpec(r *Workspace) {
	if r.Spec.Workflows.Controller.Image == "" {
		r.Spec.Workflows.Controller.Image = "willemmeints/workflow-controller:latest"
	}

	if len(r.Spec.Workflows.Controller.Resources.Limits) == 0 {
		r.Spec.Workflows.Controller.Resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		}
	}

	if len(r.Spec.Workflows.Controller.Resources.Requests) == 0 {
		r.Spec.Workflows.Controller.Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("200Mi"),
		}
	}

	if r.Spec.Workflows.Controller.Replicas == nil {
		r.Spec.Workflows.Controller.Replicas = pointer.Int32(1)
	}

	agents := []WorkflowAgentPoolSpec{}

	for _, agentPoolSpec := range r.Spec.Workflows.Agents {
		if agentPoolSpec.Image == "" {
			agentPoolSpec.Image = "willemmeints/workflow-agent:latest"
		}

		if len(agentPoolSpec.Resources.Limits) == 0 {
			agentPoolSpec.Resources.Limits = corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("16Gi"),
			}
		}

		if len(agentPoolSpec.Resources.Requests) == 0 {
			agentPoolSpec.Resources.Requests = corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			}
		}

		if agentPoolSpec.Replicas == nil {
			agentPoolSpec.Replicas = pointer.Int32(1)
		}

		agents = append(agents, agentPoolSpec)
	}

	r.Spec.Workflows.Agents = agents
}

func defaultExperimentTrackingSpec(r *Workspace) {
	if r.Spec.ExperimentTracking.Image == "" {
		r.Spec.ExperimentTracking.Image = "willemmeints/experiment-tracking:latest"
	}

	if r.Spec.ExperimentTracking.Replicas == nil {
		r.Spec.ExperimentTracking.Replicas = pointer.Int32(1)
	}

	if len(r.Spec.ExperimentTracking.Resources.Limits) == 0 {
		r.Spec.ExperimentTracking.Resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		}
	}

	if len(r.Spec.ExperimentTracking.Resources.Requests) == 0 {
		r.Spec.ExperimentTracking.Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		}
	}
}

func defaultStorageSpec(r *Workspace) {
	if r.Spec.Storage.DatabaseStorage.IsZero() {
		r.Spec.Storage.DatabaseStorage = resource.MustParse("10Gi")
	}

	if r.Spec.Storage.DatabaseBackupStorage.IsZero() {
		r.Spec.Storage.DatabaseBackupStorage = resource.MustParse("10Gi")
	}
}

//+kubebuilder:webhook:path=/validate-mlops-aigency-com-v1alpha1-workspace,mutating=false,failurePolicy=fail,sideEffects=None,groups=mlops.aigency.com,resources=workspaces,verbs=create;update,versions=v1alpha1,name=mworkspace.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Workspace{}

func (r *Workspace) ValidateCreate() error {
	validationErrors := field.ErrorList{}

	if len(validationErrors) > 0 {
		groupKind := schema.GroupKind{Group: "mlops.aigency.com", Kind: "Workspace"}
		return apierrors.NewInvalid(groupKind, r.Name, validationErrors)
	}

	return nil
}

func (r *Workspace) ValidateUpdate(old runtime.Object) error {
	validationErrors := field.ErrorList{}

	validationErrors = append(validationErrors, validateWorkflowAgentPoolNames(r)...)

	if len(validationErrors) > 0 {
		groupKind := schema.GroupKind{Group: "mlops.aigency.com", Kind: "Workspace"}
		return apierrors.NewInvalid(groupKind, r.Name, validationErrors)
	}

	return nil
}

func (r *Workspace) ValidateDelete() error {
	return nil
}

func validateWorkflowAgentPoolNames(r *Workspace) field.ErrorList {
	validationErrors := field.ErrorList{}
	var agentNames []string

	for index, agentPoolSpec := range r.Spec.Workflows.Agents {
		if slices.Contains(agentNames, agentPoolSpec.Name) {
			err := field.Invalid(
				field.NewPath("spec").Child("workflows").Child("agents").Index(index).Child("name"),
				agentPoolSpec.Name,
				"name must be unique",
			)

			validationErrors = append(validationErrors, err)
		}

		agentNames = append(agentNames, agentPoolSpec.Name)
	}

	return validationErrors
}
