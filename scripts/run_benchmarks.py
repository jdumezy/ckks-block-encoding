#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import math
import os
import platform
import re
import shutil
import subprocess
import sys
import time
from dataclasses import asdict, dataclass, field
from pathlib import Path


SECRET_H = 256
RADIX_LOGN = 16
CRT_LOGN = 15
BASIC_LOGN = 15
BASIC_MAX_SLOTS = 1 << (BASIC_LOGN - 1)
SECURITY_LIMITS: dict[int, dict[int, int]] = {
    12: {32: 19, 64: 58, 128: 86, 192: 96, 256: 101, 512: 106, 1024: 111},
    13: {32: 37, 64: 113, 128: 167, 192: 187, 256: 197, 512: 209, 1024: 214},
    14: {32: 77, 64: 229, 128: 333, 192: 370, 256: 388, 512: 416, 1024: 426},
    15: {32: 164, 64: 468, 128: 669, 192: 739, 256: 774, 512: 829, 1024: 854},
    16: {32: 349, 64: 961, 128: 1351, 192: 1481, 256: 1553, 512: 1659, 1024: 1714},
}
RADIX_OPS = ["Add", "Sub", "Eq", "Lt", "Cmp", "Clean"]
RADIX_SHAPES = ["64bit_r4", "256bit_r4"]
BASIC_OPS: dict[str, tuple[str, str, int, int]] = {
    "native": ("native", r"^BenchmarkBasicNative", 1, 1),
    "unary_lut": ("unary_lut", r"^BenchmarkBasicUnaryLUT", 1, 1),
    "binary_lut": ("binary_lut", r"^BenchmarkBasicBinaryLUT", 3, 2),
    "four_lut": ("four_lut", r"^BenchmarkBasicFourLUT", 4, 4),
    "clean": ("clean", r"^BenchmarkBasicClean", 4, 1),
}
BASIC_OPS_SPLIT: dict[str, tuple[str, str, int, int]] = {
    "native": ("native", r"^BenchmarkSplitBasicNative", 1, 1),
    "unary_lut": ("unary_lut", r"^BenchmarkSplitBasicUnaryLUT", 1, 1),
    "binary_lut": ("binary_lut", r"^BenchmarkSplitBasicBinaryLUT", 2, 2),
    "clean": ("clean", r"^BenchmarkSplitBasicClean", 4, 1),
    "to_standard": ("to_standard", r"^BenchmarkSplitToStandard", 1, 1),
    "from_standard": ("from_standard", r"^BenchmarkSplitFromStandard", 13, 1),
}
BASIC_ENCODINGS = ["BRU", "LBRU", "WH", "IND"]
BASIC_LOG2T = [2, 3, 4, 5, 6, 7, 8, 9, 10]
BASIC_SPLIT_MAX_LOG2T = 8
BASIC_SPLIT_OP_ENCODINGS: dict[str, list[str]] = {
    "to_standard": ["BRU"],
    "from_standard": ["BRU"],
}
BASIC_SPLIT_OP_MAX_LOG2T: dict[str, int] = {
    "from_standard": 7,
}
BASIC_NATIVE_LOG2T = BASIC_LOG2T
BASIC_UNARY_LOG2T = BASIC_LOG2T
BASIC_BINARY_LOG2T = [2, 3, 4, 5]
BASIC_FOUR_LOG2T = [2, 3]
BASIC_OP_LABELS = {
    "native": "Native",
    "unary_lut": "Unary LUT",
    "binary_lut": "Bivariate LUT",
    "four_lut": "4-Variate LUT",
    "clean": "Clean",
    "to_standard": "Split->Standard",
    "from_standard": "Standard->Split",
}

CRT_OPS: dict[str, tuple[str, str, int]] = {
    "add": ("add", r"^BenchmarkPackedCRTOnlineNativeBRUAdd", 1),
    "sub": ("sub", r"^BenchmarkPackedCRTOnlineNativeBRUSub", 1),
    "mul_lbru": ("mul_lbru", r"^BenchmarkPackedCRTOnlineNativeLBRUMul", 1),
    "bru_to_lbru": ("bru_to_lbru", r"^BenchmarkPackedCRTOnlineSwitch/BRUToLBRU", 1),
    "lbru_to_bru": ("lbru_to_bru", r"^BenchmarkPackedCRTOnlineSwitch/LBRUToBRU", 1),
    "clean": ("clean", r"^BenchmarkPackedCRTOnlineCleanBRU", 4),
    "packed_add": ("add", r"^BenchmarkPackedCRTOnlineNativeBRUAdd", 1),
    "packed_sub": ("sub", r"^BenchmarkPackedCRTOnlineNativeBRUSub", 1),
    "packed_mul_lbru": ("mul_lbru", r"^BenchmarkPackedCRTOnlineNativeLBRUMul", 1),
    "packed_bru_to_lbru": ("bru_to_lbru", r"^BenchmarkPackedCRTOnlineSwitch/BRUToLBRU", 1),
    "packed_lbru_to_bru": ("lbru_to_bru", r"^BenchmarkPackedCRTOnlineSwitch/LBRUToBRU", 1),
    "packed_clean": ("clean", r"^BenchmarkPackedCRTOnlineCleanBRU", 4),
}

