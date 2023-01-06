using System.Collections.ObjectModel;
using k8s.Models;
using KubeOps.Operator.Entities;
using KubeOps.Operator.Entities.Annotations;

namespace Cartographer.V1Alpha1.Entities;

/// <summary>
/// Defines a prefect environment for a project.
/// </summary>
[KubernetesEntity(Group = "mlops.aigency.com", ApiVersion = "v1alpha1", Kind = "Workspace")]
public class V1Alpha1Workspace : CustomKubernetesEntity<V1Alpha1Workspace.EnvironmentSpec>
{
    /// <summary>
    /// Specification for the prefect environment.
    /// </summary>
    public class EnvironmentSpec
    {
        /// <summary>
        /// Gets or sets the orion database desired configuration
        /// </summary>
        public WorkflowsSpec Workflows { get; set; } = new();
    }

    /// <summary>
    /// Describes the configuration of the orion database
    /// </summary>
    public class WorkflowsSpec
    {
        /// <summary>
        /// Gets or sets the image to use for the controller and agents.
        /// </summary>
        [Description("The docker image to use for the prefect orion server and agents")]
        public string Image { get; set; } = String.Empty;

        /// <summary>
        /// Gets or sets the number of agents to deploy.
        /// </summary>
        [Items(MinItems = 1, MaxItems = -1)]
        [Description("The pools to create for the workflow environment")]
        public Collection<AgentPoolSpec> AgentPools { get; set; } = new();

        /// <summary>
        /// Gets or sets the number of controllers to deploy.
        /// </summary>
        [RangeMinimum(Minimum = 1)]
        [Description("The number of orion server replicas to deploy")]
        public int ControllerReplicas { get; set; } = 1;

        /// <summary>
        /// Gets or sets the number of replicas to deploy for the database
        /// </summary>
        [RangeMinimum(Minimum = 1)]
        [Description("The number of replicas to deploy for the orion server database")]
        public int DatabaseReplicas { get; set; } = 1;

        /// <summary>
        /// Gets the size of the storage to claim for the database
        /// </summary>
        [Description("The storage space to claim for the orion server database")]
        public ResourceQuantity StorageQuota { get; set; } = new("16Gi");
    }

    /// <summary>
    /// Defines what a pool of workflow agents looks like.
    /// </summary>
    public class AgentPoolSpec
    {
        /// <summary>
        /// Gets or sets the name of the agent pool.
        /// </summary>
        [Description("The name of the agent pool")]
        public string Name { get; set; } = String.Empty;

        /// <summary>
        /// Gets or sets the name of the docker image to use for the agents in the pool.
        /// </summary>
        [Description("The docker image to use for the agents in the pool")]
        public string Image { get; set; } = String.Empty;

        /// <summary>
        /// Gets or sets the number of agents to deploy for the pool.
        /// </summary>
        [Description("Number of agents to deploy for the pool")]
        public int Replicas { get; set; } = 1;
    }
}