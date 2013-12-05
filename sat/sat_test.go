package sat

// test class, but will eventuall be turned into the sat package :-)

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
    "../sorters"
)

func TestWhichClauses(t *testing.T) {

	//sizes := []int{100,112,128,144,160,176}
	sizes := []int{500,750,1000}
    //sizes := []int{}
	//typs := []sorters.SortingNetworkType{Bubble, Bitonic, OddEven, Pairwise}
	//typs := []sorters.SortingNetworkType{Bitonic, OddEven, Pairwise}
	typs := []sorters.SortingNetworkType{OddEven, Pairwise}
	whichT := []int{1, 2, 3, 4}
	lt := Pred("AtMost")
	gt := Pred("AtLeast")

	for _, size := range sizes {
		for _, typ := range typs {
			for _, wh := range whichT {
				//k := int(0.05 * float64(size))
				k := size / 4
				//k := size - size/4
				sorter1 := sorters.CreateCardinalityNetwork(size, k, AtMost, typ)
				sorter2 := sorters.CreateCardinalityNetwork(size, k+1, AtLeast, typ)
				sorter1.RemoveOutput()
				sorter2.RemoveOutput()

				var which1 [8]bool
				var which2 [8]bool

				switch wh {
				case 1:
					which1 = [8]bool{false, false, false, true, true, true, false, false}
					which2 = [8]bool{false, true, true, false, false, false, true, false}
				case 2:
					which1 = [8]bool{false, false, false, true, true, true, false, true}
					which2 = [8]bool{false, true, true, false, false, false, true, true}
				case 3:
					which1 = [8]bool{false, true, true, true, true, true, true, false}
					which2 = [8]bool{false, true, true, true, true, true, true, false}
				case 4:
					which1 = [8]bool{false, true, true, true, true, true, true, true}
					which2 = [8]bool{false, true, true, true, true, true, true, true}
				}

				input := make([]Literal, size)
				for i, _ := range input {
					input[i] = Literal{true, Atom{Pred("Input"), i, 0}}
				}

				clauses := createEncoding(input, which1, []Literal{}, "lt", lt, sorter1)
				clauses.AddClauseSet(createEncoding(input, which2, []Literal{}, "gt", gt, sorter2))
	            g := IdGenerator(size * size)
	            g.GenerateIds(clauses)
	            g.filename = strconv.Itoa(size) + "_" + strconv.Itoa(k) + "_" + typ.String() + "_" + strconv.Itoa(wh)+".cnf"
	            g.printClausesDIMACS(clauses)
			}
		}
	}

}

//func TestGenerateSAT(t *testing.T) {
//	size := 128
//	k := size / 2
//	//typ := Bubble
//	typ := Bitonic
//	//typ := OddEven
//
//	sorter1 := CreateCardinalityNetwork(size, k, AtMost, typ)
//	sorter2 := CreateCardinalityNetwork(size, k+1, AtLeast, typ)
//	sorter1.RemoveOutput()
//	sorter2.RemoveOutput()
//
//	input := make([]Literal, size)
//	for i, _ := range input {
//		input[i] = Literal{true, Atom{Pred("Input"), i, 0}}
//	}
//
//	lt := Pred("AtMost")
//	gt := Pred("AtLeast")
//
//	which := [8]bool{false, true, true, true, true, true, true, true}
//
//	// 3,4,5
//	which = [8]bool{false, false, false, true, true, true, false, false}
//	fmt.Println(which)
//	clauses := createEncoding(input, which, []Literal{}, "lt", lt, sorter1)
//
//	// 1,2,6
//	which = [8]bool{false, true, true, false, false, false, true, false}
//	clauses.AddClauseSet(createEncoding(input, which, []Literal{}, "gt", gt, sorter2))
//
//	printSorterTikZ(sorter2, "pic.tex")
//
//	g := IdGenerator(size * size)
//	g.GenerateIds(clauses)
//	g.filename = "test.cnf"
//	g.printClausesDIMACS(clauses)
//	//g.printDebug(clauses)
//}
//
