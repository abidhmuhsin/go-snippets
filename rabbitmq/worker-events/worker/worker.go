package worker

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"abidhmuhsin.com/go-snippets/rabbitmq/worker-events/connection"
	rabbitmq "abidhmuhsin.com/go-snippets/rabbitmq/worker-events/rabbitmq"
	"abidhmuhsin.com/go-snippets/rabbitmq/worker-events/signal"
)

var Queues = []*rabbitmq.Queue{
	{Name: AsyncJobABCQueue},
	// {Name: EndTrialsQueue},
	// {Name: NotifyAboutEndingTrialsQueue, Ttl: 24 * time.Hour},
	// {Name: NotifyAboutTrialMidwaysQueue, Ttl: 15 * 24 * time.Hour},
	// {Name: SendEmailQueue},
	{Name: SendWebhookQueue},
}

func Run(conn *connection.Conn, queues []string) error {
	err := conn.RabbitMq().DeclareQueues(Queues)
	if err != nil {
		return err
	}

	log.Println("Queues:")

	for _, q := range queues {
		log.Printf("* %s", q)
	}

	log.Println("Starting to listen...")

	runners := map[rabbitmq.QueueName]func(*connection.Conn, []byte) error{
		AsyncJobABCQueue: createAsyncJobABC,
		// EndTrialsQueue:               endTrials,
		// NotifyAboutEndingTrialsQueue: notifyAboutEndingTrials,
		// NotifyAboutTrialMidwaysQueue: notifyAboutTrialMidways,
		// SendEmailQueue:               sendEmail,
		SendWebhookQueue: sendWebhook,
	}

	deliveries := make(chan amqp.Delivery)

	err = conn.RabbitMq().Consume(queues, deliveries)
	if err != nil {
		panic(err)
	}

	go func() {
		for delivery := range deliveries {
			err2 := runners[rabbitmq.QueueName(delivery.RoutingKey)](conn, delivery.Body)
			if err2 != nil {
				log.Println(err2)
			}

			err2 = delivery.Ack(false)
			if err2 != nil {
				panic(err2)
			}

		}
	}()

	signal.ExitWithSignal()

	return nil
}

func publish(conn *connection.Conn, queue rabbitmq.QueueName, payload any) error {
	publishingBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	publishing := amqp.Publishing{
		Body:        publishingBody,
		ContentType: "application/json",
	}

	return conn.RabbitMq().Publish("", string(queue), false, false, publishing)
}
