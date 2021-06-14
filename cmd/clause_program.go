package cmd

import (
	"fmt"
	"github.com/vale1410/bule/grounder"
	"strings"
)

type ClauseProgram struct {
	clauses     [][]string
	alternation [][]string // level 0 is exist, then alternating!
	conflict    bool
	idMap       map[string]int
}

// This translate from a grounded Bule program to clause representation
// Assuming the program is ground!!!
func translateFromRuleProgram(program grounder.Program) (p ClauseProgram) {

	for _, r := range program.Rules {
		clause := make([]string, 0, len(r.Literals))
		for _, l := range r.Literals {
			s := program.OutputString(l)
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

	if !textualFlag {
		if conflict {
			sb.WriteString(fmt.Sprintln("p cnf 1 1 \n 0\n"))
			return sb
		}

		sb.WriteString(fmt.Sprintln("p", "cnf", len(vars), len(cls)))
	}

	if quantificationFlag {
		for i, quantifier := range p.alternation {

			if len(quantifier) == 0 {
				continue
			}
			if textualFlag {
				if i%2 == 0 {
					sb.WriteString(fmt.Sprint("#exists "))
				} else {
					sb.WriteString(fmt.Sprint("#forall "))
				}
				for _, v := range quantifier {
					sb.WriteString(v + " ")
				}
				sb.WriteString(fmt.Sprintln("."))
			} else {
				if i%2 == 0 {
					sb.WriteString(fmt.Sprint("e "))
				} else {
					sb.WriteString(fmt.Sprint("a "))
				}
				for _, v := range quantifier {
					sb.WriteString(fmt.Sprint(vars[v], " "))
				}
				sb.WriteString(fmt.Sprintln("0"))
			}
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
					sb.WriteString(fmt.Sprint(" | "))
				}
				sb.WriteString(fmt.Sprint(lit))
			}
			sb.WriteString(fmt.Sprintln("."))
		}
	}
	return sb
}

func (p *ClauseProgram) StringBuilder() strings.Builder {
	sb := p.prepare(map[string]bool{})
	return sb
}

func (p *ClauseProgram) prepare(additionalUnits map[string]bool) strings.Builder {

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
