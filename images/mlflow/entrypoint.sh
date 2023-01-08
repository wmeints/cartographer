#!/bin/sh

# Make sure we have sensible defaults for the registry and backend storage.
MLFLOW_BACKEND_STORE="${MLFLOW_BACKEND_STORE:-file:///var/data/mlflow}"
 
mlflow server --backend-store-uri "${MLFLOW_BACKEND_STORE}"
