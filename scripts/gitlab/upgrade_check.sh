#!/bin/bash

STATUS="ok"
VERSION="1.1"
LATEST_VERSION="1.1"
CHANGE_NOTES="https://gitlab.com/gitlab-org/gitlab-foss/blob/master/CHANGELOG.md"

jq --null-input \
  --arg status "$STATUS" \
  --arg version "$VERSION" \
  --arg latest "$LATEST_VERSION" \
  --arg changeNotes "$CHANGE_NOTES" \
  '{"status": $status,"version": $version,"latest": $latest,"change_notes_url": $changeNotes}'