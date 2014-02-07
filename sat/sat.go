package sat

// todo: Call Solvers, get back result etc.

import (
	"bufio"
	"fmt"
	"github.com/vale1410/bule/sorters"
	"os"
	"strconv"
)

type Pred string

// we only allow two dimensional predicates
type Atom struct {
	P  Pred
	V1 int
	V2 int
}

type Literal struct {
	Sign bool
	Atom Atom
}

func Neg(l Literal) Literal {
	l.Sign = !l.Sign
	return l
}

type Clause struct {
	Desc     string
	Literals []Literal
}

type ClauseSet []Clause

func (cs *ClauseSet) AddClause(desc string, literals ...Literal) {
	*cs = append(*cs, Clause{desc, literals})
}

func (cs *ClauseSet) AddClauseSet(cl ClauseSet) {
	*cs = append(*cs, cl...)
}

type Gen struct {
	nextId   int
	mapping  map[Atom]int
	Filename string
	out      *os.File
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

func (l Literal) ToTex() (s string) {
	if !l.Sign {
		s += "\\bar "
	}
	s += "x_{"
	s += l.Atom.ToTex()
	s += "}"
	return
}

// Create Encoding for Sorting Network
// 0)  Omitted for clarity (ids as in paper)
// 1)  A or -D
// 2)  B or -D
// 3) -A or -B or D
// 4) -A or  C
// 5) -B or  C
// 6)  A or  B or -C
// 7)  C or -D
// -1,0,1 mean dontCare, false, true
func CreateEncoding(input []Literal, which [8]bool, output []Literal, tag string, pred Pred, sorter sorters.Sorter) (cs ClauseSet) {

	cs = make([]Clause, 0, 7*len(sorter.Comparators))

	backup := make(map[int]Literal, len(sorter.Out)+len(sorter.In))

	for i, x := range sorter.In {
		backup[x] = input[i]
	}

	for i, x := range sorter.Out {
		backup[x] = output[i]
	}

	for _, comp := range sorter.Comparators {

		if comp.D == 1 || comp.C == 0 {
			fmt.Println("something is wrong with the comparator", comp)
			panic("something is wrong with the comparator")
		}

		getLit := func(x int) Literal {
			if lit, ok := backup[x]; ok {
				return lit
			} else {
				return Literal{true, Atom{pred, x, 0}}
			}
		}

		a := getLit(comp.A)
		b := getLit(comp.B)
		c := getLit(comp.C)
		d := getLit(comp.D)

		if comp.C == 1 { // 6) A or B
			//if which[6] {
			cs.AddClause(tag+"6-", a, b)
			//}
		} else if comp.C > 0 { // 4) 5) 6)
			//4)
			if which[4] {
				cs.AddClause(tag+"4", Neg(a), c)
			}
			//5)
			if which[5] {
				cs.AddClause(tag+"5", Neg(b), c)
			}
			//6)
			if which[6] {
				cs.AddClause(tag+"6", a, b, Neg(c))
			}
		}
		if comp.D == 0 { //3)
			//if which[3] {
			cs.AddClause(tag+"3-", Neg(a), Neg(b))
			//}
		} else if comp.D > 0 { // 1) 2) 3)
			//1)
			if which[1] {
				cs.AddClause(tag+"1", a, Neg(d))
			}
			//2)
			if which[2] {
				cs.AddClause(tag+"2", b, Neg(d))
			}
			//3)
			if which[3] {
				cs.AddClause(tag+"3", Neg(a), Neg(b), d)
			}
		}

		if which[7] && (comp.D > 1 || comp.D > 1) { // 7)
			cs.AddClause(tag+"7", c, Neg(d))
		}
	}
	return
}

func IdGenerator(m int) (g Gen) {
	g.mapping = make(map[Atom]int, m)
	return
}

func (g *Gen) GenerateIds(cl ClauseSet) {
	for _, c := range cl {
		for _, l := range c.Literals {
			g.putAtom(l.Atom)
		}
	}
}

func (g *Gen) solve(cl []Clause) {
	for _, c := range cl {
		for _, l := range c.Literals {
			g.putAtom(l.Atom)
		}
	}

	g.PrintClausesDIMACS(cl)
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

func (g *Gen) Print(arg ...interface{}) {
	if g.Filename == "" {
		for _, s := range arg {
			fmt.Print(s, " ")
		}
	} else {
		var ss string
		for _, s := range arg {
			ss += fmt.Sprintf("%v", s) + " "
		}
		if _, err := g.out.Write([]byte(ss)); err != nil {
			panic(err)
		}
	}
}

func (g *Gen) Println(arg ...interface{}) {
	if g.Filename == "" {
		for _, s := range arg {
			fmt.Print(s, " ")
		}
		fmt.Println()
	} else {
		var ss string
		for _, s := range arg {
			ss += fmt.Sprintf("%v", s) + " "
		}
		ss += "\n"

		if _, err := g.out.Write([]byte(ss)); err != nil {
			panic(err)
		}
	}
}

func (g *Gen) PrintClausesDIMACS(clauses ClauseSet) {

	if g.Filename != "" {
		var err error
		g.out, err = os.Create(g.Filename)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := g.out.Close(); err != nil {
				panic(err)
			}
		}()
	}

	g.Println("p cnf", g.nextId, len(clauses))
    fmt.Println("CNF: #var", g.nextId,"#cls", len(clauses))

	for _, c := range clauses {
		for _, l := range c.Literals {
			s := strconv.Itoa(g.mapping[l.Atom])
			if l.Sign {
				g.Print(" " + s)
			} else {
				g.Print("-" + s)
			}
		}
		g.Println("0")
	}
	// close fo on exit and check for its returned error
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

func (g *Gen) printDebug(clauses []Clause) {

	symbolTable := g.generateSymbolTable()

	// first print symbol table into file
	fmt.Println("c <atom>(V1,V2).")

	for i, s := range symbolTable {
		fmt.Println("c", i, "\t:", s)
	}

	stat := make(map[string]int, 0)
	var descs []string

	for _, c := range clauses {

		count, ok := stat[c.Desc]
		if !ok {
			stat[c.Desc] = 1
			descs = append(descs, c.Desc)
		} else {
			stat[c.Desc] = count + 1
		}

		fmt.Printf("c %s\t", c.Desc)
		first := true
		for _, l := range c.Literals {
			if !first {
				fmt.Printf(",")
			}
			first = false
			if l.Sign {
				fmt.Print(" ")
			} else {
				fmt.Print("-")
			}
			fmt.Print(l.Atom.P, "(", l.Atom.V1, ",", l.Atom.V2, ")")
		}
		fmt.Println(".")
	}

	for _, key := range descs {
		fmt.Printf("c %v\t: %v\t%.1f \n", key, stat[key], 100*float64(stat[key])/float64(len(clauses)))
	}
	fmt.Printf("c %v\t: %v\t%.1f \n", "tot", len(clauses), 100.0)
}
