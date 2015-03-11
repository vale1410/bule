package main

import (
	"flag"
	"fmt"
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/translation"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var f = flag.String("f", "test.pb", "Path of to PB file.")
var out = flag.String("o", "out.cnf", "Path of output file.")
var ver = flag.Bool("ver", false, "Show version info.")

//var check_clause = flag.Bool("clause", true, "Checks if Pseudo-Boolean is not just a simple clause.")
var dbg = flag.Bool("d", false, "Print debug information.")
var dbgfile = flag.String("df", "", "File to print debug information.")
var reformat = flag.Bool("reformat", false, "Reformat PB files into correct format. Decompose = into >= and <=")
var gurobi = flag.Bool("gurobi", false, "Reformat to Gurobi input, output to stdout.")

//var model = flag.String("model", "model.lp", "path to model file")
//var solve = flag.Bool("solve", false, "Pass problem to clasp and solve.")
//var ttimeout = flag.Int("timeout", 10, "Timeout in seconds.")

var digitRegexp = regexp.MustCompile("([0-9]+ )*[0-9]+")

var dbgoutput *os.File

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
		fmt.Println(`Bule CNF Grounder: Tag 0.1 Pseudo Booleans
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
		return
	}

	pbs := parse(*f)

	if *gurobi {
		fmt.Println("Subject To")
		atoms := make(map[string]bool, len(pbs))
		for _, pb := range pbs {
			pb.Normalize(constraints.AtLeast, false)
			pb.PrintGurobi()
			for _, x := range pb.Entries {
				atoms[x.Literal.A.Id()] = true
			}
		}
		fmt.Println("Binary")
		for aS, _ := range atoms {
			fmt.Print("x" + aS + " ")
		}
		fmt.Println()
	} else if *reformat {
		for _, pb := range pbs {
			if pb.Typ == constraints.Equal {
				pb.Typ = constraints.AtLeast
				pb.Normalize(constraints.AtLeast, false)
				pb.Print10()
				pb.Typ = constraints.AtMost
			}
			pb.Normalize(constraints.AtLeast, false)
			pb.Print10()
		}
	} else {

		var clauses sat.ClauseSet

		stats := make([]int, translation.TranslationTypes)

		for i, pb := range pbs {

			pb.Id = i
			//pb.Print10()

			t := translation.Categorize(&pb)

			stats[t.Typ]++

			clauses.AddClauseSet(t.Clauses)
			//t.Clauses.PrintDebug()
			//pb.Print10()
			//fmt.Println()
		}

		printStats(stats)
		g := sat.IdGenerator(clauses.Size() * 7)
		g.Filename = *out
		//clauses.PrintDebug()
		//g.PrintDIMACS(clauses)
		g.Solve(clauses)
	}

}

func printStats(stats []int) {
	if len(stats) != int(translation.TranslationTypes) {
		panic("Stats for translation errornous")
	}
	fmt.Printf("Facts\tClause\tAMO\tEx1\tCard\tCompl\n")

	for _, x := range stats {
		fmt.Printf("%v\t", x)
	}
	fmt.Println()
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

func parse(filename string) (pbs []constraints.Threshold) {

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
	var count int
	state := 0
	t := 0

	for _, l := range lines {

		if state > 0 && (l == "" || strings.HasPrefix(l, "%") || strings.HasPrefix(l, "*")) {
			continue
		}

		elements := strings.Fields(l)

		switch state {
		case 0:
			{
				debug(l)
				var b1 error
				count, b1 = strconv.Atoi(elements[4])
				vars, b2 := strconv.Atoi(elements[2])
				if b1 != nil || b2 != nil {
					debug("cant convert to threshold:", l)
					panic("bad conversion of numbers")
				}
				debug("Found PB file with", count, "constraints and", vars, "variables")
				pbs = make([]constraints.Threshold, count)
				state = 1
			}
		case 1:
			{
				if t >= count {
					panic("Number of constraints incorrectly specified in pb input file " + filename)
				}
				pbs[t].Desc = l

				n := (len(elements) - 3) / 2
				pbs[t].Entries = make([]constraints.Entry, n)

				for i := 0; i < len(elements)-3; i++ {

					weight, b1 := strconv.ParseInt(elements[i], 10, 64)
					i++
					variable, b2 := strconv.Atoi(digitRegexp.FindString(elements[i]))

					if b1 != nil || b2 != nil {
						debug("cant convert to threshold:", l)
						panic("bad conversion of numbers")
					}
					atom := sat.NewAtomP1(sat.Pred("x"), variable)
					pbs[t].Entries[i/2] = constraints.Entry{sat.Literal{true, atom}, weight}
				}

				pbs[t].K, _ = strconv.ParseInt(elements[len(elements)-2], 10, 64)
				typS := elements[len(elements)-3]

				if typS == ">=" {
					pbs[t].Typ = constraints.AtLeast
				} else if typS == "<=" {
					pbs[t].Typ = constraints.AtMost
				} else if typS == "==" || typS == "=" {
					pbs[t].Typ = constraints.Equal
				} else {
					debug("cant convert to threshold:", l)
					panic("bad conversion of symbols" + typS)
				}
				t++
			}
		}
	}
	return
}