BENCH_LINE_RE = re.compile(
    r"^(?P<name>\S+?)(?:-\d+)?\s+(?P<n>\d+)\s+(?P<ns>\d+(?:\.\d+)?)\s+ns/op"
)
METRIC_RE = re.compile(r"(-?\d+(?:\.\d+)?)\s+([A-Za-z_][A-Za-z_0-9\-/]*)")


@dataclass
class BenchSample:
    name: str
    ns_per_op: float
    metrics: dict[str, float] = field(default_factory=dict)


@dataclass
class BenchResult:
    target: str
    op: str
    shape: str
    logN: int
    mode: str
    bench_regex: str
    package: str
    ok: bool
    returncode: int
    duration_s: float
    samples: list[BenchSample] = field(default_factory=list)
    output_tail: str = ""
    layout: str = "packed"

    def mean_ns(self) -> float:
        if not self.samples:
            return float("nan")
        return sum(s.ns_per_op for s in self.samples) / len(self.samples)

    def stddev_ns(self) -> float:
        if len(self.samples) < 2:
            return 0.0
        mean = self.mean_ns()
        return math.sqrt(sum((s.ns_per_op - mean) ** 2 for s in self.samples) / (len(self.samples) - 1))

    def metric(self, name: str) -> float | None:
        for sample in self.samples:
            if name in sample.metrics:
                return sample.metrics[name]
        return None

    def correct(self) -> bool:
        values = [sample.metrics.get("correct") for sample in self.samples if "correct" in sample.metrics]
        return bool(values) and all(value is not None and value >= 0.5 for value in values)


class Heartbeat:
    def __init__(self, path: Path):
        self.path = path
        self.path.parent.mkdir(parents=True, exist_ok=True)
        self.fh = self.path.open("a", encoding="utf-8")

    def write(self, event: str, **fields: object) -> None:
        row = {"time_utc": utc_stamp(), "event": event}
        row.update(fields)
        self.fh.write(json.dumps(row, sort_keys=True) + "\n")
        self.fh.flush()

    def close(self) -> None:
        self.fh.close()


def utc_stamp() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def run_text(cmd: list[str], workdir: Path) -> str:
    try:
        proc = subprocess.run(cmd, cwd=workdir, text=True, capture_output=True, check=False, timeout=10)
    except (OSError, subprocess.TimeoutExpired):
        return ""
    return proc.stdout.strip()


def repo_root() -> Path:
    return Path(__file__).resolve().parents[1]


def redact_text(text: str, root: Path) -> str:
    redacted = text.replace(str(root), "<repo>")
    home = str(Path.home())
    if home and home != "/":
        redacted = redacted.replace(home, "<home>")
    return redacted


