package app

import "github.com/tfharrelson/scicomp-bench/pkg/models"

type Api interface {
	Login(models.LoginRequest) (*models.AuthResponse, ApplicationError)
	SignUp(models.SignUpRequest) (*models.AuthResponse, ApplicationError)
	SubmitJob(models.SubmitJobRequest) ApplicationError
}
