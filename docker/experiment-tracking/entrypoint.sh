#!/bin/sh

ENCODED_PASSWORD=$(python -c "import urllib.parse; print(urllib.parse.quote_plus('${DB_PASS}'))")

# Make sure we have sensible defaults for the registry and backend storage.
MLFLOW_BACKEND_STORE="postgresql://${DB_USER}:${ENCODED_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
 
mlflow server --backend-store-uri "${MLFLOW_BACKEND_STORE}" --host 0.0.0.0
