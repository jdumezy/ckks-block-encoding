package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Pro7ech/lattigo/he"
	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/he/hefloat/bootstrapping"
	"github.com/Pro7ech/lattigo/ring"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
	"character-encoding/character-encoding/crt"
	"character-encoding/character-encoding/lut"
)

type iterResult struct {
	Iter          int     `json:"iter"`
	PrecisionBits float64 `json:"precision_bits"`
	Correct       bool    `json:"correct"`
	Bootstrapped  bool    `json:"bootstrapped"`
	Cleaned       bool    `json:"cleaned"`
	Level         int     `json:"level"`
}

type runResult struct {
	Experiment string       `json:"experiment"`
	Encoding   string       `json:"encoding"`
	Log2T      int          `json:"log2t"`
	Seed       int          `json:"seed"`
	Iterations []iterResult `json:"iterations"`
	Note       string       `json:"note,omitempty"`
}

func main() {
	out := flag.String("out", "", "")
	onlyExperiment := flag.String("experiment", "", "")
	onlyLog2T := flag.Int("log2t", -1, "")
	onlySeed := flag.Int("seed", -1, "")
	seeds := flag.Int("seeds", 10, "")
	maxIters := flag.Int("max-iters", 50, "")
	progress := flag.Bool("progress", true, "")
	logN := flag.Int("logn", 12, "")
	encoding := flag.String("encoding", "BRU", "")
	cleanGain := flag.Bool("clean-gain", false, "")
	cleanInjectBits := flag.Float64("clean-inject-bits", 12.4, "")
	flag.Parse()

	if *cleanGain {
		runCleanGain(*out, *encoding, *logN, *seeds, *cleanInjectBits, *onlyLog2T, *progress)
		return
	}

	enc := strings.ToUpper(*encoding)
	if _, ok := encodingByName(enc); !ok {
		die("unknown encoding %q (want BRU|LBRU|WH|IND)", enc)
	}

	env, err := newEnv(*logN)
	if err != nil {
		die("setup: %v", err)
	}
	logf(*progress, "LogN=%d LogQP=%.1f MaxLevel=%d post-boot residual level=%d encoding=%s", env.Params.LogN(), env.Params.LogQP(), env.Params.MaxLevel(), env.PostBootLevel, enc)

	encLower := strings.ToLower(enc)
	log2Ts := []int{2, 4, 6, 8}
	configs := []runConfig{}
	for _, log2t := range log2Ts {
		configs = append(configs, runConfig{Experiment: "unary_lut_" + encLower, Encoding: enc, Log2T: log2t})
		if enc != "IND" {
			configs = append(configs, runConfig{Experiment: "native_" + encLower, Encoding: enc, Log2T: log2t})
		}
	}
	configs = append(configs, runConfig{Experiment: "crt_lbru_256", Encoding: "LBRU", Log2T: 0})

	results := []runResult{}
	flush := func() {
		if *out == "" {
			return
		}
		payload := struct {
			LogN          int         `json:"logN"`
			PostBootLevel int         `json:"post_boot_level"`
			Results       []runResult `json:"results"`
		}{LogN: env.Params.LogN(), PostBootLevel: env.PostBootLevel, Results: results}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return
		}
		_ = os.WriteFile(*out, data, 0o644)
	}
	for _, cfg := range configs {
		if *onlyExperiment != "" && cfg.Experiment != *onlyExperiment {
			continue
		}
		if *onlyLog2T != -1 && cfg.Log2T != *onlyLog2T && cfg.Experiment != "crt_lbru_256" {
			continue
		}
		setup, err := buildExperiment(env, cfg)
		if err != nil {
			logf(*progress, "skip %s log2t=%d: %v", cfg.Experiment, cfg.Log2T, err)
			continue
		}
		for seed := 0; seed < *seeds; seed++ {
			if *onlySeed != -1 && seed != *onlySeed {
				continue
			}
			start := time.Now()
			res := runOne(env, setup, cfg, seed, *maxIters)
			logf(*progress, "%s log2t=%d seed=%d: %d iters in %s (last_correct=%v last_prec=%.1f)",
				cfg.Experiment, cfg.Log2T, seed, len(res.Iterations), time.Since(start).Round(time.Second),
				lastCorrect(res), lastPrecision(res))
			results = append(results, res)
			flush()
		}
	}

	payload := struct {
		LogN          int         `json:"logN"`
		PostBootLevel int         `json:"post_boot_level"`
		Results       []runResult `json:"results"`
	}{
		LogN:          env.Params.LogN(),
		PostBootLevel: env.PostBootLevel,
		Results:       results,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		die("marshal: %v", err)
	}
	if *out == "" {
		os.Stdout.Write(data)
		os.Stdout.Write([]byte("\n"))
		return
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		die("write: %v", err)
	}
}

func die(f string, args ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", args...)
	os.Exit(1)
}

func logf(enabled bool, f string, args ...any) {
	if !enabled {
		return
	}
	fmt.Fprintf(os.Stderr, f+"\n", args...)
}

func lastCorrect(r runResult) bool {
	if len(r.Iterations) == 0 {
		return false
	}
	return r.Iterations[len(r.Iterations)-1].Correct
}

func lastPrecision(r runResult) float64 {
	if len(r.Iterations) == 0 {
		return 0
	}
	return r.Iterations[len(r.Iterations)-1].PrecisionBits
}

