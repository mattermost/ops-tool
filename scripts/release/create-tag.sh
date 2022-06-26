#!/bin/bash

STATUS="tag"
BASE_NAME=$(basename "$REPO_URL" ".git")

function cloneRepo(){
    git clone -b ${BRANCH_NAME} "$REPO_URL"
    cd "${BASE_NAME}"
}

function deleteRepo(){
    cd ..
    rm -rf "${BASE_NAME}"
}

cloneRepo

if [ "$(git tag -l "${RELEASE_TAG}")" ]; then
    jq  --null-input \
        --arg tag "$RELEASE_TAG" \
        --arg branch "$BRANCH_NAME" \
        '{"status": "tag_exists",  "data": { "error":"Tag is already created!","BRANCH_NAME": $branch,"RELEASE_TAG": $tag }}'
    deleteRepo
    exit 0
fi

if [ "$BRANCH_NAME" != "$MAIN_BRANCH" ]
then
    git checkout ${BRANCH_NAME}
    retVal=$?
    if [ $retVal -ne 0 ]
    then
        jq  --null-input \
            --arg tag "$RELEASE_TAG" \
            --arg branch "$BRANCH_NAME" \
            '{"status": "missing_branch",  "data": { "error":"Can not find branch!","BRANCH_NAME": $branch,"RELEASE_TAG": $tag }}'
        deleteRepo
        exit 0
    fi
fi

git tag -l "${RELEASE_TAG}"

if [ "${DRY_RUN}" == "no" ]
then
    git push origin ${RELEASE_TAG}
    retVal=$?
else
    git push --dry_run origin ${RELEASE_TAG}
    retVal=$?
fi

deleteRepo

jq  --null-input \
        --arg tag "$RELEASE_TAG" \
        --arg branch "$BRANCH_NAME" \
        '{"status": "tag",  "data": { "BRANCH_NAME": $branch,"RELEASE_TAG": $tag }}'
    

