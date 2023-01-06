using Cartographer.V1Alpha1.Controllers;
using Cartographer.V1Alpha1.Entities;
using Cartographer.V1Alpha1.Webhooks;
using k8s;
using KubeOps.Operator;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddSingleton<IKubernetes>(sp =>
{
    // Since we can run inside or outside the cluster,
    // we need to set up a different configuration for each of the cases.
    var config = KubernetesClientConfiguration.IsInCluster() switch
    {
        true => KubernetesClientConfiguration.InClusterConfig(),
        false => KubernetesClientConfiguration.BuildConfigFromConfigFile()
    };
    
    return new Kubernetes(config);
});

// You can use assembly scanning, but it breaks silently. This breaks in your face when you screw up.
// Makes it a lot easier to debug the operator ;-)
builder.Services
    .AddKubernetesOperator(options =>
    {
        options.Name = "cartographer";
        options.EnableAssemblyScanning = false;
    })
    .AddEntity<V1Alpha1Workspace>()
    .AddController<V1Alpha1WorkspaceController>()
    .AddValidationWebhook<V1Alpha1WorkspaceValidator>();

var app = builder.Build();
app.UseKubernetesOperator();
await app.RunOperatorAsync(args);
