#!/bin/bash

flags="${1}"

docker ps $flags --format='{{json . }}' | jq --slurp -c '{data: .}'