package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tfharrelson/scicomp-bench/app/templates"
	"github.com/tfharrelson/scicomp-bench/pkg/api/actions"
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

func redirectToLogin(sse *datastar.ServerSentEventGenerator, w http.ResponseWriter) {
	if err := sse.ExecuteScript(`window.location.href = "/login`); err != nil {
		http.Error(w, "Couldn't redirect to login page", http.StatusInternalServerError)
	}
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

func RenderLogin(w http.ResponseWriter, r *http.Request) {
	if err := templates.LoginForm().Render(r.Context(), w); err != nil {
		http.Error(w, "Couldn't render login page", http.StatusInternalServerError)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Couldn't parse form", http.StatusInternalServerError)
		return
	}
	var request models.LoginRequest
	request.Username = r.FormValue("username")
	request.Password = r.FormValue("password")

	sse := datastar.NewSSE(w, r)
	if request.Username == "" || request.Password == "" {
		if err := sse.PatchElementTempl(templates.LoginError("Username and password required")); err != nil {
			http.Error(w, "Couldn't patch login error message", http.StatusInternalServerError)
			return
		}
		return
	}

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

func RenderSignup(w http.ResponseWriter, r *http.Request) {
	if err := templates.SignupForm().Render(r.Context(), w); err != nil {
		http.Error(w, "Couldn't render login page", http.StatusInternalServerError)
		return
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Couldn't parse form", http.StatusInternalServerError)
		return
	}
	var request models.SignUpRequest
	request.Username = r.FormValue("username")
	request.Email = r.FormValue("email")
	request.Password = r.FormValue("password")

	sse := datastar.NewSSE(w, r)
	if request.Username == "" || request.Email == "" || request.Password == "" {
		if err := sse.PatchElementTempl(templates.SignUpError("All fields required")); err != nil {
			http.Error(w, "Couldn't patch signup error message", http.StatusInternalServerError)
			return
		}
		return
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

	// redirect to bench page
	if err := sse.ExecuteScript(`
		window.location.href = "/submitjob"
	`); err != nil {
		http.Error(w, "Couldn't redirect to bench page", http.StatusInternalServerError)
		return
	}
}

func RenderJobForm(w http.ResponseWriter, r *http.Request) {
	// TODO: check auth
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		// redirect to login page with standard http lib
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	_, err = actions.CheckToken(tokenCookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// TODO: implement a welcome message from claims
	if err := templates.JobForm().Render(r.Context(), w); err != nil {
		http.Error(w, "Couldn't render job form", http.StatusInternalServerError)
	}
}

func SubmitJob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Couldn't parse form", http.StatusInternalServerError)
		return
	}
	fmt.Println("got submit job request")
	sse := datastar.NewSSE(w, r)
	// TODO: check auth
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		redirectToLogin(sse, w)
		return
	}
	_, err = actions.CheckToken(tokenCookie.Value)
	if err != nil {
		redirectToLogin(sse, w)
		return
	}

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

	appErr := apiStore.SubmitJob(request)
	if appErr != nil {
		if err := sse.PatchElementTempl(templates.SubmitJobError(appErr.Message())); err != nil {
			http.Error(w, "Couldn't patch job error", http.StatusInternalServerError)
			return
		}
	}
	if err := sse.PatchElementTempl(templates.SubmitJobToast("Job submitted successfully.")); err != nil {
		http.Error(w, "Couldn't patch job submitted toast", http.StatusInternalServerError)
		return
	}
}
