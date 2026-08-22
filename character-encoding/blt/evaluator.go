package blt

import (
	"fmt"

	"github.com/Pro7ech/lattigo/he"
	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

// A shallow-copied hefloat.Evaluator at logN=15 holds roughly 30 MB of RNS
// buffers, so this also bounds memory.
const DefaultWorkerPoolCapacity = 8

// Workers carry the ctx.Gen at which they were created; stale ones are
// discarded when the context's evaluator is rebuilt by EnsureGaloisKeys.
type Evaluator struct {
	Ctx     *charctx.Context
	workers chan *Worker
	slots   chan struct{}
	cap     int
}

type hoistingKey struct {
	LevelQ int
	LevelP int
}

// Workers are not safe for concurrent use; each goroutine must acquire its
// own via GetWorker.
type Worker struct {
	Eval   *hefloat.Evaluator
	LTEval *he.LinearTransformationEvaluator
	hb     map[hoistingKey]rlwe.HoistingBuffer
	bsgs   map[hoistingKey]map[int]*rlwe.Ciphertext
	gen    int
}

func NewEvaluator(ctx *charctx.Context) *Evaluator {
	return NewEvaluatorWithCapacity(ctx, DefaultWorkerPoolCapacity)
}

func NewEvaluatorWithCapacity(ctx *charctx.Context, capacity int) *Evaluator {
	if capacity < 1 {
		capacity = 1
	}
	return &Evaluator{
		Ctx:     ctx,
		workers: make(chan *Worker, capacity),
		slots:   make(chan struct{}, capacity),
		cap:     capacity,
	}
}

func (e *Evaluator) Capacity() int {
	return e.cap
}

func (e *Evaluator) newWorker() *Worker {
	sub := e.Ctx.Evaluator.ShallowCopy()
	return &Worker{
		Eval:   sub,
		LTEval: he.NewLinearTransformationEvaluator(sub),
		hb:     make(map[hoistingKey]rlwe.HoistingBuffer),
		bsgs:   make(map[hoistingKey]map[int]*rlwe.Ciphertext),
		gen:    e.Ctx.Gen,
	}
}

func (e *Evaluator) GetWorker() *Worker {
	for {
		select {
		case w := <-e.workers:
			if w.gen != e.Ctx.Gen {
				e.releaseWorker()
				continue
			}
			return w
		default:
		}

		select {
		case e.slots <- struct{}{}:
			return e.newWorker()
		case w := <-e.workers:
			if w.gen != e.Ctx.Gen {
				e.releaseWorker()
				continue
			}
			return w
		}
	}
}

func (e *Evaluator) PutWorker(w *Worker) {
	if w.gen != e.Ctx.Gen {
		e.releaseWorker()
		return
	}
	select {
	case e.workers <- w:
	default:
		e.releaseWorker()
	}
}

func (e *Evaluator) releaseWorker() {
	select {
	case <-e.slots:
	default:
	}
}

func (w *Worker) hoistingBuffer(levelQ, levelP int) rlwe.HoistingBuffer {
	key := hoistingKey{LevelQ: levelQ, LevelP: levelP}
	if buf, ok := w.hb[key]; ok {
		return buf
	}
	buf := w.LTEval.NewHoistingBuffer(levelQ, levelP)
	w.hb[key] = buf
	return buf
}

func (w *Worker) bsgsPreRotations(levelQ, levelP int) map[int]*rlwe.Ciphertext {
	key := hoistingKey{LevelQ: levelQ, LevelP: levelP}
	if preRot, ok := w.bsgs[key]; ok {
		return preRot
	}
	preRot := make(map[int]*rlwe.Ciphertext)
	w.bsgs[key] = preRot
	return preRot
}

func (e *Evaluator) Apply(in charenc.CipherBlock, ct *CompiledTransform) (charenc.CipherBlock, error) {
	if in.Spec != ct.In {
		return charenc.CipherBlock{}, fmt.Errorf("blt.Apply: input spec %+v does not match compiled In %+v", in.Spec, ct.In)
	}
	if len(in.CTs) != 1 {
		return charenc.CipherBlock{}, fmt.Errorf("blt.Apply: expected 1 ciphertext for block layout, got %d", len(in.CTs))
	}
	e.Ctx.EnsureGaloisKeys(ct.GaloisEls)
	w := e.GetWorker()
	defer e.PutWorker(w)
	out, err := e.applyRawWith(w, in.CTs[0], &ct.Raw)
	if err != nil {
		return charenc.CipherBlock{}, err
	}
	return charenc.CipherBlock{
		Spec:   ct.Out,
		Layout: in.Layout,
		CTs:    []*rlwe.Ciphertext{out},
	}, nil
}

func (e *Evaluator) ApplyRaw(in *rlwe.Ciphertext, ct *RawCompiled) (*rlwe.Ciphertext, error) {
	e.Ctx.EnsureGaloisKeys(ct.GaloisEls)
	w := e.GetWorker()
	defer e.PutWorker(w)
	return e.applyRawWith(w, in, ct)
}

// Caller must pre-ensure the required Galois keys.
func (e *Evaluator) ApplyRawWith(w *Worker, in *rlwe.Ciphertext, ct *RawCompiled) (*rlwe.Ciphertext, error) {
	return e.applyRawWith(w, in, ct)
}

func (e *Evaluator) applyRawWith(w *Worker, in *rlwe.Ciphertext, ct *RawCompiled) (*rlwe.Ciphertext, error) {
	if in.Level() < ct.InputLevel {
		return nil, fmt.Errorf("blt.applyRawWith: input level %d below compiled InputLevel %d", in.Level(), ct.InputLevel)
	}
	out := hefloat.NewCiphertext(e.Ctx.Params, 1, ct.InputLevel)

	if ct.DiagPlaintext != nil {
		if err := w.Eval.Mul(in, ct.DiagPlaintext, out); err != nil {
			return nil, fmt.Errorf("blt.applyRawWith: diag mul: %w", err)
		}
	} else if ct.LT.GiantStep != -1 {
		if err := e.applyRawBSGSWith(w, in, ct, out); err != nil {
			return nil, fmt.Errorf("blt.applyRawWith: BSGS LT evaluate: %w", err)
		}
	} else {
		buf := w.hoistingBuffer(ct.InputLevel, e.Ctx.Params.MaxLevelP())
		if err := w.LTEval.Evaluate(in, ct.LT, buf, out); err != nil {
			return nil, fmt.Errorf("blt.applyRawWith: LT evaluate: %w", err)
		}
	}
	if err := w.Eval.Rescale(out, out); err != nil {
		return nil, fmt.Errorf("blt.applyRawWith: rescale: %w", err)
	}
	if ct.BiasPlaintext != nil {
		if err := w.Eval.Add(out, ct.BiasPlaintext, out); err != nil {
			return nil, fmt.Errorf("blt.applyRawWith: bias add: %w", err)
		}
	}
	return out, nil
}

func (e *Evaluator) applyRawBSGSWith(w *Worker, in *rlwe.Ciphertext, ct *RawCompiled, out *rlwe.Ciphertext) error {
	levelQ := min(in.Level(), ct.LT.LevelQ)
	levelP := ct.LT.LevelP
	buf := w.hoistingBuffer(levelQ, levelP)
	w.Eval.FillHoistingBuffer(levelQ, levelP, in.Q[1], in.IsNTT, buf)

	_, _, rotN2 := ct.LT.BSGSIndex()
	preRot := w.bsgsPreRotations(levelQ, levelP)
	params := w.Eval.GetRLWEParameters()
	for _, i := range rotN2 {
		if i == 0 {
			continue
		}
		rot := preRot[i]
		if rot == nil {
			rot = rlwe.NewCiphertext(params, 1, levelQ, levelP)
			preRot[i] = rot
		}
		if err := w.Eval.AutomorphismHoistedLazy(levelQ, in, buf, params.GaloisElement(i), rot); err != nil {
			return err
		}
	}
	if err := w.LTEval.MultiplyByDiagMatrixBSGS(in, ct.LT, preRot, out); err != nil {
		return err
	}
	return nil
}
