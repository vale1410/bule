package sat

// test class, but will eventuall be turned into the sat package :-)

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
    "../sorters"
)

// Create Encoding for Sorting Network
// 1)  A or -D
// 2)  B or -D
// 3) -A or -B or D
// 4) -A or  C
// 5) -B or  C
// 6)  A or  B or -C
// 7)  C or -D
// -1,0,1 mean dontCare, false, true
func createEncoding(input []Literal, which [8]bool, output []Literal, tag string, pred Pred, sorter sorters.Sorter) (cs ClauseSet) {

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
				cs.AddClause(tag+"4", neg(a), c)
			}
			//5)
			if which[5] {
				cs.AddClause(tag+"5", neg(b), c)
			}
			//6)
			if which[6] {
				cs.AddClause(tag+"6", a, b, neg(c))
			}
		}
		if comp.D == 0 { //3)
			//if which[3] {
			cs.AddClause(tag+"3-", neg(a), neg(b))
			//}
		} else if comp.D > 0 { // 1) 2) 3)
			//1)
			if which[1] {
				cs.AddClause(tag+"1", a, neg(d))
			}
			//2)
			if which[2] {
				cs.AddClause(tag+"2", b, neg(d))
			}
			//3)
			if which[3] {
				cs.AddClause(tag+"3", neg(a), neg(b), d)
			}
		}

		if which[7] && (comp.D > 1 || comp.D > 1) { // 7)
			cs.AddClause(tag+"7", c, neg(d))
		}
	}
	return
}

type Clause struct {
	desc     string
	literals []Literal
}

type ClauseSet []Clause

func (cs *ClauseSet) AddClause(desc string, literals ...Literal) {
	*cs = append(*cs, Clause{desc, literals})
}

func (cs *ClauseSet) AddClauseSet(cl ClauseSet) {
	*cs = append(*cs, cl...)
}

type Literal struct {
	sign bool
	atom Atom
}

func neg(l Literal) Literal {
	l.sign = !l.sign
	return l
}

type Pred string

// we only allow two dimensional predicates
type Atom struct {
	P  Pred
	V1 int
	V2 int
}

type Gen struct {
	nextId   int
	mapping  map[Atom]int
	filename string
	out      *os.File
}

func IdGenerator(m int) (g Gen) {
	g.mapping = make(map[Atom]int, m)
	return
}

func (g *Gen) GenerateIds(cl ClauseSet) {
	for _, c := range cl {
		for _, l := range c.literals {
			g.putAtom(l.atom)
		}
	}
}

func (g *Gen) solve(cl []Clause) {
	for _, c := range cl {
		for _, l := range c.literals {
			g.putAtom(l.atom)
		}
	}

	g.printClausesDIMACS(cl)
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
	if g.filename == "" {
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
	if g.filename == "" {
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

func (g *Gen) printClausesDIMACS(clauses ClauseSet) {

	if g.filename != "" {
		var err error
		g.out, err = os.Create(g.filename)
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

	for _, c := range clauses {
		for _, l := range c.literals {
			s := strconv.Itoa(g.mapping[l.atom])
			if l.sign {
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

		count, ok := stat[c.desc]
		if !ok {
			stat[c.desc] = 1
			descs = append(descs, c.desc)
		} else {
			stat[c.desc] = count + 1
		}

		fmt.Printf("c %s\t", c.desc)
		first := true
		for _, l := range c.literals {
			if !first {
				fmt.Printf(",")
			}
			first = false
			if l.sign {
				fmt.Print(" ")
			} else {
				fmt.Print("-")
			}
			fmt.Print(l.atom.P, "(", l.atom.V1, ",", l.atom.V2, ")")
		}
		fmt.Println(".")
	}

	for _, key := range descs {
		fmt.Printf("c %v\t: %v\t%.1f \n", key, stat[key], 100*float64(stat[key])/float64(len(clauses)))
	}
	fmt.Printf("c %v\t: %v\t%.1f \n", "tot", len(clauses), 100.0)
}
