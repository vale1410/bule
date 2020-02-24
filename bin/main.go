package main

import (
	"flag"
	"fmt"
	"github.com/vale1410/bule"
)

var (
	debugFlag = flag.Int("d", 0, "Debug Level .")
	progFlag  = flag.String("f", "", "Path to file.")
)

func debug(level int, s ...interface{}) {
	if level <= *debugFlag {
		fmt.Println(s...)
	}
}

func main() {

	flag.Parse()

	p := bule.ParseProgram(*progFlag)
	bule.DebugLevel = *debugFlag

	debug(2, "\nReplace Constants")
	p.ReplaceConstants()

	debug(2, "\nExpand Generators")
	p.ExpandGenerators()

	// forget about heads now!
	debug(2, "\nRewrite Equivalences")
	p.RewriteEquivalencesAndImplications()

	// There are no equivalences and no generators anymore !

	{
		debug(2, "Grounding:")
		clauses, existQ, forallQ, maxIndex := p.Ground()

		// Do Unit Propagation

		// Find variables that need to be put in the quantifier alternation

		for i := 0; i <= maxIndex; i++ {

			if atoms, ok := forallQ[i]; ok {
				fmt.Print("a")
				for _, a := range atoms {
					fmt.Print(" ", a)
				}
				fmt.Println()
			}
			if atoms, ok := existQ[i]; ok {
				fmt.Print("e")
				for _, a := range atoms {
					fmt.Print(" ", a)
				}
				fmt.Println()
			}
		}

		for _, clause := range clauses {
			for _, lit := range clause {
				fmt.Print(lit.String(), " ")
			}
			fmt.Println()
		}
	}
}
