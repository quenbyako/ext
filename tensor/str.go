package tensor

import (
	"fmt"
	"strings"

	"github.com/quenbyako/ext/slices"
)

func strTensor(higherDimIndex []int, dims []int, at func(...int) string, pad int) (res string) {
	switch len(dims) {
	case 1:
		return strVector(dims[0], func(i0 int) string { return at(i0) }, pad)
	case 2:
		return strMatrix(dims[0], dims[1], func(i0, i1 int) string { return at(i0, i1) }, pad)
	case 3:
		for i2 := range dims[2] {
			if i2 > 0 {
				res += "\n\n"
			}
			res += fmt.Sprintf("DIM[%v]\n", strings.Join(slices.Remap(append(higherDimIndex, i2), func(i int) string { return fmt.Sprint(i) }), " "))
			res += strMatrix(dims[0], dims[1], func(i0, i1 int) string { return at(i0, i1, i2) }, pad)
		}
	default:
		for iN := range dims[len(dims)-1] {
			if iN > 0 {
				res += "\n\n"
			}

			res += strTensor(append(higherDimIndex, iN), dims[:len(dims)-1], func(i ...int) string { return at(append(i, iN)...) }, pad)
		}
	}

	return res
}

func strVector(dim0 int, at func(n0 int) string, pad int) string {
	cells := make([]string, dim0)
	for i0 := range dim0 {
		// cells[i0] = fmt.Sprintf("%-*v", pad, at(i0))
		cells[i0] = fmt.Sprintf("%*v", pad, at(i0))
	}

	return "[" + strings.Join(cells, " ") + "]"
}

func strMatrix(dim0, dim1 int, at func(n0, n1 int) string, pad int) string {
	rows := make([]string, dim1+2) // 2 for top and bottom part

	tablePart := slices.Repeat(dim0, strings.Repeat("─", pad))
	rows[0] = "┌" + strings.Join(tablePart, "┬") + "┐"
	rows[len(rows)-1] = "└" + strings.Join(tablePart, "┴") + "┘"

	for i1 := range dim1 {
		cells := make([]string, dim0)
		for i0 := range dim0 {
			// 	cells[i0] = fmt.Sprintf("%-*v", pad, at(i0, i1))
			cells[i0] = fmt.Sprintf("%*v", pad, at(i0, i1))
		}
		rows[i1+1] = "│" + strings.Join(cells, " ") + "│"
	}

	return strings.Join(rows, "\n")
}
