package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/google/uuid"
	"knative.dev/eventing-contrib/pkg/kncloudevents"
	"testing"
)

func TestReceiver(t *testing.T) {
	defaultClient, err := kncloudevents.NewDefaultClient("http://localhost:8080")
	if err != nil {
		panic(err.Error())
	}
	//ref := types.ParseURLRef("https://test")
	//target := cloudevents.Event{
	//	Context: cloudevents.EventContextV03{
	//		SpecVersion: "v3",
	//		Type:        "test1",
	//		Source:      *ref,
	//		Subject:     nil,
	//		ID:          uuid.New().String(),
	//		Extensions: map[string]interface{}{
	//			"test":     "test",
	//			"StreamId": "User-"+uuid.New().String(),
	//		},
	//	}.AsV03(),
	//	Data: string([]byte("test")),
	//}
	target := cloudevents.New(cloudevents.CloudEventsVersionV03)
	target.SetType("dev.knative.docs.sample")
	target.SetSource("https://knative.dev/eventing-contrib/cmd/heartbeats/")
	target.SetID(uuid.New().String())
	target.SetExtension("StreamId", "User-"+uuid.New().String())
	_ = target.SetData("test")

	send, event, err := defaultClient.Send(context.TODO(), target)
	fmt.Println(send, event, err)
	//time.Sleep(time.Hour)
}
