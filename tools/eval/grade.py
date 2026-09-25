#!/usr/bin/env python3
# 回读一次评测的产出。**不接受执行者的自述** —— 判据是产物本身。
#
# 两半：
#   1. 装置完整性：每个任务的工作区清单逐文件 sha256 与 `pristine.txt` 比 —— 只该多出它产出的那个包，
#      一个已存在文件都不许变（第一版基线只比文件名，那挡不住"改了内容但名字没变"）。
#   2. 判据回读：用**执行者自己那份 config**（同一个环境）打开它产出的包，读回目标状态。
#
# **只在四个执行者全部结束之后跑**：它们的守护进程与工作区就是我判据要读的那一份。
#
# 判据是数据，不是散在代码里的字符串（见 TASKS）—— 换任务就换这张表，改一处。
#
# 用法：python tools/eval/grade.py --run %TEMP%\ttrun

import argparse
import hashlib
import json
import os
import subprocess
import sys

# 每个任务的判据。`check` 的三种取值对应三种问法：
#   spec-attr  ——  产出包 → find `<field>` → get_component → `spec.<kind>.attrs.<attr> == value`
#   layout-attr —— 产出包 → find `<field>` → get_component → `layout.<attr> == value`
#   find       —— 产出包能开，且这些查询各有命中（"按表加了字段"这类结构性任务）
#   files-only —— 不产出包：只看工作区有没有被动过，答案在报告里由人读
TASKS = {
    "E1": dict(pkg="_ai_e1.tzs", check="spec-attr", field="net108", kind="field",
               attr="can_edit", value="N",
               why="规格侧的勾选位（Y/N），改对了要看它变成 N"),
    "E2": dict(pkg="_ai_e2.tzs", check="layout-attr", field="HBoxT1",
               attr="hidden", value="true",
               why="布局侧的 BOOLEAN；与 E1 是**对照组**：这里写 Y/N 会被值校验拒"),
    "E3": dict(pkg="_ai_e3.tzs", check="find", queries=["pmdlent", "pmdlsite"],
               why="任务级 field_add 加的两个字段：能开、且两个列名都有命中"),
    "E4": dict(pkg="", check="files-only",
               why="只开不写：工作区该一个文件都不多；三个答案在报告里"),
}


def inventory(root):
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


def read_pristine(path):
    out = {}
    if not os.path.isfile(path):
        return None
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            line = line.rstrip("\n")
            if not line:
                continue
            digest, rel = line.split("  ", 1)
            out[rel] = digest
    return out


def call(task_dir, verb, args=None, form=None):
    """跑一次**只读**动词（open / find_component / get_component / close）。"""
    exe = os.path.join(task_dir, "tt.exe")
    env = dict(os.environ, TT_CONFIG=os.path.join(task_dir, "config.json"))
    cmd = [exe, "dev", "tzs", verb, "--json"]
    if form:
        cmd += ["--form", form]
    argsfile = None
    if args is not None:
        # 参数文件落在任务目录、**不在 ws 里** —— 否则它会被清单比对当成"多出的文件"。
        argsfile = os.path.join(task_dir, "grade_args.json")
        with open(argsfile, "w", encoding="utf-8") as fh:
            json.dump(args, fh, ensure_ascii=False)
        cmd += ["--args-file", argsfile]
    try:
        p = subprocess.run(cmd, capture_output=True, timeout=600, env=env, cwd=task_dir)
    finally:
        if argsfile and os.path.exists(argsfile):
            os.remove(argsfile)
    out = p.stdout.decode("utf-8", "replace").strip()
    frame = None
    try:
        frame = json.loads(out.splitlines()[0]) if out else None
    except Exception:
        pass
    return frame


def result_of(frame):
    return (frame or {}).get("result") or {}


def error_of(frame):
    return (frame or {}).get("error") or {}


