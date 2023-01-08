using System.Collections.ObjectModel;
using Cartographer.Common;
using Cartographer.V1Alpha1.Entities;
using k8s;
using k8s.Models;
using KubeOps.Operator.Entities.Extensions;
using Microsoft.Rest;

namespace Cartographer.V1Alpha1.Reconcilers;

/// <summary>
/// Reconciles the components related to the orion server database.
/// </summary>
public class PostgresClusterReconciler
{
    private IKubernetes _kubernetes;
    private ILogger _logger;

    /// <summary>
    /// Initializes a new instance of <see cref="OrionDatabaseReconciler"/>
    /// </summary>
    /// <param name="kubernetes">Kubernetes client to use</param>
    /// <param name="logger">Logger to write log message to</param>
    public PostgresClusterReconciler(IKubernetes kubernetes, ILogger logger)
    {
        _logger = logger;
        _kubernetes = kubernetes;
    }

    /// <summary>
    /// Reconciles the resources related to the orion server database.
    /// </summary>
    /// <param name="entity">Workspace for which to reconcile the orion server database.</param>
    public async Task ReconcileAsync(V1Alpha1Workspace entity)
    {
        var genericClient = new GenericClient(_kubernetes,
            "postgres-operator.crunchydata.com", "v1beta1",
            "postgresclusters");

        var postgresClusterLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = "postgres-cluster",
        };

        var clusterName = "${entity.Name()}-db";

        var existingClusters = await genericClient.ListNamespacedAsync<CustomResourceList<V1Beta1PostgresCluster>>(
            entity.Namespace());

        if (existingClusters.Items.Count == 0)
        {
            _logger.CreatingPostgresCluster(clusterName, entity.Name(), entity.Namespace());

            var postgresCluster = new V1Beta1PostgresCluster
            {
                Metadata = new V1ObjectMeta
                {
                    Name = clusterName,
                    Labels = postgresClusterLabels
                },
                Spec = new V1Beta1PostgresCluster.PostgresClusterSpec
                {
                    Instances = new Collection<V1Beta1PostgresCluster.InstanceSpec>
                    {
                        new V1Beta1PostgresCluster.InstanceSpec
                        {
                            Name = "instance01",
                            DataVolumeClaimSpec = new V1Beta1PostgresCluster.DataVolumeClaimSpec
                            {
                                AccessModes = new() { "ReadWriteOnce" },
                                Resources = new()
                                {
                                    Requests = new Dictionary<string, ResourceQuantity>
                                    {
                                        ["storage"] = new("10Gi"),
                                    }
                                }
                            }
                        }
                    },
                    PostgresVersion = 14,
                    Backups = new V1Beta1PostgresCluster.BackupSpec()
                    {
                        PgBackRest = new V1Beta1PostgresCluster.BackRestSpec()
                        {
                            Repositories = new Collection<V1Beta1PostgresCluster.BackupRepositorySpec>
                            {
                                new V1Beta1PostgresCluster.BackupRepositorySpec
                                {
                                    // We'll store the backups in the first repository, see also
                                    // https://access.crunchydata.com/documentation/postgres-operator/v5/tutorial/backups/
                                    Name = "repo1",
                                    Volume = new V1Beta1PostgresCluster.BackupVolumeSpec
                                    {
                                        VolumeClaimSpec = new V1Beta1PostgresCluster.DataVolumeClaimSpec()
                                        {
                                            AccessModes = new() { "ReadWriteOnce" },
                                            Resources = new()
                                            {
                                                Requests = new Dictionary<string, ResourceQuantity>
                                                {
                                                    ["storage"] = new("10Gi")
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    },
                    Users = new Collection<V1Beta1PostgresCluster.PostgresUserSpec>
                    {
                        new V1Beta1PostgresCluster.PostgresUserSpec()
                        {
                            Name = "orion",
                            Databases = new() {"orion"}
                        },
                        new V1Beta1PostgresCluster.PostgresUserSpec()
                        {
                            Name = "mlflow",
                            Databases = new() {"mlflow"}
                        },
                    }
                }
            };

            await _kubernetes.CreateNamespacedCustomObjectAsync(
                postgresCluster.WithOwnerReference(entity),
                "postgres-operator.crunchydata.com",
                "v1beta1",
                entity.Namespace(),
                "postgresclusters");
        }
    }
}