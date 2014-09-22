package sorters

import (
    "fmt"
    "os"
    "os/exec"
    "sort"
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

func PrintSorterTikZ(sorter Sorter, filename string) {

    // group

    fmt.Println("test sorters print")

    depth := make(map[int]int, len(sorter.Comparators))
    lines := make(map[int]int, len(sorter.Comparators))

    for i, x := range sorter.In {
        depth[x] = 0
        lines[x] = i
    }

    maxDepth := 0

    groups := make([]pairSlice, 0, len(sorter.Comparators))

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
                fmt.Println("something is wrong", p)
            }

            groups[max-1] = append(groups[max-1], p)
        } else {
            panic("depth map is missing comparator")
        }
    }

    layers := make([]layerMap, 0, maxDepth)

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

    // groups contains the comparators for each depth
    // layers is a map for layering the comparators in each
    // group such they dont overlap

    //lets start drawing it :-)

    layerDist := 0.3
    groupDist := 1.0
    lineDist := 1.0

    file, ok := os.Create(filename)
    if ok != nil {
        panic("Can open file to write.")
    }

    file.Write([]byte(fmt.Sprintln(`
\documentclass{article}

\usepackage[latin1]{inputenc}
\usepackage{tikz}
\usetikzlibrary{shapes,arrows}
\begin{document}
\pagestyle{empty}
\tikzset{cross/.style = 
    {inner sep=0pt,minimum size=3pt,fill,circle}}
\centering 
\resizebox {\columnwidth} {!} {
\begin{tikzpicture}[node distance=1cm, auto]`)))

    length := 0.0

    maxLayerDist := 0

    showIds := true

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

            if showIds {
                s3 := "     \\node at (%v,%v) {%v};\n"
                file.Write([]byte(fmt.Sprintf(s3, d+layerDist, A+layerDist, comp.idA)))
                file.Write([]byte(fmt.Sprintf(s3, d+layerDist, B+layerDist, comp.idB)))
            }

        }

        length += float64(maxLayerDist)*layerDist + groupDist
        maxLayerDist = 0
    }

    for i, _ := range sorter.In {

        s1 := "    \\draw[thick] (%v,%v) to (%v,%v);\n"
        file.Write([]byte(fmt.Sprintf(s1, -layerDist, i, length-groupDist+layerDist, i)))
    }

    if showIds {
        for i, id := range sorter.In {
            s := "     \\node at (%v,%v) {%v};\n"
            file.Write([]byte(fmt.Sprintf(s, -2*layerDist, i, id)))
        }
        //for i, id := range sorter.Out {
        //  s := "\\node at (%v,%v) {%v};\n"
        //  file.Write([]byte(fmt.Sprintf(s, length-groupDist+layerDist+layerDist, i, id)))
        //}
    }

    file.Write([]byte(fmt.Sprintln(`
\end{tikzpicture}
}
\end{document}`)))
}

func printSorterDot(sorter Sorter, filename string) {

    file, ok := os.Create(filename)
    if ok != nil {
        panic("Can open file to write.")
    }
    file.Write([]byte(fmt.Sprintln("digraph {")))
    file.Write([]byte(fmt.Sprintln("  graph [rankdir = LR, splines=ortho];")))

    rank := "{rank=same; "
    for i := 0; i < len(sorter.Out); i++ {
        if sorter.Out[i] > 1 {
            rank += fmt.Sprintf(" t%v ", sorter.Out[i])
        }
    }
    rank += "}; "

    for i := 0; i < len(sorter.Out); i++ {
        file.Write([]byte(fmt.Sprintf("n%v -> t%v\n", sorter.In[i], sorter.In[i])))
    }

    file.Write([]byte(rank))
    rank = "{rank=same; "
    for i := 0; i < len(sorter.Out); i++ {
        rank += fmt.Sprintf(" t%v ", sorter.In[i])
    }
    rank += "}; "
    file.Write([]byte(rank))

    //var rank string
    for _, comp := range sorter.Comparators {
        rank = "{rank=same; "
        rank += fmt.Sprintf(" t%v t%v ", comp.A, comp.B)
        rank += "}; "
        file.Write([]byte(rank))
    }

    for _, comp := range sorter.Comparators {
        if comp.A > 1 && comp.B > 1 {
            //file.Write([]byte(fmt.Sprintf("t%v -> t%v [dir=none]\n", comp.A, comp.B)))
            file.Write([]byte(fmt.Sprintf("t%v -> t%v \n", comp.B, comp.A)))
        }
        if comp.C > 1 {
            file.Write([]byte(fmt.Sprintf("t%v -> t%v \n", comp.A, comp.C)))
        }
        if comp.D > 1 {

            file.Write([]byte(fmt.Sprintf("t%v -> t%v \n", comp.B, comp.D)))
        }
    }
    file.Write([]byte(fmt.Sprintln("}")))
    // run dot stuff
    dotPng := exec.Command("dot", "-Tpng", filename, "-O")
    _ = dotPng.Run()

    rmDot := exec.Command("rm", "-fr", filename)
    _ = rmDot.Run()
}
