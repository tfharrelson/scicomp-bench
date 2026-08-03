package app

import (
	"net/http"
	"time"

	"github.com/tfharrelson/scicomp-bench/pkg/api/actions"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

type MockApiStore struct {
	success bool
}

func NewMockApiStore(success bool) *MockApiStore {
	return &MockApiStore{success: success}
}

func mockCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "session_token",
		Value:    "xyz123securevalue",
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,                 // Prevents JavaScript access
		Secure:   true,                 // Requires HTTPS connection
		SameSite: http.SameSiteLaxMode, // Controls cross-site request behavior
	}
}

func (a *MockApiStore) Login(request models.LoginRequest) (*models.AuthResponse, ApplicationError) {
	if a.success {
		return &models.AuthResponse{Token: "xyz123securevalue"}, nil
	}
	return nil, &LoginError{http.StatusUnauthorized}
}

func (a *MockApiStore) SignUp(request models.SignUpRequest) (*models.AuthResponse, ApplicationError) {
	if a.success {
		return &models.AuthResponse{Token: "xyz123securevalue"}, nil
	}
	return nil, &SignUpError{http.StatusBadRequest}
}

func (a *MockApiStore) SubmitJob(request models.SubmitJobRequest) ApplicationError {
	if a.success {
		return nil
	}
	return &SubmitJobError{http.StatusBadRequest}
}

type InProcessApiStore struct {
	DB  db.DB
	Bus events.Bus
}

func NewInProcessApiStore(db db.DB, bus events.Bus) *InProcessApiStore {
	return &InProcessApiStore{DB: db, Bus: bus}
}

func (a *InProcessApiStore) Login(request models.LoginRequest) (*models.AuthResponse, ApplicationError) {
	resp, err := actions.Login(a.DB, request)
	if err != nil {
		return nil, &LoginError{code: http.StatusInternalServerError}
	}
	return resp, nil
}

func (a *InProcessApiStore) SignUp(request models.SignUpRequest) (*models.AuthResponse, ApplicationError) {
	resp, err := actions.SignUp(a.DB, request)
	if err != nil {
		return nil, &LoginError{code: http.StatusInternalServerError}
	}
	return resp, nil
}

func (a *InProcessApiStore) SubmitJob(request models.SubmitJobRequest) ApplicationError {
	err := actions.SubmitJob(a.Bus, request)
	if err != nil {
		return &SubmitJobError{code: http.StatusInternalServerError}
	}
	return nil
}

// TODO: fill out the distributed implementation for sending a request to an actual api http server
