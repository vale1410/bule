package sorting_network

import (
	"fmt"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/sorters"
	"math"
	"strconv"
	"testing"
)

func TestExample(test *testing.T) {

	filename := "test"
	//typ := sorters.Bubble
	typ := sorters.Pairwise
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

	t1.CreateSorter(typ)
	t2.CreateSorter(typ)

	//fmt.Println(t)

	PrintThresholdTikZ(filename+".tex", []Threshold{t1, t2})

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

func createEntries(weights []int64) (entries []Entry) {
	p := sat.Pred("x")
	entries = make([]Entry, len(weights))

	for i := 0; i < len(weights); i++ {
		l := sat.Literal{true, sat.NewAtomP1(p, i)}
		entries[i] = Entry{l, weights[i]}
	}
	return
}

func createExample1() (t Threshold) {
	weights := []int64{11, 10, 6, 5}
	t.K = 25

	t.Desc = "Simple Test"
	t.Typ = AtMost
	t.Entries = createEntries(weights)
	return
}

func createExample2() (t Threshold) {
	weights := []int64{-1, -1, -1, 1}
	t.K = 0

	t.Desc = "Simple Test"
	t.Typ = AtLeast
	t.Entries = createEntries(weights)
	return
}

func createCardinality(n int, k int64, weight int64) (t Threshold) {

	t.K = k

	t.Desc = "Cardinality Test"
	t.Typ = AtMost
	weights := make([]int64, n)
	for i := 0; i < n; i++ {
		weights[i] = weight
	}
	t.Entries = createEntries(weights)

	return
}

func createIgnasi1() (t Threshold) {

	weights := []int64{4, 3, 1, 1, 1, 1}
	t.K = 5

	t.Desc = "Ignasi 1"
	t.Typ = AtMost
	t.Entries = createEntries(weights)
	return
}

func createIgnasi2() (t Threshold) {

	weights := []int64{7, 6, 2, 2, 2, 2, 1, 1, 1, 1, 1}
	t.K = 12

	t.Desc = "Ignasi 2"
	t.Typ = AtMost
	t.Entries = createEntries(weights)
	return
}

// creats a Threshold function of size n (n must be even)
// given in "Size of OBDDs representing threshold functions"
func createJapan1(n int) (t Threshold) {

	t.Desc = "Japan 1 size " + strconv.Itoa(n)
	t.Typ = AtLeast
	//t.Entries = make([]Entry, n)
	weights := make([]int64, n)

	x := int64(1)

	for i := 0; i < n/2; i++ {
		e := n - i - 1
		t.K += x
		weights[e] = x
		//t.Entries[e] = Entry{Literal{true, Atom(e + 1)}, x}
		x = x * 2
	}

	y := x / 2

	for i := n / 2; i < n; i++ {
		e := n - i - 1
		t.K += x - y
		weights[e] = x - y
		//t.Entries[e] = Entry{Literal{true, Atom(e + 1)}, x - y}
		y = y / 2
	}

	t.Entries = createEntries(weights)
	t.K = t.K / 2

	return
}

// creates a Threshold function of size n=k*k
// given in "Size of OBDDs representing threshold functions"
func createJapan2(k int) (t Threshold) {

	n := k * k
	t.Desc = "Japan 2 size " + strconv.Itoa(n)
	t.Typ = AtLeast
	//t.Entries = make([]Entry, n)
	weights := make([]int64, n)

	x := int64(1)
	y := int64(math.Exp2(float64(k)))

	fmt.Println(y)

	for i := 0; i < k; i++ {
		for j := 0; j < k; j++ {
			e := n - i*k - j - 1
			t.K += x + y
			weights[e] = x + y
			//t.Entries[e] = Entry{Literal{true, Atom(e + 1)}, x + y}
			x = x * 2
		}
		y = y * 2
		x = int64(1)
	}

	t.Entries = createEntries(weights)
	t.K = t.K / 2

	return
}
