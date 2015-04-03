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
var stat_flag = flag.Bool("stat", false, "Do statistics.")
var cat_flag = flag.Int("cat", 1, "Categorize method 1, or 2. (default 1).")

var complex_flag = flag.String("complex", "hybrid", "Solve complex PBs with mdd/sn/hybrid. Default is hybrid")
var timeout_flag = flag.Int("timeout", 100, "Timeout of the overall solving process")
var mdd_max_flag = flag.Int("mdd-max", 1000000, "Maximal Number of MDD Nodes in processing one PB.")
var mdd_redundant_flag = flag.Bool("mdd-redundant", true, "Reduce MDD by redundant nodes.")
var opt_bound_flag = flag.Int64("opt-bound", -1, "initial bound for optimization function <= value.")
var solver_flag = flag.String("solver", "minisat", "Choose Solver: minisat/clasp/lingeling/glucose/CCandr/cmsat.")
var seed_flag = flag.Int64("seed", 31415, "Random seed.")
var opt_rewrite_flag = flag.Bool("opt-rewrite", true, "Rewrites opt with chains from AMO and other constraint.")
var amo_reuse_flag = flag.Bool("amo-reuse", false, "Reuses AMO constraints for rewriting complex PBs.")
var rewrite_same_flag = flag.Bool("rewrite-same", false, "Groups same coefficients and introduces sorter and chains for them.")
var ex_chain_flag = flag.Bool("ex-chain", false, "Rewrites PBs with matching EXK constraints.")
var amo_chain_flag = flag.Bool("amo-chain", true, "Rewrites PBs with matching AMO.")

var digitRegexp = regexp.MustCompile("([0-9]+ )*[0-9]+")

var dbgoutput *os.File

