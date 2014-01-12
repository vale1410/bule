package pb

import (
	"fmt"
	"github.com/vale1410/bule/sorters"
	"math"
	"strconv"
	"testing"
)

func TestExample(test *testing.T) {

    //Example 1
	//t := createCardinality(8,4,1)
    //filename := "cardinality_8_4_1"

    //Example 2
	//t := createCardinality(8,16,4)
    //filename := "cardinality_8_16_4"

    //Example 3
	//t := createCardinality(8,12,3)
    //filename := "cardinality_8_12_3"

    //Example 4
	//t := createExample1()
    //filename := "example1"

    //Example 5
	//t := createJapan1(10)
    //filename := "japan1_10"

    //Example 6
	t := createJapan2(3)
    filename := "japan2_3"


	//typ := sorters.OddEven
	typ := sorters.OddEven

	t.Print10()
	t.Print2()

	t.CreateSortingEncoding(typ)

    fmt.Println("sorter size comparators", len(t.Sorter.Comparators))

	sorters.PrintSorterTikZ(t.Sorter, "tmp/"+filename+".tex")
}


//func TestJapan1(test *testing.T) {
//	t := createJapan1(16)
//
//	t.Print10()
//	t.Print2()
//
//	t.CreateBags()
//
//}
//
//func TestJapan2(test *testing.T) {
//	t := createJapan2(4)
//
//	t.Print10()
//	t.Print2()
//
//	t.CreateBags()
//}

//
//func TestBinary(test *testing.T) {
//	fmt.Println()
//	n := int64(10)
//	fmt.Println(binary(n))
//}
func createExample1() (t Threshold) {
	weights := []int64{11, 10, 6, 5}
	t.K = 25

	t.Desc = "Simple Test"
	t.Typ = AtMost
	t.Entries = make([]Entry, len(weights))

	for i := 0; i < len(weights); i++ {
		t.Entries[i] = Entry{Literal{true, Atom(i)}, weights[i]}
	}
	return
}

func createCardinality(n int,k int64, weight int64) (t Threshold) {

	t.K = k

	t.Desc = "Cardinality Test"
	t.Typ = AtMost
	t.Entries = make([]Entry, n)

	for i := 0; i < n; i++ {
		t.Entries[i] = Entry{Literal{true, Atom(i)}, weight}
	}

	return
}

func createSimple(n int) (t Threshold) {

	t.K = 2 * int64(n)

	t.Desc = "Simple Test"
	t.Typ = AtMost
	t.Entries = make([]Entry, n)

	for i := 0; i < n; i++ {
		x := int64(5)
		t.Entries[i] = Entry{Literal{true, Atom(i)}, x}
	}
	return
}

func createIgnasi1() (t Threshold) {

	weights := []int64{4, 3, 1, 1, 1, 1}
	t.K = 5

	t.Desc = "Ignasi 1"
	t.Typ = AtMost
	t.Entries = make([]Entry, len(weights))

	for i := 0; i < len(weights); i++ {
		t.Entries[i] = Entry{Literal{true, Atom(i)}, weights[i]}
	}
	return
}

func createIgnasi2() (t Threshold) {

	weights := []int64{7, 6, 2, 2, 2, 2, 1, 1, 1, 1, 1}
	t.K = 12

	t.Desc = "Ignasi 2"
	t.Typ = AtMost
	t.Entries = make([]Entry, len(weights))

	for i := 0; i < len(weights); i++ {
		t.Entries[i] = Entry{Literal{true, Atom(i)}, weights[i]}
	}
	return
}

// creats a Threshold function of size n (n must be even)
// given in "Size of OBDDs representing threshold functions"
func createJapan1(n int) (t Threshold) {
	t.Desc = "Japan 1 size " + strconv.Itoa(n)
	t.Typ = AtLeast
	t.Entries = make([]Entry, n)

	x := int64(1)

	for i := 0; i < n/2; i++ {
		e := n - i - 1
		t.K += x
		t.Entries[e] = Entry{Literal{true, Atom(e)}, x}
		x = x * 2
	}

	y := x / 2

	for i := n / 2; i < n; i++ {
		e := n - i - 1
		t.K += x - y
		t.Entries[e] = Entry{Literal{true, Atom(e)}, x - y}
		y = y / 2
	}

	t.K = t.K / 2

	return
}

// creates a Threshold function of size n=k*k
// given in "Size of OBDDs representing threshold functions"
func createJapan2(k int) (t Threshold) {

	n := k * k
	t.Desc = "Japan 2 size " + strconv.Itoa(n)
	t.Typ = AtLeast
	t.Entries = make([]Entry, n)

	x := int64(1)
	y := int64(math.Exp2(float64(k)))

	fmt.Println(y)

	for i := 0; i < k; i++ {
		//y := int64(math.Exp2(float64(k - 1)))
		for j := 0; j < k; j++ {
			e := n - i*k - j - 1
			//e := i*k + j
			t.K += x + y
			t.Entries[e] = Entry{Literal{true, Atom(e)}, x + y}
			x = x * 2
		}
		y = y * 2
		x = int64(1)
	}

	t.K = t.K / 2

	return
}
