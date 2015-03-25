package sat

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/vale1410/bule/glob"
)

// from atom.Id() -> 0/1
type Assignment map[string]int

type Optimizer interface {
	Evaluate(Assignment) int64
	Translate(int64) ClauseSet
	Empty() bool
}

type Result struct {
	Solved      bool
	Satisfiable bool
	Optimal     bool
	Timeout     bool
	Value       int64
	Time        int64 // total time to solution
	Assignment  Assignment
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

func (g *Gen) Solve(cs ClauseSet, opt Optimizer) (result Result) {

	glob.A(g.Filename != "", "Set filename for SAT solving.")
	glob.A(cs.Size() > 0, "Needs to contain at least 1 clause.")

	//generate the reverse mapping

	result_chan := make(chan rawResult)
	timeout := make(chan bool, 1)

	go func() {
		time.Sleep(time.Duration(glob.Timeout_flag) * time.Second)
		timeout <- true
	}()

	finished := false
	current := cs
	result.Assignment = make(Assignment, len(g.idMap))
	result.Value = math.MaxInt64

	for !finished {

		log.Println("Writing", current.Size(), "clauses")
		g.PrintDIMACS(cs)

		log.Println("solving...", result.Value)
		go g.solveProblem(result_chan)

		select {
		case r := <-result_chan:
			result.Solved = r.solved
			if r.solved {
				result.Satisfiable = r.satisfiable
				if r.satisfiable {
					ss := strings.Split(strings.TrimSpace(r.assignment), " ")

					count := 0
					for _, x := range ss {
						x = strings.TrimSpace(x)
						if x == "" {
							continue
						}
						id, err := strconv.Atoi(x)
						if err != nil {
							glob.A(false, err.Error())
						}
						if id != 0 {
							sign := 1
							if id < 0 {
								sign = 0
								id = -id
							}

							atom := g.idMap[id]
							if g.PrimaryVars[atom.Id()] {
								count++
								result.Assignment[atom.Id()] = sign
							}

						}
					}

					glob.A(count == len(result.Assignment), "count != assignment")

					v := opt.Evaluate(result.Assignment)
					fmt.Println(v, result.Value)
					glob.A(v < result.Value, v, "<", result.Value, "no improvement ... cant be ")
					result.Value = v

					if !opt.Empty() {
						current = cs
						//current.AddClauseSet(opt.Translate(result.Value - 1))
						//current.PrintDebug()
						fmt.Println("result.Value", result.Value)

						//current.AddClauseSet(opt.Translate(result.Value - 2))
						//current.PrintDebug()
						finished = true
					} else {
						finished = true
					}

				} else {
					finished = true
					result.Optimal = true
					log.Println("lower bound proven")
				}
			}
		case <-timeout:
			finished = true
			result.Solved = false
			result.Timeout = true
		}
	}

	close(result_chan)
	close(timeout)

	return
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

type rawResult struct {
	solved      bool
	satisfiable bool
	assignment  string
}

func (g *Gen) solveProblem(result chan<- rawResult) {

	solver := exec.Command("clasp", g.Filename, "--time-limit", strconv.Itoa(glob.Timeout_flag))
	//solver := exec.Command("clasp", g.Filename)
	//solver := exec.Command("cmsat", g.Filename)
	//solver := exec.Command("minisat", g.Filename)
	stdout, _ := solver.StdoutPipe()

	solver.Start()

	r := bufio.NewReader(stdout)
	s, err := r.ReadString('\n')
	var res rawResult

	for {
		if strings.HasPrefix(s, "v ") {
			res.assignment += s[1:]
		} else if strings.HasPrefix(s, "s ") {
			if strings.Contains(s, "UNSATISFIABLE") {
				res.solved = true
				res.satisfiable = false
			} else if strings.Contains(s, "SATISFIABLE") {
				res.solved = true
				res.satisfiable = true
			} else {
				res.solved = false
				glob.D("whats up? result of sat solver does not contain proper answer!")
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
