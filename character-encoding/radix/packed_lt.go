package radix

import (
	"fmt"
)

// Packed less-than (one bit per word). Depth: 3 + ceil(log2d).
// Requires d to be a power of two so the scan stays within the word.
type CompiledLt struct {
	Spec         PaddedSpec
	LUT          *CompiledPackedSharedBivariate
	InputLevel   int
	OutputLevel  int
	Rounds       int
	AllGaloisEls []uint64
}

func (e *Evaluator) CompileLt(spec PaddedSpec, inputLevel int) (*CompiledLt, error) {
	rescale := e.Ctx.Params.LevelsConsumedPerRescaling()
	d := spec.Base.Width
	rounds := ceilLog2(d)
	if (1 << rounds) != d {
		return nil, fmt.Errorf("radix.CompileLt: d=%d must be a power of two", d)
	}

	maxStep := kogglesStoneMaxStep(d)
	if spec.WordStride < d+maxStep {
		return nil, fmt.Errorf("radix.CompileLt: WordStride=%d < d+maxStep=%d", spec.WordStride, d+maxStep)
	}
	if spec.BatchSize*spec.WordStride > spec.BlocksPerCT {
		return nil, fmt.Errorf("radix.CompileLt: BatchSize*WordStride=%d > BlocksPerCT=%d",
			spec.BatchSize*spec.WordStride, spec.BlocksPerCT)
	}

	genEval := func(v0, v1 int) []complex128 {
		if v0 < v1 {
			return []complex128{1}
		}
		return []complex128{0}
	}
	propEval := func(v0, v1 int) []complex128 {
		if v0 == v1 {
			return []complex128{1}
		}
		return []complex128{0}
	}

	lut, err := CompilePackedSharedBivariate(e.Ctx, spec, 1,
		[]func(v0, v1 int) []complex128{genEval, propEval}, inputLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileLt: LUT: %w", err)
	}

	ksGals := make([]uint64, 0, rounds)
	for r := 0; r < rounds; r++ {
		shift := positiveMod(-(1<<r)*spec.Stride, spec.MaxSlots)
		ksGals = append(ksGals, e.Ctx.Params.GaloisElement(shift))
	}
	e.Ctx.EnsureGaloisKeys(ksGals)

	galSet := map[uint64]struct{}{}
	for _, g := range ksGals {
		galSet[g] = struct{}{}
	}
	for _, g := range lut.AllGaloisEls {
		galSet[g] = struct{}{}
	}
	galEls := make([]uint64, 0, len(galSet))
	for g := range galSet {
		galEls = append(galEls, g)
	}

	outputLevel := inputLevel - (3+rounds)*rescale
	return &CompiledLt{
		Spec:         spec,
		LUT:          lut,
		InputLevel:   inputLevel,
		OutputLevel:  outputLevel,
		Rounds:       rounds,
		AllGaloisEls: galEls,
	}, nil
}

func (e *Evaluator) Lt(x, y PaddedCiphertext, plan *CompiledLt) (*LtResult, error) {
	if !samePaddedSpec(x.Spec, plan.Spec) || !samePaddedSpec(y.Spec, plan.Spec) {
		return nil, fmt.Errorf("radix.Lt: spec mismatch")
	}
	if len(x.CTs) != 1 || len(y.CTs) != 1 {
		return nil, fmt.Errorf("radix.Lt: expected 1 ct each")
	}
	e.Ctx.EnsureGaloisKeys(plan.AllGaloisEls)

	gpOuts, err := e.evalPackedSharedBivariate(x.CTs[0], y.CTs[0], plan.LUT)
	if err != nil {
		return nil, fmt.Errorf("radix.Lt: LUT: %w", err)
	}
	gPrefix, err := e.runGPScan(gpOuts[0], gpOuts[1], plan.Spec, plan.Rounds)
	if err != nil {
		return nil, fmt.Errorf("radix.Lt: scan: %w", err)
	}

	return &LtResult{Spec: plan.Spec, CT: gPrefix}, nil
}

func (e *Evaluator) DecryptLt(res *LtResult) ([]bool, error) {
	return e.decodeBits(res)
}
