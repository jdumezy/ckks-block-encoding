package lut

import (
	"fmt"
	"math"
	"math/cmplx"
	"runtime"
	"sync"
	"testing"

	"github.com/Pro7ech/lattigo/he/hefloat"
	"github.com/Pro7ech/lattigo/rlwe"

	"character-encoding/character-encoding/blt"
	"character-encoding/character-encoding/charctx"
	"character-encoding/character-encoding/charenc"
)

// Split-layout benchmarks. In the split layout, each BRU coordinate of a block
// lives in its own ciphertext: J = spec.Slots ciphertexts per block, with slot
// w of CT[c] holding the c-th coordinate of word w's block. The split layout
// has no SIMD-style block packing within a ciphertext, so every ciphertext can
// be filled to MaxSlots words. Operations decompose into per-coordinate work
// that is parallelised over the J output coordinates (and, for bivariate LUTs,
// over the J x J basis multiplications).

// encodeSplitInputs returns J = spec.Slots plaintext vectors, one per block
// coordinate. Slot w of result[c] is the c-th coordinate of codec.EncodeValue(vals[w]).
func encodeSplitInputs(spec charenc.BlockSpec, vals []int, slots int) ([][]complex128, error) {
	codec, err := charenc.NewCodec(spec)
	if err != nil {
		return nil, err
	}
	J := spec.Slots
	out := make([][]complex128, J)
	for c := 0; c < J; c++ {
		out[c] = make([]complex128, slots)
	}
	for w, v := range vals {
		if w >= slots {
			break
		}
		encoded := codec.EncodeValue(v)
		for c := 0; c < J; c++ {
			out[c][w] = encoded[c]
		}
	}
	return out, nil
}

func encryptSplitInputs(ctx *charctx.Context, spec charenc.BlockSpec, vals []int, level int) ([]*rlwe.Ciphertext, error) {
	plain, err := encodeSplitInputs(spec, vals, ctx.Params.MaxSlots())
	if err != nil {
		return nil, err
	}
	cts := make([]*rlwe.Ciphertext, len(plain))
	for c, vec := range plain {
		ct, err := encryptVector(ctx, vec, level)
		if err != nil {
			return nil, fmt.Errorf("encrypt coord %d: %w", c, err)
		}
		cts[c] = ct
	}
	return cts, nil
}

func decryptSplitInputs(ctx *charctx.Context, cts []*rlwe.Ciphertext) ([][]complex128, error) {
	out := make([][]complex128, len(cts))
	for c, ct := range cts {
		vec, err := decryptVector(ctx, ct)
		if err != nil {
			return nil, fmt.Errorf("decrypt coord %d: %w", c, err)
		}
		out[c] = vec
	}
	return out, nil
}

// splitBatch returns the natural batch size for the split layout: one word
// per slot, filling the ciphertext.
func splitBatch(_ *testing.B) int {
	return 1 << (basicBenchLogN - 1)
}

func checkSplitOutput(b *testing.B, spec charenc.BlockSpec, decoded [][]complex128, want []int) bool {
	b.Helper()
	codec, err := charenc.NewCodec(spec)
	if err != nil {
		b.Fatal(err)
	}
	J := spec.Slots
	coords := make([]complex128, J)
	for w, v := range want {
		for c := 0; c < J; c++ {
			coords[c] = decoded[c][w]
		}
		got, err := codec.DecodeValue(coords)
		if err != nil {
			b.Fatal(err)
		}
		if got != positiveMod(v, spec.Alphabet.Modulus) {
			return false
		}
	}
	return true
}

// maxAbsSplitError returns the worst per-slot error across all coord CTs for
// the first `batch` words.
func maxAbsSplitError(got, want [][]complex128, batch int) float64 {
	maxErr := 0.0
	for c := range got {
		for w := 0; w < batch; w++ {
			err := cmplx.Abs(got[c][w] - want[c][w])
			if err > maxErr {
				maxErr = err
			}
		}
	}
	return maxErr
}

func reportSplitMetrics(b *testing.B, ctx *charctx.Context, spec charenc.BlockSpec, log2t, actualT, batch, levels int, inPrec, outPrec float64, inSat, outSat, correct bool) {
	reportBasicMetrics(b, ctx, spec, log2t, actualT, batch, levels, inPrec, inSat, outPrec, outSat, correct)
	b.ReportMetric(float64(spec.Slots), "split_cts")
}

func splitWorkerCapacity() int {
	c := runtime.GOMAXPROCS(0)
	if c < 1 {
		return 1
	}
	return c
}