def grade_task(run, tid, spec, device, checks):
    d = os.path.join(run, tid)
    ws = os.path.join(d, "ws")
    print("\n== %s ==" % tid)
    print("   判据：%s" % spec["why"])

    # ---- 装置完整性
    now = inventory(ws)
    pristine = read_pristine(os.path.join(d, "pristine.txt"))
    if pristine is None:
        device.append("%s: 没有 pristine.txt —— 装置不是 setup.py 建的（判据仍会跑，"
                      "但「已存在的文件没被改过」这一条无从判定）" % tid)
        print("   !! 没有 pristine.txt，跳过清单比对")
    else:
        added = sorted(set(now) - set(pristine))
        removed = sorted(set(pristine) - set(now))
        changed = sorted(k for k in set(pristine) & set(now) if now[k] != pristine[k])
        print("   多出 %d 个：%s" % (len(added), added or "（无）"))
        print("   少了 %d 个：%s" % (len(removed), removed or "（无）"))
        print("   内容变了 %d 个：%s" % (len(changed), changed[:6] or "（无）"))
        if removed:
            checks.append("%s 少了文件：%s" % (tid, removed[:3]))
        if changed:
            checks.append("%s 改了已存在的文件：%s" % (tid, changed[:3]))

    if spec["check"] == "files-only":
        print("   （这条任务不产出包，答案在报告里由人读）")
        return

    # ---- 判据回读
    pkg = os.path.join(ws, spec["pkg"])
    if not os.path.isfile(pkg):
        print("   !! 找不到 %s" % spec["pkg"])
        checks.append("%s: 没有 %s（上面「多出」那几个就是它实际写下的东西）" % (tid, spec["pkg"]))
        return
    frame = call(d, "open", {"path": pkg})
    if not (frame or {}).get("ok"):
        print("   !! 打不开：%s" % json.dumps(error_of(frame), ensure_ascii=False)[:200])
        checks.append("%s: 产出的包打不开" % tid)
        return
    prog = result_of(frame).get("program")
    print("   打开 %s → program=%s" % (spec["pkg"], prog))

    if spec["check"] == "find":
        for q in spec["queries"]:
            f = call(d, "find_component", {"query": q}, form=prog)
            hits = [m.get("name") for m in (result_of(f).get("matches") or [])]
            print("   find_component %-10s → %s" % (q, hits or "**没有命中**"))
            if not hits:
                checks.append("%s: 找不到 %s" % (tid, q))
        call(d, "close", None, form=prog)
        return

    f = call(d, "find_component", {"query": spec["field"]}, form=prog)
    matches = result_of(f).get("matches") or []
    if not matches:
        print("   !! find_component %s 没有命中" % spec["field"])
        checks.append("%s: 找不到 %s" % (tid, spec["field"]))
        call(d, "close", None, form=prog)
        return
    el = matches[0]["path"]
    f2 = call(d, "get_component", {"path": el}, form=prog)
    res = result_of(f2)
    if spec["check"] == "spec-attr":
        got = ((res.get("spec") or {}).get(spec["kind"]) or {}).get("attrs", {}).get(spec["attr"])
    else:
        got = (res.get("layout") or {}).get(spec["attr"])
    ok = got == spec["value"]
    print("   **判据** %s 该是 %r，实际 %r  → %s"
          % (spec["attr"], spec["value"], got, "做对" if ok else "没做对"))
    if not ok:
        checks.append("%s: %s 是 %r，该是 %r" % (tid, spec["attr"], got, spec["value"]))
    call(d, "close", None, form=prog)


def main():
    ap = argparse.ArgumentParser(description="回读一次评测的产出")
    ap.add_argument("--run", default=os.path.join(os.environ.get("TEMP") or "/tmp", "ttrun"))
    ap.add_argument("--tasks", nargs="+", default=list(TASKS))
    a = ap.parse_args()

    device, checks = [], []
    for tid in a.tasks:
        spec = TASKS.get(tid)
        if spec is None:
            print("没有 %s 的判据定义" % tid)
            continue
        if not os.path.isdir(os.path.join(a.run, tid)):
            print("\n== %s ==\n   装置不存在，跳过" % tid)
            continue
        grade_task(a.run, tid, spec, device, checks)

    print("\n===== 汇总 =====")
    if device:
        # 装置问题与"任务没做对"是两件事，分开报：前者说明这一轮的数字要人看一眼再信，
        # 后者是执行者（或文档）真的没做到。混成一张表会让人把"脚本没跑对"读成"活干砸了"。
        print("装置问题 %d 个（判据仍然跑了，但这一轮是否可信要人看一眼）：" % len(device))
        for p in device:
            print("  - %s" % p)
    if checks:
        print("判据未达成 %d 个：" % len(checks))
        for p in checks:
            print("  - %s" % p)
        return 1
    if device:
        print("（判据都过了；退出码 2 只表示上面那些装置问题 —— 这一轮的数字要人看一眼再信）")
        return 2
    print("没有发现问题（判据都在上面逐条打出来了；执行者的自述不参与判定）")
    return 0


if __name__ == "__main__":
    sys.exit(main())
