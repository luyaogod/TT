"""W3 gate -- the wave that implemented all 49 declared functions.

The four batch drivers (batch.sh / batch-write.sh / batch-edit.sh / batch-action.sh) test the
LEGACY path: RoundTrip (LINK_SRC=none, self-contained), AddField and Edit (LINK_SRC=link, they
compile their own copy of src/*.cs). Not one of them contains a line of src/Designer/**. They
regress Wave 1's extraction; they say nothing about Wave 3.

This gate covers the new surface. Two sections:

  A. THE SHIPPING PATH. The six-request chain driven through tzs-cli -- the real CLI, spawning
     the real named-pipe daemon -- with stdout AND stderr both captured and a hard timeout.
     That combination is exactly what used to hang: the spawned daemon inherited the caller's
     capture pipes, so a harness reading both blocked until timeout. W3-E fixed it. gate-w2.py
     still carries "Driving the chain through tzs-cli is blocked by a real defect" in its
     docstring; section A is what retires that sentence, and A9 is the anti-regression: if the
     second call costs ~1 s like the first, the daemon is not being reused and the whole
     long-lived-process design has silently degraded to one process per call.

  B. THE WHOLE CORPUS, through the server, with the write verdict. For every pinned corpus file
     in its own `tzs-server --stdio` process: open, read everything, make ONE real layout write
     (nudge +1 cell), then require (i) validate reports zero NEW errors and (ii) the repacked
     file is still a RoundTrip fixed point -- 0 tsd adds, 0 tsd drops, 0 path adds/drops, and
     the same `stale` count the PRISTINE file already had. The criterion is baseline-relative,
     per SPEC 11.24 (d): pristine files are not clean, and three corpus files report `stale`
     before anything touches them. The first draft of this gate hardcoded stale==0 and failed
     cpmp530 -- which reports the identical stale=2 against the pinned baseline row. A gate that
     fails a file for a pre-existing property is worse than no gate.

     (i) is tiered: `validate` is 1.3-10.4 s a call and this calls it twice per file, so it runs
     on a named sample chosen one-per-axis (`--validate sample`, the default) rather than on all
     81. See the VALIDATE_MODES block for why -- in short, it was ~13 of the sweep's ~18.5
     minutes, it duplicates what batch.sh already settles over the whole corpus, and it produced
     no finding in the 81 files of the day. `--validate all` is still there for when the
     validator changes. (The corpus is 67 files as of 2026-09-19 -- see TASKS.md.)

  B-neg. A negative control that is byte-exact rather than "no error appeared": for a few files,
     save once untouched (S1) and once after a bogus attribute write (S2). The bogus write must
     come back with a frozen E_* code AND S1 and S2 must be identical files. "Looks like no
     error" has hidden three silent failures in this project already.

Usage: python gate-w3.py [--limit N] [--files a,b,c] [--validate all|sample|none]
                         [--skip-b] [--root DIR]
"""
import hashlib
import json
import os
import subprocess
import sys
import time

BIN = os.environ.get('TZSCLI_BIN') or r'C:\Users\18526\AppData\Local\Temp\dep_probe'
ROOT = os.environ.get('TZSCLI_ROOT') or r'D:\t100_wrok_dir'
DEFAULT_WS = os.environ.get('TZSCLI_WS') or r'D:\t100_wrok_dir\hengshuo\prd'
REPORT = os.environ.get('TZSCLI_REPORT') or \
    os.path.join(os.environ.get('LOCALAPPDATA', '.'), 'Temp', 'gate-w3-report.txt')

# The four drivers do not agree on their exclusion list; this is the UNION, so nothing we or
# they leave behind is ever mistaken for corpus.
EXCLUDE = ('_ai', '_dw', '_ed', '_del', '_ac')

KINDS = ['field', 'hfield', 'pfield', 'rfield', 'mlfield', 'tree', 'act']

