#!/bin/bash

docker inspect "$1" | jq -c '{data: .}'