using System.Collections.ObjectModel;
using Cartographer.V1Alpha1.Entities;
using k8s;
using k8s.Models;
using KubeOps.Operator.Entities.Extensions;

namespace Cartographer.V1Alpha1.Reconcilers;

/// <summary>
/// Reconciles the MLFlow components.
/// </summary>
public class MlFlowReconciler
{
    private readonly IKubernetes _kubernetes;
    private readonly ILogger _logger;
    
    /// <summary>
    /// Initializes a new instance of <see cref="MlFlowReconciler"/>
    /// </summary>
    /// <param name="kubernetes">Kubernetes client to use</param>
    /// <param name="logger">Logger to use</param>
    public MlFlowReconciler(IKubernetes kubernetes, ILogger logger)
    {
        _logger = logger;
        _kubernetes = kubernetes;
    }

    /// <summary>
    /// Reconciles the MLFlow components.
    /// </summary>
    /// <param name="entity"></param>
    public async Task ReconcileAsync(V1Alpha1Workspace entity)
    {
        await ReconcileMLFlowPersistentVolumeClaimAsync(entity);
        await ReconcileMlFlowServerDeploymentAsync(entity);
    }

    private async Task ReconcileMLFlowPersistentVolumeClaimAsync(V1Alpha1Workspace entity)
    {
        var persistentVolumeClaimLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = $"{entity.Name()}-mlflow-server-pvc",
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
                    Name = $"{entity.Name()}-mlflow-server-pvc",
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

    private async Task ReconcileMlFlowServerDeploymentAsync(V1Alpha1Workspace entity)
    {
        var deploymentLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = "mlflow-server"
        };

        var existingDeployments = await _kubernetes.ListNamespacedDeploymentAsync(
            entity.Namespace(), labelSelector: deploymentLabels.AsLabelSelector());

        if (existingDeployments.Items.Count == 0)
        {
            _logger.LogInformation(
                "Existing database deployment not found for {EnvironmentName}. Creating a new database deployment",
                entity.Name());

            var deployment = new V1Deployment
            {
                Metadata = new V1ObjectMeta
                {
                    Name = $"{entity.Name()}-mlflow-server",
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
                            Name = $"{entity.Name()}-mlflow-server"
                        },
                        Spec = new V1PodSpec
                        {
                            Containers = new Collection<V1Container>
                            {
                                new V1Container
                                {
                                    Name = "mlflow",
                                    Image = "willemmeints/mlflow:2.1.1",
                                    Env = new Collection<V1EnvVar>
                                    {
                                        new V1EnvVar
                                        {
                                            Name = "MLFLOW_BACKEND_STORE",
                                            ValueFrom = new V1EnvVarSource
                                            {
                                                SecretKeyRef = new V1SecretKeySelector
                                                {
                                                    Key = "mlflowDatabaseConnectionUrl",
                                                    Name = $"{entity.Name()}-environment-secrets"
                                                }
                                            }
                                        }
                                    },
                                    Ports = new Collection<V1ContainerPort>
                                    {
                                        new V1ContainerPort(containerPort: 5000, name: "http-mlflow")
                                    },
                                    VolumeMounts = new Collection<V1VolumeMount>
                                    {
                                        new V1VolumeMount(mountPath: "/var/data/mlflow", name: "data")
                                    },
                                    Resources = new V1ResourceRequirements
                                    {
                                        Requests = new Dictionary<string, ResourceQuantity>
                                        {
                                            ["cpu"] = new ResourceQuantity("100m"),
                                            ["memory"] = new ResourceQuantity("256Mi")
                                        },
                                        Limits = new Dictionary<string, ResourceQuantity>
                                        {
                                            ["cpu"] = new ResourceQuantity("500m"),
                                            ["memory"] = new ResourceQuantity("512Mi")
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
                                            claimName: $"{entity.Name()}-mlflow-server-pvc")
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
}