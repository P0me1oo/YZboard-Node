#!/usr/bin/env python3
"""在独立网络空间内运行预编译回归测试，不安装 Node 服务。"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import subprocess
import tempfile
import time


PACKAGES = ("controlplane", "service", "singbox", "xray")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--artifacts", type=Path, required=True, help="已校验的 Linux amd64 测试包目录")
    args = parser.parse_args()
    artifacts = args.artifacts.resolve(strict=True)
    if platform.system() != "Linux" or platform.machine() not in ("x86_64", "amd64"):
        raise RuntimeError("本测试包要求 Linux amd64")
    if os.readlink("/proc/self/ns/net") == os.readlink("/proc/1/ns/net"):
        raise RuntimeError("请通过 unshare --net --fork 启动，测试需要独立网络空间")

    checksums = {}
    for line in (artifacts / "SHA256SUMS").read_text(encoding="utf-8").splitlines():
        checksum, name = line.split(maxsplit=1)
        checksums[name.lstrip("*")] = checksum.lower()
    binaries = {name: artifacts / f"{name}-linux-amd64.test" for name in PACKAGES}
    for binary in binaries.values():
        if binary.resolve(strict=True).parent != artifacts or not binary.is_file():
            raise RuntimeError("测试二进制路径越界")
        with binary.open("rb") as stream:
            digest = hashlib.file_digest(stream, "sha256").hexdigest()
        if checksums.get(binary.name) != digest:
            raise RuntimeError(f"{binary.name} 校验失败")
        if not os.access(binary, os.X_OK):
            raise RuntimeError(f"{binary.name} 缺少执行权限")

    subprocess.run(["ip", "link", "set", "lo", "up"], check=True, capture_output=True)
    results = []
    try:
        with tempfile.TemporaryDirectory(prefix="yznode-regression-", dir="/var/tmp") as directory:
            env = dict(os.environ, TMPDIR=directory, GOMAXPROCS="2", GORACE="halt_on_error=1")
            for package, binary in binaries.items():
                started = time.monotonic()
                # 程序日志可能包含随机测试身份，仅保留在内存中，不输出和持久化。
                run = subprocess.run(
                    [str(binary), "-test.v", "-test.timeout=5m"],
                    cwd=directory, env=env, capture_output=True, text=True,
                    encoding="utf-8", errors="replace", timeout=330,
                )
                output = run.stdout + run.stderr
                passed = len(re.findall(r"^\s*--- PASS:", output, flags=re.MULTILINE))
                skipped = len(re.findall(r"^\s*--- SKIP:", output, flags=re.MULTILINE))
                result = {
                    "package": package,
                    "exit_code": run.returncode,
                    "passed": passed,
                    "skipped": skipped,
                    "race_detected": "WARNING: DATA RACE" in output,
                    "seconds": round(time.monotonic() - started, 3),
                }
                results.append(result)
                print(json.dumps(result, ensure_ascii=False), flush=True)
                if run.returncode or skipped:
                    raise RuntimeError(f"{package} 未完整通过，退出码 {run.returncode}，跳过 {skipped} 项")
    finally:
        # 报告只包含状态和计数，不保存测试 Token、身份、配置或程序原始日志。
        (artifacts / "regression-results.json").write_text(
            json.dumps(results, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
        )
    print("回归测试全部通过；一次性目录已清理。", flush=True)


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"回归验收失败：{error if isinstance(error, RuntimeError) else type(error).__name__}", flush=True)
        raise SystemExit(1)
