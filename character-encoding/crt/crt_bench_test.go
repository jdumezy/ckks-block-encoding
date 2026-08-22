package crt

import (
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/ring"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

func productionParameters() hefloat.ParametersLiteral {
	logN := 15
	if raw := os.Getenv("CRT_BENCH_LOGN"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 10 {
			panic(fmt.Sprintf("invalid CRT_BENCH_LOGN=%q", raw))
		}
		logN = parsed
	}
	levels := 4
	if raw := os.Getenv("CRT_BENCH_LEVELS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			panic(fmt.Sprintf("invalid CRT_BENCH_LEVELS=%q", raw))
		}
		levels = parsed
	}
	logQ := []int{55}
	for i := 0; i < levels; i++ {
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

func newBenchContext(b *testing.B) *charctx.Context {
	b.Helper()
	ctx, err := charctx.NewContext(productionParameters())
	if err != nil {
		b.Fatal(err)
	}
	return ctx
}

func benchInteger(bits int, salt int64) *big.Int {
	x := new(big.Int).Lsh(big.NewInt(1), uint(bits-1))
	x.Add(x, big.NewInt(salt))
	return x
}

func reportParameterMetrics(b *testing.B, ctx *charctx.Context) {
	b.ReportMetric(ctx.Params.LogQ(), "logQ")
	b.ReportMetric(ctx.Params.LogP(), "logP")
	b.ReportMetric(ctx.Params.LogQP(), "logQP")
	b.ReportMetric(256, "h")
}

func reportPackedSpecMetrics(b *testing.B, ctx *charctx.Context, spec PackedSpec) {
	b.ReportMetric(float64(spec.Base.Channels()), "channels")
	b.ReportMetric(float64(spec.Ciphertexts()), "packed_cts")
	b.ReportMetric(float64(spec.BatchSize), "words/op")
	b.ReportMetric(float64(spec.WordStride), "word_slots")
	b.ReportMetric(float64(spec.TotalUsed), "used_slots")
	b.ReportMetric(float64(spec.MaxSlots), "max_slots")
	b.ReportMetric(spec.Base.Log2T, "log2T")
	b.ReportMetric(float64(spec.Base.MaxPrime()), "maxPrime")
	b.ReportMetric(float64(runtime.GOMAXPROCS(0)), "gomaxprocs")
	reportParameterMetrics(b, ctx)
}

func reportLevels(b *testing.B, levels int) {
	b.ReportMetric(float64(levels), "levels/op")
	b.ReportMetric(1, "correct")
	b.ReportMetric(0, "max_error")
}

func benchmarkWorkers(b *testing.B, parallel bool) int {
	b.Helper()
	if !parallel {
		return 1
	}
	if raw := os.Getenv("CRT_BENCH_WORKERS"); raw != "" {
		workers, err := strconv.Atoi(raw)
		if err != nil || workers < 1 {
			b.Fatalf("invalid CRT_BENCH_WORKERS=%q", raw)
		}
		return workers
	}
	return runtime.GOMAXPROCS(0)
}

func benchmarkBits() []int {
	bitsList := []int{64, 256}
	if os.Getenv("CRT_BENCH_1024") == "1" {
		bitsList = append(bitsList, 1024)
	}
	return bitsList
}

// benchmarkPackedSpec builds a maximal word-batched packed CRT spec. The
// CRT_PACK_MAX_SLOTS env var can cap the used slots and therefore reduce the
// word batch size. CRT_PACK_TARGET_CTS is kept as a compatibility knob, but
// maximal word batching always uses one ciphertext.
func benchmarkPackedSpec(b *testing.B, ctx *charctx.Context, base Spec) PackedSpec {
	b.Helper()
	physical := ctx.Params.MaxSlots()
	if raw := os.Getenv("CRT_PACK_TARGET_CTS"); raw != "" {
		target, err := strconv.Atoi(raw)
		if err != nil || target < 1 {
			b.Fatalf("invalid CRT_PACK_TARGET_CTS=%q", raw)
		}
		spec, err := NewPackedSpecWithTargetCiphertexts(base, physical, target)
		if err != nil {
			b.Fatal(err)
		}
		return spec
	}
	if raw := os.Getenv("CRT_PACK_MAX_SLOTS"); raw != "" {
		packMax, err := strconv.Atoi(raw)
		if err != nil || packMax < 1 {
			b.Fatalf("invalid CRT_PACK_MAX_SLOTS=%q", raw)
		}
		spec, err := NewPackedSpecWithMaxUsedSlots(base, physical, packMax)
		if err != nil {
			b.Fatal(err)
		}
		return spec
	}
	spec, err := NewPackedSpec(base, physical)
	if err != nil {
		b.Fatal(err)
	}
	return spec
}

// benchmarkLTOptions reads CRT_LT_STRATEGY and CRT_LT_GIANT_STEP and returns
// the resulting blt.CompileOptions. The default is BSGS because packed switch
// latency is the main benchmark target.
func benchmarkLTOptions(b *testing.B) blt.CompileOptions {
	b.Helper()
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("CRT_LT_STRATEGY")))
	switch raw {
	case "", "bsgs":
		return blt.CompileOptions{Strategy: blt.LTBSGS}
	case "naive":
		return blt.CompileOptions{Strategy: blt.LTNaive}
	case "bsgs-giant", "giant":
		gsRaw := os.Getenv("CRT_LT_GIANT_STEP")
		gs, err := strconv.Atoi(gsRaw)
		if err != nil || gs < 1 {
			b.Fatalf("CRT_LT_STRATEGY=%q requires CRT_LT_GIANT_STEP>=1, got %q", raw, gsRaw)
		}
		return blt.CompileOptions{Strategy: blt.LTBSGSWithGiantStep, GiantStep: gs}
	default:
		b.Fatalf("invalid CRT_LT_STRATEGY=%q (want naive|bsgs|bsgs-giant)", raw)
	}
	return blt.CompileOptions{}
}

