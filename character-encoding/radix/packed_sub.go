package radix

import (
	"fmt"
	"math/cmplx"
	"sync"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/charctx"
)

// Packed subtractor. rawDiff = x * conj(y) (BRU(v) -> BRU(-v mod base)).
// HeadroomDelta lifts xs headroom from BRU(0) to BRU(1) so Sub's P_sub
// (which is 1 on (0,0)) does not propagate a borrow across batched words.
// Depth: 4 + ceil(log2d).
type CompiledSub struct {
	Spec          PaddedSpec
	LUT           *CompiledPackedSharedBivariate
	Block0Mask    *rlwe.Plaintext
	HeadroomDelta *rlwe.Plaintext
	InputLevel    int
	OutputLevel   int
	Rounds        int
	ConjGalEl     uint64
	AllGaloisEls  []uint64
}

func (e *Evaluator) CompileSub(spec PaddedSpec, inputLevel int) (*CompiledSub, error) {
	rescale := e.Ctx.Params.LevelsConsumedPerRescaling()
	base := spec.Base.Radix
	d := spec.Base.Width
	rounds := ceilLog2(d)

	maxStep := kogglesStoneMaxStep(d)
	if spec.WordStride < d+maxStep {
		return nil, fmt.Errorf("radix.CompileSub: WordStride=%d < d+maxStep=%d", spec.WordStride, d+maxStep)
	}
	if spec.BatchSize*spec.WordStride > spec.BlocksPerCT {
		return nil, fmt.Errorf("radix.CompileSub: BatchSize*WordStride=%d > BlocksPerCT=%d",
			spec.BatchSize*spec.WordStride, spec.BlocksPerCT)
	}

	B := spec.BlockSlots
	zeta := cmplx.Exp(complex(0, 2*3.141592653589793238462643383279502884/float64(base)))
	// deltaMaskNeg[c] = conj(zeta^{c+1}) - 1; deltaMaskPos[c] = BRU(1)[c] - BRU(0)[c].
	deltaMaskNeg := make([]complex128, B)
	deltaMaskPos := make([]complex128, B)
	for c := 0; c < B; c++ {
		power := complex(1, 0)
		for i := 0; i < c+1; i++ {
			power *= zeta
		}
		deltaMaskNeg[c] = cmplx.Conj(power) - 1
		deltaMaskPos[c] = power - 1
	}

	genEval := func(v0, v1 int) []complex128 {
		out := make([]complex128, B)
		if v0 < v1 {
			copy(out, deltaMaskNeg)
		}
		return out
	}
	propEval := func(v0, v1 int) []complex128 {
		out := make([]complex128, B)
		if v0 == v1 {
			for c := 0; c < B; c++ {
				out[c] = 1
			}
		}
		return out
	}

	lut, err := CompilePackedSharedBivariate(e.Ctx, spec, B,
		[]func(v0, v1 int) []complex128{genEval, propEval}, inputLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileSub: shared LUT: %w", err)
	}

	rawDiffLevel := inputLevel - rescale
	block0Mask, err := e.block0Mask(spec, rawDiffLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileSub: block0 mask: %w", err)
	}

	headroomDelta, err := buildPackedHeadroomDelta(e.Ctx, spec, deltaMaskPos, inputLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileSub: headroom delta: %w", err)
	}

	preShift := positiveMod(-spec.Stride, spec.MaxSlots)
	preShiftGal := e.Ctx.Params.GaloisElement(preShift)
	conjGal := e.Ctx.Params.GaloisElementForComplexConjugation()
	e.Ctx.EnsureGaloisKeys([]uint64{preShiftGal, conjGal})

	ksGals := make([]uint64, 0, rounds)
	for r := 0; r < rounds; r++ {
		shift := positiveMod(-(1<<r)*spec.Stride, spec.MaxSlots)
		ksGals = append(ksGals, e.Ctx.Params.GaloisElement(shift))
	}
	e.Ctx.EnsureGaloisKeys(ksGals)

	galSet := map[uint64]struct{}{preShiftGal: {}, conjGal: {}}
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

	outputLevel := inputLevel - (3+rounds+1)*rescale
	return &CompiledSub{
		Spec:          spec,
		LUT:           lut,
		Block0Mask:    block0Mask,
		HeadroomDelta: headroomDelta,
		InputLevel:    inputLevel,
		OutputLevel:   outputLevel,
		Rounds:        rounds,
		ConjGalEl:     conjGal,
		AllGaloisEls:  galEls,
	}, nil
}

