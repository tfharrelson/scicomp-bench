package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tfharrelson/scicomp-bench/pkg/db"
)

func setupTestHandlers() {
	InitHandlers(db.NewInMemoryDB(), nil)
}

func TestIndex(t *testing.T) {
	setupTestHandlers()

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	Index(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 (file not found in test), got %d", rec.Code)
	}
}

func TestLoginMissingFields(t *testing.T) {
	setupTestHandlers()

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"missing username", "", "password"},
		{"missing password", "user", ""},
		{"missing both", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := "username=" + tt.username + "&password=" + tt.password
			req := httptest.NewRequest("POST", "/login", strings.NewReader(form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			Login(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestSignupMissingFields(t *testing.T) {
	setupTestHandlers()

	tests := []struct {
		name     string
		username string
		email    string
		password string
	}{
		{"missing username", "", "test@test.com", "pass"},
		{"missing email", "user", "", "pass"},
		{"missing password", "user", "test@test.com", ""},
		{"missing all", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := "username=" + tt.username + "&email=" + tt.email + "&password=" + tt.password
			req := httptest.NewRequest("POST", "/signup", strings.NewReader(form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			Signup(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestSubmitJobMissingFields(t *testing.T) {
	setupTestHandlers()

	tests := []struct {
		name      string
		jobName   string
		jobType   string
		inputFile string
	}{
		{"missing job name", "", "dft", "input"},
		{"empty job name", "   ", "dft", "input"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := "job_name=" + tt.jobName + "&type=" + tt.jobType + "&input_file=" + tt.inputFile
			req := httptest.NewRequest("POST", "/submit", strings.NewReader(form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			SubmitJob(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}
