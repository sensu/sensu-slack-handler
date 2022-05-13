[![Bonsai Asset Badge](https://img.shields.io/badge/Sensu%20Slack%20Handler-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-slack-handler) [![Build Status](https://travis-ci.org/sensu/sensu-slack-handler.svg?branch=master)](https://travis-ci.org/sensu/sensu-slack-handler)

# Sensu Slack Handler

- [Overview](#overview)
- [Usage examples](#usage-examples)
  - [Help output](#help-output)
  - [Environment variables](#environment-variables)
  - [Templates](#templates)
  - [Annotations](#annotations)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
  - [Check definition](#check-definition)
- [Installation from source and contributing](#installation-from-source-and-contributing)

## Overview


The [Sensu Slack Handler][0] is a [Sensu Event Handler][3] that sends event data
to a configured Slack channel.

## Usage examples

### Help output

Help:

```
Usage:
  sensu-slack-handler [flags]
  sensu-slack-handler [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -a, --alert-on-critical             The Slack notification will alert the channel with @channel
  -c, --channel string                The channel to post messages to (default "#general")
  -t, --description-template string   The Slack notification output template, in Golang text/template format (default "{{ .Check.Output }}")
  -h, --help                          help for sensu-slack-handler
  -i, --icon-url string               A URL to an image to use as the user avatar (default "https://www.sensu.io/img/sensu-logo.png")
  -u, --username string               The username that messages will be sent as (default "sensu")
  -w, --webhook-url string            The webhook url to send messages to
```

### Environment variables

|Argument               |Environment Variable       |
|-----------------------|---------------------------|
|--webhook-url          |SLACK_WEBHOOK_URL          |
|--channel              |SLACK_CHANNEL              |
|--username             |SLACK_USERNAME             |
|--icon-url             |SLACK_ICON_URL             |
|--description-template |SLACK_DESCRIPTION_TEMPLATE |
|--alert-on-critical    |SENSU_SLACK_ALERT_CRITICAL |


**Security Note:** Care should be taken to not expose the webhook URL for this handler by specifying it
on the command line or by directly setting the environment variable in the handler definition.  It is
suggested to make use of [secrets management][7] to surface it as an environment variable.  The
handler definition above references it as a secret.  Below is an example secrets definition that make
use of the built-in [env secrets provider][8].

```yml
---
type: Secret
api_version: secrets/v1
metadata:
  name: slack-webhook-url
spec:
  provider: env
  id: SLACK_WEBHOOK_URL
```

### Templates

This handler provides options for using templates to populate the values
provided by the event in the message sent via Slack. More information on
template syntax and format can be found in [the documentation][9]


### Annotations

All arguments for this handler are tunable on a per entity or check basis based
on annotations. The annotations keyspace for this handler is
`sensu.io/plugins/slack/config`.

**NOTE**: Due to [check token substituion][10], supplying a template value such
as for `description-template` as a check annotation requires that you place the
desired template as a [golang string literal][11] (enlcosed in backticks)
within another template definition.  This does not apply to entity annotations.

#### Examples

To customize the channel for a given entity, you could use the following
sensu-agent configuration snippet:

```yml
# /etc/sensu/agent.yml example
annotations:
  sensu.io/plugins/slack/config/channel: '#monitoring'
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

### Proxy Support

This handler supports the use of the environment variables HTTP_PROXY,
HTTPS_PROXY, and NO_PROXY (or the lowercase versions thereof). HTTPS_PROXY takes
precedence over HTTP_PROXY for https requests.  The environment values may be
either a complete URL or a "host[:port]", in which case the "http" scheme is assumed.

## Installing from source and contributing

Download the latest version of the sensu-slack-handler from [releases][4],
or create an executable from this source.

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
[7]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/
[8]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/#use-env-for-secrets-management
[9]: https://docs.sensu.io/sensu-go/latest/observability-pipeline/observe-process/handler-templates/
[10]: https://docs.sensu.io/sensu-go/latest/observability-pipeline/observe-schedule/checks/#check-token-substitution
[11]: https://golang.org/ref/spec#String_literals
