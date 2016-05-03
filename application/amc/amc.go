package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/parser"
	"github.com/vale1410/bule/sat"
)

func main() {

	glob.Init()

	if *glob.Ver {
		fmt.Println(`Approximate Model Counter: Tag 0.1
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

	p := parser.New(*glob.Filename_flag)
	pbs := p.Pbs

	primaryVars := make(map[string]bool, 0)

	for i, _ := range pbs {
		for _, x := range pbs[i].Entries {
			primaryVars[x.Literal.A.Id()] = true
		}
	}

	var clauses sat.ClauseSet
	for _, pb := range pbs {
		//fmt.Println(pb)
		pb.Simplify()
		//glob.A(pb.Empty() || pb.Typ == constraints.OPT || pb.Translated, "pbs", pb.Id, "has not been translated", pb)
		if pb.Empty() {
			continue
		}
		pb.TranslateComplexThreshold()

		clauses.AddClauseSet(pb.Clauses)
	}

	if *glob.Dimacs_flag {
		clauses.PrintDebug()
	}

	if *glob.Solve_flag {
		g := sat.IdGenerator(clauses.Size() * 7)
		g.PrimaryVars = primaryVars
		//fmt.Println()
		primaryVars := make(map[string]bool, 0)

		for i, _ := range pbs {
			for _, x := range pbs[i].Entries {
				primaryVars[x.Literal.A.Id()] = true
			}
		}

		//clauses.PrintDebug()
		//g.Solve(clauses, opt, *glob.Opt_bound_flag, -opt.Offset)

		inferPrimeVars := true
		g.PrintDIMACS(clauses, inferPrimeVars)
		//fmt.Println()
	}
}
