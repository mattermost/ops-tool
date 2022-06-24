#!/bin/bash
PAGEID="1"

if [ -n "$1" ]
then
    PAGEID="$1"
fi

curl -s --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "${GITLAB_URL}/api/v4/projects/${REPO_ID}/pipelines?page=${PAGEID}" | jq -c -s '{data: .[0]}'



