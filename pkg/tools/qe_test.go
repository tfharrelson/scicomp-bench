package tools

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

type NoOpBus struct{}

func (e *NoOpBus) Publish(event *models.Event) error {
	return nil
}

func (e *NoOpBus) Subscribe(topic models.Topic, handler func(*models.Event) error) error {
	return nil
}

func qeInputFile() string {
	pseudoDir := os.Getenv("PSEUDO_DIR")
	if pseudoDir == "" {
		panic("no pseudo_dir set")
	}

	// insert pseudo_dir into the input file
	return fmt.Sprintf(`&control
    calculation = 'scf'
    restart_mode='from_scratch',
    tstress = .true.
    tprnfor = .true.
    pseudo_dir = '%s',
 /
 &system
    ibrav=  2, celldm(1) =10.20, nat=  2, ntyp= 1,
    ecutwfc =18.0,
 /
 &electrons
    mixing_mode = 'plain'
    mixing_beta = 0.7
    conv_thr =  1.0d-5
 /
ATOMIC_SPECIES
 Si  28.086  Si.us.pbe.z_4.ld1.psl.v1.0.0-high.upf
ATOMIC_POSITIONS alat
 Si 0.00 0.00 0.00
 Si 0.25 0.25 0.25
K_POINTS gamma`,
		pseudoDir,
	)
}

func TestQERun(t *testing.T) {
	// unforutnately, this one requires qe to be installed
	// and that lives outside of the go ecosystem, so this can
	// only be run on a properly configured machine for now
	t.Parallel()

	qeTool := NewQuantumEspressoTool(db.NewInMemoryDB(), &NoOpBus{})
	tests := []struct {
		name  string
		event models.Event
	}{
		{
			"happy path", models.Event{
				ID:      uuid.New(),
				Version: 1,
				Topic:   "test",
				JobName: "test",
				Type:    "dft", Payload: models.DFTEventPayload{InputFile: qeInputFile()},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := qeTool.Run(test.event)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
