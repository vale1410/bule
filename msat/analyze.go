package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var init_path = flag.String("f", "", "Path to result folder.")
var out_path = flag.String("o", "/tmp/out", "Path to output folder.")

const (
	SAT = iota
	UNSAT
	TO  // Timeout
	MEM // Memout
	ERR // Some other error
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type record struct {
	name   string
	values []Value
}

func (v Value) TimeString() string {
	if v.typ == SAT {
		return fmt.Sprintf("%.2f(1)", v.time)
	} else if v.typ == UNSAT {
		return fmt.Sprintf("%.2f(0)", v.time)
	} else if v.typ == TO {
		return "TO"
	} else if v.typ == MEM {
		return "MEM"
	} else {
		fmt.Println(v)
		panic("Could not print value! ")
	}
}

type Value struct {
	time float64
	//	conflicts int
	typ int
}

type Record struct {
	name   string
	values []Value
}

type Frame struct {
	// per encoding
	order []int // mapping to print in correct order, must be permutation!
	// per encoding
	records []Record
	// names of the instances
	insts []string

	// computed statistics for time
	// per encoding
	solved      []int
	solvedSAT   []int
	solvedUNSAT []int
	timeouts    []int
	avg         []float64
	//	std         []float64
	//hist        []float64
	//	solvedAll   []int     // value only for rows where ALL methods solved
	//	avgAll      []float64 // value only for rows where ALL methods solved
	//	stdAll      []float64 // value only for rows where ALL methods solved
}

type By func(p1, p2 *Value) bool

func (by By) Sort(values []Value) {
	rds := &valueSorter{
		values: values,
		by:     by,
	}
	sort.Sort(rds)
}

type valueSorter struct {
	values []Value
	by     func(p1, p2 *Value) bool
}

func (s *valueSorter) Len() int {
	return len(s.values)
}

func (s *valueSorter) Swap(i, j int) {
	s.values[i], s.values[j] = s.values[j], s.values[i]
}

func (s *valueSorter) Less(i, j int) bool {
	return s.by(&s.values[i], &s.values[j])
}

func main() {

	flag.Parse()

	var frame Frame
	//	var instances []string

	if *init_path == "" || *out_path == "" {
		fmt.Println("Please run with -f <input folder> -o <output folder>")
		return
	}

	folders, err := ioutil.ReadDir(*init_path)
	check(err)

	first := true
	for _, enc := range folders {
		//		fmt.Println("scanning encoding", enc.Name())
		rd := Record{name: enc.Name()}

		path := filepath.Join(*init_path, enc.Name())
		{ // check that there are only directories here
			tmp, err := os.Open(path)
			check(err)
			stat, err := tmp.Stat()
			if !stat.IsDir() {
				fmt.Println("path is not folder for encodings ", path, "\nabort!")
				os.Exit(0)
			}
		}

		res, _ := ioutil.ReadDir(path)

		rd.values = make([]Value, len(res))

		for i, r := range res {
			// parse file and write record!
			file, err := os.Open(filepath.Join(path, r.Name()))
			if first {
				frame.insts = append(frame.insts, r.Name())
			} else {
				if len(frame.insts) <= i || frame.insts[i] != r.Name() {
					fmt.Println(frame.insts)
					fmt.Println("wrong structure ", path, r.Name())
					panic("wrong structure")
				}
			}
			check(err)

			scanner := bufio.NewScanner(file)

			var v Value
			v.typ = SAT

			for scanner.Scan() {

				if scanner.Text() == "Time limit exceeded!" {
					v.typ = TO
					continue
				}

				if strings.Contains(scanner.Text(), "UNSATISFIABLE") {
					v.typ = UNSAT
					continue
				}

				if scanner.Text() == "==========" || strings.Contains(scanner.Text(), "SATISFIABLE") {
					v.typ = SAT
					continue
				}

				fields := strings.FieldsFunc(scanner.Text(), func(r rune) bool { return r == ',' })

				if len(fields) > 1 {

					switch len(fields) {

					case 9:
						time, err := strconv.ParseFloat(fields[8], 64)
						check(err)
						v.time = time

					case 10:
						time, err := strconv.ParseFloat(fields[9], 64)
						check(err)
						v.time = time

					default:
						fmt.Println("problems parsing", fields)
					}

					//					{
					//						conflicts, err := strconv.Atoi(fields[3])
					//						check(err)
					//						v.conflicts = conflicts
					//					}
				}
			}

			rd.values[i] = v
			file.Close()
		}

		frame.records = append(frame.records, rd)
		first = false
	}

	{ /// print.config  if exists

		frame.order = make([]int, len(frame.records))

		i := 0

		{
			file, err := os.Open("print.config")
			if err == nil {
				defer file.Close()
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					for j, r := range frame.records {
						if scanner.Text() == r.name {
							frame.order[i] = j
							i++
							break
						}

					}
				}
			}
		}

		for ; i < len(frame.order); i++ {
			frame.order[i] = i
		}
	}

	writeTable(frame)
	writeHist(frame)
	writeComputeStatistics(&frame)
	writePlot(frame, false)
	writePlot(frame, true)
	writeTex(frame)
	//writeTable(frame)

}

