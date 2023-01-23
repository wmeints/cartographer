package controllers

import (
	"context"

	postgres "github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *WorkspaceReconciler) reconcilePostgresCluster(ctx context.Context, workspace *mlopsv1alpha1.Workspace) error {
	logger := log.FromContext(ctx).WithValues(
		"workspace", workspace.GetName(),
		"namespace", workspace.GetNamespace())

	cluster := &postgres.PostgresCluster{}
	clusterName := workspace.GetName()

	if err := r.Get(ctx, types.NamespacedName{Name: clusterName, Namespace: workspace.GetNamespace()}, cluster); err != nil {
		if errors.IsNotFound(err) {
			// The total storage capacity is the same as the sum of the storage capacity of the experiment tracking
			// database and the workflow controller database.
			storageQuantity := workspace.Spec.Storage.DatabaseStorage
			backupQuantity := workspace.Spec.Storage.DatabaseBackupStorage

			cluster = &postgres.PostgresCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: workspace.GetNamespace(),
					Labels: map[string]string{
						"mlops.aigency.com/workspace": workspace.GetName(),
						"mlops.aigency.com/component": "postgres-cluster",
					},
				},
				Spec: postgres.PostgresClusterSpec{
					Image:           "registry.developers.crunchydata.com/crunchydata/crunchy-postgres:ubi8-14.6-2",
					PostgresVersion: 14,
					InstanceSets: []postgres.PostgresInstanceSetSpec{
						{
							Name: "db01",
							DataVolumeClaimSpec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									corev1.ReadWriteOnce,
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"storage": storageQuantity,
									},
								},
							},
						},
					},
					Backups: postgres.Backups{
						PGBackRest: postgres.PGBackRestArchive{
							Repos: []postgres.PGBackRestRepo{
								{
									Name: "repo1",
									Volume: &postgres.RepoPVC{
										VolumeClaimSpec: corev1.PersistentVolumeClaimSpec{
											AccessModes: []corev1.PersistentVolumeAccessMode{
												corev1.ReadWriteOnce,
											},
											Resources: corev1.ResourceRequirements{
												Requests: corev1.ResourceList{
													"storage": backupQuantity,
												},
											},
										},
									},
								},
							},
						},
					},
					Users: []postgres.PostgresUserSpec{
						{
							Name:      "mlflow",
							Databases: []postgres.PostgresIdentifier{"mlflow"},
						},
						{
							Name:      "prefect",
							Databases: []postgres.PostgresIdentifier{"prefect"},
						},
					},
				},
			}

			if err := ctrl.SetControllerReference(workspace, cluster, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference for postgres cluster")
				return err
			}

			if err := r.Create(ctx, cluster); err != nil {
				logger.Error(err, "Failed to create postgres cluster for the workspace")
				return err
			}

			return nil
		}

		logger.Error(err, "Failed to get postgres cluster for the workspace")
		return err
	}

	return nil
}
