# Ops Tool

Ops Tool registers itself as a slash command in Mattermost and allow users to define many subcommands and integrations only by using configuration files and bash scripts.

## Terminology

| Term | Description |
| :--  | :-- |
| __Slash Command__ | Messages that begin with / are interpreted as slash commands. The commands will send an HTTP POST request to a web service, and process a response back to Mattermost. ie `/ops` |
| __Incoming Hook__ | Mattermost incoming webhook which allows sending messages to channels and users. |
| __Intearctive Dialog__ | Integrations open dialogs by sending an HTTP POST, containing some data in the request body, to an endpoint on the Mattermost server. |
| __Provider__ | Command group. It allows logical seperation of commands. It is the first argument of the slash command. ie `/ops docker`, `/ops fun` |
| __Command__ | Configurable provider command. It is an alias of one shell script. It is the second argument of slash command. ie `/ops docker ps/`, `/ops docker image`, `/ops fun joke`. |
| __Sub Command__ | Provider command selector. By using subcommand it is possible to enforce users to provide predefined arguments. OpsTool treat them as seperate commands, and subcommand will not pass to the shell scripts as arguments. ie `/ops docker image list`. `/ops docker image delete`.`/ops docker image` command will be undefined. | 
| __Arguments__ | Anything provided after command or subcommand will be passed to the shell script as command line argument. ie `/ops docker ps` vs. `/ops docker ps -a`, `/ops gitlab version check` |
| __Scheduler__ | Cron base schedule executor to execute any command by predefined periods. |

## Overview

Ops Tool provide a mechanism to respond to Mattermost slash command webhooks. Whenever a request is received, it is processed by following logic.

1. Mattermost will send request to `/hook` endpoint whenever user trigger a slash command.
2. Controller will search config array and find mapped provider. If no provider found, it will send back error response.
3. Controller will search provider commands and find mapped command. If no command found, send back error response.
4. Controller will check command permissions for triggering user. If user is not allowed, send back forbidden response.
5. Controller will check command dialog requirements, if dialog is configured before command processing controller will send dialog request. Skip to `step 14` for dialog flows
6. Prepare request model from configuration and slash command arguments.
7. Controller will pass request mode to Handler script as environment variables and command line arguments. Controller will execute script in bash shell. 
8. Handler script should exit with `0` status and dump response model in json format in the output.
Sample: 
    ```json
    {
        "status":"ok",
        "data":[
            {"id":"12", "name":"test"},
            {"id":"12", "name":"test"},
        ]
    }
    ```
9. Ops Tool controller, will read handler script's output and builds an `OpsCommandOutput` object.
10. If response message is not needed, controller will return http status 200. (Determined by command configuration)
11. Controller will pass response model(`OpsCommandOutput`) to renderer. Renderer will render template with model and produce a response message.
12. Controller will determine message color by using status field. (Colors are determined by configuration)
13. Controller will return attachment message in the response with status code 200.

#### Interactive Dialog Flow

14. Controller will build Interactive Dialog Request and create a dialog session.
15. Controller will send request for to [Mattermost](https://developers.mattermost.com/integrate/admin-guide/admin-interactive-dialogs/) with cancel notification. Ends response with status code 300.
16. Mattermost will create a request to `/dialog` endpoint when user clicks on `Submit` or `Cancel` button.
17. Controller will determine session by using `CallbackID` identifier. If no session found, it will send notification to user to retry.(Ops Tool restart case, sessions are not persistent.)
18. Controller will remove session from session map.
19. If user did cancel the flow, controller will stop processing the request.
20. Controller will get command from session data and process the command.(Steps 6,7,8,9)
21. If response is needed, controller will send response in incoming webhook.


## Configuration Files

OpsTool has following configuration files and folders.

| Configuration | Format | Description |
| :-- | :-- | :--| 
| config/`config.yaml` | `YAML` | Main configuration file, hold tool general configuration, scheduled commands and provider configurations. |
| commands/`**/*.yaml` | `YAML` | Definitions of providers, commands and sub-commands are located in this folder|
| scripts/`**/*.sh` | 'Executable Script` | All scripts are located in this folder. |

## Playbooks

### How to create a provider

1. Create a dedicated folder under scripts and store it's scripts into that folder.
2. Create a dedicated folder under commands and store it's command definitions in that folder.
3. Configure provider commands in provider configuration file.
3. Add provider record in main configuration.
4. Restart ops tool.

### How to create a command

1. Store command scripts into provider script folder.
2. Create command definition configuration file and store in commands folder.
3. Add command configuration to provider file.
4. Restart ops tool.

## Configuration & Script Files

* [Main Configuration](./main_configuration.md)
* [Provider Configuration](./provider_configuration.md)
* [Command Configuration](./command_configuration.md)
* [Handler Script](./handler_script.md)
