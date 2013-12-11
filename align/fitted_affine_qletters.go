// This file is automatically generated. Do not edit - make changes to relevant got file.

// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"

	"fmt"
	"os"
	"text/tabwriter"
)

//line fitted_affine_type.got:17
func drawFittedAffineTableQLetters(rSeq, qSeq alphabet.QLetters, index alphabet.Index, table [][3]int, a FittedAffine) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Printf("rSeq: %s\n", rSeq)
	fmt.Printf("qSeq: %s\n", qSeq)
	for l := 0; l < 3; l++ {
		fmt.Fprintf(tw, "%c\tqSeq\t", "MUL"[l])
		for _, l := range qSeq {
			fmt.Fprintf(tw, "%c\t", l)
		}
		fmt.Fprintln(tw)

		r, c := rSeq.Len()+1, qSeq.Len()+1
		fmt.Fprint(tw, "rSeq\t")
		for i := 0; i < r; i++ {
			if i != 0 {
				fmt.Fprintf(tw, "%c\t", rSeq[i-1].L)
			}

			for j := 0; j < c; j++ {
				p := pointerFittedAffineQLetters(rSeq, qSeq, i, j, l, table, index, a, c)
				var vi interface{}
				if vi = table[i*c+j][l]; vi == minInt {
					vi = "-Inf"
				}
				if p != "" {
					fmt.Fprintf(tw, "%s % 4v\t", p, vi)
				} else {
					fmt.Fprintf(tw, "%v\t", vi)
				}
			}
			fmt.Fprintln(tw)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

func pointerFittedAffineQLetters(rSeq, qSeq alphabet.QLetters, i, j, l int, table [][3]int, index alphabet.Index, a FittedAffine, c int) string {
	if i == 0 || j == 0 {
		return ""
	}
	rVal := index[rSeq[i-1].L]
	qVal := index[qSeq[j-1].L]
	if rVal < 0 || qVal < 0 {
		return ""
	}
	switch p := i*c + j; table[p][l] {
	case table[p-c][up] + a.Matrix[rVal][gap]:
		return "⬆ u"
	case table[p-1][left] + a.Matrix[gap][qVal]:
		return "⬅ l"

	case table[p-c][diag] + a.GapOpen + a.Matrix[rVal][gap]:
		return "⬆ m"
	case table[p-1][diag] + a.GapOpen + a.Matrix[gap][qVal]:
		return "⬅ m"

	case table[p-c-1][up] + a.Matrix[rVal][qVal]:
		return "⬉ u"
	case table[p-c-1][left] + a.Matrix[rVal][qVal]:
		return "⬉ l"
	case table[p-c-1][diag] + a.Matrix[rVal][qVal]:
		return "⬉ m"
	default:
		return [3]string{"", "⬆ u", "⬅ l"}[l]
	}
}

func (a FittedAffine) alignQLetters(rSeq, qSeq alphabet.QLetters, alpha alphabet.Alphabet) ([]feat.Pair, error) {
	let := len(a.Matrix)
	la := make([]int, 0, let*let)
	for _, row := range a.Matrix {
		if len(row) != let {
			return nil, ErrMatrixNotSquare
		}
		la = append(la, row...)
	}

	index := alpha.LetterIndex()
	r, c := rSeq.Len()+1, qSeq.Len()+1
	table := make([][3]int, r*c)
	table[0] = [3]int{
		diag: 0,
		up:   minInt,
		left: minInt,
	}
	table[1] = [3]int{
		diag: minInt,
		up:   minInt,
		left: a.GapOpen + la[index[qSeq[0].L]],
	}
	for j := range table[2:c] {
		table[j+2] = [3]int{
			diag: minInt,
			up:   minInt,
			left: table[j+1][left] + la[index[qSeq[j+1].L]],
		}
	}
	table[c] = [3]int{
		diag: minInt,
		up:   a.GapOpen + la[index[rSeq[0].L]*let],
		left: minInt,
	}
	for i := 2; i < r; i++ {
		table[i*c] = [3]int{
			diag: minInt,
			up:   table[(i-1)*c][up] + la[index[rSeq[i-1].L]*let],
			left: minInt,
		}
	}

	var scores [3]int
	for i := 1; i < r; i++ {
		for j := 1; j < c; j++ {
			var (
				rVal = index[rSeq[i-1].L]
				qVal = index[qSeq[j-1].L]
			)
			if rVal < 0 || qVal < 0 {
				continue
			}
			p := i*c + j
			scores = [3]int{
				diag: table[p-c-1][diag],
				up:   table[p-c-1][up],
				left: table[p-c-1][left],
			}
			table[p][diag] = max(&scores) + la[rVal*let+qVal]

			table[p][up] = max2(
				add(table[p-c][diag], a.GapOpen+la[rVal*let]),
				add(table[p-c][up], la[rVal*let]),
			)

			table[p][left] = max2(
				add(table[p-1][diag], a.GapOpen+la[qVal]),
				add(table[p-1][left], la[qVal]),
			)
		}
	}
	if debugFittedAffine {
		drawFittedAffineTableQLetters(rSeq, qSeq, index, table, a)
	}

	var aln []feat.Pair
	score, last, layer := 0, diag, diag
	var j int
	max := minInt
	for x, v := range table[(r-1)*c : len(table)] {
		if v[diag] >= max {
			j = x
			max = v[diag]
		}
	}
	i := r - 1
	for y := 1; y < r-1; y++ {
		v := table[(y*c)+c-1][diag]
		if v >= max {
			i = y
			max = v
			j = c - 1
		}
	}
	maxI, maxJ := i, j
	for i > 0 && j > 0 {
		var (
			rVal = index[rSeq[i-1].L]
			qVal = index[qSeq[j-1].L]
		)
		if rVal < 0 || qVal < 0 {
			continue
		}
		p := i*c + j
		switch table[p][layer] {
		case table[p-c][up] + la[rVal*let]:
			if last != up && p != len(table)-1 {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p][layer] - table[p-c][up]
			i--
			layer = up
			last = up
		case table[p-1][left] + la[qVal]:
			if last != left && p != len(table)-1 {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p][layer] - table[p-1][left]
			j--
			layer = left
			last = left
		case table[p-c][diag] + a.GapOpen + la[rVal*let]:
			if last != up && p != len(table)-1 {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p][layer] - table[p-c][diag]
			i--
			layer = diag
			last = up
		case table[p-1][diag] + a.GapOpen + la[qVal]:
			if last != left && p != len(table)-1 {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p][layer] - table[p-1][diag]
			j--
			layer = diag
			last = left
		case table[p-c-1][up] + la[rVal*let+qVal]:
			if last != diag {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p][layer] - table[p-c-1][up]
			i--
			j--
			layer = up
			last = diag
		case table[p-c-1][left] + la[rVal*let+qVal]:
			if last != diag {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p][layer] - table[p-c-1][left]
			i--
			j--
			layer = left
			last = diag
		case table[p-c-1][diag] + la[rVal*let+qVal]:
			if last != diag {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p][layer] - table[p-c-1][diag]
			i--
			j--
			layer = diag
			last = diag

		default:
			panic(fmt.Sprintf("align: fitted nw affine internal error: no path at row: %d col:%d layer:%s\n", i, j, "mul"[layer:layer+1]))
		}
	}

	aln = append(aln, &featPair{
		a:     feature{start: i, end: maxI},
		b:     feature{start: j, end: maxJ},
		score: score,
	})

	for i, j := 0, len(aln)-1; i < j; i, j = i+1, j-1 {
		aln[i], aln[j] = aln[j], aln[i]
	}

	return aln, nil
}
