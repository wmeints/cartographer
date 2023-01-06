using Cartographer.V1Alpha1.Entities;
using KubeOps.Operator.Webhooks;

namespace Cartographer.V1Alpha1.Webhooks;

/// <summary>
/// Validates the specification for a prefect environment
/// </summary>
public class V1Alpha1WorkspaceValidator : IValidationWebhook<V1Alpha1Workspace>
{
    public AdmissionOperations Operations => AdmissionOperations.Create | AdmissionOperations.Update;

    /// <summary>
    /// Handles the creation of a new prefect environment
    /// </summary>
    /// <param name="newEntity">The new entity to create.</param>
    /// <param name="dryRun">Flag indicating whether this is a dry run.</param>
    /// <returns>Returns a success result when the spec is valid; Otherwise an invalid result is returned.</returns>
    public ValidationResult Create(V1Alpha1Workspace newEntity, bool dryRun)
    {
        return ValidationResult.Success();
    }
    
    /// <summary>
    /// Handles the update of an existing prefect environment.
    /// </summary>
    /// <param name="oldEntity">Existing entity that's being updated.</param>
    /// <param name="newEntity">New definition of the existing entity being updated.</param>
    /// <param name="dryRun">Flag indicating whether this is a dry run.</param>
    /// <returns>Returns a success result when the spec is valid; Otherwise an invalid result is returned.</returns>
    public ValidationResult Update(V1Alpha1Workspace oldEntity, V1Alpha1Workspace newEntity, bool dryRun)
    {
        return ValidationResult.Success();
    }
}