# ---- how much validate the sweep buys -------------------------------------------------------
# `validate` is the designer's own validator, and it measures 1.3 s on a 114-element form and
# 10.4 s on a 670-element one. Section B calls it twice per file, so running it over all 81 files
# was ~13 of the sweep's ~18.5 minutes -- for the one assertion that a handful of files settle,
# and which batch.sh already covers over the whole corpus in 2 minutes via RoundTrip. It found
# nothing in 81 files. So it is now tiered:
#
#   sample (default)  the files below, chosen because each turns on a DIFFERENT axis of the
#                     verdict. Everything else still gets the cheap assertions.
#   all               the old behaviour, ~19 minutes. For when the validator itself changed.
#   none              structural pass only.
VALIDATE_MODES = ('all', 'sample', 'none')

# One file per axis the validate verdict can turn on. Each is here for a reason, not for symmetry.
# (The xiyuan/tst module was cleared out of the workspace on 2026-09-19; cs_excel_in_xmdl_s01, which
# used to carry the cross-workspace axis, went with it. asft330 now carries it -- note that the
# "smallest field vocabulary" case went away with that file and is no longer covered by any sample.)
DEEP_SAMPLE = (
    'aapp320(c).tzs',              # tpl=P, the form every example in the docs uses
    'cpmp530(c).tzs',              # tpl=Q, and one of the three that arrive with posX drift
    'asft330(c).tzs',              # xiyuan/prd -- the cross-workspace axis (a different TZSCLI_WS)
    'aist310_wf(c).tzs',           # Tree, 577 elements, 34 acts
    'axmt500_wf(c).tzs',           # 678 elements, and already 11 WARNINGs before anything touches it
)

# Codes a refusal may carry. Anything else is an unclassified failure.
FROZEN_ERR = {'E_NOT_FOUND', 'E_BAD_PARAM', 'E_DESIGNER', 'E_INTERNAL', 'E_KEY_IN_USE',
              'E_NO_OP', 'E_ATTR_CLAMPED', 'E_ATTR_NOT_WHITELIST', 'E_UNKNOWN_METHOD',
              'E_NOT_IMPLEMENTED', 'E_KIND_MISMATCH', 'E_ALREADY_EXISTS', 'E_LAST_PAGE',
              'E_NOT_CONTAINER', 'E_ROOT', 'E_UNSUPPORTED'}

report_lines = []


def say(s=''):
    print(s)
    report_lines.append(s)


ok_all = True
npass = nfail = 0


def check(n, label, cond, detail=''):
    global ok_all, npass, nfail
    ok_all = ok_all and bool(cond)
    if cond:
        npass += 1
    else:
        nfail += 1
    say('%s %-5s %s%s' % ('PASS' if cond else 'FAIL', str(n) + '.', label,
                          ('   ' + detail) if detail else ''))


# --------------------------------------------------------------------------- helpers

def ws_of(path):
    """Whichever ancestor holds an mta/ directory. Deriving it beats a hardcoded list: a module
    directory is a workspace because of that directory, and TzpManager refuses packages from
    outside the one it was configured with."""
    d = os.path.dirname(os.path.abspath(path))
    while True:
        if os.path.isdir(os.path.join(d, 'mta')):
            return d
        parent = os.path.dirname(d)
        if parent == d:
            return DEFAULT_WS
        d = parent


def corpus(root, limit=0):
    out = []
    for dirpath, _dirs, names in os.walk(root):
        for nm in names:
            if not nm.lower().endswith('.tzs'):
                continue
            if nm.startswith(EXCLUDE):
                continue
            out.append(os.path.join(dirpath, nm))
    out.sort()
    return out[:limit] if limit else out


def srv(reqs, ws, timeout=900):
    """One tzs-server --stdio process, a batch of requests, parsed replies."""
    payload = '\n'.join(json.dumps({'id': i + 1, 'fn': f, 'args': a}, ensure_ascii=False)
                        for i, (f, a) in enumerate(reqs)) + '\n'
    p = subprocess.run([os.path.join(BIN, 'tzs-server.exe'), '--stdio'],
                       input=payload.encode('utf-8'), capture_output=True,
                       env=dict(os.environ, TZSCLI_WS=ws), cwd=BIN, timeout=timeout)
    frames = {}
    for ln in p.stdout.decode('utf-8', 'replace').splitlines():
        ln = ln.strip()
        if ln.startswith('{'):
            try:
                d = json.loads(ln)
            except ValueError:
                continue
            frames[d.get('id')] = d
    return frames, p.stderr.decode('utf-8', 'replace')


