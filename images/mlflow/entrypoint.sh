#!/bin/sh

# Make sure we have sensible defaults for the registry and backend storage.
MLFLOW_BACKEND_STORE="postgresql://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
 
mlflow server --backend-store-uri "${MLFLOW_BACKEND_STORE}" --host 0.0.0.0
