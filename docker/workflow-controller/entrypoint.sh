PREFECT_ORION_DATABASE_CONNECTION_URL="postgresql+asyncpg://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
prefect orion start --host 0.0.0.0 --port 4200 --log-level WARNING