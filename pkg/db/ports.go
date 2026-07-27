// Package db contains all database adapters for the scicomp-bench application
package db

import "github.com/tfharrelson/scicomp-bench/pkg/models"

type DB interface {
	CreateUser(string, string, string) error
	GetUserPasswordHash(string) (string, error)
	PublishEvent(*models.Event) error
}
