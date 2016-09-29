// Package main provides ...
package main

import (
	"bufio"
	"log"
	"net"
	"os"

	"github.com/rxwen/resourcepool"
)

func main() {
	pool, _ := resourcepool.NewResourcePool("127.0.0.1", "29090", func(host, port string) (interface{}, error) {
		endpoint := host + ":" + port
		con, err := net.Dial("tcp", endpoint)
		log.Println("create connection to " + endpoint)
		if err != nil {
			log.Println("dial " + endpoint + " failed, err is: " + err.Error())
		}
		return con, err
	}, func(res interface{}) error {
		log.Println("destroy connection")
		con := res.(net.Conn)
		con.Close()
		return nil
	}, 3, 1)

	for i := 0; i < 10000; i++ {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		res, err := pool.Get()
		log.Println("pool size is ", pool.Count())
		if err != nil {
			log.Println("failed to get resource, error is: " + err.Error())
			continue
		}
		con := res.(net.Conn)
		log.Println("about to write message " + text)
		n, err := con.Write([]byte("hello world " + text + " \n"))
		log.Println("write message result ", n, " ", err)
		pool.CheckError(con, err)
		err = pool.Release(res)
		log.Println("release result: ", err)
	}
}
