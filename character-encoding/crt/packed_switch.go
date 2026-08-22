package crt

import (
	"fmt"
	"sync"

	"github.com/Pro7ech/lattigo/he"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

type CompiledPackedUnary struct {
	In, Out    PackedSpec
	Transforms []*blt.RawCompiled
}

func SwitchPackedSpec(s PackedSpec, kind charenc.EncodingKind) (PackedSpec, error) {
	base, err := SwitchSpec(s.Base, kind)
	if err != nil {
		return PackedSpec{}, err
	}
	packCap := s.PackCap
	if packCap <= 0 {
		packCap = s.MaxSlots
	}
	out, err := NewPackedSpecWithMaxUsedSlots(base, s.MaxSlots, packCap)
	if err != nil {
		return PackedSpec{}, err
	}
	if !samePackedLayout(s, out) {
		return PackedSpec{}, fmt.Errorf("crt.SwitchPackedSpec: switched encoding changes packed layout")
	}
	return out, nil
}

func CompilePackedSwitch(ctx *charctx.Context, in, out PackedSpec, inputLevel int) (*CompiledPackedUnary, error) {
	return CompilePackedSwitchWithOptions(ctx, in, out, inputLevel, blt.CompileOptions{})
}

// Diagonal-only ciphertexts ignore opts and use plaintext-ciphertext multiplication.
func CompilePackedSwitchWithOptions(ctx *charctx.Context, in, out PackedSpec, inputLevel int, opts blt.CompileOptions) (*CompiledPackedUnary, error) {
	if err := checkCompatibleSwitchSpecs(in.Base, out.Base); err != nil {
		return nil, err
	}
	if in.MaxSlots != out.MaxSlots || !samePackedLayout(in, out) {
		return nil, fmt.Errorf("crt.CompilePackedSwitch: packed layout mismatch")
	}
	transforms := make([]*blt.RawCompiled, in.Ciphertexts())
	for ctIdx := range transforms {
		raw, err := compilePackedSwitchCiphertext(ctx, in, out, ctIdx, inputLevel, opts)
		if err != nil {
			return nil, fmt.Errorf("crt.CompilePackedSwitch: ciphertext %d: %w", ctIdx, err)
		}
		ctx.EnsureGaloisKeys(raw.GaloisEls)
		transforms[ctIdx] = raw
	}
	return &CompiledPackedUnary{In: in, Out: out, Transforms: transforms}, nil
}

func compilePackedSwitchCiphertext(ctx *charctx.Context, in, out PackedSpec, ctIdx int, inputLevel int, opts blt.CompileOptions) (*blt.RawCompiled, error) {
	slots := ctx.Params.MaxSlots()
	inSlots := in.CTSlots[ctIdx]
	outSlots := out.CTSlots[ctIdx]
	diagonals := he.Diagonals[complex128]{}
	bias := make([]complex128, outSlots)
	diagonalOnly := true

	for chIdx := range in.Channels {
		inCh := in.Channels[chIdx]
		if inCh.Ciphertext != ctIdx {
			continue
		}
		outCh := out.Channels[chIdx]
		tr, err := compileSwitchTransform(in.Base.Specs[chIdx], out.Base.Specs[chIdx])
		if err != nil {
			return nil, fmt.Errorf("channel %d p=%d: %w", chIdx, in.Base.Primes[chIdx], err)
		}
		if len(tr.Bias) != outCh.Slots || len(tr.Matrix) != outCh.Slots {
			return nil, fmt.Errorf("channel %d transform shape mismatch", chIdx)
		}
		for word := 0; word < in.BatchSize; word++ {
			inWordBase := word * in.WordStride
			outWordBase := word * out.WordStride
			copy(bias[outWordBase+outCh.Offset:outWordBase+outCh.Offset+outCh.Slots], tr.Bias)
			for r := 0; r < outCh.Slots; r++ {
				if len(tr.Matrix[r]) != inCh.Slots {
					return nil, fmt.Errorf("channel %d row %d has %d cols, expected %d", chIdx, r, len(tr.Matrix[r]), inCh.Slots)
				}
				outSlot := outWordBase + outCh.Offset + r
				for c, v := range tr.Matrix[r] {
					if v == 0 {
						continue
					}
					inSlot := inWordBase + inCh.Offset + c
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

	return blt.CompileDiagonalsWithOptions(diagonals, diagonalOnly, bias, outSlots, inSlots, ctx, inputLevel, opts)
}

func (e *Evaluator) EvalPackedSwitch(x PackedCiphertext, compiled *CompiledPackedUnary, parallel bool) (PackedCiphertext, error) {
	if !samePackedSpec(x.Spec, compiled.In) {
		return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedSwitch: input spec mismatch")
	}
	if len(x.CTs) != x.Spec.Ciphertexts() || len(compiled.Transforms) != x.Spec.Ciphertexts() {
		return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedSwitch: ciphertext count mismatch")
	}
	out := PackedCiphertext{Spec: compiled.Out, CTs: make([]*rlwe.Ciphertext, len(x.CTs))}
	workers := channelParallelism(parallel, len(x.CTs), e.LUT.BLT.Capacity())
	if workers == 1 {
		w := e.LUT.BLT.GetWorker()
		defer e.LUT.BLT.PutWorker(w)
		for i := range x.CTs {
			ct, err := e.LUT.BLT.ApplyRawWith(w, x.CTs[i], compiled.Transforms[i])
			if err != nil {
				return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedSwitch: ciphertext %d: %w", i, err)
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
			return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedSwitch: ciphertext %d: %w", i, err)
		}
	}
	return out, nil
}

func samePackedLayout(a, b PackedSpec) bool {
	if a.MaxSlots != b.MaxSlots || a.PackCap != b.PackCap || a.WordStride != b.WordStride || a.BatchSize != b.BatchSize || a.TotalUsed != b.TotalUsed || len(a.CTSlots) != len(b.CTSlots) || len(a.Channels) != len(b.Channels) {
		return false
	}
	for i := range a.CTSlots {
		if a.CTSlots[i] != b.CTSlots[i] {
			return false
		}
	}
	for i := range a.Channels {
		if a.Channels[i] != b.Channels[i] {
			return false
		}
	}
	return true
}
