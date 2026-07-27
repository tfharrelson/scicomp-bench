// Package models contains all domain models for the scicomp-bench application
package models

import (
	"github.com/google/uuid"
)

type Topic string

type Event struct {
	ID      uuid.UUID    `json:"id"`
	Version int          `json:"version"`
	Topic   Topic        `json:"topic"`
	JobName string       `json:"job_name"`
	Type    JobType      `json:"type"`
	Payload EventPayload `json:"payload"`
}

type EventPayload interface {
	isEventPayload()
}

type DFTEventPayload struct {
	InputFile string `json:"input_file"`
}

func (e *DFTEventPayload) isEventPayload() {}

type DummyEventPayload struct{}

func (e *DummyEventPayload) isEventPayload() {}
