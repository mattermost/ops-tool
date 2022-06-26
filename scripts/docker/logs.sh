#!/bin/bash

LOG=$(docker logs --tail 100 "$1" 2>&1)

jq --null-input \
   --arg log "$LOG" \
   '{"data": $log }'
