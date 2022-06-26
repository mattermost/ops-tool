# Handler Scripts

Handler scripts are bash scripts which builds response model and dumps that model into the stdout. 

## Response Model

| Name | Type | Description |
| :--  | :--  | :--         |
| status | string, required | Status of the script execution. It will be used by template and color determination. |
| data | JSON object/Array | It holds response data to be used by template. It can be value, object or array of objects/values. | 


## Sample Script

```bash
#!/bin/bash

STATUS="cut"
BASE_NAME=$(basename "$REPO_URL" ".git")

function deleteRepo(){
    cd ..
    rm -rf "${BASE_NAME}"
}

git clone "$REPO_URL"
cd "${BASE_NAME}"

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

git push
retVal=$?
deleteRepo
    
if [ $retVal -ne 0 ]
then
    jq  --null-input \
        --arg branch "$BRANCH_DEST" \
        --arg branch_mobile "$BRANCH_DEST_MOBILE" \
        --arg version "$SEMVER_RELEASE" \
        '{"status": "commit_error",  "data": { "error":"Please check repository permissions.", "SEMVER_RELEASE": $version,"BRANCH_DEST": $branch,"BRANCH_DEST_MOBILE": $branch_mobile }}'
    exit 0
fi

jq  --null-input \
    --arg status "$STATUS" \
    --arg branch "$BRANCH_DEST" \
    --arg branch_mobile "$BRANCH_DEST_MOBILE" \
    --arg version "$SEMVER_RELEASE" \
    '{"status": $status, "data": { "SEMVER_RELEASE": $version,"BRANCH_DEST": $branch,"BRANCH_DEST_MOBILE": $branch_mobile }}'
```

## Sample Output

```shell
$ export BRANCH_DEST=release-7.1 
$ export SEMVER_RELEASE=7.1.0 
$ export BRANCH_DEST_MOBILE=release-1.53 
$ export REPO_URL=git@github.com:/mattermost/cut-release-branch.git 
$ ./scripts/release/cut-release-branch.sh 
{
  "status": "cut",
  "data": {
    "SEMVER_RELEASE": "7.1.0",
    "BRANCH_DEST": "release-7.1",
    "BRANCH_DEST_MOBILE": "release-1.53"
  }
}
$
```
