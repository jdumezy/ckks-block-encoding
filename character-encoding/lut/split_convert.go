package lut

import (
	"fmt"
	"math"
	"math/cmplx"
	"sync"

	"github.com/Pro7ech/lattigo/he"
	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/he/hefloat/bootstrapping"
	"github.com/Pro7ech/lattigo/rlwe"
	"github.com/Pro7ech/lattigo/utils/bignum"

	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

var twoPi = 2 * math.Pi

type SplitToStandardPlan struct {
	coeff []*rlwe.Plaintext
	bias  complex128
	level int
}

func CompileSplitToStandard(ctx *charctx.Context, spec charenc.BlockSpec, level int) (*SplitToStandardPlan, error) {
	codec, err := charenc.NewCodec(spec)
	if err != nil {
		return nil, err
	}
	t := spec.Alphabet.Modulus
	J := spec.Slots

	matrix := make([][]complex128, t)
	rhs := make([]complex128, t)
	for k := 0; k < t; k++ {
		row := make([]complex128, J+1)
		encoded := codec.EncodeValue(k)
		for c := 0; c < J; c++ {
			row[c] = encoded[c]
		}
		row[J] = 1
		matrix[k] = row
		rhs[k] = complex(float64(k), 0)
	}

	sol, err := solveDenseLinearSystem(matrix, rhs)
	if err != nil {
		return nil, fmt.Errorf("CompileSplitToStandard: %w", err)
	}

	slots := ctx.Params.MaxSlots()
	defaultScale := ctx.Params.DefaultScale()
	ptScale := ctx.Params.GetScalingFactor(defaultScale, defaultScale, level)

	coeff := make([]*rlwe.Plaintext, J)
	for c := 0; c < J; c++ {
		v := sol[c]
		if v == 0 {
			continue
		}
		vec := make([]complex128, slots)
		for i := range vec {
			vec[i] = v
		}
		pt := hefloat.NewPlaintext(ctx.Params, level)
		pt.Scale = ptScale
		if err := ctx.Encoder.Encode(vec, pt); err != nil {
			return nil, fmt.Errorf("CompileSplitToStandard: encode coef %d: %w", c, err)
		}
		coeff[c] = pt
	}

	return &SplitToStandardPlan{coeff: coeff, bias: sol[J], level: level}, nil
}

func EvalSplitToStandard(ev *Evaluator, splitCTs []*rlwe.Ciphertext, plan *SplitToStandardPlan) (*rlwe.Ciphertext, error) {
	if len(splitCTs) != len(plan.coeff) {
		return nil, fmt.Errorf("EvalSplitToStandard: got %d split ciphertexts, plan expects %d", len(splitCTs), len(plan.coeff))
	}

	w := ev.BLT.GetWorker()
	defer ev.BLT.PutWorker(w)

	var acc *rlwe.Ciphertext
	first := true
	for c, pt := range plan.coeff {
		if pt == nil || splitCTs[c] == nil {
			continue
		}
		in := splitCTs[c]
		if in.Level() > plan.level {
			in = w.Eval.DropLevelNew(in, in.Level()-plan.level)
		}
		tmp := hefloat.NewCiphertext(ev.Ctx.Params, 1, plan.level)
		if err := w.Eval.Mul(in, pt, tmp); err != nil {
			return nil, fmt.Errorf("EvalSplitToStandard: mul coord %d: %w", c, err)
		}
		if first {
			acc = tmp
			first = false
		} else {
			if err := w.Eval.Add(acc, tmp, acc); err != nil {
				return nil, fmt.Errorf("EvalSplitToStandard: add coord %d: %w", c, err)
			}
		}
	}
	if first {
		return nil, fmt.Errorf("EvalSplitToStandard: every coefficient is zero")
	}
	if plan.bias != 0 {
		if err := w.Eval.Add(acc, plan.bias, acc); err != nil {
			return nil, fmt.Errorf("EvalSplitToStandard: add bias: %w", err)
		}
	}
	if err := w.Eval.Rescale(acc, acc); err != nil {
		return nil, fmt.Errorf("EvalSplitToStandard: rescale: %w", err)
	}
	return acc, nil
}

type SplitFromStandardPlan struct {
	spec charenc.BlockSpec
	t    int
	K    int
}

func CompileSplitFromStandard(spec charenc.BlockSpec) (*SplitFromStandardPlan, error) {
	if spec.Alphabet.Kind != charenc.BRU {
		return nil, fmt.Errorf("CompileSplitFromStandard: only BRU is supported as a direct target; use a follow-up encoding switch for %v", spec.Alphabet.Kind)
	}
	if !spec.Reduced {
		return nil, fmt.Errorf("CompileSplitFromStandard: only reduced BRU is supported")
	}
	return &SplitFromStandardPlan{spec: spec, t: spec.Alphabet.Modulus, K: 16}, nil
}

func EvalSplitFromStandard(btp *bootstrapping.Evaluator, standardCT *rlwe.Ciphertext, plan *SplitFromStandardPlan) ([]*rlwe.Ciphertext, error) {
	ct, err := btp.SlotsToCoeffs(standardCT, nil)
	if err != nil {
		return nil, fmt.Errorf("EvalSplitFromStandard: SlotsToCoeffs: %w", err)
	}
	if ct, _, err = btp.ScaleDown(ct); err != nil {
		return nil, fmt.Errorf("EvalSplitFromStandard: ScaleDown: %w", err)
	}
	if ct, err = btp.ModUp(ct); err != nil {
		return nil, fmt.Errorf("EvalSplitFromStandard: ModUp: %w", err)
	}
	ctReal, _, err := btp.CoeffsToSlots(ct)
	if err != nil {
		return nil, fmt.Errorf("EvalSplitFromStandard: CoeffsToSlots: %w", err)
	}

	J := plan.spec.Slots
	K := plan.K
	if K > J {
		K = J
	}

	direct, err := evalDirectCoords(btp, ctReal, K)
	if err != nil {
		return nil, err
	}

	if K == J {
		return direct, nil
	}

	qMaxNeeded := 0
	for c := K + 1; c <= J; c++ {
		r := c % K
		q := c / K
		if r == 0 {
			q -= 1
		}
		if q > qMaxNeeded {
			qMaxNeeded = q
		}
	}

	pb := he.NewPowerBasis(direct[K-1], bignum.Monomial)
	for q := 2; q <= qMaxNeeded; q++ {
		if err := pb.GenPower(q, false, btp.Evaluator); err != nil {
			return nil, fmt.Errorf("EvalSplitFromStandard: GenPower(%d) for stripe: %w", q, err)
		}
	}

	out := make([]*rlwe.Ciphertext, J)
	copy(out, direct)

	for c := K + 1; c <= J; c++ {
		r := c % K
		q := c / K
		if r == 0 {
			r = K
			q -= 1
		}
		qPow, ok := pb.Value[q]
		if !ok || qPow == nil {
			return nil, fmt.Errorf("EvalSplitFromStandard: missing power %d in basis", q)
		}
		prod, err := btp.Evaluator.MulRelinNew(direct[r-1], qPow)
		if err != nil {
			return nil, fmt.Errorf("EvalSplitFromStandard: cross-mul c=%d: %w", c, err)
		}
		if err := btp.Evaluator.Rescale(prod, prod); err != nil {
			return nil, fmt.Errorf("EvalSplitFromStandard: rescale c=%d: %w", c, err)
		}
		out[c-1] = prod
	}

	return out, nil
}

func evalDirectCoords(btp *bootstrapping.Evaluator, ctReal *rlwe.Ciphertext, K int) ([]*rlwe.Ciphertext, error) {
	direct := make([]*rlwe.Ciphertext, K)
	errs := make([]error, K)
	var wg sync.WaitGroup
	wg.Add(K)
	for i := 0; i < K; i++ {
		i := i
		go func() {
			defer wg.Done()
			eval := btp.Evaluator.ShallowCopy()
			polyEval := hefloat.NewPolynomialEvaluator(btp.BootstrappingParameters, eval)
			mod1Eval := hefloat.NewMod1Evaluator(eval, polyEval, btp.Mod1Parameters)
			tmp := ctReal.Clone()
			if i > 0 {
				if err := eval.Mul(tmp, int64(i+1), tmp); err != nil {
					errs[i] = fmt.Errorf("pre-scale: %w", err)
					return
				}
			}
			sin, err := mod1Eval.EvaluateWithAffineTransformationNew(tmp, complex(twoPi, 0), 0)
			if err != nil {
				errs[i] = fmt.Errorf("EvalMod: %w", err)
				return
			}
			sin.Scale = btp.BootstrappingParameters.DefaultScale()
			cos, err := mod1Eval.EvaluateElseWithAffineTransformationNew(tmp, complex(twoPi, 0), 0)
			if err != nil {
				errs[i] = fmt.Errorf("EvalElse: %w", err)
				return
			}
			cos.Scale = btp.BootstrappingParameters.DefaultScale()
			if err := eval.Mul(sin, 1i, sin); err != nil {
				errs[i] = fmt.Errorf("mul i: %w", err)
				return
			}
			if err := eval.Add(cos, sin, cos); err != nil {
				errs[i] = fmt.Errorf("recombine: %w", err)
				return
			}
			direct[i] = cos
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("evalDirectCoords[%d]: %w", i, err)
		}
	}
	return direct, nil
}

func powerBasis(btp *bootstrapping.Evaluator, base *rlwe.Ciphertext, n int) ([]*rlwe.Ciphertext, error) {
	if n < 1 {
		return nil, fmt.Errorf("powerBasis: n must be >= 1, got %d", n)
	}
	pb := he.NewPowerBasis(base, bignum.Monomial)
	for k := 2; k <= n; k++ {
		if err := pb.GenPower(k, false, btp.Evaluator); err != nil {
			return nil, fmt.Errorf("powerBasis: GenPower(%d): %w", k, err)
		}
	}
	out := make([]*rlwe.Ciphertext, n)
	for k := 1; k <= n; k++ {
		ct, ok := pb.Value[k]
		if !ok || ct == nil {
			return nil, fmt.Errorf("powerBasis: missing power %d", k)
		}
		out[k-1] = ct
	}
	return out, nil
}

func solveDenseLinearSystem(M [][]complex128, y []complex128) ([]complex128, error) {
	n := len(M)
	if len(y) != n {
		return nil, fmt.Errorf("solveDenseLinearSystem: matrix is %dx?, rhs has length %d", n, len(y))
	}

	A := make([][]complex128, n)
	for i := 0; i < n; i++ {
		if len(M[i]) != n {
			return nil, fmt.Errorf("solveDenseLinearSystem: row %d has %d cols, expected %d", i, len(M[i]), n)
		}
		A[i] = make([]complex128, n+1)
		copy(A[i], M[i])
		A[i][n] = y[i]
	}

	for i := 0; i < n; i++ {
		maxAbs := cmplx.Abs(A[i][i])
		pivot := i
		for j := i + 1; j < n; j++ {
			if a := cmplx.Abs(A[j][i]); a > maxAbs {
				maxAbs = a
				pivot = j
			}
		}
		if maxAbs == 0 {
			return nil, fmt.Errorf("solveDenseLinearSystem: singular matrix at column %d", i)
		}
		A[i], A[pivot] = A[pivot], A[i]

		for j := i + 1; j < n; j++ {
			factor := A[j][i] / A[i][i]
			for k := i; k <= n; k++ {
				A[j][k] -= factor * A[i][k]
			}
		}
	}

	x := make([]complex128, n)
	for i := n - 1; i >= 0; i-- {
		sum := A[i][n]
		for j := i + 1; j < n; j++ {
			sum -= A[i][j] * x[j]
		}
		x[i] = sum / A[i][i]
	}
	return x, nil
}
