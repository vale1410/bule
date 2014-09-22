package sat

// test class

import (
//  "fmt"
    "github.com/vale1410/bule/sorters"
    "os"
    "strconv"
    "testing"
)

func TestWhichClauses(t *testing.T) {

    //sizes := []int{100,112,128,144,160,176}
    //sizes := []int{500, 750, 1000}
    sizes := []int{5, 10, 15}
    //sizes := []int{}
    typs := []sorters.SortingNetworkType{sorters.Bubble, sorters.Bitonic, sorters.OddEven, sorters.Pairwise}
    //typs := []sorters.SortingNetworkType{Bitonic, OddEven, Pairwise}
    //typs := []sorters.SortingNetworkType{sorters.OddEven, sorters.Pairwise}
    //typs := []sorters.SortingNetworkType{sorters.Pairwise}
    whichT := []int{1, 2, 3, 4}
    lt := Pred("AtMost")
    gt := Pred("AtLeast")

    for _, size := range sizes {
        for _, typ := range typs {
            for _, wh := range whichT {
                //k := int(0.05 * float64(size))
                k := size / 4
                //k := size - size/4
                sorter1 := sorters.CreateCardinalityNetwork(size, k, sorters.AtMost, typ)
                sorter2 := sorters.CreateCardinalityNetwork(size, k+1, sorters.AtLeast, typ)
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
                    input[i] = Literal{true, NewAtomP1(Pred("Input"), i)}
                }

                clauses := CreateEncoding(input, which1, []Literal{}, "lt", lt, sorter1)
                clauses.AddClauseSet(CreateEncoding(input, which2, []Literal{}, "gt", gt, sorter2))
                g := IdGenerator(size * size)
                g.GenerateIds(clauses)
                g.Filename = os.TempDir() + "/" + strconv.Itoa(size) + "_" + strconv.Itoa(k) + "_" + typ.String() + "_" + strconv.Itoa(wh) + ".cnf"
                g.PrintClausesDIMACS(clauses)
            }
        }
    }

}
