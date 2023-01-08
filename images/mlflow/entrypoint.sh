#!/bin/sh

# Make sure we have sensible defaults for the registry and backend storage.
MLFLOW_BACKEND_STORE="${MLFLOW_BACKEND_STORE:-file:///var/data/mlflow}"

# You can switch of the registry by setting the MLFLOW_DISABLE_REGISTRY flag
# By default, we assume that you use the registry as well as the metrics server.
if [ -z "${MLFLOW_DISABLE_REGISTRY}" ]
then
  mlflow server --backend-store-uri "${MLFLOW_BACKEND_STORE}" --registry-store-uri "${MLFLOW_REGISTRY_STORE}"
else 
  mlflow server --backend-store-uri "${MLFLOW_BACKEND_STORE}"
fi
