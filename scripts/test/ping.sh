#!/bin/bash

jq  --null-input \
    --arg status "ok" \
    --arg data "pong" \
    '{"status": $status, "data": $data}'



