#!/usr/bin/env python3
# 为一次 `.tzs` 工具面评测准备装置：每个任务一份**全新**的工作区副本 + 该任务的配置 + 一份 tt.exe。
#
# 为什么一个任务一份整副本（而不是共用一份、或只拷 mta/）：
#   * 四个执行者要并行跑、互不干扰，而工作区是**目录级**的绑定（引擎的 `open` 只收工作区之下的包）；
#   * 只拷 `mta/` 不够 —— 加载会弹一个无消息泵的模态框，然后 90 秒超时（E_FATAL_LOAD_TIMEOUT）。
#     这是实测出来的，不是猜的。
#
# 每份副本建立后立刻记一份 `pristine.txt`（逐文件 sha256）。**判据的基线来自这一份，不来自源目录**
#  —— 源语料以后会变（实测：磁盘上已经比 pin 多出三个包），而"执行者有没有动过不该动的东西"问的是
#  它拿到手的那一份。
#
# 用法：
#   python tools/eval/setup.py --src D:\t100_wrok_dir\hengshuo\prd --exe tt.exe
#   python tools/eval/setup.py --src <工作区> --exe <tt.exe> --run %TEMP%\ttrun --tasks E1 E2 E3 E4
#
# 四个任务 × 130 MB ≈ 520 MB，几秒钟到一分钟（robocopy）。跑完把 `prompts.md` 里对应任务的提示词
# 交给一个**只拿 SKILL 的干净执行者**，等它做完再跑 `grade.py`。

import argparse
import hashlib
import json
import os
import shutil
import subprocess
import sys

DEFAULT_RUN = os.path.join(os.environ.get("TEMP") or "/tmp", "ttrun")


def inventory(root):
    """{相对路径: sha256}，路径统一用 / —— 判据只需在两份清单之间比，用哪种分隔符不影响正确性。"""
    out = {}
    for dirpath, _dirnames, filenames in os.walk(root):
        for f in filenames:
            p = os.path.join(dirpath, f)
            rel = os.path.relpath(p, root).replace("\\", "/")
            h = hashlib.sha256()
            with open(p, "rb") as fh:
                for chunk in iter(lambda: fh.read(1 << 20), b""):
                    h.update(chunk)
            out[rel] = h.hexdigest()
    return out


def write_pristine(ws, path):
    inv = inventory(ws)
    with open(path, "w", encoding="utf-8", newline="\n") as fh:
        for rel in sorted(inv):
            fh.write("%s  %s\n" % (inv[rel], rel))
    listing = hashlib.sha256("".join(r + "\n" for r in sorted(inv)).encode()).hexdigest()
    return len(inv), listing


def copy_tree(src, dst):
    """整份拷贝。优先 robocopy（Windows 上快十倍），拿不到就退化到 shutil。"""
    os.makedirs(dst, exist_ok=True)
    if os.name == "nt" and shutil.which("robocopy"):
        # 参数由 Python 直接交给进程，不经过 shell —— 所以不会踩上"`/E` 被 Git Bash 当成路径
        # 转成 `E:/`"那个坑（实测过：robocopy 报 Invalid Parameter #3）。robocopy 的退出码里
        # 0-7 是"成功/有跳过"，>=8 才是失败。
        r = subprocess.run(["robocopy", src, dst, "/E", "/NFL", "/NDL", "/NJH", "/NJS", "/NP"],
                           capture_output=True)
        if r.returncode >= 8:
            raise SystemExit("robocopy 失败 rc=%d\n%s" % (
                r.returncode, r.stdout.decode("utf-8", "replace")[-2000:]))
        return
    shutil.copytree(src, dst, dirs_exist_ok=True)


def main():
    ap = argparse.ArgumentParser(description="为一个任务准备一份评测工作区")
    ap.add_argument("--src", required=True, help="工作区来源（整份拷；一般是语料所在的模块目录，"
                                                 "它下面有 mta/）")
    ap.add_argument("--exe", required=True, help="要测的那份 tt.exe（刚构建的那个）")
    ap.add_argument("--run", default=DEFAULT_RUN, help="装置根目录（默认 %%TEMP%%\\ttrun）")
    ap.add_argument("--tasks", nargs="+", default=["E1", "E2", "E3", "E4"])
    ap.add_argument("--server-exe", default="", help="写进 config.json 的 tzs.serverExe"
                                                    "（默认：仓库里 engine/out/tzs-server.exe）")
    a = ap.parse_args()

    src = os.path.abspath(a.src)
    exe = os.path.abspath(a.exe)
    if not os.path.isdir(os.path.join(src, "mta")):
        raise SystemExit("%s 下面没有 mta/ —— 它不是工作区（引擎会拒开它下面的包）" % src)
    if not os.path.isfile(exe):
        raise SystemExit("找不到 %s" % exe)
    server = a.server_exe or os.path.join(
        # tools/eval/setup.py → 上三级才是仓库根（eval → tools → 仓库）；
        # 少退一级会得到 tools/engine/out/…（第一次真跑就是这么错的）。
        os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))),
        "engine", "out", "tzs-server.exe")
    server = server.replace("/", os.sep)

    for t in a.tasks:
        d = os.path.join(a.run, t)
        shutil.rmtree(d, ignore_errors=True)
        os.makedirs(d)
        ws = os.path.join(d, "ws")
        copy_tree(src, ws)
        shutil.copy2(exe, os.path.join(d, "tt.exe"))
        cfg = {"schemaVersion": 2,
               # 两种分隔符引擎都认（它自己会把 / 归一成 \），但写出去的东西只用一种写法 ——
               # `--run` 常常是 shell 展开的（%TEMP% 里带反斜杠、参数里带正斜杠），
               # 不归一就会写出 `…\Local/Temp/ttrun2\E1\ws` 这种混着的路径。
               "tzs": {"workspace": ws.replace("/", "\\"),
                       "serverExe": server.replace("/", "\\")}}
        with open(os.path.join(d, "config.json"), "w", encoding="utf-8") as fh:
            json.dump(cfg, fh, ensure_ascii=False, indent=2)
        n, listing = write_pristine(ws, os.path.join(d, "pristine.txt"))
        print("%s: 就绪  %d 个文件  清单指纹 %s" % (t, n, listing[:16]))

    print()
    print("装置在 %s" % a.run)
    print("下一步：把 prompts.md 里对应任务的提示词交给一个只拿 SKILL 的执行者；")
    print("        四个都做完之后再跑 `python tools/eval/grade.py --run %s`。" % a.run)


if __name__ == "__main__":
    sys.exit(main())
