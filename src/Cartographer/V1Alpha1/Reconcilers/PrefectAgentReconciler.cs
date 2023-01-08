using System.Collections.ObjectModel;
using Cartographer.Common;
using Cartographer.V1Alpha1.Entities;
using k8s;
using k8s.Models;
using KubeOps.Operator.Entities.Extensions;
using Microsoft.Rest;
using ThorstenHans.JsonPatch.Contrib;

namespace Cartographer.V1Alpha1.Reconcilers;

/// <summary>
/// Reconciles the state of the prefect agent pools
/// </summary>
public class PrefectAgentReconciler
{
    private readonly IKubernetes _kubernetes;
    private readonly ILogger _logger;

    /// <summary>
    /// Initializes a new instance of <see cref="PrefectAgentReconciler"/>
    /// </summary>
    /// <param name="kubernetes">Kubernetes client to use</param>
    /// <param name="logger">Logger to write log message to</param>
    public PrefectAgentReconciler(IKubernetes kubernetes, ILogger logger)
    {
        _logger = logger;
        _kubernetes = kubernetes;
    }

    /// <summary>
    /// Reconciles the agent deployment for the prefect environment.
    /// </summary>
    /// <param name="entity">Workspace to reconcile the agent deployment for.</param>
    public async Task ReconcileAsync(V1Alpha1Workspace entity)
    {
        foreach (var agentPoolSpec in entity.Spec.Workflows.AgentPools)
        {
            await ReconcileAgentPoolAsync(entity, agentPoolSpec);
        }

        await DeleteUnusedAgentPoolsAsync(entity);
    }

    private async ValueTask DeleteUnusedAgentPoolsAsync(V1Alpha1Workspace entity)
    {
        var statefulSetLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/component"] = "prefect-agent",
            ["mlops.aigency.com/environment"] = entity.Name()
        };

        V1StatefulSetList existingStatefulSets = await _kubernetes.ListNamespacedStatefulSetAsync(
            entity.Namespace(), labelSelector: statefulSetLabels.AsLabelSelector());

        foreach (var statefulSet in existingStatefulSets.Items)
        {
            if (!entity.Spec.Workflows.AgentPools.Any(x => x.Name == statefulSet.GetLabel("mlops.aigency.com/pool")))
            {
                await _kubernetes.DeleteNamespacedStatefulSetAsync(statefulSet.Name(), entity.Namespace());
            }
        }
    }

    private async Task ReconcileAgentPoolAsync(V1Alpha1Workspace entity, V1Alpha1Workspace.AgentPoolSpec agentPoolSpec)
    {
        var statefulSetLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/component"] = "prefect-agent",
            ["mlops.aigency.com/pool"] = agentPoolSpec.Name,
            ["mlops.aigency.com/environment"] = entity.Name()
        };

        var existingStatefulSets = await _kubernetes.ListNamespacedStatefulSetAsync(
            entity.Namespace(), labelSelector: statefulSetLabels.AsLabelSelector());

        if (existingStatefulSets.Items.Count == 0)
        {
            await CreateAgentPoolDeploymentAsync(entity, agentPoolSpec, statefulSetLabels);
            return;
        }

        var existingStatefulSet = existingStatefulSets.Items.First();

        if (existingStatefulSet.Spec.Replicas != agentPoolSpec.Replicas)
        {
            await ScaleAgentPoolAsync(entity, agentPoolSpec, existingStatefulSet);
        }
    }

    private async Task ScaleAgentPoolAsync(V1Alpha1Workspace entity, V1Alpha1Workspace.AgentPoolSpec agentPoolSpec,
        V1StatefulSet existingStatefulSet)
    {
        _logger.ScalingDeployment(
            $"{entity.Name()}-agent-{agentPoolSpec.Name}",
            entity.Name(), entity.Namespace(), agentPoolSpec.Replicas);

        var patchDocument = JsonPatchDocumentBuilder.BuildFor<V1StatefulSet>();
        patchDocument.Replace(x => x.Spec.Replicas, agentPoolSpec.Replicas);

        await _kubernetes.PatchNamespacedStatefulSetScaleAsync(
            new V1Patch(patchDocument.ToJsonPatch(), V1Patch.PatchType.JsonPatch),
            existingStatefulSet.Name(),
            existingStatefulSet.Namespace());
    }

    private async Task CreateAgentPoolDeploymentAsync(V1Alpha1Workspace entity,
        V1Alpha1Workspace.AgentPoolSpec agentPoolSpec, Dictionary<string, string> labels)
    {
        var statefulSetName = $"{entity.Name()}-agent-{agentPoolSpec.Name}";
        var deploymentImageName = agentPoolSpec.Image;

        if (string.IsNullOrEmpty(deploymentImageName))
        {
            deploymentImageName = "prefecthq/prefect:2-latest";
        }

        _logger.CreatingStatefulSet(statefulSetName, entity.Name(), entity.Namespace());

        var statefulSet = new V1StatefulSet
        {
            Metadata = new V1ObjectMeta
            {
                Name = statefulSetName,
                Labels = labels,
            },
            Spec = new V1StatefulSetSpec
            {
                ServiceName = $"{entity.Name()}-agent-{agentPoolSpec.Name}",
                Replicas = agentPoolSpec.Replicas,
                Selector = new V1LabelSelector
                {
                    MatchLabels = labels
                },
                Template = new V1PodTemplateSpec
                {
                    Metadata = new V1ObjectMeta
                    {
                        Labels = labels
                    },
                    Spec = new V1PodSpec
                    {
                        Containers = new Collection<V1Container>
                        {
                            new V1Container
                            {
                                Name = "agent",
                                Image = deploymentImageName,
                                Command = new Collection<string>
                                {
                                    "prefect", "agent", "start", "-q", agentPoolSpec.Name
                                },
                                Env = new Collection<V1EnvVar>
                                {
                                    new V1EnvVar(name: "PREFECT_API_URL",
                                        value: $"http://{entity.Name()}-orion-server:4200/api")
                                },
                                Resources = new V1ResourceRequirements
                                {
                                    Requests = agentPoolSpec.ResourceRequests,
                                    Limits = agentPoolSpec.ResourceLimits
                                }
                            }
                        }
                    }
                }
            }
        };

        await _kubernetes.CreateNamespacedStatefulSetAsync(
            statefulSet.WithOwnerReference(entity),
            entity.Namespace());
    }
}