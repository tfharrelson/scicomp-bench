// Package actions contains the domain actions of Scicomp Bench services
package actions

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte = []byte(os.Getenv("JWT_KEY"))

func Login(db db.DB, request models.LoginRequest) (*models.AuthResponse, error) {
	bcryptHash, err := db.GetUserPasswordHash(request.Username)
	if err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(bcryptHash), []byte(request.Password)); err != nil {
		return nil, err
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
		return nil, err
	}
	return &models.AuthResponse{Token: tokenString}, nil
}

func SignUp(db db.DB, request models.SignUpRequest) (*models.AuthResponse, error) {
	err := db.CreateUser(request.Username, request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	loginRequest := models.LoginRequest{
		Username: request.Username,
		Password: request.Password,
	}
	return Login(db, loginRequest)
}

type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TODO: should this be public?
func CheckToken(token string) (*CustomClaims, error) {
	// TODO: obviously fix this
	// would like to implement a public/private key pair for this
	tok, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(t *jwt.Token) (any, error) { return jwtKey, nil })
	if err != nil {
		return nil, err
	}
	claim, ok := tok.Claims.(*CustomClaims)
	if !ok {
		panic("claims are incorrectly configured")
	}
	if !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claim, nil
}
