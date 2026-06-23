#!/usr/bin/env python3
"""Run and evaluate the Sakana Fugu Go coding benchmark."""

from __future__ import annotations

import difflib
import json
import os
import re
import shutil
import statistics
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
from datetime import date
from pathlib import Path

ROOT = Path(__file__).resolve().parent
FIXTURES = ROOT / "fixtures"
PROMPTS = ROOT / "prompts"
RUN_NUMBER = int(os.getenv("RESEARCH_RUN", "1"))
RESPONSES = ROOT / f"responses-run{RUN_NUMBER}"
LOGS = ROOT / f"logs-run{RUN_NUMBER}"
API_URL = "https://api.sakana.ai/v1/chat/completions"
MODEL = "fugu"
TASKS = ("topk", "pmap", "options")
SYSTEM_PROMPT = (
    "You are an expert Go engineer. Follow the requested API exactly. "
    "Output source code only."
)


def safe(value: str) -> str:
    return re.sub(r"[^A-Za-z0-9_.-]+", "__", value)


def prepare() -> None:
    RESPONSES.mkdir(parents=True, exist_ok=True)
    LOGS.mkdir(parents=True, exist_ok=True)


def resolve_token() -> str:
    token = os.getenv("SAKANA_AI_API_KEY")
    if not token:
        raise RuntimeError("SAKANA_AI_API_KEY is not set")
    return token


def extract_code(text: str) -> str:
    blocks = re.findall(r"```(?:go)?\s*(.*?)```", text, re.DOTALL | re.IGNORECASE)
    code = max(blocks, key=len) if blocks else text
    if (start := code.find("package ")) >= 0:
        code = code[start:]
    return code.strip() + "\n"


def post_json(body: dict, token: str) -> tuple[int, dict]:
    request = urllib.request.Request(
        API_URL,
        data=json.dumps(body).encode(),
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=900) as response:
            return response.status, json.load(response)
    except urllib.error.HTTPError as error:
        return error.code, json.loads(error.read().decode())


def generate() -> None:
    prepare()
    token = resolve_token()
    requests: list[dict] = []
    for index, task in enumerate(TASKS, 1):
        directory = RESPONSES / safe(MODEL) / task
        directory.mkdir(parents=True, exist_ok=True)
        source = directory / "solution.go"
        if source.exists():
            print(f"[{index}/3] cached {MODEL} :: {task}", flush=True)
            continue
        print(f"[{index}/3] {MODEL} :: {task}", flush=True)
        prompt = (PROMPTS / f"{task}.md").read_text(encoding="utf-8").strip()
        body = {
            "model": MODEL,
            "messages": [
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": prompt},
            ],
            "temperature": 0,
            "max_completion_tokens": 10000,
            "stream": False,
        }
        started = time.perf_counter()
        record: dict = {"model": MODEL, "task": task}
        try:
            status, payload = post_json(body, token)
            record["elapsed_seconds"] = time.perf_counter() - started
            record["status"] = status
            record["usage"] = payload.get("usage", {})
            if status >= 400:
                record["error"] = payload
            else:
                choice = payload["choices"][0]
                message = choice["message"]
                content = message.get("content")
                record["finish_reason"] = choice.get("finish_reason")
                record["message_keys"] = list(message)
                (directory / "response.json").write_text(
                    json.dumps(payload, ensure_ascii=False, indent=2),
                    encoding="utf-8",
                )
                if isinstance(content, str) and content.strip():
                    source.write_text(extract_code(content), encoding="utf-8")
                else:
                    record["error"] = "empty content"
        except Exception as error:  # noqa: BLE001
            record["elapsed_seconds"] = time.perf_counter() - started
            record["error"] = repr(error)
        if "error" in record:
            (directory / "error.json").write_text(
                json.dumps(record, ensure_ascii=False, indent=2), encoding="utf-8"
            )
        requests.append(record)
        print(
            f"  {record['elapsed_seconds']:.2f}s {record.get('usage', {})}",
            flush=True,
        )
    (LOGS / "requests.json").write_text(
        json.dumps(requests, ensure_ascii=False, indent=2), encoding="utf-8"
    )


