/*
Copyright © 2020 Valentin Mayer-Eichberger <valentin@mayer-eichberger.de>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"bufio"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// dimacsCmd represents the dimacs command
var dimacsCmd = &cobra.Command{
	Use:   "dimacs",
	Short: "Grounds to dimacs completely.",
	Long: `This is a copy from the old grounder that.

		usage: ./ground <filename> [<unit>],

		unit is a sequence of literals, e.g. -3 2 5 -4
		minus meaning negated literal.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("c dimacs grounding called.")
		if len(args) == 0 {
			fmt.Println("usage: ./ground <filename> [units]")
			os.Exit(1)
		}

		units := make(map[string]bool, 0)

		for i, s := range args {
			if i < 2 {
				continue
			}
			if strings.HasPrefix(s, "-") {
				s = "~" + strings.TrimLeft(s, "-")
			}
			units[s] = true
		}
		fmt.Println("run with args")
		run(args[0], units)
	},
}

func init() {
	rootCmd.AddCommand(dimacsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dimacsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dimacsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(filename string, units map[string]bool) {

	printInfoFlag := true
	dimacs := true

	qvars := make(map[string]bool, 0)
	count := 1

	cls := [][]string{}
	last := ""
	alternation := [][]string{}

	// open a file or stream
	var scanner *bufio.Scanner
	file, err := os.Open(filename)
	if err != nil {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	//adjust the capacity to your need (max characters in line)
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {

		s := scanner.Text()

		if !strings.HasPrefix(s, "e ")  && !strings.HasPrefix(s, "a ")  {
			s = strings.ReplaceAll(s, " ", "")
			s = strings.ReplaceAll(s, "],", "] ")
			s = strings.ReplaceAll(s, "].", "]")
		}

		fields := strings.Fields(s)

		if len(fields) == 0 || strings.HasPrefix(fields[0], "%") {
			continue
		}

		first := fields[0]

		if first == "c" {
			continue
		}

		// merge consecutive e's and a's
		if first == "e" || first == "a" {
			for _, v := range fields[1:] {
				qvars[v] = true
			}
			if last == first {
				alternation[len(alternation)-1] = append(alternation[len(alternation)-1], fields[1:]...)
			} else {
				alternation = append(alternation, fields)
			}
			last = first
			continue
		}

		clause := fields

		if len(clause) == 1 {
			units[clause[0]] = true
		} else {
			cls = append(cls, clause)
		}
	}

	size := 0

	var cls2 [][]string
	conflict := false

	for size < len(units) {
		//fmt.Println("units", units)
		size = len(units)
		cls2 = make([][]string, 0, len(cls))

		for _, clause := range cls {
			clause2 := make([]string, 0, len(clause))
			keepClause := true

			//fmt.Println("clause", clause)
			for _, lit := range clause {
				if _, b := units[lit]; b {
					keepClause = false
				}
				//fmt.Println(units, lit, neg(lit))
				if _, b := units[neg(lit)]; !b {
					clause2 = append(clause2, lit)
				} else {
					//fmt.Println("remove", lit, "from", clause)
				}
			}
			//fmt.Println("clause2", clause2)
			if len(clause2) == 1 {
				units[clause2[0]] = true
			} else if len(clause2) == 0 {
				fmt.Println("c conflict:", clause)
				conflict = true
			}

			if keepClause && len(clause2) > 1 {
				cls2 = append(cls2, clause2)

			}
		}
		cls = cls2
	}

	vars := make(map[string]int, 0)
	{ // generate id's for variables
		for _, quantifier := range alternation {
			for i, v := range quantifier {
				if i == 0 {
					continue
				}
				if _, b := vars[v]; !b {
					vars[v] = count
					count++
				}
			}
		}
		for lit := range units {
			v := pos(lit)
			if _, b := vars[v]; !b {
				vars[v] = count
				count++
			}
		}

		for _, clause := range cls {
			for _, lit := range clause {
				v := pos(lit)
				if _, b := vars[v]; !b {
					vars[v] = count
					count++
				}
			}
		}
	}

	// TODO:remove units from all levels and dont give them in the formula
	//	{ // remove units from outermost E quantifier
	//		if len(alternation) > 0 && alternation[0][0] == "e" {
	//			for i := 1; i < len(alternation[0]); i++ {
	//				if units[alternation[0][i]] {
	//					alternation[0] = append(alternation[0][:i], alternation[0][i+1:]...)
	//					i--
	//				}
	//			}
	//		}
	//	}

	{ // create innermost EXIST for auxiliary variables
		var aux []string
		for v := range vars {
			if !qvars[v] {
				aux = append(aux, v)
			}
		}

		if len(aux) > 0 {
			if len(alternation) > 0 && alternation[len(alternation)-1][0] == "e" {
				alternation[len(alternation)-1] = append(alternation[len(alternation)-1], aux...)
			} else {
				alternation = append(alternation, append([]string{"e"}, aux...))
			}
		}
	}

	if dimacs {

		if printInfoFlag {
			varids := make([]string, len(vars)+1)
			for v, i := range vars {
				varids[i] = v
			}
			for i, v := range varids {
				if i > 0 {
					fmt.Println("c", i, v)
				}
			}
		}

		if conflict {
			fmt.Println("p cnf 1 1 \n 0\n")
			return
		}

		if printInfoFlag {
			fmt.Println("p", "cnf", len(vars), len(cls)+len(units))
		} else {
			fmt.Println("p", "cnf", len(vars)-len(units), len(cls))
		}
		for _, quantifier := range alternation {
			//fmt.Println(quantifier)
			if len(quantifier) == 1 {
				continue
			}
			for i, v := range quantifier {
				if i == 0 {
					fmt.Print(v, " ")
					continue
				}
				if !printInfoFlag && units[v] == true {
					continue
				}
				fmt.Print(vars[v], " ")
			}
			fmt.Println("0")
		}

		if printInfoFlag {
			for lit := range units {
				if strings.HasPrefix(lit, "~") {
					fmt.Print("-")
				}
				fmt.Print(vars[pos(lit)], " ")
				fmt.Println(0)
			}
		}

		for _, clause := range cls {

			for _, lit := range clause {
				if strings.HasPrefix(lit, "~") {
					fmt.Print("-")
				}
				fmt.Print(vars[pos(lit)], " ")
			}
			fmt.Println("0")
		}

	} else {

		for _, quantifier := range alternation {
			for _, v := range quantifier {
				fmt.Print(v, " ")
			}
			fmt.Println()
		}

		//     fmt.Println("c units")
		for unit := range units {
			fmt.Println(unit)

		}
		//		fmt.Println("c clauses")
		for _, clause := range cls {

			for _, v := range clause {
				fmt.Print(v, " ")
			}
			fmt.Println()
		}

	}

}

func pos(s string) string {
	if strings.HasPrefix(s, "~") {
		return strings.TrimLeft(s, "~")
	} else {
		return s
	}
}
func neg(s string) string {
	if strings.HasPrefix(s, "~") {
		return strings.TrimLeft(s, "~")
	} else {
		return "~" + s
	}
}
