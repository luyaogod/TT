import os, sys

HERE = os.path.dirname(os.path.abspath(__file__))

def norm(p):
    p = p.strip()
    if p.startswith('/d/'):
        p = 'D:/' + p[3:]
    p = os.path.normcase(os.path.normpath(p.replace('\\', '/')))
    # The tsv was written by bash from `find`, so its non-ASCII bytes are in the console
    # code page, not UTF-8 -- reading them back mangles the Chinese filenames differently
    # than the manifest's. Drop non-ASCII so the comparison is about the file set, not
    # about which decoder won.
    return ''.join(c for c in p if ord(c) < 128)

man = set()
for ln in open(os.path.join(HERE, 'corpus.manifest'), encoding='utf-8'):
    if ln.startswith('#') or not ln.strip():
        continue
    man.add(norm(ln.split(None, 2)[2]))

ran = set()
for ln in open(os.path.join(HERE, 'batch-results.tsv'), encoding='utf-8', errors='replace'):
    f = ln.split('|')
    if len(f) > 13 and f[13].strip():
        ran.add(norm(f[13]))

print('manifest %d   ran %d' % (len(man), len(ran)))
print()
print('in manifest, not run:')
for p in sorted(man - ran):
    print('   ', p)
print()
print('run, not in manifest:')
for p in sorted(ran - man):
    print('   ', p)
