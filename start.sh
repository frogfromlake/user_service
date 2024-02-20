#!/bin/sh

# This file serves as a reference to show how to use ENTRYPOINT in a Dockerfile
# as well as entrypoint with command in a docker-compose file.
set -e

echo "Starting User Service"
exec "$@"