// splitLTPlan is the compiled form of an affine block linear transform
// out_r = sum_c matrix[r][c] * basis[c] + bias[r], with every coefficient
// held as a slot-constant scalar at the chosen operation level. Non-zero
// coefficients are nudged onto Lattigo's scaled scalar path so every per-row
// reduction is followed by a single Rescale, regardless of whether individual
// coefficients are integer-valued.
type splitLTPlan struct {
	coeff  [][]complex128  // [R][N]; zero entries denote zero coefficients
	bias   []complex128    // length R; added as a CT-scalar at the pre-rescale scale
	zeroPT *rlwe.Plaintext // zero plaintext at the same level/scale (for all-zero rows)
	level  int             // ciphertext level at which Mul(basis, coeff) is applied
}

// compileSplitLTPlan records every non-zero coefficient of an (R x N) matrix
// as a CKKS scalar at the chosen operation level. Exact Gaussian integers get
// a one-ulp real nudge so Lattigo uses the same scaled multiplication path as
// non-integer scalars.
func compileSplitLTPlan(ctx *charctx.Context, matrix [][]complex128, bias []complex128, level int) (*splitLTPlan, error) {
	R := len(matrix)
	if R == 0 {
		return nil, fmt.Errorf("compileSplitLTPlan: matrix has no rows")
	}
	N := len(matrix[0])
	if len(bias) != R {
		return nil, fmt.Errorf("compileSplitLTPlan: bias length %d != %d rows", len(bias), R)
	}
	slots := ctx.Params.MaxSlots()
	defaultScale := ctx.Params.DefaultScale()
	ptScale := ctx.Params.GetScalingFactor(defaultScale, defaultScale, level)
	coeff := make([][]complex128, R)
	for r := 0; r < R; r++ {
		if len(matrix[r]) != N {
			return nil, fmt.Errorf("compileSplitLTPlan: row %d has %d cols, expected %d", r, len(matrix[r]), N)
		}
		coeff[r] = make([]complex128, N)
		for c := 0; c < N; c++ {
			v := matrix[r][c]
			if v == 0 {
				continue
			}
			coeff[r][c] = splitScaledScalar(v)
		}
	}
	zeroPT := hefloat.NewPlaintext(ctx.Params, level)
	zeroPT.Scale = ptScale
	if err := ctx.Encoder.Encode(make([]complex128, slots), zeroPT); err != nil {
		return nil, fmt.Errorf("compileSplitLTPlan: encode zero plaintext: %w", err)
	}
	return &splitLTPlan{coeff: coeff, bias: bias, zeroPT: zeroPT, level: level}, nil
}

func splitScaledScalar(v complex128) complex128 {
	if real(v) == math.Trunc(real(v)) && imag(v) == math.Trunc(imag(v)) {
		return complex(math.Nextafter(real(v), math.Inf(1)), imag(v))
	}
	return v
}

// evalSplitLT applies a compiled LT to a uniform-level basis, in parallel
// across chunks of non-zero row terms. Caller guarantees that every non-nil
// basis CT is at exactly plan.level.
func evalSplitLT(ev *Evaluator, basis []*rlwe.Ciphertext, plan *splitLTPlan) ([]*rlwe.Ciphertext, error) {
	R := len(plan.coeff)
	rowCols := make([][]int, R)
	totalTerms := 0
	for r, row := range plan.coeff {
		for c, coeff := range row {
			if c >= len(basis) {
				return nil, fmt.Errorf("split LT row %d col %d: basis has length %d", r, c, len(basis))
			}
			if coeff == 0 || basis[c] == nil {
				continue
			}
			rowCols[r] = append(rowCols[r], c)
			totalTerms++
		}
	}

	capacity := ev.BLT.Capacity()
	if totalTerms == 0 || capacity <= 1 || totalTerms < capacity*4 {
		return evalSplitLTRows(ev, basis, plan)
	}

	type job struct {
		row        int
		start, end int
	}
	type result struct {
		row int
		ct  *rlwe.Ciphertext
		err error
	}

	workers := min(capacity, totalTerms)
	chunkSize := splitLTChunkSize(totalTerms, workers)
	jobs := []job{}
	for r, cols := range rowCols {
		for start := 0; start < len(cols); start += chunkSize {
			end := min(start+chunkSize, len(cols))
			jobs = append(jobs, job{row: r, start: start, end: end})
		}
	}
	if len(jobs) == 0 {
		return evalSplitLTRows(ev, basis, plan)
	}

	jobCh := make(chan job)
	resultCh := make(chan result, len(jobs))
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			w := ev.BLT.GetWorker()
			defer ev.BLT.PutWorker(w)
			for jb := range jobCh {
				ct, err := evalSplitLTPartial(w, ev, basis, plan, rowCols[jb.row][jb.start:jb.end], jb.row)
				resultCh <- result{row: jb.row, ct: ct, err: err}
			}
		}()
	}
	go func() {
		for _, jb := range jobs {
			jobCh <- jb
		}
		close(jobCh)
		wg.Wait()
		close(resultCh)
	}()

	partials := make([][]*rlwe.Ciphertext, R)
	var firstErr error
	for res := range resultCh {
		if res.err != nil {
			if firstErr == nil {
				firstErr = res.err
			}
			continue
		}
		partials[res.row] = append(partials[res.row], res.ct)
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return finishSplitLTPartials(ev, basis, plan, partials)
}

