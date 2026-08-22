package radix

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/ring"

	"character-encoding/character-encoding/charctx"
)

// packedBenchParams returns CKKS parameters sized for the depth of the
// op under bench. chainLen sets the level count -- pick the tight value
// from addSubChainLen / cmpChainLen.
func packedBenchParams(logN, chainLen int) hefloat.ParametersLiteral {
	logQ := []int{60}
	for i := 0; i < chainLen-1; i++ {
		logQ = append(logQ, 40)
	}
	return hefloat.ParametersLiteral{
		LogN:            logN,
		LogQ:            logQ,
		LogP:            []int{60},
		Xs:              &ring.Ternary{H: 256},
		LogDefaultScale: 40,
	}
}

type opShape struct {
	name         string
	wordBits     int
	bitsPerDigit int
	logN         int
}

func opShapesAt(logN int) []opShape {
	shapes := []opShape{
		{"64bit_r4", 64, 2, logN},
		{"256bit_r4", 256, 2, logN},
		{"64bit_r16", 64, 4, logN},
	}
	if logN >= 16 {
		shapes = append(shapes, opShape{"256bit_r16", 256, 4, logN})
	}
	return shapes
}

// chainLen for Add/Sub at a given d.
func addSubChainLen(d int) int { return 4 + ceilLog2(d) + 1 }

// chainLen for Eq/Lt/Cmp at a given d.
func cmpChainLen(d int) int { return 3 + ceilLog2(d) + 1 }

func compileSpec(b *testing.B, shape opShape, chainLen int, batch int) (*charctx.Context, *Evaluator, Spec, PaddedSpec) {
	b.Helper()
	spec, err := NewSpec(shape.wordBits, shape.bitsPerDigit)
	if err != nil {
		b.Fatal(err)
	}
	ctx, err := charctx.NewContext(packedBenchParams(shape.logN, chainLen))
	if err != nil {
		b.Fatal(err)
	}
	ev := NewEvaluator(ctx)
	if batch == 0 {
		batch = MaxBatchSize(spec, ctx.Params.MaxSlots())
	}
	padded, err := NewPaddedSpecBatched(spec, ctx.Params.MaxSlots(), batch)
	if err != nil {
		b.Fatalf("NewPaddedSpecBatched(batch=%d): %v", batch, err)
	}
	return ctx, ev, spec, padded
}

func randInputs(spec Spec, batch int, seed int64) ([]*big.Int, []*big.Int) {
	rng := rand.New(rand.NewSource(seed))
	mod := spec.Modulus()
	xs := make([]*big.Int, batch)
	ys := make([]*big.Int, batch)
	for i := 0; i < batch; i++ {
		xs[i] = new(big.Int).Rand(rng, mod)
		ys[i] = new(big.Int).Rand(rng, mod)
	}
	return xs, ys
}

func reportParameterMetrics(b *testing.B, ctx *charctx.Context) {
	b.ReportMetric(ctx.Params.LogQ(), "logQ")
	b.ReportMetric(ctx.Params.LogP(), "logP")
	b.ReportMetric(ctx.Params.LogQP(), "logQP")
	b.ReportMetric(256, "h")
}

func reportPackedMetrics(b *testing.B, ctx *charctx.Context, spec PaddedSpec, levels int) {
	b.ReportMetric(float64(spec.BatchSize), "words/op")
	b.ReportMetric(float64(levels), "levels/op")
	reportParameterMetrics(b, ctx)
}

func reportCorrect(b *testing.B, correct bool) {
	if correct {
		b.ReportMetric(1, "correct")
	} else {
		b.ReportMetric(0, "correct")
	}
}

func reportBoolAccuracy(b *testing.B, correct bool) {
	if correct {
		b.ReportMetric(0, "max_error")
	} else {
		b.ReportMetric(1, "max_error")
	}
	reportCorrect(b, correct)
}

func reportPaddedAccuracy(b *testing.B, maxErr *big.Int) {
	b.ReportMetric(float64(maxErr.Int64()), "max_error")
	reportCorrect(b, maxErr.Sign() == 0)
}

func paddedMaxError(b *testing.B, ctx *charctx.Context, ct PaddedCiphertext, want []*big.Int) *big.Int {
	b.Helper()
	pt, err := DecryptPadded(ctx, ct)
	if err != nil {
		b.Fatal(err)
	}
	got, err := DecodePaddedBatch(pt)
	if err != nil {
		b.Fatal(err)
	}
	if len(got) != len(want) {
		b.Fatalf("decoded %d words, want %d", len(got), len(want))
	}
	maxErr := big.NewInt(0)
	for i := range got {
		err := new(big.Int).Sub(got[i], want[i])
		err.Abs(err)
		if err.Cmp(maxErr) > 0 {
			maxErr = err
		}
	}
	return maxErr
}

