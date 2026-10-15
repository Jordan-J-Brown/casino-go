#!/bin/bash

# Start platform
cd casino || exit 1
make platform || exit 1
for i in {0..3}; do
  echo "Sleeping for 5 seconds..."
  sleep 5

  echo "Checking postgres health. Attempt: $((i+1))"
  status="$(docker inspect postgres --format='{{ .State.Health.Status }}')"
  echo "Postgres status: $status"
  if [[ "$status" == 'healthy' ]]; then
    break
  fi
  if [[ "$i" == '2' ]]; then \
    echo "Postgres health check failed"
    exit 1
  fi
done

make migrate-up || exit 1

# Start casino
make build || exit 1
make up || exit 1
cd ../ || exit 1

# Start engine
make build-engine engine=mines || exit 1
make engine-up engine=mines || exit 1