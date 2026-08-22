package radix

import (
	"fmt"
	"math/cmplx"
	"sync"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/charctx"
)

// Mask is 1 on BRU-coord slots of every digit except each word's LSB block,
// so no spurious carry enters digit 0 of any word.
func buildPackedBlock0Mask(ctx *charctx.Context, spec PaddedSpec, level int) (*rlwe.Plaintext, error) {
	values := make([]complex128, ctx.Params.MaxSlots())
	d := spec.Base.Width
	for w := 0; w < spec.BatchSize; w++ {
		for off := 1; off < d; off++ {
			base := (w*spec.WordStride + off) * spec.Stride
			for c := 0; c < spec.BlockSlots; c++ {
				values[base+c] = 1
			}
		}
	}
	return encodeConstantPlaintext(ctx, values, level)
}

// Packed G/P Kogge-Stone adder. Depth: 4 + ceil(log2d) levels.
type CompiledAdd struct {
	Spec         PaddedSpec
	LUT          *CompiledPackedSharedBivariate
	Block0Mask   *rlwe.Plaintext
	InputLevel   int
	OutputLevel  int
	Rounds       int
	AllGaloisEls []uint64
}

func (e *Evaluator) CompileAdd(spec PaddedSpec, inputLevel int) (*CompiledAdd, error) {
	rescale := e.Ctx.Params.LevelsConsumedPerRescaling()
	base := spec.Base.Radix
	d := spec.Base.Width
	rounds := ceilLog2(d)

	// Largest KS rotation must wrap into BRU(0) padding, not real data.
	maxStep := kogglesStoneMaxStep(d)
	if spec.WordStride < d+maxStep {
		return nil, fmt.Errorf("radix.CompileAdd: WordStride=%d < d+maxStep=%d", spec.WordStride, d+maxStep)
	}
	if spec.BatchSize*spec.WordStride > spec.BlocksPerCT {
		return nil, fmt.Errorf(
			"radix.CompileAdd: BatchSize*WordStride=%d > BlocksPerCT=%d",
			spec.BatchSize*spec.WordStride, spec.BlocksPerCT,
		)
	}

	B := spec.BlockSlots
	zeta := cmplx.Exp(complex(0, 2*3.141592653589793238462643383279502884/float64(base)))
	deltaMask := make([]complex128, B)
	for c := 0; c < B; c++ {
		power := complex(1, 0)
		for i := 0; i < c+1; i++ {
			power *= zeta
		}
		deltaMask[c] = power - 1
	}

	genEval := func(v0, v1 int) []complex128 {
		out := make([]complex128, B)
		if v0+v1 >= base {
			copy(out, deltaMask)
		}
		return out
	}
	propEval := func(v0, v1 int) []complex128 {
		out := make([]complex128, B)
		if v0+v1 == base-1 {
			for c := 0; c < B; c++ {
				out[c] = 1
			}
		}
		return out
	}

	lut, err := CompilePackedSharedBivariate(e.Ctx, spec, B,
		[]func(v0, v1 int) []complex128{genEval, propEval}, inputLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileAdd: shared LUT: %w", err)
	}

	rawSumLevel := inputLevel - rescale
	block0Mask, err := e.block0Mask(spec, rawSumLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileAdd: block0 mask: %w", err)
	}

	preShift := positiveMod(-spec.Stride, spec.MaxSlots)
	preShiftGal := e.Ctx.Params.GaloisElement(preShift)
	e.Ctx.EnsureGaloisKeys([]uint64{preShiftGal})

	ksGals := make([]uint64, 0, rounds)
	for r := 0; r < rounds; r++ {
		shift := positiveMod(-(1<<r)*spec.Stride, spec.MaxSlots)
		ksGals = append(ksGals, e.Ctx.Params.GaloisElement(shift))
	}
	e.Ctx.EnsureGaloisKeys(ksGals)

	galSet := map[uint64]struct{}{preShiftGal: {}}
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
	return &CompiledAdd{
		Spec:         spec,
		LUT:          lut,
		Block0Mask:   block0Mask,
		InputLevel:   inputLevel,
		OutputLevel:  outputLevel,
		Rounds:       rounds,
		AllGaloisEls: galEls,
	}, nil
}

func (e *Evaluator) Add(x, y PaddedCiphertext, plan *CompiledAdd) (PaddedCiphertext, error) {
	if !samePaddedSpec(x.Spec, plan.Spec) || !samePaddedSpec(y.Spec, plan.Spec) {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: spec mismatch")
	}
	if len(x.CTs) != plan.Spec.CTCount || len(y.CTs) != plan.Spec.CTCount {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: CT count mismatch")
	}
	if plan.Spec.CTCount != 1 {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: multi-CT layout not yet supported")
	}

	e.Ctx.EnsureGaloisKeys(plan.AllGaloisEls)
	xCT, yCT := x.CTs[0], y.CTs[0]
	preShift := positiveMod(-plan.Spec.Stride, plan.Spec.MaxSlots)

	var rawSum *rlwe.Ciphertext
	var gpOuts []*rlwe.Ciphertext
	var errSum, errLUT error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		w := e.BLT.GetWorker()
		defer e.BLT.PutWorker(w)
		rawSum = hefloat.NewCiphertext(e.Ctx.Params, 1, minInt(xCT.Level(), yCT.Level()))
		if err := w.Eval.MulRelin(xCT, yCT, rawSum); err != nil {
			errSum = err
			return
		}
		if err := w.Eval.Rescale(rawSum, rawSum); err != nil {
			errSum = err
		}
	}()
	go func() {
		defer wg.Done()
		// Pre-shift by -Stride so the LUT emits carry-into directly.
		w1 := e.BLT.GetWorker()
		xs, err := w1.Eval.RotateNew(xCT, preShift)
		e.BLT.PutWorker(w1)
		if err != nil {
			errLUT = fmt.Errorf("rotate x: %w", err)
			return
		}
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
	if errSum != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: rawSum: %w", errSum)
	}
	if errLUT != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: shared LUT: %w", errLUT)
	}
	if len(gpOuts) != 2 {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: expected 2 LUT outputs, got %d", len(gpOuts))
	}
	gPrefix, err := e.runGPScan(gpOuts[0], gpOuts[1], plan.Spec, plan.Rounds)
	if err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: scan: %w", err)
	}

	wFinal := e.BLT.GetWorker()
	defer e.BLT.PutWorker(wFinal)

	delta := hefloat.NewCiphertext(e.Ctx.Params, 1, rawSum.Level())
	if err := wFinal.Eval.Mul(rawSum, plan.Block0Mask, delta); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: delta mask mul: %w", err)
	}
	if err := wFinal.Eval.Rescale(delta, delta); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: delta rescale: %w", err)
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
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: correction mul: %w", err)
	}
	if err := wFinal.Eval.Rescale(correction, correction); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: correction rescale: %w", err)
	}

	if rawSum.Level() > correction.Level() {
		e.Ctx.Evaluator.DropLevel(rawSum, rawSum.Level()-correction.Level())
	}
	if err := wFinal.Eval.Add(rawSum, correction, rawSum); err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Add: final add: %w", err)
	}

	return PaddedCiphertext{Spec: plan.Spec, CTs: []*rlwe.Ciphertext{rawSum}}, nil
}
