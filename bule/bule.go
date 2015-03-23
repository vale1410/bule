package main

import (
	"flag"
	"fmt"
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
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
var debug_flag = flag.Bool("d", false, "Print debug information.")
var debug_filename = flag.String("df", "", "File to print debug information.")
var reformat_flag = flag.Bool("reformat", false, "Reformat PB files into correct format. Decompose = into >= and <=")
var gurobi_flag = flag.Bool("gurobi", false, "Reformat to Gurobi input, output to stdout.")
var solve_flag = flag.Bool("solve", false, "Dont solve just categorize and analyze the constriants.")

var complex_flag = flag.String("complex", "hybrid", "Solve complex PBs with mdd/sn/hybrid. Default is hybrid")
var timeout_flag = flag.Int("timeout", 100, "Timeout of the overall solving process")
var mdd_max_flag = flag.Int("mdd_max", 300000, "Maximal Number of MDD Nodes in processing one PB.")
var mdd_redundant_flag = flag.Bool("mdd_redundant", true, "Reduce MDD by redundant nodes.")

var digitRegexp = regexp.MustCompile("([0-9]+ )*[0-9]+")

var dbgoutput *os.File

func main() {
	flag.Parse()

	if *debug_filename != "" {
		var err error
		glob.Debug_filename = *debug_filename
		glob.Debug_output, err = os.Create(*debug_filename)
		if err != nil {
			panic(err)
		}
		defer glob.Debug_output.Close()
	}

	if *ver {
		fmt.Println(`Bule CNF Grounder: Tag 0.8 Pseudo Booleans
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
		return
	}

	// put all configuration here
	glob.Debug_flag = *debug_flag
	glob.Complex_flag = *complex_flag
	glob.Timeout_flag = *timeout_flag
	glob.MDD_max_flag = *mdd_max_flag
	glob.MDD_redundant_flag = *mdd_redundant_flag

	glob.D("Running Debug Mode...")

	opt, pbs, err := parse(*filename_flag)
	if err != nil {
		err.Error()
	}

	if !opt.Empty() {
		glob.D("Ignoring optimization statement")
	}

	if *gurobi_flag {
		fmt.Println("Subject To")
		atoms := make(map[string]bool, len(pbs))
		for _, pb := range pbs {
			pb.Normalize(constraints.GE, false)
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
			if pb.Typ == constraints.EQ {
				pb.Typ = constraints.GE
				pb.Normalize(constraints.GE, false)
				pb.Print10()
				pb.Typ = constraints.LE
			}
			pb.Normalize(constraints.GE, false)
			pb.Print10()
		}

	} else {

		var clauses sat.ClauseSet

		//if *stats {
		//// stats start
		//fmt.Print(*filename_flag, ";", len(primaryVars), ";", len(pbs), ";")
		//for i, x := range stats {
		//	if i > 0 {
		//		fmt.Printf("%v;", x)
		//	}
		//}
		//fmt.Println()
		//printStats(stats)
		//// stats end
		//	}

		if !*solve_flag { // do statistics

			ppbs := make([]*constraints.Threshold, len(pbs))
			for i, _ := range pbs {
				pbs[i].SortWeight()
				ppbs[i] = &pbs[i]
			}
			constraints.Categorize2(ppbs)

		} else if *solve_flag {
			stats := make([]int, constraints.TranslationTypes)
			primaryVars := make(map[string]bool, 0)
			for i, _ := range pbs {
				for _, x := range pbs[i].Entries {
					primaryVars[x.Literal.A.Id()] = true
				}
				t := constraints.Categorize1(&pbs[i])
				stats[t.Typ]++
				clauses.AddClauseSet(t.Clauses)
			}

			fmt.Print(*filename_flag)
			g := sat.IdGenerator(clauses.Size() * 7)
			g.Filename = *out
			g.PrimaryVars = primaryVars
			//clauses.PrintDebug()
			g.Solve(clauses)
		}
	}

}

func printStats(stats []int) {
	if len(stats) != int(constraints.TranslationTypes) {
		panic("Stats for translation errornous")
	}
	fmt.Printf("Facts\tClause\tAMO\tEx1\tCard\tMDD\tSN\n")

	for i, x := range stats {
		if i > 0 {
			fmt.Printf("%v\t", x)
		}
	}
	fmt.Println()
}

func parse(filename string) (opt constraints.Threshold, pbs []constraints.Threshold, err error) {

	input, err := ioutil.ReadFile(filename)

	if err != nil {
		return opt, pbs, err
	}

	output, err := os.Create(*out)
	if err != nil {
		err.Error()
	}
	defer output.Close()

	//TODO: use a buffered reader to prevent the whole file being in memory
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
				glob.D(l)
				var b1 error
				count, b1 = strconv.Atoi(elements[4])
				vars, b2 := strconv.Atoi(elements[2])
				if b1 != nil || b2 != nil {
					glob.D("cant convert to threshold:", l)
					panic("bad conversion of numbers")
				}
				glob.D("Found PB file with", count, "constraints and", vars, "variables")
				pbs = make([]constraints.Threshold, count)
				state = 1
			}
		case 1:
			{
				glob.A(t <= count, "Number of constraints incorrectly specified in pb input file ", filename)

				var n int  // number of entries
				var f int  // index of entry
				var o bool //optimization
				var pb constraints.Threshold

				if "min:" == elements[0] {
					o = true
					n = (len(elements) - 2) / 2
					f = 1
				} else {
					o = false
					n = (len(elements) - 3) / 2
					f = 0
				}

				pb.Entries = make([]constraints.Entry, n)

				for i := f; i < 2*n; i++ {

					weight, b1 := strconv.ParseInt(elements[i], 10, 64)
					i++
					if b1 != nil {
						glob.D("cant convert to threshold:", l)
						panic("bad conversion of numbers")
					}
					atom := sat.NewAtomP(sat.Pred(elements[i]))
					pb.Entries[(i-f)/2] = constraints.Entry{sat.Literal{true, atom}, weight}
				}

				if o {
					pb.Typ = constraints.OPT
					opt = pb
				} else {
					pb.K, _ = strconv.ParseInt(elements[len(elements)-2], 10, 64)
					typS := elements[len(elements)-3]

					if typS == ">=" {
						pb.Typ = constraints.GE
					} else if typS == "<=" {
						pb.Typ = constraints.LE
					} else if typS == "==" || typS == "=" {
						pb.Typ = constraints.EQ
					} else {
						glob.D("cant convert to threshold:", l)
						panic("bad conversion of symbols" + typS)
					}
					pb.Id = t
					pbs[t] = pb
					t++
				}
			}
		}
	}

	glob.A(t == count, t, count, "Number of constraints incorrectly specified in pb input file ", filename)

	return
}