func bitsCorrect(b *testing.B, got []bool, want []bool) bool {
	b.Helper()
	if len(got) != len(want) {
		b.Fatalf("decoded %d bits, want %d", len(got), len(want))
	}
	correct := true
	for i := range got {
		if got[i] != want[i] {
			correct = false
			break
		}
	}
	return correct
}

func benchAddAtShape(b *testing.B, shape opShape, batch int) {
	levels := addSubChainLen(spec_widthFor(shape)) - 1
	ctx, ev, spec, padded := compileSpec(b, shape, levels+1, batch)
	plan, err := ev.CompileAdd(padded, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xs, ys := randInputs(spec, padded.BatchSize, 1)
	ptX, _ := EncodePaddedBatch(padded, xs)
	ptY, _ := EncodePaddedBatch(padded, ys)
	ctX, _ := EncryptPadded(ctx, ptX, ctx.Params.MaxLevel())
	ctY, _ := EncryptPadded(ctx, ptY, ctx.Params.MaxLevel())
	out, err := ev.Add(ctX, ctY, plan)
	if err != nil {
		b.Fatal(err)
	}
	want := make([]*big.Int, len(xs))
	for i := range want {
		want[i], _ = SimAdd(spec, xs[i], ys[i])
	}
	maxErr := paddedMaxError(b, ctx, out, want)
	b.ResetTimer()
	reportPackedMetrics(b, ctx, padded, levels)
	reportPaddedAccuracy(b, maxErr)
	for i := 0; i < b.N; i++ {
		if _, err := ev.Add(ctX, ctY, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func benchSubAtShape(b *testing.B, shape opShape, batch int) {
	levels := addSubChainLen(spec_widthFor(shape)) - 1
	ctx, ev, spec, padded := compileSpec(b, shape, levels+1, batch)
	plan, err := ev.CompileSub(padded, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xs, ys := randInputs(spec, padded.BatchSize, 2)
	ptX, _ := EncodePaddedBatch(padded, xs)
	ptY, _ := EncodePaddedBatch(padded, ys)
	ctX, _ := EncryptPadded(ctx, ptX, ctx.Params.MaxLevel())
	ctY, _ := EncryptPadded(ctx, ptY, ctx.Params.MaxLevel())
	out, err := ev.Sub(ctX, ctY, plan)
	if err != nil {
		b.Fatal(err)
	}
	want := make([]*big.Int, len(xs))
	for i := range want {
		want[i], _ = SimSub(spec, xs[i], ys[i])
	}
	maxErr := paddedMaxError(b, ctx, out, want)
	b.ResetTimer()
	reportPackedMetrics(b, ctx, padded, levels)
	reportPaddedAccuracy(b, maxErr)
	for i := 0; i < b.N; i++ {
		if _, err := ev.Sub(ctX, ctY, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func benchEqAtShape(b *testing.B, shape opShape, batch int) {
	levels := cmpChainLen(spec_widthFor(shape)) - 1
	ctx, ev, spec, padded := compileSpec(b, shape, levels+1, batch)
	plan, err := ev.CompileEq(padded, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xs, ys := randInputs(spec, padded.BatchSize, 3)
	ptX, _ := EncodePaddedBatch(padded, xs)
	ptY, _ := EncodePaddedBatch(padded, ys)
	ctX, _ := EncryptPadded(ctx, ptX, ctx.Params.MaxLevel())
	ctY, _ := EncryptPadded(ctx, ptY, ctx.Params.MaxLevel())
	out, err := ev.Eq(ctX, ctY, plan)
	if err != nil {
		b.Fatal(err)
	}
	got, err := ev.DecryptEq(out)
	if err != nil {
		b.Fatal(err)
	}
	want := make([]bool, len(xs))
	for i := range want {
		v, _ := SimEq(spec, xs[i], ys[i])
		want[i] = v == 1
	}
	correct := bitsCorrect(b, got, want)
	b.ResetTimer()
	reportPackedMetrics(b, ctx, padded, levels)
	reportBoolAccuracy(b, correct)
	for i := 0; i < b.N; i++ {
		if _, err := ev.Eq(ctX, ctY, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func benchLtAtShape(b *testing.B, shape opShape, batch int) {
	levels := cmpChainLen(spec_widthFor(shape)) - 1
	ctx, ev, spec, padded := compileSpec(b, shape, levels+1, batch)
	plan, err := ev.CompileLt(padded, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xs, ys := randInputs(spec, padded.BatchSize, 4)
	ptX, _ := EncodePaddedBatch(padded, xs)
	ptY, _ := EncodePaddedBatch(padded, ys)
	ctX, _ := EncryptPadded(ctx, ptX, ctx.Params.MaxLevel())
	ctY, _ := EncryptPadded(ctx, ptY, ctx.Params.MaxLevel())
	out, err := ev.Lt(ctX, ctY, plan)
	if err != nil {
		b.Fatal(err)
	}
	got, err := ev.DecryptLt(out)
	if err != nil {
		b.Fatal(err)
	}
	want := make([]bool, len(xs))
	for i := range want {
		v, _ := SimCmp(spec, xs[i], ys[i])
		want[i] = v[1] == 1
	}
	correct := bitsCorrect(b, got, want)
	b.ResetTimer()
	reportPackedMetrics(b, ctx, padded, levels)
	reportBoolAccuracy(b, correct)
	for i := 0; i < b.N; i++ {
		if _, err := ev.Lt(ctX, ctY, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func benchCmpAtShape(b *testing.B, shape opShape, batch int) {
	levels := cmpChainLen(spec_widthFor(shape)) - 1
	ctx, ev, spec, padded := compileSpec(b, shape, levels+1, batch)
	plan, err := ev.CompileCmp(padded, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xs, ys := randInputs(spec, padded.BatchSize, 5)
	ptX, _ := EncodePaddedBatch(padded, xs)
	ptY, _ := EncodePaddedBatch(padded, ys)
	ctX, _ := EncryptPadded(ctx, ptX, ctx.Params.MaxLevel())
	ctY, _ := EncryptPadded(ctx, ptY, ctx.Params.MaxLevel())
	out, err := ev.Cmp(ctX, ctY, plan)
	if err != nil {
		b.Fatal(err)
	}
	gotEq, gotLt, gotGt, err := ev.DecryptCmp(out)
	if err != nil {
		b.Fatal(err)
	}
	wantEq := make([]bool, len(xs))
	wantLt := make([]bool, len(xs))
	wantGt := make([]bool, len(xs))
	for i := range xs {
		v, _ := SimCmp(spec, xs[i], ys[i])
		wantEq[i] = v[0] == 1
		wantLt[i] = v[1] == 1
		wantGt[i] = v[2] == 1
	}
	correct := bitsCorrect(b, gotEq, wantEq) && bitsCorrect(b, gotLt, wantLt) && bitsCorrect(b, gotGt, wantGt)
	b.ResetTimer()
	reportPackedMetrics(b, ctx, padded, levels)
	reportBoolAccuracy(b, correct)
	for i := 0; i < b.N; i++ {
		if _, err := ev.Cmp(ctX, ctY, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func cleanChainLen() int { return 5 }

func benchCleanAtShape(b *testing.B, shape opShape, batch int) {
	levels := cleanChainLen() - 1
	ctx, ev, spec, padded := compileSpec(b, shape, levels+1, batch)
	plan, err := ev.CompileClean(padded, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xs, _ := randInputs(spec, padded.BatchSize, 6)
	ptX, _ := EncodePaddedBatch(padded, xs)
	ctX, _ := EncryptPadded(ctx, ptX, ctx.Params.MaxLevel())
	out, err := ev.Clean(ctX, plan)
	if err != nil {
		b.Fatal(err)
	}
	maxErr := paddedMaxError(b, ctx, out, xs)
	b.ResetTimer()
	reportPackedMetrics(b, ctx, padded, levels)
	reportPaddedAccuracy(b, maxErr)
	for i := 0; i < b.N; i++ {
		if _, err := ev.Clean(ctX, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func spec_widthFor(shape opShape) int { return shape.wordBits / shape.bitsPerDigit }

// LogN = 15, h = 256, max LogQP = 774 at 128-bit security.
func BenchmarkAdd_LogN15_64R4(b *testing.B)    { benchAddAtShape(b, opShapesAt(15)[0], 0) }
func BenchmarkAdd_LogN15_256R4(b *testing.B)   { benchAddAtShape(b, opShapesAt(15)[1], 0) }
func BenchmarkAdd_LogN15_64R16(b *testing.B)   { benchAddAtShape(b, opShapesAt(15)[2], 0) }
func BenchmarkSub_LogN15_64R4(b *testing.B)    { benchSubAtShape(b, opShapesAt(15)[0], 0) }
func BenchmarkSub_LogN15_256R4(b *testing.B)   { benchSubAtShape(b, opShapesAt(15)[1], 0) }
func BenchmarkSub_LogN15_64R16(b *testing.B)   { benchSubAtShape(b, opShapesAt(15)[2], 0) }
func BenchmarkEq_LogN15_64R4(b *testing.B)     { benchEqAtShape(b, opShapesAt(15)[0], 0) }
func BenchmarkEq_LogN15_256R4(b *testing.B)    { benchEqAtShape(b, opShapesAt(15)[1], 0) }
func BenchmarkEq_LogN15_64R16(b *testing.B)    { benchEqAtShape(b, opShapesAt(15)[2], 0) }
func BenchmarkLt_LogN15_64R4(b *testing.B)     { benchLtAtShape(b, opShapesAt(15)[0], 0) }
func BenchmarkLt_LogN15_256R4(b *testing.B)    { benchLtAtShape(b, opShapesAt(15)[1], 0) }
func BenchmarkLt_LogN15_64R16(b *testing.B)    { benchLtAtShape(b, opShapesAt(15)[2], 0) }
func BenchmarkCmp_LogN15_64R4(b *testing.B)    { benchCmpAtShape(b, opShapesAt(15)[0], 0) }
func BenchmarkCmp_LogN15_256R4(b *testing.B)   { benchCmpAtShape(b, opShapesAt(15)[1], 0) }
func BenchmarkCmp_LogN15_64R16(b *testing.B)   { benchCmpAtShape(b, opShapesAt(15)[2], 0) }
func BenchmarkClean_LogN15_64R4(b *testing.B)  { benchCleanAtShape(b, opShapesAt(15)[0], 0) }
func BenchmarkClean_LogN15_256R4(b *testing.B) { benchCleanAtShape(b, opShapesAt(15)[1], 0) }
func BenchmarkClean_LogN15_64R16(b *testing.B) { benchCleanAtShape(b, opShapesAt(15)[2], 0) }

// LogN = 16, h = 256, max LogQP = 1553 at 128-bit security.
func BenchmarkAdd_LogN16_64R4(b *testing.B)     { benchAddAtShape(b, opShapesAt(16)[0], 0) }
func BenchmarkAdd_LogN16_256R4(b *testing.B)    { benchAddAtShape(b, opShapesAt(16)[1], 0) }
func BenchmarkAdd_LogN16_64R16(b *testing.B)    { benchAddAtShape(b, opShapesAt(16)[2], 0) }
func BenchmarkAdd_LogN16_256R16(b *testing.B)   { benchAddAtShape(b, opShapesAt(16)[3], 0) }
func BenchmarkSub_LogN16_64R4(b *testing.B)     { benchSubAtShape(b, opShapesAt(16)[0], 0) }
func BenchmarkSub_LogN16_256R4(b *testing.B)    { benchSubAtShape(b, opShapesAt(16)[1], 0) }
func BenchmarkSub_LogN16_64R16(b *testing.B)    { benchSubAtShape(b, opShapesAt(16)[2], 0) }
func BenchmarkSub_LogN16_256R16(b *testing.B)   { benchSubAtShape(b, opShapesAt(16)[3], 0) }
func BenchmarkEq_LogN16_64R4(b *testing.B)      { benchEqAtShape(b, opShapesAt(16)[0], 0) }
func BenchmarkEq_LogN16_256R4(b *testing.B)     { benchEqAtShape(b, opShapesAt(16)[1], 0) }
func BenchmarkEq_LogN16_64R16(b *testing.B)     { benchEqAtShape(b, opShapesAt(16)[2], 0) }
func BenchmarkEq_LogN16_256R16(b *testing.B)    { benchEqAtShape(b, opShapesAt(16)[3], 0) }
func BenchmarkLt_LogN16_64R4(b *testing.B)      { benchLtAtShape(b, opShapesAt(16)[0], 0) }
func BenchmarkLt_LogN16_256R4(b *testing.B)     { benchLtAtShape(b, opShapesAt(16)[1], 0) }
func BenchmarkLt_LogN16_64R16(b *testing.B)     { benchLtAtShape(b, opShapesAt(16)[2], 0) }
func BenchmarkLt_LogN16_256R16(b *testing.B)    { benchLtAtShape(b, opShapesAt(16)[3], 0) }
func BenchmarkCmp_LogN16_64R4(b *testing.B)     { benchCmpAtShape(b, opShapesAt(16)[0], 0) }
func BenchmarkCmp_LogN16_256R4(b *testing.B)    { benchCmpAtShape(b, opShapesAt(16)[1], 0) }
func BenchmarkCmp_LogN16_64R16(b *testing.B)    { benchCmpAtShape(b, opShapesAt(16)[2], 0) }
func BenchmarkCmp_LogN16_256R16(b *testing.B)   { benchCmpAtShape(b, opShapesAt(16)[3], 0) }
func BenchmarkClean_LogN16_64R4(b *testing.B)   { benchCleanAtShape(b, opShapesAt(16)[0], 0) }
func BenchmarkClean_LogN16_256R4(b *testing.B)  { benchCleanAtShape(b, opShapesAt(16)[1], 0) }
func BenchmarkClean_LogN16_64R16(b *testing.B)  { benchCleanAtShape(b, opShapesAt(16)[2], 0) }
func BenchmarkClean_LogN16_256R16(b *testing.B) { benchCleanAtShape(b, opShapesAt(16)[3], 0) }
