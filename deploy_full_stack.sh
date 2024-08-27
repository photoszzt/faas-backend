#!/bin/sh

echo "Deploying OpenFaaS core services and Galvanic backend"
docker stack deploy galvanic --compose-file docker-compose-full.yml
