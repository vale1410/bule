package circuit

import (
	"fmt"
	"testing"
)

func TestExample(test *testing.T) {

	fmt.Println("check")
    buildXOR()
}

func buildXOR() (c Circuit) {

	var in1, in2 Gate

	in1.Typ = Input
	in2.Typ = Input

	nIn1 := build(NEG, &in1)
	nIn2 := build(NEG, &in2)

	and1 := build(AND, &in1, nIn2)
	and2 := build(AND, nIn1, &in2)

	or := build(OR, and1, and2)

	c.Name = "xor"
	c.In = []*Gate{&in1, &in2}
	c.Out = []*Gate{or}
    fmt.Println(c)
	return
}

