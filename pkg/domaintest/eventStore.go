package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/flags"
	"github.com/jdextraze/go-gesclient/projections"
	uuid "github.com/satori/go.uuid"
	"net"
	"os"
	"time"
)

var UserCredentials = client.NewUserCredentials("admin", "changeit")

type Aggregate interface {
	GetAggregateName() string
	GetVersion() int
	GetId() string
}

type Event interface {
	GetType() string
}

func SendEvent(event Event, agg Aggregate) error {
	v4, err := uuid.NewV4()
	if err != nil {
		return err
	}
	marshal, err := json.Marshal(event)
	if err != nil {
		return err
	}
	data := client.NewEventData(v4, event.GetType(), true, marshal, nil)

	set := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.Init(set)
	err = set.Parse(os.Args)
	if err != nil {
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	conn, err := flags.CreateConnection(hostname)
	if err != nil {
		return err
	}
	if err := conn.ConnectAsync().Wait(); err != nil {
		return err
	}
	defer conn.Close()

	task, err := conn.AppendToStreamAsync(getStreamName(agg), agg.GetVersion(), []*client.EventData{data},
		UserCredentials)
	if err != nil {
		return err
	} else if err := task.Error(); err != nil {
		return err
	} else {
		result := task.Result().(*client.WriteResult)
		println(result.String())
	}
	return nil
}

func getStreamName(agg Aggregate) string {
	return fmt.Sprintf("%s@%s", agg.GetAggregateName(), agg.GetId())
}

func LoadProjection(agg Aggregate) (interface{}, error) {
	addr, err := net.ResolveTCPAddr(`tcp`, `localhost:2113`)
	if err != nil {
		return nil, err
	}
	manager := projections.NewManager(addr, time.Second*10)
	async := manager.GetPartitionStateAsync(agg.GetAggregateName(), getStreamName(agg), UserCredentials)
	if err := async.Error(); err != nil {
		return nil, err
	} else {
		return async.Result(), nil
	}
}
