package thriftclientpool_test

import (
	"github.com/rxwen/thrift-client-pool"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClientPool(t *testing.T) {
	assert := assert.New(t)
	pool, err := thriftclientpool.NewThriftClientPool("fakehost", "9090", func(host, port string) (interface{}, error) {
		return "fake_connection", nil
	})

	assert.Nil(err)
	con, err := pool.Get()
	assert.Nil(err)
	assert.NotNil(con)

	err = pool.Release("fake_connection2")
	assert.NotNil(err)
	err = pool.Release(con)
	assert.Nil(err)
}
