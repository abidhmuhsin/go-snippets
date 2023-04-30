package worker

import (
	"encoding/json"
	"log"

	"abidhmuhsin.com/go-snippets/rabbitmq/worker-events/connection"
	rabbitmq "abidhmuhsin.com/go-snippets/rabbitmq/worker-events/rabbitmq"
)

// Define queue name string -- used to access queue.
const AsyncJobABCQueue rabbitmq.QueueName = "createEvent"

// Define the payload type using an imported refrence type from the job's implementation package.
// type AsyncJobABCPayload = abcJobService.CreatePayload
// Using a local type here
type AsyncJobABCPayload struct {
	Name    string
	Counter int
}

// Publisher method. Used to publish a new job for given Queue (AsyncJobABCQueue)
func CreateAsyncJobABC(conn *connection.Conn, payload *AsyncJobABCPayload) error {
	return publish(conn, AsyncJobABCQueue, payload)
}

// Listner method. Used to execute any new jobs received for given Queue (AsyncJobABCQueue) in the worker.
// May call any external service implementations to process the given job as necessary.
func createAsyncJobABC(conn *connection.Conn, b []byte) error {
	payload := &AsyncJobABCPayload{}

	err := json.Unmarshal(b, payload)
	if err != nil {
		return err
	}
	// Handle the job via job's implementation package.
	// err = abcJobService.PerformJob(conn, payload)
	// if err != nil {
	// 	return err
	// }
	log.Println("Received asyncjobABC", payload)

	return nil
}