func writeTex(frame Frame) {
	name := "all.tex"
	f, err := os.Create(filepath.Join(*out_path, name))
	check(err)
	defer f.Close()
	_, err = f.WriteString(`
	\documentclass{article}
	\usepackage{longtable}
	\usepackage{pgfplots}
	\usepackage{pgfplotstable,booktabs}
	\usepackage{tikzscale}

	\begin{document}

	\begin{figure}
%	\includegraphics{plot.tikz}
	\includegraphics{plot_log.tikz}
	\centering
	\include{statistics}%
	\end{figure}
	\include{table}
	\includegraphics{hist.tikz}

	\end{document}

	`)

}

func writeTable(frame Frame) {
	name := "table.tex"
	f, err := os.Create(filepath.Join(*out_path, name))
	check(err)
	defer f.Close()

	f.WriteString("\\begin{longtable}{ ll")
	for _, _ = range frame.order {
		f.WriteString(fmt.Sprintf("r"))
	}
	f.WriteString("}\n\\toprule\n")

	f.WriteString("id & name ")
	//f.WriteString("id ")
	for _, r := range frame.order {
		f.WriteString(fmt.Sprintf("& %v  ", frame.records[r].name))
	}
	f.WriteString("\\\\\n")
	f.WriteString("\\midrule\n")

	for i, id := range frame.insts {
		if len(id) >= 10 {
			f.WriteString(fmt.Sprintf(" %v & \\verb|%v| ", i, id[:10]))
		} else {
			f.WriteString(fmt.Sprintf(" %v & \\verb|%v| ", i, id))
		}

		//f.WriteString(fmt.Sprintf(" %v ", i))
		for _, r := range frame.order {
			f.WriteString(fmt.Sprintf(" & %v ", frame.records[r].values[i].TimeString()))
		}
		f.WriteString("\\\\\n")
	}
	f.WriteString("\\bottomrule\n")
	f.WriteString("\\end{longtable}\n")
}