func splitLTChunkSize(totalTerms, workers int) int {
	if workers < 1 {
		return totalTerms
	}
	targetChunks := workers * 4
	if targetChunks < 1 {
		targetChunks = 1
	}
	chunkSize := (totalTerms + targetChunks - 1) / targetChunks
	if chunkSize < 1 {
		return 1
	}
	return chunkSize
}

func evalSplitLTRows(ev *Evaluator, basis []*rlwe.Ciphertext, plan *splitLTPlan) ([]*rlwe.Ciphertext, error) {
	R := len(plan.coeff)
	out := make([]*rlwe.Ciphertext, R)
	errs := make([]error, R)
	var wg sync.WaitGroup
	wg.Add(R)
	for r := 0; r < R; r++ {
		r := r
		go func() {
			defer wg.Done()
			ct, err := evalSplitLTRow(ev, basis, plan, r)
			if err != nil {
				errs[r] = err
				return
			}
			out[r] = ct
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func evalSplitLTPartial(w *blt.Worker, ev *Evaluator, basis []*rlwe.Ciphertext, plan *splitLTPlan, cols []int, r int) (*rlwe.Ciphertext, error) {
	var acc *rlwe.Ciphertext
	first := true
	for _, c := range cols {
		coeff := plan.coeff[r][c]
		tmp := hefloat.NewCiphertext(ev.Ctx.Params, 1, plan.level)
		if err := w.Eval.Mul(basis[c], coeff, tmp); err != nil {
			return nil, fmt.Errorf("split LT row %d col %d: %w", r, c, err)
		}
		if first {
			acc = tmp
			first = false
		} else {
			if err := w.Eval.Add(acc, tmp, acc); err != nil {
				return nil, fmt.Errorf("split LT row %d add: %w", r, err)
			}
		}
	}
	if first {
		return nil, fmt.Errorf("split LT row %d: empty partial", r)
	}
	return acc, nil
}

func finishSplitLTPartials(ev *Evaluator, basis []*rlwe.Ciphertext, plan *splitLTPlan, partials [][]*rlwe.Ciphertext) ([]*rlwe.Ciphertext, error) {
	R := len(plan.coeff)
	out := make([]*rlwe.Ciphertext, R)
	errs := make([]error, R)
	var wg sync.WaitGroup
	wg.Add(R)
	for r := 0; r < R; r++ {
		r := r
		go func() {
			defer wg.Done()
			w := ev.BLT.GetWorker()
			defer ev.BLT.PutWorker(w)
			ct, err := finishSplitLTRow(w, ev, basis, plan, partials[r], r)
			if err != nil {
				errs[r] = err
				return
			}
			out[r] = ct
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func finishSplitLTRow(w *blt.Worker, ev *Evaluator, basis []*rlwe.Ciphertext, plan *splitLTPlan, partials []*rlwe.Ciphertext, r int) (*rlwe.Ciphertext, error) {
	var acc *rlwe.Ciphertext
	if len(partials) > 0 {
		acc = partials[0]
		for _, partial := range partials[1:] {
			if err := w.Eval.Add(acc, partial, acc); err != nil {
				return nil, fmt.Errorf("split LT row %d partial add: %w", r, err)
			}
		}
	} else {
		// All-zero row. Manufacture a zero CT at the right scale by multiplying
		// any non-nil basis entry with the cached zero plaintext.
		var anchor *rlwe.Ciphertext
		for _, ct := range basis {
			if ct != nil {
				anchor = ct
				break
			}
		}
		if anchor == nil {
			return nil, fmt.Errorf("split LT row %d: no basis ciphertexts available", r)
		}
		acc = hefloat.NewCiphertext(ev.Ctx.Params, 1, plan.level)
		if err := w.Eval.Mul(anchor, plan.zeroPT, acc); err != nil {
			return nil, fmt.Errorf("split LT row %d zero anchor: %w", r, err)
		}
	}
	if plan.bias[r] != 0 {
		if err := w.Eval.Add(acc, plan.bias[r], acc); err != nil {
			return nil, fmt.Errorf("split LT row %d bias: %w", r, err)
		}
	}
	if err := w.Eval.Rescale(acc, acc); err != nil {
		return nil, fmt.Errorf("split LT row %d rescale: %w", r, err)
	}
	return acc, nil
}

func evalSplitLTRow(ev *Evaluator, basis []*rlwe.Ciphertext, plan *splitLTPlan, r int) (*rlwe.Ciphertext, error) {
	w := ev.BLT.GetWorker()
	defer ev.BLT.PutWorker(w)

	row := plan.coeff[r]
	var acc *rlwe.Ciphertext
	first := true
	for c, coeff := range row {
		if coeff == 0 || basis[c] == nil {
			continue
		}
		tmp := hefloat.NewCiphertext(ev.Ctx.Params, 1, plan.level)
		if err := w.Eval.Mul(basis[c], coeff, tmp); err != nil {
			return nil, fmt.Errorf("split LT row %d col %d: %w", r, c, err)
		}
		if first {
			acc = tmp
			first = false
		} else {
			if err := w.Eval.Add(acc, tmp, acc); err != nil {
				return nil, fmt.Errorf("split LT row %d add: %w", r, err)
			}
		}
	}
	if first {
		// All-zero row. Manufacture a zero CT at the right scale by multiplying
		// any non-nil basis entry with the cached zero plaintext.
		var anchor *rlwe.Ciphertext
		for _, ct := range basis {
			if ct != nil {
				anchor = ct
				break
			}
		}
		if anchor == nil {
			return nil, fmt.Errorf("split LT row %d: no basis ciphertexts available", r)
		}
		acc = hefloat.NewCiphertext(ev.Ctx.Params, 1, plan.level)
		if err := w.Eval.Mul(anchor, plan.zeroPT, acc); err != nil {
			return nil, fmt.Errorf("split LT row %d zero anchor: %w", r, err)
		}
	}
	if plan.bias[r] != 0 {
		if err := w.Eval.Add(acc, plan.bias[r], acc); err != nil {
			return nil, fmt.Errorf("split LT row %d bias: %w", r, err)
		}
	}
	if err := w.Eval.Rescale(acc, acc); err != nil {
		return nil, fmt.Errorf("split LT row %d rescale: %w", r, err)
	}
	return acc, nil
}

// dropBasisToLevel returns basis CTs aligned to `level`, allocating new
// ciphertexts only for those whose current level is higher than `level`.
func dropBasisToLevel(w *blt.Worker, basis []*rlwe.Ciphertext, level int) []*rlwe.Ciphertext {
	out := make([]*rlwe.Ciphertext, len(basis))
	for c, ct := range basis {
		if ct == nil {
			continue
		}
		if ct.Level() > level {
			out[c] = w.Eval.DropLevelNew(ct, ct.Level()-level)
		} else {
			out[c] = ct
		}
	}
	return out
}

// evalSplitNative computes the per-coord pointwise law: CT_out[c] = CT_x[c] * CT_y[c]
// followed by Rescale, parallelised over c.
func evalSplitNative(ev *Evaluator, xs, ys []*rlwe.Ciphertext) ([]*rlwe.Ciphertext, error) {
	out := make([]*rlwe.Ciphertext, len(xs))
	if err := ev.parallelMulRelin(xs, ys, out); err != nil {
		return nil, err
	}
	return out, nil
}

// evalSplitBinary applies a compiled binary tensor LT. The pairwise basis
// muls run in parallel, the J leading-axis and J trailing-axis inputs are
// dropped to the product level, and the R output rows reduce in parallel.
func evalSplitBinary(ev *Evaluator, xs, ys []*rlwe.Ciphertext, plan *splitLTPlan) ([]*rlwe.Ciphertext, error) {
	J := len(xs)
	if len(ys) != J {
		return nil, fmt.Errorf("evalSplitBinary: input arity mismatch: %d vs %d", J, len(ys))
	}
	N := (1 + J) * (1 + J)
	if len(plan.coeff) == 0 || len(plan.coeff[0]) != N {
		return nil, fmt.Errorf("evalSplitBinary: plan basis size %d, expected %d", len(plan.coeff[0]), N)
	}

	pairs := J * J
	lefts := make([]*rlwe.Ciphertext, pairs)
	rights := make([]*rlwe.Ciphertext, pairs)
	for i := 0; i < J; i++ {
		for j := 0; j < J; j++ {
			lefts[i*J+j] = xs[i]
			rights[i*J+j] = ys[j]
		}
	}
	prods := make([]*rlwe.Ciphertext, pairs)
	if err := ev.parallelMulRelin(lefts, rights, prods); err != nil {
		return nil, fmt.Errorf("evalSplitBinary: basis mul: %w", err)
	}

	w := ev.BLT.GetWorker()
	droppedXs := dropBasisToLevel(w, xs, plan.level)
	droppedYs := dropBasisToLevel(w, ys, plan.level)
	ev.BLT.PutWorker(w)

	basis := make([]*rlwe.Ciphertext, N)
	for j := 0; j < J; j++ {
		basis[j+1] = droppedYs[j]
	}
	for i := 0; i < J; i++ {
		basis[(i+1)*(1+J)] = droppedXs[i]
	}
	for i := 0; i < J; i++ {
		for j := 0; j < J; j++ {
			basis[(i+1)*(1+J)+(j+1)] = prods[i*J+j]
		}
	}
	return evalSplitLT(ev, basis, plan)
}

func benchSplitNative(b *testing.B, enc basicEncoding, log2t int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	ctx := newBasicContext(b, 1)
	ev := NewEvaluatorWithWorkerCapacity(ctx, splitWorkerCapacity())
	batch := splitBatch(b)

	xs := randValues(t, batch, 11)
	ys := randValues(t, batch, 12)
	if enc.kind == charenc.IND {
		copy(ys, xs)
	}
	wantValues := make([]int, batch)
	for i := range wantValues {
		wantValues[i] = nativeLaw(enc.kind, t, xs[i], ys[i])
	}
	xCTs, err := encryptSplitInputs(ctx, spec, xs, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	yCTs, err := encryptSplitInputs(ctx, spec, ys, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xPlain, _ := encodeSplitInputs(spec, xs, ctx.Params.MaxSlots())
	wantPlain, _ := encodeSplitInputs(spec, wantValues, ctx.Params.MaxSlots())
	inDecoded, err := decryptSplitInputs(ctx, xCTs)
	if err != nil {
		b.Fatal(err)
	}
	out, err := evalSplitNative(ev, xCTs, yCTs)
	if err != nil {
		b.Fatal(err)
	}
	outDecoded, err := decryptSplitInputs(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	inPrec, inSat := precisionBits(maxAbsSplitError(inDecoded, xPlain, batch))
	outPrec, outSat := precisionBits(maxAbsSplitError(outDecoded, wantPlain, batch))
	correct := checkSplitOutput(b, spec, outDecoded, wantValues)
	b.ReportAllocs()
	b.ResetTimer()
	reportSplitMetrics(b, ctx, spec, log2t, t, batch, 1, inPrec, outPrec, inSat, outSat, correct)
	for i := 0; i < b.N; i++ {
		if _, err := evalSplitNative(ev, xCTs, yCTs); err != nil {
			b.Fatal(err)
		}
	}
}

func benchSplitUnaryLUT(b *testing.B, enc basicEncoding, log2t int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	ctx := newBasicContext(b, 1)
	ev := NewEvaluatorWithWorkerCapacity(ctx, splitWorkerCapacity())
	batch := splitBatch(b)

	table := randomLUTTable(t, 1)
	fn := func(x int) int { return table[positiveMod(x, t)] }

	codec, err := charenc.NewCodec(spec)
	if err != nil {
		b.Fatal(err)
	}
	tr, err := blt.CompileUnary(codec, codec, fn)
	if err != nil {
		b.Fatal(err)
	}
	plan, err := compileSplitLTPlan(ctx, tr.Matrix, tr.Bias, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}

	xs := randValues(t, batch, 21)
	wantValues := make([]int, batch)
	for i, x := range xs {
		wantValues[i] = fn(x)
	}
	xCTs, err := encryptSplitInputs(ctx, spec, xs, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xPlain, _ := encodeSplitInputs(spec, xs, ctx.Params.MaxSlots())
	wantPlain, _ := encodeSplitInputs(spec, wantValues, ctx.Params.MaxSlots())
	inDecoded, err := decryptSplitInputs(ctx, xCTs)
	if err != nil {
		b.Fatal(err)
	}
	out, err := evalSplitLT(ev, xCTs, plan)
	if err != nil {
		b.Fatal(err)
	}
	outDecoded, err := decryptSplitInputs(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	inPrec, inSat := precisionBits(maxAbsSplitError(inDecoded, xPlain, batch))
	outPrec, outSat := precisionBits(maxAbsSplitError(outDecoded, wantPlain, batch))
	correct := checkSplitOutput(b, spec, outDecoded, wantValues)
	b.ReportAllocs()
	b.ResetTimer()
	reportSplitMetrics(b, ctx, spec, log2t, t, batch, 1, inPrec, outPrec, inSat, outSat, correct)
	for i := 0; i < b.N; i++ {
		if _, err := evalSplitLT(ev, xCTs, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func benchSplitBinaryLUT(b *testing.B, enc basicEncoding, log2t int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	ctx := newBasicContext(b, 2)
	ev := NewEvaluatorWithWorkerCapacity(ctx, splitWorkerCapacity())
	batch := splitBatch(b)

	table := randomLUTTable(t, 2)
	fn := func(xy []int) int { return table[lutTableIndex(t, xy)] }

	codec, err := charenc.NewCodec(spec)
	if err != nil {
		b.Fatal(err)
	}
	J := spec.Slots
	coef, err := interpolateAugmentedTensor([]charenc.Codec{codec, codec}, codec, []int{J, J}, fn)
	if err != nil {
		b.Fatal(err)
	}
	// The augmented-tensor coefficient at position (0, 0) acts as the row's
	// additive constant. Lift it into the LT plan's bias vector so the
	// per-row reduction handles it via a scalar add.
	R := len(coef)
	bias := make([]complex128, R)
	matrix := make([][]complex128, R)
	for r := 0; r < R; r++ {
		bias[r] = coef[r][0]
		matrix[r] = make([]complex128, len(coef[r]))
		copy(matrix[r], coef[r])
		matrix[r][0] = 0
	}

	level := ctx.Params.MaxLevel() - ctx.Params.LevelsConsumedPerRescaling()
	plan, err := compileSplitLTPlan(ctx, matrix, bias, level)
	if err != nil {
		b.Fatal(err)
	}

	xs := randValues(t, batch, 31)
	ys := randValues(t, batch, 32)
	wantValues := make([]int, batch)
	tmp := make([]int, 2)
	for i := range xs {
		tmp[0] = xs[i]
		tmp[1] = ys[i]
		wantValues[i] = fn(tmp)
	}
	xCTs, err := encryptSplitInputs(ctx, spec, xs, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	yCTs, err := encryptSplitInputs(ctx, spec, ys, ctx.Params.MaxLevel())
	if err != nil {
		b.Fatal(err)
	}
	xPlain, _ := encodeSplitInputs(spec, xs, ctx.Params.MaxSlots())
	wantPlain, _ := encodeSplitInputs(spec, wantValues, ctx.Params.MaxSlots())
	inDecoded, err := decryptSplitInputs(ctx, xCTs)
	if err != nil {
		b.Fatal(err)
	}
	out, err := evalSplitBinary(ev, xCTs, yCTs, plan)
	if err != nil {
		b.Fatal(err)
	}
	outDecoded, err := decryptSplitInputs(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	inPrec, inSat := precisionBits(maxAbsSplitError(inDecoded, xPlain, batch))
	outPrec, outSat := precisionBits(maxAbsSplitError(outDecoded, wantPlain, batch))
	correct := checkSplitOutput(b, spec, outDecoded, wantValues)
	b.ReportAllocs()
	b.ResetTimer()
	reportSplitMetrics(b, ctx, spec, log2t, t, batch, 2, inPrec, outPrec, inSat, outSat, correct)
	for i := 0; i < b.N; i++ {
		if _, err := evalSplitBinary(ev, xCTs, yCTs, plan); err != nil {
			b.Fatal(err)
		}
	}
}

type splitCleanPlan struct {
	kind    charenc.EncodingKind
	toIND   *splitLTPlan
	fromIND *splitLTPlan
}

func compileSplitCleanPlan(ctx *charctx.Context, kind charenc.EncodingKind, spec charenc.BlockSpec, t int) (*splitCleanPlan, error) {
	if kind == charenc.IND || kind == charenc.WH {
		return &splitCleanPlan{kind: kind}, nil
	}
	indCodec, err := charenc.NewIND(t, true, 0)
	if err != nil {
		return nil, err
	}
	codec, err := charenc.NewCodec(spec)
	if err != nil {
		return nil, err
	}
	fnID := func(x int) int { return x }
	toIND, err := blt.CompileUnary(codec, indCodec, fnID)
	if err != nil {
		return nil, err
	}
	fromIND, err := blt.CompileUnary(indCodec, codec, fnID)
	if err != nil {
		return nil, err
	}
	rescale := ctx.Params.LevelsConsumedPerRescaling()
	toLevel := ctx.Params.MaxLevel()
	fromLevel := ctx.Params.MaxLevel() - 3*rescale
	toPlan, err := compileSplitLTPlan(ctx, toIND.Matrix, toIND.Bias, toLevel)
	if err != nil {
		return nil, err
	}
	fromPlan, err := compileSplitLTPlan(ctx, fromIND.Matrix, fromIND.Bias, fromLevel)
	if err != nil {
		return nil, err
	}
	return &splitCleanPlan{kind: kind, toIND: toPlan, fromIND: fromPlan}, nil
}

// evalSplitClean cleans a split ciphertext block. IND and WH use a per-coord
// polynomial cleaner; BRU and LBRU pivot through an IND representation via
// two split linear transforms.
func evalSplitClean(ev *Evaluator, x []*rlwe.Ciphertext, plan *splitCleanPlan) ([]*rlwe.Ciphertext, error) {
	switch plan.kind {
	case charenc.IND:
		return evalSplitINDPolyClean(ev, x)
	case charenc.WH:
		return evalSplitWHPolyClean(ev, x)
	}
	ind, err := evalSplitLT(ev, x, plan.toIND)
	if err != nil {
		return nil, fmt.Errorf("split clean: to IND: %w", err)
	}
	cleaned, err := evalSplitINDPolyClean(ev, ind)
	if err != nil {
		return nil, fmt.Errorf("split clean: IND poly: %w", err)
	}
	return evalSplitLT(ev, cleaned, plan.fromIND)
}

// evalSplitINDPolyClean applies 3*x^2 - 2*x^3 to every coord ciphertext in
// parallel. Matches evalBasicINDClean's polynomial.
func evalSplitINDPolyClean(ev *Evaluator, cts []*rlwe.Ciphertext) ([]*rlwe.Ciphertext, error) {
	return evalSplitPerCoordPoly(ev, cts, indPolyCoord)
}

// evalSplitWHPolyClean applies 3*x - x^3 to every coord ciphertext in parallel.
// Matches evalBasicWHClean's polynomial (including its final scale doubling).
func evalSplitWHPolyClean(ev *Evaluator, cts []*rlwe.Ciphertext) ([]*rlwe.Ciphertext, error) {
	return evalSplitPerCoordPoly(ev, cts, whPolyCoord)
}

func evalSplitPerCoordPoly(ev *Evaluator, cts []*rlwe.Ciphertext, poly func(*blt.Worker, *charctx.Context, *rlwe.Ciphertext) (*rlwe.Ciphertext, error)) ([]*rlwe.Ciphertext, error) {
	n := len(cts)
	out := make([]*rlwe.Ciphertext, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			w := ev.BLT.GetWorker()
			defer ev.BLT.PutWorker(w)
			ct, err := poly(w, ev.Ctx, cts[i])
			if err != nil {
				errs[i] = err
				return
			}
			out[i] = ct
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func indPolyCoord(w *blt.Worker, ctx *charctx.Context, ind *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	x2 := hefloat.NewCiphertext(ctx.Params, 1, ind.Level())
	if err := w.Eval.MulRelin(ind, ind, x2); err != nil {
		return nil, err
	}
	if err := w.Eval.Rescale(x2, x2); err != nil {
		return nil, err
	}
	indAtX2 := ind
	if indAtX2.Level() > x2.Level() {
		indAtX2 = w.Eval.DropLevelNew(indAtX2, indAtX2.Level()-x2.Level())
	}
	x3 := hefloat.NewCiphertext(ctx.Params, 1, x2.Level())
	if err := w.Eval.MulRelin(x2, indAtX2, x3); err != nil {
		return nil, err
	}
	if err := w.Eval.Rescale(x3, x3); err != nil {
		return nil, err
	}
	x2AtX3 := x2
	if x2AtX3.Level() > x3.Level() {
		x2AtX3 = w.Eval.DropLevelNew(x2AtX3, x2AtX3.Level()-x3.Level())
	}
	term2 := hefloat.NewCiphertext(ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(x2AtX3, 3, term2); err != nil {
		return nil, err
	}
	term3 := hefloat.NewCiphertext(ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(x3, 2, term3); err != nil {
		return nil, err
	}
	clean := hefloat.NewCiphertext(ctx.Params, 1, x3.Level())
	if err := w.Eval.Sub(term2, term3, clean); err != nil {
		return nil, err
	}
	return clean, nil
}

func whPolyCoord(w *blt.Worker, ctx *charctx.Context, x *rlwe.Ciphertext) (*rlwe.Ciphertext, error) {
	x2 := hefloat.NewCiphertext(ctx.Params, 1, x.Level())
	if err := w.Eval.MulRelin(x, x, x2); err != nil {
		return nil, err
	}
	if err := w.Eval.Rescale(x2, x2); err != nil {
		return nil, err
	}
	xAtX2 := x
	if xAtX2.Level() > x2.Level() {
		xAtX2 = w.Eval.DropLevelNew(xAtX2, xAtX2.Level()-x2.Level())
	}
	x3 := hefloat.NewCiphertext(ctx.Params, 1, x2.Level())
	if err := w.Eval.MulRelin(x2, xAtX2, x3); err != nil {
		return nil, err
	}
	if err := w.Eval.Rescale(x3, x3); err != nil {
		return nil, err
	}
	xAtX3 := x
	if xAtX3.Level() > x3.Level() {
		xAtX3 = w.Eval.DropLevelNew(xAtX3, xAtX3.Level()-x3.Level())
	}
	term1 := hefloat.NewCiphertext(ctx.Params, 1, x3.Level())
	if err := w.Eval.Mul(xAtX3, 3, term1); err != nil {
		return nil, err
	}
	out := hefloat.NewCiphertext(ctx.Params, 1, x3.Level())
	if err := w.Eval.Sub(term1, x3, out); err != nil {
		return nil, err
	}
	out.Scale = out.Scale.Mul(rlwe.NewScale(2))
	return out, nil
}

func benchSplitClean(b *testing.B, enc basicEncoding, log2t int) {
	spec, t, err := basicSpec(enc.kind, log2t)
	if err != nil {
		b.Fatal(err)
	}
	levels := 4
	if enc.kind == charenc.IND || enc.kind == charenc.WH {
		levels = 2
	}
	ctx := newBasicContext(b, levels)
	ev := NewEvaluatorWithWorkerCapacity(ctx, splitWorkerCapacity())
	batch := splitBatch(b)

	plan, err := compileSplitCleanPlan(ctx, enc.kind, spec, t)
	if err != nil {
		b.Fatal(err)
	}

	xs := randValues(t, batch, 41)
	exact, err := encodeSplitInputs(spec, xs, ctx.Params.MaxSlots())
	if err != nil {
		b.Fatal(err)
	}
	noisy := make([][]complex128, len(exact))
	for c := range exact {
		noisy[c] = append([]complex128(nil), exact[c]...)
		for i := range noisy[c] {
			noise := math.Ldexp(math.Sin(float64((c*1000+i)*17+3)), -12)
			noisy[c][i] += complex(noise, -0.5*noise)
		}
	}
	xCTs := make([]*rlwe.Ciphertext, len(noisy))
	for c, vec := range noisy {
		ct, err := encryptVector(ctx, vec, ctx.Params.MaxLevel())
		if err != nil {
			b.Fatal(err)
		}
		xCTs[c] = ct
	}
	inDecoded, err := decryptSplitInputs(ctx, xCTs)
	if err != nil {
		b.Fatal(err)
	}
	out, err := evalSplitClean(ev, xCTs, plan)
	if err != nil {
		b.Fatal(err)
	}
	outDecoded, err := decryptSplitInputs(ctx, out)
	if err != nil {
		b.Fatal(err)
	}
	inPrec, inSat := precisionBits(maxAbsSplitError(inDecoded, exact, batch))
	outPrec, outSat := precisionBits(maxAbsSplitError(outDecoded, exact, batch))
	correct := checkSplitOutput(b, spec, outDecoded, xs)
	b.ReportAllocs()
	b.ResetTimer()
	reportSplitMetrics(b, ctx, spec, log2t, t, batch, levels, inPrec, outPrec, inSat, outSat, correct)
	for i := 0; i < b.N; i++ {
		if _, err := evalSplitClean(ev, xCTs, plan); err != nil {
			b.Fatal(err)
		}
	}
}

func runSplitBenchmarks(b *testing.B, fn func(*testing.B, basicEncoding, int)) {
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

func BenchmarkSplitBasicNative(b *testing.B) {
	runSplitBenchmarks(b, benchSplitNative)
}

func BenchmarkSplitBasicUnaryLUT(b *testing.B) {
	runSplitBenchmarks(b, benchSplitUnaryLUT)
}

func BenchmarkSplitBasicBinaryLUT(b *testing.B) {
	runSplitBenchmarks(b, benchSplitBinaryLUT)
}

func BenchmarkSplitBasicClean(b *testing.B) {
	runSplitBenchmarks(b, benchSplitClean)
}
