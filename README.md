# sheryl

## Overview

Sheryl is a CLI tool for integrating shell scripts.

To date, a variety of shell scripts and tools have been developed, each playing a sophisticated role.

With this Sheryl, you can bundle tools with such specialties to complete complex tasks quickly and easily.

The setup is a single yaml.

## Install

```bash
go install github.com/gari8/sheryl/cmd/sheryl@latest
```

## Usage
### Yaml Settings
Please, Check `docs/step.yaml`.
The contents are as follows.

Please, Set environment variables with `env`.

If there are multiple environment variables with the same name, the one described in this yaml takes precedence.

Please, See `steps`.

This very intuitive shell script allows you to easily execute each step in order from top to bottom.

```yaml
env:
  REQUEST_PATH: https://google.com
steps:
  - name: list
    cmd: |
      ls -A1
  - name: show
    cmd: |
      echo 'cmd: {{ .list.cmd }}'
  - name: ps
    cmd: |
      ps aux | grep {{ .show.pid }}
  - name: curl
    retries: 3
    interval: 1s
    cmd: |
      curl -I $REQUEST_PATH
```

You can also use Go's template syntax only within a cmd entry, as in `echo 'cmd: {{ .list.cmd }}'`.

In the above example, `step.name = show` uses the cmd entry of `step.name = list`, and `step.name = ps` uses the pid information of `step.name = show`.

### Executing Command

Executing command is easy.

Simply execute the run subcommand specifying the configuration file using the c, config option as shown below.

```bash
sheryl run -c docs/step.yaml -o json
```
