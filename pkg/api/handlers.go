package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
	"golang.org/x/crypto/bcrypt"
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

	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
	}

	// persist it to db
	err = db.CreateUser(request.Username, request.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
	}
}

var jwtKey []byte = []byte(os.Getenv("JWT_KEY"))

func Login(w http.ResponseWriter, r *http.Request, db db.DB) {
	w.Header().Set("Content-Type", "application/json")
	var request models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, `{"message": "Bad Request"}`, http.StatusBadRequest)
	}

	bcryptHash, err := db.GetUserPasswordHash(request.Username)
	if err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(bcryptHash), []byte(request.Password)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
	expirationTime := time.Now().Add(15 * time.Minute)
	claimsMap := map[string]any{
		"username": request.Username,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claimsMap))
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
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

	// download the file from the header
	var payload models.EventPayload
	switch request.Type {
	case models.DFTJob:
		payload = &models.DFTEventPayload{InputFile: request.InputFile}
	case models.DummyJob:
		payload = &models.DummyEventPayload{}
	default:
		http.Error(w, `{"message": "Bad Request"}`, http.StatusBadRequest)
		return
	}

	err := eventBus.Publish(&models.Event{
		ID:      uuid.New(),
		Version: 1,
		JobName: request.JobName,
		Type:    request.Type,
		Payload: payload,
	})
	if err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
}
