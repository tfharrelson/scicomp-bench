package models

type JobType string

const (
	DummyJob JobType = "dummy"
	DFTJob   JobType = "dft"
)

type Status string

const (
	Requested Status = "requested"
	Started   Status = "started"
	Completed Status = "completed"
)
