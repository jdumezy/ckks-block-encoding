package radix

import (
	"fmt"

	"github.com/Pro7ech/lattigo/rlwe"
)

// Bit for word w lives at slot (w*WordStride + d - 1) * Stride.
type PerWordBitResult struct {
	Spec PaddedSpec
	CT   *rlwe.Ciphertext
}

type EqResult = PerWordBitResult
type LtResult = PerWordBitResult

func (e *Evaluator) decodeBits(res *PerWordBitResult) ([]bool, error) {
	pt := e.Ctx.Decryptor.DecryptNew(res.CT)
	values := make([]complex128, e.Ctx.Params.MaxSlots())
	if err := e.Ctx.Encoder.Decode(pt, values); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	out := make([]bool, res.Spec.BatchSize)
	d := res.Spec.Base.Width
	for w := 0; w < res.Spec.BatchSize; w++ {
		slot := (w*res.Spec.WordStride + d - 1) * res.Spec.Stride
		out[w] = real(values[slot]) > 0.5
	}
	return out, nil
}
