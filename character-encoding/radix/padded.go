package radix

import (
	"fmt"
	"math/big"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

// Each digit occupies Stride = base^2 slots (BRU reduced encoding in the
// first BlockSlots, then zero padding). Unused blocks must hold BRU(0), not
// literal zero, so the LUT's interpolated average does not leak into the scan.
// Each word gets WordStride = d + maxKoggeStoneStep blocks; the headroom
// isolates words from the cyclic rotations of the Kogge-Stone scan.
type PaddedSpec struct {
	Base        Spec
	BlockSlots  int
	Stride      int
	MaxSlots    int
	BlocksPerCT int
	WordStride  int
	BatchSize   int
	Digits      []PaddedDigit
	CTCount     int
}

type PaddedDigit struct {
	Ciphertext int
	Index      int
}

type PaddedPlaintext struct {
	Spec   PaddedSpec
	Values [][]complex128
}

type PaddedCiphertext struct {
	Spec PaddedSpec
	CTs  []*rlwe.Ciphertext
}

func NewPaddedSpec(base Spec, maxSlots int) (PaddedSpec, error) {
	return NewPaddedSpecBatched(base, maxSlots, 1)
}

func NewPaddedSpecBatched(base Spec, maxSlots, batchSize int) (PaddedSpec, error) {
	if err := base.Validate(); err != nil {
		return PaddedSpec{}, err
	}
	if batchSize < 1 {
		return PaddedSpec{}, fmt.Errorf("radix.NewPaddedSpecBatched: batchSize must be >= 1, got %d", batchSize)
	}
	blockSlots := base.Radix - 1
	stride := base.Radix * base.Radix
	if stride > maxSlots {
		return PaddedSpec{}, fmt.Errorf("radix.NewPaddedSpecBatched: stride %d exceeds maxSlots %d", stride, maxSlots)
	}
	blocksPerCT := maxSlots / stride
	if blocksPerCT < 1 {
		return PaddedSpec{}, fmt.Errorf("radix.NewPaddedSpecBatched: maxSlots %d cannot hold one block of stride %d", maxSlots, stride)
	}

	maxStep := kogglesStoneMaxStep(base.Width)
	wordStride := base.Width + maxStep
	if batchSize*wordStride > blocksPerCT {
		return PaddedSpec{}, fmt.Errorf(
			"radix.NewPaddedSpecBatched: %d words at WordStride=%d need %d blocks, have %d (max batchSize = %d)",
			batchSize, wordStride, batchSize*wordStride, blocksPerCT, blocksPerCT/wordStride,
		)
	}

	digits := make([]PaddedDigit, batchSize*base.Width)
	for w := 0; w < batchSize; w++ {
		for i := 0; i < base.Width; i++ {
			digits[w*base.Width+i] = PaddedDigit{
				Ciphertext: 0,
				Index:      w*wordStride + i,
			}
		}
	}
	return PaddedSpec{
		Base:        base,
		BlockSlots:  blockSlots,
		Stride:      stride,
		MaxSlots:    maxSlots,
		BlocksPerCT: blocksPerCT,
		WordStride:  wordStride,
		BatchSize:   batchSize,
		Digits:      digits,
		CTCount:     1,
	}, nil
}

func MaxBatchSize(base Spec, maxSlots int) int {
	if err := base.Validate(); err != nil {
		return 0
	}
	stride := base.Radix * base.Radix
	if stride > maxSlots {
		return 0
	}
	blocksPerCT := maxSlots / stride
	wordStride := base.Width + kogglesStoneMaxStep(base.Width)
	if wordStride <= 0 {
		return 0
	}
	return blocksPerCT / wordStride
}

// max_step = 2^(ceil(log2 d) - 1); 1 for d <= 1.
func kogglesStoneMaxStep(d int) int {
	if d <= 1 {
		return 1
	}
	rounds := ceilLog2(d)
	return 1 << (rounds - 1)
}

// Digits are word-major: digit i is in batched word i/Width at position i%Width.
func (s PaddedSpec) SlotOf(i int) int {
	return s.Digits[i].Index * s.Stride
}

func EncodePadded(s PaddedSpec, x *big.Int) (PaddedPlaintext, error) {
	if s.BatchSize != 1 {
		return PaddedPlaintext{}, fmt.Errorf("radix.EncodePadded: spec has BatchSize=%d, use EncodePaddedBatch", s.BatchSize)
	}
	return EncodePaddedBatch(s, []*big.Int{x})
}

// Word w occupies blocks [w*WordStride, w*WordStride+d) followed by maxStep
// BRU(0) headroom blocks. Every other slot is also BRU(0) so the LUT sees a
// valid encoding everywhere.
func EncodePaddedBatch(s PaddedSpec, xs []*big.Int) (PaddedPlaintext, error) {
	if len(xs) != s.BatchSize {
		return PaddedPlaintext{}, fmt.Errorf("radix.EncodePaddedBatch: got %d words, BatchSize=%d", len(xs), s.BatchSize)
	}
	bru, err := charenc.NewBRU(s.Base.Radix, true)
	if err != nil {
		return PaddedPlaintext{}, err
	}
	zeroBRU := bru.EncodeValue(0)
	values := make([][]complex128, s.CTCount)
	for c := range values {
		values[c] = make([]complex128, s.MaxSlots)
		for k := 0; k < s.BlocksPerCT; k++ {
			base := k * s.Stride
			copy(values[c][base:base+s.BlockSlots], zeroBRU)
		}
	}
	for w, x := range xs {
		word, err := Encode(s.Base, x)
		if err != nil {
			return PaddedPlaintext{}, fmt.Errorf("word %d: %w", w, err)
		}
		for i, digit := range word.Digits {
			place := s.Digits[w*s.Base.Width+i]
			encoded := bru.EncodeValue(digit)
			base := place.Index * s.Stride
			copy(values[place.Ciphertext][base:base+s.BlockSlots], encoded)
		}
	}
	return PaddedPlaintext{Spec: s, Values: values}, nil
}

func EncryptPadded(ctx *charctx.Context, pt PaddedPlaintext, level int) (PaddedCiphertext, error) {
	if pt.Spec.MaxSlots > ctx.Params.MaxSlots() {
		return PaddedCiphertext{}, fmt.Errorf("radix.EncryptPadded: spec needs %d slots, context supports %d", pt.Spec.MaxSlots, ctx.Params.MaxSlots())
	}
	cts := make([]*rlwe.Ciphertext, pt.Spec.CTCount)
	for i, vals := range pt.Values {
		padded := make([]complex128, ctx.Params.MaxSlots())
		copy(padded, vals)
		plain := hefloat.NewPlaintext(ctx.Params, level)
		if err := ctx.Encoder.Encode(padded, plain); err != nil {
			return PaddedCiphertext{}, fmt.Errorf("encode CT %d: %w", i, err)
		}
		ct := hefloat.NewCiphertext(ctx.Params, 1, level)
		if err := ctx.Encryptor.Encrypt(plain, ct); err != nil {
			return PaddedCiphertext{}, fmt.Errorf("encrypt CT %d: %w", i, err)
		}
		cts[i] = ct
	}
	return PaddedCiphertext{Spec: pt.Spec, CTs: cts}, nil
}

func DecryptPadded(ctx *charctx.Context, ct PaddedCiphertext) (PaddedPlaintext, error) {
	values := make([][]complex128, ct.Spec.CTCount)
	for i, c := range ct.CTs {
		pt := ctx.Decryptor.DecryptNew(c)
		decoded := make([]complex128, ctx.Params.MaxSlots())
		if err := ctx.Encoder.Decode(pt, decoded); err != nil {
			return PaddedPlaintext{}, fmt.Errorf("decode CT %d: %w", i, err)
		}
		values[i] = decoded[:ct.Spec.MaxSlots]
	}
	return PaddedPlaintext{Spec: ct.Spec, Values: values}, nil
}

func DecodePadded(pt PaddedPlaintext) (*big.Int, error) {
	if pt.Spec.BatchSize != 1 {
		return nil, fmt.Errorf("radix.DecodePadded: spec has BatchSize=%d, use DecodePaddedBatch", pt.Spec.BatchSize)
	}
	out, err := DecodePaddedBatch(pt)
	if err != nil {
		return nil, err
	}
	return out[0], nil
}

func DecodePaddedBatch(pt PaddedPlaintext) ([]*big.Int, error) {
	bru, err := charenc.NewBRU(pt.Spec.Base.Radix, true)
	if err != nil {
		return nil, err
	}
	out := make([]*big.Int, pt.Spec.BatchSize)
	d := pt.Spec.Base.Width
	for w := 0; w < pt.Spec.BatchSize; w++ {
		digits := make([]int, d)
		for i := 0; i < d; i++ {
			place := pt.Spec.Digits[w*d+i]
			base := place.Index * pt.Spec.Stride
			digit, err := bru.DecodeValue(pt.Values[place.Ciphertext][base : base+pt.Spec.BlockSlots])
			if err != nil {
				return nil, fmt.Errorf("word %d digit %d: %w", w, i, err)
			}
			digits[i] = digit
		}
		x, err := Decode(Word{Spec: pt.Spec.Base, Digits: digits})
		if err != nil {
			return nil, fmt.Errorf("word %d: %w", w, err)
		}
		out[w] = x
	}
	return out, nil
}

func positiveMod(x, m int) int {
	v := x % m
	if v < 0 {
		v += m
	}
	return v
}

func samePaddedSpec(a, b PaddedSpec) bool {
	if a.Base != b.Base {
		return false
	}
	if a.MaxSlots != b.MaxSlots || a.Stride != b.Stride || a.BlocksPerCT != b.BlocksPerCT {
		return false
	}
	if a.WordStride != b.WordStride || a.BatchSize != b.BatchSize || a.CTCount != b.CTCount {
		return false
	}
	return true
}
