#!/bin/bash

DATA="$HELLO $WORLD"


jq  --null-input \
    --arg status "ok" \
    --arg data "$DATA" \
    '{"status": $status, "data": $data}'




