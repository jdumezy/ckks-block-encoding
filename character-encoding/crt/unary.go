package crt

import (
	"fmt"
	"sync"

	"github.com/Pro7ech/lattigo/he"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
	"character-encoding/character-encoding/lut"
)

type UnaryFunc func(p, x int) int

// Reduces mod p between stages so each f sees a valid residue.
func ComposeUnaryFuncs(funcs ...UnaryFunc) (UnaryFunc, error) {
	if len(funcs) == 0 {
		return nil, fmt.Errorf("crt.ComposeUnaryFuncs: no functions")
	}
	for i, f := range funcs {
		if f == nil {
			return nil, fmt.Errorf("crt.ComposeUnaryFuncs: nil function at index %d", i)
		}
	}
	return func(p, x int) int {
		y := positiveMod(x, p)
		for _, f := range funcs {
			y = positiveMod(f(p, y), p)
		}
		return y
	}, nil
}

func ComposeUnaryTables(s Spec, tableSets ...[][]int) ([][]int, error) {
	if len(tableSets) == 0 {
		return nil, fmt.Errorf("crt.ComposeUnaryTables: no table sets")
	}
	for i, tables := range tableSets {
		if err := checkUnaryTables(s, tables); err != nil {
			return nil, fmt.Errorf("crt.ComposeUnaryTables: table set %d: %w", i, err)
		}
	}
	out := make([][]int, s.Channels())
	for chIdx, p := range s.Primes {
		table := make([]int, p)
		for x := 0; x < p; x++ {
			y := x
			for _, tables := range tableSets {
				y = positiveMod(tables[chIdx][y], p)
			}
			table[x] = y
		}
		out[chIdx] = table
	}
	return out, nil
}

type CompiledUnaryLUT struct {
	Spec     Spec
	Channels []*lut.CompiledUnary
}

func CompileUnaryLUT(ctx *charctx.Context, s Spec, f UnaryFunc, inputLevel int) (*CompiledUnaryLUT, error) {
	if f == nil {
		return nil, fmt.Errorf("crt.CompileUnaryLUT: nil LUT function")
	}
	return compileUnaryLUTWithChannelFunc(ctx, s, func(_ int, p int, x int) int {
		return f(p, x)
	}, inputLevel)
}

func CompileComposedUnaryLUT(ctx *charctx.Context, s Spec, inputLevel int, funcs ...UnaryFunc) (*CompiledUnaryLUT, error) {
	fused, err := ComposeUnaryFuncs(funcs...)
	if err != nil {
		return nil, err
	}
	return CompileUnaryLUT(ctx, s, fused, inputLevel)
}

// tables[i] must have length p_i; entries are reduced mod p_i.
func CompileUnaryLUTFromTables(ctx *charctx.Context, s Spec, tables [][]int, inputLevel int) (*CompiledUnaryLUT, error) {
	if err := checkUnaryTables(s, tables); err != nil {
		return nil, err
	}
	return compileUnaryLUTWithChannelFunc(ctx, s, func(chIdx int, _ int, x int) int {
		return tables[chIdx][x]
	}, inputLevel)
}

func CompileComposedUnaryLUTFromTables(ctx *charctx.Context, s Spec, inputLevel int, tableSets ...[][]int) (*CompiledUnaryLUT, error) {
	fused, err := ComposeUnaryTables(s, tableSets...)
	if err != nil {
		return nil, err
	}
	return CompileUnaryLUTFromTables(ctx, s, fused, inputLevel)
}