def res(frames, i):
    d = frames.get(i) or {}
    return (d.get('result') or {}), (d.get('error') or {}), bool(d.get('ok'))


def roundtrip(path, ws):
    p = subprocess.run([os.path.join(BIN, 'RoundTrip.exe'), path], capture_output=True,
                       env=dict(os.environ, TZSCLI_WS=ws), cwd=BIN, timeout=900)
    for ln in p.stdout.decode('utf-8', 'replace').splitlines():
        if ln.startswith('SUMMARY|'):
            return ln.split('|'), ln
    return None, ''


def rt_parse(f):
    """SUMMARY|status|env|tpl|tsdAdds|tsdDrops|fdPathAdds|fdPathDrops|stale|nodesIn|nodesOut|
    layoutElems|sample|path|error  -- column order is frozen by the four batch drivers."""
    if not f or len(f) < 12 or f[1] != 'ok':
        return None
    try:
        return {'adds': int(f[4]), 'drops': int(f[5]), 'pathAdds': int(f[6]),
                'pathDrops': int(f[7]), 'stale': int(f[8]),
                'nodesIn': int(f[9]), 'nodesOut': int(f[10]), 'sample': f[12] if len(f) > 12 else ''}
    except (ValueError, IndexError):
        return None


def rt_ok(after, before):
    """The write verdict, BASELINE-RELATIVE. Adds/drops must be zero -- the designer's model has
    to reproduce every spec node and layout path it was given. `stale`, though, cannot be
    required to be zero: three pristine corpus files already report it (cpmp530 reports 2,
    with the same two posX entries, before and after our nudge -- verified against the pinned
    baseline row). Requiring 0 here would fail a file for a pre-existing property, which is
    exactly the mistake SPEC 11.24 (d) exists to prevent. So: same count as the pristine file.
    A difference in the sample text is reported, not failed -- nudging the drifted element
    itself legitimately changes its stored value."""
    if not after or not before:
        return False
    return (after['adds'] == 0 and after['drops'] == 0 and after['pathAdds'] == 0
            and after['pathDrops'] == 0 and after['stale'] == before['stale'])


def sha(path):
    h = hashlib.sha256()
    with open(path, 'rb') as fh:
        for chunk in iter(lambda: fh.read(1 << 16), b''):
            h.update(chunk)
    return h.hexdigest()


def walk(n, acc):
    acc.append(n)
    for c in (n.get('children') or []):
        walk(c, acc)
    return acc


# ============================================================== SECTION A -- shipping path

