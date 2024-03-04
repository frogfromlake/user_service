#!/bin/sh

# (re)start application and its dependencies
docker compose -f docker-compose.yaml down --volumes
docker compose -f docker-compose.yaml up -d --build