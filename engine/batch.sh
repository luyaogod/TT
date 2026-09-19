#!/usr/bin/env bash
# Batch fixed-point check over the whole .tzs corpus.
#
# For every package: can the designer load it headlessly, and does re-saving it through
# its own model reproduce exactly the same spec nodes and layout paths? One process per
# file, so a package that pops a modal dialog and hangs costs us its timeout and nothing
# more. A failing file also prints its first few differing names, so it is diagnosable
# from the TSV alone.
#
#   ./batch.sh                      whole default corpus
#   ./batch.sh /d/t100_wrok_dir     explicit root
#   ./batch.sh /d/t100_wrok_dir 60  per-file timeout in seconds
#
# Results land in batch-results.tsv next to this script.
# Columns: status, env, code_template, tsdAdds, tsdDrops, fdPathAdds, fdPathDrops,
#          tsdNodesIn, tsdNodesOut, layoutElems, sample, path, error

set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="${1:-/d/t100_wrok_dir}"
TMO="${2:-120}"
OUT="$HERE/batch-results.tsv"
EXE="/c/Users/18526/AppData/Local/Temp/dep_probe/RoundTrip.exe"

if [ ! -x "$EXE" ]; then echo "missing $EXE -- run ./build.sh first"; exit 1; fi

# Exclude the drivers' own scratch output. The three write-side drivers already did this;
# the read side did not, so its baseline silently included the _ai_*.tzs files left behind by
# earlier AddField runs -- which made the baseline unreproducible, since those files can be
# regenerated (or not) between runs.
mapfile -d '' FILES < <(find "$ROOT" -iname '*.tzs' ! -name '_ai*' ! -name '_dw*' ! -name '_ed*' ! -name '_del*' ! -name '_ac*' -print0 | sort -z)
TOTAL=${#FILES[@]}
echo "corpus: $ROOT"
echo "packages: $TOTAL   per-file timeout: ${TMO}s"
echo

: > "$OUT"
n=0 ok=0 fail=0
for f in "${FILES[@]}"; do
  n=$((n+1))
  w="$(cygpath -w "$f")"

  # The workspace is whichever ancestor holds an mta/ directory -- that is what makes a
  # module directory a workspace, and TzpManager refuses packages from outside the one it
  # was configured with. Deriving it beats hardcoding a per-corpus path list.
  ws=""; d="$(dirname "$f")"
  while [ "$d" != "/" ] && [ -n "$d" ]; do
    [ -d "$d/mta" ] && { ws="$d"; break; }
    d="$(dirname "$d")"
  done
  [ -z "$ws" ] && ws=/d/t100_wrok_dir/hengshuo/prd

  line="$(TZSCLI_QUIET=1 TZSCLI_WS="$(cygpath -w "$ws")" timeout "$TMO" "$EXE" "$w" 2>&1 | grep '^SUMMARY|' | tail -1)"
  [ -z "$line" ] && line="SUMMARY|TIMEOUT|||||||||||$w|exceeded ${TMO}s"

  echo "$line" >> "$OUT"
  IFS='|' read -r _ status env tpl ta td fa fd st ni no el sample path err <<< "$line"

  base="$(basename "$f")"
  if [ "$status" = "ok" ] && [ "$ta" = 0 ] && [ "$td" = 0 ] && [ "$fa" = 0 ] && [ "$fd" = 0 ]; then
    ok=$((ok+1))
    printf '%3d/%d  OK    %-28s env=%-2s tpl=%-4s nodes=%-6s stale=%s\n' "$n" "$TOTAL" "$base" "$env" "$tpl" "$ni" "$st"
  else
    fail=$((fail+1))
    printf '%3d/%d  FAIL  %-28s %s\n' "$n" "$TOTAL" "$base" "$err$sample"
  fi
done

echo
echo "=== $ok clean, $fail with differences or errors (of $TOTAL) ==="
echo "results: $OUT"
