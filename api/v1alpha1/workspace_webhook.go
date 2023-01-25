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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
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
	workspacelog.Info("Providing defaults for workspace", "workspaceName", r.Name)

	defaultWorkflowsSpec(r)
	defaultExperimentTrackingSpec(r)
	defaultStorageSpec(r)
	defaultComputeClusterSpec(r)
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
