package resourcepool

import (
	"container/list"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const DefaultPoolSize = 32
const DefaultGetTimeoutSecond = 3

type ResourcePool struct {
	lock         sync.Mutex
	host         string
	port         string
	creationFunc ClientCreationFunc
	closeFunc    ClientCloseFunc
	maxSize      int
	getTimeout   int
	busyList     list.List
	idleList     chan interface{}
}

// ClientCreationFunc is the function used for creating new client.
type ClientCreationFunc func(host, port string) (interface{}, error)

// ClientCloseFunc is the function used for closing client.
type ClientCloseFunc func(interface{}) error

// AddServer adds a new server to the pool.
func NewResourcePool(host, port string, fnCreation ClientCreationFunc, fnClose ClientCloseFunc, maxSize, getTimeout int) (*ResourcePool, error) {
	pool := ResourcePool{
		maxSize:      maxSize,
		host:         host,
		port:         port,
		creationFunc: fnCreation,
		closeFunc:    fnClose,
		getTimeout:   getTimeout,
		idleList:     make(chan interface{}, maxSize),
	}
	return &pool, nil
}

// Get retrives a connection from the pool.
func (pool *ResourcePool) Get() (interface{}, error) {
	var res interface{}
	var err error
	if pool.maxSize == 0 {
		fmt.Println("max size is 0")
		return pool.creationFunc(pool.host, pool.port)
	}
	fmt.Println("total size: ", pool.Count(), "busy list len", pool.busyList.Len())
	select {
	// try get without block to see if resource is already available
	case res = <-pool.idleList:
		fmt.Println("there is idle resource available now")
	default:
		fmt.Println("there is no idle resource available now")
		res = nil
	}

	if res != nil {
		pool.lock.Lock()
		defer pool.lock.Unlock()
		pool.busyList.PushBack(res)
		return res, nil
	} else if pool.Count() < pool.maxSize {
		go func() {
			res, err = pool.creationFunc(pool.host, pool.port)
			if err != nil {
				log.Println("resource creation failed: ", err)
			} else {
				pool.idleList <- res
			}
		}()
	}

	if pool.getTimeout != 0 {
		select {
		case res = <-pool.idleList:
		case <-time.After(time.Second * time.Duration(pool.getTimeout)):
			res = nil
			err = errors.New("get resource timed out")
		}
	} else {
		select {
		case res = <-pool.idleList:
		}
	}
	if err == nil {
		pool.lock.Lock()
		defer pool.lock.Unlock()
		pool.busyList.PushBack(res)
	}
	return res, err
}

// Release puts the connection back to the pool.
func (pool *ResourcePool) Release(c interface{}) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	element := pool.busyList.Front()
	for {
		if element == nil {
			return errors.New("the resource isn't found in the pool, is it a managed resource?")
		}
		if c == element.Value {
			select {
			case pool.idleList <- c:
				pool.busyList.Remove(element)
				return nil
			default:
				pool.closeFunc(c)
				return errors.New("the resource can't be put to idle list")
			}
		}
		element = element.Next()
	}
}

// CheckError destroies the connection when necessary by checking error.
func (pool *ResourcePool) CheckError(c interface{}, err error) error {
	if err == nil {
		return nil
	}
	pool.lock.Lock()
	defer pool.lock.Unlock()
	element := pool.busyList.Front()
	for {
		if element == nil {
			return errors.New("the resource isn't found in the pool, is it a managed resource?")
		}
		if c == element.Value {
			log.Println("encountered an error, destory the connection now")
			pool.busyList.Remove(element)
			pool.closeFunc(c)
			return nil
		}
		element = element.Next()
	}
	return nil
}

// Destroy disconnects all connectsions.
func (pool *ResourcePool) Destroy() error {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if pool.busyList.Len() < 1 {
		return errors.New("not all managed resources are free, can't destory now")
	}
	close(pool.idleList)
	for res := range pool.idleList {
		pool.closeFunc(res)
	}
	return nil
}

// Replace replaces existing connections to oldServer with connections to newServer.
func (pool *ResourcePool) Replace(oldHost, oldPort, newHost, newPort string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
}

// Count returns total number of connections in the pool.
func (pool *ResourcePool) Count() int {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return len(pool.idleList) + pool.busyList.Len()
}