// Writes BRU(1)-BRU(0) into the slots xs reads from BRU(0) padding after
// preshift (block 0 and blocks d+1..WordStride-1 of every word).
func buildPackedHeadroomDelta(ctx *charctx.Context, spec PaddedSpec, deltaPos []complex128, level int) (*rlwe.Plaintext, error) {
	values := make([]complex128, ctx.Params.MaxSlots())
	d := spec.Base.Width
	writeBlock := func(blockIdx int) {
		base := blockIdx * spec.Stride
		for c := 0; c < spec.BlockSlots; c++ {
			values[base+c] = deltaPos[c]
		}
	}
	for w := 0; w < spec.BatchSize; w++ {
		writeBlock(w * spec.WordStride)
		for off := d + 1; off < spec.WordStride; off++ {
			writeBlock(w*spec.WordStride + off)
		}
	}
	pt := hefloat.NewPlaintext(ctx.Params, level)
	if err := ctx.Encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return pt, nil
}

func (e *Evaluator) Sub(x, y PaddedCiphertext, plan *CompiledSub) (PaddedCiphertext, error) {
	if !samePaddedSpec(x.Spec, plan.Spec) || !samePaddedSpec(y.Spec, plan.Spec) {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: spec mismatch")
	}
	if len(x.CTs) != plan.Spec.CTCount || len(y.CTs) != plan.Spec.CTCount {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: CT count mismatch")
	}
	if plan.Spec.CTCount != 1 {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: multi-CT layout not yet supported")
	}

	e.Ctx.EnsureGaloisKeys(plan.AllGaloisEls)
	xCT, yCT := x.CTs[0], y.CTs[0]
	preShift := positiveMod(-plan.Spec.Stride, plan.Spec.MaxSlots)

	var rawDiff *rlwe.Ciphertext
	var gpOuts []*rlwe.Ciphertext
	var errDiff, errLUT error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		w := e.BLT.GetWorker()
		defer e.BLT.PutWorker(w)
		conjY, err := w.Eval.ConjugateNew(yCT)
		if err != nil {
			errDiff = fmt.Errorf("conjugate y: %w", err)
			return
		}
		rawDiff = hefloat.NewCiphertext(e.Ctx.Params, 1, minInt(xCT.Level(), conjY.Level()))
		if err := w.Eval.MulRelin(xCT, conjY, rawDiff); err != nil {
			errDiff = err
			return
		}
		if err := w.Eval.Rescale(rawDiff, rawDiff); err != nil {
			errDiff = err
		}
	}()
	go func() {
		defer wg.Done()
		w1 := e.BLT.GetWorker()
		xs, err := w1.Eval.RotateNew(xCT, preShift)
		if err != nil {
			e.BLT.PutWorker(w1)
			errLUT = fmt.Errorf("rotate x: %w", err)
			return
		}
		if err := w1.Eval.Add(xs, plan.HeadroomDelta, xs); err != nil {
			e.BLT.PutWorker(w1)
			errLUT = fmt.Errorf("headroom delta: %w", err)
			return
		}
		e.BLT.PutWorker(w1)
		w2 := e.BLT.GetWorker()
		ys, err := w2.Eval.RotateNew(yCT, preShift)
		e.BLT.PutWorker(w2)
		if err != nil {
			errLUT = fmt.Errorf("rotate y: %w", err)
			return
		}
		gpOuts, errLUT = e.evalPackedSharedBivariate(xs, ys, plan.LUT)
	}()
	wg.Wait()
	if errDiff != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: rawDiff: %w", errDiff)
	}
	if errLUT != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: shared LUT: %w", errLUT)
	}
	if len(gpOuts) != 2 {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: expected 2 LUT outputs, got %d", len(gpOuts))
	}
	gPrefix, err := e.runGPScan(gpOuts[0], gpOuts[1], plan.Spec, plan.Rounds)
	if err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: scan: %w", err)
	}

	wFinal := e.BLT.GetWorker()
	defer e.BLT.PutWorker(wFinal)

	delta := hefloat.NewCiphertext(e.Ctx.Params, 1, rawDiff.Level())
	if err := wFinal.Eval.Mul(rawDiff, plan.Block0Mask, delta); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: delta mask mul: %w", err)
	}
	if err := wFinal.Eval.Rescale(delta, delta); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: delta rescale: %w", err)
	}

	commonLevel := minInt(gPrefix.Level(), delta.Level())
	if gPrefix.Level() > commonLevel {
		e.Ctx.Evaluator.DropLevel(gPrefix, gPrefix.Level()-commonLevel)
	}
	if delta.Level() > commonLevel {
		e.Ctx.Evaluator.DropLevel(delta, delta.Level()-commonLevel)
	}
	correction := hefloat.NewCiphertext(e.Ctx.Params, 1, commonLevel)
	if err := wFinal.Eval.MulRelin(gPrefix, delta, correction); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: correction mul: %w", err)
	}
	if err := wFinal.Eval.Rescale(correction, correction); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: correction rescale: %w", err)
	}

	if rawDiff.Level() > correction.Level() {
		e.Ctx.Evaluator.DropLevel(rawDiff, rawDiff.Level()-correction.Level())
	}
	if err := wFinal.Eval.Add(rawDiff, correction, rawDiff); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Sub: final add: %w", err)
	}

	return PaddedCiphertext{Spec: plan.Spec, CTs: []*rlwe.Ciphertext{rawDiff}}, nil
}
