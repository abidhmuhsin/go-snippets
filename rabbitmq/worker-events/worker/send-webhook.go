package worker

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"

	"abidhmuhsin.com/go-snippets/rabbitmq/worker-events/connection"
	rabbitmq "abidhmuhsin.com/go-snippets/rabbitmq/worker-events/rabbitmq"
)

// ENV variables
const (
	WEBHOOK_URL = "https://abid.dev"
)

// Define queue name string -- used to access queue.
const SendWebhookQueue rabbitmq.QueueName = "sendWebhook"

// Define constants to identify different Webhook events within a webhook
const (
	SendWebhookEventCleanupFreeUserData Event = "maintenance.cleanup_userdata"
	SendWebhookEventUserSignedUp        Event = "user.signed_up"
)

type Event string

type SendWebhookPayload struct {
	Data  any
	Event Event
}

// Publisher method. Used to publish a new job for given Queue (SendWebhookQueue)
func SendWebhook(conn *connection.Conn, payload *SendWebhookPayload) {
	if WEBHOOK_URL == "" {
		return
	}

	err := publish(conn, SendWebhookQueue, payload)
	if err != nil {
		// log/track error
		log.Println(err)
	}
}

// Listner method. Used to execute any new jobs received for given Queue (SendWebhookQueue) in the worker.
func sendWebhook(conn *connection.Conn, b []byte) error {
	payload := &SendWebhookPayload{}

	err := json.Unmarshal(b, payload)
	if err != nil {
		return err
	}

	log.Printf("Delivering webhook for event '%s'", payload.Event)

	rawDataByteSlice, err := json.Marshal(payload.Data)
	if err != nil {
		return err
	}

	rawDataMap := map[string]any{}

	err = json.Unmarshal(rawDataByteSlice, &rawDataMap)
	if err != nil {
		return err
	}

	dataMap := map[string]any{}

	for k, v := range rawDataMap {
		dataMap[strings.ToLower(k[0:1])+k[1:]] = v
	}

	bodyData := map[string]any{
		"data":  dataMap,
		"event": payload.Event,
	}

	bodyByteSlice, err := json.Marshal(bodyData)
	if err != nil {
		return err
	}

	_, err = conn.HttpClient().Post(WEBHOOK_URL, "application/json", bytes.NewBuffer(bodyByteSlice))
	if err != nil {
		return err
	}

	return nil
}
