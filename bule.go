package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/parser"
	"github.com/vale1410/bule/sat"
)

func main() {
	glob.Init()
	if *glob.Ver {
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
		*glob.Filename_flag = flag.Args()[0]
	}

	if *glob.Debug_filename != "" {

		var err error
		glob.Debug_file, err = os.Create(*glob.Debug_filename)

		if err != nil {
			panic(err)
		}
		defer glob.Debug_file.Close()
	}

	glob.D("Running Debug Mode...")

	problem := parser.New(*glob.Filename_flag)

	if *glob.Pbo_flag {
		problem.PrintPBO()
		return
	}

	if *glob.Gringo_flag {
		problem.PrintGringo()
		return
	}

	if *glob.Gurobi_flag {
		problem.PrintGurobi()
		return
	}

	pbs := problem.Pbs[1:] // opt is just a pointer to first in pbs.
	opt := problem.Opt

	primaryVars := make(map[string]bool, 0)

	for i, _ := range pbs {
		for _, x := range pbs[i].Entries {
			primaryVars[x.Literal.A.Id()] = true
		}
	}

	var clauses sat.ClauseSet

	// Categorize Version 1 (deprecated)
	switch *glob.Cat_flag {
	case 1:
		{
			for _, pb := range pbs {
				pb.Print10()
				pb.CategorizeTranslate1()
				clauses.AddClauseSet(pb.Clauses)
			}
		}
	case 2:
		{
			constraints.CategorizeTranslate2(pbs)
			for _, pb := range pbs {
				clauses.AddClauseSet(pb.Clauses)
			}
		}
	default:
		panic("Category not implemented")
	}

	if *glob.Dimacs_flag {
		clauses.PrintDebug()
	}

	if *glob.Solve_flag {
		g := sat.IdGenerator(clauses.Size()*7 + 1)
		g.PrimaryVars = primaryVars
		opt.NormalizePositiveCoefficients()
		opt.Offset = opt.K
		glob.A(opt.Positive(), "opt only has positive coefficients")
		g.Solve(clauses, opt, *glob.Opt_bound_flag, -opt.Offset)
		//fmt.Println()
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
	fmt.Print(*glob.Filename_flag, ";")
	for i := trans; i < constraints.TranslationTypes; i++ {
		if i > 0 {
			fmt.Printf("%v;", stats[i])
		}
	}
	fmt.Println()
}
