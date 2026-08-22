package crt

import (
	"fmt"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

type CompiledPackedClean struct {
	In    PackedSpec
	IND   PackedSpec
	ToIND *CompiledPackedUnary
	From  *CompiledPackedUnary
}

func CompilePackedClean(ctx *charctx.Context, in PackedSpec, inputLevel int, opts blt.CompileOptions) (*CompiledPackedClean, error) {
	rescale := ctx.Params.LevelsConsumedPerRescaling()
	if inputLevel < 4*rescale {
		return nil, fmt.Errorf("crt.CompilePackedClean: input level %d too low for four-level cleaning", inputLevel)
	}
	ind, err := SwitchPackedSpec(in, charenc.IND)
	if err != nil {
		return nil, fmt.Errorf("crt.CompilePackedClean: IND spec: %w", err)
	}
	toIND, err := CompilePackedSwitchWithOptions(ctx, in, ind, inputLevel, opts)
	if err != nil {
		return nil, fmt.Errorf("crt.CompilePackedClean: to IND: %w", err)
	}
	fromIND, err := CompilePackedSwitchWithOptions(ctx, ind, in, inputLevel-3*rescale, opts)
	if err != nil {
		return nil, fmt.Errorf("crt.CompilePackedClean: from IND: %w", err)
	}
	return &CompiledPackedClean{In: in, IND: ind, ToIND: toIND, From: fromIND}, nil
}

func (e *Evaluator) EvalPackedClean(x PackedCiphertext, compiled *CompiledPackedClean, parallel bool) (PackedCiphertext, error) {
	if !samePackedSpec(x.Spec, compiled.In) {
		return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedClean: input spec mismatch")
	}
	ind, err := e.EvalPackedSwitch(x, compiled.ToIND, parallel)
	if err != nil {
		return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedClean: to IND: %w", err)
	}
	cleanIND := PackedCiphertext{Spec: compiled.IND, CTs: make([]*rlwe.Ciphertext, len(ind.CTs))}
	workers := channelParallelism(parallel, len(ind.CTs), e.LUT.BLT.Capacity())
	if workers == 1 {
		w := e.LUT.BLT.GetWorker()
		for i, ct := range ind.CTs {
			cleaned, err := e.evalINDCleanCiphertext(w, ct)
			if err != nil {
				e.LUT.BLT.PutWorker(w)
				return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedClean: clean ciphertext %d: %w", i, err)
			}
			cleanIND.CTs[i] = cleaned
		}
		e.LUT.BLT.PutWorker(w)
	} else {
		errs := make([]error, len(ind.CTs))
		done := make(chan struct{}, workers)
		for workerID := 0; workerID < workers; workerID++ {
			go func(workerID int) {
				w := e.LUT.BLT.GetWorker()
				defer e.LUT.BLT.PutWorker(w)
				defer func() { done <- struct{}{} }()
				for i := workerID; i < len(ind.CTs); i += workers {
					cleaned, err := e.evalINDCleanCiphertext(w, ind.CTs[i])
					if err != nil {
						errs[i] = err
						return
					}
					cleanIND.CTs[i] = cleaned
				}
			}(workerID)
		}
		for i := 0; i < workers; i++ {
			<-done
		}
		for i, err := range errs {
			if err != nil {
				return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedClean: clean ciphertext %d: %w", i, err)
			}
		}
	}
	out, err := e.EvalPackedSwitch(cleanIND, compiled.From, parallel)
	if err != nil {
		return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedClean: from IND: %w", err)
	}
	return out, nil
}

func (e *Evaluator) evalINDCleanCiphertext(w *blt.Worker, x *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	x2 := hefloat.NewCiphertext(e.Ctx.Params, 1, x.Level())
	if err := w.Eval.MulRelin(x, x, x2); err != nil {
		return nil, err
	}
	if err := w.Eval.Rescale(x2, x2); err != nil {
		return nil, err
	}
	xAtX2 := x
	if xAtX2.Level() > x2.Level() {
		xAtX2 = w.Eval.DropLevelNew(xAtX2, xAtX2.Level()-x2.Level())
	}
	x3 := hefloat.NewCiphertext(e.Ctx.Params, 1, x2.Level())
	if err := w.Eval.MulRelin(x2, xAtX2, x3); err != nil {
		return nil, err
	}
	if err := w.Eval.Rescale(x3, x3); err != nil {
		return nil, err
	}
	x2AtX3 := x2
	if x2AtX3.Level() > x3.Level() {
		x2AtX3 = w.Eval.DropLevelNew(x2AtX3, x2AtX3.Level()-x3.Level())
	}
	term2 := hefloat.NewCiphertext(e.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(x2AtX3, 3, term2); err != nil {
		return nil, err
	}
	term3 := hefloat.NewCiphertext(e.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(x3, 2, term3); err != nil {
		return nil, err
	}
	out := hefloat.NewCiphertext(e.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Sub(term2, term3, out); err != nil {
		return nil, err
	}
	return out, nil
}
