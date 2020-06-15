[![Bonsai Asset Badge](https://img.shields.io/badge/Sensu%20Slack%20Handler-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-slack-handler) [![Build Status](https://travis-ci.org/sensu/sensu-slack-handler.svg?branch=master)](https://travis-ci.org/sensu/sensu-slack-handler)

# Sensu Slack Handler

- [Overview](#overview)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
  - [Check definition](#check-definition)
- [Installation from source and
  contributing](#installation-from-source-and-contributing)

## Overview


The [Sensu Slack Handler][0] is a [Sensu Event Handler][3] that sends event data
to a configured Slack channel.

## Usage examples

Help:

```
Usage:
  sensu-slack-handler [flags]
  sensu-slack-handler [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -c, --channel string                The channel to post messages to (default "#general")
  -t, --description-template string   The Slack notification output template, in Golang text/template format (default "{{ .Check.Output }}")
  -h, --help                          help for sensu-slack-handler
  -i, --icon-url string               A URL to an image to use as the user avatar (default "https://www.sensu.io/img/sensu-logo.png")
  -u, --username string               The username that messages will be sent as (default "sensu")
  -w, --webhook-url string            The webhook url to send messages to
```

## Configuration

### Asset registration

Assets are the best way to make use of this handler. If you're not using an asset, please consider doing so! If you're using sensuctl 5.13 or later, you can use the following command to add the asset:

`sensuctl asset add sensu/sensu-slack-handler`

If you're using an earlier version of sensuctl, you can download the asset
definition from [this project's Bonsai Asset Index
page][6].

### Handler definition

Create the handler using the following handler definition:

```yml
---
api_version: core/v2
type: Handler
metadata:
  namespace: default
  name: slack
spec:
  type: pipe
  command: sensu-slack-handler --channel '#general' --username 'sensu'
  filters:
  - is_incident
  runtime_assets:
  - sensu/sensu-slack-handler
  secrets:
  - name: SLACK_WEBHOOK_URL
    secret: slack-webhook-url
  timeout: 10
```
**Note**: The library used in the Sensu SDK for this plugin requires that if your Slack webhook URL is listed as an environment variable, the URL cannot be surrounded by quotes. 

**Security Note**: The Slack webhook URL should always be treated as a security
sensitive configuration option and in this example, it is loaded into the
handler configuration as an environment variable using a [secret][5]. Command
arguments are commonly readable from the process table by other unprivaledged
users on a system (ex: ps and top commands), so it's a better practise to read
in sensitive information via environment variables or configuration files on
disk. The --webhook-url flag is provided as an override for testing purposes.

### Check definition

```
api_version: core/v2
type: CheckConfig
metadata:
  namespace: default
  name: dummy-app-healthz
spec:
  command: check-http -u http://localhost:8080/healthz
  subscriptions:
  - dummy
  publish: true
  interval: 10
  handlers:
  - slack
```

### Customizing configuration options via checks and entities

All configuration options of this handler can be overridden via the annotations
of checks and entities. For example, to customize the channel for a given
entity, you could use the following sensu-agent configuration snippet:

```yml
# /etc/sensu/agent.yml example
annotations:
  sensu.io/plugins/slack/config/channel: '#monitoring'
```

### Proxy Support

This handler supports the use of the environment variables HTTP_PROXY,
HTTPS_PROXY, and NO_PROXY (or the lowercase versions thereof). HTTPS_PROXY takes
precedence over HTTP_PROXY for https requests.  The environment values may be
either a complete URL or a "host[:port]", in which case the "http" scheme is assumed.

## Installing from source and contributing

Download the latest version of the sensu-slack-handler from [releases][4],
or create an executable script from this source.

### Compiling

From the local path of the sensu-slack-handler repository:
```
go build
```

To contribute to this plugin, see [CONTRIBUTING](https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md)

[0]: https://github.com/sensu/sensu-slack-handler
[1]: https://github.com/sensu/sensu-go
[3]: https://docs.sensu.io/sensu-go/latest/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/sensu/sensu-slack-handler/releases
[5]: https://docs.sensu.io/sensu-go/latest/reference/secrets/
[6]: https://bonsai.sensu.io/assets/sensu/sensu-slack-handler