def display_path(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return redact_text(str(path), root)


def host_info(root: Path) -> dict[str, str]:
    info = {
        "kernel": f"{platform.system()} {platform.release()}",
        "machine": platform.machine(),
        "python": platform.python_version(),
        "cpu_count": str(os.cpu_count() or 0),
        "go": run_text(["go", "version"], root) or "not found",
    }
    try:
        with open("/proc/cpuinfo", encoding="utf-8") as fh:
            for line in fh:
                if line.startswith("model name"):
                    info["cpu"] = line.split(":", 1)[1].strip()
                    break
    except OSError:
        info["cpu"] = platform.processor() or "unknown"
    try:
        with open("/proc/meminfo", encoding="utf-8") as fh:
            for line in fh:
                if line.startswith("MemTotal"):
                    kib = int(line.split()[1])
                    info["memory"] = f"{kib // 1024 // 1024} GiB"
                    break
    except OSError:
        info["memory"] = "unknown"
    return info


def parse_csv(raw: str) -> list[str]:
    return [part.strip() for part in raw.split(",") if part.strip()]


def parse_bench_output(stdout: str) -> list[BenchSample]:
    samples: list[BenchSample] = []
    for line in stdout.splitlines():
        match = BENCH_LINE_RE.match(line.strip())
        if not match:
            continue
        metrics: dict[str, float] = {}
        for metric in METRIC_RE.finditer(line):
            label = metric.group(2)
            if label == "ns/op":
                continue
            metrics[label] = float(metric.group(1))
        samples.append(BenchSample(match.group("name"), float(match.group("ns")), metrics))
    return samples


def shape_bits_radix(shape: str) -> tuple[int, int]:
    bits, radix = shape.split("_")
    return int(bits.replace("bit", "")), int(radix.replace("r", ""))


def shape_width(shape: str) -> int:
    bits, radix = shape_bits_radix(shape)
    return bits // int(math.log2(radix))


def is_prime(n: int) -> bool:
    if n < 2:
        return False
    if n % 2 == 0:
        return n == 2
    d = 3
    while d * d <= n:
        if n % d == 0:
            return False
        d += 2
    return True


def previous_prime(n: int) -> int:
    for p in range(n, 1, -1):
        if is_prime(p):
            return p
    return 2


def basic_modulus(encoding: str, log2t: int) -> int:
    t = 1 << log2t
    if encoding.upper() == "LBRU":
        return previous_prime(t)
    return t


def basic_feature_slots(encoding: str, log2t: int, arity: int) -> int:
    t = basic_modulus(encoding, log2t)
    if arity <= 1:
        return t - 1
    return t**arity


def basic_shape(encoding: str, log2t: int) -> str:
    return f"{encoding}_log2t{log2t}"


def basic_log2ts_for_op(args: argparse.Namespace, op: str) -> list[int]:
    override = {
        "native": args.basic_native_log2t,
        "unary_lut": args.basic_unary_log2t,
        "binary_lut": args.basic_binary_log2t,
        "four_lut": args.basic_four_log2t,
        "clean": args.basic_clean_log2t,
    }.get(op, "")
    raw = override or args.basic_log2t
    return [int(part) for part in parse_csv(raw)]


def chain_len(op: str, width: int) -> int:
    if op == "Clean":
        return 5
    rounds = max(1, math.ceil(math.log2(width)))
    if op in {"Add", "Sub"}:
        return 4 + rounds + 1
    return 3 + rounds + 1


def radix_params(length: int) -> tuple[str, str, int]:
    log_q = f"[60, 40 x {length - 1}]"
    log_p = "[60]"
    return log_q, log_p, 60 + 40 * (length - 1) + 60


def crt_params(levels: float | None) -> tuple[str, str, int]:
    consumed = int(levels) if levels is not None else 1
    log_q = f"[55, 40 x {consumed}]"
    log_p = "[60]"
    return log_q, log_p, 55 + 40 * consumed + 60


def max_secure_log_qp(logN: int) -> int | None:
    return SECURITY_LIMITS.get(logN, {}).get(SECRET_H)


def secure(logN: int, log_qp: float) -> bool:
    max_qp = max_secure_log_qp(logN)
    return max_qp is not None and log_qp <= max_qp


def measured_log_qp(result: BenchResult, fallback: int) -> float:
    return result.metric("logQP") or float(fallback)


def fmt_ms(ns: float) -> str:
    if math.isnan(ns):
        return "no data"
    ms = ns / 1_000_000
    if ms >= 1000:
        return f"{ms / 1000:.2f} s"
    if ms >= 1:
        return f"{ms:.1f} ms"
    return f"{ms * 1000:.0f} us"


def output_tail(text: str, limit: int = 4000) -> str:
    if len(text) <= limit:
        return text
    return text[-limit:]


def redacted_output_tail(text: str, root: Path, limit: int = 4000) -> str:
    return output_tail(redact_text(text, root), limit)


def bench_wrapper(args: argparse.Namespace, threads: int) -> list[str]:
    if threads == 1 and args.bench_core is not None:
        return ["taskset", "-c", str(args.bench_core)]
    return []


def multi_thread_count(args: argparse.Namespace) -> int:
    return os.cpu_count() or 1


def run_go_bench(
    args: argparse.Namespace,
    root: Path,
    heartbeat: Heartbeat,
    package: str,
    bench_regex: str,
    count: int,
    threads: int,
    timeout_s: int,
    env_extra: dict[str, str],
    gocache: Path,
    meta: dict[str, object],
) -> BenchResult:
    env = os.environ.copy()
    env["GOCACHE"] = str(gocache)
    env["GOMAXPROCS"] = str(threads)
    env.update(env_extra)
    cmd = bench_wrapper(args, threads) + [
        "go",
        "test",
        package,
        "-run",
        "^$",
        "-bench",
        bench_regex,
        "-benchtime",
        "1x",
        "-count",
        str(count),
        "-timeout",
        f"{timeout_s}s",
    ]
    tracked_keys = set(env_extra) | {"GOCACHE", "GOMAXPROCS"}
    tracked_env = {}
    for key in sorted(tracked_keys):
        if key in env:
            tracked_env[key] = redact_text(env[key], root)
    heartbeat.write("benchmark_start", command=cmd, env=tracked_env, **meta)
    start = time.monotonic()
    try:
        proc = subprocess.run(cmd, cwd=root, env=env, text=True, capture_output=True, check=False, timeout=timeout_s)
        combined = proc.stdout + ("\n--- STDERR ---\n" + proc.stderr if proc.stderr else "")
        returncode = proc.returncode
    except subprocess.TimeoutExpired as exc:
        stdout = exc.stdout if isinstance(exc.stdout, str) else ""
        stderr = exc.stderr if isinstance(exc.stderr, str) else ""
        combined = stdout + ("\n--- STDERR ---\n" + stderr if stderr else "")
        returncode = -9
    duration = time.monotonic() - start
    samples = parse_bench_output(combined)
    correctness_ok = all(sample.metrics.get("correct", 1) >= 0.5 for sample in samples)
    ok = returncode == 0 and bool(samples) and correctness_ok
    heartbeat.write(
        "benchmark_finish",
        ok=ok,
        returncode=returncode,
        duration_s=round(duration, 3),
        samples=[asdict(sample) for sample in samples],
        correctness_ok=correctness_ok,
        output_tail=redacted_output_tail(combined, root),
        **meta,
    )
    return BenchResult(
        target=str(meta["target"]),
        op=str(meta["op"]),
        shape=str(meta["shape"]),
        logN=int(meta["logN"]),
        mode=str(meta["mode"]),
        bench_regex=bench_regex,
        package=package,
        ok=ok,
        returncode=returncode,
        duration_s=duration,
        samples=samples,
        output_tail=redacted_output_tail(combined, root),
        layout=str(meta.get("layout", "packed")),
    )


def radix_name(op: str, logN: int, shape: str) -> str:
    return f"Benchmark{op}_LogN{logN}_{shape.replace('bit_r', 'R')}"


def run_radix(args: argparse.Namespace, root: Path, heartbeat: Heartbeat, mode: str, logN: int) -> list[BenchResult]:
    threads = 1 if mode == "single" else multi_thread_count(args)
    shapes = parse_csv(args.shapes) or RADIX_SHAPES
    ops = parse_csv(args.ops) or RADIX_OPS
    results: list[BenchResult] = []
    for op in ops:
        for shape in shapes:
            name = radix_name(op, logN, shape)
            print(f"radix {mode} LogN={logN} {op} {shape}", flush=True)
            results.append(
                run_go_bench(
                    args,
                    root,
                    heartbeat,
                    "./character-encoding/radix",
                    f"^{name}$",
                    args.count,
                    threads,
                    args.timeout,
                    {},
                    args.gocache,
                    {"target": "radix", "op": op, "shape": shape, "logN": logN, "mode": mode},
                )
            )
    return results


def run_crt(args: argparse.Namespace, root: Path, heartbeat: Heartbeat, mode: str, logN: int) -> list[BenchResult]:
    threads = 1 if mode == "single" else multi_thread_count(args)
    submode = "sequential" if mode == "single" else "parallel"
    ops = parse_csv(args.crt_ops)
    bits_list = [int(raw) for raw in parse_csv(args.crt_bits)]
    results: list[BenchResult] = []
    for op in ops:
        label, prefix, levels = CRT_OPS[op]
        env = {"CRT_BENCH_LOGN": str(logN), "CRT_BENCH_WORKERS": str(threads), "CRT_BENCH_LEVELS": str(levels)}
        for bits in bits_list:
            regex = f"{prefix}/{bits}bits/{submode}$"
            print(f"crt {mode} LogN={logN} {label} {bits} bits", flush=True)
            results.append(
                run_go_bench(
                    args,
                    root,
                    heartbeat,
                    "./character-encoding/crt",
                    regex,
                    args.count,
                    threads,
                    args.timeout,
                    env,
                    args.gocache,
                    {"target": "crt", "op": label, "shape": f"{bits}bits", "logN": logN, "mode": mode},
                )
            )
    return results


def run_basic(args: argparse.Namespace, root: Path, heartbeat: Heartbeat, mode: str) -> list[BenchResult]:
    threads = 1 if mode == "single" else multi_thread_count(args)
    submode = "sequential"
    layout = args.layout
    op_table = BASIC_OPS if layout == "packed" else BASIC_OPS_SPLIT
    ops = parse_csv(args.basic_ops) or list(BASIC_OPS)
    encodings = [enc.upper() for enc in (parse_csv(args.basic_encodings) or BASIC_ENCODINGS)]
    results: list[BenchResult] = []
    for op in ops:
        if op not in op_table:
            print(f"skip basic {mode} {op}: not supported in --layout={layout}", flush=True)
            continue
        label, prefix, _levels, arity = op_table[op]
        log2ts = basic_log2ts_for_op(args, op)
        op_encodings = encodings
        if layout == "split":
            override_encs = BASIC_SPLIT_OP_ENCODINGS.get(op)
            if override_encs is not None:
                op_encodings = [enc for enc in encodings if enc in override_encs]
                if not op_encodings:
                    print(
                        f"skip basic {mode} {label}: split {label} is restricted to "
                        f"{override_encs}, none selected via --basic-encodings",
                        flush=True,
                    )
                    continue
        for encoding in op_encodings:
            for log2t in log2ts:
                op_max_log2t = BASIC_SPLIT_MAX_LOG2T
                if layout == "split":
                    cap = BASIC_SPLIT_OP_MAX_LOG2T.get(op)
                    if cap is not None and cap < op_max_log2t:
                        op_max_log2t = cap
                if layout == "split" and log2t > op_max_log2t:
                    print(
                        f"skip basic {mode} {label} {encoding} log2t={log2t}: "
                        f"split {label} capped at log2t <= {op_max_log2t}",
                        flush=True,
                    )
                    continue
                t = basic_modulus(encoding, log2t)
                feature_slots = basic_feature_slots(encoding, log2t, arity)
                if feature_slots > BASIC_MAX_SLOTS:
                    print(
                        f"skip basic {mode} {label} {encoding} log2t={log2t}: "
                        f"{feature_slots} tensor slots exceed {BASIC_MAX_SLOTS}",
                        flush=True,
                    )
                    continue
                if (op == "unary_lut" or (op == "clean" and encoding not in {"IND", "WH"})) and t > args.basic_max_lut_size:
                    print(
                        f"skip basic {mode} {label} {encoding} log2t={log2t}: "
                        f"t={t} exceeds --basic-max-lut-size={args.basic_max_lut_size}",
                        flush=True,
                    )
                    continue
                if arity > 1 and feature_slots > args.basic_max_tensor_size:
                    print(
                        f"skip basic {mode} {label} {encoding} log2t={log2t}: "
                        f"{feature_slots} tensor basis exceeds --basic-max-tensor-size={args.basic_max_tensor_size}",
                        flush=True,
                    )
                    continue
                regex = f"{prefix}/{encoding}/log2t{log2t}/{submode}$"
                env = {
                    "BASIC_BENCH_ENCODING": encoding,
                    "BASIC_BENCH_LOG2T": str(log2t),
                }
                print(f"basic {mode} {layout} LogN={BASIC_LOGN} {label} {encoding} log2t={log2t}", flush=True)
                results.append(
                    run_go_bench(
                        args,
                        root,
                        heartbeat,
                        "./character-encoding/lut",
                        regex,
                        args.count,
                        threads,
                        args.timeout,
                        env,
                        args.gocache,
                        {
                            "target": "basic",
                            "op": label,
                            "shape": basic_shape(encoding, log2t),
                            "logN": BASIC_LOGN,
                            "mode": mode,
                            "layout": layout,
                        },
                    )
                )
    return results


def markdown(results: list[BenchResult], host: dict[str, str], started_at: str, finished_at: str) -> str:
    layouts = sorted({result.layout for result in results}) or ["packed"]
    lines = [
        "# Benchmark Summary",
        "",
        f"- Started: {started_at}",
        f"- Finished: {finished_at}",
        f"- Layout(s): {', '.join(layouts)}",
        f"- CPU: {host.get('cpu', 'unknown')} ({host.get('cpu_count', '0')} logical cores)",
        f"- Memory: {host.get('memory', 'unknown')}",
        f"- Kernel: {host.get('kernel', 'unknown')}",
        f"- Go: {host.get('go', 'unknown')}",
        "",
    ]
    append_results_tables(lines, "Single Thread", [result for result in results if result.mode == "single"])
    append_results_tables(lines, "Multi Thread", [result for result in results if result.mode == "multi"])
    other_modes = [result for result in results if result.mode not in {"single", "multi"}]
    if other_modes:
        append_results_tables(lines, "Other", other_modes)
    append_parameter_tables(lines, results)
    return "\n".join(lines) + "\n"


def result_groups(results: list[BenchResult]) -> list[tuple[str, list[BenchResult]]]:
    groups: list[tuple[str, list[BenchResult]]] = []
    for target in ("basic", "crt", "radix"):
        target_results = [result for result in results if result.target == target]
        if not target_results:
            continue
        if target == "basic":
            seen_labels: set[str] = set()
            for op_table in (BASIC_OPS, BASIC_OPS_SPLIT):
                for op in op_table.values():
                    label = op[0]
                    if label in seen_labels:
                        continue
                    seen_labels.add(label)
                    op_results = [result for result in target_results if result.op == label]
                    if op_results:
                        groups.append((f"Basic {BASIC_OP_LABELS.get(label, label)}", op_results))
        else:
            groups.append((target.upper() if target == "crt" else target.capitalize(), target_results))
    for target in sorted({result.target for result in results} - {"basic", "crt", "radix"}):
        groups.append((target.capitalize(), [result for result in results if result.target == target]))
    return groups


def append_parameter_tables(lines: list[str], results: list[BenchResult]) -> None:
    seen: set[tuple[str, int, str, str]] = set()
    for title, group in result_groups(results):
        rows: list[str] = []
        for result in group:
            if result.target == "radix":
                width = shape_width(result.shape)
                log_q, log_p, fallback_qp = radix_params(chain_len(result.op, width))
                qp = measured_log_qp(result, fallback_qp)
                key = (result.target, result.logN, result.op, result.shape)
                label = f"{result.op} {result.shape}"
            elif result.target == "basic":
                log_q, log_p, fallback_qp = crt_params(result.metric("levels/op"))
                qp = measured_log_qp(result, fallback_qp)
                key = (result.target, result.logN, result.op, result.shape)
                label = f"basic {result.op} {result.shape}"
            else:
                log_q, log_p, fallback_qp = crt_params(result.metric("levels/op"))
                qp = measured_log_qp(result, fallback_qp)
                key = (result.target, result.logN, result.op, "crt")
                label = f"CRT {result.op}"
            if key in seen:
                continue
            seen.add(key)
            max_qp = max_secure_log_qp(result.logN)
            rows.append(f"| {result.target} | {result.logN} | {label} | `{log_q}` | `{log_p}` | {qp:.1f} | {max_qp if max_qp is not None else 'unknown'} | {'yes' if secure(result.logN, qp) else 'no'} |")
        if not rows:
            continue
        lines += [
            "",
            f"## {title} Parameter Check",
            "",
            "| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |",
            "| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |",
        ]
        lines.extend(rows)


def append_results_tables(lines: list[str], mode_title: str, results: list[BenchResult]) -> None:
    for title, group in result_groups(results):
        layouts = sorted({result.layout for result in group})
        if len(layouts) <= 1:
            append_results_table(lines, f"{mode_title} {title} Results", group)
            continue
        for layout in layouts:
            subset = [result for result in group if result.layout == layout]
            if subset:
                append_results_table(lines, f"{mode_title} {layout.capitalize()} {title} Results", subset)


def precision_text(result: BenchResult, metric: str, saturated_metric: str) -> str:
    precision = result.metric(metric)
    saturated = (result.metric(saturated_metric) or 0.0) >= 0.5
    if precision is None:
        return ""
    if saturated:
        return f">={precision:.1f} bits"
    return f"{precision:.1f} bits"


def append_results_table(lines: list[str], title: str, results: list[BenchResult]) -> None:
    if not results:
        return
    is_basic = all(result.target == "basic" for result in results)
    has_output_level = any(result.metric("output_level") is not None for result in results)
    lines += [f"## {title}", ""]
    if is_basic:
        header = "| Target | LogN | Operation | Shape | Levels"
        align = "| :----- | ---: | :-------- | :---- | -----:"
        if has_output_level:
            header += " | Remaining"
            align += " | --------:"
        header += " | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |"
        align += " | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |"
        lines += [header, align]
    else:
        header = "| Target | LogN | Operation | Shape | Levels"
        align = "| :----- | ---: | :-------- | :---- | -----:"
        if has_output_level:
            header += " | Remaining"
            align += " | --------:"
        header += " | Samples | Mean | Stddev | Amortized | Correct | Max Error |"
        align += " | ------: | ---: | -----: | --------: | :------ | --------: |"
        lines += [header, align]
    for result in results:
        divisor = result.metric("words/op") or result.metric("packed_cts") or 1.0
        amortized = result.mean_ns() / divisor if divisor else result.mean_ns()
        levels = result.metric("levels/op")
        level_text = f"{levels:.0f}" if levels is not None else ""
        output_level = result.metric("output_level")
        output_level_text = f"{output_level:.0f}" if output_level is not None else ""
        correct_text = "yes" if result.correct() else "no"
        max_error = result.metric("max_error")
        max_error_text = "" if max_error is None else f"{max_error:.0f}"
        prefix = (
            f"| {result.target} | {result.logN} | {result.op} | {result.shape} | "
            f"{level_text}"
        )
        if has_output_level:
            prefix += f" | {output_level_text}"
        prefix += (
            f" | {len(result.samples)} | {fmt_ms(result.mean_ns())} | {fmt_ms(result.stddev_ns())} | "
            f"{fmt_ms(amortized)}"
        )
        if is_basic:
            input_precision_text = precision_text(result, "input_precision_bits", "input_precision_saturated")
            output_precision_text = precision_text(result, "precision_bits", "precision_saturated")
            delta = result.metric("precision_delta_bits")
            if delta is not None:
                gain_loss_text = f"{delta:+.1f} bits" if delta != 0 else "0.0 bits"
            else:
                gain_loss_text = ""
            lines.append(
                f"{prefix} | {input_precision_text} | {output_precision_text} | "
                f"{gain_loss_text} | {correct_text} | {max_error_text} |"
            )
        else:
            lines.append(f"{prefix} | {correct_text} | {max_error_text} |")
    lines.append("")


def result_secure(result: BenchResult) -> bool:
    if result.target == "radix":
        _, _, fallback_qp = radix_params(chain_len(result.op, shape_width(result.shape)))
    elif result.target == "basic":
        _, _, fallback_qp = crt_params(result.metric("levels/op"))
    else:
        _, _, fallback_qp = crt_params(result.metric("levels/op"))
    qp = measured_log_qp(result, fallback_qp)
    return secure(result.logN, qp)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run the character-encoding benchmarks and write heartbeat details.")
    parser.add_argument("--target", choices=["radix", "crt", "basic", "both", "all"], default="both")
    parser.add_argument("--mode", choices=["single", "multi", "both"], default="both")
    parser.add_argument(
        "--layout",
        choices=["packed", "split", "both"],
        default="packed",
        help="Ciphertext layout. 'packed' keeps every block coordinate in one ciphertext (default); 'split' uses one ciphertext per block coordinate (basic ops + standard<->split converters); 'both' runs both layouts in sequence so a single invocation covers packed basic/CRT/radix plus split basic and converters.",
    )
    parser.add_argument("--logn", action="append", type=int, help="CRT LogN override. Radix benchmarks always use LogN=16.")
    parser.add_argument("--count", type=int, default=10)
    parser.add_argument("--timeout", type=int, default=1800)
    parser.add_argument("--ops", default="")
    parser.add_argument("--shapes", default="")
    parser.add_argument("--crt-ops", default="add,sub,mul_lbru,bru_to_lbru,lbru_to_bru,clean")
    parser.add_argument("--crt-bits", default="64,256")
    parser.add_argument("--basic-ops", default="native,unary_lut,binary_lut,four_lut,clean,to_standard,from_standard")
    parser.add_argument("--basic-encodings", default="BRU,LBRU,WH,IND")
    parser.add_argument("--basic-log2t", default=",".join(str(k) for k in BASIC_LOG2T))
    parser.add_argument("--basic-native-log2t", default=",".join(str(k) for k in BASIC_NATIVE_LOG2T), help="Log2T sweep override for basic native benchmarks")
    parser.add_argument("--basic-unary-log2t", default=",".join(str(k) for k in BASIC_UNARY_LOG2T), help="Log2T sweep override for basic unary LUT benchmarks")
    parser.add_argument("--basic-binary-log2t", default=",".join(str(k) for k in BASIC_BINARY_LOG2T), help="Log2T sweep override for basic bivariate LUT benchmarks")
    parser.add_argument("--basic-four-log2t", default=",".join(str(k) for k in BASIC_FOUR_LOG2T), help="Log2T sweep override for basic 4-variate LUT benchmarks")
    parser.add_argument("--basic-clean-log2t", default="", help="Log2T sweep override for basic cleaning benchmarks")
    parser.add_argument(
        "--basic-max-tensor-size",
        type=int,
        default=4096,
        help="largest t^arity tensor basis to compile for bivariate/four-variate basic LUTs",
    )
    parser.add_argument(
        "--basic-max-lut-size",
        type=int,
        default=1024,
        help="largest t to compile for unary LUT and cleaning basic benchmarks",
    )
    parser.add_argument("--out-dir", type=Path, default=Path("bench_results"))
    parser.add_argument("--gocache", type=Path, default=Path("/tmp/go-build-character-encoding"))
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--bench-core", type=int, default=None, help="Physical core to pin single-thread benchmarks to via taskset -c. Unset by default (no pinning). Ignored for multi-thread benchmarks.")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    root = repo_root()
    if shutil.which("go") is None:
        print("go not found on PATH", file=sys.stderr)
        return 2
    unknown_crt = [op for op in parse_csv(args.crt_ops) if op not in CRT_OPS]
    if unknown_crt:
        print(f"unknown CRT op(s): {', '.join(unknown_crt)}", file=sys.stderr)
        return 2
    known_basic = set(BASIC_OPS) | set(BASIC_OPS_SPLIT)
    unknown_basic = [op for op in parse_csv(args.basic_ops) if op not in known_basic]
    if unknown_basic:
        print(f"unknown basic op(s): {', '.join(unknown_basic)}", file=sys.stderr)
        return 2
    unknown_basic_encodings = [enc for enc in parse_csv(args.basic_encodings) if enc.upper() not in BASIC_ENCODINGS]
    if unknown_basic_encodings:
        print(f"unknown basic encoding(s): {', '.join(unknown_basic_encodings)}", file=sys.stderr)
        return 2

    modes = ["single", "multi"] if args.mode == "both" else [args.mode]
    radix_logns = [RADIX_LOGN]
    crt_logns = sorted(set(args.logn)) if args.logn else [CRT_LOGN]

    stamp = time.strftime("%Y%m%dT%H%M%SZ", time.gmtime())
    out_dir = (root / args.out_dir / stamp).resolve()

    host = host_info(root)
    started_at = utc_stamp()
    print(f"output: {display_path(root, out_dir)}", flush=True)
    print(f"host: {host.get('cpu', 'unknown')} | {host.get('cpu_count', '0')} cores | {host.get('memory', 'unknown')}", flush=True)
    if args.dry_run:
        return 0

    args.gocache.mkdir(parents=True, exist_ok=True)
    out_dir.mkdir(parents=True, exist_ok=True)

    heartbeat = Heartbeat(out_dir / "heartbeat.jsonl")
    heartbeat.write("run_start", args={k: redact_text(str(v), root) for k, v in vars(args).items()}, host=host)
    results: list[BenchResult] = []
    layouts = ["packed", "split"] if args.layout == "both" else [args.layout]
    original_layout = args.layout
    try:
        for mode in modes:
            for layout in layouts:
                args.layout = layout
                if args.target in {"radix", "both", "all"}:
                    if layout == "split":
                        print("skip radix: --layout=split is only implemented for basic targets", flush=True)
                    else:
                        for logN in radix_logns:
                            results.extend(run_radix(args, root, heartbeat, mode, logN))
                if args.target in {"crt", "both", "all"}:
                    if layout == "split":
                        print("skip crt: --layout=split is only implemented for basic targets", flush=True)
                    else:
                        for logN in crt_logns:
                            results.extend(run_crt(args, root, heartbeat, mode, logN))
                if args.target in {"basic", "all"}:
                    results.extend(run_basic(args, root, heartbeat, mode))
        args.layout = original_layout
    finally:
        finished_at = utc_stamp()
        heartbeat.write("run_finish", result_count=len(results), failed=sum(1 for r in results if not r.ok))
        heartbeat.close()

    (out_dir / "results.json").write_text(json.dumps([asdict(result) for result in results], indent=2), encoding="utf-8")
    (out_dir / "summary.md").write_text(markdown(results, host, started_at, finished_at), encoding="utf-8")
    failures = [result for result in results if not result.ok]
    insecure_results = [result for result in results if not result_secure(result)]
    if failures:
        print(f"{len(failures)} benchmark invocation(s) failed; see {display_path(root, out_dir / 'heartbeat.jsonl')}", file=sys.stderr)
        return 1
    if insecure_results:
        print(f"{len(insecure_results)} benchmark invocation(s) used insecure parameters; see {display_path(root, out_dir / 'summary.md')}", file=sys.stderr)
        return 1
    print(f"wrote {display_path(root, out_dir / 'summary.md')}")
    print(f"wrote {display_path(root, out_dir / 'results.json')}")
    print(f"wrote {display_path(root, out_dir / 'heartbeat.jsonl')}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
