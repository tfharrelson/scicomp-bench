package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"

	"github.com/tfharrelson/scicomp-bench/pkg/api"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

var (
	localDB       db.DB
	localEventBus events.Bus
)

func InitHandlers(d db.DB, bus events.Bus) {
	localDB = d
	localEventBus = bus
}

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("app", "resources", "index.html"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	apiReq := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(body))
	apiReq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	api.Login(rec, apiReq, localDB)

	if rec.Code == http.StatusOK {
		for _, cookie := range rec.Result().Cookies() {
			http.SetCookie(w, cookie)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Login failed", rec.Code)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		http.Error(w, "All fields required", http.StatusBadRequest)
		return
	}

	body := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, username, email, password)
	apiReq := httptest.NewRequest("POST", "/api/v1/signup", bytes.NewBufferString(body))
	apiReq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	api.SignUp(rec, apiReq, localDB)

	if rec.Code == http.StatusOK {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	msg, _ := io.ReadAll(rec.Body)
	http.Error(w, string(msg), rec.Code)
}

func SubmitJob(w http.ResponseWriter, r *http.Request) {
	jobName := strings.TrimSpace(r.FormValue("job_name"))
	jobType := models.JobType(r.FormValue("type"))
	inputFile := r.FormValue("input_file")

	if jobName == "" {
		http.Error(w, "Job name required", http.StatusBadRequest)
		return
	}

	body := fmt.Sprintf(`{"type":"%s","job_name":"%s","input_file":"%s"}`, jobType, jobName, inputFile)
	apiReq := httptest.NewRequest("POST", "/api/v1/submit", bytes.NewBufferString(body))
	apiReq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	api.SubmitJob(rec, apiReq, localDB, localEventBus)

	if rec.Code == http.StatusOK {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	msg, _ := io.ReadAll(rec.Body)
	http.Error(w, string(msg), rec.Code)
}
