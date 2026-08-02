package app

import (
	"log"
	"net/http"

	"github.com/tfharrelson/scicomp-bench/pkg/api"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
)

func CreateApp() http.Handler {
	mux := http.NewServeMux()

	config := api.NewConfig()
	localDB := db.NewInMemoryDB()
	bus, err := events.NewNatsEventBus(config.EventBus.URL, config.EventBus.MaxReconnects, config.EventBus.WaitTime, localDB)
	if err != nil {
		log.Fatalf("Failed creating event bus: %v", err)
	}

	InitHandlers(localDB, bus)

	mux.HandleFunc("GET /", Index)
	mux.HandleFunc("POST /login", Login)
	mux.HandleFunc("POST /signup", Signup)
	mux.HandleFunc("POST /submit", SubmitJob)

	return mux
}