def command(args: list[str], cwd: Path, timeout: int = 60) -> dict:
    try:
        result = subprocess.run(
            args, cwd=cwd, capture_output=True, text=True, timeout=timeout
        )
        return {
            "ok": result.returncode == 0,
            "returncode": result.returncode,
            "stdout": result.stdout,
            "stderr": result.stderr,
        }
    except subprocess.TimeoutExpired as error:
        stdout = error.stdout.decode() if isinstance(error.stdout, bytes) else error.stdout
        stderr = error.stderr.decode() if isinstance(error.stderr, bytes) else error.stderr
        return {
            "ok": False,
            "returncode": 124,
            "stdout": stdout or "",
            "stderr": (stderr or "") + f"\ntimeout after {timeout}s",
        }


def test_results(output: str) -> dict:
    actions: dict[str, str] = {}
    for line in output.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        name = event.get("Test")
        if name and event.get("Action") in {"pass", "fail"}:
            actions[name] = event["Action"]
    leaves = {
        name
        for name in actions
        if not any(other.startswith(name + "/") for other in actions)
    }
    passed = sorted(name for name in leaves if actions[name] == "pass")
    failed = sorted(name for name in leaves if actions[name] == "fail")
    return {"passed": passed, "failed": failed, "passed_count": len(passed)}


def benchmark_metrics(output: str) -> dict:
    matches = re.findall(
        r"^(Benchmark\S+)-\d+\s+\d+\s+([\d.]+)\s+ns/op"
        r"(?:\s+([\d.]+)\s+B/op)?(?:\s+([\d.]+)\s+allocs/op)?",
        output,
        re.MULTILINE,
    )
    if not matches:
        return {}
    name, ns, bytes_op, allocs = matches[-1]
    return {
        "name": name,
        "ns_per_op": float(ns),
        "bytes_per_op": float(bytes_op or 0),
        "allocs_per_op": float(allocs or 0),
    }


def evaluate_one(task: str, requests: list[dict]) -> dict:
    directory = RESPONSES / safe(MODEL) / task
    source = directory / "solution.go"
    result: dict = {"model": MODEL, "task": task, "available": source.exists()}
    metadata = directory / "response.json"
    error = directory / "error.json"
    if metadata.exists():
        payload = json.loads(metadata.read_text(encoding="utf-8"))
        result["usage"] = payload.get("usage", {})
        result["finish_reason"] = payload["choices"][0].get("finish_reason")
    if error.exists():
        result["generation_error"] = json.loads(error.read_text(encoding="utf-8"))
    for request in requests:
        if request["task"] == task:
            result["elapsed_seconds"] = request["elapsed_seconds"]
            result.setdefault("usage", request.get("usage", {}))
            break
    if not source.exists():
        return result

    code = source.read_text(encoding="utf-8")
    result["loc"] = sum(
        1
        for line in code.splitlines()
        if line.strip() and not line.lstrip().startswith("//")
    )
    with tempfile.TemporaryDirectory(prefix=f"fugu-go-{task}-") as temp:
        work = Path(temp)
        shutil.copytree(FIXTURES / task, work, dirs_exist_ok=True)
        shutil.copy2(source, work / "solution.go")
        formatted = command(["gofmt", "-d", "solution.go"], work)
        result["gofmt_clean"] = formatted["stdout"] == ""
        result["gofmt_diff"] = formatted["stdout"]
        result["build"] = command(["go", "build", "./..."], work)
        if not result["build"]["ok"]:
            return result
        test = command(
            ["go", "test", "-json", "-count=1", "-timeout=15s", "./..."],
            work,
            timeout=20,
        )
        result["test"] = {**test, **test_results(test["stdout"])}
        result["vet"] = command(["go", "vet", "./..."], work)
        result["lint"] = command(
            [
                "golangci-lint",
                "run",
                "--no-config",
                "--default=standard",
                "--timeout=30s",
                "--output.text.colors=false",
            ],
            work,
            timeout=40,
        )
        before = (work / "solution.go").read_text(encoding="utf-8")
        fix = command(["go", "fix", "./..."], work)
        after = (work / "solution.go").read_text(encoding="utf-8")
        diff = "\n".join(
            difflib.unified_diff(
                before.splitlines(),
                after.splitlines(),
                fromfile="before.go",
                tofile="after.go",
                lineterm="",
            )
        )
        result["go_fix"] = {**fix, "changed": before != after, "diff": diff}
        if test["ok"]:
            result["race"] = command(
                ["go", "test", "-race", "-count=1", "-timeout=30s", "./..."],
                work,
                timeout=40,
            )
        else:
            result["race"] = {"ok": False, "stderr": "skipped: tests failed"}
        if test["ok"] and result["race"]["ok"]:
            bench = command(
                ["go", "test", "-run=^$", "-bench=.", "-benchmem", "-count=3", "./..."],
                work,
                timeout=240,
            )
            result["benchmark"] = {
                **bench,
                "metrics": benchmark_metrics(bench["stdout"]),
            }
    return result


