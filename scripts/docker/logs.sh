#!/bin/bash

LOG=$(docker logs "$1" 2>&1)

jq --null-input \
   --arg log "$LOG" \
   '{"data": $log }'
