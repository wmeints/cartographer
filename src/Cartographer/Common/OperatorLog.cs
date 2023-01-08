namespace Cartographer.Common;

/// <summary>
/// Provides statically typed logging messages for the application
/// </summary>
public static partial class OperatorLog
{
    [LoggerMessage(
        EventId = 1001,
        Level = LogLevel.Information,
        Message = "Creating deployment {DeploymentName} for workspace {WorkspaceName} in namespace {Namespace}"
    )]
    public static partial void CreatingDeployment(this ILogger logger, string deploymentName, string workspaceName, string namespaceName);

    [LoggerMessage(
        EventId = 1002,
        Level = LogLevel.Information,
        Message = "Creating service {ServiceName} for workspace {WorkspaceName} in namespace {Namespace}"
    )]
    public static partial void CreatingService(this ILogger logger, string serviceName, string workspaceName, string namespaceName);

    [LoggerMessage(
        EventId = 1002,
        Level = LogLevel.Information,
        Message = "Creating stateful set {StatefulSetName} for workspace {WorkspaceName} in namespace {Namespace}"
    )]
    public static partial void CreatingStatefulSet(this ILogger logger, string statefulSetName, string workspaceName, string namespaceName);

    [LoggerMessage(
        EventId = 1003,
        Level = LogLevel.Information,
        Message = "Creating postgres cluster {ClusterName} for workspace {WorkspaceName} in namespace {Namespace}"
    )]
    public static partial void CreatingPostgresCluster(this ILogger logger, string clusterName, string workspaceName, string namespaceName);

    [LoggerMessage(
        EventId = 1003,
        Level = LogLevel.Information,
        Message = "Scaling deployment {DeploymentName} for workspace {WorkspaceName} in namespace {Namespace} to {Replicas} replicas"
    )]
    public static partial void ScalingDeployment(this ILogger logger, string deploymentName, string workspaceName, string namespaceName, int replicas);
}