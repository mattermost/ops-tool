# Provider Configuration

Providers are used for command groupping. Those:

1. Groups commands under one provider.
2. Providers can be enabled only for some users.(Teams are not supported yet.)
3. Group level environment variables can be defined and they will be valid for all commands in the group.
4. Multiple providers and/or commands can be defined in same file. (It is yaml array.)
5. To return list of supported commands.

## Provider Configuration File Structure

```yaml
- command: "gitlab"
  name: "GitLab Commands"
  description: "Get supported Gitlab Commands"
  vars:
    - name: GITLAB_URL
      value: https://git.internal.mattermost.com
    - name: GITLAB_TOKEN
      value: TOKEN
  provides:
    - commands/gitlab/gitlab_version.yaml
    - commands/gitlab/gitlab_runners.yaml
    - commands/gitlab/gitlab_pipelines.yaml
  users:
    - JulienTant
    - phoinixgrr
    - pfltdv


```

### Configuration
| Name | Type | Description |
| :--  | :--  | :--         |
| command | string | Provider slash command name. `/ops gitlab` |
| name | string | Provider name which is used at help commands and internal logs.| 
| description | string | Provider description. It is used at help command output |
| vars | list of [OpsVariable](#opsvariable) | Hard-coded bash environment variables that will be passed to handler scripts. |
| provides | list of string | Command configuration file path list. |
| users | list of nickname | If `users` property defined, only allowed users can see and execute commands of this provider.  |

If relative paths are used in command configuration file paths, consider relative path of the script to the executable file. 

