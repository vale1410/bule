package sat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vale1410/bule/glob"
)

// from atom.Id() -> 0/1
type Assignment map[string]int

type Optimizer interface {
	Evaluate(Assignment) int64
	Translate(int64) ClauseSet
	String() string
	Empty() bool
}

type Result struct {
	Solved      bool
	Satisfiable bool
	Optimal     bool
	Timeout     bool
	M           string
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

func (g *Gen) refresh() {
	g.mapping = make(map[string]int)
	g.idMap = make([]Atom, 1)
	g.nextId = 0
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

func (g *Gen) Solve(cs ClauseSet, opt Optimizer, init int64) (result Result) {

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

	result.Value = math.MaxInt64

	if !opt.Empty() && init >= 0 {
		glob.D("init set", init)
		opt_clauses := opt.Translate(init)
		fmt.Println("opt cls", opt_clauses.Size())
		current.AddClauseSet(opt_clauses)
		result.Value = init + 1
	}

	result.Assignment = make(Assignment, len(g.idMap))

	time_total := time.Now()

	iterations := 0

	for !finished {
		iterations++

		//glob.D("Writing", current.Size(), "clauses")
		fmt.Println("tot cls", current.Size())
		//g.PrintDIMACS(current)
		//current.PrintDebug()

		if opt.Empty() {
			glob.D("solving...")
		} else {
			glob.D("solving for opt <= ", maxS(result.Value-1), "...")
			//glob.D(opt.String())
		}
		time_before := time.Now()
		go g.solveProblem(current, result_chan)

		select {
		case r := <-result_chan:
			result.Solved = r.solved
			fmt.Printf("Time: %.3f s\n", time.Since(time_before).Seconds())
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

					if !opt.Empty() {
						v := opt.Evaluate(result.Assignment)
						if v <= 0 {
							glob.D("SAT for opt = 0")
							finished = true
							result.Optimal = true
							fmt.Println("OPTIMIUM:  0")
							result.M = "OPTIMUM"
						} else {
							glob.A(v < result.Value, v, "<", result.Value, "no improvement ... cant be ")
							result.Value = v
							fmt.Println("SAT for opt =", result.Value)
							result.M = "SAT"
							current = cs
							//g.printAssignment(result.Assignment)
							//fmt.Println()
							opt_clauses := opt.Translate(result.Value - 1)
							fmt.Println("opt cls", opt_clauses.Size())
							//opt_clauses.PrintDebug()
							current.AddClauseSet(opt_clauses)
						}
					} else {
						fmt.Println("SAT")
						result.M = "SAT"
						finished = true
					}

				} else {
					finished = true
					result.Optimal = true
					if !opt.Empty() {
						glob.D("UNSAT at", maxS(result.Value-1), ", lower bound proven for ", maxS(result.Value))
						fmt.Println("OPTIMIUM: ", maxS(result.Value))
						result.M = "OPTIMUM"
					} else {
						fmt.Println("UNSAT")
						result.M = "UNSAT"
					}
				}
			} else {
				result.Solved = false
				glob.D("Result received not solved, why?")
				result.M = "ERROR"
				finished = true
			}
		case <-timeout:
			fmt.Println("TIMEOUT")
			result.M = "TIMEOUT"
			finished = true
			result.Solved = false
			result.Timeout = true
		}
	}

	close(result_chan)
	close(timeout)

	fmt.Printf("cTIME: %.3f s\n", time.Since(time_total).Seconds())
	fmt.Printf("xxx: %v;%v;%v;%.2f;%v;%v;%v\n", glob.Filename_flag, result.M, maxS(result.Value), time.Since(time_total).Seconds(), iterations, cs.Size(), current.Size()-cs.Size())

	return
}

func maxS(v int64) string {
	if v > math.MaxInt64/2 {
		return "+∞"
	} else {
		return strconv.Itoa(int(v))
	}
}

