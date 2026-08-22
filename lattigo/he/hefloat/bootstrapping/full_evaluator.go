package bootstrapping

import (
	"fmt"
	"math"
	"math/big"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"
)

type FullEvaluator struct {
	Btp       *Evaluator
	Params    hefloat.Parameters
	Mod1Par   hefloat.Mod1Parameters
	C2SMat    *hefloat.DFTMatrix
	S2CMat    *hefloat.DFTMatrix
	BypassMat *hefloat.DFTMatrix
	Evk       *rlwe.MemEvaluationKeySet
	EvalRound bool

	ev     *hefloat.Evaluator
	dft    *hefloat.DFTEvaluator
	mod1Ev *hefloat.Mod1Evaluator
}

func NewFullEvaluator(btp *Evaluator, btpParams Parameters, sk *rlwe.SecretKey, evalRound bool) (*FullEvaluator, error) {
	params := btpParams.BootstrappingParameters

	mod1Par, err := hefloat.NewMod1ParametersFromLiteral(params, btpParams.Mod1ParametersLiteral)
	if err != nil {
		return nil, fmt.Errorf("FullEvaluator: Mod1ParametersFromLiteral: %w", err)
	}

	qDiff := mod1Par.QDiff
	qDiv := mod1Par.ScalingFactor().Float64() / math.Exp2(math.Round(math.Log2(float64(params.Q()[0]))))
	if qDiv > 1 {
		qDiv = 1
	}
	c2sScaling := new(big.Float).SetFloat64(qDiv / (mod1Par.Mod1Interval() * qDiff))
	scale := params.DefaultScale().Float64()
	offset := mod1Par.ScalingFactor().Float64() / mod1Par.MessageRatio()
	stCScaling := new(big.Float).SetFloat64(scale / offset)

	encoder := hefloat.NewEncoder(params)

	c2sLit := btpParams.CoeffsToSlotsParameters
	if c2sLit.Scaling == nil {
		c2sLit.Scaling = c2sScaling
	} else {
		c2sLit.Scaling = new(big.Float).Mul(c2sLit.Scaling, c2sScaling)
	}
	c2sMat, err := hefloat.NewDFTMatrixFromLiteral(params, c2sLit, encoder)
	if err != nil {
		return nil, fmt.Errorf("FullEvaluator: C2S matrix: %w", err)
	}

	s2cLit := btpParams.SlotsToCoeffsParameters
	if s2cLit.Scaling == nil {
		s2cLit.Scaling = stCScaling
	} else {
		s2cLit.Scaling = new(big.Float).Mul(s2cLit.Scaling, stCScaling)
	}
	s2cMat, err := hefloat.NewDFTMatrixFromLiteral(params, s2cLit, encoder)
	if err != nil {
		return nil, fmt.Errorf("FullEvaluator: S2C matrix: %w", err)
	}

	var bypassMat *hefloat.DFTMatrix
	bypassLit := buildBypassLiteral(c2sLit, mod1Par, s2cLit.LevelQ, params.MaxLevelP())
	if evalRound {
		bypassLit.Scaling = new(big.Float).SetFloat64(qDiv)
		bypassMat, err = hefloat.NewDFTMatrixFromLiteral(params, bypassLit, encoder)
		if err != nil {
			return nil, fmt.Errorf("FullEvaluator: Bypass matrix: %w", err)
		}
	}

	galSet := map[uint64]struct{}{}
	for _, g := range c2sLit.GaloisElements(params) {
		galSet[g] = struct{}{}
	}
	for _, g := range s2cLit.GaloisElements(params) {
		galSet[g] = struct{}{}
	}
	if evalRound {
		for _, g := range bypassLit.GaloisElements(params) {
			galSet[g] = struct{}{}
		}
	}
	galSet[params.GaloisElementForComplexConjugation()] = struct{}{}

	kgen := rlwe.NewKeyGenerator(params)
	missing := []uint64{}
	for g := range galSet {
		if _, err := btp.GetGaloisKey(g); err != nil {
			missing = append(missing, g)
		}
	}
	allKeys := []*rlwe.GaloisKey{}
	for _, k := range btp.MemEvaluationKeySet.GaloisKeys {
		allKeys = append(allKeys, k)
	}
	for _, k := range kgen.GenGaloisKeysNew(missing, sk) {
		allKeys = append(allKeys, k)
	}
	rlk, err := btp.GetRelinearizationKey()
	if err != nil {
		return nil, fmt.Errorf("FullEvaluator: missing relinearization key: %w", err)
	}
	evk := rlwe.NewMemEvaluationKeySet(rlk, allKeys...)

	ev := hefloat.NewEvaluator(params, evk)
	dft := hefloat.NewDFTEvaluator(params, ev)
	polyEv := hefloat.NewPolynomialEvaluator(params, ev)
	mod1Ev := hefloat.NewMod1Evaluator(ev, polyEv, mod1Par)

	return &FullEvaluator{
		Btp:       btp,
		Params:    params,
		Mod1Par:   mod1Par,
		C2SMat:    c2sMat,
		S2CMat:    s2cMat,
		BypassMat: bypassMat,
		Evk:       evk,
		EvalRound: evalRound,
		ev:        ev,
		dft:       dft,
		mod1Ev:    mod1Ev,
	}, nil
}