func compileUnaryLUTWithChannelFunc(ctx *charctx.Context, s Spec, f func(chIdx, p, x int) int, inputLevel int) (*CompiledUnaryLUT, error) {
	channels := make([]*lut.CompiledUnary, s.Channels())
	for i, p := range s.Primes {
		p := p
		i := i
		table := lut.UnaryTable{
			In:  s.Specs[i],
			Out: s.Specs[i],
			Eval: func(x int) int {
				return positiveMod(f(i, p, x), p)
			},
		}
		compiled, err := lut.CompileUnary(table, ctx, inputLevel)
		if err != nil {
			return nil, fmt.Errorf("crt.CompileUnaryLUT: channel %d p=%d: %w", i, p, err)
		}
		channels[i] = compiled
	}
	return &CompiledUnaryLUT{Spec: s, Channels: channels}, nil
}

func (e *Evaluator) EvalUnaryLUT(x Ciphertext, compiled *CompiledUnaryLUT, parallel bool) (Ciphertext, error) {
	if compiled == nil {
		return Ciphertext{}, fmt.Errorf("crt.EvalUnaryLUT: nil compiled LUT")
	}
	if !sameSpec(x.Spec, compiled.Spec) {
		return Ciphertext{}, fmt.Errorf("crt.EvalUnaryLUT: input spec mismatch")
	}
	if len(x.Blocks) != x.Spec.Channels() || len(compiled.Channels) != x.Spec.Channels() {
		return Ciphertext{}, fmt.Errorf("crt.EvalUnaryLUT: channel count mismatch")
	}
	out := Ciphertext{Spec: x.Spec, Blocks: make([]charenc.CipherBlock, x.Spec.Channels())}
	workers := channelParallelism(parallel, x.Spec.Channels(), e.LUT.BLT.Capacity())
	if workers == 1 {
		for i := range x.Blocks {
			block, err := e.LUT.EvalUnary(x.Blocks[i], compiled.Channels[i])
			if err != nil {
				return Ciphertext{}, fmt.Errorf("crt.EvalUnaryLUT: channel %d: %w", i, err)
			}
			out.Blocks[i] = block
		}
		return out, nil
	}

	var wg sync.WaitGroup
	errs := make([]error, x.Spec.Channels())
	wg.Add(workers)
	for workerID := 0; workerID < workers; workerID++ {
		go func(workerID int) {
			defer wg.Done()
			for i := workerID; i < len(x.Blocks); i += workers {
				block, err := e.LUT.EvalUnary(x.Blocks[i], compiled.Channels[i])
				if err != nil {
					errs[i] = err
					return
				}
				out.Blocks[i] = block
			}
		}(workerID)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return Ciphertext{}, fmt.Errorf("crt.EvalUnaryLUT: channel %d: %w", i, err)
		}
	}
	return out, nil
}

type CompiledPackedUnaryLUT struct {
	Spec       PackedSpec
	Transforms []*blt.RawCompiled
}

func CompilePackedUnaryLUT(ctx *charctx.Context, s PackedSpec, f UnaryFunc, inputLevel int) (*CompiledPackedUnaryLUT, error) {
	return CompilePackedUnaryLUTWithOptions(ctx, s, f, inputLevel, blt.CompileOptions{})
}

func CompilePackedUnaryLUTWithOptions(ctx *charctx.Context, s PackedSpec, f UnaryFunc, inputLevel int, opts blt.CompileOptions) (*CompiledPackedUnaryLUT, error) {
	if f == nil {
		return nil, fmt.Errorf("crt.CompilePackedUnaryLUT: nil LUT function")
	}
	return compilePackedUnaryLUTWithChannelFunc(ctx, s, func(_ int, p int, x int) int {
		return f(p, x)
	}, inputLevel, opts)
}

func CompilePackedComposedUnaryLUT(ctx *charctx.Context, s PackedSpec, inputLevel int, opts blt.CompileOptions, funcs ...UnaryFunc) (*CompiledPackedUnaryLUT, error) {
	fused, err := ComposeUnaryFuncs(funcs...)
	if err != nil {
		return nil, err
	}
	return CompilePackedUnaryLUTWithOptions(ctx, s, fused, inputLevel, opts)
}