func reportLTOptions(b *testing.B, opts blt.CompileOptions) {
	b.Helper()
	b.ReportMetric(float64(opts.Strategy), "lt_strategy")
	if opts.Strategy == blt.LTBSGSWithGiantStep {
		b.ReportMetric(float64(opts.GiantStep), "lt_giant_step")
	}
}

func encryptedPackedOperands(b *testing.B, ctx *charctx.Context, spec PackedSpec) (PackedCiphertext, PackedCiphertext, []*big.Int, []*big.Int) {
	b.Helper()
	level := ctx.Params.MaxLevel()
	xs := make([]*big.Int, spec.BatchSize)
	ys := make([]*big.Int, spec.BatchSize)
	for i := 0; i < spec.BatchSize; i++ {
		xs[i] = benchInteger(spec.Base.Bits, int64(0x12345+i))
		ys[i] = benchInteger(spec.Base.Bits, int64(0x6789+3*i))
	}
	px, err := EncodePackedBatch(spec, xs)
	if err != nil {
		b.Fatal(err)
	}
	py, err := EncodePackedBatch(spec, ys)
	if err != nil {
		b.Fatal(err)
	}
	cx, err := EncryptPacked(ctx, px, level)
	if err != nil {
		b.Fatal(err)
	}
	cy, err := EncryptPacked(ctx, py, level)
	if err != nil {
		b.Fatal(err)
	}
	return cx, cy, xs, ys
}

func encryptedPackedOperand(b *testing.B, ctx *charctx.Context, spec PackedSpec) (PackedCiphertext, []*big.Int) {
	b.Helper()
	level := ctx.Params.MaxLevel()
	xs := make([]*big.Int, spec.BatchSize)
	for i := 0; i < spec.BatchSize; i++ {
		xs[i] = benchInteger(spec.Base.Bits, int64(0x12345+i))
	}
	px, err := EncodePackedBatch(spec, xs)
	if err != nil {
		b.Fatal(err)
	}
	cx, err := EncryptPacked(ctx, px, level)
	if err != nil {
		b.Fatal(err)
	}
	return cx, xs
}

func checkPackedResidues(b *testing.B, ctx *charctx.Context, ct PackedCiphertext, want func(word, channel, p int) int) {
	b.Helper()
	pt, err := DecryptPacked(ctx, ct)
	if err != nil {
		b.Fatal(err)
	}
	residues, err := DecodePackedBatchResidues(pt)
	if err != nil {
		b.Fatal(err)
	}
	for w := range residues {
		for ch, p := range ct.Spec.Base.Primes {
			expected := positiveMod(want(w, ch, p), p)
			if residues[w][ch] != expected {
				b.Fatalf("word %d channel %d residue=%d, want %d", w, ch, residues[w][ch], expected)
			}
		}
	}
}

