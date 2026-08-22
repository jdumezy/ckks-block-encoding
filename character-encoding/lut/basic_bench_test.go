package lut

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/Pro7ech/lattigo/he"
	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/ring"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

const (
	basicBenchLogN       = 15
	basicRandomLUTSeed   = int64(0x5eed5eed)
	basicRandomLUTStride = int64(1_000_003)
)

type basicEncoding struct {
	name string
	kind charenc.EncodingKind
}

var basicEncodings = []basicEncoding{
	{name: "BRU", kind: charenc.BRU},
	{name: "LBRU", kind: charenc.LBRU},
	{name: "WH", kind: charenc.WH},
	{name: "IND", kind: charenc.IND},
}

func basicBenchParams(levels int) hefloat.ParametersLiteral {
	logQ := []int{55}
	for i := 0; i < levels; i++ {
		logQ = append(logQ, 40)
	}
	return hefloat.ParametersLiteral{
		LogN:            basicBenchLogN,
		LogQ:            logQ,
		LogP:            []int{60},
		Xs:              &ring.Ternary{H: 256},
		LogDefaultScale: 40,
	}
}

func basicLog2Ts() []int {
	raw := strings.TrimSpace(os.Getenv("BASIC_BENCH_LOG2T"))
	if raw == "" {
		return []int{2, 3, 4, 5, 6, 7, 8}
	}
	out := []int{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, err := strconv.Atoi(part)
		if err != nil {
			panic(fmt.Sprintf("invalid BASIC_BENCH_LOG2T=%q", raw))
		}
		out = append(out, k)
	}
	return out
}

func selectedBasicEncodings() []basicEncoding {
	raw := strings.TrimSpace(os.Getenv("BASIC_BENCH_ENCODING"))
	if raw == "" {
		return basicEncodings
	}
	want := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		want[strings.ToUpper(strings.TrimSpace(part))] = true
	}
	out := []basicEncoding{}
	for _, enc := range basicEncodings {
		if want[enc.name] {
			out = append(out, enc)
		}
	}
	if len(out) == 0 {
		panic(fmt.Sprintf("invalid BASIC_BENCH_ENCODING=%q", raw))
	}
	return out
}

func basicSpec(kind charenc.EncodingKind, log2t int) (charenc.BlockSpec, int, error) {
	t := 1 << log2t
	switch kind {
	case charenc.BRU:
		c, err := charenc.NewBRU(t, true)
		if err != nil {
			return charenc.BlockSpec{}, 0, err
		}
		return c.Spec(), t, nil
	case charenc.LBRU:
		t = previousPrime(t)
		c, err := charenc.NewLBRU(t, 0, true)
		if err != nil {
			return charenc.BlockSpec{}, 0, err
		}
		return c.Spec(), t, nil
	case charenc.WH:
		c, err := charenc.NewWH(t, true)
		if err != nil {
			return charenc.BlockSpec{}, 0, err
		}
		return c.Spec(), t, nil
	case charenc.IND:
		c, err := charenc.NewIND(t, true, 0)
		if err != nil {
			return charenc.BlockSpec{}, 0, err
		}
		return c.Spec(), t, nil
	default:
		return charenc.BlockSpec{}, 0, fmt.Errorf("unsupported encoding %v", kind)
	}
}

func previousPrime(n int) int {
	for p := n; p >= 2; p-- {
		if isSmallPrime(p) {
			return p
		}
	}
	return 2
}

func isSmallPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	for d := 3; d*d <= n; d += 2 {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func positiveMod(x, m int) int {
	r := x % m
	if r < 0 {
		r += m
	}
	return r
}

func basicBatch(b *testing.B, spec charenc.BlockSpec, featureSlots int) int {
	b.Helper()
	blockSlots := max(spec.Slots, featureSlots)
	batch := (1 << (basicBenchLogN - 1)) / blockSlots
	if batch < 1 {
		b.Skipf("block layout needs %d slots, LogN=%d supports %d", blockSlots, basicBenchLogN, 1<<(basicBenchLogN-1))
	}
	return batch
}

func newBasicContext(b *testing.B, levels int) *charctx.Context {
	b.Helper()
	ctx, err := charctx.NewContext(basicBenchParams(levels))
	if err != nil {
		b.Fatal(err)
	}
	return ctx
}

func encodePackedBlocks(spec charenc.BlockSpec, values []int, slots int) ([]complex128, error) {
	return encodePackedBlocksStrided(spec, values, spec.Slots, slots)
}

func encodePackedBlocksStrided(spec charenc.BlockSpec, values []int, stride, slots int) ([]complex128, error) {
	codec, err := charenc.NewCodec(spec)
	if err != nil {
		return nil, err
	}
	out := make([]complex128, slots)
	for i, v := range values {
		copy(out[i*stride:i*stride+spec.Slots], codec.EncodeValue(v))
	}
	return out, nil
}

func encryptVector(ctx *charctx.Context, values []complex128, level int) (*rlwe.Ciphertext, error) {
	padded := make([]complex128, ctx.Params.MaxSlots())
	copy(padded, values)
	pt := hefloat.NewPlaintext(ctx.Params, level)
	if err := ctx.Encoder.Encode(padded, pt); err != nil {
		return nil, err
	}
	ct := hefloat.NewCiphertext(ctx.Params, 1, level)
	if err := ctx.Encryptor.Encrypt(pt, ct); err != nil {
		return nil, err
	}
	return ct, nil
}

func decryptVector(ctx *charctx.Context, ct *rlwe.Ciphertext) ([]complex128, error) {
	pt := ctx.Decryptor.DecryptNew(ct)
	values := make([]complex128, ctx.Params.MaxSlots())
	if err := ctx.Encoder.Decode(pt, values); err != nil {
		return nil, err
	}
	return values, nil
}

func maxAbsError(got, want []complex128, n int) float64 {
	maxErr := 0.0
	for i := 0; i < n; i++ {
		if err := cmplx.Abs(got[i] - want[i]); err > maxErr {
			maxErr = err
		}
	}
	return maxErr
}

func maxAbsErrorStrided(got, want []complex128, batch, blockSlots, stride int) float64 {
	maxErr := 0.0
	for block := 0; block < batch; block++ {
		base := block * stride
		for j := 0; j < blockSlots; j++ {
			if err := cmplx.Abs(got[base+j] - want[base+j]); err > maxErr {
				maxErr = err
			}
		}
	}
	return maxErr
}

func precisionBits(err float64) (bits float64, saturated bool) {
	if err <= 0 {
		return 200, true
	}
	p := -math.Log2(err)
	if p > 200 {
		return 200, true
	}
	return p, false
}

func checkDecodedBlocks(b *testing.B, spec charenc.BlockSpec, decoded []complex128, want []int) bool {
	return checkDecodedBlocksStrided(b, spec, decoded, want, spec.Slots)
}

func checkDecodedBlocksStrided(b *testing.B, spec charenc.BlockSpec, decoded []complex128, want []int, stride int) bool {
	b.Helper()
	codec, err := charenc.NewCodec(spec)
	if err != nil {
		b.Fatal(err)
	}
	ok := true
	for i, w := range want {
		got, err := codec.DecodeValue(decoded[i*stride : i*stride+spec.Slots])
		if err != nil {
			b.Fatal(err)
		}
		if got != positiveMod(w, spec.Alphabet.Modulus) {
			ok = false
			break
		}
	}
	return ok
}

func randValues(t, batch int, seed int64) []int {
	rng := rand.New(rand.NewSource(seed))
	out := make([]int, batch)
	for i := range out {
		out[i] = rng.Intn(t)
	}
	return out
}

func randomLUTTable(t, arity int) []int {
	size := 1
	for i := 0; i < arity; i++ {
		size *= t
	}
	rng := rand.New(rand.NewSource(basicRandomLUTSeed + int64(t)*basicRandomLUTStride + int64(arity)))
	table := make([]int, size)
	for i := range table {
		table[i] = rng.Intn(t)
	}
	if t > 1 && len(table) > 1 {
		allSame := true
		for _, v := range table[1:] {
			if v != table[0] {
				allSame = false
				break
			}
		}
		if allSame {
			table[len(table)-1] = (table[0] + 1) % t
		}
	}
	return table
}

func lutTableIndex(t int, xs []int) int {
	idx := 0
	for _, x := range xs {
		idx = idx*t + positiveMod(x, t)
	}
	return idx
}

func reportBasicMetrics(b *testing.B, ctx *charctx.Context, spec charenc.BlockSpec, log2t int, actualT int, batch int, levels int, inPrec float64, inSat bool, outPrec float64, outSat bool, correct bool) {
	b.ReportMetric(float64(batch), "blocks/op")
	b.ReportMetric(float64(batch), "words/op")
	b.ReportMetric(float64(spec.Slots), "block_slots")
	b.ReportMetric(float64(log2t), "requested_log2t")
	b.ReportMetric(math.Log2(float64(actualT)), "log2T")
	b.ReportMetric(float64(levels), "levels/op")
	b.ReportMetric(float64(runtime.GOMAXPROCS(0)), "gomaxprocs")
	b.ReportMetric(ctx.Params.LogQ(), "logQ")
	b.ReportMetric(ctx.Params.LogP(), "logP")
	b.ReportMetric(ctx.Params.LogQP(), "logQP")
	b.ReportMetric(256, "h")
	b.ReportMetric(inPrec, "input_precision_bits")
	b.ReportMetric(outPrec, "precision_bits")
	if inSat {
		b.ReportMetric(1, "input_precision_saturated")
	}
	if outSat {
		b.ReportMetric(1, "precision_saturated")
	}
	if !inSat && !outSat {
		b.ReportMetric(outPrec-inPrec, "precision_delta_bits")
	}
	if correct {
		b.ReportMetric(1, "correct")
		b.ReportMetric(0, "max_error")
	} else {
		b.ReportMetric(0, "correct")
		b.ReportMetric(1, "max_error")
	}
}

func reportBasicCleanMetrics(b *testing.B, ctx *charctx.Context, spec charenc.BlockSpec, log2t int, actualT int, batch int, levels int, inPrec float64, inSat bool, outPrec float64, outSat bool, correct bool) {
	reportBasicMetrics(b, ctx, spec, log2t, actualT, batch, levels, inPrec, inSat, outPrec, outSat, correct)
}

func nativeLaw(kind charenc.EncodingKind, t, x, y int) int {
	switch kind {
	case charenc.BRU:
		return positiveMod(x+y, t)
	case charenc.LBRU:
		return positiveMod(x*y, t)
	case charenc.WH:
		return x ^ y
	case charenc.IND:
		return x
	default:
		return 0
	}
}

func benchBasicNative(b *testing.B, enc basicEncoding, log2t int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	batch := basicBatch(b, spec, spec.Slots)
	ctx := newBasicContext(b, 1)
	ev := NewEvaluatorWithWorkerCapacity(ctx, 1)
	xs := randValues(t, batch, 11)
	ys := randValues(t, batch, 12)
	if enc.kind == charenc.IND {
		copy(ys, xs)
	}
	wantValues := make([]int, batch)
	for i := range wantValues {
		wantValues[i] = nativeLaw(enc.kind, t, xs[i], ys[i])
	}
	xPlain, err := encodePackedBlocks(spec, xs, ctx.Params.MaxSlots())
	if err != nil {
		b.Fatal(err)
	}
	yPlain, err := encodePackedBlocks(spec, ys, ctx.Params.MaxSlots())
	if err != nil {
		b.Fatal(err)
	}
	wantPlain, err := encodePackedBlocks(spec, wantValues, ctx.Params.MaxSlots())
	if err != nil {
		b.Fatal(err)
	}
	ctX, err := encryptVector(ctx, xPlain, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	ctY, err := encryptVector(ctx, yPlain, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	inDecoded, err := decryptVector(ctx, ctX)
	if err != nil {
		b.Fatal(err)
	}
	out, err := evalBasicNative(ev, ctX, ctY)
	if err != nil {
		b.Fatal(err)
	}
	outDecoded, err := decryptVector(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	used := batch * spec.Slots
	inPrec, inSat := precisionBits(maxAbsError(inDecoded, xPlain, used))
	outPrec, outSat := precisionBits(maxAbsError(outDecoded, wantPlain, used))
	correct := checkDecodedBlocks(b, spec, outDecoded, wantValues)
	b.ReportAllocs()
	b.ResetTimer()
	reportBasicMetrics(b, ctx, spec, log2t, t, batch, 1, inPrec, inSat, outPrec, outSat, correct)
	for i := 0; i < b.N; i++ {
		if _, err := evalBasicNative(ev, ctX, ctY); err != nil {
			b.Fatal(err)
		}
	}
}

func evalBasicNative(ev *Evaluator, x, y *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	w := ev.BLT.GetWorker()
	defer ev.BLT.PutWorker(w)
	level := min(x.Level(), y.Level())
	out := hefloat.NewCiphertext(ev.Ctx.Params, 1, level)
	if err := w.Eval.MulRelin(x, y, out); err != nil {
		return nil, err
	}
	if err := w.Eval.Rescale(out, out); err != nil {
		return nil, err
	}
	return out, nil
}

type packedUnaryPlan struct {
	raw      *blt.RawCompiled
	outSpec  charenc.BlockSpec
	inSlots  int
	outSlots int
}

func compilePackedUnary(ctx *charctx.Context, in, out charenc.BlockSpec, batch int, inputLevel int, f func(int) int) (*packedUnaryPlan, error) {
	inCodec, err := charenc.NewCodec(in)
	if err != nil {
		return nil, err
	}
	outCodec, err := charenc.NewCodec(out)
	if err != nil {
		return nil, err
	}
	tr, err := blt.CompileUnary(inCodec, outCodec, f)
	if err != nil {
		return nil, err
	}
	raw, err := compilePackedBlockTransform(ctx, tr.Matrix, tr.Bias, batch, in.Slots, out.Slots, inputLevel)
	if err != nil {
		return nil, err
	}
	ctx.EnsureGaloisKeys(raw.GaloisEls)
	return &packedUnaryPlan{raw: raw, outSpec: out, inSlots: batch * in.Slots, outSlots: batch * out.Slots}, nil
}

func compilePackedBlockTransform(ctx *charctx.Context, matrix [][]complex128, bias []complex128, batch, inBlockSlots, outBlockSlots int, inputLevel int) (*blt.RawCompiled, error) {
	return compilePackedBlockTransformStrided(ctx, matrix, bias, batch, inBlockSlots, outBlockSlots, inBlockSlots, outBlockSlots, inputLevel)
}

func compilePackedBlockTransformStrided(ctx *charctx.Context, matrix [][]complex128, bias []complex128, batch, inBlockSlots, outBlockSlots, inStride, outStride int, inputLevel int) (*blt.RawCompiled, error) {
	slots := ctx.Params.MaxSlots()
	inSlots := batch * inStride
	outSlots := batch * outStride
	diagonals := he.Diagonals[complex128]{}
	fullBias := make([]complex128, outSlots)
	diagonalOnly := true
	for block := 0; block < batch; block++ {
		inBase := block * inStride
		outBase := block * outStride
		copy(fullBias[outBase:outBase+outBlockSlots], bias)
		for r := 0; r < outBlockSlots; r++ {
			for c, v := range matrix[r] {
				if v == 0 {
					continue
				}
				outSlot := outBase + r
				inSlot := inBase + c
				if inSlot != outSlot {
					diagonalOnly = false
				}
				d := (inSlot - outSlot) % slots
				if d < 0 {
					d += slots
				}
				row, ok := diagonals[d]
				if !ok {
					row = make([]complex128, slots)
					diagonals[d] = row
				}
				row[outSlot] = v
			}
		}
	}
	return blt.CompileDiagonalsWithOptions(diagonals, diagonalOnly, fullBias, outSlots, inSlots, ctx, inputLevel, blt.CompileOptions{})
}

func benchBasicUnaryLUT(b *testing.B, enc basicEncoding, log2t int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	batch := basicBatch(b, spec, spec.Slots)
	ctx := newBasicContext(b, 1)
	ev := NewEvaluatorWithWorkerCapacity(ctx, 1)
	table := randomLUTTable(t, 1)
	fn := func(x int) int { return table[positiveMod(x, t)] }
	plan, err := compilePackedUnary(ctx, spec, spec, batch, ctx.Params.MaxLevel(), fn)
	if err != nil {
		b.Fatal(err)
	}
	xs := randValues(t, batch, 21)
	wantValues := make([]int, batch)
	for i, x := range xs {
		wantValues[i] = fn(x)
	}
	xPlain, _ := encodePackedBlocks(spec, xs, ctx.Params.MaxSlots())
	wantPlain, _ := encodePackedBlocks(spec, wantValues, ctx.Params.MaxSlots())
	ctX, err := encryptVector(ctx, xPlain, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	inDecoded, _ := decryptVector(ctx, ctX)
	out, err := ev.BLT.ApplyRaw(ctX, plan.raw)
	if err != nil {
		b.Fatal(err)
	}
	outDecoded, err := decryptVector(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	used := batch * spec.Slots
	inPrec, inSat := precisionBits(maxAbsError(inDecoded, xPlain, used))
	outPrec, outSat := precisionBits(maxAbsError(outDecoded, wantPlain, used))
	correct := checkDecodedBlocks(b, spec, outDecoded, wantValues)
	b.ReportAllocs()
	b.ResetTimer()
	reportBasicMetrics(b, ctx, spec, log2t, t, batch, 1, inPrec, inSat, outPrec, outSat, correct)
	for i := 0; i < b.N; i++ {
		if _, err := ev.BLT.ApplyRaw(ctX, plan.raw); err != nil {
			b.Fatal(err)
		}
	}
}

type packedTensorPlan struct {
	arity       int
	spec        charenc.BlockSpec
	batch       int
	spreads     []*blt.RawCompiled
	finalLT     *blt.RawCompiled
	mulSchedule []mulStep
	mulLevels   [][]int
	allGalois   []uint64
}

func compilePackedTensor(ctx *charctx.Context, spec charenc.BlockSpec, arity, batch, inputLevel int, fn func([]int) int) (*packedTensorPlan, error) {
	J := spec.Slots
	featureSlots := 1
	for i := 0; i < arity; i++ {
		featureSlots *= 1 + J
	}
	inSlots := batch * featureSlots
	packedSlots := batch * featureSlots
	outSlots := batch * featureSlots
	spreads := make([]*blt.RawCompiled, arity)
	for i := 0; i < arity; i++ {
		raw, err := compilePackedAugmentedSpread(ctx, J, arity, i, batch, inputLevel)
		if err != nil {
			return nil, fmt.Errorf("spread %d: %w", i, err)
		}
		spreads[i] = raw
	}
	inCodecs := make([]charenc.Codec, arity)
	for i := 0; i < arity; i++ {
		c, err := charenc.NewCodec(spec)
		if err != nil {
			return nil, err
		}
		inCodecs[i] = c
	}
	outCodec, err := charenc.NewCodec(spec)
	if err != nil {
		return nil, err
	}
	Js := make([]int, arity)
	for i := range Js {
		Js[i] = J
	}
	coef, err := interpolateAugmentedTensor(inCodecs, outCodec, Js, fn)
	if err != nil {
		return nil, err
	}
	schedule, mulLevels, packedDepth := mulSchedulePlan(arity)
	packedLevel := inputLevel - (1+packedDepth)*ctx.Params.LevelsConsumedPerRescaling()
	finalLT, err := compilePackedBlockTransformStrided(ctx, coef, nil, batch, featureSlots, J, featureSlots, featureSlots, packedLevel)
	if err != nil {
		return nil, fmt.Errorf("final LT: %w", err)
	}
	if inSlots > ctx.Params.MaxSlots() || packedSlots > ctx.Params.MaxSlots() || outSlots > ctx.Params.MaxSlots() {
		return nil, fmt.Errorf("packed tensor dimensions exceed slots")
	}
	all := map[uint64]struct{}{}
	for _, raw := range spreads {
		for _, g := range raw.GaloisEls {
			all[g] = struct{}{}
		}
	}
	for _, g := range finalLT.GaloisEls {
		all[g] = struct{}{}
	}
	allGalois := make([]uint64, 0, len(all))
	for g := range all {
		allGalois = append(allGalois, g)
	}
	ctx.EnsureGaloisKeys(allGalois)
	return &packedTensorPlan{arity: arity, spec: spec, batch: batch, spreads: spreads, finalLT: finalLT, mulSchedule: schedule, mulLevels: mulLevels, allGalois: allGalois}, nil
}

func compilePackedAugmentedSpread(ctx *charctx.Context, J, arity, axis, batch, inputLevel int) (*blt.RawCompiled, error) {
	slots := ctx.Params.MaxSlots()
	augSlots := make([]int, arity)
	featureSlots := 1
	for i := range augSlots {
		augSlots[i] = 1 + J
		featureSlots *= augSlots[i]
	}
	inSlots := batch * featureSlots
	outSlots := batch * featureSlots
	diagonals := he.Diagonals[complex128]{}
	bias := make([]complex128, outSlots)
	diagonalOnly := true
	for block := 0; block < batch; block++ {
		inBase := block * featureSlots
		outBase := block * featureSlots
		for fp := 0; fp < featureSlots; fp++ {
			tuple := unpackIndex(fp, augSlots)
			outSlot := outBase + fp
			a := tuple[axis]
			if a == 0 {
				bias[outSlot] = 1
				continue
			}
			inSlot := inBase + a - 1
			if inSlot != outSlot {
				diagonalOnly = false
			}
			d := (inSlot - outSlot) % slots
			if d < 0 {
				d += slots
			}
			row, ok := diagonals[d]
			if !ok {
				row = make([]complex128, slots)
				diagonals[d] = row
			}
			row[outSlot] = 1
		}
	}
	return blt.CompileDiagonalsWithOptions(diagonals, diagonalOnly, bias, outSlots, inSlots, ctx, inputLevel, blt.CompileOptions{})
}

func (p *packedTensorPlan) eval(ev *Evaluator, inputs []*rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	if len(inputs) != p.arity {
		return nil, fmt.Errorf("got %d inputs, expected %d", len(inputs), p.arity)
	}
	ev.Ctx.EnsureGaloisKeys(p.allGalois)
	reg := make([]*rlwe.Ciphertext, p.arity+len(p.mulSchedule))
	if err := ev.parallelApplyRaw(inputs, p.spreads, reg[:p.arity]); err != nil {
		return nil, err
	}
	for _, group := range p.mulLevels {
		lefts := make([]*rlwe.Ciphertext, len(group))
		rights := make([]*rlwe.Ciphertext, len(group))
		for j, stepIdx := range group {
			step := p.mulSchedule[stepIdx]
			lefts[j] = reg[step.Left]
			rights[j] = reg[step.Right]
		}
		prods := make([]*rlwe.Ciphertext, len(group))
		if err := ev.parallelMulRelin(lefts, rights, prods); err != nil {
			return nil, err
		}
		for j, stepIdx := range group {
			reg[p.arity+stepIdx] = prods[j]
		}
	}
	w := ev.BLT.GetWorker()
	defer ev.BLT.PutWorker(w)
	return ev.BLT.ApplyRawWith(w, reg[len(reg)-1], p.finalLT)
}

func benchBasicTensorLUT(b *testing.B, enc basicEncoding, log2t, arity int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	featureSlots := 1
	for i := 0; i < arity; i++ {
		featureSlots *= spec.Alphabet.Modulus
	}
	batch := basicBatch(b, spec, featureSlots)
	levels := 3
	if arity > 2 {
		levels = 4
	}
	ctx := newBasicContext(b, levels)
	ev := NewEvaluatorWithWorkerCapacity(ctx, 8)
	table := randomLUTTable(t, arity)
	fn := func(xs []int) int { return table[lutTableIndex(t, xs)] }
	plan, err := compilePackedTensor(ctx, spec, arity, batch, ctx.Params.MaxLevel(), fn)
	if err != nil {
		b.Fatal(err)
	}
	stride := featureSlots
	inputValues := make([][]int, arity)
	inputPlain := make([][]complex128, arity)
	inputCTs := make([]*rlwe.Ciphertext, arity)
	for i := 0; i < arity; i++ {
		inputValues[i] = randValues(t, batch, int64(31+i))
		inputPlain[i], _ = encodePackedBlocksStrided(spec, inputValues[i], stride, ctx.Params.MaxSlots())
		inputCTs[i], err = encryptVector(ctx, inputPlain[i], ctx.Params.MaxLevel())
		if err != nil {
			b.Fatal(err)
		}
	}
	wantValues := make([]int, batch)
	tmp := make([]int, arity)
	for block := 0; block < batch; block++ {
		for i := 0; i < arity; i++ {
			tmp[i] = inputValues[i][block]
		}
		wantValues[block] = fn(tmp)
	}
	wantPlain, _ := encodePackedBlocksStrided(spec, wantValues, stride, ctx.Params.MaxSlots())
	inDecoded, _ := decryptVector(ctx, inputCTs[0])
	out, err := plan.eval(ev, inputCTs)
	if err != nil {
		b.Fatal(err)
	}
	outDecoded, err := decryptVector(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	inPrec, inSat := precisionBits(maxAbsErrorStrided(inDecoded, inputPlain[0], batch, spec.Slots, stride))
	outPrec, outSat := precisionBits(maxAbsErrorStrided(outDecoded, wantPlain, batch, spec.Slots, stride))
	correct := checkDecodedBlocksStrided(b, spec, outDecoded, wantValues, stride)
	b.ReportAllocs()
	b.ResetTimer()
	reportBasicMetrics(b, ctx, spec, log2t, t, batch, levels, inPrec, inSat, outPrec, outSat, correct)
	for i := 0; i < b.N; i++ {
		if _, err := plan.eval(ev, inputCTs); err != nil {
			b.Fatal(err)
		}
	}
}

func benchBasicClean(b *testing.B, enc basicEncoding, log2t int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	batch := basicBatch(b, spec, spec.Slots)
	levels := 4
	if enc.kind == charenc.IND || enc.kind == charenc.WH {
		levels = 2
	}
	ctx := newBasicContext(b, levels)
	ev := NewEvaluatorWithWorkerCapacity(ctx, 1)
	cleanPlan, err := compileBasicCleanPlan(ctx, enc.kind, spec, t, batch)
	if err != nil {
		b.Fatal(err)
	}
	xs := randValues(t, batch, 41)
	exactPlain, _ := encodePackedBlocks(spec, xs, ctx.Params.MaxSlots())
	noisyPlain := append([]complex128(nil), exactPlain...)
	for i := 0; i < batch*spec.Slots; i++ {
		noise := math.Ldexp(math.Sin(float64(i*17+3)), -12)
		noisyPlain[i] += complex(noise, -0.5*noise)
	}
	ctX, err := encryptVector(ctx, noisyPlain, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	inDecoded, _ := decryptVector(ctx, ctX)
	out, err := evalBasicClean(ev, ctX, cleanPlan)
	if err != nil {
		b.Fatal(err)
	}
	outDecoded, err := decryptVector(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	used := batch * spec.Slots
	inPrec, inSat := precisionBits(maxAbsError(inDecoded, exactPlain, used))
	outPrec, outSat := precisionBits(maxAbsError(outDecoded, exactPlain, used))
	correct := checkDecodedBlocks(b, spec, outDecoded, xs)
	b.ReportAllocs()
	b.ResetTimer()
	reportBasicCleanMetrics(b, ctx, spec, log2t, t, batch, levels, inPrec, inSat, outPrec, outSat, correct)
	for i := 0; i < b.N; i++ {
		if _, err := evalBasicClean(ev, ctX, cleanPlan); err != nil {
			b.Fatal(err)
		}
	}
}

type basicCleanPlan struct {
	kind    charenc.EncodingKind
	toIND   *blt.RawCompiled
	fromIND *blt.RawCompiled
}

func compileBasicCleanPlan(ctx *charctx.Context, kind charenc.EncodingKind, spec charenc.BlockSpec, t, batch int) (*basicCleanPlan, error) {
	if kind == charenc.IND || kind == charenc.WH {
		return &basicCleanPlan{kind: kind}, nil
	}
	indCodec, err := charenc.NewIND(t, true, 0)
	if err != nil {
		return nil, err
	}
	indSpec := indCodec.Spec()
	fnID := func(x int) int { return x }
	toIND, err := compilePackedUnary(ctx, spec, indSpec, batch, ctx.Params.MaxLevel(), fnID)
	if err != nil {
		return nil, err
	}
	fromIND, err := compilePackedUnary(ctx, indSpec, spec, batch, ctx.Params.MaxLevel()-3, fnID)
	if err != nil {
		return nil, err
	}
	return &basicCleanPlan{kind: kind, toIND: toIND.raw, fromIND: fromIND.raw}, nil
}

func evalBasicClean(ev *Evaluator, x *rlwe.Ciphertext, plan *basicCleanPlan) (*rlwe.Ciphertext, error) {
	if plan.kind == charenc.IND {
		return evalBasicINDClean(ev, x)
	}
	if plan.kind == charenc.WH {
		return evalBasicWHClean(ev, x)
	}
	ind, err := ev.BLT.ApplyRaw(x, plan.toIND)
	if err != nil {
		return nil, err
	}
	cleanIND, err := evalBasicINDClean(ev, ind)
	if err != nil {
		return nil, err
	}
	return ev.BLT.ApplyRaw(cleanIND, plan.fromIND)
}

func evalBasicINDClean(ev *Evaluator, ind *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	w := ev.BLT.GetWorker()
	x2 := hefloat.NewCiphertext(ev.Ctx.Params, 1, ind.Level())
	if err := w.Eval.MulRelin(ind, ind, x2); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	if err := w.Eval.Rescale(x2, x2); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	indAtX2 := ind
	if indAtX2.Level() > x2.Level() {
		indAtX2 = w.Eval.DropLevelNew(indAtX2, indAtX2.Level()-x2.Level())
	}
	x3 := hefloat.NewCiphertext(ev.Ctx.Params, 1, x2.Level())
	if err := w.Eval.MulRelin(x2, indAtX2, x3); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	if err := w.Eval.Rescale(x3, x3); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	x2AtX3 := x2
	if x2AtX3.Level() > x3.Level() {
		x2AtX3 = w.Eval.DropLevelNew(x2AtX3, x2AtX3.Level()-x3.Level())
	}
	term2 := hefloat.NewCiphertext(ev.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(x2AtX3, 3, term2); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	term3 := hefloat.NewCiphertext(ev.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(x3, 2, term3); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	cleanIND := hefloat.NewCiphertext(ev.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Sub(term2, term3, cleanIND); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	ev.BLT.PutWorker(w)
	return cleanIND, nil
}

func evalBasicWHClean(ev *Evaluator, x *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	w := ev.BLT.GetWorker()
	x2 := hefloat.NewCiphertext(ev.Ctx.Params, 1, x.Level())
	if err := w.Eval.MulRelin(x, x, x2); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	if err := w.Eval.Rescale(x2, x2); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	xAtX2 := x
	if xAtX2.Level() > x2.Level() {
		xAtX2 = w.Eval.DropLevelNew(xAtX2, xAtX2.Level()-x2.Level())
	}
	x3 := hefloat.NewCiphertext(ev.Ctx.Params, 1, x2.Level())
	if err := w.Eval.MulRelin(x2, xAtX2, x3); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	if err := w.Eval.Rescale(x3, x3); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	xAtX3 := x
	if xAtX3.Level() > x3.Level() {
		xAtX3 = w.Eval.DropLevelNew(xAtX3, xAtX3.Level()-x3.Level())
	}
	term1 := hefloat.NewCiphertext(ev.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(xAtX3, 3, term1); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	out := hefloat.NewCiphertext(ev.Ctx.Params, 1, x3.Level())
	if err := w.Eval.Sub(term1, x3, out); err != nil {
		ev.BLT.PutWorker(w)
		return nil, err
	}
	out.Scale = out.Scale.Mul(rlwe.NewScale(2))
	ev.BLT.PutWorker(w)
	return out, nil
}

func runBasicBenchmarks(b *testing.B, fn func(*testing.B, basicEncoding, int)) {
	for _, enc := range selectedBasicEncodings() {
		enc := enc
		b.Run(enc.name, func(b *testing.B) {
			for _, log2t := range basicLog2Ts() {
				log2t := log2t
				b.Run(fmt.Sprintf("log2t%d", log2t), func(b *testing.B) {
					b.Run("sequential", func(b *testing.B) { fn(b, enc, log2t) })
				})
			}
		})
	}
}

func BenchmarkBasicNative(b *testing.B) {
	runBasicBenchmarks(b, benchBasicNative)
}

func BenchmarkBasicUnaryLUT(b *testing.B) {
	runBasicBenchmarks(b, benchBasicUnaryLUT)
}

func BenchmarkBasicBinaryLUT(b *testing.B) {
	runBasicBenchmarks(b, func(b *testing.B, enc basicEncoding, log2t int) {
		benchBasicTensorLUT(b, enc, log2t, 2)
	})
}

func BenchmarkBasicFourLUT(b *testing.B) {
	runBasicBenchmarks(b, func(b *testing.B, enc basicEncoding, log2t int) {
		benchBasicTensorLUT(b, enc, log2t, 4)
	})
}

func BenchmarkBasicClean(b *testing.B) {
	runBasicBenchmarks(b, benchBasicClean)
}
