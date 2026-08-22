package crt

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func positiveMod(x, m int) int {
	r := x % m
	if r < 0 {
		r += m
	}
	return r
}
