package thriftpool

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/rxwen/resourcepool"
	"github.com/rxwen/srvresolver"
)

// CreateThriftPool function creates a connection for specified thrift servie.
func CreateThriftPool(endpoint string,
	clientFunc func(thrift.TTransport, thrift.TProtocolFactory) interface{},
	poolSize int, timeoutSecond int) (*resourcepool.ResourcePool, error) {
	pool, err := resourcepool.NewResourcePool("", "", func(host, port string) (interface{}, error) {
		host, port, err := srvresolver.ResolveSRV(endpoint)
		if err != nil {
			return nil, err
		}
		verificationServiceEndpoint := host + ":" + port
		protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
		useTransport, err := thrift.NewTSocket(verificationServiceEndpoint)
		if err != nil {
			return nil, err
		}
		err = useTransport.Open()
		client := clientFunc(useTransport, protocolFactory)
		return client, err
	}, func(interface{}) error {
		return nil
	}, poolSize, timeoutSecond)
	return pool, err
}
