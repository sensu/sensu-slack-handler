package main

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFormattedEventAction(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")

	action := formattedEventAction(event)
	assert.Equal("RESOLVED", action)

	event.Check.Status = 1
	action = formattedEventAction(event)
	assert.Equal("ALERT", action)
}

func TestChomp(t *testing.T) {
	assert := assert.New(t)

	trimNewline := chomp("hello\n")
	assert.Equal("hello", trimNewline)

	trimCarriageReturn := chomp("hello\r")
	assert.Equal("hello", trimCarriageReturn)

	trimBoth := chomp("hello\r\n")
	assert.Equal("hello", trimBoth)

	trimLots := chomp("hello\r\n\r\n\r\n")
	assert.Equal("hello", trimLots)
}

func TestEventKey(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	eventKey := eventKey(event)
	assert.Equal("entity1/check1", eventKey)
}

func TestEventSummary(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Output = "disk is full"

	eventKey := eventSummary(event, 100)
	assert.Equal("entity1/check1:disk is full", eventKey)

	eventKey = eventSummary(event, 5)
	assert.Equal("entity1/check1:disk ...", eventKey)
}

func TestFormattedMessage(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Output = "disk is full"
	event.Check.Status = 1
	formattedMsg := formattedMessage(event)
	assert.Equal("ALERT - entity1/check1:disk is full", formattedMsg)
}

func TestMessageColor(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")

	event.Check.Status = 0
	color := messageColor(event)
	assert.Equal("good", color)

	event.Check.Status = 1
	color = messageColor(event)
	assert.Equal("warning", color)

	event.Check.Status = 2
	color = messageColor(event)
	assert.Equal("danger", color)
}

func TestMessageStatus(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")

	event.Check.Status = 0
	status := messageStatus(event)
	assert.Equal("Resolved", status)

	event.Check.Status = 1
	status = messageStatus(event)
	assert.Equal("Warning", status)

	event.Check.Status = 2
	status = messageStatus(event)
	assert.Equal("Critical", status)
}

func TestSendMessage(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")

	var apiStub = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		expectedBody := `{"channel":"#test","attachments":[{"color":"good","fallback":"RESOLVED - entity1/check1:","title":"Description","fields":[{"title":"Status","value":"Resolved","short":false},{"title":"Entity","value":"entity1","short":true},{"title":"Check","value":"check1","short":true}],"blocks":null}],"replace_original":false,"delete_original":false}`
		assert.Equal(expectedBody, string(body))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"ok": true}`))
		require.NoError(t, err)
	}))

	config.slackwebHookURL = apiStub.URL
	config.slackChannel = "#test"
	config.slackDescriptionTemplate = "{{ .Check.Output }}"
	err := sendMessage(event)
	assert.NoError(err)
}

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	config := HandlerConfig{}
	event := corev2.FixtureEvent("entity1", "check1")
	config.slackDescriptionTemplate = "Sensu Event Details"
	config.slackUsername = "Dummy user"
	config.slackChannel = "Test"
	config.slackIconURL = "https://www.sensu.io/img/sensu-logo.png"
	_ = os.Setenv("SLACK_WEBHOOK_URL", "http://example.com/webhook")
	config.slackwebHookURL = os.Getenv("SLACK_WEBHOOK_URL")
	assert.NoError(checkArgs(event))
}
