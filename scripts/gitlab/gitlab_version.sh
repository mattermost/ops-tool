#!/bin/bash

STATUS="ok"
VERSION="1.1"
LATEST_VERSION="1.1"
CHANGE_NOTES_RAW="https://gitlab.com/gitlab-org/gitlab-foss/-/raw/master/CHANGELOG.md"
CHANGE_NOTES="https://gitlab.com/gitlab-org/gitlab-foss/blob/master/CHANGELOG.md"

GITLAB_VERSION=$(curl -s --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "${GITLAB_URL}/api/v4/version" | jq -r '.version' )
VERSION="${GITLAB_VERSION%-*}"

if [ "$1" == "check" ]
then
  RELEASE_VERSION=$(curl -s $CHANGE_NOTES_RAW | grep -E '^##\s[0-9]+\.[0-9]+\.[0-9]+\s\(' | head -n1)
  RELEASE_VERSION="${RELEASE_VERSION% (*}"
  LATEST_VERSION="${RELEASE_VERSION#\#\# *}"

  if [ $LATEST_VERSION != $VERSION ] 
  then
      STATUS="upgrade"
  fi

  jq --null-input \
    --arg status "$STATUS" \
    --arg version "$VERSION" \
    --arg latest "$LATEST_VERSION" \
    --arg changeNotes "$CHANGE_NOTES" \
    '{"status": $status, "data": {"version": $version,"latest": $latest,"change_notes_url": $changeNotes}}'
else
  jq --null-input \
    --arg status "$STATUS" \
    --arg version "$VERSION" \
    '{"status": "ok", "data": { "version": $version }}'
fi