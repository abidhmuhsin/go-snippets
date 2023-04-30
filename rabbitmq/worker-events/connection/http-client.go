package connection

import (
	"net/http"
	"time"
)

func (c *Conn) HttpClient() *http.Client {
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return c.httpClient
}
