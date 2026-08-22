package bootstrapping

import (
	"fmt"

	"github.com/Pro7ech/lattigo/rlwe"
)

type EvaluationKeys struct {
	*rlwe.MemEvaluationKeySet
	EvkDenseToSparse *rlwe.EvaluationKey
	EvkSparseToDense *rlwe.EvaluationKey
}

func (p Parameters) GenEvaluationKeys(sk *rlwe.SecretKey) (*EvaluationKeys, error) {
	if p.ResidualParameters.N() != p.BootstrappingParameters.N() {
		return nil, fmt.Errorf("bootstrapping: ResidualParameters and BootstrappingParameters must share the same ring degree")
	}

	params := p.BootstrappingParameters
	kgen := rlwe.NewKeyGenerator(params)

	rlk := kgen.GenRelinearizationKeyNew(sk)
	galEls := p.GaloisElements(params)
	gks := kgen.GenGaloisKeysNew(galEls, sk)

	keys := &EvaluationKeys{
		MemEvaluationKeySet: rlwe.NewMemEvaluationKeySet(rlk, gks...),
	}

	if p.EphemeralSecretWeight > 0 {
		skSparse, err := p.genSparseSecretKey(sk)
		if err != nil {
			return nil, err
		}
		keys.EvkDenseToSparse = kgen.GenEvaluationKeyNew(sk, skSparse)
		keys.EvkSparseToDense = kgen.GenEvaluationKeyNew(skSparse, sk)
	}

	return keys, nil
}

func (p Parameters) genSparseSecretKey(skDense *rlwe.SecretKey) (*rlwe.SecretKey, error) {
	params := p.BootstrappingParameters

	literal := rlwe.ParametersLiteral{
		LogN:       params.LogN(),
		Q:          params.Q()[:1],
		P:          params.P()[:1],
		NTTFlag:    true,
		RingType:   params.RingType(),
		LogNthRoot: params.LogNthRoot(),
	}

	sparseParams, err := rlwe.NewParametersFromLiteral(literal)
	if err != nil {
		return nil, fmt.Errorf("bootstrapping: build sparse params: %w", err)
	}

	kgenSparse := rlwe.NewKeyGenerator(sparseParams)
	return kgenSparse.GenSecretKeyWithHammingWeightNew(p.EphemeralSecretWeight), nil
}
