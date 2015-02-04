package threshold

import (
	"fmt"
	"github.com/vale1410/bule/sorters"
	"os"
	"sort"
	"strconv"
)

type pair struct {
	A, B     int
	idA, idB int
}

type pairSlice []pair
type layerMap map[pair]int

func (l pairSlice) Len() int      { return len(l) }
func (l pairSlice) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l pairSlice) Less(i, j int) bool {

	if l[i].B-l[i].A > l[j].B-l[j].A {
		return true
	} else if l[i].B-l[i].A == l[j].B-l[j].A {
		return l[i].A < l[j].A
	} else {
		return false
	}
}

//\usepackage{float}
//\floatstyle{boxed}
//\restylefloat{figure}

func writeProlog(file *os.File) {
	file.Write([]byte(fmt.Sprintln(`
\documentclass{article}

\usepackage[latin1]{inputenc}
\usepackage{tikz}
\usetikzlibrary{shapes,arrows}
\usetikzlibrary{decorations.pathreplacing}
\begin{document}
\pagestyle{empty}`)))
}

func writeEpilog(file *os.File) {
	file.Write([]byte(fmt.Sprintln(`
\end{document}`)))
}

func PrintThresholdTikZ(filename string, ts []Threshold) {

	file, ok := os.Create(filename)
	if ok != nil {
		panic("Can open file to write.")
	}
	defer file.Close()

	writeProlog(file)

	for i, t := range ts {
		file.Write([]byte(fmt.Sprintf("\\section{Example %v}", i+1)))
		t.writeDescription(file)
		t.writeTikz(file)
	}

	writeEpilog(file)
}

// groups contains the comparators for each depth
// layers is a map for layering the comparators in each
// group such they dont overlap
func prepareData(sorter sorters.Sorter) (groups []pairSlice, layers []layerMap) {

	depth := make(map[int]int, len(sorter.Comparators))
	lines := make(map[int]int, len(sorter.Comparators))

	for i, x := range sorter.In {
		depth[x] = 0
		lines[x] = i
	}

	maxDepth := 0

	groups = make([]pairSlice, 0, len(sorter.Comparators))

	for _, x := range sorter.Comparators {
		if max, ok := depth[x.A]; ok {
			if depth[x.B] > max {
				max = depth[x.B]
			}
			max = max + 1
			depth[x.C] = max
			depth[x.D] = max
			lines[x.C] = lines[x.A]
			lines[x.D] = lines[x.B]
			if max > maxDepth {
				maxDepth = max
				group := make(pairSlice, 0, len(sorter.In))
				groups = append(groups, group)
			}

			p := pair{lines[x.A], lines[x.B], x.C, x.D}

			if p.A >= p.B {

				fmt.Println("something is wrong with comparator", x)
				fmt.Println("something is and pair", p)
			}

			groups[max-1] = append(groups[max-1], p)
		} else {
			panic("depth map is missing comparator")
		}
	}

	layers = make([]layerMap, 0, maxDepth)

	for _, group := range groups {
		sort.Sort(group)

		layer := make(layerMap, len(group))

		l := 0

		for len(layer) < len(group) {

			used := make([]bool, len(sorter.In))

			for _, p := range group {

				if _, ok := layer[p]; !ok {

					fits := true

					for i := p.A; i <= p.B; i++ {
						if used[i] {
							fits = false
						}
					}
					if fits {
						layer[p] = l
						for i := p.A; i <= p.B; i++ {
							used[i] = true
						}
					}
				}
			}

			l++
		}
		layers = append(layers, layer)
		//fmt.Println(group, layer)
	}
	return
}

