package main

import (
	"flag"
	"fmt"
	"github.com/vale1410/bule/sat"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var f = flag.String("f", "qwh-5-10.pls", "Instance.")
var out = flag.String("o", "out.cnf", "Path of output file.")
var ver = flag.Bool("ver", false, "Show version info.")
var dbg = flag.Bool("d", false, "Print debug information.")
var dbgfile = flag.String("df", "", "File to print debug information.")

var digitRegexp = regexp.MustCompile("([0-9]+ )*[0-9]+")

var dbgoutput *os.File

type QGC [][]int

func main() {
	flag.Parse()

	if *dbgfile != "" {
		var err error
		dbgoutput, err = os.Create(*dbgfile)
		if err != nil {
			panic(err)
		}
		defer dbgoutput.Close()
	}

	debug("Running Debug Mode...")

	if *ver {
		fmt.Println(`QGC-Translator: Tag 0.1
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
		return
	}

	g := parse(*f)

	clauses := translateToSAT(g)

	s := sat.IdGenerator(len(clauses))
	s.GenerateIds(clauses)
	s.Filename = *out
	s.PrintClausesDIMACS(clauses)
}

func debug(arg ...interface{}) {
	if *dbg {
		if *dbgfile == "" {
			fmt.Print("dbg: ")
			for _, s := range arg {
				fmt.Print(s, " ")
			}
			fmt.Println()
		} else {
			ss := "dbg: "
			for _, s := range arg {
				ss += fmt.Sprintf("%v", s) + " "
			}
			ss += "\n"

			if _, err := dbgoutput.Write([]byte(ss)); err != nil {
				panic(err)
			}
		}
	}
}

func translateToSAT(g QGC) (clauses sat.ClauseSet) {

	fmt.Println(g)

	n := len(g)

	var litIn []sat.Literal
	p := sat.Pred("v")

	for i = 0; i < n; i++ {
		for j = 0; j < n; j++ {
			litIn = make([]sat.Literal, n)
			//s := "at least one"
			for k = 0; k < n; k++ {
				litIn[i][j] = sat.Literal{true, sat.Atom{i, j, k}}
			}
			clauses.AddClause(s, litIn...)
		}
	}

	//p := sat.Pred("vc")

	////at least constraint for each edge

	//s := "at least one"
	//for _, e := range vc.Edges {
	//	l1 := sat.Literal{true, sat.Atom{p, e.a, 0}}
	//	l2 := sat.Literal{true, sat.Atom{p, e.b, 0}}
	//	clauses.AddClause(s, l1, l2)
	//}

	////global counter

	//sorter := sorters.CreateCardinalityNetwork(vc.NVertex, vc.Size, sorters.AtMost, sorters.Pairwise)
	//sorter.RemoveOutput()

	//litIn := make([]sat.Literal, vc.NVertex)

	//for i, _ := range litIn {
	//	litIn[i] = sat.Literal{true, sat.Atom{p, i + 1, 0}}
	//}

	//which := [8]bool{false, false, false, true, true, true, false, false}
	//pred := sat.Pred("aux")
	//clauses.AddClauseSet(sat.CreateEncoding(litIn, which, []sat.Literal{}, "atMost", pred, sorter))

	return
}

func parse(filename string) (g QGC) {

	input, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Please specifiy correct path to instance. File does not exist: ", filename)
		panic(err)
	}

	output, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	lines := strings.Split(string(input), "\n")

	// 0 : first line, 1 : rest of the lines
	state := 0
	t := 0

	for ln, l := range lines {

		if state > 0 && (l == "" || strings.HasPrefix(l, "*")) {
			continue
		}

		elements := strings.Fields(l)

		switch state {
		case 0:
			{
				debug(l)

				n, b := strconv.Atoi(elements[1])

				if b != nil || elements[0] != "order" {
					debug("no proper stuff to read:", l)
					panic("bad conversion of numbers")
				}

				debug("Size of problem", n)

				g = make(QGC, n)
				for i, _ := range g {
					g[i] = make([]int, n)
				}
				state = 1
			}
		case 1:
			{
				if t > len(g) {
					debug(t, " ", l)
					panic("incorrect number of elements.")
				}

				for i, p := range elements {

					a, b := strconv.Atoi(p)

					if b != nil {
						debug("cant convert to instance:", l)
						panic("bad conversion of numbers")
					}

					g[ln-1][i] = a
				}

			}
		}
	}
	return
}
