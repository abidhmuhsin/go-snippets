package connection

import "abidhmuhsin.com/go-snippets/rabbitmq/worker-events/rabbitmq"

func (c *Conn) RabbitMq() *rabbitmq.RabbitMq {
	if c.rabbitMq == nil {
		c.rabbitMq = rabbitmq.New()
	}

	return c.rabbitMq
}
