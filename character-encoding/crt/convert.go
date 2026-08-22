package crt

import (
	"fmt"
	"math"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
	"character-encoding/character-encoding/lut"
)

type CompiledUnary struct {
	In, Out  Spec
	Channels []*lut.CompiledUnary
}

func SwitchSpec(s Spec, kind charenc.EncodingKind) (Spec, error) {
	return NewSpec(s.Primes, kind, s.Reduced, s.Bits, s.Log2T)
}

func CompileSwitch(ctx *charctx.Context, in, out Spec, inputLevel int) (*CompiledUnary, error) {
	if err := checkCompatibleSwitchSpecs(in, out); err != nil {
		return nil, err
	}
	channels := make([]*lut.CompiledUnary, in.Channels())
	for i, p := range in.Primes {
		compiled, err := compileSwitchChannel(ctx, in.Specs[i], out.Specs[i], inputLevel)
		if err != nil {
			return nil, fmt.Errorf("crt.CompileSwitch: channel %d p=%d: %w", i, p, err)
		}
		channels[i] = compiled
	}
	return &CompiledUnary{In: in, Out: out, Channels: channels}, nil
}

func compileSwitchChannel(ctx *charctx.Context, in, out charenc.BlockSpec, inputLevel int) (*lut.CompiledUnary, error) {
	tr, err := compileSwitchTransform(in, out)
	if err != nil {
		return nil, err
	}
	compiled, err := blt.Compile(tr, ctx, inputLevel)
	if err != nil {
		return nil, err
	}
	ctx.EnsureGaloisKeys(compiled.GaloisEls)
	return &lut.CompiledUnary{
		Table: lut.UnaryTable{
			In:   in,
			Out:  out,
			Eval: func(x int) int { return x },
		},
		Transform: compiled,
	}, nil
}

func checkCompatibleSwitchSpecs(in, out Spec) error {
	if in.Channels() != out.Channels() {
		return fmt.Errorf("crt.CompileSwitch: channel count mismatch")
	}
	if in.Reduced != out.Reduced {
		return fmt.Errorf("crt.CompileSwitch: reduced/non-reduced mismatch")
	}
	if !in.Reduced {
		return fmt.Errorf("crt.CompileSwitch: only reduced BRU <-> LBRU conversions are implemented")
	}
	for i := range in.Primes {
		if in.Primes[i] != out.Primes[i] {
			return fmt.Errorf("crt.CompileSwitch: channel %d prime mismatch", i)
		}
	}
	return nil
}

func compileSwitchTransform(in, out charenc.BlockSpec) (blt.Transform, error) {
	p := in.Alphabet.Modulus
	if p != out.Alphabet.Modulus {
		return blt.Transform{}, fmt.Errorf("modulus mismatch")
	}
	if in.Alphabet.Kind == charenc.BRU && out.Alphabet.Kind == charenc.LBRU {
		matrix, bias, err := bruToLBRUMatrix(p, out.Alphabet.Generator)
		if err != nil {
			return blt.Transform{}, err
		}
		return blt.Transform{In: in, Out: out, Matrix: matrix, Bias: bias}, nil
	}
	if in.Alphabet.Kind == charenc.LBRU && out.Alphabet.Kind == charenc.BRU {
		matrix, bias, err := lbruToBRUMatrix(p, in.Alphabet.Generator)
		if err != nil {
			return blt.Transform{}, err
		}
		return blt.Transform{In: in, Out: out, Matrix: matrix, Bias: bias}, nil
	}
	if in.Alphabet.Kind == charenc.BRU && out.Alphabet.Kind == charenc.IND {
		matrix, bias, err := bruToINDMatrix(p)
		if err != nil {
			return blt.Transform{}, err
		}
		return blt.Transform{In: in, Out: out, Matrix: matrix, Bias: bias}, nil
	}
	if in.Alphabet.Kind == charenc.IND && out.Alphabet.Kind == charenc.BRU {
		matrix, bias, err := indToBRUMatrix(p)
		if err != nil {
			return blt.Transform{}, err
		}
		return blt.Transform{In: in, Out: out, Matrix: matrix, Bias: bias}, nil
	}
	if in.Alphabet.Kind == charenc.LBRU && out.Alphabet.Kind == charenc.IND {
		matrix, bias, err := lbruToINDMatrix(p, in.Alphabet.Generator)
		if err != nil {
			return blt.Transform{}, err
		}
		return blt.Transform{In: in, Out: out, Matrix: matrix, Bias: bias}, nil
	}
	if in.Alphabet.Kind == charenc.IND && out.Alphabet.Kind == charenc.LBRU {
		matrix, bias, err := indToLBRUMatrix(p, out.Alphabet.Generator)
		if err != nil {
			return blt.Transform{}, err
		}
		return blt.Transform{In: in, Out: out, Matrix: matrix, Bias: bias}, nil
	}
	inCodec, err := charenc.NewCodec(in)
	if err != nil {
		return blt.Transform{}, fmt.Errorf("input codec: %w", err)
	}
	outCodec, err := charenc.NewCodec(out)
	if err != nil {
		return blt.Transform{}, fmt.Errorf("output codec: %w", err)
	}
	values := make([][]complex128, p)
	for x := range values {
		values[x] = outCodec.EncodeValue(x)
	}
	matrix, bias, err := inCodec.Interpolate(values, out)
	if err != nil {
		return blt.Transform{}, fmt.Errorf("interpolate %v -> %v: %w", in.Alphabet.Kind, out.Alphabet.Kind, err)
	}
	return blt.Transform{In: in, Out: out, Matrix: matrix, Bias: bias}, nil
}

