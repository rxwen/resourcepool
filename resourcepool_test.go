package resourcepool_test

import (
	"log"
	"testing"

	"github.com/rxwen/resourcepool"
	"github.com/stretchr/testify/assert"
)

func TestResourcePool(t *testing.T) {
	assert := assert.New(t)
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		log.Println("create new resource")
		return "fake_connection", nil
	}, func(interface{}) error {
		return nil
	}, 3, 1)

	assert.Equal(0, pool.Count())

	assert.Nil(err)
	con, err := pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(1, pool.Count())
	err = pool.Release(con)
	assert.NotNil(con)
	assert.Equal(1, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(1, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(2, pool.Count())
	err = pool.Release(con)
	assert.NotNil(con)
	assert.Equal(2, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(2, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(3, pool.Count())
	con, err = pool.Get()
	assert.NotNil(err) // should timeout
	assert.Equal(3, pool.Count())

	err = pool.Release("not managed resource")
	assert.NotNil(err) // should report error
	assert.Equal(3, pool.Count())

	err = pool.Release(con)
	assert.NotNil(err)
	assert.Equal(3, pool.Count())
	err = pool.Release(con)
	assert.Equal(3, pool.Count())
}