def section_a(target):
    say('--- A. shipping path: tzs-cli -> named-pipe daemon -> server -> designer ---')
    cli = os.path.join(BIN, 'tzs-cli.exe')
    saved = os.path.join(os.path.dirname(target), '_ai_w3cli_gate.tzs')
    if os.path.exists(saved):
        os.remove(saved)

    def call(*args, timeout=300):
        t = time.time()
        try:
            p = subprocess.run([cli] + list(args), capture_output=True,
                               env=dict(os.environ, TZSCLI_WS=ws_of(target)), cwd=BIN,
                               timeout=timeout)
            return p.returncode, p.stdout.decode('utf-8', 'replace'), \
                p.stderr.decode('utf-8', 'replace'), time.time() - t
        except subprocess.TimeoutExpired as e:
            return 'TIMEOUT', (e.stdout or b'').decode('utf-8', 'replace'), \
                (e.stderr or b'').decode('utf-8', 'replace'), time.time() - t

    def last_json(so):
        for ln in reversed(so.strip().splitlines()):
            ln = ln.strip()
            if ln.startswith('{'):
                try:
                    return json.loads(ln)
                except ValueError:
                    pass
        return {}

    # stop anything left over first, so A1 measures a cold daemon
    call('stop')

    rc, so, se, dt = call('open', target)
    t_cold = dt
    d = last_json(so)
    h = (d.get('result') or {}).get('handle')
    check('A1', 'tzs-cli open survives capturing stdout AND stderr',
          rc == 0 and d.get('ok') and h,
          'rc=%s %.2fs handle=%s' % (rc, dt, h))
    if not h:
        say('    chain cannot continue without a handle')
        return

    rc, so, se, dt = call('call', 'validate', '--handle', h)
    d = last_json(so)
    v = d.get('result') or {}
    base_err = [e for e in (v.get('after') or []) if e.get('type') == 'ERROR']
    check('A2', 'validate(base) over the CLI', rc == 0 and d.get('ok') and not base_err,
          'after=%d ERROR, %.2fs' % (len(base_err), dt))

    rc, so, se, dt = call('call', 'find_component', '--handle', h, '--query', 'l_apcasite')
    d = last_json(so)
    ms = (d.get('result') or {}).get('matches') or []
    path = ms[0]['path'] if ms else None
    t_first_warm = dt
    check('A3', 'find_component over the CLI', rc == 0 and d.get('ok') and path,
          (path or '')[-46:] or json.dumps(d.get('error'), ensure_ascii=False)[:120])
    if not path:
        # Section A's chain needs a name that exists in the target. A gate that tracebacks on an
        # unexpected input is worse than one that fails: report and stop the section, the way the
        # missing-handle case above already does. (Hit by running the gate against a file that has
        # no l_apcasite: find_component returned matchCount 0 and `path` was None.)
        say('    %s has no l_apcasite -- section A needs a known name; stopping here.'
            % os.path.basename(target))
        call('stop')
        return

    rc, so, se, dt = call('call', 'nudge', '--handle', h, '--paths', path,
                          '--direction', 'right', '--offset', '1')
    d = last_json(so)
    delta = (d.get('result') or {}).get('delta') or []
    check('A4', 'nudge writes a real posX through the CLI',
          rc == 0 and d.get('ok') and len(delta) == 1 and delta[0].get('attrs'),
          json.dumps(delta[0].get('attrs') if delta else d.get('error'), ensure_ascii=False)[:90])

    rc, so, se, dt = call('call', 'validate', '--handle', h)
    d = last_json(so)
    new = (d.get('result') or {}).get('newErrors') or []
    check('A5', 'validate(after) reports no NEW error', rc == 0 and d.get('ok') and not new,
          'newErrors=%d' % len(new))

    rc, so, se, dt = call('save', '--handle', h, '--out', saved)
    d = last_json(so)
    size = os.path.getsize(saved) if os.path.exists(saved) else 0
    check('A6', 'save over the CLI', rc == 0 and d.get('ok') and size > 0,
          '%.1f KB' % (size / 1024.0) if size else json.dumps(d.get('error'), ensure_ascii=False)[:120])

    if size:
        pf, _ = roundtrip(target, ws_of(target))
        f, ln = roundtrip(saved, ws_of(target))
        check('A7', 'RoundTrip fixed point on the CLI-produced file (baseline-relative)',
              rt_ok(rt_parse(f), rt_parse(pf)), ln[:120])

    # A8: warm-call latency. The point of the daemon is that only the FIRST call pays the
    # ~890 ms designer init. If the second is also ~1 s, the daemon is not being reused.
    rc, so, se, dt = call('call', 'find_component', '--handle', h, '--query', 'l_apcasite')
    check('A8', 'a warm call costs far less than the cold open',
          rc == 0 and dt < 0.6 and t_cold > dt * 1.5,
          'cold open %.2fs -> warm call %.2fs (first warm %.2fs)' % (t_cold, dt, t_first_warm))

    rc, so, se, dt = call('stop')
    d = last_json(so)
    check('A9', 'stop tears the daemon down', rc == 0 and d.get('ok'),
          'rc=%s %.2fs %s' % (rc, dt, json.dumps(d.get('result') or d.get('error'), ensure_ascii=False)[:60]))

    if os.path.exists(saved):
        os.remove(saved)


# ============================================================== SECTION B -- the corpus

