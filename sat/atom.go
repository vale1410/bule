package sat

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type Pred string

type Atom interface {
	Id() string
}

type AtomP struct {
	P Pred
}

type Atom1 struct {
	V int
}

type AtomP1 struct {
	P Pred
	V int
}

type AtomP2 struct {
	P  Pred
	V1 int
	V2 int
}

type AtomP3 struct {
	P  Pred
	V1 int
	V2 int
	V3 int
}

func (a Atom1) Id() string {
	return strconv.Itoa(a.V)
}

func (a AtomP) Id() string {
	return string(a.P)
}

func (a AtomP1) Id() string {
	return string(a.P) + "(" + strconv.Itoa(a.V) + ")"
}

func (a AtomP2) Id() string {
	return string(a.P) + "(" + strconv.Itoa(a.V1) + "," + strconv.Itoa(a.V2) + ")"
}

func (a AtomP3) Id() string {
	return string(a.P) + "(" + strconv.Itoa(a.V1) + "," + strconv.Itoa(a.V2) + "," + strconv.Itoa(a.V3) + ")"
}

func NewAtomP(p Pred) Atom {
	return AtomP{p}
}

func NewAtom1(v int) Atom {
	return Atom1{v}
}

func NewAtomP1(p Pred, v int) Atom {
	return AtomP1{p, v}
}

func NewAtomP2(p Pred, v1, v2 int) Atom {
	return AtomP2{p, v1, v2}
}

func NewAtomP3(p Pred, v1, v2, v3 int) Atom {
	return AtomP3{p, v1, v2, v3}
}

// TODO; once demanded do a generic atom type
//type AtomPN struct {

type Literal struct {
	Sign bool
	A    Atom
}

func Neg(l Literal) Literal {
	l.Sign = !l.Sign
	return l
}

func (l Literal) ToTxt() (s string) {
	if !l.Sign {
		s += "~"
	} else {
		s += " "
	}
	s += "x"
	s += l.A.Id()
	s += " "
	return
}

func (l Literal) ToTex() (s string) {
	if !l.Sign {
		s += "\\bar "
	}
	s += "x_{"
	s += l.A.Id()
	s += "}"
	return
}

type Gen struct {
	nextId   int
	mapping  map[string]int
	Filename string
	out      *os.File
}

func IdGenerator(m int) (g Gen) {
	g.mapping = make(map[string]int, m)
	return
}

func (g *Gen) putAtom(a Atom) {
	if id, b := g.mapping[a.Id()]; !b {
		g.nextId++
		id = g.nextId
		g.mapping[a.Id()] = id
	}
}

func (g *Gen) getId(a Atom) (id int) {
	id, b := g.mapping[a.Id()]

	if !b {
		g.nextId++
		id = g.nextId
		g.mapping[a.Id()] = id
	}

	return id
}

func (g *Gen) printSymbolTable(filename string) {

	symbolFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}
	// close on exit and check for its returned error
	defer func() {
		if err := symbolFile.Close(); err != nil {
			panic(err)
		}
	}()

	// make a write buffer
	w := bufio.NewWriter(symbolFile)

	for i, s := range g.mapping {
		// write a chunk
		if _, err := w.Write([]byte(fmt.Sprintln(i, "\t:", s))); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}

}
