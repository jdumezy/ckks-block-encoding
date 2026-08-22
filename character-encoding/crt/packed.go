package crt

import (
	"fmt"
	"math/big"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
)

type PackedChannel struct {
	Ciphertext int
	Offset     int
	Slots      int
}

type PackedSpec struct {
	Base       Spec
	MaxSlots   int
	PackCap    int
	WordStride int
	BatchSize  int
	Channels   []PackedChannel
	CTSlots    []int
	TotalUsed  int
}

type PackedPlaintext struct {
	Spec   PackedSpec
	Values [][]complex128
}

type PackedCiphertext struct {
	Spec PackedSpec
	CTs  []*rlwe.Ciphertext
}

func NewPackedSpec(base Spec, maxSlots int) (PackedSpec, error) {
	return NewPackedSpecWithMaxUsedSlots(base, maxSlots, maxSlots)
}

func NewPackedSpecWithMaxUsedSlots(base Spec, physicalMaxSlots, packMaxSlots int) (PackedSpec, error) {
	if physicalMaxSlots <= 0 {
		return PackedSpec{}, fmt.Errorf("crt.NewPackedSpecWithMaxUsedSlots: physicalMaxSlots must be positive")
	}
	if packMaxSlots <= 0 {
		return PackedSpec{}, fmt.Errorf("crt.NewPackedSpecWithMaxUsedSlots: packMaxSlots must be positive")
	}
	if packMaxSlots > physicalMaxSlots {
		return PackedSpec{}, fmt.Errorf("crt.NewPackedSpecWithMaxUsedSlots: packMaxSlots %d exceeds physicalMaxSlots %d", packMaxSlots, physicalMaxSlots)
	}
	channels := make([]PackedChannel, base.Channels())
	offset := 0
	for i, spec := range base.Specs {
		channels[i] = PackedChannel{
			Ciphertext: 0,
			Offset:     offset,
			Slots:      spec.Slots,
		}
		offset += spec.Slots
	}
	wordStride := offset
	if wordStride <= 0 {
		return PackedSpec{}, fmt.Errorf("crt.NewPackedSpecWithMaxUsedSlots: empty word layout")
	}
	if wordStride > packMaxSlots {
		return PackedSpec{}, fmt.Errorf("crt.NewPackedSpecWithMaxUsedSlots: one CRT word needs %d slots, pack cap is %d", wordStride, packMaxSlots)
	}
	batchSize := packMaxSlots / wordStride
	if batchSize < 1 {
		return PackedSpec{}, fmt.Errorf("crt.NewPackedSpecWithMaxUsedSlots: pack cap %d cannot fit one %d-slot CRT word", packMaxSlots, wordStride)
	}
	totalUsed := batchSize * wordStride
	return PackedSpec{
		Base:       base,
		MaxSlots:   physicalMaxSlots,
		PackCap:    packMaxSlots,
		WordStride: wordStride,
		BatchSize:  batchSize,
		Channels:   channels,
		CTSlots:    []int{totalUsed},
		TotalUsed:  totalUsed,
	}, nil
}

func NewPackedSpecWithTargetCiphertexts(base Spec, physicalMaxSlots, targetCTs int) (PackedSpec, error) {
	if physicalMaxSlots <= 0 {
		return PackedSpec{}, fmt.Errorf("crt.NewPackedSpecWithTargetCiphertexts: physicalMaxSlots must be positive")
	}
	if targetCTs <= 0 {
		return PackedSpec{}, fmt.Errorf("crt.NewPackedSpecWithTargetCiphertexts: targetCTs must be positive")
	}
	return NewPackedSpecWithMaxUsedSlots(base, physicalMaxSlots, physicalMaxSlots)
}

func (s PackedSpec) Ciphertexts() int {
	return len(s.CTSlots)
}

func (s PackedSpec) slotOf(word, channel, coord int) (int, int, error) {
	if word < 0 || word >= s.BatchSize {
		return 0, 0, fmt.Errorf("word %d outside batch size %d", word, s.BatchSize)
	}
	if channel < 0 || channel >= len(s.Channels) {
		return 0, 0, fmt.Errorf("channel %d outside channel count %d", channel, len(s.Channels))
	}
	ch := s.Channels[channel]
	if coord < 0 || coord >= ch.Slots {
		return 0, 0, fmt.Errorf("coord %d outside channel %d slot count %d", coord, channel, ch.Slots)
	}
	return ch.Ciphertext, word*s.WordStride + ch.Offset + coord, nil
}

