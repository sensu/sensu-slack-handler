package main

import (
	"errors"
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugins-go-library/sensu"
	"os"
	"strings"

	"github.com/bluele/slack"
)

type HandlerConfigOption struct {
	Value string
	Path  string
	Env   string
}

type HandlerConfig struct {
	sensu.PluginConfig
	SlackWebhookUrl HandlerConfigOption
	SlackChannel    HandlerConfigOption
	SlackUsername   HandlerConfigOption
	SlackIconUrl    HandlerConfigOption
}

var (
	stdin  *os.File
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
			Path:      "webhook-url",
			Env:       "SENSU_SLACK_WEHBOOK_URL",
			Argument:  "webhook-url",
			Shorthand: "w",
			Default:   "",
			Usage:     "The webhook url to send messages to, defaults to value of SLACK_WEBHOOK_URL env variable",
			Value:     &config.SlackWebhookUrl.Value,
		},
		{
			Path:      "channel",
			Env:       "SENSU_SLACK_CHANNEL",
			Argument:  "channel",
			Shorthand: "c",
			Default:   "#general",
			Usage:     "The channel to post messages to",
			Value:     &config.SlackChannel.Value,
		},
		{
			Path:      "username",
			Env:       "SENSU_SLACK_USERNAME",
			Argument:  "username",
			Shorthand: "u",
			Default:   "sensu",
			Usage:     "The username that messages will be sent as",
			Value:     &config.SlackUsername.Value,
		},
		{
			Path:      "icon-url",
			Env:       "SENSU_SLACK_ICON_URL",
			Argument:  "icon-url",
			Shorthand: "i",
			Default:   "http://s3-us-west-2.amazonaws.com/sensuapp.org/sensu.png",
			Usage:     "A URL to an image to use as the user avatar",
			Value:     &config.SlackIconUrl.Value,
		},
	}
)

func main() {
	goHandler, _ := sensu.NewGoHandler(&config.PluginConfig, slackConfigOptions, checkArgs, executeHandler)
	err := goHandler.Execute()
	if err != nil {
		fmt.Printf("Error executing plugin: %s", err)
	}
}

func checkArgs(event *types.Event) error {
	if len(config.SlackWebhookUrl.Value) == 0 {
		return fmt.Errorf("--webhook-url or SENSU_SLACK_WEHBOOK_URL environment variable is required")
	}

	if event.Timestamp <= 0 {
		return errors.New("timestamp is missing or must be greater than zero")
	}

	if event.Entity == nil {
		return errors.New("entity is missing from event")
	}

	if !event.HasCheck() {
		return errors.New("check is missing from event")
	}

	if err := event.Entity.Validate(); err != nil {
		return err
	}

	if err := event.Check.Validate(); err != nil {
		return err
	}

	return nil
}

func executeHandler(event *types.Event) error {
	if err := sendMessage(event); err != nil {
		return errors.New(err.Error())
	}

	return nil
}

func formattedEventAction(event *types.Event) string {
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

func eventKey(event *types.Event) string {
	return fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
}

func eventSummary(event *types.Event, maxLength int) string {
	output := chomp(event.Check.Output)
	if len(event.Check.Output) > maxLength {
		output = output[0:maxLength] + "..."
	}
	return fmt.Sprintf("%s:%s", eventKey(event), output)
}

func formattedMessage(event *types.Event) string {
	return fmt.Sprintf("%s - %s", formattedEventAction(event), eventSummary(event, 100))
}

func messageColor(event *types.Event) string {
	switch event.Check.Status {
	case 0:
		return "good"
	case 2:
		return "danger"
	default:
		return "warning"
	}
}

func messageStatus(event *types.Event) string {
	switch event.Check.Status {
	case 0:
		return "Resolved"
	case 2:
		return "Critical"
	default:
		return "Warning"
	}
}

func messageAttachment(event *types.Event) *slack.Attachment {
	attachment := &slack.Attachment{
		Title:    "Description",
		Text:     event.Check.Output,
		Fallback: formattedMessage(event),
		Color:    messageColor(event),
		Fields: []*slack.AttachmentField{
			&slack.AttachmentField{
				Title: "Status",
				Value: messageStatus(event),
				Short: false,
			},
			&slack.AttachmentField{
				Title: "Entity",
				Value: event.Entity.Name,
				Short: true,
			},
			&slack.AttachmentField{
				Title: "Check",
				Value: event.Check.Name,
				Short: true,
			},
		},
	}
	return attachment
}

func sendMessage(event *types.Event) error {
	hook := slack.NewWebHook(config.SlackWebhookUrl.Value)
	return hook.PostMessage(&slack.WebHookPostPayload{
		Attachments: []*slack.Attachment{messageAttachment(event)},
		Channel:     config.SlackChannel.Value,
		IconUrl:     config.SlackIconUrl.Value,
		Username:    config.SlackUsername.Value,
	})
}
