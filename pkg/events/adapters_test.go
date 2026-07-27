package events

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

func TestNatsEventBus(t *testing.T) {
	t.Parallel()

	opts := natsserver.Options{Port: -1}
	ns, err := natsserver.NewServer(&opts)
	if err != nil {
		t.Fatalf("Failed to start nats server: %v", err)
	}

	go ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		t.Fatalf("nats server failed to start")
	}

	defer ns.Shutdown()

	clientURL, err := url.Parse(ns.ClientURL())
	if err != nil {
		t.Fatalf("Failed to parse nats server url: %v", err)
	}
	eventBus, err := NewNatsEventBus(clientURL, 0, 1, db.NewInMemoryDB())
	if err != nil {
		t.Fatalf("Failed to create nats event bus: %v", err)
	}

	event := models.Event{
		ID:      uuid.New(),
		Version: 1,
		Topic:   "test",
		JobName: "test",
		Type:    "dft",
		Payload: &models.DFTEventPayload{
			InputFile: "testfile.txt",
		},
	}

	dataChan := make(chan string)
	err = eventBus.Subscribe("test", func(e *models.Event) error {
		dataChan <- e.Payload.(*models.DFTEventPayload).InputFile
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to event: %v", err)
	}

	err = eventBus.Publish(&event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	result := <-dataChan

	if result != "testfile.txt" {
		t.Errorf("Expected result to be 'testfile.txt', got %s", result)
	}
}