type Env struct {
	Params        hefloat.Parameters
	BTPParams     bootstrapping.Parameters
	SK            *rlwe.SecretKey
	BTP           *bootstrapping.Evaluator
	CharCtx       *charctx.Context
	Boot          *bootstrapping.FullEvaluator
	PostBootLevel int
}

func newEnv(logN int) (*Env, error) {
	btpParams := buildBootstrapParams(logN)
	params := btpParams.BootstrappingParameters
	kgen := rlwe.NewKeyGenerator(params)
	sk, _ := kgen.GenKeyPairNew()
	evk, err := btpParams.GenEvaluationKeys(sk)
	if err != nil {
		return nil, fmt.Errorf("bootstrap evk: %w", err)
	}
	btp, err := bootstrapping.NewEvaluator(btpParams, evk)
	if err != nil {
		return nil, fmt.Errorf("bootstrap evaluator: %w", err)
	}
	ctx, err := newCharCtx(params, sk)
	if err != nil {
		return nil, err
	}
	boot, err := bootstrapping.NewFullEvaluator(btp, btpParams, sk, true)
	if err != nil {
		return nil, fmt.Errorf("bootstrap wrapper: %w", err)
	}
	post, err := measurePostBootLevel(params, sk, boot)
	if err != nil {
		return nil, fmt.Errorf("measure post-boot level: %w", err)
	}
	return &Env{
		Params:        params,
		BTPParams:     btpParams,
		SK:            sk,
		BTP:           btp,
		CharCtx:       ctx,
		Boot:          boot,
		PostBootLevel: post,
	}, nil
}

func buildBootstrapParams(logN int) bootstrapping.Parameters {
	residualQ := []int{55, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40}
	bootS2C := []int{39, 39, 39}
	bootMod1 := []int{60, 60, 60, 60, 60, 60, 60, 60}
	bootC2S := []int{58, 58}
	logQ := append([]int{}, residualQ...)
	logQ = append(logQ, bootS2C...)
	logQ = append(logQ, bootMod1...)
	logQ = append(logQ, bootC2S...)
	logP := []int{61, 61, 61, 61, 61}

	params, err := hefloat.NewParametersFromLiteral(hefloat.ParametersLiteral{
		LogN:            logN,
		LogQ:            logQ,
		LogP:            logP,
		LogDefaultScale: 40,
		Xs:              &ring.Ternary{H: 256},
	})
	if err != nil {
		panic(fmt.Sprintf("buildBootstrapParams: %v", err))
	}

	c2s := hefloat.DFTMatrixLiteral{
		Type:     hefloat.HomomorphicEncode,
		Format:   hefloat.RepackImagAsReal,
		LogSlots: params.LogMaxSlots(),
		LevelQ:   params.MaxLevel(),
		LevelP:   params.MaxLevelP(),
		Levels:   []int{2, 2},
	}
	mod1 := hefloat.Mod1ParametersLiteral{
		LevelQ:          params.MaxLevel() - c2s.Depth(true),
		LogScale:        60,
		Mod1Type:        hefloat.CosDiscrete,
		Mod1Degree:      30,
		Mod1Interval:    16,
		DoubleAngle:     3,
		LogMessageRatio: 8,
	}
	s2c := hefloat.DFTMatrixLiteral{
		Type:     hefloat.HomomorphicDecode,
		Format:   hefloat.RepackImagAsReal,
		LogSlots: params.LogMaxSlots(),
		LevelQ:   mod1.LevelQ - mod1.Depth(),
		LevelP:   params.MaxLevelP(),
		Levels:   []int{1, 1, 1},
	}
	return bootstrapping.Parameters{
		ResidualParameters:      params,
		BootstrappingParameters: params,
		SlotsToCoeffsParameters: s2c,
		CoeffsToSlotsParameters: c2s,
		Mod1ParametersLiteral:   mod1,
		EphemeralSecretWeight:   32,
	}
}

func newCharCtx(params hefloat.Parameters, sk *rlwe.SecretKey) (*charctx.Context, error) {
	ctx := &charctx.Context{Params: params}
	ctx.KeyGen = rlwe.NewKeyGenerator(params)
	ctx.SecretKey = sk
	rlk := ctx.KeyGen.GenRelinearizationKeyNew(sk)
	ctx.EvK = rlwe.NewMemEvaluationKeySet(rlk)
	ctx.Encoder = hefloat.NewEncoder(params)
	ctx.Encryptor = rlwe.NewEncryptor(params, sk)
	ctx.Decryptor = rlwe.NewDecryptor(params, sk)
	ctx.Evaluator = hefloat.NewEvaluator(params, ctx.EvK)
	ctx.LTEvaluator = he.NewLinearTransformationEvaluator(ctx.Evaluator)
	return ctx, nil
}

func measurePostBootLevel(params hefloat.Parameters, sk *rlwe.SecretKey, boot *bootstrapping.FullEvaluator) (int, error) {
	encoder := hefloat.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, sk)
	values := make([]complex128, params.MaxSlots())
	for i := range values {
		values[i] = complex(0.5, 0)
	}
	pt := hefloat.NewPlaintext(params, 1)
	if err := encoder.Encode(values, pt); err != nil {
		return 0, err
	}
	ct := hefloat.NewCiphertext(params, 1, pt.Level())
	if err := encryptor.Encrypt(pt, ct); err != nil {
		return 0, err
	}
	out, err := boot.Bootstrap(ct)
	if err != nil {
		return 0, err
	}
	return out.Level(), nil
}