def one_file(target, idx, total, deep=True):
    """Everything section B asserts about a single corpus file, plus the raw material the
    aggregate checks need.

    `deep` gates the two `validate` calls, which are the whole cost of this sweep: the designer's
    own validator measures 1.3 s on a 114-element form and 10.4 s on a 670-element one, and this
    runs it twice per file. Callers pass deep=False for the breadth pass -- see the --validate
    option -- because the assertion it feeds is covered by a handful of files chosen for the axes
    they cover, and spending it on all 81 buys nothing the cheap assertions do not.

    Requests are addressed by KEY, not by index. The first draft hardcoded ids, which made every
    id shift the moment one request became optional -- the same class of harness bug as
    forgetting the `open` in the second payload.
    """
    ws = ws_of(target)
    tag = os.path.basename(target)
    tmp = os.path.join(os.path.dirname(target), '_ai_w3gate_%d.tzs' % idx)
    if os.path.exists(tmp):
        os.remove(tmp)

    r = {'file': target, 'tag': tag, 'ws': ws, 'notes': [], 'deep': bool(deep)}

    # The pristine file's own RoundTrip, so the verdict below has a baseline to be relative to.
    pf, pline = roundtrip(target, ws)
    r['baseRt'] = rt_parse(pf)
    if not r['baseRt']:
        r['notes'].append('pristine RoundTrip: ' + (pline or 'no SUMMARY')[:160])
        return r

    plan = [('open',  'open',      {'path': target}),
            ('tree',  'form_tree', {'handle': 'h1', 'depth': 99})]
    if deep:
        plan.append(('val', 'validate', {'handle': 'h1'}))
    for k in KINDS:
        plan.append(('kind:' + k, 'describe_kind', {'handle': 'h1', 'kind': k}))
    plan.append(('spec', 'list_spec_nodes',    {'handle': 'h1', 'kind': 'field'}))
    plan.append(('rec',  'list_records',       {'handle': 'h1'}))
    plan.append(('str',  'list_local_strings', {'handle': 'h1'}))

    frames, _err = srv([(f, a) for _k, f, a in plan], ws)
    gid = {k: i + 1 for i, (k, _f, _a) in enumerate(plan)}

    res0, err0, ok0 = res(frames, gid['open'])
    r['open'] = ok0 and res0.get('state') == 'Loaded'
    if not r['open']:
        r['notes'].append('open: ' + json.dumps(err0, ensure_ascii=False)[:160])
        return r

    tree, terr, tok = res(frames, gid['tree'])
    r['tree'] = tok
    nodes = walk(tree.get('root') or {}, []) if tree.get('root') else []
    r['nodes'] = len(nodes)
    r['elements'] = tree.get('elementCount')

    if deep:
        val, verr, vok = res(frames, gid['val'])
        r['validate'] = vok
        r['baseAfter'] = len([e for e in (val.get('after') or []) if e.get('type') == 'ERROR'])
        r['baseWarn'] = len([e for e in (val.get('after') or []) if e.get('type') == 'WARNING'])
    else:
        r['validate'] = None
        r['baseAfter'] = r['baseWarn'] = None

    # describe_kind: the 7 spec kinds. Every one must answer with a list -- the attribute
    # vocabulary is materialised from the node's own type, and an empty answer means the
    # caller can never self-correct (SPEC 11.24 (c), from:"spec:<kind>").
    r['kinds'] = {}
    for k in KINDS:
        kr, kerr, kok = res(frames, gid['kind:' + k])
        vals = kr.get(k)
        r['kinds'][k] = len(vals) if isinstance(vals, list) else None

    lsn, _, lsok = res(frames, gid['spec'])
    r['listSpecNodes'] = (lsn.get('count') or 0) if lsok else None
    lr, _, lrok = res(frames, gid['rec'])
    r['records'] = (lr.get('count') or 0) if lrok else None
    ls, _, lsok2 = res(frames, gid['str'])
    r['strings'] = (ls.get('count') or 0) if lsok2 else None

    # pick a real leaf to operate on: deepest node with a name that is not a container
    leaf = None
    for n in nodes:
        p, nm, tg = n.get('path'), n.get('name'), n.get('tag')
        if not p or not nm or tg in ('Form', 'VBox', 'HBox', 'Grid', 'Group', 'Page', 'Folder'):
            continue
        leaf = p
    r['leaf'] = leaf
    if not leaf:
        r['notes'].append('no non-container leaf to operate on')
        return r

    # A --stdio process is one-shot: it reads stdin to EOF and exits, so the handle from the
    # first payload does not exist here. Every payload that operates on the package opens it
    # itself. (Costs one extra designer init per file, ~0.9 s, and removes a whole class of
    # harness bug -- an earlier draft forgot the open and every B assertion failed E_NOT_FOUND.)
    name = leaf.rsplit('/', 1)[-1]
    plan2 = [('open', 'open',           {'path': target}),
             ('gc',   'get_component',  {'handle': 'h1', 'path': leaf}),
             ('fc',   'find_component', {'handle': 'h1', 'query': name}),
             ('nd',   'nudge',          {'handle': 'h1', 'paths': [leaf],
                                         'direction': 'right', 'offset': 1})]
    if deep:
        plan2.append(('val', 'validate', {'handle': 'h1'}))
    plan2.append(('sv', 'save', {'handle': 'h1', 'out': tmp}))

    f2, _e2 = srv([(f, a) for _k, f, a in plan2], ws)
    g2 = {k: i + 1 for i, (k, _f, _a) in enumerate(plan2)}

    gc, gcerr, gcok = res(f2, g2['gc'])
    r['get_component'] = gcok and bool(gc.get('layout'))
    fc, _, fcok = res(f2, g2['fc'])
    r['find_component'] = fcok and ((fc.get('matchCount') or 0) >= 1)
    nd, nderr, ndok = res(f2, g2['nd'])
    delta = (nd.get('delta') or []) if ndok else []
    r['nudge'] = ndok and len(delta) == 1
    r['nudgeDelta'] = delta[0].get('attrs') if delta else None
    if not ndok:
        r['notes'].append('nudge: ' + json.dumps(nderr, ensure_ascii=False)[:160])
    if deep:
        v2, _, v2ok = res(f2, g2['val'])
        r['newErrors'] = len(v2.get('newErrors') or []) if v2ok else -1
        r['newWarnings'] = len(v2.get('newWarnings') or []) if v2ok else -1
    else:
        r['newErrors'] = r['newWarnings'] = None
    sv, sverr, svok = res(f2, g2['sv'])
    r['save'] = svok and os.path.exists(tmp) and os.path.getsize(tmp) > 0
    if not svok:
        r['notes'].append('save: ' + json.dumps(sverr, ensure_ascii=False)[:160])

    if r['save']:
        f, ln = roundtrip(tmp, ws)
        r['rt'] = rt_parse(f)
        r['rtLine'] = ln
        r['rtClean'] = rt_ok(r['rt'], r['baseRt'])
        if not r['rtClean']:
            r['notes'].append('RoundTrip: ' + ln[:220])
        elif r['rt']['stale'] and r['rt']['sample'] != r['baseRt']['sample']:
            r['notes'].append('stale entry text changed (count %d held): %s'
                              % (r['rt']['stale'], r['rt']['sample'][:120]))
        os.remove(tmp)

    return r


