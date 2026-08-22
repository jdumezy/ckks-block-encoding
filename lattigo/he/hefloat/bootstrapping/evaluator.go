package bootstrapping

import (
	"fmt"
	"math"
	"math/big"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/ring"
	"github.com/Pro7ech/lattigo/rlwe"
)

const twoPi = 2 * math.Pi

type Evaluator struct {
	Parameters
	*hefloat.Evaluator
	DFTEvaluator  *hefloat.DFTEvaluator
	Mod1Evaluator *hefloat.Mod1Evaluator
	*EvaluationKeys

	Mod1Parameters hefloat.Mod1Parameters
	S2CDFTMatrix   *hefloat.DFTMatrix
	C2SDFTMatrix   *hefloat.DFTMatrix
}

func NewEvaluator(p Parameters, evk *EvaluationKeys) (*Evaluator, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if p.ResidualParameters.N() != p.BootstrappingParameters.N() {
		return nil, fmt.Errorf("bootstrapping: ResidualParameters and BootstrappingParameters must share the same ring degree")
	}

	eval := &Evaluator{Parameters: p}

	var err error
	if eval.Mod1Parameters, err = hefloat.NewMod1ParametersFromLiteral(p.BootstrappingParameters, p.Mod1ParametersLiteral); err != nil {
		return nil, err
	}

	params := p.BootstrappingParameters
	encoder := hefloat.NewEncoder(params)

	K := eval.Mod1Parameters.Mod1Interval()
	qDiff := eval.Mod1Parameters.QDiff
	qDiv := eval.Mod1Parameters.ScalingFactor().Float64() / math.Exp2(math.Round(math.Log2(float64(params.Q()[0]))))
	if qDiv > 1 {
		qDiv = 1
	}

	c2sScaling := new(big.Float).SetFloat64(qDiv / (K * qDiff))
	if p.CoeffsToSlotsParameters.Scaling == nil {
		eval.CoeffsToSlotsParameters.Scaling = c2sScaling
	} else {
		eval.CoeffsToSlotsParameters.Scaling = new(big.Float).Mul(p.CoeffsToSlotsParameters.Scaling, c2sScaling)
	}

	if p.SlotsToCoeffsParameters.Scaling == nil {
		eval.SlotsToCoeffsParameters.Scaling = new(big.Float).SetFloat64(1.0)
	}

	if eval.C2SDFTMatrix, err = hefloat.NewDFTMatrixFromLiteral(params, eval.CoeffsToSlotsParameters, encoder); err != nil {
		return nil, err
	}
	if eval.S2CDFTMatrix, err = hefloat.NewDFTMatrixFromLiteral(params, eval.SlotsToCoeffsParameters, encoder); err != nil {
		return nil, err
	}

	eval.EvaluationKeys = evk
	eval.Evaluator = hefloat.NewEvaluator(params, evk.MemEvaluationKeySet)
	eval.DFTEvaluator = hefloat.NewDFTEvaluator(params, eval.Evaluator)
	polyEval := hefloat.NewPolynomialEvaluator(params, eval.Evaluator)
	eval.Mod1Evaluator = hefloat.NewMod1Evaluator(eval.Evaluator, polyEval, eval.Mod1Parameters)

	return eval, nil
}

func (eval *Evaluator) newHoistingBuffer(levelQ, levelP int) rlwe.HoistingBuffer {
	return eval.Evaluator.Evaluator.NewHoistingBuffer(levelQ, levelP)
}

func (eval *Evaluator) ScaleDown(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Scale, error) {
	params := eval.BootstrappingParameters
	r := params.RingQ()

	for ctIn.Level() != 0 && checkMessageRatio(ctIn, eval.Mod1Parameters.MessageRatio(), r) {
		ctIn.ResizeQ(ctIn.Level() - 1)
	}

	currentModulus := r.AtLevel(ctIn.Level()).Modulus()
	currentMessageRatio := rlwe.NewScale(currentModulus)
	currentMessageRatio = currentMessageRatio.Div(ctIn.Scale)

	targetMessageRatio := rlwe.NewScale(eval.Mod1Parameters.MessageRatio())
	scaleUp := currentMessageRatio.Div(targetMessageRatio)

	if scaleUp.Cmp(rlwe.NewScale(0.5)) == -1 {
		return nil, nil, fmt.Errorf("bootstrapping.ScaleDown: initial Q/Scale = %f < 0.5*Q[0]/MessageRatio = %f", currentMessageRatio.Float64(), targetMessageRatio.Float64())
	}

	scaleUpBigint := scaleUp.BigInt()
	if err := eval.Evaluator.Mul(ctIn, scaleUpBigint, ctIn); err != nil {
		return nil, nil, err
	}
	ctIn.Scale = ctIn.Scale.Mul(rlwe.NewScale(scaleUpBigint))

	targetScale := new(big.Float).SetPrec(256).SetInt(big.NewInt(int64(r[0].Modulus)))
	targetScale.Quo(targetScale, new(big.Float).SetFloat64(eval.Mod1Parameters.MessageRatio()))

	if ctIn.Level() != 0 {
		if err := eval.Evaluator.RescaleTo(ctIn, rlwe.NewScale(targetScale), ctIn); err != nil {
			return nil, nil, err
		}
	}

	errScale := ctIn.Scale.Div(rlwe.NewScale(targetScale))
	return ctIn, &errScale, nil
}

