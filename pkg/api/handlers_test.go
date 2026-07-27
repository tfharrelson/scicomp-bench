package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tfharrelson/scicomp-bench/pkg/db"
)

func decodeResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	if len(w.Body.Bytes()) == 0 {
		return nil
	}
	var result map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		fmt.Printf("hit error while decoding body %s: %s", w.Body.Bytes(), err)
		panic(err)
	}
	return result
}

func TestHealth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/health", strings.NewReader(""))
	w := httptest.NewRecorder()
	Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code to be %d, got %d", http.StatusOK, w.Code)
	}

	body := decodeResponse(w)
	if body["status"] != "ok" {
		t.Errorf("expected status to be 'ok', got %v", body["status"])
	}
}

func TestSignUp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"happy path", `{"username":"test","email":"test@test.com","password":"password"}`, http.StatusOK},
		{"bad request", `{"bad":"body"}`, http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/users", strings.NewReader(test.body))
			w := httptest.NewRecorder()
			db := db.NewInMemoryDB()

			SignUp(w, req, db)

			if w.Code != test.status {
				t.Errorf("expected status code to be %d, got %d", test.status, w.Code)
			}
			body := decodeResponse(w)
			if body != nil {
				if body["message"] != "Bad Request" {
					t.Errorf("expected error to be 'Bad Request', got %v", body["message"])
				}
			}
		})
	}
}
