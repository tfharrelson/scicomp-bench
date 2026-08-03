package api

import (
	"encoding/json"
	"net/http"

	"github.com/tfharrelson/scicomp-bench/pkg/api/actions"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status": "ok"}`))
	if err != nil {
		// TODO: figure out how to handle this properly - prob need to retry
		panic(err)
	}
}

func SignUp(w http.ResponseWriter, r *http.Request, db db.DB) {
	w.Header().Set("Content-Type", "application/json")
	var request models.SignUpRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&request)
	if err != nil {
		http.Error(w, `{"message": "Bad Request"}`, http.StatusBadRequest)
		return
	}

	resp, err := actions.SignUp(db, request)
	if err != nil {
		println(err.Error())
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request, db db.DB) {
	w.Header().Set("Content-Type", "application/json")
	var request models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, `{"message": "Bad Request"}`, http.StatusBadRequest)
	}

	resp, err := actions.Login(db, request)
	if err != nil {
		// TODO: create different error types to differentiate response codes
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
}

func SubmitJob(w http.ResponseWriter, r *http.Request, db db.DB, eventBus events.Bus) {
	w.Header().Set("Content-Type", "application/json")
	var request models.SubmitJobRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, `{"message": "Bad Request"}`, http.StatusBadRequest)
		return
	}

	if err := actions.SubmitJob(eventBus, request); err != nil {
		// TODO: handle mapping of different error types to different error http codes
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
}