func checkMessageRatio(ct *rlwe.Ciphertext, msgRatio float64, r ring.RNSRing) bool {
	level := ct.Level()
	currentModulus := r.AtLevel(level).Modulus()
	currentMessageRatio := rlwe.NewScale(currentModulus).Div(ct.Scale)
	return currentMessageRatio.Cmp(rlwe.NewScale(r[level].Modulus).Mul(rlwe.NewScale(msgRatio))) > -1
}

func (eval *Evaluator) ModUp(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if eval.EvkDenseToSparse != nil {
		if err := eval.Evaluator.ApplyEvaluationKey(ctIn, eval.EvkDenseToSparse, ctIn); err != nil {
			return nil, err
		}
	}

	params := eval.BootstrappingParameters
	ringQ := params.RingQ().AtLevel(ctIn.Level())

	for i := range ctIn.Q {
		ringQ.INTT(ctIn.Q[i], ctIn.Q[i])
	}

	ctIn.ResizeQ(params.MaxLevel())

	levelQ := params.MaxLevel()
	ringQ = params.RingQ().AtLevel(levelQ)

	Q := ringQ.ModuliChain()
	q := Q[0]
	brcQ := ringQ.BRedConstants()

	N := ringQ.N()

	for j := 0; j < N; j++ {
		coeff := ctIn.Q[0][0][j]
		pos, neg := uint64(1), uint64(0)
		if coeff >= (q >> 1) {
			coeff = q - coeff
			pos, neg = 0, 1
		}
		for i := 1; i <= levelQ; i++ {
			tmp := ring.BRedAdd(coeff, Q[i], brcQ[i])
			ctIn.Q[0][i][j] = tmp*pos + (Q[i]-tmp)*neg
		}
	}

	if eval.EvkSparseToDense != nil {
		if err := eval.modUpSparse(ctIn, ringQ, params); err != nil {
			return nil, err
		}
	} else {
		for j := 0; j < N; j++ {
			coeff := ctIn.Q[1][0][j]
			pos, neg := uint64(1), uint64(0)
			if coeff >= (q >> 1) {
				coeff = q - coeff
				pos, neg = 0, 1
			}
			for i := 1; i <= levelQ; i++ {
				tmp := ring.BRedAdd(coeff, Q[i], brcQ[i])
				ctIn.Q[1][i][j] = tmp*pos + (Q[i]-tmp)*neg
			}
		}

		ringQ.NTT(ctIn.Q[0], ctIn.Q[0])
		ringQ.NTT(ctIn.Q[1], ctIn.Q[1])

		if scale := (eval.Mod1Parameters.ScalingFactor().Float64() / eval.Mod1Parameters.MessageRatio()) / ctIn.Scale.Float64(); scale > 1 {
			scalar := uint64(math.Round(scale))
			ringQ.MulScalar(ctIn.Q[0], scalar, ctIn.Q[0])
			ringQ.MulScalar(ctIn.Q[1], scalar, ctIn.Q[1])
			ctIn.Scale = ctIn.Scale.Mul(rlwe.NewScale(scale))
		}
	}

	return ctIn, eval.Evaluator.Trace(ctIn, eval.CoeffsToSlotsParameters.LogSlots, ctIn)
}

