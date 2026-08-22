package radix

import (
	"fmt"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

// R and N are the per-block output and input slot counts.
func compileSingleBlockBLT(ctx *charctx.Context, spec PaddedSpec, perBlockMat [][]complex128, perBlockBias []complex128, R, N int, inputLevel int) (*blt.RawCompiled, error) {
	slots := spec.MaxSlots
	K := spec.BlocksPerCT

	diag := map[int][]complex128{}
	addDiag := func(d int) []complex128 {
		row, ok := diag[d]
		if !ok {
			row = make([]complex128, slots)
			diag[d] = row
		}
		return row
	}
	for r := 0; r < R; r++ {
		for c := 0; c < N; c++ {
			v := perBlockMat[r][c]
			if v == 0 {
				continue
			}
			d := positiveMod(c-r, slots)
			row := addDiag(d)
			for k := 0; k < K; k++ {
				row[k*spec.Stride+r] = v
			}
		}
	}
	bias := make([]complex128, slots)
	if perBlockBias != nil {
		for k := 0; k < K; k++ {
			for r := 0; r < R; r++ {
				bias[k*spec.Stride+r] = perBlockBias[r]
			}
		}
	}

	return blt.CompileDiagonals(diag, false, bias, K*spec.Stride, K*spec.Stride, ctx, inputLevel)
}

// Applies the 3x^2 - 2x^3 indicator sharpener in IND, switching BRU<->IND
// either side. Depth: 4 levels.
type CompiledClean struct {
	Spec         PaddedSpec
	ToIND        *blt.RawCompiled
	FromIND      *blt.RawCompiled
	InputLevel   int
	OutputLevel  int
	AllGaloisEls []uint64
}

func (c *CompiledClean) GaloisElements() []uint64 { return c.AllGaloisEls }

func (e *Evaluator) CompileClean(spec PaddedSpec, inputLevel int) (*CompiledClean, error) {
	rescale := e.Ctx.Params.LevelsConsumedPerRescaling()
	if inputLevel < 4*rescale {
		return nil, fmt.Errorf("radix.CompileClean: input level %d below 4 rescales", inputLevel)
	}
	t := spec.Base.Radix
	bruT, err := charenc.NewBRU(t, true)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileClean: BRU codec: %w", err)
	}
	indT, err := charenc.NewIND(t, true, 0)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileClean: IND codec: %w", err)
	}

	toIND, err := compilePerDigitSwitch(e.Ctx, spec, bruT, indT, inputLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileClean: BRU->IND: %w", err)
	}
	sharpLevel := inputLevel - rescale - 2*rescale
	fromIND, err := compilePerDigitSwitch(e.Ctx, spec, indT, bruT, sharpLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompileClean: IND->BRU: %w", err)
	}

	galSet := map[uint64]struct{}{}
	for _, g := range toIND.GaloisEls {
		galSet[g] = struct{}{}
	}
	for _, g := range fromIND.GaloisEls {
		galSet[g] = struct{}{}
	}
	galEls := make([]uint64, 0, len(galSet))
	for g := range galSet {
		galEls = append(galEls, g)
	}
	if len(galEls) > 0 {
		e.Ctx.EnsureGaloisKeys(galEls)
	}

	return &CompiledClean{
		Spec:         spec,
		ToIND:        toIND,
		FromIND:      fromIND,
		InputLevel:   inputLevel,
		OutputLevel:  sharpLevel - rescale,
		AllGaloisEls: galEls,
	}, nil
}

func (e *Evaluator) Clean(x PaddedCiphertext, plan *CompiledClean) (PaddedCiphertext, error) {
	if !samePaddedSpec(x.Spec, plan.Spec) {
		return PaddedCiphertext{}, fmt.Errorf("radix.Clean: spec mismatch")
	}
	if len(x.CTs) != 1 {
		return PaddedCiphertext{}, fmt.Errorf("radix.Clean: multi-CT layout not supported")
	}
	e.Ctx.EnsureGaloisKeys(plan.AllGaloisEls)

	ind, err := e.BLT.ApplyRaw(x.CTs[0], plan.ToIND)
	if err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Clean: to IND: %w", err)
	}

	sharpened, err := e.sharpenIND(ind)
	if err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Clean: sharpen: %w", err)
	}

	out, err := e.BLT.ApplyRaw(sharpened, plan.FromIND)
	if err != nil {
		return PaddedCiphertext{}, fmt.Errorf("radix.Clean: from IND: %w", err)
	}
	return PaddedCiphertext{Spec: plan.Spec, CTs: []*rlwe.Ciphertext{out}}, nil
}

// out = 3x^2 - 2x^3; pushes values near {0,1} toward exact {0,1}.
func (e *Evaluator) sharpenIND(x *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	w := e.BLT.GetWorker()
	defer e.BLT.PutWorker(w)

	x2 := hefloat.NewCiphertext(e.Ctx.Params, 1, x.Level())
	if err := w.Eval.MulRelin(x, x, x2); err != nil {
		return nil, fmt.Errorf("x^2 mul: %w", err)
	}
	if err := w.Eval.Rescale(x2, x2); err != nil {
		return nil, fmt.Errorf("x^2 rescale: %w", err)
	}

	xAtX2 := x
	if xAtX2.Level() > x2.Level() {
		xAtX2 = w.Eval.DropLevelNew(xAtX2, xAtX2.Level()-x2.Level())
	}
	x3 := hefloat.NewCiphertext(e.Ctx.Params, 1, x2.Level())
	if err := w.Eval.MulRelin(x2, xAtX2, x3); err != nil {
		return nil, fmt.Errorf("x^3 mul: %w", err)
	}
	if err := w.Eval.Rescale(x3, x3); err != nil {
		return nil, fmt.Errorf("x^3 rescale: %w", err)
	}

	x2AtX3 := x2
	if x2AtX3.Level() > x3.Level() {
		x2AtX3 = w.Eval.DropLevelNew(x2AtX3, x2AtX3.Level()-x3.Level())
	}
	term2 := hefloat.NewCiphertext(e.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(x2AtX3, 3, term2); err != nil {
		return nil, fmt.Errorf("3*x^2: %w", err)
	}
	term3 := hefloat.NewCiphertext(e.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(x3, 2, term3); err != nil {
		return nil, fmt.Errorf("2*x^3: %w", err)
	}
	out := hefloat.NewCiphertext(e.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Sub(term2, term3, out); err != nil {
		return nil, fmt.Errorf("sub: %w", err)
	}
	return out, nil
}

func compilePerDigitSwitch(ctx *charctx.Context, spec PaddedSpec, in, out charenc.Codec, inputLevel int) (*blt.RawCompiled, error) {
	t := spec.Base.Radix
	values := make([][]complex128, t)
	for v := 0; v < t; v++ {
		values[v] = out.EncodeValue(v)
	}
	A, bias, err := in.Interpolate(values, out.Spec())
	if err != nil {
		return nil, fmt.Errorf("compilePerDigitSwitch: interpolate: %w", err)
	}
	return compileSingleBlockBLT(ctx, spec, A, bias, out.Spec().Slots, in.Spec().Slots, inputLevel)
}
