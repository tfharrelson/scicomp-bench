package models

type SimulationSnapshot struct {
	Energy   float64  `json:"energy"`
	Atoms    []Atom   `json:"atoms"`
	Forces   []Force  `json:"forces"`
	UnitCell UnitCell `json:"cell"`
}

type Atom struct {
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Mass     float64  `json:"mass"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Force struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type UnitCell struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Alpha float64 `json:"alpha"`
	Beta  float64 `json:"beta"`
	Gamma float64 `json:"gamma"`
}