func (eval *Evaluator) modUpSparse(ctIn *rlwe.Ciphertext, ringQ ring.RNSRing, params hefloat.Parameters) error {
	ringP := params.RingP()
	levelQ := params.MaxLevel()
	levelP := params.MaxLevelP()

	Q := ringQ.ModuliChain()
	P := ringP.ModuliChain()
	q := Q[0]
	brcQ := ringQ.BRedConstants()
	brcP := ringP.BRedConstants()

	ks := eval.Evaluator.Evaluator
	buf := ks.NewHoistingBuffer(levelQ, levelP)

	N := ringQ.N()

	for j := 0; j < N; j++ {
		coeff := ctIn.Q[1][0][j]
		pos, neg := uint64(1), uint64(0)
		if coeff > (q >> 1) {
			coeff = q - coeff
			pos, neg = 0, 1
		}
		for i := 0; i <= levelQ; i++ {
			tmp := ring.BRedAdd(coeff, Q[i], brcQ[i])
			buf[0].Q[i][j] = tmp*pos + (Q[i]-tmp)*neg
		}
		for i := 0; i <= levelP; i++ {
			tmp := ring.BRedAdd(coeff, P[i], brcP[i])
			buf[0].P[i][j] = tmp*pos + (P[i]-tmp)*neg
		}
	}

	for i := len(buf) - 1; i >= 0; i-- {
		ringQ.NTT(buf[0].Q, buf[i].Q)
	}
	for i := len(buf) - 1; i >= 0; i-- {
		ringP.NTT(buf[0].P, buf[i].P)
	}

	ringQ.NTT(ctIn.Q[0], ctIn.Q[0])

	if scale := (eval.Mod1Parameters.ScalingFactor().Float64() / eval.Mod1Parameters.MessageRatio()) / ctIn.Scale.Float64(); scale > 1 {
		scalar := uint64(math.Round(scale))
		for i := len(buf) - 1; i >= 0; i-- {
			ringQ.MulScalar(buf[0].Q, scalar, buf[i].Q)
		}
		for i := len(buf) - 1; i >= 0; i-- {
			ringP.MulScalar(buf[0].P, scalar, buf[i].P)
		}
		ringQ.MulScalar(ctIn.Q[0], scalar, ctIn.Q[0])
		ctIn.Scale = ctIn.Scale.Mul(rlwe.NewScale(scale))
	}

	ctTmpQ := []ring.RNSPoly{ringQ.NewRNSPoly(), ctIn.Q[1]}
	ctTmp := &rlwe.Ciphertext{
		Vector:   &ring.Vector{Q: ctTmpQ},
		MetaData: ctIn.MetaData,
	}

	ks.GadgetProductHoisted(levelQ, buf, &eval.EvkSparseToDense.GadgetCiphertext, ctTmp)
	ringQ.Add(ctIn.Q[0], ctTmp.Q[0], ctIn.Q[0])

	return nil
}

func (eval *Evaluator) CoeffsToSlots(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, *rlwe.Ciphertext, error) {
	buf := eval.newHoistingBuffer(eval.C2SDFTMatrix.LevelQ, eval.C2SDFTMatrix.LevelP)
	return eval.DFTEvaluator.CoeffsToSlotsNew(ctIn, eval.C2SDFTMatrix, buf)
}

func (eval *Evaluator) SlotsToCoeffs(ctReal, ctImag *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	buf := eval.newHoistingBuffer(eval.S2CDFTMatrix.LevelQ, eval.S2CDFTMatrix.LevelP)
	return eval.DFTEvaluator.SlotsToCoeffsNew(ctReal, ctImag, eval.S2CDFTMatrix, buf)
}

func (eval *Evaluator) EvalMod(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	ct, err := eval.Mod1Evaluator.EvaluateWithAffineTransformationNew(ctIn, complex(twoPi, 0), 0)
	if err != nil {
		return nil, err
	}
	ct.Scale = eval.BootstrappingParameters.DefaultScale()
	return ct, nil
}

func (eval *Evaluator) EvalElse(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	ct, err := eval.Mod1Evaluator.EvaluateElseWithAffineTransformationNew(ctIn, complex(twoPi, 0), 0)
	if err != nil {
		return nil, err
	}
	ct.Scale = eval.BootstrappingParameters.DefaultScale()
	return ct, nil
}

func (eval *Evaluator) BootstrapToUnitCircle(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	ct, err := eval.SlotsToCoeffs(ctIn, nil)
	if err != nil {
		return nil, fmt.Errorf("bootstrapping.BootstrapToUnitCircle: S2C: %w", err)
	}

	if ct, _, err = eval.ScaleDown(ct); err != nil {
		return nil, fmt.Errorf("bootstrapping.BootstrapToUnitCircle: ScaleDown: %w", err)
	}

	if ct, err = eval.ModUp(ct); err != nil {
		return nil, fmt.Errorf("bootstrapping.BootstrapToUnitCircle: ModUp: %w", err)
	}

	ctReal, _, err := eval.CoeffsToSlots(ct)
	if err != nil {
		return nil, fmt.Errorf("bootstrapping.BootstrapToUnitCircle: C2S: %w", err)
	}

	ctSin, err := eval.EvalMod(ctReal)
	if err != nil {
		return nil, fmt.Errorf("bootstrapping.BootstrapToUnitCircle: EvalMod: %w", err)
	}

	ctCos, err := eval.EvalElse(ctReal)
	if err != nil {
		return nil, fmt.Errorf("bootstrapping.BootstrapToUnitCircle: EvalElse: %w", err)
	}

	if err = eval.Evaluator.Mul(ctSin, 1i, ctSin); err != nil {
		return nil, fmt.Errorf("bootstrapping.BootstrapToUnitCircle: mul i: %w", err)
	}

	if err = eval.Evaluator.Add(ctCos, ctSin, ctCos); err != nil {
		return nil, fmt.Errorf("bootstrapping.BootstrapToUnitCircle: add: %w", err)
	}

	return ctCos, nil
}
