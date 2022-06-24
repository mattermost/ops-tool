#!/bin/bash
LOG=$(curl --location -s --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "${GITLAB_URL}/api/v4/projects/${REPO_ID}/jobs/${1}/trace" | sed -r "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[m|K]//g")

jq --null-input \
   --arg log "$LOG" \
   '{"data": $log }'




