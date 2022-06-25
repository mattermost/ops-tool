#!/bin/bash

RES=$(cosign verify --key="$COSIGN_PUB_KEY" "$1" 2>&1)

JSON=$(echo "$RES" | tail -n1)
if jq -e . >/dev/null 2>&1 <<<"$JSON"; then
    STATUS="success"
    TEXT=$(echo "$RES" | sed \$d)
    JSON=$(jq . <<< "$JSON")

else
    STATUS="error"
    JSON="{}"
    TEXT="$RES"
fi

jq --null-input \
   --arg status "$STATUS" \
   --arg text "$TEXT" \
   --argjson json "$JSON" \
   '{"status": $status, "data": {"text":$text, "json":$json}}'