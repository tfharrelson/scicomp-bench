package api

import (
	"net/http"

	"github.com/tfharrelson/scicomp-bench/pkg/db"
)

func CreateApp() {
	mux := http.NewServeMux()

	db := db.NewInMemoryDB()
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
	// mux.HandleFunc("POST /api/v1/submit", SubmitJob)
	// mux.HandleFunc("GET /api/v1/stream", StreamEvents)
	// mux.HandleFunc("GET /api/v1/jobs", GetJobs)
}
