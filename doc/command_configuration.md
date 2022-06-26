# Command Configuration

Commands configurations defines command mapping rules, runtime environment of the handler scripts, dialog definitions and response message templates.

## Provider Configuration File Structure

```yaml
- command: "cut"
  name: "Cut Release Branch"
  description: "Prepare repositories for the next release."
  vars:
    - name: REPO_URL
      value: git@git.internal.mattermost.com:ci/cut-release-branch.git
  users:
    - JulienTant
    - phoinixgrr
    - pfltdv
  dialog:
    title: Cut Release Branch
    url: https://ops-tool.cloud.mattermost.com/api/v4/actions/dialogs/open
    callbackUrl: https://3b3a-185-148-86-26.eu.ngrok.io/dialog
    hook: https://ops-tool.cloud.mattermost.com/hooks/83d13meogjbrtxkydt1mfijd9w
    introduction_text: Prepare repositories for the next release.
    elements:
      - name: BRANCH_DEST 
        title: Release Branch Name
        type: text
        subtype: text
        optional: false
        default: release-7.0
        help_text: Determines the branch name will be created.
      - name: BRANCH_DEST_MOBILE 
        title: Mobile Release Branch Name
        type: text
        subtype: text
        optional: false
        default: release-1.53
        help_text: Determines the branch name for mobile project to be created.
      - name: SEMVER_RELEASE 
        title: Server Version
        type: text
        subtype: text
        optional: false
        default: 7.0.0
        help_text: Version number.
  exec: 
    - scripts/release/cut-release-branch.sh
  response:
    type: "ephemeral"
    colors:
    - color: "#0000ff"
      status: "cut"
    - color: "#ff0000"
    template: |
      {{ if eq .Status "cut" }}
        Release Cut pipelines are successfully started!
        
      {{- else -}}
        {{ .Data.error }}

      {{ end }}
        
        * __Version__: {{ .Data.SEMVER_RELEASE }}
        * __Branch__: {{ .Data.BRANCH_DEST }}
        * __Mobile Branch__: {{ .Data.BRANCH_DEST_MOBILE }}

```

### Configuration

| Name | Type | Description |
| :--  | :--  | :--         |
| command | string | Provider slash command name. `/ops gitlab` |
| name | string | Provider name which is used at help commands and internal logs.| 
| description | string | Provider description. It is used at help command output |
| vars | list of [OpsVariable](#opsvariable) | Hard-coded bash environment variables that will be passed to handler scripts. |
| users | list of string | If `users` property defined, only allowed users can execute this command otherwise everyone can trigger this command.  |
| dialog | [Dialog](#dialog) | If defined, command handler will show a dialog to a user. |
| exec | List of string | Path of script handlers. Multiple script can be defined, ops tool will execute all of them sequentally and use output of last script. |
| response | [OpsResponse](#opsresponse) | If defined, controller will send a message to the user. |


If relative paths are used in script paths, consider relative path of the script to the executable file. 

### Dialog 
Dialog structure uses similar structure with [Interactive Dialogs](https://developers.mattermost.com/integrate/admin-guide/admin-interactive-dialogs/). 

| Name | Type | Description |
| :--  | :--  | :--         |
| title | string | Dialog Title. |
| url | string | Mattermost server's dialog open endpoint url. `/api/v4/actions/dialogs/open` | 
| callbackUrl | string | OpsTool's public dialog endpoint url. `/hook` | 
| hook | string | Mattermost's incoming webhook url. |
| introduction_text | string |Markdown-formatted introduction text which is displayed above the dialog elements.|
| elements | Elements | Up to 5 elements allowed per dialog. If none are supplied the dialog box acts as a simple confirmation. Details are [here](https://developers.mattermost.com/integrate/admin-guide/admin-interactive-dialogs/#elements) |

 ### Response

| Name | Type | Description |
| :--  | :--  | :--         |
| type | string | Response notification type. `ephemeral` or `in_channel`. If response is ephemeral, it will send back only to the user. |
| colors | [ColorSelector](#colorselector) | Color map bind to the status values. If status field is missing it will be default color.  |
| template | Go Template String | Go Template to be used to render the response from response model. Details are [here](https://pkg.go.dev/html/template) |

 ### ColorSelector

| Name | Type | Description |
| :--  | :--  | :--         |
| color | string | Hex representation of the color.`#ff0000` |
| status | status | Color activator. If status field of the response model is equal to the this value, this color will be used at the rendered message.  |

To make any color default, remove status field and add that color as last item of the array.