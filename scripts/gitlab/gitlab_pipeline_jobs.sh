#!/bin/bash

curl -s --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "${GITLAB_URL}/api/v4/projects/${REPO_ID}/pipelines/${1}/jobs" | jq -c -s '{data: .[0]}'



