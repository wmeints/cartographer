using System.Collections.ObjectModel;
using System.Text.Json.Serialization;
using k8s.Models;
using KubeOps.Operator.Entities;
using KubeOps.Operator.Entities.Annotations;

namespace Cartographer.V1Alpha1.Entities;

/// <summary>
/// Defines the postgres cluster to use for the workspace.
/// </summary>
[IgnoreEntity]
[KubernetesEntity(Group = "postgres-operator.crunchydata.com", ApiVersion = "v1beta1", Kind = "PostgresCluster",
    PluralName = "postgresclusters")]
public class V1Beta1PostgresCluster : CustomKubernetesEntity<V1Beta1PostgresCluster.PostgresClusterSpec>
{
    /// <summary>
    /// Initializes a new instance of <see cref="V1Beta1PostgresCluster"/>.
    /// </summary>
    public V1Beta1PostgresCluster()
    {
        Kind = "PostgresCluster";
        ApiVersion = "postgres-operator.crunchydata.com/v1beta1";
    }

    /// <summary>
    /// Defines what a postgres cluster looks like
    /// </summary>
    public class PostgresClusterSpec
    {
        /// <summary>
        /// Gets or sets the postgres version to use
        /// </summary>
        [JsonPropertyName("postgresVersion")] public int PostgresVersion { get; set; } = 14;

        /// <summary>
        /// Gets or sets the instances to deploy
        /// </summary>
        [JsonPropertyName("instances")] public Collection<InstanceSpec> Instances { get; set; } = new();

        /// <summary>
        /// Gets or sets the backup settings
        /// </summary>
        [JsonPropertyName("backups")] public BackupSpec Backups { get; set; } = new();

        /// <summary>
        /// Gets or sets the users and associated databases
        /// </summary>
        [JsonPropertyName("users")] public Collection<PostgresUserSpec> Users { get; set; } = new();
    }

    /// <summary>
    /// Defines the structure of a database instance
    /// </summary>
    public class InstanceSpec
    {
        /// <summary>
        /// Gets or sets the name of the database.
        /// </summary>
        public string Name { get; set; } = String.Empty;

        /// <summary>
        /// Gets or sets the volume specification for the database
        /// </summary>
        public DataVolumeClaimSpec DataVolumeClaimSpec { get; set; } = new();
    }

    /// <summary>
    /// Defines the structure of a postgres cluster data volume
    /// </summary>
    public class DataVolumeClaimSpec
    {
        /// <summary>
        /// Gets or sets the access modes supported by the volume
        /// </summary>
        public Collection<string> AccessModes { get; set; } = new();

        /// <summary>
        /// Gets or sets the resource requests for the volume
        /// </summary>
        public V1ResourceRequirements Resources { get; set; } = new();
    }

    /// <summary>
    /// Defines the structure of the backup for the postgres cluster
    /// </summary>
    public class BackupSpec
    {
        /// <summary>
        /// Gets or sets the configuration for PGBackRest.
        /// </summary>
        /// <remarks>Learn more about pgbackrest here: https://pgbackrest.org/</remarks>
        [JsonPropertyName("pgbackrest")]
        public BackRestSpec PgBackRest { get; set; } = new();
    }

    /// <summary>
    /// Defines the PGBackRest backup configuration
    /// </summary>
    public class BackRestSpec
    {
        /// <summary>
        /// Gets or sets the list of repositories where the backups should be stored.
        /// </summary>
        [JsonPropertyName("repos")]
        public Collection<BackupRepositorySpec> Repositories { get; set; } = new();
    }

    /// <summary>
    /// Defines the configuration for a backup repository
    /// </summary>
    public class BackupRepositorySpec
    {
        /// <summary>
        /// Gets or sets the name of the backup repository (must be repo1...repo4)
        /// </summary>
        public string Name { get; set; } = String.Empty;

        /// <summary>
        /// Gets or sets the volume configuration.
        /// </summary>
        public BackupVolumeSpec Volume { get; set; } = new();
    }

    /// <summary>
    /// Defines the backup volume configuration
    /// </summary>
    public class BackupVolumeSpec
    {
        /// <summary>
        /// Gets or sets the backup volume specification
        /// </summary>
        public DataVolumeClaimSpec VolumeClaimSpec { get; set; } = new();
    }

    public class PostgresUserSpec
    {
        public string Name { get; set; } = String.Empty;
        public Collection<string> Databases { get; set; } = new();
        public string Options { get; set; } = String.Empty;
    }
}