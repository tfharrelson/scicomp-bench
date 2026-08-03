package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tfharrelson/scicomp-bench/app/templates"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

func setupTestHandlers() {
	InitHandlers(db.NewInMemoryDB(), &mockEventBus{}, NewMockApiStore(true))
}

type mockEventBus struct{}

func (m *mockEventBus) Publish(event *models.Event) error { return nil }
func (m *mockEventBus) Subscribe(topic models.Topic, handler func(*models.Event) error) error {
	return nil
}

func TestIndex(t *testing.T) {
	setupTestHandlers()

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	Index(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "<!doctype html>") {
		t.Error("expected HTML response")
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

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}

			body := rec.Body.String()
			if !strings.Contains(body, "Username and password required") {
				t.Error("expected error message")
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

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}

			body := rec.Body.String()
			if !strings.Contains(body, "All fields required") {
				t.Error("expected error message")
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

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}

			body := rec.Body.String()
			if !strings.Contains(body, "Job name required") {
				t.Error("expected error message")
			}
		})
	}
}

func TestLoginForm(t *testing.T) {
	buf := &strings.Builder{}
	err := templates.LoginForm().Render(context.Background(), buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()

	if !strings.Contains(html, `data-on:submit="@post('/login')"`) {
		t.Error("expected Datastar form submission")
	}
	if !strings.Contains(html, `id="login-error"`) {
		t.Error("expected error element")
	}
	if !strings.Contains(html, `id="login-success"`) {
		t.Error("expected success element")
	}
	if !strings.Contains(html, `name="username"`) {
		t.Error("expected username field")
	}
	if !strings.Contains(html, `name="password"`) {
		t.Error("expected password field")
	}
}

func TestSignupForm(t *testing.T) {
	buf := &strings.Builder{}
	err := templates.SignupForm().Render(context.Background(), buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()

	if !strings.Contains(html, `data-on:submit="@post('/signup')"`) {
		t.Error("expected Datastar form submission")
	}
	if !strings.Contains(html, `id="signup-error"`) {
		t.Error("expected error element")
	}
	if !strings.Contains(html, `id="signup-success"`) {
		t.Error("expected success element")
	}
	if !strings.Contains(html, `name="username"`) {
		t.Error("expected username field")
	}
	if !strings.Contains(html, `name="email"`) {
		t.Error("expected email field")
	}
	if !strings.Contains(html, `name="password"`) {
		t.Error("expected password field")
	}
}

func TestJobForm(t *testing.T) {
	buf := &strings.Builder{}
	err := templates.JobForm().Render(context.Background(), buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()

	if !strings.Contains(html, `data-on:submit="@post('/submit')"`) {
		t.Error("expected Datastar form submission")
	}
	if !strings.Contains(html, `id="job-error"`) {
		t.Error("expected error element")
	}
	if !strings.Contains(html, `id="job-success"`) {
		t.Error("expected success element")
	}
	if !strings.Contains(html, `name="job_name"`) {
		t.Error("expected job_name field")
	}
	if !strings.Contains(html, `name="type"`) {
		t.Error("expected type field")
	}
	if !strings.Contains(html, `name="input_file"`) {
		t.Error("expected input_file field")
	}
}

func TestIndexPage(t *testing.T) {
	buf := &strings.Builder{}
	err := templates.IndexPage().Render(context.Background(), buf)
	if err != nil {
		t.Fatal(err)
	}

	html := buf.String()

	if !strings.Contains(html, "<!doctype html>") {
		t.Error("expected DOCTYPE")
	}
	if !strings.Contains(html, `data-datastar`) {
		t.Error("expected Datastar body attribute")
	}
	if !strings.Contains(html, `data-on:submit="@post('/login')"`) {
		t.Error("expected login form in index")
	}
	if !strings.Contains(html, `data-on:submit="@post('/signup')"`) {
		t.Error("expected signup form in index")
	}
	if !strings.Contains(html, `data-on:submit="@post('/submit')"`) {
		t.Error("expected job form in index")
	}
}
