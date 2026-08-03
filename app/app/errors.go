package app

import (
	"fmt"
	"net/http"
)

type ApplicationError interface {
	Message() string
	Code() int
	Error() string
}

func fmtError(e ApplicationError) string {
	return fmt.Sprintf("LoginError: %s, code: %d", e.Message(), e.Code())
}

type LoginError struct {
	code int
}

func (e *LoginError) Message() string {
	switch e.code {
	case http.StatusUnauthorized:
		return "Username and password don't match"
	case http.StatusBadRequest:
		return "Username or password is empty"
	default:
		return "Login failed for some unknown reason"
	}
}

func (e *LoginError) Code() int {
	return e.code
}

func (e *LoginError) Error() string {
	return fmtError(e)
}

type SignUpError struct {
	code int
}

func (e *SignUpError) Message() string {
	switch e.code {
	case http.StatusBadRequest:
		return "Username, email, or password is empty"
	default:
		return "Signup failed for some unknown reason"
	}
}

func (e *SignUpError) Code() int {
	return e.code
}

func (e *SignUpError) Error() string {
	return fmtError(e)
}

type SubmitJobError struct {
	code int
}

func (e *SubmitJobError) Message() string {
	switch e.code {
	case http.StatusBadRequest:
		return "Job name or input file is empty"
	default:
		return fmt.Sprintf("Job submission failed internally, code: %d", e.code)
	}
}

func (e *SubmitJobError) Code() int {
	return e.code
}

func (e *SubmitJobError) Error() string {
	return fmtError(e)
}
