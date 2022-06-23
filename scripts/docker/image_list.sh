#!/bin/bash

docker image list --format='{"ID":"{{ .ID }}", "Repository": "{{ .Repository }}", "Tag":"{{ .Tag }}", "CreatedAt": "{{ .CreatedAt }}", "Size": "{{ .Size }}"}' | jq --slurp | jq -c '{data: .}'