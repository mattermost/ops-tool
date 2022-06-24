#!/bin/bash
curl -s --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "${GITLAB_URL}/api/v4/runners/all" | jq -c -s '.[0] | sort_by(.runner_type) | {data: .}'


