package lut

import (
	"fmt"
	"strings"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

// CompiledTensor holds the augmented-packing schedule for a d-variate LUT
// (d in {2,3,4}): each input is spread into the common feature layout of size
// N = product(1+J_i) (axis index 0 denotes the constant copy 1), a balanced
// multiplication tree packs the d spreads into one ciphertext, and one final
// BLT writes the output encoding.
type CompiledTensor struct {
	Table        MultiTable
	Arity        int
	InputLevel   int
	Spreads      []*blt.RawCompiled
	FinalLT      *blt.RawCompiled
	MulSchedule  []mulStep
	MulLevels    [][]int
	PackedLevel  int
	AllGaloisEls []uint64
}

type mulStep struct {
	Left, Right int
}

func CompileMultivariate(table MultiTable, ctx *charctx.Context, inputLevel int) (*CompiledTensor, error) {
	d := len(table.In)
	if d < 2 || d > 4 {
		return nil, fmt.Errorf("lut.CompileMultivariate: arity %d outside [2,4]", d)
	}
	for i, spec := range table.In {
		if !spec.Reduced {
			return nil, fmt.Errorf("lut.CompileMultivariate: input %d must be reduced", i)
		}
	}
	rescale := ctx.Params.LevelsConsumedPerRescaling()
	mulDepth := mulTreeDepth(d)
	totalDepth := (1 + mulDepth + 1) * rescale
	if inputLevel < totalDepth {
		return nil, fmt.Errorf("lut.CompileMultivariate: arity %d needs %d levels, input has %d", d, totalDepth, inputLevel)
	}

	inCodecs := make([]charenc.Codec, d)
	J := make([]int, d)
	for i, spec := range table.In {
		c, err := charenc.NewCodec(spec)
		if err != nil {
			return nil, fmt.Errorf("lut.CompileMultivariate: input %d codec: %w", i, err)
		}
		inCodecs[i] = c
		J[i] = c.Spec().Slots
	}
	outCodec, err := charenc.NewCodec(table.Out)
	if err != nil {
		return nil, fmt.Errorf("lut.CompileMultivariate: output codec: %w", err)
	}
	R := outCodec.Spec().Slots
	N := 1
	for _, j := range J {
		N *= 1 + j
	}
	if N > ctx.Params.MaxSlots() {
		return nil, fmt.Errorf("lut.CompileMultivariate: tensor size %d exceeds %d slots; raise LogN", N, ctx.Params.MaxSlots())
	}

	spreads := make([]*blt.RawCompiled, d)
	for i := 0; i < d; i++ {
		mat, bias := augmentedSpreadMatrix(J, i)
		raw, err := blt.CompileMatrix(mat, bias, N, J[i], ctx, inputLevel)
		if err != nil {
			return nil, fmt.Errorf("lut.CompileMultivariate: spread %d: %w", i, err)
		}
		spreads[i] = raw
	}

	schedule, mulLevels, packedDepth := mulSchedulePlan(d)
	packedLevel := inputLevel - (1+packedDepth)*rescale

	coef, err := interpolateAugmentedTensor(inCodecs, outCodec, J, table.Eval)
	if err != nil {
		return nil, fmt.Errorf("lut.CompileMultivariate: interpolate: %w", err)
	}
	finalLT, err := blt.CompileMatrix(coef, nil, R, N, ctx, packedLevel)
	if err != nil {
		return nil, fmt.Errorf("lut.CompileMultivariate: final LT: %w", err)
	}

	all := make(map[uint64]struct{}, 64)
	for _, s := range spreads {
		for _, g := range s.GaloisEls {
			all[g] = struct{}{}
		}
	}
	for _, g := range finalLT.GaloisEls {
		all[g] = struct{}{}
	}
	allGalEls := make([]uint64, 0, len(all))
	for g := range all {
		allGalEls = append(allGalEls, g)
	}
	ctx.EnsureGaloisKeys(allGalEls)

	return &CompiledTensor{
		Table:        table,
		Arity:        d,
		InputLevel:   inputLevel,
		Spreads:      spreads,
		FinalLT:      finalLT,
		MulSchedule:  schedule,
		MulLevels:    mulLevels,
		PackedLevel:  packedLevel,
		AllGaloisEls: allGalEls,
	}, nil
}

func mulTreeDepth(d int) int {
	if d <= 1 {
		return 0
	}
	depth := 0
	for n := d; n > 1; n = (n + 1) / 2 {
		depth++
	}
	return depth
}

func mulSchedulePlan(d int) ([]mulStep, [][]int, int) {
	switch d {
	case 2:
		return []mulStep{{0, 1}}, [][]int{{0}}, 1
	case 3:
		return []mulStep{{0, 1}, {3, 2}}, [][]int{{0}, {1}}, 2
	case 4:
		return []mulStep{{0, 1}, {2, 3}, {4, 5}}, [][]int{{0, 1}, {2}}, 2
	default:
		return nil, nil, 0
	}
}

// augmentedSpreadMatrix produces the (N x J[i]) matrix and bias for input i:
// for slot p with tuple (a_0..a_{d-1}), the row picks x[a_i - 1] when a_i > 0
// and the bias is 1 when a_i = 0.
func augmentedSpreadMatrix(J []int, i int) ([][]complex128, []complex128) {
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
		tuple := unpackIndex(p, augSlots)
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

func unpackIndex(p int, dims []int) []int {
	out := make([]int, len(dims))
	for i := len(dims) - 1; i >= 0; i-- {
		out[i] = p % dims[i]
		p /= dims[i]
	}
	return out
}

// interpolateAugmentedTensor builds the (R x N) matrix mapping augmented
// tensor features to the output encoding. The basis B is square invertible
// because reduced inputs give |input tuples| = product T_i = product(1+J_i).
func interpolateAugmentedTensor(inCodecs []charenc.Codec, out charenc.Codec, J []int, fn func([]int) int) ([][]complex128, error) {
	d := len(inCodecs)
	T := make([]int, d)
	augSlots := make([]int, d)
	totalInputs := 1
	totalFeatures := 1
	for i, c := range inCodecs {
		T[i] = c.Spec().Alphabet.Modulus
		augSlots[i] = 1 + J[i]
		totalInputs *= T[i]
		totalFeatures *= augSlots[i]
	}
	if totalInputs != totalFeatures {
		return nil, fmt.Errorf("interpolateAugmentedTensor: |inputs|=%d != |features|=%d", totalInputs, totalFeatures)
	}

	enc := make([][][]complex128, d)
	for i, c := range inCodecs {
		enc[i] = make([][]complex128, T[i])
		for v := 0; v < T[i]; v++ {
			enc[i][v] = c.EncodeValue(v)
		}
	}

	B := make([][]complex128, totalInputs)
	for vp := 0; vp < totalInputs; vp++ {
		v := unpackIndex(vp, T)
		row := make([]complex128, totalFeatures)
		for fp := 0; fp < totalFeatures; fp++ {
			a := unpackIndex(fp, augSlots)
			prod := complex(1, 0)
			for i := 0; i < d; i++ {
				if a[i] == 0 {
					continue
				}
				prod *= enc[i][v[i]][a[i]-1]
			}
			row[fp] = prod
		}
		B[vp] = row
	}

	R := out.Spec().Slots
	coef := make([][]complex128, R)
	for r := range coef {
		coef[r] = make([]complex128, totalFeatures)
	}
	keyParts := make([]string, len(inCodecs))
	for i, c := range inCodecs {
		s := c.Spec()
		keyParts[i] = fmt.Sprintf("%s-m=%d-r=%v-g=%d-o=%d",
			s.Alphabet.Kind, s.Alphabet.Modulus, s.Reduced,
			s.Alphabet.Generator, s.Alphabet.Omitted)
	}
	basisKey := fmt.Sprintf("interpolateAugmentedTensor/arity=%d/[%s]", len(inCodecs),
		strings.Join(keyParts, ","))
	basisLU, err := charenc.LUFactorizeCached(basisKey, B)
	if err != nil {
		return nil, fmt.Errorf("interpolateAugmentedTensor: factorize: %w", err)
	}
	for r := 0; r < R; r++ {
		yr := make([]complex128, totalInputs)
		for vp := 0; vp < totalInputs; vp++ {
			v := unpackIndex(vp, T)
			yr[vp] = out.EncodeValue(fn(v))[r]
		}
		sol, err := basisLU.Solve(yr)
		if err != nil {
			return nil, fmt.Errorf("interpolateAugmentedTensor: solve r=%d: %w", r, err)
		}
		copy(coef[r], sol)
	}
	return coef, nil
}