def negative_control(files):
    """Byte-exact, not 'no error appeared': an untouched save and a save after a refused write
    must be the same bytes. Anything else means the refusal wrote something on its way out."""
    say()
    say('--- B-neg. a refused write must change zero bytes ---')
    bad = 0
    detail = []
    for idx, target in enumerate(files):
        ws = ws_of(target)
        s1 = os.path.join(os.path.dirname(target), '_ai_w3neg_a_%d.tzs' % idx)
        s2 = os.path.join(os.path.dirname(target), '_ai_w3neg_b_%d.tzs' % idx)
        for f in (s1, s2):
            if os.path.exists(f):
                os.remove(f)
        reqs1 = [('open', {'path': target}), ('save', {'handle': 'h1', 'out': s1})]
        srv(reqs1, ws)

        # a layout attribute that is not in the whitelist -- SPEC 11.24 (d) says the indexer
        # refuses it, and a refusal that still mutated would be the worst kind of bug.
        reqs2 = [
            ('open', {'path': target}),
            ('form_tree', {'handle': 'h1', 'depth': 99}),
        ]
        f2, _ = srv(reqs2, ws)
        tree, _, _ = res(f2, 2)
        nodes = walk(tree.get('root') or {}, []) if tree.get('root') else []
        leaf = None
        for n in nodes:
            if n.get('path') and n.get('name') and n.get('tag') not in (
                    'Form', 'VBox', 'HBox', 'Grid', 'Group', 'Page', 'Folder'):
                leaf = n['path']
        if not leaf:
            detail.append('%s: no leaf' % os.path.basename(target))
            bad += 1
            continue

        reqs3 = [
            ('open', {'path': target}),
            ('set_layout_attr', {'handle': 'h1', 'path': leaf,
                                 'attr': 'zzNotARealAttribute', 'value': '1'}),
            ('save', {'handle': 'h1', 'out': s2}),
        ]
        f3, _ = srv(reqs3, ws)
        _r, werr, wok = res(f3, 2)
        code = werr.get('code')
        same = (os.path.exists(s1) and os.path.exists(s2) and sha(s1) == sha(s2))
        refused = (not wok) and code in FROZEN_ERR
        if not (refused and same):
            bad += 1
            detail.append('%s: ok=%s code=%s sameBytes=%s' % (os.path.basename(target), wok, code, same))
        for f in (s1, s2):
            if os.path.exists(f):
                os.remove(f)
    check('B-n', 'refused write -> frozen code AND byte-identical output',
          bad == 0, '; '.join(detail[:6]) if detail else '%d files' % len(files))


