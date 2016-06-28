package resourcepool_test

import (
	"github.com/rxwen/resourcepool"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResourcePool(t *testing.T) {
	assert := assert.New(t)
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		return "fake_connection", nil
	}, func(interface{}) error {
		return nil
	}, 5)

	assert.Equal(0, pool.Count())

	assert.Nil(err)
	con, err := pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(1, pool.Count())

	err = pool.Release("fake_connection2")
	assert.NotNil(err)
	assert.Equal(1, pool.Count())
	pool.Shrink()
	assert.Equal(1, pool.Count())
	err = pool.Release(con)
	assert.Equal(1, pool.Count())
	pool.Shrink()
	assert.Equal(0, pool.Count())
	assert.Nil(err)
}
