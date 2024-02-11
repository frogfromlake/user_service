#!/bin/sh

set -e

echo "Running DB migrations"
/streamfair_user_svc/migrate -path /streamfair_user_svc/migration -database "$DB_SOURCE_USER_SERVICE" -verbose up

echo "Starting User Service"
exec "$@"