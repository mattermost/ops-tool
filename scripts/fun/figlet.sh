#!/bin/bash

ART=$(figlet -f banner "$@")

jq --null-input \
   --arg art "$ART" \
   '{status:"ok", "data": $art }'
