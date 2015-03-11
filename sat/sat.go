package sat

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Gen struct {
	nextId   int
	mapping  map[string]int
	Filename string
	out      *os.File
}

func IdGenerator(m int) (g Gen) {
	g.mapping = make(map[string]int, m)
	return
}

//#########################################################3

func (g *Gen) Solve(cs ClauseSet) {

	// check a filename

	if g.Filename == "" { // in the future dont
		panic("generator has not filled filenmane, is needed for SAT solving")
	}

	g.PrintDIMACS(cs)

	//generate the reverse mapping

	result := make(chan Result)
	timeout := make(chan bool, 1)
	ttimeout := 10 //timeout in seconds

	go func() {
		time.Sleep(time.Duration(ttimeout) * time.Second)
		timeout <- true
	}()

	go g.solveProblem(result)

	select {
	case r := <-result:

		if r.satisfiable {
			//parseResult(r.s, assignment)
			fmt.Println("SATISFIABLE", r.s)
		} else {
			fmt.Println("UNSATISFIABLE")
		}
	case <-timeout:
		fmt.Println("what are you waiting for? timeout")
	}

	close(result)
	close(timeout)

	//print output from mapping

	//fmt.Printf("%v %v\n", current, optimal)
	//for _, x := range assignment {
	//	fmt.Printf("%v ", x)
	//}
	//fmt.Printf("\n")
}

//func parseResult(s string, assignment []bool) bool {
//	ss := strings.Split(string(s), " ")
//
//	ok := len(assignment) == len(ss)
//
//	if ok {
//		for _, x := range ss {
//
//			if strings.HasPrefix(x, "assign") {
//				numbers := digitRegexp.FindAllString(x, -1)
//
//				if 2 == len(numbers) {
//
//					customer, b1 := strconv.Atoi(numbers[0])
//					warehouse, b2 := strconv.Atoi(numbers[1])
//					if b1 != nil || b2 != nil {
//						panic("bad conversion of numbers in result")
//					}
//					assignment[customer] = warehouse
//				} else {
//					ok = false
//					break
//				}
//			}
//		}
//	}
//
//	return ok
//}

func (g *Gen) solveProblem(result chan<- Result) {

	//	gringo := exec.Command("gringo", *out, *model)
	clasp := exec.Command("clasp", g.Filename)
	//clasp.Stdin, _ = gringo.StdoutPipe()

	satFilter := NewSATFilter(result)
	clasp.Stdout = &satFilter

	_ = clasp.Run()

}

type Result struct {
	satisfiable bool
	s           string
}

type SATFilter struct {
	result chan<- Result
	backup string // keep string of values
}

func NewSATFilter(result chan<- Result) (cf SATFilter) {
	cf.result = result
	return
}

func (cf *SATFilter) Write(p []byte) (n int, err error) {

	s := string(p)

	fmt.Println("out:", string(p))

	if strings.HasPrefix(s, "v ") {
		cf.backup += s[2:]
	} else if strings.HasPrefix(s, "s ") {
		result := Result{true, cf.backup}
		if strings.Contains(s, "UNSATISFIABLE") {
			result.satisfiable = false
		} else if strings.Contains(s, "SATISFIABLE") {
			result.satisfiable = true
		} else {
			fmt.Println(s)
			panic("whats up? result of sat solver does not contain proper anwser!")
		}
		cf.result <- result
	}

	return len(p), nil
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
