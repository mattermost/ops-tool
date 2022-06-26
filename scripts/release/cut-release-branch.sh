#!/bin/bash

STATUS="cut"
BASE_NAME=$(basename "$REPO_URL" ".git")

function cloneRepo(){
    git clone "$REPO_URL"
    cd "${BASE_NAME}"
}

function deleteRepo(){
    cd ..
    rm -rf "${BASE_NAME}"
}

cloneRepo

sed -i "s/BRANCH_DEST=.*/BRANCH_DEST=${BRANCH_DEST}/g" cut-release.vars
sed -i "s/BRANCH_DEST_MOBILE=.*/BRANCH_DEST_MOBILE=${BRANCH_DEST_MOBILE}/g" cut-release.vars
sed -i "s/SEMVER_RELEASE=.*/SEMVER_RELEASE=${SEMVER_RELEASE}/g" cut-release.vars

COMMIT_LOG=$(git commit -am "Bump version to ${SEMVER_RELEASE}, mobile ${BRANCH_DEST_MOBILE}")
retVal=$?
if [ $retVal -ne 0 ]
then
    jq  --null-input \
        --arg branch "$BRANCH_DEST" \
        --arg branch_mobile "$BRANCH_DEST_MOBILE" \
        --arg version "$SEMVER_RELEASE" \
        '{"status": "commit_error",  "data": { "error":"It is already configured.", "SEMVER_RELEASE": $version,"BRANCH_DEST": $branch,"BRANCH_DEST_MOBILE": $branch_mobile }}'
    deleteRepo
    exit 0
fi

git --dry-run push

deleteRepo

jq  --null-input \
    --arg status "$STATUS" \
    --arg branch "$BRANCH_DEST" \
    --arg branch_mobile "$BRANCH_DEST_MOBILE" \
    --arg version "$SEMVER_RELEASE" \
    '{"status": $status, "data": { "SEMVER_RELEASE": $version,"BRANCH_DEST": $branch,"BRANCH_DEST_MOBILE": $branch_mobile }}'