func main() {
	flag.Parse()

	if *debug_filename != "" {

		var err error

		glob.Debug_output, err = os.Create(*debug_filename)

		if err != nil {
			panic(err)
		}
		defer glob.Debug_output.Close()
	}

	if *ver {
		fmt.Println(`Bule CNF Grounder: Tag 0.91 Pseudo Booleans
Copyright (C) NICTA and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
		return
	}

	// put all configuration here
	glob.Filename_flag = *filename_flag
	glob.Debug_flag = *debug_flag
	glob.Complex_flag = *complex_flag
	glob.Timeout_flag = *timeout_flag
	glob.MDD_max_flag = *mdd_max_flag
	glob.MDD_redundant_flag = *mdd_redundant_flag
	glob.Solver_flag = *solver_flag
	glob.Seed_flag = *seed_flag
	glob.Opt_rewrite_flag = *opt_rewrite_flag
	glob.Amo_reuse_flag = *amo_reuse_flag
	glob.Rewrite_same_flag = *rewrite_same_flag
	glob.Ex_chain_flag = *ex_chain_flag
	glob.Amo_chain_flag = *amo_chain_flag
	glob.Opt_bound_flag = *opt_bound_flag

	glob.D("Running Debug Mode...")

	pbs, err := parse(*filename_flag)
	opt := pbs[0] // per convention first in pbs is opt statement (possibly empty)
	if !opt.Empty() && opt.SumWeights() <= *opt_bound_flag {
		glob.D("opt.SumWeights <= *opt_bound", opt.SumWeights(), "<=", *opt_bound_flag)
		*opt_bound_flag = -1
	}

	if !opt.Empty() {
		opt.Normalize(constraints.LE, true)
	}

	if err != nil {
		err.Error()
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

		stats := make([]int, constraints.TranslationTypes)
		primaryVars := make(map[string]bool, 0)

		if opt.Empty() {
			for i, _ := range pbs {
				for _, x := range pbs[i].Entries {
					primaryVars[x.Literal.A.Id()] = true
				}
			}
		} else {
			for _, x := range opt.Entries {
				primaryVars[x.Literal.A.Id()] = true
			}
		}

		switch *cat_flag {
		case 1:

			for _, pb := range pbs[1:] {
				pb.Categorize1()
			}

		case 2:

			var tmp_opt constraints.Threshold
			if !glob.Opt_rewrite_flag {
				tmp_opt = (*opt).Copy()
				*opt = constraints.Threshold{}
			}

			constraints.Categorize2(pbs)

			if !glob.Opt_rewrite_flag {
				*opt = tmp_opt
			}

			if !opt.Empty() { // add new variables of auxiliaries added in the transformation
				for _, x := range opt.Entries {
					primaryVars[x.Literal.A.Id()] = true
				}
			}
		default:
			glob.A(false, "Categorization of constraints does not exist:", *cat_flag)
		}

		var clauses sat.ClauseSet
		for _, pb := range pbs {
			glob.A(pb.Empty() || pb.Typ == constraints.OPT || pb.Translated, "pbs", pb.Id, "has not been translated", pb)
			stats[pb.TransTyp]++
			//fmt.Println(pb.Id, pb.Clauses.Size())
			clauses.AddClauseSet(pb.Clauses)
		}
		//clauses.PrintDebug()

		if *stat_flag {
			//fmt.Print(*filename_flag, ";", len(primaryVars), ";", len(pbs), ";")
			//for i, x := range stats {
			//	if i > 0 {
			//		fmt.Printf("%v;", x)
			//	}
			//}
			//fmt.Println()
			printStats(stats)
		}

		if *solve_flag {
			g := sat.IdGenerator(clauses.Size() * 7)
			g.Filename = *out
			g.PrimaryVars = primaryVars
			g.Solve(clauses, opt, *opt_bound_flag)
			fmt.Println()
		}
	}
}

func printStats(stats []int) {

	glob.A(len(stats) == int(constraints.TranslationTypes), "Stats for translation errornous")

	trans := constraints.Facts
	for i := trans; i < constraints.TranslationTypes; i++ {
		if i > 0 {
			fmt.Printf("%v\t", constraints.TranslationType(i))
		}
	}
	fmt.Println()
	fmt.Print(glob.Filename_flag, ";")
	for i := trans; i < constraints.TranslationTypes; i++ {
		if i > 0 {
			fmt.Printf("%v;", stats[i])
		}
	}
	fmt.Println()
}

// returns list of *pb; first one is optimization statement, possibly empty
func parse(filename string) (pbs []*constraints.Threshold, err error) {

	input, err := ioutil.ReadFile(filename)

	if err != nil {
		return pbs, err
	}

	//TODO: use a buffered reader to prevent the whole file being in memory
	lines := strings.Split(string(input), "\n")

	// 0 : first line, 1 : rest of the lines
	var count int
	state := 1
	t := 0
	pbs = make([]*constraints.Threshold, 0, len(lines))

	for _, l := range lines {

		if l == "" || strings.HasPrefix(l, "%") || strings.HasPrefix(l, "*") {
			continue
		}

		elements := strings.Fields(l)

		if len(elements) == 1 { // quick hack to ignore single element lines (not neccessary)
			continue
		}

		switch state {
		case 0: // deprecated: for parsing the "header" of pb files, now parser is flexible
			{
				glob.D(l)
				var b1 error
				count, b1 = strconv.Atoi(elements[4])
				vars, b2 := strconv.Atoi(elements[2])
				if b1 != nil || b2 != nil {
					glob.D("cant convert to threshold:", l)
					panic("bad conversion of numbers")
				}
				glob.D("File PB file with", count, "constraints and", vars, "variables")
				state = 1
			}
		case 1:
			{

				var n int  // number of entries
				var f int  // index of entry
				var o bool //optimization
				var pb constraints.Threshold

				offset_back := 0
				if elements[len(elements)-1] != ";" {
					offset_back = 1
				}

				if elements[0] == "min:" || elements[0] == "Min" {
					o = true
					n = (len(elements) + offset_back - 2) / 2
					f = 1
				} else {
					o = false
					n = (len(elements) + offset_back - 3) / 2
					f = 0
				}

				pb.Entries = make([]constraints.Entry, n)

				for i := f; i < 2*n; i++ {

					weight, b1 := strconv.ParseInt(elements[i], 10, 64)
					i++
					if b1 != nil {
						glob.D("cant convert to threshold:", elements[i], "\nin PB\n", l)
						panic("bad conversion of numbers")
					}
					atom := sat.NewAtomP(sat.Pred(elements[i]))
					pb.Entries[(i-f)/2] = constraints.Entry{sat.Literal{true, atom}, weight}
				}
				// fake empty opt in case it does not exist
				if t == 0 && !o {
					pbs = append(pbs, &constraints.Threshold{})
					t++
				}
				pb.Id = t
				if o {
					pb.Typ = constraints.OPT
					glob.D("Scanned optimization statement")
				} else {
					pb.K, err = strconv.ParseInt(elements[len(elements)-2+offset_back], 10, 64)

					if err != nil {
						glob.A(false, " cant parse threshold, error", err.Error(), pb.K)
					}
					typS := elements[len(elements)-3+offset_back]

					if typS == ">=" {
						pb.Typ = constraints.GE
					} else if typS == "<=" {
						pb.Typ = constraints.LE
					} else if typS == "==" || typS == "=" {
						pb.Typ = constraints.EQ
					} else {
						glob.A(false, "cant convert to threshold, equationtype typS:", typS)
					}
				}

				pbs = append(pbs, &pb)
				t++
				//fmt.Println(pb.Id)
				//pb.Print10()
			}
		}
	}

	glob.A(len(pbs) == t, "Id of constraint must correspond to position")
	glob.D("Scanned", t, "PB constraints (including opt)")
	return
}
