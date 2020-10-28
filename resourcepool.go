package resourcepool

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type ResourcePool struct {
	host         string
	port         string
	creationFunc ClientCreationFunc
	closeFunc    ClientCloseFunc
	maxSize      int
	getTimeout   int
	idleList     chan interface{}
}

// ClientCreationFunc is the function used for creating new client.
type ClientCreationFunc func(host, port string) (interface{}, error)

// ClientCloseFunc is the function used for closing client.
type ClientCloseFunc func(interface{}) error

// AddServer adds a new server to the pool.
func NewResourcePool(host, port string, fnCreation ClientCreationFunc,
	fnClose ClientCloseFunc, maxSize, getTimeout int) (*ResourcePool, error) {
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
func (pool *ResourcePool) Get(waittime ...int) (interface{}, error) {
	var res interface{}
	var err error
	timetowait := 3000
	if len(waittime) > 0 {
		timetowait = waittime[0]
	}
	select {
	// try get without block to see if resource is already available
	case res = <-pool.idleList:
	case <-time.After(time.Duration(timetowait) * time.Millisecond):
		log.WithField("wait time", timetowait).Info("wait timed out, create new")
	}
	if res == nil {
		res, err = pool.creationFunc(pool.host, pool.port)
		if err != nil {
			log.WithError(err).Error("resource creation failed")
		}
	}
	return res, err
}

// Release puts the connection back to the pool.
// deprecated, use Putback instead
func (pool *ResourcePool) Release(c interface{}) error {
	log.Warn("deprecated Release, use Putback instead")
	if c == nil {
		log.Info("release nil resource, ignore")
		return nil
	}
	if len(pool.idleList) >= pool.maxSize {
		pool.closeFunc(c)
		return nil
	}
	select {
	case pool.idleList <- c:
	}
	return nil

}

func (pool *ResourcePool) Putback(c interface{}, destroy bool) error {
	if c == nil {
		log.Info("release nil resource, ignore")
		return nil
	}
	if destroy || len(pool.idleList) >= pool.maxSize {
		pool.closeFunc(c)
		return nil
	}
	select {
	case pool.idleList <- c:
	}
	return nil

}

// CheckError destroies the connection when necessary by checking error.
// deprecated, not useful anymore
func (pool *ResourcePool) CheckError(c interface{}, err error) error {
	log.Warn("deprecated CheckError, use Putback instead")
	if err == nil {
		return nil
	}
	log.Info("encountered an error, destory the connection")
	pool.closeFunc(c)
	return nil
}

// Destroy disconnects all connectsions.
func (pool *ResourcePool) Destroy() error {
	close(pool.idleList)
	for res := range pool.idleList {
		pool.closeFunc(res)
	}
	return nil
}

// Replace replaces existing connections to oldServer with connections to newServer.
func (pool *ResourcePool) Replace(oldHost, oldPort, newHost, newPort string) {
}

// Count returns total number of connections in the pool.
func (pool *ResourcePool) Count() int {
	return len(pool.idleList)
}

func (pool *ResourcePool) Dump(reason string) {
	log.WithFields(log.Fields{
		"port":   pool.port,
		"idle":   len(pool.idleList),
		"reason": reason,
	}).Info("pool status")
}
