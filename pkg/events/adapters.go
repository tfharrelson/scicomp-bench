package events

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

type NatsEventBus struct {
	nc *nats.Conn
	db db.DB
}

func (e *NatsEventBus) Publish(event *models.Event) error {
	// figure out the data part of the event publishing
	data, err := marshalEventToNats(event)
	if err != nil {
		return err
	}
	err = e.db.PublishEvent(event)
	if err != nil {
		return err
	}
	return e.nc.Publish(string(event.Topic), data)
}

func (e *NatsEventBus) Subscribe(topic models.Topic, handler func(*models.Event) error) error {
	// TODO: figure out what to do with the returned subscription
	_, err := e.nc.Subscribe(string(topic), func(msg *nats.Msg) {
		if msg.Reply != "" {
			err := msg.InProgress()
			if err != nil {
				panic(err)
			}
		}

		event, err := unmarshalMsg(msg)
		if err != nil {
			panic(err)
		}

		err = handler(event)
		if err != nil {
			fmt.Println(err)
			if msg.Reply != "" {
				if err := msg.Nak(); err != nil {
					panic(err)
				}
			}
		}

		if msg.Reply != "" {
			if err := msg.Ack(); err != nil {
				panic(err)
			}
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func NewNatsEventBus(
	URL *url.URL,
	MaxReconnects int,
	WaitTime int,
	DB db.DB,
) (*NatsEventBus, error) {
	waitTime := time.Duration(WaitTime)
	nc, err := nats.Connect(
		URL.String(),
		nats.MaxReconnects(MaxReconnects),
		nats.ReconnectWait(waitTime),
		nats.RetryOnFailedConnect(true),
		nats.DisconnectHandler(func(_ *nats.Conn) {
			fmt.Println("Disconnected")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			fmt.Println("Reconnected")
		}),
	)
	if err != nil {
		return nil, err
	}
	return &NatsEventBus{
		nc: nc,
		db: DB,
	}, nil
}

func unmarshalMsg(msg *nats.Msg) (*models.Event, error) {
	event := new(models.Event)
	event.Topic = models.Topic(msg.Subject)

	rawData := make(map[string]interface{})
	if err := json.Unmarshal(msg.Data, &rawData); err != nil {
		return nil, err
	}

	parsedUUID, err := uuid.Parse(rawData["id"].(string))
	if err != nil {
		return nil, err
	}
	event.ID = parsedUUID
	version, ok := rawData["version"].(float64)
	if !ok {
		fmt.Println(rawData["version"])
		return nil, fmt.Errorf("invalid version")
	}
	event.Version = int(version)
	event.JobName = rawData["job_name"].(string)
	switch rawData["type"] {
	case "dft":
		event.Type = models.DFTJob
	case "dummy":
		event.Type = models.DummyJob
	default:
		return nil, fmt.Errorf("unknown job type")
	}

	rawPayload, ok := rawData["payload"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload")
	}
	switch rawData["type"] {
	case "dft":
		payload, err := constructDFTPayload(rawPayload)
		if err != nil {
			return nil, err
		}
		event.Payload = payload
	default:
		return nil, fmt.Errorf("unknown payload type")
	}
	return event, nil
}

func constructDFTPayload(rawPayload map[string]interface{}) (*models.DFTEventPayload, error) {
	payload := new(models.DFTEventPayload)
	payload.InputFile = rawPayload["input_file"].(string)
	return payload, nil
}

func marshalEventToNats(event *models.Event) ([]byte, error) {
	return json.Marshal(struct {
		models.Event
		Topic string `json:"-"`
	}{
		Event: *event,
	})
}