func (t *Threshold) writeTikz(file *os.File) {

	sorter := t.Sorter

	groups, layers := prepareData(sorter)

	showIds := false

	//lets start drawing it :-)

	layerDist := 0.3
	groupDist := 1.0
	lineDist := 1.0

	symbolsTex := make(map[int]string, 3)
	symbolsTex[-1] = "\\ast"
	symbolsTex[0] = "0"
	symbolsTex[1] = "1"

	//file.Write([]byte("\\begin{figure}[h!]"))
	file.Write([]byte(fmt.Sprintln(`
\tikzset{cross/.style =
    {inner sep=0pt,minimum size=3pt,fill,circle}}
\centering`)))
	file.Write([]byte(fmt.Sprintln(`\resizebox {\columnwidth} {!} {`)))
	file.Write([]byte(fmt.Sprintln(`\begin{tikzpicture}[node distance=1cm, auto]`)))

	length := 0.0

	maxLayerDist := 0

	lineLength := make([]float64, len(sorter.In))

	for i, group := range groups {

		layer := layers[i]

		for _, comp := range group {

			if layer[comp] > maxLayerDist {
				maxLayerDist = layer[comp]
			}

			d := length + float64(layer[comp])*layerDist
			A := float64(comp.A) * lineDist
			B := float64(comp.B) * lineDist
			s1 := "     \\draw[thick] (%v,%v) to (%v,%v);\n"
			s2 := "     \\node[cross] at (%v,%v) {};\n"
			file.Write([]byte(fmt.Sprintf(s1, d, A, d, B)))
			file.Write([]byte(fmt.Sprintf(s2, d, A)))
			file.Write([]byte(fmt.Sprintf(s2, d, B)))

			s3 := "     \\node at (%v,%v) {$%v$};\n"

			if comp.idA < 2 {
				lineLength[comp.A] = d + layerDist
				file.Write([]byte(fmt.Sprintf(s3, d+2*layerDist, A, symbolsTex[comp.idA])))
			}

			if comp.idB < 2 {
				lineLength[comp.B] = d + layerDist
				file.Write([]byte(fmt.Sprintf(s3, d+2*layerDist, B, symbolsTex[comp.idB])))
			}

			if showIds {
				//debug
				file.Write([]byte(fmt.Sprintf(s3, d+layerDist, A+layerDist, comp.idA)))
				file.Write([]byte(fmt.Sprintf(s3, d+layerDist, B+layerDist, comp.idB)))
			}

		}

		length += float64(maxLayerDist)*layerDist + groupDist
		maxLayerDist = 0
	}

	for i, _ := range sorter.In {
		s1 := "    \\draw[thick] (%v,%v) to (%v,%v);\n"
		hight := float64(i) * lineDist
		var d float64
		if lineLength[i] > 0.0 {
			d = lineLength[i]
		} else {
			d = length - groupDist + layerDist
		}
		file.Write([]byte(fmt.Sprintf(s1, -layerDist, hight, d, hight)))
	}

	//if showIds {

	// i is level
	pos := 0

	for i, bag := range t.Bags {
		start := pos

		col1 := -3 * layerDist
		col2 := -2 * layerDist

		for _, lit := range bag {
			s := "     \\node at (%v,%v) {$%v$};\n"
			file.Write([]byte(fmt.Sprintf(s, col2, pos, lit.ToTex())))
			pos++
		}

		if len(bag) > 0 {
			s1 := "\\draw[thick,decorate,decoration={brace,amplitude=5pt},xshift=-4pt,yshift=0pt] (%v,%v) -- (%v,%v) node [black,midway,xshift=-5pt] {$2^%v$};"
			file.Write([]byte(fmt.Sprintf(s1, col1, start, col1, pos-1, i)))
		}

	}
	//}

	file.Write([]byte(fmt.Sprintln("\\end{tikzpicture}")))
	file.Write([]byte(fmt.Sprintln("}")))
	//  file.Write([]byte(fmt.Sprintln("\\end{figure}")))

}

func BinaryStr(n int64) (s string) {
	bin := binary(n)
	for i := len(bin) - 1; i >= 0; i-- {
		s += strconv.Itoa(bin[i])
	}
	return
}

func (t *Threshold) Print2() {
	fmt.Println(t.Desc)

	first := true
	for _, x := range t.Entries {
		l := x.Literal
		if !first {
			fmt.Printf("+ ")
		}
		first = false

		fmt.Print(BinaryStr(x.Weight))
		fmt.Print(x.Weight)
		fmt.Print(l.ToTex())
	}
	switch t.Typ {
	case AtMost:
		fmt.Print(" <= ")
	case AtLeast:
		fmt.Print(" >= ")
	case Equal:
		fmt.Print(" == ")
	}

	fmt.Print(BinaryStr(t.K))

	fmt.Println()
	fmt.Println()
}

