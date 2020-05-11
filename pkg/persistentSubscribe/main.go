package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/flags"
	"knative.dev/eventing-contrib/pkg/kncloudevents"
	"log"
	"os"
	"time"
)

var bus chan *client.RecordedEvent

func main() {
	environ := os.Environ()
	for _, s := range environ {
		fmt.Println(s)
	}
	bus = make(chan *client.RecordedEvent, 1000)
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.Init(set)
	eventHost := os.Getenv("EVENT_HOST")
	if len(eventHost) <= 0 {
		eventHost = "tcp://admin:changeit@localhost:1113"
	}
	set.Set("endpoint", eventHost)
	flag.Parse()

	//gesclient.Debug()

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

	metadataAsync, err := conn.GetStreamMetadataAsync("User-a9315235-af58-4744-ac19-b87787ed8eb2", client.NewUserCredentials("admin", "changeit"))
	fmt.Println(metadataAsync, err, metadataAsync.Result())
	async, err := conn.ConnectToPersistentSubscriptionAsync("allEvent", "test", eventHandler, subHandler,
		client.NewUserCredentials("admin", "changeit"), 1, false)
	if err != nil {
		panic(err.Error())
	}
	if err := async.Error(); err != nil {
		panic(err.Error())
	}
	subscription := async.Result().(client.PersistentSubscription)
	fmt.Printf("开启持久订阅:%+v \n", subscription)

	//go sendEventToNative()
	time.Sleep(time.Hour)
}

func sendEventToNative() {
	redirectUrl := os.Getenv("REDIRECT_URL")
	if len(redirectUrl) <= 0 {
		redirectUrl = "http://event-display:8080"
	}

	defaultClient, err := kncloudevents.NewDefaultClient(redirectUrl)
	//defaultClient.StartReceiver()
	if err != nil {
		panic(err.Error())
	}
	for {
		select {
		case e := <-bus:
			fmt.Println("sendEventToNative", e)
			_, _, err := defaultClient.Send(context.TODO(), newNativeEvent(e))
			if err != nil {
				panic(err.Error())
			}
		}
	}
}

func newNativeEvent(e *client.RecordedEvent) cloudevents.Event {
	eventSource := fmt.Sprintf("https://knative.dev/eventing-contrib/cmd/heartbeats/#%s/%s", "testNS", "testPOD")
	target := cloudevents.New(cloudevents.CloudEventsVersionV03)
	target.SetType(e.EventType())
	target.SetSource(eventSource)
	target.SetID(e.EventId().String())
	target.SetExtension("StreamId", e.EventStreamId())
	_ = target.SetData(e.Data())
	//ref := types.ParseURLRef(eventSource)
	//event := cloudevents.Event{
	//	Context: cloudevents.EventContextV03{
	//		SpecVersion: "v3",
	//		Type:        e.EventType(),
	//		Source:      *ref,
	//		Subject:     nil,
	//		ID:          e.EventId().String(),
	//		Extensions: map[string]interface{}{
	//			"test":     "test",
	//			"StreamId": e.EventStreamId(),
	//		},
	//	}.AsV03(),
	//	Data: string(e.Data()),
	//}
	return target
	//return event
}

func eventHandler(s client.PersistentSubscription, r *client.ResolvedEvent) error {
	fmt.Printf("=================================\n")
	fmt.Printf("收到事件:%+v \n", r)

	fmt.Printf("收到事件内容:%+v \n", r.Event().String())
	fmt.Printf("收到事件内容.data:%s \n", r.Event().Data())

	err := s.Acknowledge([]client.ResolvedEvent{*r})
	fmt.Printf("=============err:%+v====================\n", err)
	if err != nil {
		fmt.Printf("err:%v \n", err.Error())
		return err
	}
	bus <- r.Event()
	return nil
}

func subHandler(s client.PersistentSubscription, dr client.SubscriptionDropReason, err error) error {
	fmt.Printf("---------------------------------\n")
	fmt.Printf("subhandler处理:%+v,err:%+v \n", dr.String(), err)
	fmt.Printf("---------------------------------\n")
	s.Stop()
	return nil
}
