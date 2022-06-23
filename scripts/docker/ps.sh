#!/bin/bash

flags="${DOCKER_PS_FLAGS}"

docker ps $flags --format='{{json . }}' | jq --slurp -c '{data: .}'