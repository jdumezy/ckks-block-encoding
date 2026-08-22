package lut

import "character-encoding/character-encoding/charenc"

type UnaryTable struct {
	In   charenc.BlockSpec
	Out  charenc.BlockSpec
	Eval func(x int) int
}

type MultiTable struct {
	In   []charenc.BlockSpec
	Out  charenc.BlockSpec
	Eval func(xs []int) int
}
