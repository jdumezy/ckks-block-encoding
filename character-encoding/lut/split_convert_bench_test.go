package lut

import (
	"fmt"
	"math/big"
	"math/cmplx"
	"runtime"
	"testing"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/he/hefloat/bootstrapping"
	"github.com/Pro7ech/lattigo/ring"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

func benchSplitToStandard(b *testing.B, enc basicEncoding, log2t int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	ctx := newBasicContext(b, 1)
	ev := NewEvaluatorWithWorkerCapacity(ctx, splitWorkerCapacity())
	batch := splitBatch(b)

	plan, err := CompileSplitToStandard(ctx, spec, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}

	xs := randValues(t, batch, 51)
	xCTs, err := encryptSplitInputs(ctx, spec, xs, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}

	out, err := EvalSplitToStandard(ev, xCTs, plan)
	if err != nil {
		b.Fatal(err)
	}

	decoded, err := decryptVector(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	correct := true
	for w, v := range xs {
		got := int(real(decoded[w]) + 0.5)
		if got != positiveMod(v, t) {
			correct = false
			break
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.ReportMetric(float64(batch), "words/op")
	b.ReportMetric(float64(spec.Slots), "split_cts")
	b.ReportMetric(float64(log2t), "requested_log2t")
	b.ReportMetric(float64(1), "levels/op")
	b.ReportMetric(float64(runtime.GOMAXPROCS(0)), "gomaxprocs")
	b.ReportMetric(ctx.Params.LogQ(), "logQ")
	b.ReportMetric(ctx.Params.LogP(), "logP")
	b.ReportMetric(ctx.Params.LogQP(), "logQP")
	b.ReportMetric(float64(out.Level()), "output_level")
	if correct {
		b.ReportMetric(1, "correct")
		b.ReportMetric(0, "max_error")
	} else {
		b.ReportMetric(0, "correct")
		b.ReportMetric(1, "max_error")
	}
	for i := 0; i < b.N; i++ {
		if _, err := EvalSplitToStandard(ev, xCTs, plan); err != nil {
			b.Fatal(err)
		}
	}
}

const splitFromStandardLogN = 15

func splitFromStandardBootstrappingParameters(tb testing.TB, t int) bootstrapping.Parameters {
	tb.Helper()
	logN := splitFromStandardLogN

	q0 := []int{50}
	qCircuit := []int{40, 40, 40, 40, 40, 40}
	qEvalMod := []int{40, 40, 40, 40, 40, 40, 40}
	qCoeffsToSlots := []int{40, 40}

	logQ := append([]int{}, q0...)
	logQ = append(logQ, qCircuit...)
	logQ = append(logQ, qEvalMod...)
	logQ = append(logQ, qCoeffsToSlots...)

	logP := []int{40, 40, 40}

	params, err := hefloat.NewParametersFromLiteral(hefloat.ParametersLiteral{
		LogN:            logN,
		LogQ:            logQ,
		LogP:            logP,
		LogDefaultScale: 40,
		Xs:              &ring.Ternary{H: 256},
	})
	if err != nil {
		tb.Fatal(err)
	}

	coeffsToSlots := hefloat.DFTMatrixLiteral{
		Type:     hefloat.HomomorphicEncode,
		Format:   hefloat.RepackImagAsReal,
		LogSlots: params.LogMaxSlots(),
		LevelQ:   params.MaxLevel(),
		LevelP:   params.MaxLevelP(),
		Levels:   []int{1, 1},
	}

	slotsToCoeffs := hefloat.DFTMatrixLiteral{
		Type:     hefloat.HomomorphicDecode,
		LogSlots: params.LogMaxSlots(),
		LevelQ:   params.MaxLevel(),
		LevelP:   params.MaxLevelP(),
		Levels:   []int{1, 1},
		Scaling:  new(big.Float).SetFloat64(1.0 / float64(t)),
	}

	mod1Literal := hefloat.Mod1ParametersLiteral{
		LevelQ:          params.MaxLevel() - coeffsToSlots.Depth(true),
		LogScale:        40,
		Mod1Type:        hefloat.CosDiscrete,
		Mod1Degree:      30,
		Mod1Interval:    16,
		DoubleAngle:     2,
		LogMessageRatio: 0,
	}

	return bootstrapping.Parameters{
		ResidualParameters:      params,
		BootstrappingParameters: params,
		SlotsToCoeffsParameters: slotsToCoeffs,
		CoeffsToSlotsParameters: coeffsToSlots,
		Mod1ParametersLiteral:   mod1Literal,
		EphemeralSecretWeight:   32,
	}
}

func benchSplitFromStandard(b *testing.B, enc basicEncoding, log2t int) {
	if enc.kind != charenc.BRU {
		b.Skipf("split_from_standard only targets BRU, got %v", enc.kind)
	}
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}

	btpParams := splitFromStandardBootstrappingParameters(b, t)
	params := btpParams.BootstrappingParameters

	kgen := rlwe.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPairNew()

	evk, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		b.Fatal(err)
	}

	btp, err := bootstrapping.NewEvaluator(btpParams, evk)
	if err != nil {
		b.Fatal(err)
	}

	encoder := hefloat.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, pk)
	decryptor := rlwe.NewDecryptor(params, sk)

	batch := params.MaxSlots()
	xs := randValues(t, batch, 61)
	values := make([]complex128, params.MaxSlots())
	for i := range values {
		values[i] = complex(float64(xs[i]), 0)
	}

	pt := hefloat.NewPlaintext(params, params.MaxLevel())
	if err := encoder.Encode(values, pt); err != nil {
		b.Fatal(err)
	}
	ct := hefloat.NewCiphertext(params, 1, params.MaxLevel())
	if err := encryptor.Encrypt(pt, ct); err != nil {
		b.Fatal(err)
	}

	plan, err := CompileSplitFromStandard(spec)
	if err != nil {
		b.Fatal(err)
	}

	splitCTs, err := EvalSplitFromStandard(btp, ct, plan)
	if err != nil {
		b.Fatal(err)
	}

	codec, err := charenc.NewCodec(spec)
	if err != nil {
		b.Fatal(err)
	}
	decoded := make([][]complex128, len(splitCTs))
	for c, out := range splitCTs {
		vec, err := decryptVector(&charctx.Context{Params: params, Encoder: encoder, Decryptor: decryptor}, out)
		if err != nil {
			b.Fatal(err)
		}
		decoded[c] = vec
	}

	J := spec.Slots
	coords := make([]complex128, J)
	correct := true
	maxErr := 0.0
	want := make([]complex128, J)
	for w := 0; w < batch; w++ {
		for c := 0; c < J; c++ {
			coords[c] = decoded[c][w]
		}
		encWant := codec.EncodeValue(xs[w])
		copy(want, encWant)
		for c := 0; c < J; c++ {
			if e := cmplx.Abs(coords[c] - want[c]); e > maxErr {
				maxErr = e
			}
		}
		got, err := codec.DecodeValue(coords)
		if err != nil {
			correct = false
			continue
		}
		if got != xs[w] {
			correct = false
		}
	}

	outputLevel := splitCTs[0].Level()
	for _, ct := range splitCTs {
		if ct.Level() < outputLevel {
			outputLevel = ct.Level()
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.ReportMetric(float64(batch), "words/op")
	b.ReportMetric(float64(spec.Slots), "split_cts")
	b.ReportMetric(float64(log2t), "requested_log2t")
	b.ReportMetric(float64(params.MaxLevel()-outputLevel), "levels/op")
	b.ReportMetric(float64(runtime.GOMAXPROCS(0)), "gomaxprocs")
	b.ReportMetric(params.LogQ(), "logQ")
	b.ReportMetric(params.LogP(), "logP")
	b.ReportMetric(params.LogQP(), "logQP")
	b.ReportMetric(float64(outputLevel), "output_level")
	b.ReportMetric(float64(params.MaxLevel()), "max_level")
	if correct {
		b.ReportMetric(1, "correct")
		b.ReportMetric(0, "max_error")
	} else {
		b.ReportMetric(0, "correct")
		b.ReportMetric(maxErr, "max_error")
	}
	for i := 0; i < b.N; i++ {
		if _, err := EvalSplitFromStandard(btp, ct, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func runSplitConvertBenchmarks(b *testing.B, fn func(*testing.B, basicEncoding, int)) {
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

func BenchmarkSplitToStandard(b *testing.B) {
	runSplitConvertBenchmarks(b, benchSplitToStandard)
}

func BenchmarkSplitFromStandard(b *testing.B) {
	runSplitConvertBenchmarks(b, benchSplitFromStandard)
}
