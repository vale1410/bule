package sat

import (
	"strconv"
	"strings"
)

type Pred string

type Atom interface {
	Id() string
	Dom() uint
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

func (a Atom1) Dom() uint  { return 1 }
func (a AtomP) Dom() uint  { return 1 }
func (a AtomP1) Dom() uint { return 1 }
func (a AtomP2) Dom() uint { return 1 }
func (a AtomP3) Dom() uint { return 1 }

func (a Atom1) Id() string  { return strconv.Itoa(a.V) }
func (a AtomP) Id() string  { return string(a.P) }
func (a AtomP1) Id() string { return string(a.P) + "(" + strconv.Itoa(a.V) + ")" }
func (a AtomP2) Id() string {
	return string(a.P) + "(" + strconv.Itoa(a.V1) + "," + strconv.Itoa(a.V2) + ")"
}
func (a AtomP3) Id() string {
	return string(a.P) + "(" + strconv.Itoa(a.V1) + "," + strconv.Itoa(a.V2) + "," + strconv.Itoa(a.V3) + ")"
}

func NewAtom1(v int) Atom                   { return Atom1{v} }
func NewAtomP(p Pred) Atom                  { return AtomP{p} }
func NewAtomP1(p Pred, v int) Atom          { return AtomP1{p, v} }
func NewAtomP2(p Pred, v1, v2 int) Atom     { return AtomP2{p, v1, v2} }
func NewAtomP3(p Pred, v1, v2, v3 int) Atom { return AtomP3{p, v1, v2, v3} }

// TODO; once demanded do a generic atom type
//type AtomPN struct {

type Literal struct {
	Sign bool
	A    Atom
}

func NewLit(s string) Literal { return Literal{true, AtomP{Pred(s)}} }

func Neg(l Literal) Literal {
	l.Sign = !l.Sign
	return l
}

func (l Literal) ToTxt() (s string) {
	if !l.Sign {
		s += " ~"
	} else {
		s += " "
	}
	//s += "x"
	s += l.A.Id()
	//s += " "
	return
}

func (l Literal) ToPBO() (s string) {
	if !l.Sign {
		panic("PBO prining not accept negated literals!")
	}
	s = strings.Replace(l.A.Id(), ",", "_", -1)
	s = strings.Replace(s, "(", "_", -1)
	s = strings.Replace(s, ")", "", -1)
	return
}

func (l Literal) ToTex() (s string) {
	if !l.Sign {
		s += "\\bar "
	}
	//s += "x_{"
	s += l.A.Id()
	//s += "}"
	return
}