type runConfig struct {
	Experiment string
	Encoding   string
	Log2T      int
}

func encodingByName(name string) (charenc.EncodingKind, bool) {
	switch strings.ToUpper(name) {
	case "BRU":
		return charenc.BRU, true
	case "LBRU":
		return charenc.LBRU, true
	case "WH":
		return charenc.WH, true
	case "IND":
		return charenc.IND, true
	default:
		return 0, false
	}
}

func previousPrime(n int) int {
	isPrime := func(x int) bool {
		if x < 2 {
			return false
		}
		if x%2 == 0 {
			return x == 2
		}
		for d := 3; d*d <= x; d += 2 {
			if x%d == 0 {
				return false
			}
		}
		return true
	}
	for p := n; p >= 2; p-- {
		if isPrime(p) {
			return p
		}
	}
	return 2
}

func nextPowerOfTwo(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

func makeCodec(kind charenc.EncodingKind, log2t int) (charenc.Codec, int, error) {
	t := 1 << log2t
	switch kind {
	case charenc.BRU:
		c, err := charenc.NewBRU(t, true)
		return c, t, err
	case charenc.LBRU:
		t = previousPrime(t)
		c, err := charenc.NewLBRU(t, 0, true)
		return c, t, err
	case charenc.WH:
		t = nextPowerOfTwo(t)
		c, err := charenc.NewWH(t, true)
		return c, t, err
	case charenc.IND:
		c, err := charenc.NewIND(t, true, 0)
		return c, t, err
	default:
		return nil, 0, fmt.Errorf("unknown encoding %v", kind)
	}
}

func nativeOp(kind charenc.EncodingKind, t, x, y int) int {
	switch kind {
	case charenc.BRU:
		r := (x + y) % t
		if r < 0 {
			r += t
		}
		return r
	case charenc.LBRU:
		return (x * y) % t
	case charenc.WH:
		return x ^ y
	default:
		return x
	}
}

func nativeY(kind charenc.EncodingKind, t int) int {
	switch kind {
	case charenc.LBRU:
		for y := 2; y < t; y++ {
			if y%t != 1 {
				return y
			}
		}
		return 1
	default:
		return 1
	}
}

type experimentSetup struct {
	cfg runConfig

	encKind       charenc.EncodingKind
	blockSpec     charenc.BlockSpec
	blockT        int
	blockCodec    charenc.Codec
	blockBatch    int
	lutEval       *lut.Evaluator
	lutSeedBase   int64
	nativeUsed    int
	nativeYInt    int
	cleanToIND    *blt.RawCompiled
	cleanFromIND  *blt.RawCompiled
	cleanInLevel  int
	cleanOutLevel int

	crtSpec     crt.PackedSpec
	crtBaseSpec crt.Spec
	crtEval     *crt.Evaluator
}

func buildExperiment(env *Env, cfg runConfig) (*experimentSetup, error) {
	setup := &experimentSetup{cfg: cfg}
	isUnary := strings.HasPrefix(cfg.Experiment, "unary_lut_")
	isNative := strings.HasPrefix(cfg.Experiment, "native_")
	switch {
	case isUnary || isNative:
		kind, ok := encodingByName(cfg.Encoding)
		if !ok {
			return nil, fmt.Errorf("unknown encoding %q", cfg.Encoding)
		}
		setup.encKind = kind
		codec, t, err := makeCodec(kind, cfg.Log2T)
		if err != nil {
			return nil, fmt.Errorf("codec %s log2t=%d: %w", cfg.Encoding, cfg.Log2T, err)
		}
		setup.blockCodec = codec
		setup.blockSpec = codec.Spec()
		setup.blockT = t
		batch := env.Params.MaxSlots() / setup.blockSpec.Slots
		if batch < 1 {
			batch = 1
		}
		setup.blockBatch = batch
		setup.nativeUsed = batch * setup.blockSpec.Slots
		setup.nativeYInt = nativeY(kind, t)
		setup.lutEval = lut.NewEvaluatorWithWorkerCapacity(env.CharCtx, 1)

		if isUnary {
			setup.lutSeedBase = int64(cfg.Log2T)*9001 + 1234567 + int64(kind)*7777
		}
		if cfg.Log2T < 8 {
			setup.cleanInLevel = env.PostBootLevel
			setup.cleanOutLevel = setup.cleanInLevel - 4
			toIND, err := compileToIND(env.CharCtx, setup.blockSpec, codec, t, batch, setup.cleanInLevel)
			if err != nil {
				return nil, fmt.Errorf("compile toIND: %w", err)
			}
			fromIND, err := compileFromIND(env.CharCtx, setup.blockSpec, codec, t, batch, setup.cleanInLevel-3)
			if err != nil {
				return nil, fmt.Errorf("compile fromIND: %w", err)
			}
			setup.cleanToIND = toIND
			setup.cleanFromIND = fromIND
		}

	case cfg.Experiment == "crt_lbru_256":
		base, err := crt.NewSpecForBits(256, charenc.LBRU, 0)
		if err != nil {
			return nil, fmt.Errorf("CRT 256 spec: %w", err)
		}
		spec, err := crt.NewPackedSpec(base, env.Params.MaxSlots())
		if err != nil {
			return setup, fmt.Errorf("packed CRT spec (LogN=%d): %w", env.Params.LogN(), err)
		}
		setup.crtBaseSpec = base
		setup.crtSpec = spec
		setup.crtEval = crt.NewEvaluator(env.CharCtx, 1)
	default:
		return nil, fmt.Errorf("unknown experiment %q", cfg.Experiment)
	}
	return setup, nil
}

func maxIntMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func randomTable(t int, seed int64) []int {
	rng := rand.New(rand.NewSource(seed))
	out := make([]int, t)
	for i := range out {
		out[i] = i
	}
	for i := t - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func compileToIND(ctx *charctx.Context, srcSpec charenc.BlockSpec, srcCodec charenc.Codec, t, batch, inputLevel int) (*blt.RawCompiled, error) {
	indCodec, err := charenc.NewIND(t, true, 0)
	if err != nil {
		return nil, err
	}
	tr, err := blt.CompileUnary(srcCodec, indCodec, func(x int) int { return x })
	if err != nil {
		return nil, err
	}
	return compileTiledBLT(ctx, tr, srcSpec, indCodec.Spec(), batch, inputLevel)
}

func compileFromIND(ctx *charctx.Context, dstSpec charenc.BlockSpec, dstCodec charenc.Codec, t, batch, inputLevel int) (*blt.RawCompiled, error) {
	indCodec, err := charenc.NewIND(t, true, 0)
	if err != nil {
		return nil, err
	}
	tr, err := blt.CompileUnary(indCodec, dstCodec, func(x int) int { return x })
	if err != nil {
		return nil, err
	}
	return compileTiledBLT(ctx, tr, indCodec.Spec(), dstSpec, batch, inputLevel)
}

func compileTiledBLT(ctx *charctx.Context, tr blt.Transform, inSpec, outSpec charenc.BlockSpec, batch, inputLevel int) (*blt.RawCompiled, error) {
	slots := ctx.Params.MaxSlots()
	diagonals := he.Diagonals[complex128]{}
	var fullBias []complex128
	if tr.Bias != nil {
		fullBias = make([]complex128, batch*outSpec.Slots)
	}
	diagonalOnly := true
	for block := 0; block < batch; block++ {
		inBase := block * inSpec.Slots
		outBase := block * outSpec.Slots
		if tr.Bias != nil {
			copy(fullBias[outBase:outBase+outSpec.Slots], tr.Bias)
		}
		for r := 0; r < outSpec.Slots; r++ {
			for c, v := range tr.Matrix[r] {
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
	raw, err := blt.CompileDiagonalsWithOptions(diagonals, diagonalOnly, fullBias, batch*outSpec.Slots, batch*inSpec.Slots, ctx, inputLevel, blt.CompileOptions{})
	if err != nil {
		return nil, err
	}
	ctx.EnsureGaloisKeys(raw.GaloisEls)
	return raw, nil
}

func cleanCT(env *Env, setup *experimentSetup, ct *rlwe.Ciphertext) (*rlwe.Ciphertext, bool, error) {
	if ct.Level() != setup.cleanInLevel {
		newCT, err := env.Boot.Bootstrap(ct)
		if err != nil {
			return nil, false, fmt.Errorf("clean bootstrap: %w", err)
		}
		ct = newCT
	}
	ind, err := setup.lutEval.BLT.ApplyRaw(ct, setup.cleanToIND)
	if err != nil {
		return nil, false, fmt.Errorf("clean toIND: %w", err)
	}
	cleaned, err := evalINDClean(env, ind)
	if err != nil {
		return nil, false, fmt.Errorf("clean INDClean: %w", err)
	}
	out, err := setup.lutEval.BLT.ApplyRaw(cleaned, setup.cleanFromIND)
	if err != nil {
		return nil, false, fmt.Errorf("clean fromIND: %w", err)
	}
	return out, true, nil
}

func evalINDClean(env *Env, ind *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	ev := env.CharCtx.Evaluator
	x2 := hefloat.NewCiphertext(env.Params, 1, ind.Level())
	if err := ev.MulRelin(ind, ind, x2); err != nil {
		return nil, err
	}
	if err := ev.Rescale(x2, x2); err != nil {
		return nil, err
	}
	indAtX2 := ind
	if indAtX2.Level() > x2.Level() {
		indAtX2 = ev.DropLevelNew(indAtX2, indAtX2.Level()-x2.Level())
	}
	x3 := hefloat.NewCiphertext(env.Params, 1, x2.Level())
	if err := ev.MulRelin(x2, indAtX2, x3); err != nil {
		return nil, err
	}
	if err := ev.Rescale(x3, x3); err != nil {
		return nil, err
	}
	x2AtX3 := x2
	if x2AtX3.Level() > x3.Level() {
		x2AtX3 = ev.DropLevelNew(x2AtX3, x2AtX3.Level()-x3.Level())
	}
	term2 := hefloat.NewCiphertext(env.Params, 1, x3.Level())
	if err := ev.Mul(x2AtX3, 3, term2); err != nil {
		return nil, err
	}
	term3 := hefloat.NewCiphertext(env.Params, 1, x3.Level())
	if err := ev.Mul(x3, 2, term3); err != nil {
		return nil, err
	}
	out := hefloat.NewCiphertext(env.Params, 1, x3.Level())
	if err := ev.Sub(term2, term3, out); err != nil {
		return nil, err
	}
	return out, nil
}

func compileTiledUnaryLUT(ctx *charctx.Context, spec charenc.BlockSpec, codec charenc.Codec, t, batch int, table []int, inputLevel int) (*blt.RawCompiled, error) {
	fn := func(x int) int {
		idx := x % t
		if idx < 0 {
			idx += t
		}
		return table[idx]
	}
	tr, err := blt.CompileUnary(codec, codec, fn)
	if err != nil {
		return nil, err
	}
	return compileTiledBLT(ctx, tr, spec, spec, batch, inputLevel)
}

func runOne(env *Env, setup *experimentSetup, cfg runConfig, seed, maxIters int) runResult {
	res := runResult{Experiment: cfg.Experiment, Encoding: cfg.Encoding, Log2T: cfg.Log2T, Seed: seed}
	switch {
	case strings.HasPrefix(cfg.Experiment, "unary_lut_"):
		runUnaryLUT(env, setup, seed, maxIters, &res)
	case strings.HasPrefix(cfg.Experiment, "native_") && cfg.Experiment != "native_":
		runNative(env, setup, seed, maxIters, &res)
	case cfg.Experiment == "crt_lbru_256":
		runCRTLBRU(env, setup, seed, maxIters, &res)
	}
	return res
}

func runUnaryLUT(env *Env, setup *experimentSetup, seed, maxIters int, res *runResult) {
	rng := rand.New(rand.NewSource(int64(seed) + 7919))
	xs := make([]int, setup.blockBatch)
	for i := range xs {
		xs[i] = rng.Intn(setup.blockT)
	}
	expected := append([]int(nil), xs...)
	plain := encodeBatch(setup.blockCodec, setup.blockSpec, xs, env.Params.MaxSlots(), setup.blockBatch)
	ct, err := encryptVector(env.CharCtx, plain, env.PostBootLevel)
	if err != nil {
		res.Note = fmt.Sprintf("encrypt: %v", err)
		return
	}
	for iter := 0; iter < maxIters; iter++ {
		bootstrapped := false
		if ct.Level() < 1 {
			newCT, err := env.Boot.Bootstrap(ct)
			if err != nil {
				res.Note = fmt.Sprintf("iter %d bootstrap: %v", iter, err)
				return
			}
			ct = newCT
			bootstrapped = true
		}
		iterSeed := setup.lutSeedBase + int64(seed)*131 + int64(iter)*977
		table := randomTable(setup.blockT, iterSeed)
		lutInst, err := compileTiledUnaryLUT(env.CharCtx, setup.blockSpec, setup.blockCodec, setup.blockT, setup.blockBatch, table, ct.Level())
		if err != nil {
			res.Note = fmt.Sprintf("iter %d compile LUT: %v", iter, err)
			return
		}
		newCT, err := setup.lutEval.BLT.ApplyRaw(ct, lutInst)
		if err != nil {
			res.Note = fmt.Sprintf("iter %d ApplyRaw: %v", iter, err)
			return
		}
		ct = newCT
		for i := range expected {
			idx := expected[i] % setup.blockT
			if idx < 0 {
				idx += setup.blockT
			}
			expected[i] = table[idx]
		}
		processIterResult(env, setup, &ct, expected, iter, bootstrapped, res)
		if !res.Iterations[len(res.Iterations)-1].Correct {
			return
		}
	}
}

func runNative(env *Env, setup *experimentSetup, seed, maxIters int, res *runResult) {
	rng := rand.New(rand.NewSource(int64(seed) + 4129))
	xs := make([]int, setup.blockBatch)
	for i := range xs {
		xs[i] = rng.Intn(setup.blockT)
	}
	expected := append([]int(nil), xs...)
	plain := encodeBatch(setup.blockCodec, setup.blockSpec, xs, env.Params.MaxSlots(), setup.blockBatch)
	ct, err := encryptVector(env.CharCtx, plain, env.PostBootLevel)
	if err != nil {
		res.Note = fmt.Sprintf("encrypt: %v", err)
		return
	}
	yInts := make([]int, setup.blockBatch)
	for i := range yInts {
		yInts[i] = setup.nativeYInt
	}
	yPlain := encodeBatch(setup.blockCodec, setup.blockSpec, yInts, env.Params.MaxSlots(), setup.blockBatch)

	w := env.CharCtx.Evaluator
	for iter := 0; iter < maxIters; iter++ {
		bootstrapped := false
		if ct.Level() < 1 {
			newCT, err := env.Boot.Bootstrap(ct)
			if err != nil {
				res.Note = fmt.Sprintf("iter %d bootstrap: %v", iter, err)
				return
			}
			ct = newCT
			bootstrapped = true
		}
		yCT, err := encryptVector(env.CharCtx, yPlain, ct.Level())
		if err != nil {
			res.Note = fmt.Sprintf("iter %d encrypt y: %v", iter, err)
			return
		}
		out := hefloat.NewCiphertext(env.Params, 1, ct.Level())
		if err := w.MulRelin(ct, yCT, out); err != nil {
			res.Note = fmt.Sprintf("iter %d mul: %v", iter, err)
			return
		}
		if err := w.Rescale(out, out); err != nil {
			res.Note = fmt.Sprintf("iter %d rescale: %v", iter, err)
			return
		}
		ct = out
		for i := range expected {
			expected[i] = nativeOp(setup.encKind, setup.blockT, expected[i], setup.nativeYInt)
		}
		processIterResult(env, setup, &ct, expected, iter, bootstrapped, res)
		if !res.Iterations[len(res.Iterations)-1].Correct {
			return
		}
	}
}

func processIterResult(env *Env, setup *experimentSetup, ctPtr **rlwe.Ciphertext, expected []int, iter int, bootstrapped bool, res *runResult) {
	ct := *ctPtr
	decoded, err := decryptVector(env.CharCtx, ct)
	if err != nil {
		res.Note = fmt.Sprintf("iter %d decrypt: %v", iter, err)
		res.Iterations = append(res.Iterations, iterResult{Iter: iter, PrecisionBits: 0, Correct: false, Bootstrapped: bootstrapped, Level: ct.Level()})
		return
	}
	expectedPlain := encodeBatch(setup.blockCodec, setup.blockSpec, expected, env.Params.MaxSlots(), setup.blockBatch)
	prec, _ := precisionBits(maxAbsError(decoded, expectedPlain, setup.nativeUsed))
	correct := decodeAndCheckBlocks(setup.blockCodec, decoded, expected, setup.blockSpec.Slots)
	cleaned := false
	if !correct && setup.cleanToIND != nil {
		cleanedCT, ok, err := cleanCT(env, setup, ct)
		if err == nil && ok {
			cleanedDecoded, derr := decryptVector(env.CharCtx, cleanedCT)
			if derr == nil {
				cleanedPrec, _ := precisionBits(maxAbsError(cleanedDecoded, expectedPlain, setup.nativeUsed))
				cleanedCorrect := decodeAndCheckBlocks(setup.blockCodec, cleanedDecoded, expected, setup.blockSpec.Slots)
				if cleanedCorrect {
					ct = cleanedCT
					prec = cleanedPrec
					correct = true
					cleaned = true
					*ctPtr = cleanedCT
				}
			}
		}
	}
	res.Iterations = append(res.Iterations, iterResult{Iter: iter, PrecisionBits: prec, Correct: correct, Bootstrapped: bootstrapped, Cleaned: cleaned, Level: ct.Level()})
}

func runCRTLBRU(env *Env, setup *experimentSetup, seed, maxIters int, res *runResult) {
	rng := rand.New(rand.NewSource(int64(seed) + 1009))
	xWord := new(big.Int).Lsh(big.NewInt(1), uint(setup.crtBaseSpec.Bits-1))
	xWord.Add(xWord, big.NewInt(int64(rng.Intn(1<<20)+1)))
	yResidues := make([]int, setup.crtBaseSpec.Channels())
	for i, p := range setup.crtBaseSpec.Primes {
		if p == 2 {
			yResidues[i] = 1
		} else {
			yResidues[i] = 2 % p
		}
	}
	xResidues := make([]int, setup.crtBaseSpec.Channels())
	for i, p := range setup.crtBaseSpec.Primes {
		xResidues[i] = int(new(big.Int).Mod(xWord, big.NewInt(int64(p))).Int64())
	}

	xs := make([]*big.Int, setup.crtSpec.BatchSize)
	for i := range xs {
		xs[i] = xWord
	}
	plain, err := crt.EncodePackedBatch(setup.crtSpec, xs)
	if err != nil {
		res.Note = fmt.Sprintf("encode x: %v", err)
		return
	}
	xPCT, err := crt.EncryptPacked(env.CharCtx, plain, env.PostBootLevel)
	if err != nil {
		res.Note = fmt.Sprintf("encrypt x: %v", err)
		return
	}

	yBig := crtFromResidues(setup.crtBaseSpec.Primes, yResidues)
	ys := make([]*big.Int, setup.crtSpec.BatchSize)
	for i := range ys {
		ys[i] = yBig
	}
	yPlain, err := crt.EncodePackedBatch(setup.crtSpec, ys)
	if err != nil {
		res.Note = fmt.Sprintf("encode y: %v", err)
		return
	}

	expectedResidues := append([]int(nil), xResidues...)

	for iter := 0; iter < maxIters; iter++ {
		bootstrapped := false
		levelMin := minCTLevel(xPCT)
		if levelMin < 1 {
			refreshed, err := bootstrapPacked(env, xPCT)
			if err != nil {
				res.Note = fmt.Sprintf("iter %d bootstrap: %v", iter, err)
				return
			}
			xPCT = refreshed
			bootstrapped = true
		}
		yPCT, err := crt.EncryptPacked(env.CharCtx, yPlain, minCTLevel(xPCT))
		if err != nil {
			res.Note = fmt.Sprintf("iter %d encrypt y: %v", iter, err)
			return
		}
		outPCT, err := setup.crtEval.EvalPackedNativeProduct(xPCT, yPCT, false)
		if err != nil {
			res.Note = fmt.Sprintf("iter %d mul: %v", iter, err)
			return
		}
		xPCT = outPCT
		for i, p := range setup.crtBaseSpec.Primes {
			expectedResidues[i] = (expectedResidues[i] * yResidues[i]) % p
		}
		prec, correct := evaluateCRTPrecision(env, setup, xPCT, expectedResidues)
		res.Iterations = append(res.Iterations, iterResult{Iter: iter, PrecisionBits: prec, Correct: correct, Bootstrapped: bootstrapped, Level: minCTLevel(xPCT)})
		if !correct {
			return
		}
	}
}

func bootstrapPacked(env *Env, x crt.PackedCiphertext) (crt.PackedCiphertext, error) {
	out := crt.PackedCiphertext{Spec: x.Spec, CTs: make([]*rlwe.Ciphertext, len(x.CTs))}
	for i, ct := range x.CTs {
		newCT, err := env.Boot.Bootstrap(ct)
		if err != nil {
			return crt.PackedCiphertext{}, fmt.Errorf("bootstrap ciphertext %d: %w", i, err)
		}
		out.CTs[i] = newCT
	}
	return out, nil
}

func evaluateCRTPrecision(env *Env, setup *experimentSetup, ct crt.PackedCiphertext, expectedResidues []int) (float64, bool) {
	pt, err := crt.DecryptPacked(env.CharCtx, ct)
	if err != nil {
		return 0, false
	}
	all, err := crt.DecodePackedBatchResidues(pt)
	if err != nil {
		return 0, false
	}
	correct := true
	for _, residues := range all {
		for ch, want := range expectedResidues {
			expected := want % setup.crtBaseSpec.Primes[ch]
			if expected < 0 {
				expected += setup.crtBaseSpec.Primes[ch]
			}
			if residues[ch] != expected {
				correct = false
				break
			}
		}
		if !correct {
			break
		}
	}

	expectedSlots := make([][]complex128, ct.Spec.Ciphertexts())
	for i := range expectedSlots {
		expectedSlots[i] = make([]complex128, ct.Spec.MaxSlots)
	}
	for w := 0; w < ct.Spec.BatchSize; w++ {
		for chIdx, ch := range ct.Spec.Channels {
			codec, err := codecForPrimeCRT(setup.crtBaseSpec, chIdx)
			if err != nil {
				return 0, false
			}
			vals := codec.EncodeValue(expectedResidues[chIdx])
			base := w*ct.Spec.WordStride + ch.Offset
			copy(expectedSlots[ch.Ciphertext][base:base+ch.Slots], vals)
		}
	}
	maxErr := 0.0
	for i, decodedSlots := range pt.Values {
		for j := 0; j < ct.Spec.TotalUsed; j++ {
			if e := cmplx.Abs(decodedSlots[j] - expectedSlots[i][j]); e > maxErr {
				maxErr = e
			}
		}
	}
	prec, _ := precisionBits(maxErr)
	return prec, correct
}

func codecForPrimeCRT(s crt.Spec, channelIdx int) (charenc.Codec, error) {
	p := s.Primes[channelIdx]
	switch s.Kind {
	case charenc.BRU:
		return charenc.NewBRU(p, s.Reduced)
	case charenc.LBRU:
		return charenc.NewLBRU(p, 0, s.Reduced)
	default:
		return nil, fmt.Errorf("codecForPrimeCRT: unsupported kind %v", s.Kind)
	}
}

func crtFromResidues(primes []int, residues []int) *big.Int {
	M := big.NewInt(1)
	for _, p := range primes {
		M.Mul(M, big.NewInt(int64(p)))
	}
	out := big.NewInt(0)
	for i, p := range primes {
		Pi := big.NewInt(int64(p))
		Mi := new(big.Int).Quo(M, Pi)
		MiInv := new(big.Int).ModInverse(Mi, Pi)
		if MiInv == nil {
			continue
		}
		term := new(big.Int).Mul(big.NewInt(int64(residues[i])), Mi)
		term.Mul(term, MiInv)
		out.Add(out, term)
	}
	out.Mod(out, M)
	return out
}

func minCTLevel(x crt.PackedCiphertext) int {
	if len(x.CTs) == 0 {
		return 0
	}
	m := x.CTs[0].Level()
	for _, c := range x.CTs[1:] {
		if c.Level() < m {
			m = c.Level()
		}
	}
	return m
}

func encodeBatch(codec charenc.Codec, spec charenc.BlockSpec, values []int, slots, batch int) []complex128 {
	out := make([]complex128, slots)
	for i, v := range values {
		if i >= batch {
			break
		}
		copy(out[i*spec.Slots:i*spec.Slots+spec.Slots], codec.EncodeValue(v))
	}
	return out
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
		if e := cmplx.Abs(got[i] - want[i]); e > maxErr {
			maxErr = e
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

func decodeAndCheckBlocks(codec charenc.Codec, decoded []complex128, expected []int, blockSlots int) bool {
	for i, want := range expected {
		got, err := codec.DecodeValue(decoded[i*blockSlots : i*blockSlots+blockSlots])
		if err != nil {
			return false
		}
		if got != want {
			return false
		}
	}
	return true
}

type cleanGainResult struct {
	Experiment      string  `json:"experiment"`
	Encoding        string  `json:"encoding"`
	Log2T           int     `json:"log2t"`
	Seed            int     `json:"seed"`
	InjectBits      float64 `json:"inject_bits"`
	PrecisionBefore float64 `json:"precision_before"`
	PrecisionAfter  float64 `json:"precision_after"`
	GainBits        float64 `json:"gain_bits"`
	CorrectBefore   bool    `json:"correct_before"`
	CorrectAfter    bool    `json:"correct_after"`
	LevelBefore     int     `json:"level_before"`
	LevelAfter      int     `json:"level_after"`
	Note            string  `json:"note,omitempty"`
}

func runCleanGain(outPath, encoding string, logN, seeds int, injectBits float64, onlyLog2T int, progress bool) {
	enc := strings.ToUpper(encoding)
	kind, ok := encodingByName(enc)
	if !ok {
		die("unknown encoding %q", enc)
	}
	env, err := newEnv(logN)
	if err != nil {
		die("setup: %v", err)
	}
	logf(progress, "LogN=%d LogQP=%.1f MaxLevel=%d post-boot residual level=%d encoding=%s clean-inject-bits=%.2f", env.Params.LogN(), env.Params.LogQP(), env.Params.MaxLevel(), env.PostBootLevel, enc, injectBits)

	log2Ts := []int{2, 4, 6, 8}
	results := []cleanGainResult{}
	flush := func() {
		if outPath == "" {
			return
		}
		data, err := json.MarshalIndent(struct {
			LogN          int               `json:"logN"`
			PostBootLevel int               `json:"post_boot_level"`
			Results       []cleanGainResult `json:"results"`
		}{LogN: env.Params.LogN(), PostBootLevel: env.PostBootLevel, Results: results}, "", "  ")
		if err != nil {
			return
		}
		_ = os.WriteFile(outPath, data, 0o644)
	}

	for _, log2t := range log2Ts {
		if onlyLog2T != -1 && onlyLog2T != log2t {
			continue
		}
		cfg := runConfig{Experiment: "clean_gain_" + strings.ToLower(enc), Encoding: enc, Log2T: log2t}
		setup, err := buildCleanGainSetup(env, cfg, kind, log2t)
		if err != nil {
			logf(progress, "skip log2t=%d: %v", log2t, err)
			continue
		}
		for seed := 0; seed < seeds; seed++ {
			start := time.Now()
			res := measureCleanGain(env, setup, cfg, seed, injectBits)
			logf(progress, "%s log2t=%d seed=%d in %s: before=%.2f after=%.2f gain=%+.2f bits (correct_after=%v)",
				cfg.Experiment, log2t, seed, time.Since(start).Round(time.Millisecond),
				res.PrecisionBefore, res.PrecisionAfter, res.GainBits, res.CorrectAfter)
			results = append(results, res)
			flush()
		}
	}
}

func buildCleanGainSetup(env *Env, cfg runConfig, kind charenc.EncodingKind, log2t int) (*experimentSetup, error) {
	setup := &experimentSetup{cfg: cfg, encKind: kind}
	codec, t, err := makeCodec(kind, log2t)
	if err != nil {
		return nil, err
	}
	setup.blockCodec = codec
	setup.blockSpec = codec.Spec()
	setup.blockT = t
	batch := env.Params.MaxSlots() / setup.blockSpec.Slots
	if batch < 1 {
		batch = 1
	}
	setup.blockBatch = batch
	setup.nativeUsed = batch * setup.blockSpec.Slots
	setup.lutEval = lut.NewEvaluatorWithWorkerCapacity(env.CharCtx, 1)
	setup.cleanInLevel = env.PostBootLevel
	setup.cleanOutLevel = setup.cleanInLevel - 4
	toIND, err := compileToIND(env.CharCtx, setup.blockSpec, codec, t, batch, setup.cleanInLevel)
	if err != nil {
		return nil, fmt.Errorf("compile toIND: %w", err)
	}
	fromIND, err := compileFromIND(env.CharCtx, setup.blockSpec, codec, t, batch, setup.cleanInLevel-3)
	if err != nil {
		return nil, fmt.Errorf("compile fromIND: %w", err)
	}
	setup.cleanToIND = toIND
	setup.cleanFromIND = fromIND
	return setup, nil
}

func measureCleanGain(env *Env, setup *experimentSetup, cfg runConfig, seed int, injectBits float64) cleanGainResult {
	res := cleanGainResult{Experiment: cfg.Experiment, Encoding: cfg.Encoding, Log2T: cfg.Log2T, Seed: seed, InjectBits: injectBits}
	rng := rand.New(rand.NewSource(int64(seed) + 13337))
	xs := make([]int, setup.blockBatch)
	for i := range xs {
		xs[i] = rng.Intn(setup.blockT)
	}
	plain := encodeBatch(setup.blockCodec, setup.blockSpec, xs, env.Params.MaxSlots(), setup.blockBatch)
	expectedPlain := append([]complex128(nil), plain...)
	noiseMag := math.Pow(2, -injectBits)
	noisy := make([]complex128, len(plain))
	for i := 0; i < setup.nativeUsed; i++ {
		theta := 2 * math.Pi * rng.Float64()
		r := noiseMag * rng.Float64()
		noisy[i] = plain[i] + complex(r*math.Cos(theta), r*math.Sin(theta))
	}
	ct, err := encryptVector(env.CharCtx, noisy, setup.cleanInLevel)
	if err != nil {
		res.Note = fmt.Sprintf("encrypt: %v", err)
		return res
	}
	res.LevelBefore = ct.Level()
	decBefore, err := decryptVector(env.CharCtx, ct)
	if err != nil {
		res.Note = fmt.Sprintf("decrypt before: %v", err)
		return res
	}
	precBefore, _ := precisionBits(maxAbsError(decBefore, expectedPlain, setup.nativeUsed))
	res.PrecisionBefore = precBefore
	res.CorrectBefore = decodeAndCheckBlocks(setup.blockCodec, decBefore, xs, setup.blockSpec.Slots)
	cleaned, _, err := cleanCT(env, setup, ct)
	if err != nil {
		res.Note = fmt.Sprintf("clean: %v", err)
		return res
	}
	res.LevelAfter = cleaned.Level()
	decAfter, err := decryptVector(env.CharCtx, cleaned)
	if err != nil {
		res.Note = fmt.Sprintf("decrypt after: %v", err)
		return res
	}
	precAfter, _ := precisionBits(maxAbsError(decAfter, expectedPlain, setup.nativeUsed))
	res.PrecisionAfter = precAfter
	res.CorrectAfter = decodeAndCheckBlocks(setup.blockCodec, decAfter, xs, setup.blockSpec.Slots)
	res.GainBits = precAfter - precBefore
	return res
}
