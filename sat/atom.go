package sat

import (
	"fmt"
	"os"
	"strconv"
)

type Literal struct {
	Sign bool
	Atom Atom
}

func Neg(l Literal) Literal {
	l.Sign = !l.Sign
	return l
}

type Atom interface {
	Id() string
}

type Pred string

type AtomP struct {
	P Pred
}

type Atom1 struct {
	V1 int
}

type AtomP1 struct {
	P  Pred
	V1 int
}

type AtomP2 struct {
	P  Pred
	V1 int
	V2 int
}

type AtomP3 struct {
	V1 int
	V2 int
	V3 int
}

type AtomPN struct {
	P  Pred
	V1 int
	V2 int
	V3 int
}

func (atom Atom) ToTex() (s string) {
	return strconv.Itoa(atom.V1)
}

func (l Literal) ToTxt() (s string) {
	if !l.Sign {
		s += "~"
	} else {
		s += " "
	}
	s += "x"
	s += l.Atom.ToTex()
	s += " "
	return
}

type Gen struct {
	nextId   int
	mapping  map[Atom]int
	Filename string
	out      *os.File
}
func IdGenerator(m int) (g Gen) {
	g.mapping = make(map[Atom]int, m)
	return
}

func (l Literal) ToTex() (s string) {
	if !l.Sign {
		s += "\\bar "
	}
	s += "x_{"
	s += l.Atom.ToTex()
	s += "}"
	return
}
func (g *Gen) putAtom(a Atom) {
	if id, b := g.mapping[a]; !b {
		g.nextId++
		id = g.nextId
		g.mapping[a] = id
	}
}

func (g *Gen) getId(a Atom) (id int) {
	id, b := g.mapping[a]

	if !b {
		g.nextId++
		id = g.nextId
		g.mapping[a] = id
	}

	return id
}

func (g *Gen) generateSymbolTable() []string {

	symbolTable := make([]string, len(g.mapping)+1)

	for atom, cnfId := range g.mapping {
		s := string(atom.P) + "("
		s += strconv.Itoa(atom.V1)
		s += ","
		s += strconv.Itoa(atom.V2)
		s += ")"
		symbolTable[cnfId] = s
	}

	return symbolTable
}

func (g *Gen) printSymbolTable(filename string) {

	symbolTable := g.generateSymbolTable()
	symbolFile, err := os.Create(filename)

	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := symbolFile.Close(); err != nil {
			panic(err)
		}
	}()

	// make a write buffer
	w := bufio.NewWriter(symbolFile)

	for i, s := range symbolTable {
		// write a chunk
		if _, err := w.Write([]byte(fmt.Sprintln(i, "\t:", s))); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}

}
