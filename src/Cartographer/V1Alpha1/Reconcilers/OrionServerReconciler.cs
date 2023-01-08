using System.Collections.ObjectModel;
using Cartographer.Common;
using Cartographer.V1Alpha1.Entities;
using k8s;
using k8s.Models;
using KubeOps.Operator.Entities.Extensions;

namespace Cartographer.V1Alpha1.Reconcilers;

public class OrionServerReconciler
{
    private IKubernetes _kubernetes;
    private ILogger _logger;

    /// <summary>
    /// Initializes a new instance of <see cref="OrionServerReconciler"/>
    /// </summary>
    /// <param name="kubernetes">Kubernetes client to use</param>
    /// <param name="logger">Logger to write log message to</param>
    public OrionServerReconciler(IKubernetes kubernetes, ILogger logger)
    {
        _logger = logger;
        _kubernetes = kubernetes;
    }

    /// <summary>
    /// Reconciles the orion server resources.
    /// </summary>
    /// <param name="entity">Workspace to reconcile the orion server resources for.</param>
    public async Task ReconcileAsync(V1Alpha1Workspace entity)
    {
        await ReconcileOrionServerDeploymentAsync(entity);
        await ReconcileOrionServerServiceAsync(entity);
    }

    private async Task ReconcileOrionServerDeploymentAsync(V1Alpha1Workspace entity)
    {
        var deploymentLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/component"] = "orion-server",
            ["mlops.aigency.com/environment"] = entity.Name()
        };

        var deploymentName = $"{entity.Name()}-orion-server";

        var existingDeployments = await _kubernetes.ListNamespacedDeploymentAsync(
            entity.Namespace(), labelSelector: deploymentLabels.AsLabelSelector());

        if (existingDeployments.Items.Count == 0)
        {
            _logger.CreatingDeployment(deploymentName, entity.Name(), entity.Namespace());

            var deploymentImageName = entity.Spec.Workflows.Image;

            if (string.IsNullOrEmpty(deploymentImageName))
            {
                deploymentImageName = "prefecthq/prefect:2-latest";
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
                            Name = $"{entity.Name()}-orion-server"
                        },
                        Spec = new V1PodSpec
                        {
                            Containers = new Collection<V1Container>
                            {
                                new V1Container
                                {
                                    Name = "orion-server",
                                    Image = deploymentImageName,
                                    Command = new Collection<string> { "prefect", "orion", "start" },
                                    Env = new Collection<V1EnvVar>
                                    {
                                        new V1EnvVar
                                        {
                                            Name = "DB_NAME",
                                            ValueFrom = new V1EnvVarSource
                                            {
                                                SecretKeyRef = new V1SecretKeySelector
                                                {
                                                    Name = $"{entity.Name()}-db-pguser-orion",
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
                                                    Name = $"{entity.Name()}-db-pguser-orion",
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
                                                    Name = $"{entity.Name()}-db-pguser-orion",
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
                                                    Name = $"{entity.Name()}-db-pguser-orion",
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
                                                    Name = $"{entity.Name()}-db-pguser-orion",
                                                    Key = "password"
                                                }
                                            }
                                        },
                                        new V1EnvVar
                                        {
                                            Name = "PREFECT_ORION_DATABASE_CONNECTION_URL",
                                            Value = "postgresql+asyncpg://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)"
                                        },
                                    },
                                    Ports = new Collection<V1ContainerPort>
                                    {
                                        new V1ContainerPort(containerPort: 4200, name: "http-orion")
                                    },
                                    Resources = new V1ResourceRequirements
                                    {
                                        Requests = new Dictionary<string, ResourceQuantity>
                                        {
                                            ["cpu"] = new ResourceQuantity("200m"),
                                            ["memory"] = new ResourceQuantity("512Mi")
                                        },
                                        Limits = new Dictionary<string, ResourceQuantity>
                                        {
                                            ["cpu"] = new ResourceQuantity("500m"),
                                            ["memory"] = new ResourceQuantity("1Gi")
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            };

            await _kubernetes.CreateNamespacedDeploymentAsync(
                deployment.WithOwnerReference(entity), entity.Namespace());
        }
    }

    private async Task ReconcileOrionServerServiceAsync(V1Alpha1Workspace entity)
    {
        var serviceLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/component"] = "orion-server",
            ["mlops.aigency.com/environment"] = entity.Name()
        };

        var serviceName = $"{entity.Name()}-orion-server";

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
                            Name = "http-orion",
                            Port = 4200,
                            Protocol = "TCP",
                            TargetPort = 4200
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