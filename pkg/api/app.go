package api

import (
	"log"
	"net/http"

	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
)

func CreateApp() {
	mux := http.NewServeMux()

	config := NewConfig()
	db := db.NewInMemoryDB()
	bus, err := events.NewNatsEventBus(config.EventBus.URL, config.EventBus.MaxReconnects, config.EventBus.WaitTime, db)
	if err != nil {
		log.Fatalf("Failed creating event bus: %v", err)
	}
	mux.HandleFunc("GET /api/v1/health", Health)
	mux.HandleFunc(
		"POST /api/v1/signup",
		func(w http.ResponseWriter, r *http.Request) {
			SignUp(w, r, db)
		},
	)
	mux.HandleFunc(
		"POST /api/v1/login",
		func(w http.ResponseWriter, r *http.Request) {
			Login(w, r, db)
		},
	)
	mux.HandleFunc("POST /api/v1/submit", func(w http.ResponseWriter, r *http.Request) {
		SubmitJob(w, r, db, bus)
	})
	// mux.HandleFunc("GET /api/v1/stream", StreamEvents)
	// mux.HandleFunc("GET /api/v1/jobs", GetJobs)
}
