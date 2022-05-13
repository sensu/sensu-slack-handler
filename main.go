package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bluele/slack"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu-community/sensu-plugin-sdk/templates"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// HandlerConfig contains the Slack handler configuration
type HandlerConfig struct {
	sensu.PluginConfig
	slackwebHookURL          string
	slackChannel             string
	slackUsername            string
	slackIconURL             string
	slackDescriptionTemplate string
	slackAlertCritical       bool
}

const (
	webHookURL          = "webhook-url"
	channel             = "channel"
	username            = "username"
	iconURL             = "icon-url"
	descriptionTemplate = "description-template"
	alertCritical       = "alert-on-critical"

	defaultChannel       = "#general"
	defaultIconURL       = "https://www.sensu.io/img/sensu-logo.png"
	defaultUsername      = "sensu"
	defaultTemplate      = "{{ .Check.Output }}"
	defaultAlert    bool = false
)

var (
	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-slack-handler",
			Short:    "The Sensu Go Slack handler for notifying a channel",
			Keyspace: "sensu.io/plugins/slack/config",
		},
	}

	slackConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      webHookURL,
			Env:       "SLACK_WEBHOOK_URL",
			Argument:  webHookURL,
			Shorthand: "w",
			Secret:    true,
			Usage:     "The webhook url to send messages to",
			Value:     &config.slackwebHookURL,
		},
		{
			Path:      channel,
			Env:       "SLACK_CHANNEL",
			Argument:  channel,
			Shorthand: "c",
			Default:   defaultChannel,
			Usage:     "The channel to post messages to",
			Value:     &config.slackChannel,
		},
		{
			Path:      username,
			Env:       "SLACK_USERNAME",
			Argument:  username,
			Shorthand: "u",
			Default:   defaultUsername,
			Usage:     "The username that messages will be sent as",
			Value:     &config.slackUsername,
		},
		{
			Path:      iconURL,
			Env:       "SLACK_ICON_URL",
			Argument:  iconURL,
			Shorthand: "i",
			Default:   defaultIconURL,
			Usage:     "A URL to an image to use as the user avatar",
			Value:     &config.slackIconURL,
		},
		{
			Path:      descriptionTemplate,
			Env:       "SLACK_DESCRIPTION_TEMPLATE",
			Argument:  descriptionTemplate,
			Shorthand: "t",
			Default:   defaultTemplate,
			Usage:     "The Slack notification output template, in Golang text/template format",
			Value:     &config.slackDescriptionTemplate,
		},
		{
			Path:      alertCritical,
			Env:       "SLACK_ALERT_CRITICAL",
			Argument:  alertCritical,
			Shorthand: "a",
			Default:   defaultAlert,
			Usage:     "The Slack notification will alert the channel with @channel",
			Value:     &config.slackAlertCritical,
		},
	}
)

func main() {
	goHandler := sensu.NewGoHandler(&config.PluginConfig, slackConfigOptions, checkArgs, sendMessage)
	goHandler.Execute()
}

func checkArgs(_ *corev2.Event) error {
	// Support deprecated environment variables
	if webhook := os.Getenv("SENSU_SLACK_WEBHOOK_URL"); webhook != "" {
		config.slackwebHookURL = webhook
	}
	if channel := os.Getenv("SENSU_SLACK_CHANNEL"); channel != "" && config.slackChannel == defaultChannel {
		config.slackChannel = channel
	}
	if username := os.Getenv("SENSU_SLACK_USERNAME"); username != "" && config.slackUsername == defaultUsername {
		config.slackUsername = username
	}
	if icon := os.Getenv("SENSU_SLACK_ICON_URL"); icon != "" && config.slackIconURL == defaultIconURL {
		config.slackIconURL = icon
	}
	if alert := os.Getenv("SENSU_SLACK_ALERT_CRITICAL"); alertCritical != "" && !config.slackAlertCritical {
		config.slackAlertCritical, _ = strconv.ParseBool(alert)
	}

	if len(config.slackwebHookURL) == 0 {
		return fmt.Errorf("--%s or SLACK_WEBHOOK_URL environment variable is required", webHookURL)
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
		if config.slackAlertCritical {
			return "<!channel> Critical"
		} else {
			return "Critical"
		}
	default:
		return "Warning"
	}
}

func messageAttachment(event *corev2.Event) *slack.Attachment {
	description, err := templates.EvalTemplate("description", config.slackDescriptionTemplate, event)
	if err != nil {
		fmt.Printf("%s: Error processing template: %s", config.PluginConfig.Name, err)
	}

	description = strings.Replace(description, `\n`, "\n", -1)
	attachment := &slack.Attachment{
		Title:    "Description",
		Text:     description,
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
	hook := slack.NewWebHook(config.slackwebHookURL)
	err := hook.PostMessage(&slack.WebHookPostPayload{
		Attachments: []*slack.Attachment{messageAttachment(event)},
		Channel:     config.slackChannel,
		IconUrl:     config.slackIconURL,
		Username:    config.slackUsername,
	})
	if err != nil {
		return fmt.Errorf("Failed to send Slack message: %v", err)
	}

	// FUTURE: send to AH
	fmt.Printf("Notification sent to Slack channel %s\n", config.slackChannel)

	return nil
}
