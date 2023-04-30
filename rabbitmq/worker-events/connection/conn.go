package connection

import (
	"net/http"

	"abidhmuhsin.com/go-snippets/rabbitmq/worker-events/rabbitmq"
)

type Conn struct {
	_parent    *Conn
	httpClient *http.Client
	rabbitMq   *rabbitmq.RabbitMq
}

func New() *Conn {
	return &Conn{}
}
