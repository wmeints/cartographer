package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
)

func defaultWorkflowsSpec(r *Workspace) {
	defaultWorklowsControllerSpec(r)
	defaultWorkflowsWorkerPoolSpecs(r)
}

func defaultWorklowsControllerSpec(r *Workspace) {
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
}

func defaultWorkflowsWorkerPoolSpecs(r *Workspace) {
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
