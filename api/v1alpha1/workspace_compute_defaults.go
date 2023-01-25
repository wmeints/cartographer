package v1alpha1

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
)

func defaultComputeClusterSpec(r *Workspace) {
	defaultComputerClusterControllerSpec(r)
	defaultComputeWorkerPoolSpecs(r)
}

func defaultComputerClusterControllerSpec(r *Workspace) {
	if r.Spec.Compute.Controller.Image == "" {
		r.Spec.Compute.Controller.Image = fmt.Sprintf("rayproject/ray:%s", r.Spec.Compute.RayVersion)
	}

	if r.Spec.Compute.Controller.Replicas == nil {
		r.Spec.Compute.Controller.Replicas = pointer.Int32(1)
	}

	if len(r.Spec.Compute.Controller.Resources.Requests) == 0 {
		r.Spec.Compute.Controller.Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("500Mi"),
		}
	}

	if len(r.Spec.Compute.Controller.Resources.Limits) == 0 {
		r.Spec.Compute.Controller.Resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		}
	}
}

func defaultComputeWorkerPoolSpecs(r *Workspace) {
	workerGroups := []ComputeWorkerPoolSpec{}

	for _, workerPoolSpec := range r.Spec.Compute.WorkerPools {
		if workerPoolSpec.Image == "" {
			workerPoolSpec.Image = fmt.Sprintf("rayproject/ray:%s", r.Spec.Compute.RayVersion)
		}

		if workerPoolSpec.MinReplicas == nil {
			workerPoolSpec.MinReplicas = pointer.Int32(1)
		}

		if workerPoolSpec.MaxReplicas == nil {
			workerPoolSpec.MaxReplicas = pointer.Int32(1)
		}

		if len(workerPoolSpec.Resources.Requests) == 0 {
			workerPoolSpec.Resources.Requests = corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			}
		}

		if len(workerPoolSpec.Resources.Limits) == 0 {
			workerPoolSpec.Resources.Limits = corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			}
		}

		// Make sure the limits are equal to the rquests otherwise Ray will not work.
		if !reflect.DeepEqual(workerPoolSpec.Resources.Limits, workerPoolSpec.Resources.Requests) {
			workerPoolSpec.Resources.Requests = workerPoolSpec.Resources.Limits
		}

		workerGroups = append(workerGroups, workerPoolSpec)
	}

	r.Spec.Compute.WorkerPools = workerGroups
}
