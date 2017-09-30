package rabbitmqpool

import (
	"fmt"

	"github.com/rxwen/resourcepool"
	"github.com/rxwen/srvresolver"
	"github.com/streadway/amqp"
)

// CreateRabbitmqConnectionPool function creates a connection for specified Rabbitmq service.
func CreateRabbitmqConnectionPool(username, password, RabbitmqService string, poolSize int, timeoutSecond int) (*resourcepool.ResourcePool, error) {
	RabbitmqPool, err := resourcepool.NewResourcePool("", "", func(host, port string) (interface{}, error) {
		server, port, err := srvresolver.ResolveSRV(RabbitmqService)
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("amqp://%s:%s@%s:%s/", "guest", "guest", server, port)
		c, err := amqp.Dial(url)
		return c, err
	}, func(c interface{}) error {
		c.(*amqp.Connection).Close()
		return nil
	}, poolSize, timeoutSecond)
	return RabbitmqPool, err
}
