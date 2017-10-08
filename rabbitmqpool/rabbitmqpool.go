package rabbitmqpool

import (
	"fmt"
	"strings"

	"github.com/rxwen/resourcepool"
	"github.com/rxwen/srvresolver"
	"github.com/streadway/amqp"
)

// CreateRabbitmqConnectionPool function creates a connection for specified Rabbitmq service.
func CreateRabbitmqConnectionPool(rabbitmqService string, poolSize int, timeoutSecond int) (*resourcepool.ResourcePool, error) {
	if rabbitmqService[len(rabbitmqService)-1] == '/' {
		rabbitmqService = rabbitmqService[0 : len(rabbitmqService)-1]
	}
	RabbitmqPool, err := resourcepool.NewResourcePool("", "", func(host, port string) (interface{}, error) {
		s := strings.Split(rabbitmqService, "@")
		server, port, err := srvresolver.ResolveSRV(s[1])
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("%s@%s:%s/", s[0], server, port)
		c, err := amqp.Dial(url)
		return c, err
	}, func(c interface{}) error {
		c.(*amqp.Connection).Close()
		return nil
	}, poolSize, timeoutSecond)
	return RabbitmqPool, err
}
