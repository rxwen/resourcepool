package redispool

import (
	"net"

	"github.com/garyburd/redigo/redis"
	"github.com/rxwen/resourcepool"
	"github.com/rxwen/srvresolver"
)

// CreateRedisConnectionPool function creates a connection for specified redis service.
func CreateRedisConnectionPool(redisService string, poolSize int, timeoutSecond int) (*resourcepool.ResourcePool, error) {
	redisPool, err := resourcepool.NewResourcePool("", "", func(host, port string) (interface{}, error) {
		redisServer, redisPort, err := srvresolver.ResolveSRV(redisService)
		if err != nil {
			return nil, err
		}
		c, err := redis.Dial("tcp", net.JoinHostPort(redisServer, redisPort))
		return c, err
	}, func(c interface{}) error {
		c.(redis.Conn).Close()
		return nil
	}, poolSize, timeoutSecond)
	return redisPool, err
}
