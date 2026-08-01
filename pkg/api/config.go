package api

import (
	"net/url"
	"os"
	"strconv"

	natsserver "github.com/nats-io/nats-server/v2/server"
)

type Env string

const (
	Prod  Env = "prod"
	Dev   Env = "dev"
	Stage Env = "stage"
)

type Config struct {
	EventBus EventBusConfig
	DB       DBConfig
}

type EventBusConfig struct {
	URL           *url.URL
	MaxReconnects int
	WaitTime      int
}

type DBConfig struct{}

func NewConfig() *Config {
	env := os.Getenv("ENV")
	switch env {
	case "prod":
		return newProdConfig()
	case "dev":
		return newDevConfig()
	case "stage":
		return newStageConfig()
	default:
		return newDevConfig()
	}
}

func newProdConfig() *Config {
	brokerUrl, err := url.Parse(os.Getenv("BROKER_URL"))
	if err != nil {
		panic(err)
	}
	// _ := os.Getenv("DB_URL")
	maxReconnects, err := strconv.Atoi(os.Getenv("BROKER_MAX_RECONNECTS"))
	if err != nil {
		panic(err)
	}
	waitTime, err := strconv.Atoi(os.Getenv("BROKER_WAIT_TIME"))
	if err != nil {
		panic(err)
	}

	return &Config{
		EventBus: EventBusConfig{
			URL:           brokerUrl,
			MaxReconnects: maxReconnects,
			WaitTime:      waitTime,
		},
		DB: DBConfig{},
	}
}

func newDevConfig() *Config {
	opts := natsserver.Options{Port: -1}
	ns, err := natsserver.NewServer(&opts)
	if err != nil {
		panic(err)
	}

	natsURL, err := url.Parse(ns.ClientURL())
	if err != nil {
		panic(err)
	}

	return &Config{
		EventBus: EventBusConfig{
			URL:           natsURL,
			MaxReconnects: 0,
			WaitTime:      1,
		},
		DB: DBConfig{},
	}
}

func newStageConfig() *Config {
	// TODO: implement stage config
	return newDevConfig()
}
