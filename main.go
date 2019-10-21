package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugins-go-library/sensu"

	"github.com/bluele/slack"
)

type HandlerConfig struct {
	sensu.PluginConfig
	SlackWebhookUrl          string
	SlackChannel             string
	SlackUsername            string
	SlackIconUrl             string
	redactMatch              string
	redact                   bool
	SlackIncludeCheckLabels  bool
	SlackIncludeEntityLabels bool
}

const (
	webHookUrl      = "webhook-url"
	channel         = "channel"
	userName        = "username"
	iconUrl         = "icon-url"
	incCheckLabels  = "include-check-labels"
	incEntityLabels = "include-entity-labels"
	redactMatch     = "redact-match"
	redact          = "redact"
)

var (
	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-slack-handler",
			Short:    "The Sensu Go Slack handler for notifying a channel",
			Timeout:  10,
			Keyspace: "sensu.io/plugins/slack/config",
		},
	}

	slackConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      webHookUrl,
			Env:       "SENSU_SLACK_WEBHOOK_URL",
			Argument:  webHookUrl,
			Shorthand: "w",
			Default:   "",
			Usage:     "The webhook url to send messages to, defaults to value of SLACK_WEBHOOK_URL env variable",
			Value:     &config.SlackWebhookUrl,
		},
		{
			Path:      redactMatch,
			Env:       "SENSU_SLACK_REDACTMATCH",
			Argument:  redactMatch,
			Shorthand: "m",
			Default:   "(?i).*(pass|key).*",
			Usage:     "Regex to redact values of matching labels",
			Value:     &config.redactMatch,
		},
		{
			Path:      redact,
			Env:       "SENSU_SLACK_REDACTMATCH",
			Argument:  redact,
			Shorthand: "r",
			Default:   false,
			Usage:     "Enable redaction of labels",
			Value:     &config.redact,
		},
		{
			Path:      channel,
			Env:       "SENSU_SLACK_CHANNEL",
			Argument:  channel,
			Shorthand: "c",
			Default:   "#general",
			Usage:     "The channel to post messages to",
			Value:     &config.SlackChannel,
		},
		{
			Path:      userName,
			Env:       "SENSU_SLACK_USERNAME",
			Argument:  userName,
			Shorthand: "u",
			Default:   "sensu",
			Usage:     "The username that messages will be sent as",
			Value:     &config.SlackUsername,
		},
		{
			Path:      iconUrl,
			Env:       "SENSU_SLACK_ICON_URL",
			Argument:  iconUrl,
			Shorthand: "i",
			Default:   "http://s3-us-west-2.amazonaws.com/sensuapp.org/sensu.png",
			Usage:     "A URL to an image to use as the user avatar",
			Value:     &config.SlackIconUrl,
		},
		{
			Path:      incCheckLabels,
			Env:       "SENSU_SLACK_INCLUDE_CHECK_LABELS",
			Argument:  incCheckLabels,
			Shorthand: "l",
			Default:   false,
			Usage:     "Include check labels in slack message?",
			Value:     &config.SlackIncludeCheckLabels,
		},
		{
			Path:      incEntityLabels,
			Env:       "SENSU_SLACK_INCLUDE_ENTITY_LABELS",
			Argument:  incEntityLabels,
			Shorthand: "e",
			Default:   false,
			Usage:     "Include entity labels in slack message?",
			Value:     &config.SlackIncludeEntityLabels,
		},
	}
)

func main() {
	goHandler := sensu.NewGoHandler(&config.PluginConfig, slackConfigOptions, checkArgs, sendMessage)
	goHandler.Execute()
}

func checkArgs(_ *corev2.Event) (e error) {
	if len(config.SlackWebhookUrl) == 0 {
		return fmt.Errorf("--webhook-url or SENSU_SLACK_WEBHOOK_URL environment variable is required")
	}

	// validate the regex compiles, if not catch the panic and return error
	defer func() {
		if r := recover(); r != nil {
			e = fmt.Errorf("regexp (%s) specified by SENSU_SLACK_REDACT or --redact is invalid", config.redactMatch)
		}
		return
	}()
	regexp.MustCompile(config.redactMatch)
	return
}

