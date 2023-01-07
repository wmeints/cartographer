using Cartographer.V1Alpha1.Entities;
using Cartographer.V1Alpha1.Reconcilers;
using k8s;
using k8s.Models;
using KubeOps.Operator.Controller;
using KubeOps.Operator.Controller.Results;
using KubeOps.Operator.Rbac;

namespace Cartographer.V1Alpha1.Controllers;

[EntityRbac(typeof(V1Alpha1Workspace), Verbs = RbacVerb.All)]
[EntityRbac(typeof(V1StatefulSet), Verbs = RbacVerb.All)]
[EntityRbac(typeof(V1Deployment), Verbs = RbacVerb.All)]
[EntityRbac(typeof(V1Service), Verbs = RbacVerb.All)]
[EntityRbac(typeof(V1Secret), Verbs = RbacVerb.All)]
public class V1Alpha1WorkspaceController : IResourceController<V1Alpha1Workspace>
{
    private readonly ILogger<V1Alpha1WorkspaceController> _logger;
    private readonly OrionDatabaseReconciler _orionDatabaseReconciler;
    private readonly OrionServerReconciler _orionServerReconciler;
    private readonly EnvironmentSecretsReconciler _environmentSecretsReconciler;
    private readonly PrefectAgentReconciler _prefectAgentReconciler;
    
    public V1Alpha1WorkspaceController(ILogger<V1Alpha1WorkspaceController> logger, IKubernetes kubernetes)
    {
        _logger = logger;
        _orionDatabaseReconciler = new OrionDatabaseReconciler(kubernetes, logger);
        _orionServerReconciler = new OrionServerReconciler(kubernetes, logger);
        _environmentSecretsReconciler = new EnvironmentSecretsReconciler(kubernetes, logger);
        _prefectAgentReconciler = new PrefectAgentReconciler(kubernetes, logger);
    }

    public async Task<ResourceControllerResult?> ReconcileAsync(V1Alpha1Workspace entity)
    {
        _logger.LogInformation("Reconciling the workspace {EnvironmentName}", entity.Name());

        await _environmentSecretsReconciler.ReconcileAsync(entity);
        await _orionDatabaseReconciler.ReconcileAsync(entity);
        await _orionServerReconciler.ReconcileAsync(entity);
        await _prefectAgentReconciler.ReconcileAsync(entity);
        
        return null;
    }

    public Task StatusModifiedAsync(V1Alpha1Workspace entity)
    {
        _logger.LogInformation("entity {Name} called {StatusModifiedAsyncName}", 
            entity.Name(), nameof(StatusModifiedAsync));

        return Task.CompletedTask;
    }

    public Task DeletedAsync(V1Alpha1Workspace entity)
    {
        _logger.LogInformation("entity {Name} called {DeletedAsyncName}", entity.Name(), nameof(DeletedAsync));

        return Task.CompletedTask;
    }
}