package crt

import (
	"fmt"
	"sync"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
	"character-encoding/character-encoding/lut"
)

type Evaluator struct {
	Ctx *charctx.Context
	LUT *lut.Evaluator
}

func NewEvaluator(ctx *charctx.Context, workerCapacity int) *Evaluator {
	return &Evaluator{
		Ctx: ctx,
		LUT: lut.NewEvaluatorWithWorkerCapacity(ctx, workerCapacity),
	}
}

// Consumes one level. In BRU this is CRT addition; in LBRU CRT multiplication.
func (e *Evaluator) EvalNativeProduct(x, y Ciphertext, parallel bool) (Ciphertext, error) {
	if err := checkBinaryOperands(x, y); err != nil {
		return Ciphertext{}, err
	}
	out := Ciphertext{Spec: x.Spec, Blocks: make([]charenc.CipherBlock, x.Spec.Channels())}
	workers := channelParallelism(parallel, x.Spec.Channels(), e.LUT.BLT.Capacity())
	if workers == 1 {
		w := e.LUT.BLT.GetWorker()
		defer e.LUT.BLT.PutWorker(w)
		for i := range x.Blocks {
			block, err := e.evalNativeChannel(w, x.Blocks[i], y.Blocks[i])
			if err != nil {
				return Ciphertext{}, fmt.Errorf("crt.EvalNativeProduct: channel %d: %w", i, err)
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
			w := e.LUT.BLT.GetWorker()
			defer e.LUT.BLT.PutWorker(w)
			for i := workerID; i < len(x.Blocks); i += workers {
				block, err := e.evalNativeChannel(w, x.Blocks[i], y.Blocks[i])
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
			return Ciphertext{}, fmt.Errorf("crt.EvalNativeProduct: channel %d: %w", i, err)
		}
	}
	return out, nil
}

func (e *Evaluator) evalNativeChannel(w *blt.Worker, x, y charenc.CipherBlock) (charenc.CipherBlock, error) {
	if len(x.CTs) != 1 || len(y.CTs) != 1 {
		return charenc.CipherBlock{}, fmt.Errorf("expected one ciphertext per channel")
	}
	level := min(x.CTs[0].Level(), y.CTs[0].Level())
	outCT := hefloat.NewCiphertext(e.Ctx.Params, 1, level)
	if err := w.Eval.MulRelin(x.CTs[0], y.CTs[0], outCT); err != nil {
		return charenc.CipherBlock{}, err
	}
	if err := w.Eval.Rescale(outCT, outCT); err != nil {
		return charenc.CipherBlock{}, err
	}
	return charenc.CipherBlock{Spec: x.Spec, Layout: x.Layout, CTs: []*rlwe.Ciphertext{outCT}}, nil
}

func (e *Evaluator) EvalSwitch(x Ciphertext, compiled *CompiledUnary, parallel bool) (Ciphertext, error) {
	if x.Spec.Channels() != compiled.In.Channels() {
		return Ciphertext{}, fmt.Errorf("crt.EvalSwitch: channel count mismatch")
	}
	if len(x.Blocks) != x.Spec.Channels() {
		return Ciphertext{}, fmt.Errorf("crt.EvalSwitch: ciphertext block count mismatch")
	}
	if compiled.Out.Channels() != compiled.In.Channels() {
		return Ciphertext{}, fmt.Errorf("crt.EvalSwitch: compiled channel count mismatch")
	}
	out := Ciphertext{Spec: compiled.Out, Blocks: make([]charenc.CipherBlock, compiled.Out.Channels())}
	workers := channelParallelism(parallel, x.Spec.Channels(), e.LUT.BLT.Capacity())
	if workers == 1 {
		for i := range x.Blocks {
			block, err := e.LUT.EvalUnary(x.Blocks[i], compiled.Channels[i])
			if err != nil {
				return Ciphertext{}, fmt.Errorf("crt.EvalSwitch: channel %d: %w", i, err)
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
			return Ciphertext{}, fmt.Errorf("crt.EvalSwitch: channel %d: %w", i, err)
		}
	}
	return out, nil
}

type CompiledBinary struct {
	Spec     Spec
	Channels []*lut.CompiledTensor
}

func CompileBinary(ctx *charctx.Context, s Spec, f func(p int, x, y int) int, inputLevel int) (*CompiledBinary, error) {
	channels := make([]*lut.CompiledTensor, s.Channels())
	for i, p := range s.Primes {
		table := lut.MultiTable{
			In:  []charenc.BlockSpec{s.Specs[i], s.Specs[i]},
			Out: s.Specs[i],
			Eval: func(xs []int) int {
				return f(p, xs[0], xs[1])
			},
		}
		compiled, err := lut.CompileBinary(table, ctx, inputLevel)
		if err != nil {
			return nil, fmt.Errorf("crt.CompileBinary: channel %d p=%d: %w", i, p, err)
		}
		channels[i] = compiled
	}
	return &CompiledBinary{Spec: s, Channels: channels}, nil
}

func (e *Evaluator) EvalBinaryLUT(x, y Ciphertext, compiled *CompiledBinary, parallel bool) (Ciphertext, error) {
	if err := checkBinaryOperands(x, y); err != nil {
		return Ciphertext{}, err
	}
	if compiled.Spec.Channels() != x.Spec.Channels() {
		return Ciphertext{}, fmt.Errorf("crt.EvalBinaryLUT: compiled channel count mismatch")
	}
	out := Ciphertext{Spec: x.Spec, Blocks: make([]charenc.CipherBlock, x.Spec.Channels())}
	workers := channelParallelism(parallel, x.Spec.Channels(), e.LUT.BLT.Capacity())
	if workers == 1 {
		for i := range x.Blocks {
			block, err := e.LUT.EvalBinary(x.Blocks[i], y.Blocks[i], compiled.Channels[i])
			if err != nil {
				return Ciphertext{}, fmt.Errorf("crt.EvalBinaryLUT: channel %d: %w", i, err)
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
				block, err := e.LUT.EvalBinary(x.Blocks[i], y.Blocks[i], compiled.Channels[i])
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
			return Ciphertext{}, fmt.Errorf("crt.EvalBinaryLUT: channel %d: %w", i, err)
		}
	}
	return out, nil
}

func channelParallelism(parallel bool, channels, capacity int) int {
	if !parallel || channels <= 1 {
		return 1
	}
	if capacity < 1 {
		capacity = 1
	}
	return min(channels, capacity)
}

func checkBinaryOperands(x, y Ciphertext) error {
	if x.Spec.Channels() != y.Spec.Channels() {
		return fmt.Errorf("CRT channel count mismatch")
	}
	if len(x.Blocks) != x.Spec.Channels() || len(y.Blocks) != y.Spec.Channels() {
		return fmt.Errorf("CRT ciphertext block count mismatch")
	}
	for i := range x.Blocks {
		if x.Blocks[i].Spec != y.Blocks[i].Spec {
			return fmt.Errorf("channel %d spec mismatch", i)
		}
	}
	return nil
}
