package tools

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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
	Name string  `xml:"name,attr"`
}

type QEAtomicStructure struct {
	AtomicPositions QEAtomicPositions `xml:"atomic_positions"`
	Cell            QECell            `xml:"cell"`
}

type QEAtomicPositions struct {
	Atom []QEAtom `xml:"atom"`
	Name string   `xml:"name,attr"`
}

type QEAtom struct {
	Position string `xml:",chardata"`
	Name     string `xml:"name,attr"`
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

	// convert the qe response to the domain model and persist it
	speciesMap := make(map[string]float64)
	for _, atomicSpecies := range qeResponse.Output.AtomicSpecies {
		fmt.Printf("species = %+v\n", atomicSpecies)
		speciesMap[atomicSpecies.Species.Name] = atomicSpecies.Species.Mass
	}
	var atoms []models.Atom
	for _, atom := range qeResponse.Output.AtomicStructure.AtomicPositions.Atom {
		vec, err := parseQEVector(atom.Position)
		if err != nil {
			return err
		}
		mass, ok := speciesMap[atom.Name]
		if !ok {
			return fmt.Errorf("no mass found for atom %s, map = %+v", atom.Name, speciesMap)
		}
		atoms = append(atoms, models.Atom{
			Type: atom.Name,
			Mass: mass,
			Position: models.Vector{
				X: vec[0],
				Y: vec[1],
				Z: vec[2],
			},
		})
	}

	a, err := parseQEVector(qeResponse.Output.AtomicStructure.Cell.A1)
	if err != nil {
		return err
	}
	b, err := parseQEVector(qeResponse.Output.AtomicStructure.Cell.A2)
	if err != nil {
		return err
	}
	c, err := parseQEVector(qeResponse.Output.AtomicStructure.Cell.A3)
	if err != nil {
		return err
	}
	cell := models.UnitCell{
		A: models.Vector{X: a[0], Y: a[1], Z: a[2]},
		B: models.Vector{X: b[0], Y: b[1], Z: b[2]},
		C: models.Vector{X: c[0], Y: c[1], Z: c[2]},
	}

	forces, err := parseQEForces(qeResponse.Output.Forces)
	if err != nil {
		return err
	}

	simulation := models.SimulationSnapshot{
		ID:       event.ID,
		Energy:   qeResponse.Output.TotalEnergy.Energy,
		Atoms:    atoms,
		Forces:   forces,
		UnitCell: cell,
	}

	if err := t.db.PersistSimulation(simulation); err != nil {
		return err
	}

	return nil
}

func parseQEVector(vecString string) ([]float64, error) {
	vecStrings := strings.Fields(vecString)
	if len(vecStrings) != 3 {
		return nil, fmt.Errorf(
			"invalid number of positional dimensions in vector %d, expected 3, in %s",
			len(vecStrings),
			vecString,
		)
	}
	x, err := strconv.ParseFloat(vecStrings[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x position: %v", vecString)
	}
	y, err := strconv.ParseFloat(vecStrings[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse y position: %v", vecString)
	}
	z, err := strconv.ParseFloat(vecStrings[2], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse z position: %v", vecString)
	}
	return []float64{x, y, z}, nil
}

func parseQEForces(qeForces QEMatrix) ([]models.Vector, error) {
	valueStrings := strings.Fields(qeForces.Values)
	if len(valueStrings)%3 != 0 {
		return nil, fmt.Errorf("invalid number of elements in force values")
	}
	values := make([]float64, len(valueStrings))
	for i, value := range valueStrings {
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse force value: %v", value)
		}
		values[i] = value
	}

	// initialize the forces
	var forces []models.Vector
	for i := 0; i < len(values)/3; i++ {
		forces = append(forces, models.Vector{})
	}

	switch qeForces.Order {
	case "C": // row major order
		for i := 0; i < len(values)/3; i++ {
			forces[i].X = values[i*3]
			forces[i].Y = values[i*3+1]
			forces[i].Z = values[i*3+2]
		}
	default: // column major order
		for i := 0; i < len(values)/3; i++ {
			forces[i].X = values[i]
			forces[i].Y = values[i+len(values)/3]
			forces[i].Z = values[i+2*len(values)/3]
		}
	}
	return forces, nil
}