def evaluate() -> None:
    requests = json.loads((LOGS / "requests.json").read_text(encoding="utf-8"))
    rows = []
    for task in TASKS:
        print(f"validate {MODEL} :: {task}", flush=True)
        rows.append(evaluate_one(task, requests))
        (LOGS / "evaluation.json").write_text(
            json.dumps(rows, ensure_ascii=False, indent=2), encoding="utf-8"
        )


def repeat_reliability() -> None:
    source = RESPONSES / safe(MODEL) / "pmap" / "solution.go"
    outcome = {"ok": False, "stderr": "skipped: pmap source is unavailable"}
    if source.exists():
        with tempfile.TemporaryDirectory(prefix="repeat-fugu-pmap-") as temp:
            work = Path(temp)
            shutil.copytree(FIXTURES / "pmap", work, dirs_exist_ok=True)
            shutil.copy2(source, work / "solution.go")
            outcome = command(
                ["go", "test", "-race", "-v", "-count=20", "-timeout=90s", "./..."],
                work,
                timeout=100,
            )
    (LOGS / f"repeat-{safe(MODEL)}.log").write_text(
        outcome.get("stdout", "") + outcome.get("stderr", ""), encoding="utf-8"
    )
    (LOGS / "repeat-summary.json").write_text(
        json.dumps({MODEL: outcome}, ensure_ascii=False, indent=2), encoding="utf-8"
    )


def summarize() -> None:
    items = json.loads((LOGS / "evaluation.json").read_text(encoding="utf-8"))
    passed = sum(item.get("test", {}).get("passed_count", 0) for item in items)
    pmap = next(item for item in items if item["task"] == "pmap")
    elapsed = [item.get("elapsed_seconds", 900.0) for item in items]
    modern_tasks = sum(
        bool(item.get("build", {}).get("ok"))
        and item.get("go_fix", {}).get("changed") is False
        for item in items
    )
    summary = {
        "functionality": passed / 19,
        "effectiveness": sum(
            bool(item.get("test", {}).get("ok")) for item in items
        )
        / 3,
        "reliability": (
            pmap.get("test", {}).get("passed_count", 0) / 6
            + sum(bool(item.get("race", {}).get("ok")) for item in items) / 3
        )
        / 2,
        "usability": (
            sum(bool(item.get("gofmt_clean")) for item in items) / 3
            + sum(bool(item.get("vet", {}).get("ok")) for item in items) / 3
            + sum(bool(item.get("lint", {}).get("ok")) for item in items) / 3
        )
        / 3,
        "response_seconds": statistics.median(elapsed),
        "modern": modern_tasks / 3,
        "passed": passed,
        "requirements": 19,
        "pmap_passed": pmap.get("test", {}).get("passed_count", 0),
        "race_tasks": sum(bool(item.get("race", {}).get("ok")) for item in items),
        "modern_tasks": modern_tasks,
        "loc": sum(item.get("loc", 0) for item in items),
        "completion_tokens": sum(
            item.get("usage", {}).get("completion_tokens", 0) for item in items
        ),
        "prompt_tokens": sum(
            item.get("usage", {}).get("prompt_tokens", 0) for item in items
        ),
    }
    output = {
        "date": date.today().isoformat(),
        "model": MODEL,
        "temperature_note": "The API accepts but ignores temperature.",
        "pricing_note": (
            "Fugu pricing depends on the highest-tier routed underlying model; "
            "the Standard subscription was used."
        ),
        "result": summary,
    }
    (ROOT / f"run{RUN_NUMBER}-summary.json").write_text(
        json.dumps(output, ensure_ascii=False, indent=2), encoding="utf-8"
    )


def main() -> None:
    action = sys.argv[1] if len(sys.argv) > 1 else "all"
    if action in {"generate", "all"}:
        generate()
    if action in {"evaluate", "all"}:
        evaluate()
    if action in {"repeat", "all"}:
        repeat_reliability()
    if action in {"summarize", "all"}:
        summarize()


if __name__ == "__main__":
    main()
