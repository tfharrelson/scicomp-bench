// Package tools contains all tool implementations for the scicomp-bench application
package tools

import "github.com/tfharrelson/scicomp-bench/pkg/models"

type Tool interface {
	Run(models.Event) error
}
