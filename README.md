[![wercker status](https://app.wercker.com/status/0023c298ffb864891b0d3bab0951bba9/m/master "wercker status")](https://app.wercker.com/project/byKey/0023c298ffb864891b0d3bab0951bba9)


resourcepool is a pool for resource management. It can be used to manage resources like database connections, redis connections, thrift connections. It has following features:

- support multiple backend servers
- support backend server replacement on the fly

Example
```
	pool, err := resourcepool.NewResourcePool("fakehost", "9090", func(host, port string) (interface{}, error) {
		log.Println("create new resource")
		return &FakeResource{}, nil
	}, func(r interface{}) error {
		log.Println("close resource ")
		log.Println(r)
		return nil
	}, 3, 1)

    destroy := false
	r1, _ := pool.Get()
	r2, _ := pool.Get()
	con, _ := pool.Get()
	defer func() { pool.Putback(r1, destroy) }() // use a func to wrap the putback to avoid evalute destroy value now
	defer func() { pool.Putback(r2, destroy) }()
	defer func() { pool.Putback(con, destroy) }()
    // use resource

    // if any error found during use the resource, and the resource is deemed not reuseable anymore, set destroy flag
    // to ture, indicate the resource shall not be managed by resource pool anymore
    // an example is, our code get a network connection from the pool, but found the connection is broken and it should
    // not be putback to the pool
    destroy = true

```

