package mqttpool

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/rxwen/resourcepool"
)

// CreateRabbitmqConnectionPool function creates a connection for specified Rabbitmq service.
func CreateMqttConnectionPool(mqttService string, poolSize int, timeoutSecond int) (*resourcepool.ResourcePool, error) {
	MqttPool, err := resourcepool.NewResourcePool("", "", func(host, port string) (interface{}, error) {
		connOpts := mqtt.NewClientOptions()
		connOpts.ProtocolVersion = 4
		connOpts.AddBroker(mqttService)
		c := mqtt.NewClient(connOpts)
		if token := c.Connect(); token.Wait() && token.Error() != nil {
			return nil, token.Error()
		}
		return c, nil
	}, func(c interface{}) error {
		c.(mqtt.Client).Disconnect(0)
		return nil
	}, poolSize, timeoutSecond)
	return MqttPool, err
}