func EncodePacked(s PackedSpec, x *big.Int) (PackedPlaintext, error) {
	xs := make([]*big.Int, s.BatchSize)
	for i := range xs {
		xs[i] = x
	}
	return EncodePackedBatch(s, xs)
}

func EncodePackedBatch(s PackedSpec, xs []*big.Int) (PackedPlaintext, error) {
	if len(xs) != s.BatchSize {
		return PackedPlaintext{}, fmt.Errorf("crt.EncodePackedBatch: got %d words, BatchSize=%d", len(xs), s.BatchSize)
	}
	values := make([][]complex128, s.Ciphertexts())
	for i := range values {
		values[i] = make([]complex128, s.MaxSlots)
	}
	for w, x := range xs {
		if x == nil {
			return PackedPlaintext{}, fmt.Errorf("crt.EncodePackedBatch: nil input at word %d", w)
		}
		for i, p := range s.Base.Primes {
			codec, err := codecForPrime(s.Base.Kind, p, s.Base.Reduced)
			if err != nil {
				return PackedPlaintext{}, err
			}
			residue := int(new(big.Int).Mod(x, big.NewInt(int64(p))).Int64())
			ch := s.Channels[i]
			ctIdx, slot, err := s.slotOf(w, i, 0)
			if err != nil {
				return PackedPlaintext{}, err
			}
			copy(values[ctIdx][slot:slot+ch.Slots], codec.EncodeValue(residue))
		}
	}
	return PackedPlaintext{Spec: s, Values: values}, nil
}

func EncryptPacked(ctx *charctx.Context, pt PackedPlaintext, level int) (PackedCiphertext, error) {
	if pt.Spec.MaxSlots > ctx.Params.MaxSlots() {
		return PackedCiphertext{}, fmt.Errorf("crt.EncryptPacked: packed spec needs %d slots, context supports %d", pt.Spec.MaxSlots, ctx.Params.MaxSlots())
	}
	cts := make([]*rlwe.Ciphertext, len(pt.Values))
	for i, values := range pt.Values {
		if len(values) > ctx.Params.MaxSlots() {
			return PackedCiphertext{}, fmt.Errorf("crt.EncryptPacked: ciphertext %d has %d slots, context supports %d", i, len(values), ctx.Params.MaxSlots())
		}
		padded := make([]complex128, ctx.Params.MaxSlots())
		copy(padded, values)
		plain := hefloat.NewPlaintext(ctx.Params, level)
		if err := ctx.Encoder.Encode(padded, plain); err != nil {
			return PackedCiphertext{}, fmt.Errorf("crt.EncryptPacked: encode ciphertext %d: %w", i, err)
		}
		ct := hefloat.NewCiphertext(ctx.Params, 1, level)
		if err := ctx.Encryptor.Encrypt(plain, ct); err != nil {
			return PackedCiphertext{}, fmt.Errorf("crt.EncryptPacked: encrypt ciphertext %d: %w", i, err)
		}
		cts[i] = ct
	}
	return PackedCiphertext{Spec: pt.Spec, CTs: cts}, nil
}

func DecryptPacked(ctx *charctx.Context, ct PackedCiphertext) (PackedPlaintext, error) {
	values := make([][]complex128, len(ct.CTs))
	for i, c := range ct.CTs {
		pt := ctx.Decryptor.DecryptNew(c)
		decoded := make([]complex128, ctx.Params.MaxSlots())
		if err := ctx.Encoder.Decode(pt, decoded); err != nil {
			return PackedPlaintext{}, fmt.Errorf("crt.DecryptPacked: decode ciphertext %d: %w", i, err)
		}
		values[i] = decoded[:ct.Spec.MaxSlots]
	}
	return PackedPlaintext{Spec: ct.Spec, Values: values}, nil
}

func DecodePackedResidues(pt PackedPlaintext) ([]int, error) {
	all, err := DecodePackedBatchResidues(pt)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("crt.DecodePackedResidues: empty batch")
	}
	return all[0], nil
}

func DecodePackedBatchResidues(pt PackedPlaintext) ([][]int, error) {
	if len(pt.Values) != pt.Spec.Ciphertexts() {
		return nil, fmt.Errorf("crt.DecodePackedBatchResidues: got %d ciphertexts, expected %d", len(pt.Values), pt.Spec.Ciphertexts())
	}
	all := make([][]int, pt.Spec.BatchSize)
	for w := range all {
		residues, err := decodePackedWordResidues(pt, w)
		if err != nil {
			return nil, err
		}
		all[w] = residues
	}
	return all, nil
}

