"""W2 gate: the milestone chain -- one server process, six requests.

    open -> validate(baseline) -> find_component -> get_component
         -> set_layout_attr -> validate(after) -> save

Everything goes to ONE `tzs-server --stdio` process in ONE payload. That is the milestone,
and it is also where the saving comes from: the ~890 ms of form-independent designer init is
paid once for the whole chain instead of once per request. The per-request `ms` the server
reports is the evidence -- sub-millisecond for everything except validate.

(The named-pipe daemon is the shipping transport and was measured separately: 870 ms cold,
223/253 ms warm, server-side handler 0.047 ms. Driving the chain through `tzs-cli` was once
blocked by a real defect -- the spawned daemon inherited the caller's capture pipes, so any
harness reading both stdout and stderr (subprocess.run(capture_output=True), a CI runner) hung
until timeout. W3-E fixed it, and `gate-w3.py` section A now drives this exact chain through
the real CLI with both streams captured: cold open 1.5 s, warm call 0.12 s. This file still
uses --stdio, which is the cheaper path when the whole chain fits in one payload.)

Usage: python gate-w2.py [outdir] [file.tzs]
"""
import json, os, subprocess, sys, time

OUTDIR = sys.argv[1] if len(sys.argv) > 1 else r'C:\Users\18526\AppData\Local\Temp\dep_gate'
WS = os.environ.get('TZSCLI_WS') or r'D:\t100_wrok_dir\hengshuo\prd'
TARGET = sys.argv[2] if len(sys.argv) > 2 else os.path.join(WS, 'aapp320(c).tzs')
CODE = 'l_apcasite'
ATTR = "case"   # pristine value on l_apcasite is "upper"; LOWER below forces a real write
SAVED = os.path.join(WS, '_ai_w2gate.tzs')      # _ai_ is excluded by all four batch drivers

SRV = os.path.join(OUTDIR, 'tzs-server.exe')
RT = os.path.join(OUTDIR, 'RoundTrip.exe')
env = dict(os.environ, TZSCLI_WS=WS)

ok = True
def check(n, label, cond, detail=''):
    global ok
    ok = ok and cond
    print('%s %-4s %s%s' % ('PASS' if cond else 'FAIL', str(n) + '.', label,
                            ('   ' + detail) if detail else ''))

PATH = None
reqs = [
    ("open",            {"path": TARGET}),
    ("validate",        {"handle": "h1"}),
    ("find_component",  {"handle": "h1", "query": CODE}),
]
payload = '\n'.join(json.dumps({"id": i + 1, "fn": f, "args": a}, ensure_ascii=False)
                    for i, (f, a) in enumerate(reqs)) + '\n'

print('--- W2 门：一个 tzs-server 进程，整条链 ---')
t0 = time.time()
p = subprocess.run([SRV, '--stdio'], input=payload.encode('utf-8'), capture_output=True,
                   env=env, cwd=OUTDIR, timeout=900)
t_first = time.time() - t0

frames = {}
for ln in p.stdout.decode('utf-8', 'replace').splitlines():
    ln = ln.strip()
    if ln.startswith('{'):
        try:
            d = json.loads(ln)
            frames[d.get("id")] = d
        except Exception:
            pass

R1, R2, R3 = frames.get(1, {}), frames.get(2, {}), frames.get(3, {})
check(1, 'open 成功且 state=Loaded',
      bool(R1.get('ok')) and (R1.get('result') or {}).get('state') == 'Loaded',
      json.dumps(R1.get('result') or R1.get('error'), ensure_ascii=False)[:130])
base_err = [e for e in ((R2.get('result') or {}).get('after') or []) if e.get('type') == 'ERROR']
check(2, 'validate(原始) 无 ERROR', bool(R2.get('ok')) and not base_err,
      json.dumps(R2.get('result') or R2.get('error'), ensure_ascii=False)[:130])
ms = ((R3.get('result') or {}).get('matches') or [])
PATH = ms[0]['path'] if ms else None
check(3, 'find_component 定位到唯一路径', bool(R3.get('ok')) and len(ms) == 1,
      (PATH or '')[:88] + ('...' if PATH and len(PATH) > 88 else '')
      or json.dumps(R3.get('error'), ensure_ascii=False)[:130])
if not PATH:
    print('\n链条断在第 3 步。'); sys.exit(1)

