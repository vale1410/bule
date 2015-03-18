package sat

import (
	"bufio"
	"fmt"
	"github.com/vale1410/bule/config"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

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
		// id check code:
		//		if g.mapping[g.idMap[id].Id()] != id {
		//			panic("wrong id mapping stuff")
		//		}
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

func (g *Gen) Solve(cs ClauseSet) {

	// check a filename

	if g.Filename == "" { // in the future dont
		panic("generator has not filled filenmane, is needed for SAT solving")
	}

	g.PrintDIMACS(cs)

	//generate the reverse mapping

	result := make(chan Result)
	timeout := make(chan bool, 1)
	ttimeout := config.Timeout_flag //timeout in seconds

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
					//fmt.Println(id)
					if id < 0 {
						assignment[-id] = false
					} else {
						assignment[id] = true
					}
				}
			}
			//g.printAssignment(assignment)
		} else {
			fmt.Print(";UNSATISFIABLE")
		}
	case <-timeout:
		fmt.Print(";TIMEOUT")
	}

	close(result)
	close(timeout)

	//print output from mapping

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

	clasp := exec.Command("clasp", g.Filename)
	stdout, _ := clasp.StdoutPipe()

	clasp.Start()

	r := bufio.NewReader(stdout)
	line, isPrefix, err := r.ReadLine()

	assignment := ""

	for err == nil && !isPrefix {
		s := string(line)
		if strings.HasPrefix(s, "v ") {
			assignment += s[1:]
		} else if strings.HasPrefix(s, "s ") {
			r := Result{true, assignment}
			if strings.Contains(s, "UNSATISFIABLE") {
				r.satisfiable = false
			} else if strings.Contains(s, "SATISFIABLE") {
				r.satisfiable = true
			} else {
				fmt.Println(s)
				panic("whats up? result of sat solver does not contain proper answer!")
			}
			result <- r
		}
		line, isPrefix, err = r.ReadLine()
	}
	if isPrefix {
		fmt.Println("buffer size to small")
		return
	}
	if err != io.EOF {
		fmt.Println(err)
		return
	}
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
