package thriftpool

import (
	"errors"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/rxwen/resourcepool"
	"github.com/rxwen/srvresolver"
)

// CreateThriftPool function creates a connection for specified thrift servie.
func CreateThriftPool(protocolType, transportType, endpoint string,
	clientFunc func(thrift.TTransport, thrift.TProtocolFactory) interface{},
	poolSize int, timeoutSecond int) (*resourcepool.ResourcePool, error) {
	pool, err := resourcepool.NewResourcePool("", "", func(host, port string) (interface{}, error) {
		host, port, err := srvresolver.ResolveSRV(endpoint)
		if err != nil {
			return nil, err
		}

		protocolFactory, transport, err := CreateThriftClient(protocolType, transportType, host, port)

		if err != nil {
			return nil, err
		}
		err = transport.Open()
		client := clientFunc(transport, protocolFactory)
		return client, err
	}, func(interface{}) error {
		return nil
	}, poolSize, timeoutSecond)
	return pool, err
}

// CreateThriftServer creates a thrift client.
func CreateThriftClient(protocolType, transportType, host, port string) (
	protocolFactory thrift.TProtocolFactory, transport thrift.TTransport, err error) {

	switch protocolType {
	case "binary":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	default:
		err = errors.New("unsupported protocol " + protocolType)
	}
	if err != nil {
		return nil, nil, err
	}
	switch transportType {
	case "socket":
		endpoint := host + ":" + port
		transport, err = thrift.NewTSocket(endpoint)
	default:
		err = errors.New("unsupported transport " + transportType)
	}
	return
}

// CreateThriftServer creates a thrift server.
func CreateThriftServer(protocolType, transportType, endpoint string,
	processor thrift.TProcessor) (
	server *thrift.TSimpleServer, err error) {
	host, port, err := srvresolver.ResolveSRV(endpoint)
	if err != nil {
		return nil, err
	}
	transportFactory := thrift.NewTTransportFactory()

	var protocolFactory thrift.TProtocolFactory
	switch protocolType {
	case "binary":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	default:
		err = errors.New("unsupported protocol " + protocolType)
	}
	if err != nil {
		return nil, err
	}

	var transport thrift.TServerTransport
	switch transportType {
	case "socket":
		endpoint := host + ":" + port
		transport, err = thrift.NewTServerSocket(endpoint)
	default:
		err = errors.New("unsupported transport " + transportType)
	}
	if err != nil {
		return nil, err
	}

	server = thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	return server, err
}