func (g *Gen) printAssignment(assignment Assignment) {

	count := 2
	fmt.Print("Vars: ")
	for idS := range g.PrimaryVars {
		if value, b := assignment[idS]; value == 1 && b {
			count++
			if count%10 == 0 {
				fmt.Println()
			} else if count == 19 {
				fmt.Println("\n... ")
				break
			}

			fmt.Print(idS)
			fmt.Print(":", value, " ")
		}
	}
	fmt.Println()
	count = 0

	first := true
	for idS, value := range assignment {
		if value == 1 && !g.PrimaryVars[idS] {
			if first {
				fmt.Print("Aux: ")
				count = 2
				first = false
			}
			count++
			if count%10 == 0 {
				fmt.Println()
			} else if count == 19 {
				fmt.Println("\n... ")
				break
			}

			fmt.Print(idS)
			fmt.Print(":", value, " ")
		}
	}
	fmt.Println()
}

type rawResult struct {
	solved      bool
	satisfiable bool
	assignment  string
}

func (g *Gen) solveProblem(clauses ClauseSet, result chan<- rawResult) {

	var solver *exec.Cmd

	//switch glob.Solver_flag {
	//case "minisat":
	//	//solver = exec.Command("minisat", g.Filename, "-rnd-seed=", strconv.FormatInt(glob.Seed_flag, 10))
	//	solver = exec.Command("minisat", g.Filename)
	//case "glucose":
	//	solver = exec.Command("glucose", g.Filename)
	//case "clasp":
	//	solver = exec.Command("clasp", g.Filename, "--time-limit", strconv.Itoa(glob.Timeout_flag))
	//case "lingeling":
	//	solver = exec.Command("lingeling", g.Filename)
	//case "cmsat":
	//	solver = exec.Command("cmsat", g.Filename)
	//case "local":
	//	solver = exec.Command("CCAnr", g.Filename, strconv.FormatInt(glob.Seed_flag, 10))
	//default:
	//	glob.A(false, "Solver not available", glob.Solver_flag)
	//}
	switch glob.Solver_flag {
	case "minisat":
		//solver = exec.Command("minisat", g.Filename, "-rnd-seed=", strconv.FormatInt(glob.Seed_flag, 10))
		solver = exec.Command("minisat")
	case "glucose":
		solver = exec.Command("glucose")
	case "clasp":
		solver = exec.Command("clasp", "--time-limit", strconv.Itoa(glob.Timeout_flag))
	case "lingeling":
		solver = exec.Command("lingeling")
	case "cmsat":
		solver = exec.Command("cmsat")
	case "local":
		solver = exec.Command("CCAnr", strconv.FormatInt(glob.Seed_flag, 10))
	default:
		glob.A(false, "Solver not available", glob.Solver_flag)
	}

	stdin, err := solver.StdinPipe()
	if err != nil {
		panic(err)
	}
	stdout, err := solver.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = solver.Start()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer stdin.Close()
		defer wg.Done()
		g.generateIds(clauses)
		io.Copy(stdin, bytes.NewReader([]byte(fmt.Sprintf("p cnf %v %v\n", g.nextId, len(clauses.list)))))
		for _, c := range clauses.list {
			io.Copy(stdin, bytes.NewReader(g.toBytes(c)))
		}
	}()

	var res rawResult
	go func() {
		defer wg.Done()
		r := bufio.NewReader(stdout)
		s, err := r.ReadString('\n')
		//fmt.Print(s)

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
			}
			s, err = r.ReadString('\n')
			//	fmt.Print(s)
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err.Error())
			}
		}
	}()

	wg.Wait()
	err_tmp := solver.Wait()
	//fmt.Println(res)
	if err_tmp != nil {
		fmt.Println("return value:", err_tmp.Error())
	}

	//if err = solver.Process.Kill(); err != nil {
	//	panic(err.Error())
	//}
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

	g.refresh()

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
		g.Print(string(g.toBytes(c)))
	}
}

func (g *Gen) toBytes(clause Clause) []byte {
	var buf bytes.Buffer
	for _, l := range clause.Literals {
		if l.Sign {
			buf.WriteString(" ")
		} else {
			buf.WriteString(" -")
		}
		buf.WriteString(strconv.Itoa(g.mapping[l.A.Id()]))
	}
	buf.WriteString(" 0\n")
	return buf.Bytes()
}

func (g *Gen) PrintMapping() {

	for i, s := range g.mapping {
		fmt.Println("c", i, "\t:", s)
	}

}
