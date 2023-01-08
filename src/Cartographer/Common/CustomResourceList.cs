using k8s;
using k8s.Models;

namespace Cartographer.Common;

public class CustomResourceList<T> : KubernetesObject where T : IKubernetesObject
{
    public V1ListMeta Metadata { get; set; } = new();
    public List<T> Items { get; set; } = new();
}