func CompilePackedUnaryLUTFromTables(ctx *charctx.Context, s PackedSpec, tables [][]int, inputLevel int) (*CompiledPackedUnaryLUT, error) {
	return CompilePackedUnaryLUTFromTablesWithOptions(ctx, s, tables, inputLevel, blt.CompileOptions{})
}

func CompilePackedUnaryLUTFromTablesWithOptions(ctx *charctx.Context, s PackedSpec, tables [][]int, inputLevel int, opts blt.CompileOptions) (*CompiledPackedUnaryLUT, error) {
	if err := checkUnaryTables(s.Base, tables); err != nil {
		return nil, err
	}
	return compilePackedUnaryLUTWithChannelFunc(ctx, s, func(chIdx int, _ int, x int) int {
		return tables[chIdx][x]
	}, inputLevel, opts)
}

func CompilePackedComposedUnaryLUTFromTables(ctx *charctx.Context, s PackedSpec, inputLevel int, tableSets ...[][]int) (*CompiledPackedUnaryLUT, error) {
	return CompilePackedComposedUnaryLUTFromTablesWithOptions(ctx, s, inputLevel, blt.CompileOptions{}, tableSets...)
}

func CompilePackedComposedUnaryLUTFromTablesWithOptions(ctx *charctx.Context, s PackedSpec, inputLevel int, opts blt.CompileOptions, tableSets ...[][]int) (*CompiledPackedUnaryLUT, error) {
	fused, err := ComposeUnaryTables(s.Base, tableSets...)
	if err != nil {
		return nil, err
	}
	return CompilePackedUnaryLUTFromTablesWithOptions(ctx, s, fused, inputLevel, opts)
}

func compilePackedUnaryLUTWithChannelFunc(ctx *charctx.Context, s PackedSpec, f func(chIdx, p, x int) int, inputLevel int, opts blt.CompileOptions) (*CompiledPackedUnaryLUT, error) {
	transforms := make([]*blt.RawCompiled, s.Ciphertexts())
	for ctIdx := range transforms {
		raw, err := compilePackedUnaryLUTCiphertext(ctx, s, ctIdx, f, inputLevel, opts)
		if err != nil {
			return nil, fmt.Errorf("crt.CompilePackedUnaryLUT: ciphertext %d: %w", ctIdx, err)
		}
		ctx.EnsureGaloisKeys(raw.GaloisEls)
		transforms[ctIdx] = raw
	}
	return &CompiledPackedUnaryLUT{Spec: s, Transforms: transforms}, nil
}

func compilePackedUnaryLUTCiphertext(ctx *charctx.Context, s PackedSpec, ctIdx int, f func(chIdx, p, x int) int, inputLevel int, opts blt.CompileOptions) (*blt.RawCompiled, error) {
	slots := ctx.Params.MaxSlots()
	usedSlots := s.CTSlots[ctIdx]
	diagonals := he.Diagonals[complex128]{}
	bias := make([]complex128, usedSlots)
	diagonalOnly := true

	for chIdx := range s.Channels {
		ch := s.Channels[chIdx]
		if ch.Ciphertext != ctIdx {
			continue
		}
		p := s.Base.Primes[chIdx]
		tr, err := compileUnaryLUTTransform(s.Base.Specs[chIdx], func(x int) int {
			return positiveMod(f(chIdx, p, x), p)
		})
		if err != nil {
			return nil, fmt.Errorf("channel %d p=%d: %w", chIdx, p, err)
		}
		if len(tr.Bias) != ch.Slots || len(tr.Matrix) != ch.Slots {
			return nil, fmt.Errorf("channel %d transform shape mismatch", chIdx)
		}
		for word := 0; word < s.BatchSize; word++ {
			wordBase := word * s.WordStride
			copy(bias[wordBase+ch.Offset:wordBase+ch.Offset+ch.Slots], tr.Bias)
			for r := 0; r < ch.Slots; r++ {
				if len(tr.Matrix[r]) != ch.Slots {
					return nil, fmt.Errorf("channel %d row %d has %d cols, expected %d", chIdx, r, len(tr.Matrix[r]), ch.Slots)
				}
				outSlot := wordBase + ch.Offset + r
				for c, v := range tr.Matrix[r] {
					if v == 0 {
						continue
					}
					inSlot := wordBase + ch.Offset + c
					if inSlot != outSlot {
						diagonalOnly = false
					}
					d := (inSlot - outSlot) % slots
					if d < 0 {
						d += slots
					}
					row, ok := diagonals[d]
					if !ok {
						row = make([]complex128, slots)
						diagonals[d] = row
					}
					row[outSlot] = v
				}
			}
		}
	}

	return blt.CompileDiagonalsWithOptions(diagonals, diagonalOnly, bias, usedSlots, usedSlots, ctx, inputLevel, opts)
}

