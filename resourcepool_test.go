package resourcepool_test

import (
	"errors"
	"log"
	"testing"

	"github.com/rxwen/resourcepool"
	"github.com/stretchr/testify/assert"
)

type FakeResource struct {
	Name string
}

func TestResourcePool(t *testing.T) {
	assert := assert.New(t)
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		log.Println("create new resource")
		return &FakeResource{}, nil
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

func TestResourcePoolCheckError(t *testing.T) {
	assert := assert.New(t)
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		log.Println("create new resource")
		return &FakeResource{}, nil
	}, func(interface{}) error {
		return nil
	}, 3, 1)

	_, err = pool.Get()
	_, err = pool.Get()
	con, err := pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	defer pool.Release(con)
	fe := errors.New("fake error")
	count1 := pool.Count()
	log.Println(pool.Count())
	assert.Nil(pool.CheckError(con, fe))
	count2 := pool.Count()
	assert.NotNil(pool.CheckError(con, fe))
	count3 := pool.Count()
	assert.NotEqual(count1, count2)
	assert.Equal(count3, count2)
}
