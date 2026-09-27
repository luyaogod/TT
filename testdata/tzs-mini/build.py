#!/usr/bin/env python3
"""重新生成 testdata/tzs-mini/ws —— 仓库自带的**最小 .tzs 工作区**。

为什么要它：`.tzs` 的语料回归从前只有"有真客户语料的那台机器"能跑（一份完整工作区 130 MB，
而且 `mta/` + `tbl/` 是数据不是代码）。这份夹具把那份语料**缩到 3.3 MB 并入库**，
于是任何人 `cd engine && ./build.sh` 之后就能跑那几条回归。

用法：
    python testdata/tzs-mini/build.py --src D:\\t100_wrok_dir\\hengshuo\\prd
    python testdata/tzs-mini/build.py --src <语料根> --out <别处>

`--src` 必须是**一个工作区**（该目录下有 `mta/`），包摊在它顶层。

## 这份工作区里为什么是这几样（每一条都是实测，别"顺手简化"）

  mta/   全量**减去三个大文件**（zooms.xml 4.8M / subroutines.xml 3M / messages.xml 2.6M）。
         **其余一个都不能少**：裁到只剩 tables.xml + ver + mod-fd.spec + core-br.spec 时，
         引擎照样回 `SUMMARY|ok`，但 `aapt300` 的布局元素从 515 悄悄变成 512 —— 静默少东西，
         正是这个仓库最怕的那种失败。留下 `tsd.xsd` 与否不影响这三个包。
  <模块>/tbl/<表>.tbl
         只拷**被包引用到的表**。模块名取 `mta/tables.xml` 里该表 `module` 属性的小写，
         **不是**包名前三位（`cpmp530` 的表 `pmdo_t` 住在 `apm/`）。
  <包>.tzs
         用例本身，见下面的 PACKAGES。

## 两个路径陷阱（踩过）

  * 工作区字符串必须用**反斜杠**。设计器会归一化**包路径**、不会归一化工作区字符串，
    所以 `TZSCLI_WS=C:/ws` + 包任意形式一律 `NotInCurrentWorkspaceException`。
    包路径用正斜杠没关系。完整规则见 skills/tt-dev-tzs/SKILL.md §4.4。
  * 包必须**在工作区目录之下**（纯字符串前缀比较），所以别把包拷到 %TEMP% 再喂给引擎。
"""
import argparse
import hashlib
import os
import re
import shutil
import sys
import zipfile

HERE = os.path.dirname(os.path.abspath(__file__))
DEFAULT_OUT = os.path.join(HERE, "ws")

# 三个包是按"小 / 中 / 大 + 特征"挑的，不是随手抓的（数见 README.md 的表）：
PACKAGES = [
    "cpmp530(c).tzs",   # 最小：12 个规格节点 / 3 种 kind / 9 个容器
    "aapp320(c).tzs",   # 中：5 种 kind / 21 种控件 / 3 个标签页
    "aapt300(c).tzs",   # 大：6 种 kind（含 tree）/ 26 种控件 / 12 页 / 26 个动作 / 65 个容器
]

# 去掉的只是这三份**目录性**数据；它们不参与表单加载，实测不影响上面三个包的任何计数。
DROP = ["zooms.xml", "subroutines.xml", "messages.xml"]


def referenced_tables(pkg):
    """包里出现过的表名。两种写法都要收：`table="x"`（属性）与 `<table ... name="x">`。"""
    found = set()
    with zipfile.ZipFile(pkg) as z:
        for name in z.namelist():
            text = z.read(name).decode("utf-8", "replace")
            found |= set(re.findall(r'table="([a-zA-Z0-9_]+)"', text))
            found |= set(re.findall(r'<table[^>]*name="([a-zA-Z0-9_]+)"', text))
    found.discard("true")
    found.discard("false")
    return found


