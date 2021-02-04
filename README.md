# Jira command(er)

This is a small utility to automate some release/tasks bookkeeping in JIRA.

## Available commands

```
Deployment management with Jira

Usage:
  jirc [command]

Available Commands:
  build       Register build in Jira
  deploy      Register deployment in Jira, update issues
  help        Help about any command
  ping        Ping Jira

Flags:
      --config string   config file (default is $HOME/.jirc.yaml)
  -h, --help            help for jirc
  -j, --jira string     Jira base URL (default "http://jira.staffconfig.ru")
  -p, --pass string     Jira password
  -u, --user string     Jira username

Use "jirc [command] --help" for more information about a command.

```

### build
```
For each task adds label in form Server-full-x.x.x and creates release
Server-x.x.x or Server-app-x.x.x when build is for distinct server project.

Usage:
  jirc build -a | BUILD_NUMBER TASK [TASK]... [flags]

Flags:
  -h, --help            help for build
  -a, --show-apps-map   show applicaitons map and exit

Global Flags:
...
```

### deploy

```

Sets project release status to "released", adds deploy date to release description.
Transition release issues to "Testing" status if available and assign back to reporters
if transition reached or succeeded.

Usage:
  jirc deploy PROJECT BUILD_NUMBER [APPLICATION] [flags]

Flags:
  -h, --help              help for deploy
  -a, --no-assign-back    do not assign issues back to reporters
  -s, --skip-transition   skip issues transition to Тестирование (implies -a)

Global Flags:
...
```
  
### ping 

```
Ping jira server

Usage:
  jirc ping [flags]

Flags:
  -h, --help   help for ping

Global Flags:
...
```
