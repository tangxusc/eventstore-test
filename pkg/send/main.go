package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/flags"
	uuid "github.com/satori/go.uuid"
	"log"
	"os"
)

func main() {
	flags.Init(flag.CommandLine)

	flag.Parse()

	gesclient.Debug()
	hostname, err := os.Hostname()
	if err != nil {
		panic(err.Error())
	}
	conn, err := flags.CreateConnection(hostname)
	if err != nil {
		panic(err.Error())
	}
	conn.Connected().Add(func(evt client.Event) error { log.Printf("Connected: %+v", evt); return nil })
	conn.Disconnected().Add(func(evt client.Event) error { log.Printf("Disconnected: %+v", evt); return nil })
	conn.Reconnecting().Add(func(evt client.Event) error { log.Printf("Reconnecting: %+v", evt); return nil })
	conn.Closed().Add(func(evt client.Event) error { log.Fatalf("Connection closed: %+v", evt); return nil })
	conn.ErrorOccurred().Add(func(evt client.Event) error { log.Printf("Error: %+v", evt); return nil })
	conn.AuthenticationFailed().Add(func(evt client.Event) error { log.Printf("Auth failed: %+v", evt); return nil })

	if err := conn.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	defer conn.Disconnected()

	u := &UserCreated{
		Id:       "2",
		UserName: "test2",
		Password: "222222",
	}
	marshal, err := json.Marshal(u)
	if err != nil {
		panic(err.Error())
	}
	v4, err := uuid.NewV4()
	if err != nil {
		panic(err.Error())
	}
	data := client.NewEventData(v4, "UserCreated", true, marshal, nil)
	s := fmt.Sprintf("User-%v", u.Id)
	task, err := conn.AppendToStreamAsync(s, client.ExpectedVersion_Any, []*client.EventData{data},
		client.NewUserCredentials("admin", "changeit"))
	if err != nil {
		panic(err.Error())
	} else if err := task.Error(); err != nil {
		panic(err.Error())
	} else {
		result := task.Result().(*client.WriteResult)
		println(result.String())
	}

}

type UserCreated struct {
	Id       string `json:"id"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
}
