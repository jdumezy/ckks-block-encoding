package radix

import (
	"sync"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
)

// Scale is set so ct-pt mul + rescale lands at the default scale.
func encodeConstantPlaintext(ctx *charctx.Context, values []complex128, level int) (*rlwe.Plaintext, error) {
	defaultScale := ctx.Params.DefaultScale()
	pt := hefloat.NewPlaintext(ctx.Params, level)
	pt.Scale = ctx.Params.GetScalingFactor(defaultScale, defaultScale, level)
	if err := ctx.Encoder.Encode(values, pt); err != nil {
		return nil, err
	}
	return pt, nil
}

type Evaluator struct {
	Ctx *charctx.Context
	BLT *blt.Evaluator

	maskMu sync.Mutex
	masks  map[block0MaskKey]*rlwe.Plaintext
}

type block0MaskKey struct {
	d, batchSize, wordStride, stride, blockSlots, maxSlots, level int
}

func NewEvaluator(ctx *charctx.Context) *Evaluator {
	return &Evaluator{Ctx: ctx, BLT: blt.NewEvaluator(ctx), masks: map[block0MaskKey]*rlwe.Plaintext{}}
}

func NewEvaluatorWithWorkerCapacity(ctx *charctx.Context, capacity int) *Evaluator {
	return &Evaluator{Ctx: ctx, BLT: blt.NewEvaluatorWithCapacity(ctx, capacity), masks: map[block0MaskKey]*rlwe.Plaintext{}}
}

type GaloisKeyHolder interface {
	GaloisElements() []uint64
}

func (c *CompiledAdd) GaloisElements() []uint64 { return c.AllGaloisEls }
func (c *CompiledSub) GaloisElements() []uint64 { return c.AllGaloisEls }
func (c *CompiledEq) GaloisElements() []uint64  { return c.AllGaloisEls }
func (c *CompiledLt) GaloisElements() []uint64  { return c.AllGaloisEls }
func (c *CompiledCmp) GaloisElements() []uint64 { return c.AllGaloisEls }

// Registers the union of Galois keys for all plans in one pass; avoids
// per-op EnsureGaloisKeys calls that each invalidate the worker pool.
func (e *Evaluator) PrewarmGaloisKeys(plans ...GaloisKeyHolder) {
	seen := map[uint64]struct{}{}
	union := make([]uint64, 0)
	for _, p := range plans {
		for _, g := range p.GaloisElements() {
			if _, ok := seen[g]; ok {
				continue
			}
			seen[g] = struct{}{}
			union = append(union, g)
		}
	}
	if len(union) > 0 {
		e.Ctx.EnsureGaloisKeys(union)
	}
}

func (e *Evaluator) block0Mask(spec PaddedSpec, level int) (*rlwe.Plaintext, error) {
	key := block0MaskKey{
		d:          spec.Base.Width,
		batchSize:  spec.BatchSize,
		wordStride: spec.WordStride,
		stride:     spec.Stride,
		blockSlots: spec.BlockSlots,
		maxSlots:   spec.MaxSlots,
		level:      level,
	}
	e.maskMu.Lock()
	if pt, ok := e.masks[key]; ok {
		e.maskMu.Unlock()
		return pt, nil
	}
	e.maskMu.Unlock()
	pt, err := buildPackedBlock0Mask(e.Ctx, spec, level)
	if err != nil {
		return nil, err
	}
	e.maskMu.Lock()
	if existing, ok := e.masks[key]; ok {
		e.maskMu.Unlock()
		return existing, nil
	}
	e.masks[key] = pt
	e.maskMu.Unlock()
	return pt, nil
}
