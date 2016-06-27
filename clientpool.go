package thriftclientpool

import (
	"container/list"
	"errors"
	"sync"
)

type ThriftClientPool struct {
	lock         sync.Mutex
	host         string
	port         string
	creationFunc ClientCreationFunc
	idleList     list.List
	busyList     list.List
}

// ClientCreationFunc is the function used for creating new client.
type ClientCreationFunc func(host, port string) (interface{}, error)

// AddServer adds a new server to the pool.
func NewThriftClientPool(host, port string, fn ClientCreationFunc) (*ThriftClientPool, error) {
	pool := ThriftClientPool{
		//lock : sync.Mutex
		host:         host,
		port:         port,
		creationFunc: fn,
	}
	return &pool, nil
}

// Get retrives a connection from the pool.
func (pool *ThriftClientPool) Get() (interface{}, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if pool.idleList.Len() > 0 {
		client := pool.idleList.Front()
		pool.idleList.Remove(client)
		pool.busyList.PushBack(client.Value)
		return client.Value, nil
	} else {
		client, err := pool.creationFunc(pool.host, pool.port)

		if err != nil {
			return nil, err
		}
		pool.busyList.PushBack(client)
		return client, nil
	}
}

// Release puts the connection back to the pool.
func (pool *ThriftClientPool) Release(c interface{}) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	element := pool.busyList.Front()
	for {
		if element == nil {
			return errors.New("the client isn't found in the pool, is it a managed client?")
		}
		if c == element.Value {
			pool.busyList.Remove(element)
			pool.idleList.PushBack(element.Value)
			return nil
		}
		element = element.Next()
	}
}

// Destroy disconnects all connectsions.
func (pool *ThriftClientPool) Destroy() {
	pool.lock.Lock()
	defer pool.lock.Unlock()
}

// Replace replaces existing connections to oldServer with connections to newServer.
func (pool *ThriftClientPool) Replace(oldHost, oldPort, newHost, newPort string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
}
