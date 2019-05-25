package main

import (
	"fmt"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugins-go-library/sensu"
	"strings"

	"github.com/bluele/slack"
)

type HandlerConfig struct {
	sensu.PluginConfig
	SlackWebhookUrl string
	SlackChannel    string
	SlackUsername   string
	SlackIconUrl    string
}

const (
	webHookUrl = "webhook-url"
	channel    = "channel"
	userName   = "username"
	iconUrl    = "icon-url"
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
			Env:       "SENSU_SLACK_WEHBOOK_URL",
			Argument:  webHookUrl,
			Shorthand: "w",
			Default:   "",
			Usage:     "The webhook url to send messages to, defaults to value of SLACK_WEBHOOK_URL env variable",
			Value:     &config.SlackWebhookUrl,
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
	}
)

func main() {
	goHandler := sensu.NewGoHandler(&config.PluginConfig, slackConfigOptions, checkArgs, sendMessage)
	goHandler.Execute()
}

func checkArgs(_ *corev2.Event) error {
	if len(config.SlackWebhookUrl) == 0 {
		return fmt.Errorf("--webhook-url or SENSU_SLACK_WEBHOOK_URL environment variable is required")
	}

	return nil
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
