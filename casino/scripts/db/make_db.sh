# ! /bin/bash

source ./scripts/db/env.sh
docker compose -f './compose/docker-compose.local.yaml' up -d