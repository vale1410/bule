package sat

// todo: Call Solvers, get back result etc.

import (
	"fmt"
	"os"
	"strconv"
)

func (g *Gen) GenerateIds(cl ClauseSet) {
	for _, c := range cl {
		for _, l := range c.Literals {
			g.putAtom(l.A)
		}
	}
}

func (g *Gen) solve(cl []Clause) {
	for _, c := range cl {
		for _, l := range c.Literals {
			g.putAtom(l.A)
		}
	}

	g.PrintClausesDIMACS(cl)
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
	//fmt.Println("CNF: #var", g.nextId, "#cls", len(clauses))

	for _, c := range clauses {
		for _, l := range c.Literals {
			s := strconv.Itoa(g.mapping[l.A.Id()])
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


func (g *Gen) PrintDebug(clauses []Clause) {

	// first print symbol table into file
	fmt.Println("c <atom>(V1,V2).")

	for i, s := range g.mapping {
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
			fmt.Print(l.ToTxt())
		}
		fmt.Println(".")
	}

	for _, key := range descs {
		fmt.Printf("c %v\t: %v\t%.1f \n", key, stat[key], 100*float64(stat[key])/float64(len(clauses)))
	}
	fmt.Printf("c %v\t: %v\t%.1f \n", "tot", len(clauses), 100.0)
}
