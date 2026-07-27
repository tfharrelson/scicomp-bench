package models

type JobInput interface {
	isJob() bool
}

type DFTJobInput struct {
	SmilesString string `json:"smiles"`
}

func (j *DFTJobInput) isJob() bool {
	return true
}
