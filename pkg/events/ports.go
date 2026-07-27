package events

import "github.com/tfharrelson/scicomp-bench/pkg/models"

type Bus interface {
	Publish(*models.Event) error
	Subscribe(models.Topic, func(*models.Event) error) error
}