func (t *Threshold) writeDescription(file *os.File) {
	//file.Write([]byte("\\caption{"))
	file.Write([]byte("The translation of the PB through sorters:\n"))
	t.WriteFormula(10, file)
	file.Write([]byte("and in binary representation:\n"))
	t.WriteFormula(2, file)
	file.Write([]byte(fmt.Sprintf("adding the Tare on both sides. $$T = %v_{10} = %v_2$$\n", t.Tare, BinaryStr(t.Tare))))
	//file.Write([]byte(fmt.Sprintf("}\n")))
}

func (t *Threshold) WriteFormula(base int, file *os.File) {

	var w string

	file.Write([]byte("$$"))
	for _, x := range t.Entries {
		if base == 2 {
			w = BinaryStr(x.Weight)
		} else if base == 10 {
			w = fmt.Sprintf("%v", x.Weight)
		} else {
			panic("only base 2 and 10 supported")
		}
		if x.Weight < 0 {
			file.Write([]byte(fmt.Sprintf("%v \\cdot %v ", w, x.Literal.ToTex())))
		} else {
			file.Write([]byte(fmt.Sprintf("+%v \\cdot %v ", w, x.Literal.ToTex())))
		}
	}
	switch t.Typ {
	case AtMost:
		file.Write([]byte(" \\leq "))
	case AtLeast:
		file.Write([]byte(" \\geq "))
	case Equal:
		file.Write([]byte(" = "))
	}
	if base == 2 {
		w = BinaryStr(t.K)
	} else if base == 10 {
		w = fmt.Sprintf("%v", t.K)
	} else {
		panic("only base 2 and 10 supported")
	}
	file.Write([]byte(fmt.Sprintf("%v $$\n", w)))
}

func (t *Threshold) PrintGurobi() {
	for _, x := range t.Entries {
		l := x.Literal

		if x.Weight > 0 {
			fmt.Printf(" +")
		}

		if x.Weight != 1 {
			fmt.Printf(" ")
			fmt.Print(x.Weight)
		}
		fmt.Print(l.ToTxt())
	}
	switch t.Typ {
	case AtMost:
		fmt.Print(" <= ")
	case AtLeast:
		fmt.Print(" >= ")
	case Equal:
		fmt.Print(" = ")
	}
	fmt.Println(t.K)

}

func (t *Threshold) Print10() {
	for _, x := range t.Entries {
		l := x.Literal

		if x.Weight > 0 {
			fmt.Printf("+")
		}

		fmt.Print(x.Weight)
		fmt.Print(l.ToTex())
	}
	switch t.Typ {
	case AtMost:
		fmt.Print("<= ")
	case AtLeast:
		fmt.Print(">= ")
	case Equal:
		fmt.Print("== ")
	}
	fmt.Println(t.K, ";")

}

func (t *Threshold) PrintGringo() {

	if len(t.Entries) > 0 {

		switch t.Typ {
		case AtMost:
			fmt.Print(":- ", t.K+1, " [ ")
		case AtLeast:
			fmt.Print(":- [ ")
		case Equal:
			fmt.Print(":- not ", t.K, " [ ")
		case Optimization:
			fmt.Print("#minimize[")
		}

		for i, x := range t.Entries {
			if i != 0 {
				fmt.Print(" , ")
			}
			if x.Weight != 1 {
				fmt.Print(x.Literal.ToTxt(), "=", x.Weight)
			} else {
				fmt.Print(x.Literal.ToTxt())
			}
		}

		switch t.Typ {
		case AtMost:
			fmt.Print(" ]")
		case AtLeast:
			fmt.Print(" ] ", t.K-1)
		case Equal:
			fmt.Print(" ] ", t.K)
		case Optimization:
			fmt.Print("]")
		}
		fmt.Println(".")
	}

}