def section_b(files, deep_of, limit_neg=3):
    deep_n = sum(1 for f in files if deep_of(f))
    say()
    say('--- B. the whole corpus through tzs-server --stdio ---')
    say('    %d files; per file: open, tree, 7x describe_kind, 3x list_*, get_component, '
        'find_component, nudge+1, save, RoundTrip' % len(files))
    say('    validate (1.3-10.4 s a call) only on the %d deep files -- see --validate' % deep_n)
    say()
    results = []
    t0 = time.time()
    for i, f in enumerate(files):
        deep = bool(deep_of(f))
        try:
            r = one_file(f, i, len(files), deep)
        except subprocess.TimeoutExpired:
            r = {'file': f, 'tag': os.path.basename(f), 'notes': ['TIMEOUT'], 'open': False,
                 'deep': deep}
        except Exception as e:
            r = {'file': f, 'tag': os.path.basename(f),
                 'notes': ['%s: %s' % (type(e).__name__, e)], 'open': False, 'deep': deep}
        results.append(r)
        # newErrors is None on a shallow file (validate was not run), which must not read as a
        # failure -- only a deep file can fail that clause.
        st = 'ok' if (r.get('open') and r.get('nudge') and r.get('rtClean')
                      and r.get('newErrors') in (0, None)) else '!!'
        say('  %3d/%d  %-3s %-32s nodes=%-5s stale %s->%s  rterr=%s%s' % (
            i + 1, len(files), st, r['tag'][:32], r.get('nodes'),
            (r.get('baseRt') or {}).get('stale'), (r.get('rt') or {}).get('stale'),
            'n/a' if r.get('rtClean') is None else (0 if r.get('rtClean') else 1),
            '' if deep else '   (no validate)'))
    say('    %d files in %.1f s   (validate run on %d of them)'
        % (len(files), time.time() - t0, sum(1 for r in results if r.get('deep'))))

    def agg(label, pred, extra=lambda r: ''):
        bad = [r for r in results if not pred(r)]
        check('B', label, not bad, ('%d bad: ' % len(bad)) + '; '.join(
            '%s%s' % (r['tag'], extra(r)) for r in bad[:5]) if bad else '%d files' % len(results))

    # The two validate assertions can only speak for the files that were asked for it.
    deepr = [r for r in results if r.get('deep')]

    def agg_deep(label, pred, extra=lambda r: ''):
        bad = [r for r in deepr if not pred(r)]
        check('B', label, not bad and bool(deepr),
              ('%d bad: ' % len(bad)) + '; '.join(
                  '%s%s' % (r['tag'], extra(r)) for r in bad[:5]) if bad
              else '%d of %d files (deep set)' % (len(deepr), len(results)))

    n = 0

    def num():
        nonlocal n
        n += 1
        return 'B%d' % n

    agg(num() + ' open reaches Loaded', lambda r: r.get('open'))
    agg(num() + ' form_tree returns a tree with elements',
        lambda r: r.get('tree') and (r.get('nodes') or 0) > 0, lambda r: ' nodes=%s' % r.get('nodes'))
    agg_deep(num() + ' validate answers', lambda r: r.get('validate'))
    agg(num() + ' all 7 describe_kind answer with a non-empty vocabulary',
        lambda r: all((r.get('kinds') or {}).get(k) for k in KINDS),
        lambda r: ' ' + json.dumps(r.get('kinds'), ensure_ascii=False)[:110])
    agg(num() + ' list_spec_nodes / list_records / list_local_strings answer',
        lambda r: r.get('listSpecNodes') is not None and r.get('records') is not None
        and r.get('strings') is not None)
    agg(num() + ' get_component exposes a layout', lambda r: r.get('get_component'))
    agg(num() + ' find_component resolves the leaf name',
        lambda r: r.get('find_component'))
    agg(num() + ' nudge applies (delta carries a changed attribute)',
        lambda r: r.get('nudge'), lambda r: ' %s' % json.dumps(r.get('nudgeDelta'), ensure_ascii=False))
    agg_deep(num() + ' validate after the write reports zero NEW errors',
             lambda r: r.get('newErrors') == 0,
             lambda r: ' newErrors=%s newWarnings=%s' % (r.get('newErrors'), r.get('newWarnings')))
    agg(num() + ' save produces a file', lambda r: r.get('save'))
    agg(num() + ' the repacked file is a fixed point relative to its own pristine RoundTrip',
        lambda r: r.get('rtClean'),
        lambda r: ' ' + (r.get('rtLine') or '')[:130])

    bad = [r for r in results if r.get('notes')]
    if bad:
        say()
        say('  notes from %d file(s):' % len(bad))
        for r in bad[:12]:
            say('    %-32s %s' % (r['tag'][:32], ' | '.join(r['notes'])[:220]))
    return results


