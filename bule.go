package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
)

var (
	filename_flag = flag.String("f", "test.pb", "Path of to PB file.")
	cnf_tmp_flag  = flag.String("out", "", "If set: output cnf to this file.")
	ver           = flag.Bool("ver", false, "Show version info.")

	debug_flag     = flag.Bool("d", false, "Print debug information.")
	debug_filename = flag.String("df", "", "File to print debug information.")
	pbo_flag       = flag.Bool("pbo", false, "Reformat to pbo format, output to stdout.")
	gringo_flag    = flag.Bool("gringo", false, "Reformat to Gringo format, output to stdout.")
	gurobi_flag    = flag.Bool("gurobi", false, "Reformat to Gurobi input, output to stdout.")
	solve_flag     = flag.Bool("solve", true, "Dont solve just categorize and analyze the constraints.")
	dimacs_flag    = flag.Bool("dimacs", false, "Print readable format of clauses.")
	stat_flag      = flag.Bool("stat", false, "Extended statistics on types of PBs in problem.")
	cat_flag       = flag.Int("cat", 2, "Categorize method 1, or 2. (default 2, historic: 1).")

	complex_flag         = flag.String("complex", "hybrid", "Solve complex PBs with mdd/sn/hybrid. Default is hybrid")
	timeout_flag         = flag.Int("timeout", 600, "Timeout of the overall solving process")
	mdd_max_flag         = flag.Int("mdd-max", 2000000, "Maximal number of MDD Nodes in processing one PB.")
	mdd_redundant_flag   = flag.Bool("mdd-redundant", true, "Reduce MDD by redundant nodes.")
	opt_bound_flag       = flag.Int64("opt-bound", math.MaxInt64, "Initial bound for optimization function <= given value. Negative values allowed.")
	opt_half_flag        = flag.Bool("opt-half", false, "Sets opt-bound to half the sum of the weights of the optimization function.")
	solver_flag          = flag.String("solver", "minisat", "Choose Solver: minisat/clasp/lingeling/glucose/CCandr/cmsat.")
	seed_flag            = flag.Int64("seed", 42, "Random seed initializer.")
	amo_reuse_flag       = flag.Bool("amo-reuse", false, "Reuses AMO constraints for rewriting complex PBs.")
	rewrite_opt_flag     = flag.Bool("opt-rewrite", true, "Rewrites opt with chains from AMO and other constraint.")
	rewrite_same_flag    = flag.Bool("rewrite-same", false, "Groups same coefficients and introduces sorter and chains for them.")
	rewrite_equal_flag   = flag.Bool("rewrite-equal", false, "Rewrites complex == constraints into >= and <=.")
	ex_chain_flag        = flag.Bool("ex-chain", false, "Rewrites PBs with matching EXK constraints.")
	amo_chain_flag       = flag.Bool("amo-chain", true, "Rewrites PBs with matching AMO.")
	search_strategy_flag = flag.String("search", "iterative", "Search objective iterative or binary.")
)

type Problem struct {
	opt *constraints.Threshold
	pbs []*constraints.Threshold
}

func (p *Problem) PrintPBO() {
	atoms := make(map[string]bool, len(p.pbs))

	for _, pb := range p.pbs {
		for _, x := range pb.Entries {
			atoms[x.Literal.A.Id()] = true
		}
	}
	fmt.Printf("* #variable= %v #constraint= %v\n", len(atoms), len(p.pbs)-1)

	for _, pb := range p.pbs {
		pb.PrintPBO()
	}
}

func (p *Problem) PrintGringo() {
	fmt.Println("#hide.")
	atoms := make(map[string]bool, len(p.pbs))

	for _, pb := range p.pbs {
		pb.PrintGringo()
		for _, x := range pb.Entries {
			atoms[x.Literal.A.Id()] = true
		}
	}
	for x, _ := range atoms {
		fmt.Println("{", x, "}.")
	}
}

