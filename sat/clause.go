package sat

import (
	"fmt"
)

type Clause struct {
	Desc     string
	Literals []Literal
}

type ClauseSet struct {
	list []Clause
}

func NewClauseSet(n int) (cs ClauseSet) {
	cs.list = make([]Clause, n)
	return cs
}

func (cs *ClauseSet) Size() int {
    return len(cs.list)
}


func (cs *ClauseSet) AddClause(literals ...Literal) {
	cs.list = append(cs.list, Clause{"-", literals})
}

func (cs *ClauseSet) AddTaggedClause(tag string, literals ...Literal) {
	cs.list = append(cs.list, Clause{tag, literals})
}

func (cs *ClauseSet) AddClauseSet(cl ClauseSet) {
	cs.list = append(cs.list, cl.list...)
}

func (cs *ClauseSet) Print() {

	stat := make(map[string]int, 0)
	var descs []string

	for _, c := range cs.list {

		count, ok := stat[c.Desc]
		if !ok {
			stat[c.Desc] = 1
			descs = append(descs, c.Desc)
		} else {
			stat[c.Desc] = count + 1
		}

		fmt.Printf("c %s\t", c.Desc)
		first := true
		for _, l := range c.Literals {
			if !first {
				fmt.Printf(",")
			}
			first = false
			fmt.Print(l.ToTxt())
		}
		fmt.Println(".")
	}

	for _, key := range descs {
		fmt.Printf("c %v\t: %v\t%.1f \n", key, stat[key], 100*float64(stat[key])/float64(len(cs.list)))
	}
	fmt.Printf("c %v\t: %v\t%.1f \n", "tot", len(cs.list), 100.0)
}