func formattedEventAction(event *corev2.Event) string {
	switch event.Check.Status {
	case 0:
		return "RESOLVED"
	default:
		return "ALERT"
	}
}

func chomp(s string) string {
	return strings.Trim(strings.Trim(strings.Trim(s, "\n"), "\r"), "\r\n")
}

func eventKey(event *corev2.Event) string {
	return fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
}

func eventSummary(event *corev2.Event, maxLength int) string {
	output := chomp(event.Check.Output)
	if len(event.Check.Output) > maxLength {
		output = output[0:maxLength] + "..."
	}
	return fmt.Sprintf("%s:%s", eventKey(event), output)
}

func formattedMessage(event *corev2.Event) string {
	return fmt.Sprintf("%s - %s", formattedEventAction(event), eventSummary(event, 100))
}

func attachCheckLabels(event *corev2.Event, attachment *slack.Attachment, config HandlerConfig) {
	re := regexp.MustCompile(config.redactMatch)
	if event.Check.Labels == nil {
		return
	}

	buf := bytes.Buffer{}
	for k, v := range event.Check.Labels {
		if config.redact && re.MatchString(k) {
			v = "**REDACTED**"
		}
		fmt.Fprintf(&buf, "%s=%s\n", k, v)
	}

	attachment.Fields = append(attachment.Fields, &slack.AttachmentField{
		Title: "Check Labels",
		Value: buf.String(),
		Short: false,
	})

	return
}

func attachEntityLabels(event *corev2.Event, attachment *slack.Attachment, config HandlerConfig) {
	re := regexp.MustCompile(config.redactMatch)
	if event.Entity.Labels == nil {
		return
	}

	buf := bytes.Buffer{}
	for k, v := range event.Entity.Labels {
		if config.redact && re.MatchString(k) {
			v = "**REDACTED**"
		}
		fmt.Fprintf(&buf, "%s=%s\n", k, v)
	}

	attachment.Fields = append(attachment.Fields, &slack.AttachmentField{
		Title: "Entity Labels",
		Value: buf.String(),
		Short: false,
	})

	return
}

func messageColor(event *corev2.Event) string {
	switch event.Check.Status {
	case 0:
		return "good"
	case 2:
		return "danger"
	default:
		return "warning"
	}
}

func messageStatus(event *corev2.Event) string {
	switch event.Check.Status {
	case 0:
		return "Resolved"
	case 2:
		return "Critical"
	default:
		return "Warning"
	}
}

func messageAttachment(event *corev2.Event) *slack.Attachment {
	attachment := &slack.Attachment{
		Title:    "Description",
		Text:     event.Check.Output,
		Fallback: formattedMessage(event),
		Color:    messageColor(event),
		Fields: []*slack.AttachmentField{
			{
				Title: "Status",
				Value: messageStatus(event),
				Short: false,
			},
			{
				Title: "Entity",
				Value: event.Entity.Name,
				Short: true,
			},
			{
				Title: "Check",
				Value: event.Check.Name,
				Short: true,
			},
		},
	}

	if config.SlackIncludeEntityLabels {
		attachEntityLabels(event, attachment, config)
	}

	if config.SlackIncludeCheckLabels {
		attachCheckLabels(event, attachment, config)
	}

	return attachment
}

func sendMessage(event *corev2.Event) error {
	hook := slack.NewWebHook(config.SlackWebhookUrl)
	return hook.PostMessage(&slack.WebHookPostPayload{
		Attachments: []*slack.Attachment{messageAttachment(event)},
		Channel:     config.SlackChannel,
		IconUrl:     config.SlackIconUrl,
		Username:    config.SlackUsername,
	})
}