func BenchmarkPackedCRTOnlineNativeBRUAdd(b *testing.B) {
	for _, bits := range benchmarkBits() {
		for _, parallel := range []bool{false, true} {
			name := fmt.Sprintf("%dbits", bits)
			if parallel {
				name += "/parallel"
			} else {
				name += "/sequential"
			}
			b.Run(name, func(b *testing.B) {
				ctx := newBenchContext(b)
				base, err := NewSpecForBits(bits, charenc.BRU, 0)
				if err != nil {
					b.Fatal(err)
				}
				spec := benchmarkPackedSpec(b, ctx, base)
				cx, cy, xs, ys := encryptedPackedOperands(b, ctx, spec)
				workers := benchmarkWorkers(b, parallel)
				ev := NewEvaluator(ctx, workers)
				out, err := ev.EvalPackedNativeProduct(cx, cy, parallel)
				if err != nil {
					b.Fatal(err)
				}
				checkPackedResidues(b, ctx, out, func(word, _ int, p int) int {
					x := int(new(big.Int).Mod(xs[word], big.NewInt(int64(p))).Int64())
					y := int(new(big.Int).Mod(ys[word], big.NewInt(int64(p))).Int64())
					return x + y
				})
				b.ReportAllocs()
				b.ResetTimer()
				reportPackedSpecMetrics(b, ctx, spec)
				reportLevels(b, 1)
				b.ReportMetric(float64(workers), "workers")
				for i := 0; i < b.N; i++ {
					if _, err := ev.EvalPackedNativeProduct(cx, cy, parallel); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkPackedCRTOnlineNativeBRUSub(b *testing.B) {
	for _, bits := range benchmarkBits() {
		for _, parallel := range []bool{false, true} {
			name := fmt.Sprintf("%dbits", bits)
			if parallel {
				name += "/parallel"
			} else {
				name += "/sequential"
			}
			b.Run(name, func(b *testing.B) {
				ctx := newBenchContext(b)
				ctx.EnsureGaloisKeys([]uint64{ctx.Params.GaloisElementForComplexConjugation()})
				base, err := NewSpecForBits(bits, charenc.BRU, 0)
				if err != nil {
					b.Fatal(err)
				}
				spec := benchmarkPackedSpec(b, ctx, base)
				cx, cy, xs, ys := encryptedPackedOperands(b, ctx, spec)
				workers := benchmarkWorkers(b, parallel)
				ev := NewEvaluator(ctx, workers)
				out, err := evalPackedBRUSub(ev, cx, cy, parallel)
				if err != nil {
					b.Fatal(err)
				}
				checkPackedResidues(b, ctx, out, func(word, _ int, p int) int {
					x := int(new(big.Int).Mod(xs[word], big.NewInt(int64(p))).Int64())
					y := int(new(big.Int).Mod(ys[word], big.NewInt(int64(p))).Int64())
					return x - y
				})
				b.ReportAllocs()
				b.ResetTimer()
				reportPackedSpecMetrics(b, ctx, spec)
				reportLevels(b, 1)
				b.ReportMetric(float64(workers), "workers")
				for i := 0; i < b.N; i++ {
					if _, err := evalPackedBRUSub(ev, cx, cy, parallel); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func evalPackedBRUSub(ev *Evaluator, x, y PackedCiphertext, parallel bool) (PackedCiphertext, error) {
	if err := checkPackedBinaryOperands(x, y); err != nil {
		return PackedCiphertext{}, err
	}
	out := PackedCiphertext{Spec: x.Spec, CTs: make([]*rlwe.Ciphertext, len(x.CTs))}
	workers := channelParallelism(parallel, len(x.CTs), ev.LUT.BLT.Capacity())
	if workers == 1 {
		w := ev.LUT.BLT.GetWorker()
		defer ev.LUT.BLT.PutWorker(w)
		for i := range x.CTs {
			ct, err := evalPackedBRUSubCiphertext(ev, w, x.CTs[i], y.CTs[i])
			if err != nil {
				return PackedCiphertext{}, fmt.Errorf("crt.evalPackedBRUSub: ciphertext %d: %w", i, err)
			}
			out.CTs[i] = ct
		}
		return out, nil
	}

	errs := make([]error, len(x.CTs))
	done := make(chan struct{}, workers)
	for workerID := 0; workerID < workers; workerID++ {
		go func(workerID int) {
			w := ev.LUT.BLT.GetWorker()
			defer ev.LUT.BLT.PutWorker(w)
			defer func() { done <- struct{}{} }()
			for i := workerID; i < len(x.CTs); i += workers {
				ct, err := evalPackedBRUSubCiphertext(ev, w, x.CTs[i], y.CTs[i])
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
			return PackedCiphertext{}, fmt.Errorf("crt.evalPackedBRUSub: ciphertext %d: %w", i, err)
		}
	}
	return out, nil
}

func evalPackedBRUSubCiphertext(ev *Evaluator, w *blt.Worker, x, y *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	conjY, err := w.Eval.ConjugateNew(y)
	if err != nil {
		return nil, err
	}
	return ev.evalPackedNativeCiphertext(w, x, conjY)
}

func BenchmarkPackedCRTOnlineNativeLBRUMul(b *testing.B) {
	for _, bits := range benchmarkBits() {
		for _, parallel := range []bool{false, true} {
			name := fmt.Sprintf("%dbits", bits)
			if parallel {
				name += "/parallel"
			} else {
				name += "/sequential"
			}
			b.Run(name, func(b *testing.B) {
				ctx := newBenchContext(b)
				base, err := NewSpecForBits(bits, charenc.LBRU, 0)
				if err != nil {
					b.Fatal(err)
				}
				spec := benchmarkPackedSpec(b, ctx, base)
				cx, cy, xs, ys := encryptedPackedOperands(b, ctx, spec)
				workers := benchmarkWorkers(b, parallel)
				ev := NewEvaluator(ctx, workers)
				out, err := ev.EvalPackedNativeProduct(cx, cy, parallel)
				if err != nil {
					b.Fatal(err)
				}
				checkPackedResidues(b, ctx, out, func(word, _ int, p int) int {
					x := int(new(big.Int).Mod(xs[word], big.NewInt(int64(p))).Int64())
					y := int(new(big.Int).Mod(ys[word], big.NewInt(int64(p))).Int64())
					return x * y
				})
				b.ReportAllocs()
				b.ResetTimer()
				reportPackedSpecMetrics(b, ctx, spec)
				reportLevels(b, 1)
				b.ReportMetric(float64(workers), "workers")
				for i := 0; i < b.N; i++ {
					if _, err := ev.EvalPackedNativeProduct(cx, cy, parallel); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkPackedCRTOnlineCleanBRU(b *testing.B) {
	for _, bits := range benchmarkBits() {
		for _, parallel := range []bool{false, true} {
			name := fmt.Sprintf("%dbits", bits)
			if parallel {
				name += "/parallel"
			} else {
				name += "/sequential"
			}
			b.Run(name, func(b *testing.B) {
				ctx := newBenchContext(b)
				bruBase, err := NewSpecForBits(bits, charenc.BRU, 0)
				if err != nil {
					b.Fatal(err)
				}
				bru := benchmarkPackedSpec(b, ctx, bruBase)
				opts := benchmarkLTOptions(b)
				level := ctx.Params.MaxLevel()
				clean, err := CompilePackedClean(ctx, bru, level, opts)
				if err != nil {
					b.Fatal(err)
				}
				cx, xs := encryptedPackedOperand(b, ctx, bru)
				workers := benchmarkWorkers(b, parallel)
				ev := NewEvaluator(ctx, workers)
				out, err := ev.EvalPackedClean(cx, clean, parallel)
				if err != nil {
					b.Fatal(err)
				}
				checkPackedResidues(b, ctx, out, func(word, _ int, p int) int {
					return int(new(big.Int).Mod(xs[word], big.NewInt(int64(p))).Int64())
				})
				b.ReportAllocs()
				b.ResetTimer()
				reportPackedSpecMetrics(b, ctx, bru)
				reportLevels(b, 4)
				reportLTOptions(b, opts)
				b.ReportMetric(float64(workers), "workers")
				for i := 0; i < b.N; i++ {
					if _, err := ev.EvalPackedClean(cx, clean, parallel); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkPackedCRTOnlineSwitch(b *testing.B) {
	for _, bits := range benchmarkBits() {
		for _, direction := range []struct {
			name string
			from charenc.EncodingKind
			to   charenc.EncodingKind
		}{
			{name: "BRUToLBRU", from: charenc.BRU, to: charenc.LBRU},
			{name: "LBRUToBRU", from: charenc.LBRU, to: charenc.BRU},
		} {
			for _, parallel := range []bool{false, true} {
				name := fmt.Sprintf("%s/%dbits", direction.name, bits)
				if parallel {
					name += "/parallel"
				} else {
					name += "/sequential"
				}
				b.Run(name, func(b *testing.B) {
					ctx := newBenchContext(b)
					inBase, err := NewSpecForBits(bits, direction.from, 0)
					if err != nil {
						b.Fatal(err)
					}
					in := benchmarkPackedSpec(b, ctx, inBase)
					out, err := SwitchPackedSpec(in, direction.to)
					if err != nil {
						b.Fatal(err)
					}
					opts := benchmarkLTOptions(b)
					level := ctx.Params.MaxLevel()
					compiled, err := CompilePackedSwitchWithOptions(ctx, in, out, level, opts)
					if err != nil {
						b.Fatal(err)
					}
					cx, xs := encryptedPackedOperand(b, ctx, in)
					workers := benchmarkWorkers(b, parallel)
					ev := NewEvaluator(ctx, workers)
					outCT, err := ev.EvalPackedSwitch(cx, compiled, parallel)
					if err != nil {
						b.Fatal(err)
					}
					checkPackedResidues(b, ctx, outCT, func(word, _ int, p int) int {
						return int(new(big.Int).Mod(xs[word], big.NewInt(int64(p))).Int64())
					})
					b.ReportAllocs()
					b.ResetTimer()
					reportPackedSpecMetrics(b, ctx, in)
					reportLevels(b, 1)
					reportLTOptions(b, opts)
					b.ReportMetric(float64(workers), "workers")
					for i := 0; i < b.N; i++ {
						if _, err := ev.EvalPackedSwitch(cx, compiled, parallel); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}
