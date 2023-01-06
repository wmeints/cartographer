using Cartographer.V1Alpha1.Entities;
using k8s;
using k8s.Models;
using KubeOps.Operator.Entities.Extensions;

namespace Cartographer.V1Alpha1.Reconcilers;

public class EnvironmentSecretsReconciler
{
    private IKubernetes _kubernetes;
    private ILogger _logger;

    /// <summary>
    /// Initializes a new instance of <see cref="OrionDatabaseReconciler"/>
    /// </summary>
    /// <param name="kubernetes">Kubernetes client to use</param>
    /// <param name="logger">Logger to write log message to</param>
    public EnvironmentSecretsReconciler(IKubernetes kubernetes, ILogger logger)
    {
        _logger = logger;
        _kubernetes = kubernetes;
    }

    public async Task ReconcileAsync(V1Alpha1Workspace entity)
    {
        var environmentSecretLabels = new Dictionary<string, string>
        {
            ["mlops.aigency.com/environment"] = entity.Name(),
            ["mlops.aigency.com/component"] = $"{entity.Name()}-environment-secrets",
        };

        var existingSecrets = await _kubernetes.ListNamespacedSecretAsync(
            entity.Namespace(), labelSelector: environmentSecretLabels.AsLabelSelector());

        if (existingSecrets.Items.Count == 0)
        {
            _logger.LogInformation(
                "Environment secrets for {EnvironmentName} not found. Creating a new set of secrets",
                entity.Name());

            var databasePassword = GeneratePassword();
            
            var secret = new V1Secret
            {
                Metadata = new V1ObjectMeta
                {
                    Name = $"{entity.Name()}-environment-secrets",
                    Labels = environmentSecretLabels
                },
                StringData = new Dictionary<string, string>
                {
                    ["orionDatabasePassword"] = databasePassword,
                    ["orionDatabaseConnectionUrl"] = $"postgresql+asyncpg://postgres:{databasePassword}@{entity.Name()}-orion-database:5432/orion" 
                }
            };

            await _kubernetes.CreateNamespacedSecretAsync(
                secret.WithOwnerReference(entity),
                entity.Namespace());
        }
    }

    private string GeneratePassword()
    {
        return Guid.NewGuid().ToString();
    }
}