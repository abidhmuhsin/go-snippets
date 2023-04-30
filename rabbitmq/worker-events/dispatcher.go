package main

import (
	"log"
	"time"

	"abidhmuhsin.com/go-snippets/rabbitmq/worker-events/connection"
	worker "abidhmuhsin.com/go-snippets/rabbitmq/worker-events/worker"
)

func main() {
	// Start worker. has signal.ExitWithSignal() inside which blocks main thread and keeps listening
	conn := &connection.Conn{}

	for i := 1; i <= 100; i++ {
		time.Sleep(300 * time.Millisecond)
		worker.CreateAsyncJobABC(conn,
			&worker.AsyncJobABCPayload{
				Name:    "test event",
				Counter: i})
		time.Sleep(300 * time.Millisecond)
		worker.SendWebhook(conn, &worker.SendWebhookPayload{
			Data:  map[string]interface{}{"Name": "test nAME"},
			Event: worker.SendWebhookEventUserSignedUp,
		})
		log.Println("Dispatched: ", i)
	}

}