def module_map(src):
    """表名 → 模块目录名（小写）。tables.xml 里 module 属性与 name 属性的先后不固定，两种都试。"""
    text = open(os.path.join(src, "mta", "tables.xml"), encoding="utf-8", errors="replace").read()
    pairs = re.findall(r"<table\b([^>]*)>", text)
    out = {}
    for attrs in pairs:
        n = re.search(r'name="([a-zA-Z0-9_]+)"', attrs)
        m = re.search(r'module="([^"]+)"', attrs)
        if n and m:
            out[n.group(1)] = m.group(1).lower()
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--src", required=True, help="真工作区（该目录下有 mta/）")
    ap.add_argument("--out", default=DEFAULT_OUT)
    args = ap.parse_args()

    src = os.path.abspath(args.src)
    out = os.path.abspath(args.out)
    if not os.path.isdir(os.path.join(src, "mta")):
        sys.exit(f"{src} 不是一个工作区：它下面没有 mta/")

    if os.path.isdir(out):
        shutil.rmtree(out)
    os.makedirs(out)

    # mta/ —— 全量减三
    os.makedirs(os.path.join(out, "mta"))
    n_mta = 0
    for f in sorted(os.listdir(os.path.join(src, "mta"))):
        if f in DROP:
            continue
        shutil.copy(os.path.join(src, "mta", f), os.path.join(out, "mta", f))
        n_mta += 1

    mod = module_map(src)
    n_tbl = 0
    for pkg in PACKAGES:
        src_pkg = os.path.join(src, pkg)
        if not os.path.exists(src_pkg):
            sys.exit(f"找不到包 {src_pkg}")
        shutil.copy(src_pkg, os.path.join(out, pkg))
        for t in sorted(referenced_tables(src_pkg)):
            m = mod.get(t)
            if not m:
                sys.exit(f"tables.xml 里没有表 {t}（包 {pkg}）—— 源语料与包不配套？")
            s = os.path.join(src, m, "tbl", t + ".tbl")
            if not os.path.exists(s):
                sys.exit(f"找不到 {s}（包 {pkg} 用到 {t}）")
            d = os.path.join(out, m, "tbl")
            os.makedirs(d, exist_ok=True)
            shutil.copy(s, d)
            n_tbl += 1

    files = [os.path.join(r, f) for r, _, fs in os.walk(out) for f in fs]
    kb = sum(os.path.getsize(p) for p in files) // 1024
    print(f"{out}\n  mta/ {n_mta} 个 + tbl/ {n_tbl} 个 + 包 {len(PACKAGES)} 个 "
          f"= {len(files)} 个文件 / {kb} KB")

    # 钉住：一份**清单**，由 TestMiniCorpusIsPinned 逐条比对。
    #
    # 为什么必须有：那三条回归的判据是**基线相对**的（`requireFixedPoint` 拿 pristine 自己
    # 当基线），所以一份重新生成错了的夹具会**自洽通过** —— 实测过一次：mta/ 裁到只剩四个文件时
    # `aapt300` 的布局元素从 515 静默变成 512，而 `SUMMARY` 还是 `ok`。
    # 清单把"这一份就是那一份"变成可判的。
    #
    # 路径用**仓库相对**的正斜杠：夹具在仓库里，不是机器特定的语料 ——
    # 所以这条断言不需要 engine/corpus.manifest 那种"指着另一份语料就跳过"的退路。
    rows = []
    for p in sorted(files):
        rel = os.path.relpath(p, out).replace("\\", "/")
        h = hashlib.sha256()
        with open(p, "rb") as fh:
            for chunk in iter(lambda: fh.read(1 << 20), b""):
                h.update(chunk)
        rows.append(f"{h.hexdigest()[:16]}  {os.path.getsize(p):>8}  {rel}")
    manifest = os.path.join(os.path.dirname(out.rstrip("\\/")), "manifest.txt")
    with open(manifest, "w", encoding="utf-8", newline="\n") as fh:
        fh.write("# 最小语料的固定清单 —— 由 build.py 生成，内部/开发者不改。\n"
                 "# 重新生成：python testdata/tzs-mini/build.py --src <真工作区>\n"
                 "# 形状：<sha256-16>  <字节>  <仓库相对路径>\n"
                 "# 为什么要有它：见 build.py 里写它那一段注释。\n")
        fh.write("\n".join(rows) + "\n")
    print(f"  manifest.txt {len(rows)} 条")


if __name__ == "__main__":
    main()
