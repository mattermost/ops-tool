#!/bin/bash

case "$1" in
  "digest-only")
    RES=$(docker image inspect --format='{{index .RepoDigests 0}}' "$2")
    ;;
  *)
    RES=$(docker image inspect "$1")
    ;;
esac

jq --null-input \
   --arg res "$RES" \
   '{"data": $res }'