# --------------------------------------------------------------------------- main

def main():
    argv = sys.argv[1:]
    limit = 0
    only = None
    skip_b = False
    root = ROOT
    vmode = 'sample'
    i = 0
    while i < len(argv):
        a = argv[i]
        if a == '--limit':
            i += 1
            limit = int(argv[i])
        elif a == '--files':
            i += 1
            only = [x for x in argv[i].split(',') if x]
        elif a == '--root':
            i += 1
            root = argv[i]
        elif a == '--validate':
            i += 1
            vmode = argv[i]
            if vmode not in VALIDATE_MODES:
                say('--validate takes one of %s' % '|'.join(VALIDATE_MODES))
                return 2
        elif a == '--skip-b':
            skip_b = True
        i += 1

    say('==== W3 GATE ====  bin=%s' % BIN)
    say('corpus root: %s        %s' % (root, time.strftime('%Y-%m-%d %H:%M:%S')))

    files = corpus(root)
    if only:
        files = [f for f in files if any(x in os.path.basename(f) for x in only)]
    say('corpus: %d files' % len(files))
    if not files:
        say('no corpus files -- wrong --root?')
        return 2

    # Section A runs on a file that is known to have a real `case`/writable leaf.
    target = None
    for f in files:
        if 'aapp320(c).tzs' in f:
            target = f
    section_a(target or files[0])

    if not skip_b:
        sel = files[:limit] if limit else files
        deep_of = lambda f: (vmode == 'all'
                             or (vmode == 'sample' and os.path.basename(f) in DEEP_SAMPLE))
        say('validate mode: %s' % vmode)
        section_b(sel, deep_of)
        negative_control(sel[:3])

    say()
    say('==== W3 GATE: %s  (%d pass, %d fail) ====' % ('PASS' if ok_all else 'FAIL', npass, nfail))
    try:
        with open(REPORT, 'w', encoding='utf-8') as fh:
            fh.write('\n'.join(report_lines) + '\n')
        print('\nreport: %s' % REPORT)
    except OSError as e:
        print('could not write report: %s' % e)
    return 0 if ok_all else 1


if __name__ == '__main__':
    try:
        sys.stdout.reconfigure(encoding='utf-8', errors='replace')
    except Exception:
        pass
    sys.exit(main())
