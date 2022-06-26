# Main Configuration

Ops Tool searches for `config.yaml` file at current folder or under `config` directory. Following configuration properties are supported:

## Configuration File Structure

```yaml
listen: "0.0.0.0:8080"
token: "MM_SLASH_COMMAND_TOKEN"           
commands:
  - commands/docker/docker.yaml                
  - commands/fun/fun.yaml
scheduler:
  - name: "Joke"
    channel: "fun"
    provider: "fun"
    command: "joke"
    args:
      - "random"
    cron: "*/5 * * * *"
    hook: "MM_INCOMING_HOOK_URL"
```

### Configuration
| Name | Type | Description |
| :--  | :--  | :--         |
| listen | string | OpsTool creates a HTTP socket and listens on that ip and port. ie. `0.0.0.0:8080` |
| token | string | Mattermost slash command token. Details are [here](https://developers.mattermost.com/integrate/admin-guide/admin-slash-commands/#custom-slash-command).| 
| commands | list of string | Provider configuration file path list. |
| scheduler | list of [OpsSchedule](#schedule) | Scheduled command execution definitions. |

If relative paths are used in provider configuration file paths, consider relative path of the script to the executable file. 

### Schedule
| Name | Type | Description |
| :--  | :--  | :--         |
| name | string | Scheduled command name. |
| channel | string | Channel name to send message. To send directly to the users use `@` as prefix. Ie. to send message to `super_user`, use `@super_user` as channel name.| 
| provider | string | Name of the provider. It is the first argument of slash command. Ie. for `/ops fun joke random` slash command, `fun` is provider, `joke` is command and `random` is argument. |
| args | list of string | If arguments will be passed to script provide them in this property. |
| cron | cronstring | Cron schedule definition. |
| hook | string | Mattermost incoming webhook url.Details are [here](https://developers.mattermost.com/integrate/admin-guide/admin-webhooks-incoming/#simple-incoming-webhook) |