func writeComputeStatistics(frame *Frame) {
	name := "statistics.tex"
	f, err := os.Create(filepath.Join(*out_path, name))
	check(err)
	defer f.Close()

	frame.solved = make([]int, len(frame.records))
	frame.solvedSAT = make([]int, len(frame.records))
	frame.solvedUNSAT = make([]int, len(frame.records))
	frame.timeouts = make([]int, len(frame.records))
	frame.avg = make([]float64, len(frame.records))
	//frame.hist = make([]int, 0, len(frame.order)*len(frame.records))
	//  frame.std = make([]float64, len(frame.records))
	//	frame.solvedAll = make([]int, len(frame.records))
	//	frame.avgAll = make([]float64, len(frame.records))
	//	frame.stdAll = make([]float64, len(frame.records))

	allSolved := make([]bool, len(frame.insts))

	for i, _ := range frame.insts {
		allSolved[i] = true
		for r, record := range frame.records {
			switch record.values[i].typ {
			case SAT:
				{
					frame.solved[r]++
					frame.solvedSAT[r]++
					frame.avg[r] += record.values[i].time
					//		frame.hist = append(frame.hist, record.values[i].time)
				}
			case UNSAT:
				{
					frame.solved[r]++
					frame.solvedUNSAT[r]++
					frame.avg[r] += record.values[i].time
					//		frame.hist = append(frame.hist, record.values[i].time)
				}
			case TO:
				{
					frame.timeouts[r]++
				}
			default:
				{
					fmt.Println("something is wrong in aggregation of values for statistics:\n", record.values[i])
					os.Exit(1)
				}

			}
			if record.values[i].typ != SAT && record.values[i].typ != UNSAT {
				allSolved[i] = false
			}
		}
	}
	for r, _ := range frame.records {
		frame.avg[r] = frame.avg[r] / float64(frame.solved[r])
	}

	//	for i, _ := range frame.insts {
	//		for r, record := range frame.records {
	//			if record.values[i].typ == SAT || record.values[i].typ == UNSAT {
	//				frame.std[r] +=
	//					(frame.avg[r] - record.values[i].time) * (frame.avg[r] - record.values[i].time)
	//			}
	//		}
	//	}
	//	for r, _ := range frame.records {
	//		frame.std[r] = frame.std[r] / float64(frame.solved[r])
	//		frame.std[r] = math.Sqrt(frame.std[r])
	//	}

	//	for i, _ := range frame.insts {
	//		if allSolved[i] {
	//			for r, record := range frame.records {
	//				frame.solvedAll[r]++
	//				frame.avgAll[r] += record.values[i].time
	//			}
	//		}
	//	}
	//	for r, _ := range frame.records {
	//		frame.avgAll[r] = frame.avgAll[r] / float64(frame.solvedAll[r])
	//	}
	//	for i, _ := range frame.insts {
	//		if allSolved[i] {
	//			for r, record := range frame.records {
	//				frame.stdAll[r] +=
	//					(frame.avgAll[r] - record.values[i].time) * (frame.avgAll[r] - record.values[i].time)
	//			}
	//		}
	//	}
	//	for r, _ := range frame.records {
	//		frame.stdAll[r] = frame.stdAll[r] / float64(frame.solvedAll[r])
	//		frame.stdAll[r] = math.Sqrt(frame.stdAll[r])
	//	}
	allSAT := true
	allUNSAT := true

	for _, r := range frame.order {
		if frame.solved[r] != frame.solvedSAT[r] {
			allSAT = false
		}
	}

	for _, r := range frame.order {
		if frame.solved[r] != frame.solvedUNSAT[r] {
			allUNSAT = false
		}
	}

	{ // Write to File; Observe that now we use frame.order

		f.WriteString("\\begin{tabular}{l")
		for _, _ = range frame.order {
			f.WriteString(fmt.Sprintf("r"))
		}
		f.WriteString("}\n\\toprule\n")

		f.WriteString(" ")
		for _, r := range frame.order {
			f.WriteString(fmt.Sprintf(" & %v ", frame.records[r].name))
		}
		f.WriteString("\\\\\n \\midrule\n")

		if !allUNSAT && !allSAT {
			f.WriteString("solved ")
			for _, r := range frame.order {
				f.WriteString(fmt.Sprint(" & ", frame.solved[r], " "))
			}
			f.WriteString("\\\\\n")
		}

		if !allUNSAT {
			f.WriteString("SAT ")
			for _, r := range frame.order {
				f.WriteString(fmt.Sprint(" & ", frame.solvedSAT[r], " "))
			}
			f.WriteString("\\\\\n")
		}

		if !allSAT {
			f.WriteString("UNSAT ")
			for _, r := range frame.order {
				f.WriteString(fmt.Sprint(" & ", frame.solvedUNSAT[r], " "))
			}
			f.WriteString("\\\\\n")
		}

		f.WriteString("TOs")
		for _, r := range frame.order {
			f.WriteString(fmt.Sprint(" & ", frame.timeouts[r], " "))
		}
		f.WriteString("\\\\\n")

		f.WriteString("avg ")
		for _, r := range frame.order {
			f.WriteString(fmt.Sprintf(" & %.2f ", frame.avg[r]))
		}
		f.WriteString("\\\\\n")

		//		f.WriteString("\\midrule\n")

		//		f.WriteString("std ")
		//		for _, r := range frame.order {
		//			f.WriteString(fmt.Sprintf(" & %.2f ", frame.std[r]))
		//		}
		//		f.WriteString("\\\\\n")

		//		f.WriteString("solvedA ")
		//		for _, r := range frame.order {
		//			f.WriteString(fmt.Sprint(" & ", frame.solvedAll[r], " "))
		//		}
		//		f.WriteString("\\\\\n")
		//
		//		f.WriteString("avgA ")
		//		for _, r := range frame.order {
		//			f.WriteString(fmt.Sprintf(" & %.2f ", frame.avgAll[r]))
		//		}
		//		f.WriteString("\\\\\n")
		//
		//		f.WriteString("stdA ")
		//		for _, r := range frame.order {
		//			f.WriteString(fmt.Sprintf(" & %.2f ", frame.stdAll[r]))
		//		}
		//		f.WriteString("\\\\ \n")

		f.WriteString("\\bottomrule\n")
		f.WriteString("\\end{tabular}\n")
	}
}