func decodePackedWordResidues(pt PackedPlaintext, word int) ([]int, error) {
	out := make([]int, pt.Spec.Base.Channels())
	for i, p := range pt.Spec.Base.Primes {
		codec, err := codecForPrime(pt.Spec.Base.Kind, p, pt.Spec.Base.Reduced)
		if err != nil {
			return nil, err
		}
		ch := pt.Spec.Channels[i]
		ctIdx, slot, err := pt.Spec.slotOf(word, i, 0)
		if err != nil {
			return nil, err
		}
		if ctIdx >= len(pt.Values) || slot+ch.Slots > len(pt.Values[ctIdx]) {
			return nil, fmt.Errorf("crt.DecodePackedBatchResidues: word %d channel %d outside packed plaintext", word, i)
		}
		v, err := codec.DecodeValue(pt.Values[ctIdx][slot : slot+ch.Slots])
		if err != nil {
			return nil, fmt.Errorf("crt.DecodePackedBatchResidues: word %d channel %d: %w", word, i, err)
		}
		out[i] = v
	}
	return out, nil
}

func (e *Evaluator) EvalPackedNativeProduct(x, y PackedCiphertext, parallel bool) (PackedCiphertext, error) {
	if err := checkPackedBinaryOperands(x, y); err != nil {
		return PackedCiphertext{}, err
	}
	out := PackedCiphertext{Spec: x.Spec, CTs: make([]*rlwe.Ciphertext, len(x.CTs))}
	workers := channelParallelism(parallel, len(x.CTs), e.LUT.BLT.Capacity())
	if workers == 1 {
		w := e.LUT.BLT.GetWorker()
		defer e.LUT.BLT.PutWorker(w)
		for i := range x.CTs {
			ct, err := e.evalPackedNativeCiphertext(w, x.CTs[i], y.CTs[i])
			if err != nil {
				return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedNativeProduct: ciphertext %d: %w", i, err)
			}
			out.CTs[i] = ct
		}
		return out, nil
	}

	errs := make([]error, len(x.CTs))
	done := make(chan struct{}, workers)
	for workerID := 0; workerID < workers; workerID++ {
		go func(workerID int) {
			w := e.LUT.BLT.GetWorker()
			defer e.LUT.BLT.PutWorker(w)
			defer func() { done <- struct{}{} }()
			for i := workerID; i < len(x.CTs); i += workers {
				ct, err := e.evalPackedNativeCiphertext(w, x.CTs[i], y.CTs[i])
				if err != nil {
					errs[i] = err
					return
				}
				out.CTs[i] = ct
			}
		}(workerID)
	}
	for i := 0; i < workers; i++ {
		<-done
	}
	for i, err := range errs {
		if err != nil {
			return PackedCiphertext{}, fmt.Errorf("crt.EvalPackedNativeProduct: ciphertext %d: %w", i, err)
		}
	}
	return out, nil
}

func (e *Evaluator) evalPackedNativeCiphertext(w *blt.Worker, x, y *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	level := min(x.Level(), y.Level())
	out := hefloat.NewCiphertext(e.Ctx.Params, 1, level)
	if err := w.Eval.MulRelin(x, y, out); err != nil {
		return nil, err
	}
	if err := w.Eval.Rescale(out, out); err != nil {
		return nil, err
	}
	return out, nil
}

func checkPackedBinaryOperands(x, y PackedCiphertext) error {
	if !samePackedSpec(x.Spec, y.Spec) {
		return fmt.Errorf("packed CRT spec mismatch")
	}
	if len(x.CTs) != x.Spec.Ciphertexts() || len(y.CTs) != y.Spec.Ciphertexts() {
		return fmt.Errorf("packed CRT ciphertext count mismatch")
	}
	return nil
}

func samePackedSpec(a, b PackedSpec) bool {
	if a.MaxSlots != b.MaxSlots || a.PackCap != b.PackCap || a.WordStride != b.WordStride || a.BatchSize != b.BatchSize || a.TotalUsed != b.TotalUsed || a.Base.Kind != b.Base.Kind || a.Base.Reduced != b.Base.Reduced || a.Base.Channels() != b.Base.Channels() {
		return false
	}
	if len(a.CTSlots) != len(b.CTSlots) || len(a.Channels) != len(b.Channels) {
		return false
	}
	for i := range a.Base.Primes {
		if a.Base.Primes[i] != b.Base.Primes[i] {
			return false
		}
	}
	for i := range a.CTSlots {
		if a.CTSlots[i] != b.CTSlots[i] {
			return false
		}
	}
	for i := range a.Channels {
		if a.Channels[i] != b.Channels[i] {
			return false
		}
	}
	return true
}
