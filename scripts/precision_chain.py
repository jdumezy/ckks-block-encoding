#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from collections import defaultdict
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
GO_PKG = "./character-encoding/cmd/precision_chain"

COLORBLIND_PALETTE = [
    "#0072B2",
    "#D55E00",
    "#009E73",
    "#CC79A7",
    "#F0E442",
    "#56B4E9",
    "#E69F00",
    "#000000",
]
MARKERS = ["o", "s", "^", "D", "v", "P", "X", "*"]
LINESTYLES = ["-", "--", "-.", ":", (0, (3, 1, 1, 1)), (0, (5, 1)), (0, (1, 1)), (0, (3, 5, 1, 5))]


def run_go(out_path: Path, seeds: int, max_iters: int, logn: int, encoding: str, extra: list[str]) -> None:
    cmd = [
        "go", "run", GO_PKG,
        "-seeds", str(seeds),
        "-max-iters", str(max_iters),
        "-logn", str(logn),
        "-encoding", encoding,
        "-out", str(out_path),
    ] + extra
    print("$", " ".join(cmd), flush=True)
    t0 = time.time()
    subprocess.run(cmd, cwd=REPO, check=True)
    print(f"Go run finished in {time.time() - t0:.1f}s", flush=True)


def load(path: Path) -> dict:
    with path.open() as fh:
        return json.load(fh)


def group(results: list[dict]) -> dict[tuple[str, int], list[list[dict]]]:
    out: dict[tuple[str, int], list[list[dict]]] = defaultdict(list)
    for r in results:
        out[(r["experiment"], r["log2t"])].append(r["iterations"])
    return out


def average_curve(runs: list[list[dict]]) -> tuple[list[int], list[float]]:
    max_iter = max((len(r) for r in runs), default=0)
    iters: list[int] = []
    means: list[float] = []
    for i in range(max_iter):
        vals = [r[i]["precision_bits"] for r in runs if i < len(r) and r[i]["correct"]]
        if not vals:
            break
        iters.append(i)
        means.append(sum(vals) / len(vals))
    return iters, means


def cleaning_iters(runs: list[list[dict]]) -> list[int]:
    seen: set[int] = set()
    for r in runs:
        for it in r:
            if it.get("cleaned"):
                seen.add(it["iter"])
    return sorted(seen)


def experiment_families(results: list[dict]) -> dict[str, list[tuple[str, int]]]:
    families: dict[str, list[tuple[str, int]]] = defaultdict(list)
    for r in results:
        exp = r["experiment"]
        log2t = r["log2t"]
        if exp.startswith("unary_lut_"):
            family = "unary_lut"
        elif exp.startswith("native_") and exp != "native":
            family = "native"
        else:
            family = exp
        key = (exp, log2t)
        if key not in families[family]:
            families[family].append(key)
    for v in families.values():
        v.sort()
    return families


def family_title(family: str, encoding: str, logn: int) -> str:
    if family == "unary_lut":
        return f"Chained unary {encoding} LUTs (LogN={logn}, packed)"
    if family == "native":
        return f"Chained native {encoding} multiplication (LogN={logn}, packed)"
    if family == "crt_lbru_256":
        return f"Chained 256-bit CRT LBRU multiplication (LogN={logn}, packed)"
    return family


def plot(grouped: dict[tuple[str, int], list[list[dict]]], families: dict[str, list[tuple[str, int]]], encoding: str, logn: int, out_dir: Path) -> None:
    import matplotlib.pyplot as plt
    out_dir.mkdir(parents=True, exist_ok=True)
    for family, keys in families.items():
        fig, ax = plt.subplots(figsize=(8, 5))
        any_data = False
        for idx, (exp, log2t) in enumerate(keys):
            runs = grouped.get((exp, log2t), [])
            iters, means = average_curve(runs)
            if not iters:
                continue
            color = COLORBLIND_PALETTE[idx % len(COLORBLIND_PALETTE)]
            marker = MARKERS[idx % len(MARKERS)]
            linestyle = LINESTYLES[idx % len(LINESTYLES)]
            label = f"log2t={log2t}" if family != "crt_lbru_256" else "256-bit LBRU"
            ax.plot(iters, means, color=color, marker=marker, markevery=max(1, len(iters) // 25),
                    linestyle=linestyle, linewidth=1.6, markersize=6, label=label)
            cleans = cleaning_iters(runs)
            if cleans:
                clean_xs = [c for c in cleans if c < len(means)]
                clean_ys = [means[c] for c in clean_xs]
                ax.scatter(clean_xs, clean_ys, marker="*", s=180, color=color,
                           edgecolors="black", linewidths=1.0, zorder=5)
            any_data = True
        if not any_data:
            plt.close(fig)
            continue
        ax.set_title(family_title(family, encoding, logn))
        ax.set_xlabel("Iteration (chained operations)")
        ax.set_ylabel("Precision (bits)")
        ax.grid(True, linestyle=":", alpha=0.6)
        ax.legend(loc="best", framealpha=0.9)
        fig.tight_layout()
        path = out_dir / f"precision_{family}_{encoding.lower()}.png"
        fig.savefig(path, dpi=140)
        plt.close(fig)
        print(f"wrote {path}", flush=True)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--seeds", type=int, default=10)
    ap.add_argument("--max-iters", type=int, default=100)
    ap.add_argument("--logn", type=int, default=16)
    ap.add_argument("--encoding", default="BRU", choices=["BRU", "LBRU", "WH", "IND"])
    ap.add_argument("--results", type=Path, default=None)
    ap.add_argument("--out-dir", type=Path, default=REPO / "bench_results" / "precision_chain")
    ap.add_argument("--go-extra", default="")
    args = ap.parse_args()

    out_dir: Path = args.out_dir
    out_dir.mkdir(parents=True, exist_ok=True)
    json_path = args.results or (out_dir / f"results_{args.encoding.lower()}.json")

    if args.results is None or not args.results.exists():
        extras = [a for a in args.go_extra.split(",") if a]
        run_go(json_path, args.seeds, args.max_iters, args.logn, args.encoding, extras)

    payload = load(json_path)
    grouped = group(payload["results"])
    families = experiment_families(payload["results"])
    plot(grouped, families, args.encoding, payload["logN"], out_dir)
    return 0


if __name__ == "__main__":
    sys.exit(main())
