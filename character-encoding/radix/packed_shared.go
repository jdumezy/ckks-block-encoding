package radix

import (
	"fmt"
	"sync"

	"github.com/Pro7ech/lattigo/he"
	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

// N bivariate LUTs sharing spread + basis-mul; FinalLT differs per output.
type CompiledPackedSharedBivariate struct {
	Spec         PaddedSpec
	R            int
	Spread0      *blt.RawCompiled
	Spread1      *blt.RawCompiled
	FinalLTs     []*blt.RawCompiled
	InputLevel   int
	PackedLevel  int
	AllGaloisEls []uint64
}

// eval[i] returns R raw slot values for output i over the (v0, v1) alphabet.
func CompilePackedSharedBivariate(
	ctx *charctx.Context,
	spec PaddedSpec,
	R int,
	evals []func(v0, v1 int) []complex128,
	inputLevel int,
) (*CompiledPackedSharedBivariate, error) {
	rescale := ctx.Params.LevelsConsumedPerRescaling()
	totalDepth := 3 * rescale
	if inputLevel < totalDepth {
		return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: needs %d levels, input has %d", totalDepth, inputLevel)
	}
	if R <= 0 || R > spec.Stride {
		return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: invalid R=%d (Stride=%d)", R, spec.Stride)
	}
	if len(evals) == 0 {
		return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: at least one eval required")
	}

	b := spec.Base.Radix
	B := spec.BlockSlots
	N := spec.Stride
	K := spec.BlocksPerCT
	slots := spec.MaxSlots
	if K*N > slots {
		return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: K*N=%d > MaxSlots=%d", K*N, slots)
	}

	perBlockMat0, perBlockBias0 := augmentedSpreadMatrixLocal([]int{B, B}, 0)
	perBlockMat1, perBlockBias1 := augmentedSpreadMatrixLocal([]int{B, B}, 1)

	buildSpreadDiagonals := func(mat [][]complex128) he.Diagonals[complex128] {
		diagonals := he.Diagonals[complex128]{}
		for r := 0; r < N; r++ {
			for c := 0; c < B; c++ {
				v := mat[r][c]
				if v == 0 {
					continue
				}
				d := positiveMod(c-r, slots)
				row, ok := diagonals[d]
				if !ok {
					row = make([]complex128, slots)
					diagonals[d] = row
				}
				for k := 0; k < K; k++ {
					row[k*N+r] = v
				}
			}
		}
		return diagonals
	}
	expandBias := func(bias []complex128) []complex128 {
		out := make([]complex128, K*N)
		for k := 0; k < K; k++ {
			for r := 0; r < N; r++ {
				out[k*N+r] = bias[r]
			}
		}
		return out
	}

	sp0, err := blt.CompileDiagonals(buildSpreadDiagonals(perBlockMat0), false, expandBias(perBlockBias0), K*N, K*N, ctx, inputLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: spread0: %w", err)
	}
	sp1, err := blt.CompileDiagonals(buildSpreadDiagonals(perBlockMat1), false, expandBias(perBlockBias1), K*N, K*N, ctx, inputLevel)
	if err != nil {
		return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: spread1: %w", err)
	}

	bru, err := charenc.NewBRU(b, true)
	if err != nil {
		return nil, err
	}
	enc := make([][]complex128, b)
	for v := 0; v < b; v++ {
		enc[v] = bru.EncodeValue(v)
	}
	totalInputs := b * b
	basis := make([][]complex128, totalInputs)
	for vp := 0; vp < totalInputs; vp++ {
		v0 := vp / b
		v1 := vp % b
		row := make([]complex128, N)
		for fp := 0; fp < N; fp++ {
			a := unpackIndexLocal(fp, []int{1 + B, 1 + B})
			prod := complex(1, 0)
			if a[0] != 0 {
				prod *= enc[v0][a[0]-1]
			}
			if a[1] != 0 {
				prod *= enc[v1][a[1]-1]
			}
			row[fp] = prod
		}
		basis[vp] = row
	}

	packedLevel := inputLevel - 2*rescale

	basisKey := fmt.Sprintf("packedSharedBivariate/bru-%d-bru-%d/reduced/B=%d", b, b, B)
	basisLU, err := charenc.LUFactorizeCached(basisKey, basis)
	if err != nil {
		return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: factorize basis: %w", err)
	}

	finalLTs := make([]*blt.RawCompiled, len(evals))
	for outIdx, eval := range evals {
		perBlockCoef := make([][]complex128, R)
		for r := range perBlockCoef {
			perBlockCoef[r] = make([]complex128, N)
		}
		for r := 0; r < R; r++ {
			yr := make([]complex128, totalInputs)
			for vp := 0; vp < totalInputs; vp++ {
				v0 := vp / b
				v1 := vp % b
				slotsForVP := eval(v0, v1)
				if len(slotsForVP) != R {
					return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: eval %d returned %d slots, expected %d", outIdx, len(slotsForVP), R)
				}
				yr[vp] = slotsForVP[r]
			}
			sol, err := basisLU.Solve(yr)
			if err != nil {
				return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: eval %d solve r=%d: %w", outIdx, r, err)
			}
			copy(perBlockCoef[r], sol)
		}

		finalDiags := he.Diagonals[complex128]{}
		for r := 0; r < R; r++ {
			for n := 0; n < N; n++ {
				v := perBlockCoef[r][n]
				if v == 0 {
					continue
				}
				d := positiveMod(n-r, slots)
				row, ok := finalDiags[d]
				if !ok {
					row = make([]complex128, slots)
					finalDiags[d] = row
				}
				for k := 0; k < K; k++ {
					row[k*N+r] = v
				}
			}
		}
		finalLT, err := blt.CompileDiagonals(finalDiags, false, nil, K*N, K*N, ctx, packedLevel)
		if err != nil {
			return nil, fmt.Errorf("radix.CompilePackedSharedBivariate: eval %d finalLT: %w", outIdx, err)
		}
		finalLTs[outIdx] = finalLT
	}

	galSet := map[uint64]struct{}{}
	for _, g := range sp0.GaloisEls {
		galSet[g] = struct{}{}
	}
	for _, g := range sp1.GaloisEls {
		galSet[g] = struct{}{}
	}
	for _, finalLT := range finalLTs {
		for _, g := range finalLT.GaloisEls {
			galSet[g] = struct{}{}
		}
	}
	galEls := make([]uint64, 0, len(galSet))
	for g := range galSet {
		galEls = append(galEls, g)
	}
	ctx.EnsureGaloisKeys(galEls)

	return &CompiledPackedSharedBivariate{
		Spec:         spec,
		R:            R,
		Spread0:      sp0,
		Spread1:      sp1,
		FinalLTs:     finalLTs,
		InputLevel:   inputLevel,
		PackedLevel:  packedLevel,
		AllGaloisEls: galEls,
	}, nil
}

func (e *Evaluator) evalPackedSharedBivariate(in0, in1 *rlwe.Ciphertext, plan *CompiledPackedSharedBivariate) ([]*rlwe.Ciphertext, error) {
	e.Ctx.EnsureGaloisKeys(plan.AllGaloisEls)

	var spread0, spread1 *rlwe.Ciphertext
	var err0, err1 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		spread0, err0 = e.BLT.ApplyRaw(in0, plan.Spread0)
	}()
	go func() {
		defer wg.Done()
		spread1, err1 = e.BLT.ApplyRaw(in1, plan.Spread1)
	}()
	wg.Wait()
	if err0 != nil {
		return nil, fmt.Errorf("radix.evalPackedSharedBivariate: spread0: %w", err0)
	}
	if err1 != nil {
		return nil, fmt.Errorf("radix.evalPackedSharedBivariate: spread1: %w", err1)
	}

	w := e.BLT.GetWorker()
	level := minInt(spread0.Level(), spread1.Level())
	basis := hefloat.NewCiphertext(e.Ctx.Params, 1, level)
	if err := w.Eval.MulRelin(spread0, spread1, basis); err != nil {
		e.BLT.PutWorker(w)
		return nil, fmt.Errorf("radix.evalPackedSharedBivariate: mul: %w", err)
	}
	if err := w.Eval.Rescale(basis, basis); err != nil {
		e.BLT.PutWorker(w)
		return nil, fmt.Errorf("radix.evalPackedSharedBivariate: rescale: %w", err)
	}
	e.BLT.PutWorker(w)

	outs := make([]*rlwe.Ciphertext, len(plan.FinalLTs))
	errs := make([]error, len(plan.FinalLTs))
	var wg2 sync.WaitGroup
	wg2.Add(len(plan.FinalLTs))
	for i := range plan.FinalLTs {
		i := i
		go func() {
			defer wg2.Done()
			w := e.BLT.GetWorker()
			defer e.BLT.PutWorker(w)
			out, err := e.BLT.ApplyRawWith(w, basis, plan.FinalLTs[i])
			if err != nil {
				errs[i] = err
				return
			}
			outs[i] = out
		}()
	}
	wg2.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("radix.evalPackedSharedBivariate: finalLT: %w", err)
		}
	}
	return outs, nil
}
