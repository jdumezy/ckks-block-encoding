package radix

import (
	"fmt"
	"sync"

	"github.com/Pro7ech/lattigo/rlwe"
)

// Returns (eq, lt, gt) per word; gt = 1 - eq - lt is level-free.
// Depth: 3 + ceil(log2d).
type CompiledCmp struct {
	Spec         PaddedSpec
	LUT          *CompiledPackedSharedBivariate
	InputLevel   int
	OutputLevel  int
	Rounds       int
	AllGaloisEls []uint64
}

func (e *Evaluator) CompileCmp(spec PaddedSpec, inputLevel int) (*CompiledCmp, error) {
	rescale := e.Ctx.Params.LevelsConsumedPerRescaling()
	d := spec.Base.Width
	rounds := ceilLog2(d)
	if (1 << rounds) != d {
		return nil, fmt.Errorf("radix.CompileCmp: d=%d must be a power of two", d)
	}

	maxStep := kogglesStoneMaxStep(d)
	if spec.WordStride < d+maxStep {
		return nil, fmt.Errorf("radix.CompileCmp: WordStride=%d < d+maxStep=%d", spec.WordStride, d+maxStep)
	}
	if spec.BatchSize*spec.WordStride > spec.BlocksPerCT {
		return nil, fmt.Errorf("radix.CompileCmp: BatchSize*WordStride=%d > BlocksPerCT=%d",
			spec.BatchSize*spec.WordStride, spec.BlocksPerCT)
	}

	pEqEval := func(v0, v1 int) []complex128 {
		if v0 == v1 {
			return []complex128{1}
		}
		return []complex128{0}
	}
	gLtEval := func(v0, v1 int) []complex128 {
		if v0 < v1 {
			return []complex128{1}
		}
		return []complex128{0}
	}
	pLtEval := func(v0, v1 int) []complex128 {
		if v0 == v1 {
			return []complex128{1}
		}
		return []complex128{0}
	}

	lut, err := CompilePackedSharedBivariate(e.Ctx, spec, 1,
		[]func(v0, v1 int) []complex128{pEqEval, gLtEval, pLtEval}, inputLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileCmp: LUT: %w", err)
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
	return &CompiledCmp{
		Spec:         spec,
		LUT:          lut,
		InputLevel:   inputLevel,
		OutputLevel:  outputLevel,
		Rounds:       rounds,
		AllGaloisEls: galEls,
	}, nil
}

// Bits live at slot (w*WordStride + d - 1) * Stride.
type CmpResult struct {
	Spec       PaddedSpec
	Eq, Lt, Gt *rlwe.Ciphertext
}

func (e *Evaluator) Cmp(x, y PaddedCiphertext, plan *CompiledCmp) (*CmpResult, error) {
	if !samePaddedSpec(x.Spec, plan.Spec) || !samePaddedSpec(y.Spec, plan.Spec) {
		return nil, fmt.Errorf("radix.Cmp: spec mismatch")
	}
	if len(x.CTs) != 1 || len(y.CTs) != 1 {
		return nil, fmt.Errorf("radix.Cmp: expected 1 ct each")
	}
	e.Ctx.EnsureGaloisKeys(plan.AllGaloisEls)

	outs, err := e.evalPackedSharedBivariate(x.CTs[0], y.CTs[0], plan.LUT)
	if err != nil {
		return nil, fmt.Errorf("radix.Cmp: LUT: %w", err)
	}
	pEq := outs[0]
	gLt := outs[1]
	pLt := outs[2]

	var eqFinal, ltFinal *rlwe.Ciphertext
	var errEq, errLt error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		eqFinal, errEq = e.runEqScan(pEq, plan.Spec, plan.Rounds)
	}()
	go func() {
		defer wg.Done()
		ltFinal, errLt = e.runGPScan(gLt, pLt, plan.Spec, plan.Rounds)
	}()
	wg.Wait()
	if errEq != nil {
		return nil, fmt.Errorf("radix.Cmp: Eq scan: %w", errEq)
	}
	if errLt != nil {
		return nil, fmt.Errorf("radix.Cmp: Lt scan: %w", errLt)
	}

	// gt = 1 - eq - lt.
	w := e.BLT.GetWorker()
	defer e.BLT.PutWorker(w)
	gt, err := w.Eval.AddNew(eqFinal, ltFinal)
	if err != nil {
		return nil, fmt.Errorf("radix.Cmp: eq+lt: %w", err)
	}
	if err := w.Eval.Mul(gt, -1.0, gt); err != nil {
		return nil, fmt.Errorf("radix.Cmp: negate: %w", err)
	}
	if err := w.Eval.Add(gt, 1.0, gt); err != nil {
		return nil, fmt.Errorf("radix.Cmp: +1: %w", err)
	}

	return &CmpResult{
		Spec: plan.Spec,
		Eq:   eqFinal,
		Lt:   ltFinal,
		Gt:   gt,
	}, nil
}

func (e *Evaluator) DecryptCmp(res *CmpResult) (eq, lt, gt []bool, err error) {
	eq, err = e.decodeBits(&PerWordBitResult{Spec: res.Spec, CT: res.Eq})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("eq: %w", err)
	}
	lt, err = e.decodeBits(&PerWordBitResult{Spec: res.Spec, CT: res.Lt})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lt: %w", err)
	}
	gt, err = e.decodeBits(&PerWordBitResult{Spec: res.Spec, CT: res.Gt})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gt: %w", err)
	}
	return eq, lt, gt, nil
}
