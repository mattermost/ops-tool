#!/bin/bash

ART=$(figlet -f banner "$ARG_TEXT")

jq --null-input \
   --arg art "$ART" \
   '{status:"ok", "data": $art }'
