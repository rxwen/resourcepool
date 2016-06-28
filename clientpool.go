package resourcepool

import (
	"container/list"
	"errors"
	"net"
	"sync"
)

const DefaultPoolSize = 32

type ResourcePool struct {
	lock         sync.Mutex
	host         string
	port         string
	creationFunc ClientCreationFunc
	closeFunc    ClientCloseFunc
	idleList     list.List
	busyList     list.List
}

// ClientCreationFunc is the function used for creating new client.
type ClientCreationFunc func(host, port string) (interface{}, error)

// ClientCloseFunc is the function used for closing client.
type ClientCloseFunc func(interface{}) error

// AddServer adds a new server to the pool.
func NewResourcePool(host, port string, fnCreation ClientCreationFunc, fnClose ClientCloseFunc, maxSize int) (*ResourcePool, error) {
	pool := ResourcePool{
		//lock : sync.Mutex
		host:         host,
		port:         port,
		creationFunc: fnCreation,
		closeFunc:    fnClose,
	}
	return &pool, nil
}

// Get retrives a connection from the pool.
func (pool *ResourcePool) Get() (interface{}, error) {
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
func (pool *ResourcePool) Release(c interface{}) error {
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

// Shrink disconnects all idle connectsions.
func (pool *ResourcePool) Shrink() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	for client := pool.idleList.Front(); client != nil; client = client.Next() {
		pool.closeFunc(client.Value)
		pool.idleList.Remove(client)
	}
}

// Destroy disconnects all connectsions.
func (pool *ResourcePool) Destroy() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.Shrink()

	for client := pool.busyList.Front(); client != nil; client = client.Next() {
		pool.closeFunc(client.Value)
		pool.idleList.Remove(client)
	}
}

// Replace replaces existing connections to oldServer with connections to newServer.
func (pool *ResourcePool) Replace(oldHost, oldPort, newHost, newPort string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
}

// Count returns total number of connections in the pool.
func (pool *ResourcePool) Count() int {
	return pool.idleList.Len() + pool.busyList.Len()
}
