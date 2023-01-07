using System.Collections.ObjectModel;
using Cartographer.V1Alpha1.Entities;
using k8s;
using k8s.Models;
using KubeOps.Operator.Controller.Results;
using KubeOps.Operator.Entities.Extensions;
using Microsoft.Rest;

namespace Cartographer.V1Alpha1.Reconcilers;

/// <summary>
/// Reconciles the components related to the orion server database.
/// </summary>
public class OrionDatabaseReconciler
{
    private IKubernetes _kubernetes;
    private ILogger _logger;

    /// <summary>
    /// Initializes a new instance of <see cref="OrionDatabaseReconciler"/>
    /// </summary>
    /// <param name="kubernetes">Kubernetes client to use</param>
    /// <param name="logger">Logger to write log message to</param>
    public OrionDatabaseReconciler(IKubernetes kubernetes, ILogger logger)
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
        await ReconcileDatabasePersistentVolumeClaim(entity);
        await ReconcileDatabaseDeploymentAsync(entity);
        await ReconcileDatabaseServiceAsync(entity);
    }

    private async Task ReconcileDatabasePersistentVolumeClaim(V1Alpha1Workspace entity)
    {
        var persistentVolumeClaimLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = $"{entity.Name()}-orion-database-pvc",
        };

        var existingPersistentVolumeClaims = await _kubernetes.ListNamespacedPersistentVolumeClaimAsync(
            entity.Namespace(), labelSelector: persistentVolumeClaimLabels.AsLabelSelector());

        if (existingPersistentVolumeClaims.Items.Count == 0)
        {
            _logger.LogInformation(
                "Existing database PVC not found for {EnvironmentName}. Creating a new PVC for the environment",
                entity.Name());

            var persistentVolumeClaim = new V1PersistentVolumeClaim
            {
                Metadata = new V1ObjectMeta
                {
                    Name = $"{entity.Name()}-orion-database-pvc",
                    Labels = persistentVolumeClaimLabels
                },
                Spec = new V1PersistentVolumeClaimSpec
                {
                    Resources = new V1ResourceRequirements
                    {
                        Requests = new Dictionary<string, ResourceQuantity>
                        {
                            ["storage"] = entity.Spec.Workflows.StorageQuota,
                        }
                    },
                    StorageClassName = "standard",
                    AccessModes = new Collection<string>
                    {
                        "ReadWriteOnce"
                    }
                }
            };

            await _kubernetes.CreateNamespacedPersistentVolumeClaimAsync(
                persistentVolumeClaim.WithOwnerReference(entity), entity.Namespace());
        }
    }

    private async Task ReconcileDatabaseDeploymentAsync(V1Alpha1Workspace entity)
    {
        var deploymentLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = "orion-database"
        };

        var existingDeployment = await _kubernetes.ListNamespacedDeploymentAsync(
            entity.Namespace(), labelSelector: deploymentLabels.AsLabelSelector());

        if (existingDeployment.Items.Count == 0)
        {
            _logger.LogInformation(
                "Existing database deployment not found for {EnvironmentName}. Creating a new database deployment",
                entity.Name());


            var deployment = new V1Deployment
            {
                Metadata = new V1ObjectMeta
                {
                    Name = $"{entity.Name()}-orion-database",
                    Labels = deploymentLabels
                },
                Spec = new V1DeploymentSpec
                {
                    Replicas = entity.Spec.Workflows.DatabaseReplicas,
                    Selector = new V1LabelSelector
                    {
                        MatchLabels = deploymentLabels
                    },
                    Template = new V1PodTemplateSpec
                    {
                        Metadata = new V1ObjectMeta
                        {
                            Labels = deploymentLabels,
                            Name = $"{entity.Name()}-orion-database"
                        },
                        Spec = new V1PodSpec
                        {
                            Containers = new Collection<V1Container>
                            {
                                new V1Container
                                {
                                    Name = "postgres",
                                    Image = "postgres:14",
                                    Env = new Collection<V1EnvVar>
                                    {
                                        new V1EnvVar
                                        {
                                            Name = "POSTGRES_PASSWORD",
                                            ValueFrom = new V1EnvVarSource
                                            {
                                                SecretKeyRef = new V1SecretKeySelector
                                                {
                                                    Key = "orionDatabasePassword",
                                                    Name = $"{entity.Name()}-environment-secrets"
                                                }
                                            }
                                        },
                                        new V1EnvVar(name: "POSTGRES_DB", value: "orion"),
                                        new V1EnvVar(name: "PGDATA", value: "/var/data/postgresql")
                                    },
                                    Ports = new Collection<V1ContainerPort>
                                    {
                                        new V1ContainerPort(containerPort: 5432, name: "tcp-postgres")
                                    },
                                    VolumeMounts = new Collection<V1VolumeMount>
                                    {
                                        new V1VolumeMount(mountPath: "/var/data/postgresql", name: "data")
                                    },
                                    Resources = new V1ResourceRequirements
                                    {
                                        Requests = new Dictionary<string, ResourceQuantity>
                                        {
                                            ["cpu"] = new ResourceQuantity("500m"),
                                            ["memory"] = new ResourceQuantity("512Mi")
                                        },
                                        Limits = new Dictionary<string, ResourceQuantity>
                                        {
                                            ["cpu"] = new ResourceQuantity("1"),
                                            ["memory"] = new ResourceQuantity("1Gi")
                                        },
                                    }
                                }
                            },
                            Volumes = new Collection<V1Volume>
                            {
                                new V1Volume
                                {
                                    Name = "data",
                                    PersistentVolumeClaim =
                                        new V1PersistentVolumeClaimVolumeSource(
                                            claimName: $"{entity.Name()}-orion-database-pvc")
                                }
                            }
                        }
                    }
                }
            };

            await _kubernetes.CreateNamespacedDeploymentAsync(
                deployment.WithOwnerReference(entity),
                entity.Namespace());
        }
    }

    private async Task ReconcileDatabaseServiceAsync(V1Alpha1Workspace entity)
    {
        var serviceLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = $"orion-database"
        };

        var existingServices = await _kubernetes.ListNamespacedServiceAsync(
            entity.Namespace(), labelSelector: serviceLabels.AsLabelSelector());

        if (existingServices.Items.Count == 0)
        {
            _logger.LogInformation(
                "Existing orion database service not found for {EnvironmentName}. Creating a new orion database service",
                entity.Name());

            var service = new V1Service
            {
                Metadata = new V1ObjectMeta
                {
                    Name = $"{entity.Name()}-orion-database",
                    Labels = serviceLabels,
                },
                Spec = new V1ServiceSpec
                {
                    Type = "ClusterIP",
                    Selector = serviceLabels,
                    Ports = new Collection<V1ServicePort>
                    {
                        new V1ServicePort
                        {
                            Name = "tcp-postgres",
                            Port = 5432,
                            Protocol = "TCP",
                            TargetPort = 5432
                        }
                    }
                }
            };

            await _kubernetes.CreateNamespacedServiceAsync(
                service.WithOwnerReference(entity),
                entity.Namespace());
        }
    }
}