func buildBypassLiteral(c2s hefloat.DFTMatrixLiteral, mod1Par hefloat.Mod1Parameters, s2cLevelQ, levelP int) hefloat.DFTMatrixLiteral {
	depth := mod1Par.Mod1Poly.Depth() + mod1Par.DoubleAngle
	maxLogSlots := c2s.LogSlots
	if maxLogSlots < 1 {
		maxLogSlots = 1
	}
	bypassDepth := depth
	if bypassDepth > maxLogSlots {
		bypassDepth = maxLogSlots
	}
	if bypassDepth < 1 {
		bypassDepth = 1
	}
	levels := make([]int, bypassDepth)
	for i := range levels {
		levels[i] = 1
	}
	return hefloat.DFTMatrixLiteral{
		Type:     hefloat.HomomorphicEncode,
		Format:   hefloat.RepackImagAsReal,
		LogSlots: c2s.LogSlots,
		LevelQ:   s2cLevelQ + bypassDepth,
		LevelP:   levelP,
		Levels:   levels,
	}
}

func (f *FullEvaluator) Bootstrap(ctIn *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	ev := f.ev
	dft := f.dft
	mod1Ev := f.mod1Ev

	ct := ctIn.Clone()
	var err error

	if ct, _, err = f.Btp.ScaleDown(ct); err != nil {
		return nil, fmt.Errorf("FullEvaluator.Bootstrap: ScaleDown: %w", err)
	}
	if ct, err = f.Btp.ModUp(ct); err != nil {
		return nil, fmt.Errorf("FullEvaluator.Bootstrap: ModUp: %w", err)
	}

	in := ct.Clone()

	c2sBuf := ev.NewHoistingBuffer(f.C2SMat.LevelQ, f.C2SMat.LevelP)
	real0, imag0, err := dft.CoeffsToSlotsNew(ct, f.C2SMat, c2sBuf)
	if err != nil {
		return nil, fmt.Errorf("FullEvaluator.Bootstrap: CoeffsToSlots: %w", err)
	}

	var ctOut *rlwe.Ciphertext
	if f.EvalRound && f.BypassMat != nil {
		K := f.Mod1Par.Mod1Interval()
		QDiff := f.Mod1Par.QDiff

		real1, err := mod1Ev.EvaluateNew(real0)
		if err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: Mod1 real: %w", err)
		}
		if err = ev.Mul(real0, K*QDiff, real0); err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: Mul real: %w", err)
		}
		if err = ev.Rescale(real0, real0); err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: Rescale real: %w", err)
		}
		if err = ev.Sub(real1, real0, real1); err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: Sub real: %w", err)
		}
		if imag0 != nil {
			imag1, err := mod1Ev.EvaluateNew(imag0)
			if err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: Mod1 imag: %w", err)
			}
			if err = ev.Mul(imag0, K*QDiff, imag0); err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: Mul imag: %w", err)
			}
			if err = ev.Rescale(imag0, imag0); err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: Rescale imag: %w", err)
			}
			if err = ev.Sub(imag1, imag0, imag1); err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: Sub imag: %w", err)
			}
			if err = ev.Mul(imag1, 1i, imag1); err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: Mul i: %w", err)
			}
			ctOut, err = ev.AddNew(real1, imag1)
			if err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: AddNew: %w", err)
			}
		} else {
			ctOut = real1
		}

		bypassBuf := ev.NewHoistingBuffer(f.BypassMat.LevelQ, f.BypassMat.LevelP)
		realB, imagB, err := dft.CoeffsToSlotsNew(in, f.BypassMat, bypassBuf)
		if err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: Bypass C2S: %w", err)
		}
		var in0 *rlwe.Ciphertext
		if imagB != nil {
			in0, err = ev.MulNew(imagB, 1i)
			if err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: bypass mul i: %w", err)
			}
			if err = ev.Add(in0, realB, in0); err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: bypass add: %w", err)
			}
		} else {
			in0 = realB
		}
		if err = ev.Add(in0, ctOut, in0); err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: combine add: %w", err)
		}

		s2cBuf := ev.NewHoistingBuffer(f.S2CMat.LevelQ, f.S2CMat.LevelP)
		ctOut, err = dft.SlotsToCoeffsNew(in0, nil, f.S2CMat, s2cBuf)
		if err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: SlotsToCoeffs: %w", err)
		}
	} else {
		real0, err = mod1Ev.EvaluateNew(real0)
		if err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: Mod1 real: %w", err)
		}
		real0.Scale = f.Params.DefaultScale()
		if imag0 != nil {
			imag0, err = mod1Ev.EvaluateNew(imag0)
			if err != nil {
				return nil, fmt.Errorf("FullEvaluator.Bootstrap: Mod1 imag: %w", err)
			}
			imag0.Scale = f.Params.DefaultScale()
		}
		s2cBuf := ev.NewHoistingBuffer(f.S2CMat.LevelQ, f.S2CMat.LevelP)
		ctOut, err = dft.SlotsToCoeffsNew(real0, imag0, f.S2CMat, s2cBuf)
		if err != nil {
			return nil, fmt.Errorf("FullEvaluator.Bootstrap: SlotsToCoeffs: %w", err)
		}
	}

	ctOut.Scale = f.Params.DefaultScale()
	return ctOut, nil
}
