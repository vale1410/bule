package sorting_network

import (
	"github.com/vale1410/bule/constraints"
	"testing"
)

func TestExample(test *testing.T) {

	filename := "test"
	//typ := sorters.Bubble
	//typ := sorters.Pairwise
	//typ := sorters.Bitonic
	//typ := sorters.OddEven

	//Example 1
	//t := createCardinality(4, 15, 5)
	//t := createCardinality(2, 1, 1)
	//t := createCardinality(9, 1, 1)

	//Example 2
	//t := createCardinality(8, 4, 1)
	//t := createCardinality(8,8,2)
	//t := createCardinality(8,16,4)
	//filename := "cardinality_8_16_4"

	//Example 3
	//t := createCardinality(8,12,3)
	//filename := "cardinality_8_12_3"

	//Example 4
	//t := createExample1()
	//filename := "example1"

	//Example 5
	//t := createJapan1(80)
	//filename := "japan1_10"

	//Example 6
	//t := createJapan2(3)
	//filename := "japan2_3"

	t1 := createIgnasi1()
	t2 := createIgnasi2()

	s1 := NewSortingNetwork(t1)
	s2 := NewSortingNetwork(t2)

	//fmt.Println(t)

	PrintThresholdTikZ(filename+".tex", []SortingNetwork{s1, s2})

}

func createIgnasi1() (t constraints.Threshold) {

	weights := []int64{4, 3, 1, 1, 1, 1}
	return constraints.CreatePB(weights, 5)
}

func createIgnasi2() (t constraints.Threshold) {

	weights := []int64{7, 6, 2, 2, 2, 2, 1, 1, 1, 1, 1}
	return constraints.CreatePB(weights, 5)
}

func TestPBOGeneration(test *testing.T) {

	//t := createJapan1(80)
	//t := createJapan2(16)

	//fmt.Printf("* #variable= %v #constraint= %v\n", len(t.Entries), 2)
	//fmt.Println("****************************************")
	//fmt.Println("* begin normalizer comments")
	//fmt.Println("* category= SAT/UNSAT-BIGINT")
	//fmt.Println("* end normalizer comments")
	//fmt.Println("****************************************")

	//t.Print10()

	//if t.Typ == AtMost {
	//  t.Typ = AtLeast
	//    //t.K++
	//} else if t.Typ == AtLeast {
	//  t.Typ = AtMost
	//    //t.K--
	//}
	//t.Print10()
}

//func TestSimple1(test *testing.T) {
//
//  t := createCardinality(2, 1, 1)
//  typ := sorters.OddEven
//  t.CreateSortingEncoding(typ)
//
//  if t.Sorter.Out[0] != 4 ||
//      len(t.Sorter.In) != 2 ||
//      len(t.Sorter.Comparators) != 1 {
//      test.Error("")
//  }
//}

//func TestJapan1(test *testing.T) {
//  t := createJapan1(16)
//
//  t.Print10()
//  t.Print2()
//
//  t.CreateBags()
//
//}
//
//func TestJapan2(test *testing.T) {
//  t := createJapan2(4)
//
//  t.Print10()
//  t.Print2()
//
//  t.CreateBags()
//}

//
//func TestBinary(test *testing.T) {
//  fmt.Println()
//  n := int64(10)
//  fmt.Println(binary(n))
//}
