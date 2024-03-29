package resourcepool_test

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/rxwen/resourcepool"
	"github.com/stretchr/testify/assert"
)

type FakeResource struct {
	Name     string
	IsClosed bool
}

func TestResourcePool(t *testing.T) {
	assert := assert.New(t)
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		log.Println("create new resource")
		return &FakeResource{
			IsClosed: false,
		}, nil
	}, func(r interface{}) error {
		if r != nil {
			re := r.(*FakeResource)
			re.IsClosed = true
		}
		return nil
	}, 3, 3)

	assert.Equal(0, pool.Count())

	assert.Nil(err)
	con, err := pool.Get(500)
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(0, pool.Count())
	assert.False(con.(*FakeResource).IsClosed)
	err = pool.Release(con)
	assert.NotNil(con)
	assert.False(con.(*FakeResource).IsClosed)
	assert.Equal(1, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.False(con.(*FakeResource).IsClosed)
	assert.Equal(0, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.False(con.(*FakeResource).IsClosed)
	assert.Equal(0, pool.Count())
	err = pool.Release(con)
	assert.NotNil(con)
	assert.False(con.(*FakeResource).IsClosed)
	assert.Equal(1, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(0, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.NotNil(con)
	assert.Equal(0, pool.Count())
	con, err = pool.Get()
	assert.Nil(err)
	assert.Equal(0, pool.Count())
	err = pool.Release(con)
	assert.Nil(err)
	assert.Equal(1, pool.Count())
	err = pool.Release(con)
	assert.Equal(2, pool.Count())
	con1, err := pool.Get()
	con2, err := pool.Get()
	con3, err := pool.Get()
	con4, err := pool.Get()
	con5, err := pool.Get()
	assert.Equal(0, pool.Count())
	err = pool.Release(con1)
	assert.Equal(1, pool.Count())
	err = pool.Release(con2)
	assert.Equal(2, pool.Count())
	assert.False(con3.(*FakeResource).IsClosed)
	err = pool.Release(con3)
	assert.False(con3.(*FakeResource).IsClosed)
	assert.Equal(3, pool.Count())
	assert.False(con4.(*FakeResource).IsClosed)
	err = pool.Release(con4)
	assert.True(con4.(*FakeResource).IsClosed)
	assert.Equal(3, pool.Count())
	assert.False(con5.(*FakeResource).IsClosed)
	err = pool.Release(con5)
	assert.True(con5.(*FakeResource).IsClosed)
	assert.Equal(3, pool.Count())
}

func TestResourcePoolCheckError(t *testing.T) {
	log.Info("start TestResourcePoolCheckError")
	assert := assert.New(t)
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		log.Println("create new resource")
		return &FakeResource{}, nil
	}, func(interface{}) error {
		log.Println("close resource")
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
	assert.Nil(pool.CheckError(con, fe))
	con = nil
	assert.Nil(pool.Release(con))
	count2 := pool.Count()
	assert.Equal(count1, count2)
	log.Info("end TestResourcePoolCheckError")
}

func TestResourcePoolReleaseOrder(t *testing.T) {
	log.Info("start TestResourcePoolReleaseOrder")
	assert := assert.New(t)
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		log.Println("create new resource")
		return &FakeResource{}, nil
	}, func(interface{}) error {
		log.Println("close resource")
		return nil
	}, 3, 1)

	con1, err := pool.Get()
	con2, err := pool.Get()
	con3, err := pool.Get()
	assert.Nil(err)
	assert.NotNil(con1)
	assert.NotNil(con2)
	assert.NotNil(con3)
	count1 := pool.Count()
	assert.Equal(count1, 0)
	pool.Release(con2)
	count1 = pool.Count()
	assert.Equal(count1, 1)
	err = pool.Release(con2)
	count1 = pool.Count()
	assert.Equal(count1, 2)
	con4, err := pool.Get()
	assert.Nil(err)
	assert.NotNil(con4)
	err = pool.Release(con4)
	assert.Nil(err)
	err = pool.Release(con1)
	assert.Nil(err)
	err = pool.Release(con3)
	assert.Nil(err)
	err = pool.Release(con3)
	assert.Nil(err)

	idleList := make(chan interface{}, 2)
	select {
	case idleList <- con1:
	default:
		log.Println("failed to add to chan")
	}
	select {
	case idleList <- con1:
	default:
		log.Println("failed to add to chan")
	}
	select {
	case idleList <- con2:
	default:
		log.Println("failed to add to chan")
	}
	select {
	case idleList <- con3:
	default:
		log.Println("failed to add to chan")
	}
	select {
	case idleList <- con4:
	default:
		log.Println("failed to add to chan")
	}
	log.Info("end TestResourcePoolReleaseOrder")
}

func putbackFunc(pool *resourcepool.ResourcePool, destroy bool) {
	r1, _ := pool.Get()
	r2, _ := pool.Get()
	con, _ := pool.Get()
	defer func() { pool.Putback(con, destroy) }()
	defer func() { pool.Putback(r1, destroy) }()
	defer func() { pool.Putback(r2, destroy) }()
}

func TestResourcePoolPutback(t *testing.T) {
	log.Info("start TestResourcePoolPutback")
	assert := assert.New(t)
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		log.Println("create new resource")
		return &FakeResource{}, nil
	}, func(r interface{}) error {
		log.Println("close resource ")
		log.Println(r)
		return nil
	}, 3, 1)
	assert.Nil(err)
	assert.NotNil(pool)
	putbackFunc(pool, false)
	assert.Equal(3, pool.Count())
	putbackFunc(pool, true)
	assert.Equal(0, pool.Count())
	putbackFunc(pool, false)
	assert.Equal(3, pool.Count())
	putbackFunc(pool, true)
	assert.Equal(0, pool.Count())
	log.Info("end TestResourcePoolPutback")
}
