package db

import (
	"errors"

	"github.com/tfharrelson/scicomp-bench/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type InMemoryDB struct {
	data   map[string]string
	events map[string]any
}

func NewInMemoryDB() *InMemoryDB {
	table := make(map[string]string)
	events := make(map[string]any)
	return &InMemoryDB{data: table, events: events}
}

func (d *InMemoryDB) CreateUser(username, email, password string) error {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	d.data[username] = string(hashedPwd)
	return nil
}

func (d *InMemoryDB) GetUserPasswordHash(username string) (string, error) {
	res, ok := d.data[username]
	if !ok {
		return "", errors.New("no user in db")
	}
	return res, nil
}

func (d *InMemoryDB) PublishEvent(event *models.Event) error {
	d.events[event.ID.String()] = event
	return nil
}
