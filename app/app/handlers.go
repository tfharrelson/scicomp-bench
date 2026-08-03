package app

import (
	"net/http"
	"strings"
	"time"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tfharrelson/scicomp-bench/app/templates"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

var (
	localDB       db.DB
	localEventBus events.Bus
	apiStore      Api
)

func InitHandlers(d db.DB, bus events.Bus, api Api) {
	localDB = d
	localEventBus = bus
	apiStore = api
}

func createCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	if err := templates.IndexPage().Render(r.Context(), w); err != nil {
		http.Error(w, "Couldn't render index page", http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	sse := datastar.NewSSE(w, r)
	if username == "" || password == "" {
		if err := sse.PatchElementTempl(templates.LoginError("Username and password required")); err != nil {
			http.Error(w, "Couldn't patch login error message", http.StatusInternalServerError)
			return
		}
		return
	}

	request := models.LoginRequest{Username: username, Password: password}
	resp, err := apiStore.Login(request)
	if err != nil {
		if err := sse.PatchElementTempl(templates.LoginError(err.Message())); err != nil {
			http.Error(w, "Couldn't patch login error message", http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, createCookie(resp.Token))
	if err := sse.PatchElementTempl(templates.LoginSuccess("Login successful!")); err != nil {
		http.Error(w, "Couldn't patch login toast", http.StatusInternalServerError)
		return
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	sse := datastar.NewSSE(w, r)
	if username == "" || email == "" || password == "" {
		if err := sse.PatchElementTempl(templates.SignUpError("All fields required")); err != nil {
			http.Error(w, "Couldn't patch signup error message", http.StatusInternalServerError)
			return
		}
		return
	}

	request := models.SignUpRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	resp, err := apiStore.SignUp(request)
	if err != nil {
		if err := sse.PatchElementTempl(templates.SignUpError(err.Message())); err != nil {
			http.Error(w, "Couldn't patch signup error message", http.StatusInternalServerError)
			return
		}
		return
	}
	if err := sse.PatchElementTempl(templates.SignUpToast("Account created! Automatically signing you in.")); err != nil {
		http.Error(w, "Couldn't patch sign up toast", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, createCookie(resp.Token))
	if err := sse.PatchElementTempl(templates.LoginSuccess("Login successful!")); err != nil {
		http.Error(w, "Couldn't patch login toast", http.StatusInternalServerError)
		return
	}
}

func SubmitJob(w http.ResponseWriter, r *http.Request) {
	// TODO: check auth
	sse := datastar.NewSSE(w, r)

	jobName := strings.TrimSpace(r.FormValue("job_name"))
	jobType := models.JobType(r.FormValue("type"))
	inputFile := r.FormValue("input_file")

	if jobName == "" {
		if err := sse.PatchElementTempl(templates.SubmitJobError("Job name required")); err != nil {
			http.Error(w, "Couldn't patch job error", http.StatusInternalServerError)
			return
		}
		return
	}

	request := models.SubmitJobRequest{
		Type:      jobType,
		JobName:   jobName,
		InputFile: inputFile,
	}

	err := apiStore.SubmitJob(request)
	if err != nil {
		if err := sse.PatchElementTempl(templates.SubmitJobError(err.Message())); err != nil {
			http.Error(w, "Couldn't patch job error", http.StatusInternalServerError)
			return
		}
	}
	if err := sse.PatchElementTempl(templates.SubmitJobToast("Job submitted successfully.")); err != nil {
		http.Error(w, "Couldn't patch job submitted toast", http.StatusInternalServerError)
		return
	}
}
