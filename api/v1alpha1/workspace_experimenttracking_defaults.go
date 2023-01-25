package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
)

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
