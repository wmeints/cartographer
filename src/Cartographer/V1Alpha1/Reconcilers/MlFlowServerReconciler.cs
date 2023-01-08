using System.Collections.ObjectModel;
using Cartographer.Common;
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
        await ReconcileMlFlowServerDeploymentAsync(entity);
        await ReconcileMlFlowServerServiceAsync(entity);
    }

    private async Task ReconcileMlFlowServerDeploymentAsync(V1Alpha1Workspace entity)
    {
        var deploymentLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = "mlflow-server"
        };

        var deploymentName = $"{entity.Name()}-mlflow-server";

        var existingDeployments = await _kubernetes.ListNamespacedDeploymentAsync(
            entity.Namespace(), labelSelector: deploymentLabels.AsLabelSelector());

        if (existingDeployments.Items.Count == 0)
        {
            _logger.CreatingDeployment(deploymentName, entity.Name(), entity.Namespace());

            var deploymentImageName = entity.Spec.ExperimentTracking.Image;

            if (string.IsNullOrEmpty(deploymentImageName))
            {
                deploymentImageName = "willemmeints/mlflow:2.1.1";
            }

            var deployment = new V1Deployment
            {
                Metadata = new V1ObjectMeta
                {
                    Name = deploymentName,
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
                                    Image = deploymentImageName,
                                    Env = new Collection<V1EnvVar>
                                    {
                                        new V1EnvVar
                                        {
                                            Name = "DB_NAME",
                                            ValueFrom = new V1EnvVarSource
                                            {
                                                SecretKeyRef = new V1SecretKeySelector
                                                {
                                                    Name = $"{entity.Name()}-db-pguser-mlflow",
                                                    Key = "dbname"
                                                }
                                            }
                                        },
                                        new V1EnvVar
                                        {
                                            Name = "DB_PORT",
                                            ValueFrom = new V1EnvVarSource
                                            {
                                                SecretKeyRef = new V1SecretKeySelector
                                                {
                                                    Name = $"{entity.Name()}-db-pguser-mlflow",
                                                    Key = "port"
                                                }
                                            }
                                        },
                                        new V1EnvVar
                                        {
                                            Name = "DB_HOST",
                                            ValueFrom = new V1EnvVarSource
                                            {
                                                SecretKeyRef = new V1SecretKeySelector
                                                {
                                                    Name = $"{entity.Name()}-db-pguser-mlflow",
                                                    Key = "host"
                                                }
                                            }
                                        },
                                        new V1EnvVar
                                        {
                                            Name = "DB_USER",
                                            ValueFrom = new V1EnvVarSource
                                            {
                                                SecretKeyRef = new V1SecretKeySelector
                                                {
                                                    Name = $"{entity.Name()}-db-pguser-mlflow",
                                                    Key = "user"
                                                }
                                            }
                                        },
                                        new V1EnvVar
                                        {
                                            Name = "DB_PASS",
                                            ValueFrom = new V1EnvVarSource
                                            {
                                                SecretKeyRef = new V1SecretKeySelector
                                                {
                                                    Name = $"{entity.Name()}-db-pguser-mlflow",
                                                    Key = "password"
                                                }
                                            }
                                        }
                                    },
                                    Ports = new Collection<V1ContainerPort>
                                    {
                                        new V1ContainerPort(containerPort: 5000, name: "http-mlflow")
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

    private async Task ReconcileMlFlowServerServiceAsync(V1Alpha1Workspace entity)
    {
        var serviceLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = "mlflow-server"
        };

        var serviceName = $"{entity.Name()}-mlflow-server";

        var existingServices = await _kubernetes.ListNamespacedServiceAsync(
            entity.Namespace(), labelSelector: serviceLabels.AsLabelSelector());

        if (existingServices.Items.Count == 0)
        {
            _logger.CreatingService(serviceName, entity.Name(), entity.Namespace());

            var service = new V1Service
            {
                Metadata = new V1ObjectMeta
                {
                    Name = serviceName,
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
                            Name = "http-mlflow",
                            Port = 5000,
                            Protocol = "TCP",
                            TargetPort = 5000
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