func compileUnaryLUTTransform(spec charenc.BlockSpec, f func(int) int) (blt.Transform, error) {
	codec, err := charenc.NewCodec(spec)
	if err != nil {
		return blt.Transform{}, err
	}
	return blt.CompileUnary(codec, codec, f)
}

func (e *Evaluator) EvalPackedUnaryLUT(x PackedCiphertext, compiled *CompiledPackedUnaryLUT, parallel bool) (PackedCiphertext, error) {
	if compiled == nil {
		return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedUnaryLUT: nil compiled LUT")
	}
	if !samePackedSpec(x.Spec, compiled.Spec) {
		return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedUnaryLUT: input spec mismatch")
	}
	if len(x.CTs) != x.Spec.Ciphertexts() || len(compiled.Transforms) != x.Spec.Ciphertexts() {
		return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedUnaryLUT: ciphertext count mismatch")
	}
	out := PackedCiphertext{Spec: x.Spec, CTs: make([]*rlwe.Ciphertext, len(x.CTs))}
	workers := channelParallelism(parallel, len(x.CTs), e.LUT.BLT.Capacity())
	if workers == 1 {
		w := e.LUT.BLT.GetWorker()
		defer e.LUT.BLT.PutWorker(w)
		for i := range x.CTs {
			ct, err := e.LUT.BLT.ApplyRawWith(w, x.CTs[i], compiled.Transforms[i])
			if err != nil {
				return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedUnaryLUT: ciphertext %d: %w", i, err)
			}
			out.CTs[i] = ct
		}
		return out, nil
	}

	var wg sync.WaitGroup
	errs := make([]error, len(x.CTs))
	wg.Add(workers)
	for workerID := 0; workerID < workers; workerID++ {
		go func(workerID int) {
			defer wg.Done()
			w := e.LUT.BLT.GetWorker()
			defer e.LUT.BLT.PutWorker(w)
			for i := workerID; i < len(x.CTs); i += workers {
				ct, err := e.LUT.BLT.ApplyRawWith(w, x.CTs[i], compiled.Transforms[i])
				if err != nil {
					errs[i] = err
					return
				}
				out.CTs[i] = ct
			}
		}(workerID)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedUnaryLUT: ciphertext %d: %w", i, err)
		}
	}
	return out, nil
}

func sameSpec(a, b Spec) bool {
	if a.Kind != b.Kind || a.Reduced != b.Reduced || a.Channels() != b.Channels() {
		return false
	}
	for i := range a.Primes {
		if a.Primes[i] != b.Primes[i] {
			return false
		}
	}
	return true
}

func checkUnaryTables(s Spec, tables [][]int) error {
	if len(tables) != s.Channels() {
		return fmt.Errorf("crt: got %d LUT tables, expected %d", len(tables), s.Channels())
	}
	for i, p := range s.Primes {
		if len(tables[i]) != p {
			return fmt.Errorf("crt: table %d has length %d, expected p=%d", i, len(tables[i]), p)
		}
	}
	return nil
}
