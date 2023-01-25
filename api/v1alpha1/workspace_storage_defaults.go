package v1alpha1

import "k8s.io/apimachinery/pkg/api/resource"

func defaultStorageSpec(r *Workspace) {
	if r.Spec.Storage.DatabaseStorage.IsZero() {
		r.Spec.Storage.DatabaseStorage = resource.MustParse("10Gi")
	}

	if r.Spec.Storage.DatabaseBackupStorage.IsZero() {
		r.Spec.Storage.DatabaseBackupStorage = resource.MustParse("10Gi")
	}
}
