package main

import (
	"abidhmuhsin.com/go-snippets/rabbitmq/worker-events/connection"
	worker "abidhmuhsin.com/go-snippets/rabbitmq/worker-events/worker"
)

func main() {
	// Start worker. has signal.ExitWithSignal() inside which blocks main thread and keeps listening
	conn := &connection.Conn{}
	worker.Run(conn, []string{string(worker.AsyncJobABCQueue), string(worker.SendWebhookQueue)})

}
