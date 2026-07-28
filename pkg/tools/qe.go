package tools

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tfharrelson/scicomp-bench/pkg/db"
	"github.com/tfharrelson/scicomp-bench/pkg/events"
	"github.com/tfharrelson/scicomp-bench/pkg/models"
)

type QuantumEspressoResponse struct {
	XMLName xml.Name              `xml:"espresso"`
	Output  QuantumEspressoOutput `xml:"output"`
}

type QuantumEspressoOutput struct {
	TotalEnergy     TotalEnergy       `xml:"total_energy"`
	AtomicSpecies   []QEAtomicSpecies `xml:"atomic_species"`
	AtomicStructure QEAtomicStructure `xml:"atomic_structure"`
	Forces          QEMatrix          `xml:"forces"`
}

type QEAtomicSpecies struct {
	Species QESpecies `xml:"species"`
}

type QESpecies struct {
	Mass float64 `xml:"mass"`
	Name string  `xml:"name"`
}

type QEAtomicStructure struct {
	AtomicPositions QEAtomicPositions `xml:"atomic_positions"`
	Cell            QECell            `xml:"cell"`
}

type QEAtomicPositions struct {
	Atom []string `xml:"atom"`
}

type QECell struct {
	A1 string `xml:"a1"`
	A2 string `xml:"a2"`
	A3 string `xml:"a3"`
}

type TotalEnergy struct {
	Energy float64 `xml:"etot"`
}

type QEMatrix struct {
	Rank    string `xml:"rank,attr"`
	RawDims string `xml:"dims,attr"`
	Order   string `xml:"order,attr,omitempty"`
	Values  string `xml:",chardata"`
}

type QuantumEspressoTool struct {
	db  db.DB
	bus events.Bus
}

func NewQuantumEspressoTool(db db.DB, eventBus events.Bus) *QuantumEspressoTool {
	return &QuantumEspressoTool{db, eventBus}
}

func (t *QuantumEspressoTool) Run(event models.Event) error {
	// need to create a temporary file to create the input spec needed to run QE
	jobName := event.JobName
	if len(jobName) == 0 {
		jobName = "default"
	}
	runName := fmt.Sprintf("qerun-%s-%s", jobName, event.ID.String())
	infile, err := os.CreateTemp("", fmt.Sprintf("%s-input", runName))
	if err != nil {
		return err
	}
	defer infile.Close()

	if event.Type != models.DFTJob {
		return fmt.Errorf("unsupported job type: %s", event.Type)
	}
	dftPayload, ok := event.Payload.(models.DFTEventPayload)
	if !ok {
		return fmt.Errorf("invalid payload type: %T, expected DFT event payload", event.Payload)
	}

	n, err := infile.Write([]byte(dftPayload.InputFile))
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("failed to write input file")
	}
	if _, err := infile.Seek(0, 0); err != nil {
		return err
	}

	// prepare output file that gets written to by QE
	outfile, err := os.CreateTemp("", fmt.Sprintf("%s-output", runName))
	if err != nil {
		return err
	}
	defer outfile.Close()

	// run the qe cli tool
	cmd := exec.Command("pw.x")
	dirName := filepath.Dir(infile.Name())
	cmd.Dir = dirName
	fmt.Printf("cmd dir: %s\nin file: %s\nout file: %s\n", dirName, infile.Name(), outfile.Name())
	cmd.Stdin = infile
	cmd.Stdout = outfile
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	// parse the xml file to get the juicy datas
	xmlFile, err := os.Open(filepath.Join(dirName, "pwscf.save/data-file-schema.xml"))
	if err != nil {
		return err
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		return err
	}

	var qeResponse QuantumEspressoResponse
	if err := xml.Unmarshal(byteValue, &qeResponse); err != nil {
		return err
	}

	return nil
}
