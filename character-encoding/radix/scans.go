package radix

import (
	"fmt"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"
)

// Kogge-Stone (G, P) prefix scan; returns G_prefix only. One mul level per round.
func (e *Evaluator) runGPScan(gIn, pIn *rlwe.Ciphertext, spec PaddedSpec, rounds int) (*rlwe.Ciphertext, error) {
	gCur := gIn
	pCur := pIn
	for r := 0; r < rounds; r++ {
		step := 1 << r
		ksShift := positiveMod(-step*spec.Stride, spec.MaxSlots)

		w := e.BLT.GetWorker()
		gShifted, err := w.Eval.RotateNew(gCur, ksShift)
		if err != nil {
			e.BLT.PutWorker(w)
			return nil, fmt.Errorf("KS round %d rotate g: %w", r, err)
		}
		pShifted, err := w.Eval.RotateNew(pCur, ksShift)
		if err != nil {
			e.BLT.PutWorker(w)
			return nil, fmt.Errorf("KS round %d rotate p: %w", r, err)
		}
		level := minInt(gCur.Level(), gShifted.Level())
		gNext := hefloat.NewCiphertext(e.Ctx.Params, 1, level)
		if err := w.Eval.MulRelin(pCur, gShifted, gNext); err != nil {
			e.BLT.PutWorker(w)
			return nil, fmt.Errorf("KS round %d mul g: %w", r, err)
		}
		if err := w.Eval.Rescale(gNext, gNext); err != nil {
			e.BLT.PutWorker(w)
			return nil, fmt.Errorf("KS round %d rescale g: %w", r, err)
		}
		gPrev := gCur
		if gPrev.Level() > gNext.Level() {
			e.Ctx.Evaluator.DropLevel(gPrev, gPrev.Level()-gNext.Level())
		}
		if err := w.Eval.Add(gNext, gPrev, gNext); err != nil {
			e.BLT.PutWorker(w)
			return nil, fmt.Errorf("KS round %d add g: %w", r, err)
		}
		if r < rounds-1 {
			pNext := hefloat.NewCiphertext(e.Ctx.Params, 1, level)
			if err := w.Eval.MulRelin(pCur, pShifted, pNext); err != nil {
				e.BLT.PutWorker(w)
				return nil, fmt.Errorf("KS round %d mul p: %w", r, err)
			}
			if err := w.Eval.Rescale(pNext, pNext); err != nil {
				e.BLT.PutWorker(w)
				return nil, fmt.Errorf("KS round %d rescale p: %w", r, err)
			}
			pCur = pNext
		}
		e.BLT.PutWorker(w)
		gCur = gNext
	}
	return gCur, nil
}

// Prefix-multiplication scan over pEq.
func (e *Evaluator) runEqScan(pEq *rlwe.Ciphertext, spec PaddedSpec, rounds int) (*rlwe.Ciphertext, error) {
	cur := pEq
	for r := 0; r < rounds; r++ {
		step := 1 << r
		ksShift := positiveMod(-step*spec.Stride, spec.MaxSlots)
		w := e.BLT.GetWorker()
		shifted, err := w.Eval.RotateNew(cur, ksShift)
		if err != nil {
			e.BLT.PutWorker(w)
			return nil, fmt.Errorf("KS round %d rotate: %w", r, err)
		}
		next := hefloat.NewCiphertext(e.Ctx.Params, 1, minInt(cur.Level(), shifted.Level()))
		if err := w.Eval.MulRelin(cur, shifted, next); err != nil {
			e.BLT.PutWorker(w)
			return nil, fmt.Errorf("KS round %d mul: %w", r, err)
		}
		if err := w.Eval.Rescale(next, next); err != nil {
			e.BLT.PutWorker(w)
			return nil, fmt.Errorf("KS round %d rescale: %w", r, err)
		}
		e.BLT.PutWorker(w)
		cur = next
	}
	return cur, nil
}
