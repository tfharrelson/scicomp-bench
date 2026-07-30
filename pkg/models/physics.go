package models

import "github.com/google/uuid"

type SimulationSnapshot struct {
	ID       uuid.UUID `json:"id"`
	Energy   float64   `json:"energy"`
	Atoms    []Atom    `json:"atoms"`
	Forces   []Vector  `json:"forces"`
	UnitCell UnitCell  `json:"cell"`
}

type Atom struct {
	Type     string  `json:"type"`
	Position Vector  `json:"position"`
	Mass     float64 `json:"mass"`
}

type Vector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type UnitCell struct {
	A Vector `json:"a"`
	B Vector `json:"b"`
	C Vector `json:"c"`
}
