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
      echo 'cmd: {{ .list.output }}'
  - name: ps
    cmd: |
      ps aux | grep {{ .show.pid }}
  - name: curl
    retries: 3
    interval: 1s
    cmd: |
      curl -I $REQUEST_PATH
```

You can also use Go's template syntax only within a cmd entry, as in `echo 'cmd: {{ .list.output }}'`.

In the above example, `step.name = show` uses the `output` information of `step.name = list`, and `step.name = ps` uses the `pid` information of `step.name = show`.

### Executing Command

Executing command is easy.

Simply execute the run subcommand specifying the configuration file using the c, config option as shown below.

```bash
sheryl run -c docs/step.yaml -o simple
```

<details>
<summary>result</summary>

```bash
sheryl run -c docs/step.yaml                                                                                                                                                                    【 main 】
【list】is success
======================================================================================
name: list
cmd: ls -A1
output: .git
.github
.gitignore
.idea
LICENSE
Makefile
README.md
bin
cmd
docs
go.mod
go.sum
pkg
======================================================================================
【show】is success
======================================================================================
name: show
cmd: echo 'cmd: .git
.github
.gitignore
.idea
LICENSE
Makefile
README.md
bin
cmd
docs
go.mod
go.sum
pkg
'
output: cmd: .git
.github
.gitignore
.idea
LICENSE
Makefile
README.md
bin
cmd
docs
go.mod
go.sum
pkg
======================================================================================
【ps】is success
======================================================================================
name: ps
cmd: ps aux | grep 47835
output: hagarihayato     47838   0.0  0.0 410724048   1200 s023  R+    8:02AM   0:00.00 grep 47835
hagarihayato     47836   0.0  0.0 410733808   1904 s023  S+    8:02AM   0:00.00 sh -c ps aux | grep 47835\012
======================================================================================
【curl】is success
======================================================================================
name: curl
cmd: curl -I $REQUEST_PATH
output:   % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0   220    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
HTTP/1.1 301 Moved Permanently
Location: https://www.google.com/
Content-Type: text/html; charset=UTF-8
Content-Security-Policy-Report-Only: object-src 'none';base-uri 'self';script-src 'nonce-GH6XtdnOGWDPwJk7JsHjeQ' 'strict-dynamic' 'report-sample' 'unsafe-eval' 'unsafe-inline' https: http:;report-uri https://csp.withgoogle.com/csp/gws/other-hp
Date: Sat, 04 Jan 2025 23:02:20 GMT
Expires: Mon, 03 Feb 2025 23:02:20 GMT
Cache-Control: public, max-age=2592000
Server: gws
Content-Length: 220
X-XSS-Protection: 0
X-Frame-Options: SAMEORIGIN
Alt-Svc: h3=":443"; ma=2592000,h3-29=":443"; ma=2592000

======================================================================================
```
</details>