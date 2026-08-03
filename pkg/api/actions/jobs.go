package actions

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

func SubmitJob(eventBus events.Bus, request models.SubmitJobRequest) error {
	var payload models.EventPayload
	switch request.Type {
	case models.DFTJob:
		payload = &models.DFTEventPayload{InputFile: request.InputFile}
	case models.DummyJob:
		payload = &models.DummyEventPayload{}
	default:
		return fmt.Errorf("unknown request type: %s", request.Type)
	}

	err := eventBus.Publish(&models.Event{
		ID:      uuid.New(),
		Version: 1,
		JobName: request.JobName,
		Type:    request.Type,
		Payload: payload,
	})
	if err != nil {
		return err
	}
	return nil
}