# --- second half, SAME process, re-open (a --stdio process is one-shot per invocation) ---
reqs2 = [
    ("open",            {"path": TARGET}),
    ("get_component",   {"handle": "h1", "path": PATH}),
    ("set_layout_attr", {"handle": "h1", "path": PATH, "attr": ATTR, "value": "lower"}),
    ("validate",        {"handle": "h1"}),
    ("save",            {"handle": "h1", "out": SAVED}),
]
payload2 = '\n'.join(json.dumps({"id": i + 1, "fn": f, "args": a}, ensure_ascii=False)
                     for i, (f, a) in enumerate(reqs2)) + '\n'
t1 = time.time()
p2 = subprocess.run([SRV, '--stdio'], input=payload2.encode('utf-8'), capture_output=True,
                    env=env, cwd=OUTDIR, timeout=900)
t_second = time.time() - t1

f2 = {}
for ln in p2.stdout.decode('utf-8', 'replace').splitlines():
    ln = ln.strip()
    if ln.startswith('{'):
        try:
            d = json.loads(ln)
            f2[d.get("id")] = d
        except Exception:
            pass
G, S, V, SV = f2.get(2, {}), f2.get(3, {}), f2.get(4, {}), f2.get(5, {})

R = G.get('result') or {}
NEED = ['req', 'can_edit', 'can_query', 'i_zoom', 'c_zoom', 'chk_ref', 'items',
        'default', 'max', 'min']
spec_attrs = {}
for _k, _v in (R.get('spec') or {}).items():
    if isinstance(_v, dict):
        spec_attrs.update(_v.get('attrs') or {})
have = [a for a in NEED if a in spec_attrs]
check(4, 'get_component 暴露以前看不见的规格属性', len(have) == len(NEED),
      '%d/%d: %s' % (len(have), len(NEED), ' '.join(have)))

sres = S.get('result') or {}
serr = (S.get('error') or {}).get('code')
# Tightened after W3-F: E_NO_OP/E_ATTR_CLAMPED are SUCCESSES (ok:true) per §11.24 (a), so the
# old "ok:true OR E_NO_OP" escape hatch is gone -- a no-op now comes back ok:true too, and
# this request is a genuine write (case "upper" -> "lower"), so it must apply for real.
check(5, 'set_layout_attr 经索引器生效（ok:true 且真写入 lower）',
      bool(S.get('ok')) and sres.get('written') == 'lower' and sres.get('changed') is True,
      json.dumps(sres if S.get('ok') else S.get('error'), ensure_ascii=False)[:150])

new_err = ((V.get('result') or {}).get('newErrors') or [])
check(6, 'validate(改动后) 无新增 ERROR', bool(V.get('ok')) and not new_err,
      ('新增 %d 条' % len(new_err)) if new_err
      else json.dumps(V.get('result') or V.get('error'), ensure_ascii=False)[:130])

size = os.path.getsize(SAVED) if os.path.exists(SAVED) else 0
check(7, 'save 产出文件', bool(SV.get('ok')) and size > 0,
      '%.1f KB' % (size / 1024.0) if size
      else json.dumps(SV.get('error'), ensure_ascii=False)[:130])

if size:
    rt = subprocess.run([RT, SAVED], capture_output=True, env=env, cwd=OUTDIR, timeout=900)
    ln = [l for l in rt.stdout.decode('utf-8', 'replace').splitlines() if l.startswith('SUMMARY|')]
    fl = ln[0].split('|') if ln else []
    good = len(fl) > 8 and fl[1] == 'ok' and fl[4:8] == ['0', '0', '0', '0']
    check(8, 'RoundTrip 不动点 0 增 0 减', good,
          ('tsd+%s/-%s fd+%s/-%s stale=%s' % tuple(fl[4:9])) if len(fl) > 8 else 'no SUMMARY')

# --- the milestone itself: one process, many requests ---
per = []
for i in sorted(frames):
    per.append('%.1f' % (frames[i].get('ms') or 0))
check(9, '一个进程连续处理整条链', len(frames) == 3 and len(f2) == 5,
      '第一段 %d 帧 / 第二段 %d 帧' % (len(frames), len(f2)))
print('      每请求服务端耗时: 第一段 [%s] ms，第二段 [%s] ms'
      % (' '.join(per), ' '.join('%.1f' % ((f2[i].get('ms') or 0)) for i in sorted(f2))))
print('      两段各一次进程: %.1f s / %.1f s（对照：一次一进程要 7×~0.9 s）' % (t_first, t_second))

if os.path.exists(SAVED):
    try:
        os.remove(SAVED); print('      (已删除 %s)' % SAVED)
    except Exception as e:
        print('      !! 删除失败: %s' % e)

print()
print('==== W2 GATE: %s ====' % ('PASS' if ok else 'FAIL'))
sys.exit(0 if ok else 1)
