package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/bluele/slack"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

type HandlerConfigOption struct {
	Value string
	Path  string
	Env   string
}

type HandlerConfig struct {
	SlackWebhookUrl HandlerConfigOption
	SlackChannel    HandlerConfigOption
	SlackUsername   HandlerConfigOption
	SlackIconUrl    HandlerConfigOption
	Timeout         int
	Keyspace        string
}

var (
	stdin  *os.File
	config = HandlerConfig{
		// default values
		SlackWebhookUrl: HandlerConfigOption{Path: "webhook-url", Env: "SENSU_SLACK_WEHBOOK_URL"},
		SlackChannel:    HandlerConfigOption{Path: "channel", Env: "SENSU_SLACK_CHANNEL"},
		SlackUsername:   HandlerConfigOption{Path: "username", Env: "SENSU_SLACK_USERNAME"},
		SlackIconUrl:    HandlerConfigOption{Path: "icon-url", Env: "SENSU_SLACK_ICON_URL"},
		Timeout:         10,
		Keyspace:        "sensu.io/plugins/slack/config",
	}
	options = []*HandlerConfigOption{
		// iterable slice of user-overridable configuration options
		&config.SlackWebhookUrl,
		&config.SlackChannel,
		&config.SlackUsername,
		&config.SlackIconUrl,
	}
)

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensu-slack-handler",
		Short: "The Sensu Go Slack handler for notifying a channel",
		RunE:  run,
	}

	/*
		Sensitive flags
		default to using envvar value
		do not mark as required
		manually test for empty value
	*/
	cmd.Flags().StringVarP(&config.SlackWebhookUrl.Value,
		"webhook-url",
		"w",
		os.Getenv("SLACK_WEBHOOK_URL"),
		"The webhook url to send messages to, defaults to value of SLACK_WEBHOOK_URL env variable")

	cmd.Flags().StringVarP(&config.SlackChannel.Value,
		"channel",
		"c",
		"#general",
		"The channel to post messages to")

	cmd.Flags().StringVarP(&config.SlackUsername.Value,
		"username",
		"u",
		"sensu",
		"The username that messages will be sent as")

	cmd.Flags().StringVarP(&config.SlackIconUrl.Value,
		"icon-url",
		"i",
		"http://s3-us-west-2.amazonaws.com/sensuapp.org/sensu.png",
		"A URL to an image to use as the user avatar")

	cmd.Flags().IntVarP(&config.Timeout,
		"timeout",
		"t",
		10,
		"The amount of seconds to wait before terminating the handler")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return errors.New("invalid argument(s) received")
	}

	// load & parse stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err.Error())
	}
	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", eventJSON)
	}

	// configuration validation & overrides
	if config.SlackWebhookUrl.Value == "" {
		_ = cmd.Help()
		return fmt.Errorf("webhook url is empty")
	}

	configurationOverrides(&config, options, event)

	if err = validateEvent(event); err != nil {
		return errors.New(err.Error())
	}

	if err = sendMessage(event); err != nil {
		return errors.New(err.Error())
	}

	return nil
}

func configurationOverrides(config *HandlerConfig, options []*HandlerConfigOption, event *types.Event) {
	if config.Keyspace == "" {
		return
	}
	for _, opt := range options {
		if opt.Path != "" {
			// compile the Annotation keyspace to look for configuration overrides
			k := path.Join(config.Keyspace, opt.Path)
			switch {
			case event.Check.Annotations[k] != "":
				opt.Value = event.Check.Annotations[k]
				log.Printf("Overriding default handler configuration with value of \"Check.Annotations.%s\" (\"%s\")\n", k, event.Check.Annotations[k])
			case event.Entity.Annotations[k] != "":
				opt.Value = event.Entity.Annotations[k]
				log.Printf("Overriding default handler configuration with value of \"Entity.Annotations.%s\" (\"%s\")\n", k, event.Entity.Annotations[k])
			}
		}
	}
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

func validateEvent(event *types.Event) error {
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
