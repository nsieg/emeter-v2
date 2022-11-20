package main

import (
	"io"
	"net/http"
	"testing"

	resty "github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
)

func Test_Send(t *testing.T) {
	// Given
	msg := "Hello"
	botId := "123"
	chatId := "456"
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	requestBody := ""

	httpmock.RegisterResponder("POST", "https://api.telegram.org/bot123/sendMessage",
		func(req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			requestBody = string(b)
			resp, _ := httpmock.NewJsonResponse(200, "{}")
			return resp, nil
		},
	)

	// When
	send(msg, botId, chatId, *client)

	// Then
	if requestBody != `{"chat_id":"456","text":"Hello"}` {
		t.Error("Wrong request body!", requestBody)
	}
	if httpmock.GetTotalCallCount() != 1 {
		t.Error("URL was called != 1 times!")
	}
}

func Test_Send_Manual(t *testing.T) {
	t.Skip("Skip manual test")

	// Given
	msg := "Test"
	botId := ""
	chatId := ""
	client := resty.New()

	// When
	send(msg, botId, chatId, *client)

	// Then
	// Manually check response in Telegram
}
