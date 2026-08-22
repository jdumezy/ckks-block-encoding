package lut

import (
	"fmt"
	"sync"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

type Evaluator struct {
	Ctx *charctx.Context
	BLT *blt.Evaluator
}

func NewEvaluator(ctx *charctx.Context) *Evaluator {
	return &Evaluator{Ctx: ctx, BLT: blt.NewEvaluator(ctx)}
}

func NewEvaluatorWithWorkerCapacity(ctx *charctx.Context, capacity int) *Evaluator {
	return &Evaluator{Ctx: ctx, BLT: blt.NewEvaluatorWithCapacity(ctx, capacity)}
}

func (e *Evaluator) EvalUnary(x charenc.CipherBlock, lut *CompiledUnary) (charenc.CipherBlock, error) {
	return e.BLT.Apply(x, lut.Transform)
}

func (e *Evaluator) EvalBinary(x, y charenc.CipherBlock, lut *CompiledTensor) (charenc.CipherBlock, error) {
	return e.evalTensor([]charenc.CipherBlock{x, y}, lut)
}

func (e *Evaluator) parallelApplyRaw(inputs []*rlwe.Ciphertext, plans []*blt.RawCompiled, dst []*rlwe.Ciphertext) error {
	n := len(plans)
	if n == 0 {
		return nil
	}
	if n == 1 {
		w := e.BLT.GetWorker()
		defer e.BLT.PutWorker(w)
		out, err := e.BLT.ApplyRawWith(w, inputs[0], plans[0])
		if err != nil {
			return err
		}
		dst[0] = out
		return nil
	}
	var wg sync.WaitGroup
	errs := make([]error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			w := e.BLT.GetWorker()
			defer e.BLT.PutWorker(w)
			out, err := e.BLT.ApplyRawWith(w, inputs[i], plans[i])
			if err != nil {
				errs[i] = err
				return
			}
			dst[i] = out
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Evaluator) parallelMulRelin(lefts, rights []*rlwe.Ciphertext, dst []*rlwe.Ciphertext) error {
	n := len(lefts)
	if n == 0 {
		return nil
	}
	if n == 1 {
		w := e.BLT.GetWorker()
		defer e.BLT.PutWorker(w)
		prod := hefloat.NewCiphertext(e.Ctx.Params, 1, min(lefts[0].Level(), rights[0].Level()))
		if err := w.Eval.MulRelin(lefts[0], rights[0], prod); err != nil {
			return err
		}
		if err := w.Eval.Rescale(prod, prod); err != nil {
			return err
		}
		dst[0] = prod
		return nil
	}
	var wg sync.WaitGroup
	errs := make([]error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			w := e.BLT.GetWorker()
			defer e.BLT.PutWorker(w)
			prod := hefloat.NewCiphertext(e.Ctx.Params, 1, min(lefts[i].Level(), rights[i].Level()))
			if err := w.Eval.MulRelin(lefts[i], rights[i], prod); err != nil {
				errs[i] = err
				return
			}
			if err := w.Eval.Rescale(prod, prod); err != nil {
				errs[i] = err
				return
			}
			dst[i] = prod
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Evaluator) EvalArity3(x, y, z charenc.CipherBlock, lut *CompiledTensor) (charenc.CipherBlock, error) {
	return e.evalTensor([]charenc.CipherBlock{x, y, z}, lut)
}

func (e *Evaluator) EvalArity4(a, b, c, d charenc.CipherBlock, lut *CompiledTensor) (charenc.CipherBlock, error) {
	return e.evalTensor([]charenc.CipherBlock{a, b, c, d}, lut)
}

func (e *Evaluator) EvalTensor(inputs []charenc.CipherBlock, lut *CompiledTensor) (charenc.CipherBlock, error) {
	return e.evalTensor(inputs, lut)
}

func (e *Evaluator) evalTensor(inputs []charenc.CipherBlock, lut *CompiledTensor) (charenc.CipherBlock, error) {
	if len(inputs) != lut.Arity {
		return charenc.CipherBlock{}, fmt.Errorf("lut.evalTensor: got %d inputs, expected %d", len(inputs), lut.Arity)
	}
	for i, in := range inputs {
		if in.Spec != lut.Table.In[i] {
			return charenc.CipherBlock{}, fmt.Errorf("lut.evalTensor: input %d spec mismatch", i)
		}
		if len(in.CTs) != 1 {
			return charenc.CipherBlock{}, fmt.Errorf("lut.evalTensor: input %d has %d ciphertexts, expected 1", i, len(in.CTs))
		}
	}

	e.Ctx.EnsureGaloisKeys(lut.AllGaloisEls)

	reg := make([]*rlwe.Ciphertext, lut.Arity+len(lut.MulSchedule))

	spreadInputs := make([]*rlwe.Ciphertext, lut.Arity)
	for i, in := range inputs {
		spreadInputs[i] = in.CTs[0]
	}
	if err := e.parallelApplyRaw(spreadInputs, lut.Spreads, reg[:lut.Arity]); err != nil {
		return charenc.CipherBlock{}, fmt.Errorf("lut.evalTensor: spread: %w", err)
	}

	for _, group := range lut.MulLevels {
		lefts := make([]*rlwe.Ciphertext, len(group))
		rights := make([]*rlwe.Ciphertext, len(group))
		for j, stepIdx := range group {
			step := lut.MulSchedule[stepIdx]
			lefts[j] = reg[step.Left]
			rights[j] = reg[step.Right]
		}
		prods := make([]*rlwe.Ciphertext, len(group))
		if err := e.parallelMulRelin(lefts, rights, prods); err != nil {
			return charenc.CipherBlock{}, fmt.Errorf("lut.evalTensor: mul stage: %w", err)
		}
		for j, stepIdx := range group {
			reg[lut.Arity+stepIdx] = prods[j]
		}
	}

	packed := reg[len(reg)-1]
	w := e.BLT.GetWorker()
	defer e.BLT.PutWorker(w)
	out, err := e.BLT.ApplyRawWith(w, packed, lut.FinalLT)
	if err != nil {
		return charenc.CipherBlock{}, fmt.Errorf("lut.evalTensor: final LT: %w", err)
	}
	return charenc.CipherBlock{
		Spec:   lut.Table.Out,
		Layout: inputs[0].Layout,
		CTs:    []*rlwe.Ciphertext{out},
	}, nil
}
