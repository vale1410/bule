package pb

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func TestExample(test *testing.T) {

	weights := []int64{11, 10, 6, 5}

	var t Threshold
	t.Desc = "Simple Test"
	t.K = 25
	t.Typ = AtMost
	t.Entries = make([]Entry, len(weights))

	for i, x := range weights {
		t.Entries[i] = Entry{Literal{true, Atom(i)}, x}
	}

	t.Print10()
	t.Print2()

}

func TestJapan1(test *testing.T) {
	t := createJapan1(8)

	t.Print10()
	t.Print2()

}

func TestJapan2(test *testing.T) {
	t := createJapan2(4)

	t.Print10()
	t.Print2()

}

func TestBinary(test *testing.T) {
	fmt.Println()
	n := int64(10)
	fmt.Println(binary(n))
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
