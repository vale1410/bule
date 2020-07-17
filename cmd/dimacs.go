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
	"bufio"
	"fmt"
	"github.com/vale1410/bule/lib"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type ClauseProgram struct {
	clauses     [][]string
	alternation [][]string // level 0 is exist, then alternating!
	units       map[string]bool
	conflict    bool
	idMap       map[string]int
}

var (
	printInfoFlag       bool
	unitPropagationFlag bool
)

// dimacsCmd represents the dimacs command
var dimacsCmd = &cobra.Command{
	Use:   "dimacs",
	Short: "Grounds to dimacs completely.",
	Long: `This is a copy from the old grounder that.

		usage: bule dimacs <filename> <units>,

		unit is a sequence of literals, e.g. -3 2 5 -4
		where minus meaning negated literal.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("c dimacs grounding called.")
		if len(args) == 0 {
			fmt.Println("usage: bule dimacs <filename> [units]")
			os.Exit(1)
		}
		units := convertArgsToUnits(args[1:])
		fmt.Println("prepare with args")
		p := parseFromFile(args[0])
		sb := p.prepare(units)
		fmt.Println(sb.String())
	},
}

func convertArgsToUnits(args []string) map[string]bool {
	units := make(map[string]bool, 0)

	for _, s := range args {
		if strings.HasPrefix(s, "-") {
			s = "~" + strings.TrimLeft(s, "-")
		}
		units[s] = true
	}
	return units
}

func init() {
	rootCmd.AddCommand(dimacsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dimacsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only prepare when this command
	// is called directly, e.g.:
	//dimacsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//	groundCmd.PersistentFlags().StringVarP(&progFlag, "file", "f", "", "Path to File")
	dimacsCmd.PersistentFlags().BoolVarP(&printInfoFlag, "info", "i", true, "Print information about units as well.")
}

// This translate from a grounded Bule program to clause representation
// Assuming the program is ground!!!
func translateFromRuleProgram(program lib.Program, units map[string]bool) (p ClauseProgram) {
	p.units = units
	fmt.Println("ARRAY: ", program.IntId2String)
	fmt.Println("BOOLs: ", program.PredicateStringTerm)
	for _, r := range program.Rules {
		if len(r.Literals) == 1 {
			p.units[program.OutputString(r.Literals[0])] = true
			continue
		}
		clause := make([]string, 0, len(r.Literals))
		for _, l := range r.Literals {
			s := program.OutputString(l)
			fmt.Println("TEST", s)
			clause = append(clause, s)
		}
		p.clauses = append(p.clauses, clause)
	}

	// Assuming quantification levels have been merged
	for _, level := range program.Alternation {
		q := make([]string, len(level))
		for i, literal := range level {
			q[i] = program.OutputString(literal)
		}
		p.alternation = append(p.alternation, q)
	}

	return
}

func parseFromFile(filename string) (p ClauseProgram) {
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

	acc := ""
	last := ""

	p.clauses = [][]string{}
	p.units = map[string]bool{}
	var alternation [][]string
	alternationDepth := 0

	for scanner.Scan() {

		s := scanner.Text()

		if !strings.HasPrefix(s, "e ") && !strings.HasPrefix(s, "a ") {
			s = strings.ReplaceAll(s, " ", ",")
			//			s = strings.ReplaceAll(s, "],", "] ")
			//			s = strings.ReplaceAll(s, "].", "]")
		}

		s = strings.ReplaceAll(s, " ", "")
		if s == "" || strings.HasPrefix(s, "%") || strings.HasPrefix(s, "c") {
			continue
		}

		if !strings.HasSuffix(s, ".") {
			acc += s
			continue
		} else {
			s = s[:len(s)-1]
			s = acc
			acc = ""
		}

		fields := strings.Split(s, ",")
		first := fields[0]

		// merge consecutive e's and a's
		if first == "e" || first == "a" {
			if last == first {
				alternation[len(alternation)-1] = append(alternation[len(alternation)-1], fields[1:]...)
			} else {
				if alternationDepth == 0 && first == "a" {
					alternation = append(alternation, fields)
				}
				alternation = append(alternation, fields)
				alternationDepth++
			}
			last = first
			continue
		}

		clause := fields

		if len(clause) == 1 {
			p.units[clause[0]] = true
		} else {
			p.clauses = append(p.clauses, clause)
		}
	}
	p.alternation = alternation
	return
}

func (p *ClauseProgram) generateIds() {

	//make Ids
	count := 1
	vars := make(map[string]int, 0)
	// generate id's for variables
	for _, quantifier := range p.alternation {
		for _, v := range quantifier {
			if _, b := vars[v]; !b {
				vars[v] = count
				count++
			}
		}
	}
	for lit := range p.units {
		v := pos(lit)
		if _, b := vars[v]; !b {
			vars[v] = count
			count++
		}
	}

	for _, clause := range p.clauses {
		for _, lit := range clause {
			v := pos(lit)
			if _, b := vars[v]; !b {
				vars[v] = count
				count++
			}
		}
	}
	p.idMap = vars
}

// Finds all variables that are not bound and adds them to the
// innermost level
func (p *ClauseProgram) createInnermostExistFromFreeVars() {

	qvars := make(map[string]bool, 0) // vars in quantifier alternation
	for _, quantifier := range p.alternation {
		for _, v := range quantifier {
			qvars[v] = true
		}
	}

	var aux []string
	for v := range p.idMap {
		if !qvars[v] {
			aux = append(aux, v)
		}
	}

	if len(aux) > 0 {
		if len(p.alternation)%2 == 1 {
			p.alternation[len(p.alternation)-1] = append(p.alternation[len(p.alternation)-1], aux...)
		} else {
			p.alternation = append(p.alternation, aux)
		}
	}
}

func (p ClauseProgram) PrintDimacs() strings.Builder {
	sb := strings.Builder{}
	vars := p.idMap
	conflict := p.conflict
	cls := p.clauses
	units := p.units

	if printInfoFlag {
		varids := make([]string, len(vars)+1)
		for v, i := range vars {
			varids[i] = v
		}
		for i, v := range varids {
			if i > 0 {
				sb.WriteString(fmt.Sprintln("c", i, v))
			}
		}
	}

	if conflict {
		sb.WriteString(fmt.Sprintln("p cnf 1 1 \n 0\n"))
		return sb
	}

	if printInfoFlag {
		sb.WriteString(fmt.Sprintln("p", "cnf", len(vars), len(cls)+len(units)))
	} else {
		sb.WriteString(fmt.Sprintln("p", "cnf", len(vars)-len(units), len(cls)))
	}
	for i, quantifier := range p.alternation {

		if len(quantifier) == 0 {
			continue
		}

		if i%2 == 0 {
			sb.WriteString(fmt.Sprint("e "))
		} else {
			sb.WriteString(fmt.Sprint("a "))
		}

		for _, v := range quantifier {
			if !printInfoFlag && units[v] == true {
				continue
			}
			sb.WriteString(fmt.Sprint(vars[v], " "))
		}
		sb.WriteString(fmt.Sprintln("0"))
	}

	if printInfoFlag {
		for lit := range units {
			if strings.HasPrefix(lit, "~") {
				sb.WriteString(fmt.Sprint("-"))
			}
			sb.WriteString(fmt.Sprint(vars[pos(lit)], " "))
			sb.WriteString(fmt.Sprintln(0))
		}
	}

	if !textualFlag {
		for _, clause := range cls {
			for _, lit := range clause {
				if strings.HasPrefix(lit, "~") {
					sb.WriteString(fmt.Sprint("-"))
				}
				sb.WriteString(fmt.Sprint(vars[pos(lit)], " "))
			}
			sb.WriteString(fmt.Sprintln("0"))
		}
	} else {
		// printout textual representation!!
		for _, clause := range cls {
			for i, lit := range clause {
				if i != 0 {
					sb.WriteString(fmt.Sprint(", "))
				}
				sb.WriteString(fmt.Sprint(lit))
			}
			sb.WriteString(fmt.Sprintln("."))
		}
	}
	return sb
}

// This is a very slow implementation of unit propagation
// units and clauses within program are updated.
func (p *ClauseProgram) unitPropagation() {

	size := 0
	var cls2 [][]string

	// Unit propagation
	for size < len(p.units) {
		//fmt.Println("units", units)
		size = len(p.units)
		cls2 = make([][]string, 0, len(p.clauses))

		for _, clause := range p.clauses {
			clause2 := make([]string, 0, len(clause))
			keepClause := true

			//fmt.Println("clause", clause)
			for _, lit := range clause {
				if _, b := p.units[lit]; b {
					keepClause = false
				}
				//fmt.Println(units, lit, negate(lit))
				if _, b := p.units[negate(lit)]; !b {
					clause2 = append(clause2, lit)
				} else {
					//fmt.Println("remove", lit, "from", clause)
				}
			}
			//fmt.Println("clause2", clause2)
			if len(clause2) == 1 {
				p.units[clause2[0]] = true
			} else if len(clause2) == 0 {
				debug(2, "c conflict:", clause)
				p.conflict = true
			}

			if keepClause && len(clause2) > 1 {
				cls2 = append(cls2, clause2)
			}
		}
		p.clauses = cls2
	}
	return
}

func (p *ClauseProgram) StringBuilder() strings.Builder {
	sb := p.prepare(map[string]bool{})
	return  sb
}

func (p *ClauseProgram) prepare(additionalUnits map[string]bool) strings.Builder {

	for unit := range additionalUnits {
		p.units[unit] = true
	}

	if unitPropagationFlag {
		p.unitPropagation()
	}
	p.generateIds()
	p.createInnermostExistFromFreeVars()
	return p.PrintDimacs()
}

func pos(s string) string {
	if strings.HasPrefix(s, "~") {
		return strings.TrimLeft(s, "~")
	} else {
		return s
	}
}
func negate(s string) string {
	if strings.HasPrefix(s, "~") {
		return strings.TrimLeft(s, "~")
	} else {
		return "~" + s
	}
}