func writeHist(frame Frame) {
	name := "hist.tikz"
	f, err := os.Create(filepath.Join(*out_path, name))
	check(err)
	defer f.Close()
	_, err = f.WriteString(`
\begin{tikzpicture}
\begin{axis}[
  ybar interval,
  xtick=,% reset from ybar interval
  xticklabel= \pgfmathprintnumber\tick--\pgfmathprintnumber\nexttick
]
\addplot+[hist={bins=5}]
table[y index=0] {
data
`)
	for i, _ := range frame.insts {
		for _, record := range frame.records {
			if record.values[i].typ == SAT || record.values[i].typ == UNSAT {
				_, err = f.WriteString(fmt.Sprintf("%.2f \n", record.values[i].time))
			}
		}
	}
	_, err = f.WriteString(`};
\end{axis}
\end{tikzpicture}
`)
}

func writePlot(frame Frame, log bool) {

	var name string
	if log {
		name = "plot_log.tikz"
	} else {
		name = "plot.tikz"
	}
	f, err := os.Create(filepath.Join(*out_path, name))
	check(err)
	defer f.Close()

	_, err = f.WriteString(`
\begin{tikzpicture}

\begin{axis}[
	%legend pos=outer north east,
	legend pos=north west,
	xlabel=Solved Instances,
	ylabel=Time in sec,
`)

	if log {
		_, err = f.WriteString("ymode=log\n")
		check(err)
	}
	_, err = f.WriteString("]\n")
	check(err)

	timeComparer := func(v1, v2 *Value) bool {
		if v1.typ == TO {
			return false
		} else if v2.typ == TO {
			return true
		} else {
			return v1.time < v2.time
		}
	}

	//	for r, record := range frame.records {
	for _, r := range frame.order {
		var record Record
		record = frame.records[r]

		By(timeComparer).Sort(record.values)

		_, err = f.WriteString("\\addplot+[")
		check(err)

		_, err = f.WriteString(fmt.Sprint("mark indices={1,", frame.solved[r]/2, ",", frame.solved[r], "}"))
		check(err)

		_, err = f.WriteString("] coordinates {\n")
		check(err)

		for i, v := range record.values {
			//		if v.time > 2.0 {
			if v.typ == SAT || v.typ == UNSAT {
				_, err = f.WriteString(fmt.Sprintf("(%v,%v)\n", i+1, v.time))
				check(err)
			}
			//		}
		}
		_, err = f.WriteString(fmt.Sprintf("};\n\\addlegendentry{%v}\n", record.name))
		check(err)
	}

	_, err = f.WriteString(`

\end{axis}
\end{tikzpicture}
`)
	check(err)
}
