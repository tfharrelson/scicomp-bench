package app

import (
	"log"
	"net/http"
	"strconv"

	natsserver "github.com/nats-io/nats-server/v2/server"

	"github.com/tfharrelson/scicomp-bench/app/resources"
	"github.com/tfharrelson/scicomp-bench/pkg/api"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
)

func CreateApp() http.Handler {
	mux := http.NewServeMux()

	// TODO: this server should have its own config separate from the api one
	config := api.NewConfig()

	// TODO: remove this once i set up nats containers and my dev environment
	portInt, err := strconv.Atoi(config.EventBus.URL.Port())
	if err != nil {
		panic(err)
	}
	opts := natsserver.Options{
		Host: config.EventBus.URL.Host,
		Port: portInt,
	}
	ns, err := natsserver.NewServer(&opts)
	if err != nil {
		panic(err)
	}
	println("created nats server")
	defer ns.Shutdown()

	localDB := db.NewInMemoryDB()
	println("created db")
	bus, err := events.NewNatsEventBus(config.EventBus.URL, config.EventBus.MaxReconnects, config.EventBus.WaitTime, localDB)
	if err != nil {
		log.Fatalf("Failed creating event bus: %v", err)
	}
	println("created event bus")

	apiStore := NewInProcessApiStore(localDB, bus)

	InitHandlers(localDB, bus, apiStore)

	mux.Handle("GET /static/", resources.Handler())
	mux.HandleFunc("GET /", Index)

	mux.HandleFunc("GET /login", RenderLogin)
	mux.HandleFunc("GET /signup", RenderSignup)
	mux.HandleFunc("GET /submitjob", RenderJobForm)
	mux.HandleFunc("POST /login", Login)
	mux.HandleFunc("POST /signup", Signup)
	mux.HandleFunc("POST /submitjob", SubmitJob)

	return mux
}
