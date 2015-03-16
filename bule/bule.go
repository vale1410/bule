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

var filename_flag = flag.String("f", "test.pb", "Path of to PB file.")
var out = flag.String("o", "out.cnf", "Path of output file.")
var ver = flag.Bool("ver", false, "Show version info.")

//var check_clause = flag.Bool("clause", true, "Checks if Pseudo-Boolean is not just a simple clause.")
var dbg = flag.Bool("d", false, "Print debug information.")
var dbgfile = flag.String("df", "", "File to print debug information.")
var reformat_flag = flag.Bool("reformat", false, "Reformat PB files into correct format. Decompose = into >= and <=")
var gurobi_flag = flag.Bool("gurobi", false, "Reformat to Gurobi input, output to stdout.")
var solve_flag = flag.Bool("solve", true, "Dont solve just categorize and analyze the constriants.")

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
		fmt.Println(`Bule CNF Grounder: Tag 0.4 Pseudo Booleans
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
		return
	}

	pbs, err := parse(*filename_flag)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if *gurobi_flag {
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
	} else if *reformat_flag {
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

		//if len(pbs) < 10 {
		//	fmt.Println()
		//	for _, pb := range pbs {
		//		pb.Print10()
		//	}
		//	fmt.Println()
		//}

		primaryVars := make(map[string]bool, 0)

		for i, pb := range pbs {

			pb.Id = i
			//pb.Print10()

			for _, x := range pb.Entries {
				primaryVars[x.Literal.A.Id()] = true
			}

			t := translation.Categorize(&pb)

			stats[t.Typ]++

			clauses.AddClauseSet(t.Clauses)
			//t.Clauses.PrintDebug()
			//pb.Print10()
			//fmt.Println()
		}

		if !*solve_flag {
			fmt.Print(*filename_flag, ";", len(primaryVars), ";", len(pbs), ";")
			for i, x := range stats {
				if i > 0 {
					fmt.Printf("%v;", x)
				}
			}
			fmt.Println()
		} else {
			printStats(stats)
			fmt.Println("\nsolve\n")
			g := sat.IdGenerator(clauses.Size() * 7)
			g.Filename = *out
			g.PrimaryVars = primaryVars
			//clauses.PrintDebug()
			//g.PrintDIMACS(clauses)
			g.Solve(clauses)
		}
	}

}

func printStats(stats []int) {
	if len(stats) != int(translation.TranslationTypes) {
		panic("Stats for translation errornous")
	}
	fmt.Printf("Facts\tClause\tAMO\tEx1\tCard\tBDD\tSN\n")

	for i, x := range stats {
		if i > 0 {
			fmt.Printf("%v\t", x)
		}
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

func parse(filename string) (pbs []constraints.Threshold, err error) {

	input, err := ioutil.ReadFile(filename)

	if err != nil {
		return pbs, err
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
					//variable, b2 := strconv.Atoi(digitRegexp.FindString(elements[i]))

					if b1 != nil {
						debug("cant convert to threshold:", l)
						panic("bad conversion of numbers")
					}
					atom := sat.NewAtomP(sat.Pred(elements[i]))
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
