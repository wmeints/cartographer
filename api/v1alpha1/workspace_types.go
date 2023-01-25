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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	// Workflows defines the configuration for the workflow spec
	Workflows WorkflowComponentSpec `json:"workflows,omitempty"`

	// ExperimentTracking defines the configuration for the MLFlow experiment tracking component
	ExperimentTracking ExperimentTrackingComponentSpec `json:"experimentTracking,omitempty"`

	Storage WorkspaceStorageSpec `json:"storage,omitempty"`

	Compute ComputeSpec `json:"compute,omitempty"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
}

// WorkflowComponentSpec defines the configuration for the workflow component
type WorkflowComponentSpec struct {
	// Controller defines the configuration for the workflow server
	Controller WorkflowControllerSpec `json:"controller,omitempty"`

	// Agents defines the agent pools to deploy
	// +kubebuilder:validation:MinItems=1
	Agents []WorkflowAgentPoolSpec `json:"agentPools,omitempty"`
}

// WorkflowControllerSpec defines the configuration for the workflow controller
type WorkflowControllerSpec struct {
	// ControllerReplicas defines the number of replicas to deploy for the workflow server
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`

	// ControllerResources defines the resource limits and requests for the workflow server
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Image defines the docker image to use for the controller
	Image string `json:"image,omitempty"`
}

// WorkflowAgentPoolSpec defines the shape of an agent pool
type WorkflowAgentPoolSpec struct {
	// Name specifies the name of the agent pool and the associated queue
	Name string `json:"name,omitempty"`

	// Image specifies a custom prefect image to use for the agent pool
	// +optional
	Image string `json:"image,omitempty"`

	// Replicas controls how many agents are deployed in the pool
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas"`

	// Resources define the resource requirements for each agent in the pool
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// ExperimentTrackingComponentSpec defines the configuration for the experiment tracking component
type ExperimentTrackingComponentSpec struct {
	// Image defines the custom docker image to use for deploying MLFlow
	Image string `json:"image,omitempty"`

	// Replicas defines the number of replicas to deploy for the experiment tracking component
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Resources define the resource requirements for the experiment tracking component
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// WorkspaceStorageSpec defines the storage configuration for the workspace
type WorkspaceStorageSpec struct {
	// DatabaseStorage defines the storage requirements for the database
	DatabaseStorage resource.Quantity `json:"database,omitempty"`

	// DatabaseBackupStorage defines the storage requirements for the database backup
	DatabaseBackupStorage resource.Quantity `json:"databaseBackup,omitempty"`
}

// ComputeSpec defines the configuration for the compute cluster
type ComputeSpec struct {
	Controller ComputeControllerSpec `json:"controller,omitempty"`
	Workers    ComputeWorkerSpec     `json:"workers,omitempty"`
	RayVersion string                `json:"rayVersion,omitempty"`
}

// ComputeControllerSpec defines the configuration for the compute cluster controller
type ComputeControllerSpec struct {
	// Replicas controls how many controllers to deploy for the compute cluster controller
	Replicas  *int32 `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ComputeWorkerSpec struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Workspace is the Schema for the workspaces API
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
