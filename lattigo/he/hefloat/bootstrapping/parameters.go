package bootstrapping

import (
	"fmt"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/ring"
)

type Parameters struct {
	ResidualParameters      hefloat.Parameters
	BootstrappingParameters hefloat.Parameters
	SlotsToCoeffsParameters hefloat.DFTMatrixLiteral
	Mod1ParametersLiteral   hefloat.Mod1ParametersLiteral
	CoeffsToSlotsParameters hefloat.DFTMatrixLiteral
	EphemeralSecretWeight   int
}

func (p Parameters) LogMaxDimensions() ring.Dimensions {
	return p.BootstrappingParameters.LogMaxDimensions()
}

func (p Parameters) LogMaxSlots() int {
	return p.BootstrappingParameters.LogMaxSlots()
}

func (p Parameters) DepthCoeffsToSlots() int {
	return p.CoeffsToSlotsParameters.Depth(true)
}

func (p Parameters) DepthEvalMod() int {
	return p.Mod1ParametersLiteral.Depth()
}

func (p Parameters) DepthSlotsToCoeffs() int {
	return p.SlotsToCoeffsParameters.Depth(true)
}

func (p Parameters) Depth() int {
	return p.DepthSlotsToCoeffs() + p.DepthEvalMod() + p.DepthCoeffsToSlots()
}

func (p Parameters) GaloisElements(params hefloat.Parameters) []uint64 {
	seen := map[uint64]struct{}{}
	add := func(g uint64) {
		if _, ok := seen[g]; ok {
			return
		}
		seen[g] = struct{}{}
	}
	for _, g := range p.CoeffsToSlotsParameters.GaloisElements(params) {
		add(g)
	}
	for _, g := range p.SlotsToCoeffsParameters.GaloisElements(params) {
		add(g)
	}
	add(params.GaloisElementForComplexConjugation())
	out := make([]uint64, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	return out
}

func (p Parameters) Validate() error {
	if p.BootstrappingParameters.MaxLevel()-p.DepthCoeffsToSlots() != p.Mod1ParametersLiteral.LevelQ {
		return fmt.Errorf("bootstrapping: Mod1Parameters.LevelQ inconsistent with CoeffsToSlots depth")
	}
	return nil
}
