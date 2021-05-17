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
	Satisfiable bool // is satisfiable (opt: there exists at least one solution)
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
}

func IdGenerator(m int) (g Gen) {
	g.mapping = make(map[string]int, m)
	g.idMap = make([]Atom, 1, m)
	return
}

func (g *Gen) refresh() {
	g.mapping = make(map[string]int)
	g.idMap = make([]Atom, glob.First_aux_id_flag)
	g.nextId = glob.First_aux_id_flag - 1
}

func (g *Gen) putAtom(a Atom) {
	if _, b := g.mapping[a.Id()]; !b {
		succ := false
		if glob.Infer_var_ids {
			id, err := strconv.Atoi(strings.TrimLeft(a.Id(), "v"))
			v := strings.TrimRight(a.Id(), "0123456789")
			if v == "v" && err == nil {
				glob.A(glob.First_aux_id_flag > id, "Inferred number ID if higher than First_aux_id. Use values for first_aux  that a larger than id in all variables v<id>.")
				succ = true
				g.mapping[a.Id()] = id
				g.idMap[id] = a
			}
		}
		if !succ {
			g.nextId++
			g.mapping[a.Id()] = g.nextId
			g.idMap = append(g.idMap, a)
		}

	}
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

func (g *Gen) Solve(cs ClauseSet, opt Optimizer, nextOpt int64, lb int64) (result Result) {

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

	if !opt.Empty() && nextOpt != math.MaxInt64 {
		glob.D("init", nextOpt, "lb", lb)
		opt_clauses := opt.Translate(nextOpt)
		fmt.Println("opt cls", opt_clauses.Size())
		current.AddClauseSet(opt_clauses)
	}
	result.Value = math.MaxInt64

	result.Assignment = make(Assignment, len(g.idMap))

	time_total := time.Now()

	iterations := 0

	for !finished {
		iterations++

		//glob.D("Writing", current.Size(), "clauses")
		fmt.Println("tot cls", current.Size())
		//current.PrintDebug()

		if opt.Empty() {
			glob.D("solving...")
		} else {
			fmt.Printf("i: %v\tcur: %v\t lb: %v\tbest: %v\n", iterations, maxS(nextOpt), lb, maxS(result.Value))
		}
		time_before := time.Now()

		if glob.Cnf_tmp_flag != "" {
			g.PrintDIMACS(current, false)
		}
		go g.solveProblem(current, result_chan)

		select {
		case r := <-result_chan:
			result.Solved = r.solved
			fmt.Printf("Time :\t%.3f s\n", time.Since(time_before).Seconds())
			if r.solved {
				if r.satisfiable {
					result.Satisfiable = true
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
						result.Value = opt.Evaluate(result.Assignment)
						g.printAssignment(result.Assignment)
						glob.D("SAT for value =", result.Value)
						finished, nextOpt = nextOptValue(lb, &result)

						if !finished {
							current = cs
							opt_clauses := opt.Translate(nextOpt)
							fmt.Println("opt cls", opt_clauses.Size())
							current.AddClauseSet(opt_clauses)
						} else {
							fmt.Println("OPTIMIUM", result.Value)
						}

					} else {
						fmt.Println("SAT")
						result.M = "SAT"
						finished = true
					}

				} else { //UNSAT
					if !opt.Empty() {
						// update lower bound
						glob.D("UNSAT for opt <=", maxS(nextOpt))

						if nextOpt == math.MaxInt64 {
							result.M = "UNSAT"
							finished = true
						} else {
							lb = nextOpt + 1

							finished, nextOpt = nextOptValue(lb, &result)

							if !finished {
								current = cs
								opt_clauses := opt.Translate(nextOpt)
								fmt.Println("opt cls", opt_clauses.Size())
								current.AddClauseSet(opt_clauses)
							} else {
								fmt.Println("OPTIMUM", result.Value)
							}
						}
					} else {
						finished = true
						result.Optimal = true
						result.M = "UNSAT"
					}
				}
			} else {
				result.Solved = false
				glob.D("Error received nothing solved, check log of solver?")
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

	//	fmt.Printf("cTIME: %.3f s\n", time.Since(time_total).Seconds())
	//	fmt.Printf("%v;%v;%v;%v;%v;%v;%v;%.2f;%v;%v;%v\n", "name", "seed", "Amo_chain", "Amo_reuse", "Rewrite_same", "result.M", "maxS(result.Value)", "time ins", "iterations", "cs.Size()", "current.Size()-cs.Size()")
	//	fmt.Printf("%v;%v;%v;%v;%v;%v;%v;%.2f;%v;%v;%v\n", glob.Filename_flag, glob.Seed_flag, glob.Amo_chain_flag, glob.Amo_reuse_flag, glob.Rewrite_same_flag, result.M, maxS(result.Value), time.Since(time_total).Seconds(), iterations, cs.Size(), current.Size()-cs.Size())
	fmt.Printf("%v;%v;%v;%.2f\n", glob.Filename_flag, result.M, maxS(result.Value), time.Since(time_total).Seconds())

	return
}

func nextOptValue(lb int64, result *Result) (finished bool, nextOpt int64) {
	switch glob.Search_strategy_flag {
	case "iterative":
		finished, nextOpt = nextOptIterative(lb, result)
	case "binary":
		finished, nextOpt = nextOptBinary(lb, result)
	default:
		glob.A(false, "Search strategy not implemented", glob.Search_strategy_flag)
	}
	return
}

func nextOptBinary(lb int64, result *Result) (bool, int64) {

	if lb == result.Value {
		result.M = "OPTIMUM"
		return true, result.Value
	} else if lb < result.Value {
		return false, (lb + result.Value) / 2
	} else {
		glob.A(false, "lb <= ub")
		return false, 0
	}
}

func nextOptIterative(lb int64, result *Result) (bool, int64) {
	if lb == result.Value {
		result.M = "OPTIMUM"
		return true, result.Value
	} else if lb < result.Value {
		return false, result.Value - 1
	} else {
		glob.A(false, "lb <= ub")
		return false, 0
	}
}

func maxS(v int64) string {
	if v > math.MaxInt64/2 {
		//return "+∞"
		return "?"
	} else {
		return strconv.Itoa(int(v))
	}
}

func (g *Gen) printAssignment(assignment Assignment) {

	count := 2
	fmt.Print("Solution: ")
	//	fmt.Println(assignment)
	//	fmt.Println(g.PrimaryVars)
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

// TODO : set seed, unite with solve.go
// @Sebastian: Second way to call SAT solvers
func (g *Gen) solveProblem(clauses ClauseSet, result chan<- rawResult) {

	var solver *exec.Cmd

	switch glob.Solver_flag {
	case "minisat":
		//solver = exec.Command("minisat", "-rnd-seed=123")
		//solver = exec.Command("minisat", "-rnd-seed="+strconv.FormatInt(glob.Seed_flag, 10))
		solver = exec.Command("minisat")
	case "glucose":
		solver = exec.Command("glucose", "-model")
	case "clasp":
		solver = exec.Command("clasp")
	case "lingeling":
		solver = exec.Command("lingeling")
	case "treengeling":
		solver = exec.Command("treengeling")
	case "plingeling":
		solver = exec.Command("plingeling")
	case "dimetheus":
		solver = exec.Command("dimetheus", "-seed="+strconv.FormatInt(glob.Seed_flag, 10))
	case "cmsat":
		solver = exec.Command("cmsat")
	case "local":
		solver = exec.Command("CCAnr", strconv.FormatInt(glob.Seed_flag, 10))
	case "microsat":
		solver = exec.Command("microsat")
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
		time_before := time.Now()
		defer stdin.Close()
		defer wg.Done()
		g.generateIds(clauses, false)
		io.Copy(stdin, bytes.NewReader([]byte(fmt.Sprintf("p cnf %v %v\n", g.nextId, len(clauses.list)))))
		for _, c := range clauses.list {
			io.Copy(stdin, bytes.NewReader(g.toBytes(c)))
		}
		fmt.Printf("Read :\t%.3f s\n", time.Since(time_before).Seconds())
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
					glob.D("whats up? Result of sat solver does not contain proper answer!")
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

	if err_tmp != nil {
		//glob.D("return value:",err_tmp() )
	}

	// TODO: why is this uncommented?
	//if err = solver.Process.Kill(); err != nil {
	//	panic(err.Error())
	//}

	result <- res
}

func (g *Gen) generateIds(cs ClauseSet, inferPrimeVars bool) { // recalculates new sat ids for each atom:
	// assuming full regeneration of Ids
	// might change existing mappings

	g.refresh()

	glob.D("c auxiliary Ids start with", g.nextId)

	for _, c := range cs.list {
		for _, l := range c.Literals {
			g.putAtom(l.A)
		}
	}
}

func (g *Gen) PrintDIMACS(cs ClauseSet, inferPrimeVars bool) {

	g.generateIds(cs, inferPrimeVars)

	var out io.Writer

	if glob.Cnf_tmp_flag != "" {
		file, err := os.Create(glob.Cnf_tmp_flag)
		if err != nil {
			panic(err)
		}
		out = file
		defer func() {
			if err := file.Close(); err != nil {
				panic(err)
			}
		}()
	} else {
		out = os.Stdout
	}

	if !glob.Infer_var_ids {
		fmt.Fprintf(out, "p cnf %d %d\n", g.nextId, len(cs.list))
	}

	for _, c := range cs.list {
		if _, err := out.Write(g.toBytes(c)); err != nil {
			panic(err)
		}

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