func bruToINDMatrix(p int) ([][]complex128, []complex128, error) {
	if p < 2 {
		return nil, nil, fmt.Errorf("prime must be >= 2")
	}
	n := p - 1
	matrix := make([][]complex128, n)
	bias := make([]complex128, n)
	scale := complex(1/float64(p), 0)
	for m := 1; m < p; m++ {
		row := make([]complex128, n)
		bias[m-1] = scale
		for k := 1; k < p; k++ {
			row[k-1] = scale * root(p, -k*m)
		}
		matrix[m-1] = row
	}
	return matrix, bias, nil
}

func indToBRUMatrix(p int) ([][]complex128, []complex128, error) {
	if p < 2 {
		return nil, nil, fmt.Errorf("prime must be >= 2")
	}
	n := p - 1
	matrix := make([][]complex128, n)
	bias := make([]complex128, n)
	for k := 1; k < p; k++ {
		row := make([]complex128, n)
		bias[k-1] = 1
		for m := 1; m < p; m++ {
			row[m-1] = root(p, k*m) - 1
		}
		matrix[k-1] = row
	}
	return matrix, bias, nil
}

func lbruToINDMatrix(p, generator int) ([][]complex128, []complex128, error) {
	if p < 2 {
		return nil, nil, fmt.Errorf("prime must be >= 2")
	}
	dlog, _, err := discreteLogTable(p, generator)
	if err != nil {
		return nil, nil, err
	}
	n := p - 1
	matrix := make([][]complex128, n)
	bias := make([]complex128, n)
	scale := complex(1/float64(n), 0)
	for m := 1; m < p; m++ {
		row := make([]complex128, n)
		row[0] = scale
		for j := 1; j < n; j++ {
			row[j] = scale * root(n, -j*dlog[m])
		}
		matrix[m-1] = row
	}
	return matrix, bias, nil
}

func indToLBRUMatrix(p, generator int) ([][]complex128, []complex128, error) {
	if p < 2 {
		return nil, nil, fmt.Errorf("prime must be >= 2")
	}
	dlog, _, err := discreteLogTable(p, generator)
	if err != nil {
		return nil, nil, err
	}
	n := p - 1
	matrix := make([][]complex128, n)
	bias := make([]complex128, n)
	matrix[0] = make([]complex128, n)
	for m := 1; m < p; m++ {
		matrix[0][m-1] = 1
	}
	for j := 1; j < n; j++ {
		row := make([]complex128, n)
		for m := 1; m < p; m++ {
			row[m-1] = root(n, j*dlog[m])
		}
		matrix[j] = row
	}
	return matrix, bias, nil
}

func bruToLBRUMatrix(p, generator int) ([][]complex128, []complex128, error) {
	if p < 2 {
		return nil, nil, fmt.Errorf("prime must be >= 2")
	}
	dlog, powers, err := discreteLogTable(p, generator)
	if err != nil {
		return nil, nil, err
	}
	_ = powers
	n := p - 1
	matrix := make([][]complex128, n)
	bias := make([]complex128, n)

	// Row 0 encodes the nonzero indicator 1_{x != 0}.
	matrix[0] = make([]complex128, n)
	bias[0] = complex(float64(n)/float64(p), 0)
	for k := 1; k < p; k++ {
		matrix[0][k-1] = complex(-1/float64(p), 0)
	}

	for j := 1; j < n; j++ {
		row := make([]complex128, n)
		gauss := complex(0, 0)
		for ell, x := range powers {
			gauss += root(n, j*ell) * root(p, -x)
		}
		scale := gauss / complex(float64(p), 0)
		for k := 1; k < p; k++ {
			row[k-1] = scale * root(n, -j*dlog[k])
		}
		matrix[j] = row
	}
	return matrix, bias, nil
}

func lbruToBRUMatrix(p, generator int) ([][]complex128, []complex128, error) {
	if p < 2 {
		return nil, nil, fmt.Errorf("prime must be >= 2")
	}
	dlog, powers, err := discreteLogTable(p, generator)
	if err != nil {
		return nil, nil, err
	}
	n := p - 1
	matrix := make([][]complex128, n)
	bias := make([]complex128, n)

	gauss := make([]complex128, n)
	gauss[0] = complex(-float64(p), 0)
	for j := 1; j < n; j++ {
		for ell, x := range powers {
			gauss[j] += root(p, x) * root(n, -j*ell)
		}
	}

	for k := 1; k < p; k++ {
		row := make([]complex128, n)
		bias[k-1] = 1
		row[0] = gauss[0] / complex(float64(n), 0)
		for j := 1; j < n; j++ {
			row[j] = gauss[j] * root(n, j*dlog[k]) / complex(float64(n), 0)
		}
		matrix[k-1] = row
	}
	return matrix, bias, nil
}

func discreteLogTable(p, generator int) ([]int, []int, error) {
	if generator == 0 {
		return nil, nil, fmt.Errorf("missing primitive root for p=%d", p)
	}
	n := p - 1
	dlog := make([]int, p)
	for i := range dlog {
		dlog[i] = -1
	}
	powers := make([]int, n)
	x := 1 % p
	for ell := 0; ell < n; ell++ {
		if dlog[x] != -1 {
			return nil, nil, fmt.Errorf("%d is not a primitive root modulo %d", generator, p)
		}
		dlog[x] = ell
		powers[ell] = x
		x = (x * generator) % p
	}
	return dlog, powers, nil
}

func root(modulus, k int) complex128 {
	theta := 2 * math.Pi * float64(positiveMod(k, modulus)) / float64(modulus)
	return complex(math.Cos(theta), math.Sin(theta))
}
