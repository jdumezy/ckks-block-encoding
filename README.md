# character-encoding

This artifact contains Go implementations and benchmark entry points for the [Character Block Encodings for Discrete CKKS](https://eprint.iacr.org/2026/1200) paper.
It is intended for benchmark reproduction, not as a general-purpose library.

Terminology warning: the paper calls the former `IND` character encoding `IDCT` to avoid confusion with IND-CPA security; code, benchmark names, and result files still use `IND`.

## Layout

- `character-encoding/`: implementation packages for character encodings, block linear transforms, lookup tables, CRT arithmetic, and packed radix arithmetic.
- `character-encoding/cmd/precision_chain/`: standalone Go binary for the long-chain precision experiments (chained unary LUTs, native multiplication, CRT multiplication, and the cleaning-gain measurement).
- `lattigo/`: local source replacement for `github.com/Pro7ech/lattigo`.
- `scripts/run_benchmarks.py`: benchmark runner with JSON, Markdown, and heartbeat output.
- `scripts/precision_chain.py`: driver and plotter for the long-chain precision experiments described in [Precision Chain Experiment](#precision-chain-experiment).
- `Dockerfile`: container environment for running the benchmark suite.

## Requirements

- Go 1.23+ and Python 3.11+. The published `bench_results/` were produced with Go 1.25.5 and Python 3.14.4.
- Matplotlib is required only for generating the precision-chain PNG plots (`python3 -m pip install matplotlib`).
- Linux is recommended for CPU and memory metadata in `heartbeat.jsonl`.

## Run Benchmarks

Quick smoke run (one radix Add benchmark, single sample; finishes in well under a minute on a modern desktop):

```sh
python3 scripts/run_benchmarks.py \
  --target radix \
  --mode single \
  --ops Add \
  --shapes 64bit_r4 \
  --count 1
```

Full run used for the benchmark tables (portable defaults, 10 samples per benchmark):

```sh
python3 scripts/run_benchmarks.py \
  --target all \
  --mode both \
  --layout both
```

This runs radix benchmarks at `LogN=16`, packed CRT benchmarks at `LogN=15`, and packed/split basic benchmarks at `LogN=15`. Expect roughly **5-6 hours** on a 16-core Zen 5 desktop; use `--count 1` for quick checks.

To reduce single-thread variance, pin single-thread benchmarks to one physical core via `--bench-core <id>` (multi-thread benchmarks ignore the flag):

```sh
python3 scripts/run_benchmarks.py \
  --target all --mode both --layout both \
  --bench-core 4
```

Notes: `split` is experimental and capped at `log2t <= 8`; `from_standard` is `BRU`-only and capped at `log2t <= 7` for the `LogN=15` chain. LUT tables use fixed seeds. Timings cover only the online homomorphic operation; setup, encryption, compilation, decryption, and correctness checks happen before `b.ResetTimer()`. Use `--help` for sweep controls.

Each run creates a timestamped directory:

```text
bench_results/<timestamp>/
  heartbeat.jsonl
  results.json
  summary.md
```

`heartbeat.jsonl` records commands, selected environment, durations, return codes, parsed samples, correctness metrics, and redacted output tails.

## Docker

Build:

```sh
docker build -t character-encoding .
```

Run the container's default smoke benchmark:

```sh
docker run --rm -v "$PWD/bench_results:/workspace/bench_results" character-encoding
```

Run a longer sweep by passing runner arguments:

```sh
docker run --rm -v "$PWD/bench_results:/workspace/bench_results" \
  character-encoding --target all --mode both --layout both
```

The container uses portable defaults and Go from `golang:1.23-bookworm`; committed artifacts record their exact Go/Python versions.

## Security Parameters

The runner reports levels, `LogQ`, `LogP`, `LogQP`, and the max 128-bit-security `LogQP` for sparse-secret `h=256`. It exits non-zero on failed commands, incorrect outputs, or insecure `LogQP`; bootstrapping converters are checked against their emitted `LogQP`, with `LogQ`/`LogP` shown as compact chain summaries.

A green outcome therefore looks like:

- the runner exit code is `0`,
- every row in `summary.md` shows `Correct: yes`, and
- every Parameter Check table shows `Secure: yes`.

`Max Error` is the maximum correctness-check deviation; integer-exact operations should report `0`.

## Reading the results

`summary.md` is grouped by mode and operation family. `results.json` contains the raw parsed samples; `heartbeat.jsonl` contains append-only command and output-tail metadata.

## Precision Chain Experiment

This experiment measures precision across 100-operation chains with EvalRound bootstrapping whenever levels are exhausted. It covers unary `BRU` LUTs and native `BRU` multiplication at `log2t in {2,4,6,8}`, plus 256-bit CRT `LBRU` multiplication, all at `LogN=16`, `h=256`, and post-bootstrap level 11. Each configuration uses three seeds. The cleaning-gain sweep injects controlled noise and measures recovery through the `to-IDCT -> polynomial -> from-IDCT` fallback path, named `to-IND`/`from-IND` in code.

Run the full experiment (3 seeds, 100 iters, 9 configurations):

```sh
python3 scripts/precision_chain.py \
  --seeds 3 --max-iters 100 --logn 16 --encoding BRU \
  --out-dir bench_results/runs/precision_chain
```

This writes PNG plots and `results_bru.json` to `bench_results/runs/precision_chain/`; the full run takes roughly 2 hours on a 16-core Zen 5 desktop. The committed `bench_results/precision_chain/` directory contains the paper's reference results and is not overwritten.

Run the cleaning-gain inject-bits sweep (single seed, four `log2t` values, nine injection levels):

```sh
for bits in 3 6 9 12 14 16 18 20 22; do
  go run ./character-encoding/cmd/precision_chain \
    -clean-gain -logn 16 -seeds 1 -encoding BRU \
    -clean-inject-bits $bits \
    -out bench_results/clean_gain/sweep_b${bits}.json
done
```

The resulting JSON files report plateau precision, envelope rate, cleaning gain, and equivalent iterations saved.

## Citation

```bibtex
@misc{character-encoding,
      author = {Jules Dumezy and Elias Suvanto},
      title = {Character Block Encodings for Discrete {CKKS}: Single-Level {LUTs} and Low-Depth Arithmetic},
      howpublished = {Cryptology {ePrint} Archive, Paper 2026/1200},
      year = {2026},
      url = {https://eprint.iacr.org/2026/1200}
}
```

## Attribution

- The local `lattigo/` source is derived from `https://github.com/Pro7ech/lattigo`.
- The functional bootstrapping for converting is partially derived from `https://github.com/jaehyungkim0/CRT-FHE`.
- The 128-bit sparse-secret security limits embedded in `scripts/run_benchmarks.py` are derived from `https://github.com/jdumezy/sparse-key-estimate.git`.

## License

This project is licensed under the Apache License 2.0. See `LICENSE`. Third-party attribution and license notices are retained in `NOTICE` and `lattigo/LICENCE`.
