package circuit

import (
//	"fmt"
)

type GateType int

const (
	Input GateType = iota
	AND
	OR
	NEG
)

type Gate struct {
	Typ GateType
	In  []*Gate
	Out []*Gate
}

type Circuit struct {
	Name string
	In   []*Gate
	Out  []*Gate
}

func (c *Circuit) printDot(filename string) {

}

func (c *Circuit) eval() {

}

func build(typ GateType, input ...*Gate) *Gate {

	var gate Gate

	gate.In = input
	gate.Typ = typ
	for _, x := range input {
		x.Out = append(x.Out, &gate)
	}

	switch typ {
	case AND:
		if len(gate.In) != 2 {
			panic("And gate has too many inputs")
		}
	case OR:
		if len(gate.In) != 2 {
			panic("Or gate has too many inputs")
		}
	case NEG:
		if len(gate.In) != 1 {
			panic("Neg gate has too many inputs")
		}
	default:
		panic("GateType  not implemented yet")
	}

	return &gate
}
