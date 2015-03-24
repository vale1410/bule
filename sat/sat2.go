package sat

import (
	"bufio"
	"fmt"
	"github.com/vale1410/bule/glob"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Assignment map[string]int

type Optimizer interface {
	Evaluate(Assignment) (int64, error)
	Translate(int64) ClauseSet
}

type Gen struct {
	nextId      int
	mapping     map[string]int
	idMap       []Atom
	PrimaryVars map[string]bool
	Filename    string
	out         *os.File
}

func IdGenerator(m int) (g Gen) {
	g.mapping = make(map[string]int, m)
	g.idMap = make([]Atom, 1, m)
	return
}

func (g *Gen) putAtom(a Atom) {
	if _, b := g.mapping[a.Id()]; !b {
		g.nextId++
		id := g.nextId
		g.mapping[a.Id()] = id
		g.idMap = append(g.idMap, a)
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

func (g *Gen) PrintSymbolTable(filename string) {

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

func (g *Gen) Solve(cs ClauseSet, opt Optimizer) {

	// check a filename

	glob.A(g.Filename != "", "No filename, is needed for SAT solving.")
	glob.A(cs.Size() > 0, "Needs to contain at least 1 clause.")
	fmt.Print(";", cs.Size())

	g.PrintDIMACS(cs)

	//generate the reverse mapping

	result := make(chan Result)
	timeout := make(chan bool, 1)
	ttimeout := glob.Timeout_flag //timeout in seconds

	go func() {
		time.Sleep(time.Duration(ttimeout) * time.Second)
		timeout <- true
	}()

	go g.solveProblem(result)

	assignment := make([]bool, len(g.idMap))

	select {
	case r := <-result:

		if r.satisfiable {
			fmt.Print(";SATISFIABLE")
			ss := strings.Split(r.s, " ")

			for _, x := range ss {
				id, _ := strconv.Atoi(x)
				if id != 0 {
					if id < 0 {
						assignment[-id] = false
					} else {
						assignment[id] = true
					}
				}
			}
		} else {
			fmt.Print(";UNSATISFIABLE")
		}
	case <-timeout:
		fmt.Print(";TIMEOUT")
	}
	//fmt.Println()

	close(result)
	close(timeout)

	//print output from mapping

}

func (g *Gen) evaluateAssignment(assignment []bool, opt Optimizer) {

}

func (g *Gen) printAssignment(assignment []bool) {

	count := -1
	fmt.Println("Primary Variables:")
	for i, x := range assignment {
		if i > 0 && g.PrimaryVars[g.idMap[i].Id()] {
			count++
			if count%10 == 0 {
				fmt.Println()
			} else if count == 19 {
				fmt.Println("\n... ")
				break
			}

			if x {
				fmt.Print(" ")
			} else {
				fmt.Print(" -")
			}
			fmt.Print(g.idMap[i].Id())
		}
	}
	fmt.Println()
	count = -1

	first := true
	for i, x := range assignment {
		if i > 0 && !g.PrimaryVars[g.idMap[i].Id()] {
			if first {
				fmt.Println("Auxiliary Variables:")
				first = false
			}
			count++
			if count%10 == 0 {
				fmt.Println()
			} else if count == 19 {
				fmt.Println("\n... ")
				break
			}
			if x {
				fmt.Print(" ")
			} else {
				fmt.Print(" -")
			}
			fmt.Print(g.idMap[i].Id())
		}
	}
	fmt.Println()
}

type Result struct {
	satisfiable bool
	s           string
}

func (g *Gen) solveProblem(result chan<- Result) {

	solver := exec.Command("clasp", g.Filename, "--time-limit", strconv.Itoa(glob.Timeout_flag))
	//solver := exec.Command("clasp", g.Filename)
	//solver := exec.Command("cmsat", g.Filename)
	//solver := exec.Command("minisat", g.Filename)
	stdout, _ := solver.StdoutPipe()

	solver.Start()

	r := bufio.NewReader(stdout)
	s, err := r.ReadString('\n')
	var res Result

	assignment := ""

	for {
		if strings.HasPrefix(s, "v ") {
			assignment += s[1:]
		} else if strings.HasPrefix(s, "s ") {
			res = Result{true, assignment}
			if strings.Contains(s, "UNSATISFIABLE") {
				res.satisfiable = false
			} else if strings.Contains(s, "SATISFIABLE") {
				res.satisfiable = true
			} else {
				fmt.Println(s)
				panic("whats up? result of sat solver does not contain proper answer!")
			}
			break
		}
		s, err = r.ReadString('\n')
		if err != nil {
			panic(err.Error())
		}
	}

	if err = solver.Process.Kill(); err != nil {
		panic(err.Error())
	}
	result <- res
}

//#########################################################3

func (g *Gen) Print(arg ...interface{}) {
	if g.Filename == "" {
		for _, s := range arg {
			fmt.Print(s, " ")
		}
	} else { //assuming the file is open!
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
	} else { //assuming the file is open!
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

func (g *Gen) generateIds(cs ClauseSet) {
	// recalculates new sat ids for each atom:
	// assuming full regeneration of Ids
	// might change existing mappings

	g.nextId = 0

	for _, c := range cs.list {
		for _, l := range c.Literals {
			g.putAtom(l.A)
		}
	}
}

func (g *Gen) PrintDIMACS(cs ClauseSet) {

	g.generateIds(cs)

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

	g.Println("p cnf", g.nextId, len(cs.list))

	for _, c := range cs.list {
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
	// close on exit and check for its returned error
}

func (g *Gen) PrintMapping() {

	for i, s := range g.mapping {
		fmt.Println("c", i, "\t:", s)
	}

}
