package radix

// For output slot p with tuple (a_0..a_{d-1}), input i contributes
// x[a_i-1] when a_i>0 and 1 when a_i==0.
func augmentedSpreadMatrixLocal(J []int, i int) ([][]complex128, []complex128) {
	d := len(J)
	augSlots := make([]int, d)
	N := 1
	for k, j := range J {
		augSlots[k] = 1 + j
		N *= 1 + j
	}
	M := make([][]complex128, N)
	bias := make([]complex128, N)
	for p := 0; p < N; p++ {
		row := make([]complex128, J[i])
		tuple := unpackIndexLocal(p, augSlots)
		a := tuple[i]
		if a == 0 {
			bias[p] = 1
		} else {
			row[a-1] = 1
		}
		M[p] = row
	}
	return M, bias
}

// p = sum_i out[i] * (prod_{j>i} dims[j]).
func unpackIndexLocal(p int, dims []int) []int {
	out := make([]int, len(dims))
	for i := len(dims) - 1; i >= 0; i-- {
		out[i] = p % dims[i]
		p /= dims[i]
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ceilLog2(n int) int {
	if n <= 1 {
		return 0
	}
	r := 0
	for (1 << r) < n {
		r++
	}
	return r
}