func (p *Problem) PrintGurobi() {
	if !p.opt.Empty() {
		fmt.Println("Minimize")
		p.opt.PrintGurobi()
	}
	fmt.Println("Subject To")
	atoms := make(map[string]bool, len(p.pbs))
	for i, pb := range p.pbs {
		if i > 0 {
			pb.Normalize(constraints.GE, false)
			pb.PrintGurobi()
			for _, x := range pb.Entries {
				atoms[x.Literal.A.Id()] = true
			}
		}
	}
	fmt.Println("Binary")
	for aS, _ := range atoms {
		fmt.Print(aS + " ")
	}
	fmt.Println()
}

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
		fmt.Println(`Bule CNF Grounder: Tag 0.95 Pseudo Booleans
Copyright (C) Data61 and Valentin Mayer-Eichberger
License GPLv2+: GNU GPL version 2 or later <http://gnu.org/licenses/gpl.html>
There is NO WARRANTY, to the extent permitted by law.`)
		return
	}

	if len(flag.Args()) >= 2 {
		fmt.Println("Command line flags not recognized", flag.Args())
		return
	}

	if len(flag.Args()) == 1 {
		glob.Filename_flag = flag.Args()[0]
	} else {
		glob.Filename_flag = *filename_flag
	}

	// put all configuration here
	glob.Debug_flag = *debug_flag
	glob.Complex_flag = *complex_flag
	glob.Timeout_flag = *timeout_flag
	glob.MDD_max_flag = *mdd_max_flag
	glob.MDD_redundant_flag = *mdd_redundant_flag
	glob.Solver_flag = *solver_flag
	glob.Seed_flag = *seed_flag
	glob.Amo_reuse_flag = *amo_reuse_flag
	glob.Rewrite_opt_flag = *rewrite_opt_flag
	glob.Rewrite_same_flag = *rewrite_same_flag
	glob.Ex_chain_flag = *ex_chain_flag
	glob.Amo_chain_flag = *amo_chain_flag
	glob.Opt_bound_flag = *opt_bound_flag
	glob.Cnf_tmp_flag = *cnf_tmp_flag
	glob.Search_strategy_flag = *search_strategy_flag

	glob.D("Running Debug Mode...")

	pbs, err := parse(glob.Filename_flag)
	opt := pbs[0] // per convention first in pbs is opt statement (possibly empty)
	p := Problem{opt, pbs}

	if *rewrite_equal_flag {
		before := len(pbs)
		for _, x := range pbs {
			//if x.Typ == constraints.EQ && x.IsComplex() {
			if x.Typ == constraints.EQ { // rewrite all constraints into this
				y := x.Copy()
				x.Typ = constraints.LE
				y.Typ = constraints.GE
				y.Id = len(pbs)
				pbs = append(pbs, &y)
			}
		}
		if len(pbs)-before > 0 {
			glob.D("c rewritten", len(pbs)-before, "equality constraints into >= and <=.")
		}
	}

	//reformatting
	if *pbo_flag {
		p.PrintPBO()
		return
	} else if *gringo_flag {
		p.PrintGringo()
		return
	} else if *gurobi_flag {
		p.PrintGurobi()
		return
	}

	if !opt.Empty() && *opt_half_flag {
		redundant := opt.Copy()
		redundant.K = opt.SumWeights() / 2
		redundant.Typ = constraints.LE
		redundant.Id = len(pbs)
		glob.D("initializing opt sumWeights/2 = ", redundant.K)
		*opt = constraints.Threshold{}
		pbs = append(pbs, &redundant)
	}

	//transform opt
	if !opt.Empty() {
		opt.NormalizePositiveCoefficients()
		opt.Offset = opt.K
		//fmt.Println("offset :", opt.Offset)
	}

	if err != nil {
		err.Error()
	}

	{

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
			if !glob.Rewrite_opt_flag {
				tmp_opt = (*opt).Copy()
				*opt = constraints.Threshold{}
			}

			constraints.Categorize2(pbs)

			if !glob.Rewrite_opt_flag {
				*opt = tmp_opt
			}
			//opt.Print10()

			if !opt.Empty() { // add new variables of auxiliaries added in the transformation
				for _, x := range opt.Entries {
					primaryVars[x.Literal.A.Id()] = true
				}
			}
			//fmt.Println(primaryVars)
		default:
			glob.A(false, "Categorization of constraints does not exist:", *cat_flag)
		}

		var clauses sat.ClauseSet
		for _, pb := range pbs {
			glob.A(pb.Empty() || pb.Typ == constraints.OPT || pb.Translated, "pbs", pb.Id, "has not been translated", pb)
			stats[pb.TransTyp]++
			clauses.AddClauseSet(pb.Clauses)
		}

		if *dimacs_flag {
			clauses.PrintDebug()
		}

		if *stat_flag {
			printStats(stats)
		}

		if *solve_flag {
			g := sat.IdGenerator(clauses.Size() * 7)
			g.PrimaryVars = primaryVars
			glob.A(opt.Positive(), "opt only has positive coefficients")
			g.Solve(clauses, opt, *opt_bound_flag, -opt.Offset)
			//fmt.Println()
		}
	}
}

func printStats(stats []int) {

	glob.A(len(stats) == int(constraints.TranslationTypes), "Stats for translation errornous")

	trans := constraints.Facts
	fmt.Print("Name;")
	for i := trans; i < constraints.TranslationTypes; i++ {
		if i > 0 {
			fmt.Printf("%v;", constraints.TranslationType(i))
		}
	}
	fmt.Println()
	fmt.Print("xxx", glob.Filename_flag, ";")
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
