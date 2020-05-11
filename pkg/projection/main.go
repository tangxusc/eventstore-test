package main

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/projections"
	"net"
	"time"
)

func main() {
	addr, err := net.ResolveTCPAddr(`tcp`, `localhost:2113`)
	if err != nil {
		panic(err.Error())
	}
	manager := projections.NewManager(addr, time.Second*10)
	async := manager.GetPartitionStateAsync("User", "User-1",
		client.NewUserCredentials("admin", "changeit"))
	if err := async.Error(); err != nil {
		panic(err.Error())
	} else {
		fmt.Println(async.Result())
	}
}
