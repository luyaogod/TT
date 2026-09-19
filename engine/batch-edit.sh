#!/usr/bin/env bash
# Batch edit-path check: for every package in the corpus, flip can_edit on a real field
# through the designer's own SpecAttributeUndoRedoCommand, then ask whether the result is
# still a fixed point.
#
# batch.sh proved the designer's model reproduces every package unchanged; batch-write.sh
# proved the add path survives every layout. This one does the same for modify: the spec
# attribute, the layout attribute it drives (noEntry), the .tsd status promotion to "u" and
# the repack all have to hold across 10-element and 1297-element forms alike.
#
#   ./batch-edit.sh                  whole default corpus
#   ./batch-edit.sh /d/t100_wrok_dir
#   ./batch-edit.sh /d/t100_wrok_dir 3     # stop after the first N (smoke test)
#
# Results land in batch-edit-results.tsv, tilde-separated (see batch-write.sh for why not tab).

set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="${1:-/d/t100_wrok_dir}"
LIMIT="${2:-0}"
OUT="$HERE/batch-edit-results.tsv"
BIN=/c/Users/18526/AppData/Local/Temp/dep_probe

for e in "$BIN/Edit.exe" "$BIN/RoundTrip.exe"; do
  [ -x "$e" ] || { echo "missing $e -- run ./build.sh first"; exit 1; }
done

mapfile -d '' FILES < <(find "$ROOT" -iname '*.tzs' ! -name '_ai*' ! -name '_dw*' ! -name '_ed*' ! -name '_del*' -print0 | sort -z)
TOTAL=${#FILES[@]}
echo "corpus: $ROOT"
echo "packages: $TOTAL   operation: flip field:can_edit on a data-bound element"
echo

: > "$OUT"
n=0 ok=0 fail=0 skip=0
for f in "${FILES[@]}"; do
  n=$((n+1))
  [ "$LIMIT" != 0 ] && [ "$n" -gt "$LIMIT" ] && break
  base="$(basename "$f")"

  ws=""; d="$(dirname "$f")"
  while [ "$d" != "/" ] && [ -n "$d" ]; do
    [ -d "$d/mta" ] && { ws="$d"; break; }
    d="$(dirname "$d")"
  done
  if [ -z "$ws" ]; then
    skip=$((skip+1)); printf '%3d  SKIP  %-28s (no workspace)\n' "$n" "$base"
    printf 'skip~no workspace~%s\n' "$(cygpath -w "$f")" >> "$OUT"; continue
  fi
  wsw="$(cygpath -w "$ws")"

  target="$(python "$HERE/pick-edit-target.py" "$(cygpath -u "$wsw")\\$base" 2>/dev/null)"
  set -- $target
  if [ $# -lt 3 ]; then
    skip=$((skip+1)); printf '%3d  SKIP  %-28s (no editable field)\n' "$n" "$base"
    printf 'skip~no editable field~%s\n' "$(cygpath -w "$f")" >> "$OUT"; continue
  fi
  epath="$1"; espec="$2"; evalue="$3"

  tmp="$ws/_edu_$$.tzs"
  edout="$(TZSCLI_RELOAD_TIMEOUT=45 TZSCLI_WS="$wsw" timeout 180 "$BIN/Edit.exe" "$(cygpath -w "$f")" "$(cygpath -w "$tmp")" set "$epath" "$espec" "$evalue" 2>&1)"
  rc=$?
  if [ "$rc" -ne 0 ] || [ ! -f "$tmp" ]; then
    emsg="$(printf '%s' "$edout" | grep -E 'Exception|error|not found|路径|超时' | head -1 | cut -c1-110)"
    [ -z "$emsg" ] && emsg="$(printf '%s' "$edout" | tail -2 | head -1 | cut -c1-110)"
    fail=$((fail+1))
    printf '%3d  FAIL  %-28s %-8s %s  (Edit) %s\n' "$n" "$base" "$espec" "$evalue" "$emsg"
    printf 'editfail~%s~%s~%s %s\n' "$espec" "$(cygpath -w "$f")" "$evalue" "$emsg" >> "$OUT"
    rm -f "$tmp"; continue
  fi

  # Baseline first: a few designer-authored packages are already a couple of posX values
  # away from what their own model would write, so the threshold has to be "no worse than
  # the source", not an absolute zero. Running the source through the same check measures it.
  stIn=$(TZSCLI_QUIET=1 TZSCLI_WS="$wsw" timeout 180 "$BIN/RoundTrip.exe" "$(cygpath -w "$f")" 2>&1 | grep '^SUMMARY|' | tail -1 | cut -d'|' -f9)
  case "$stIn" in ''|*[!0-9]*) stIn=0;; esac

  line="$(TZSCLI_QUIET=1 TZSCLI_WS="$wsw" timeout 180 "$BIN/RoundTrip.exe" "$(cygpath -w "$tmp")" 2>&1 | grep '^SUMMARY|' | tail -1)"
  rm -f "$tmp"
  [ -z "$line" ] && line="SUMMARY|TIMEOUT|||||||||||$(cygpath -w "$f")|exceeded"

  IFS='|' read -r _ status env tpl ta td fa fd st ni no el sample path err <<< "$line"
  if [ "$status" = "ok" ] && [ "$ta" = 0 ] && [ "$td" = 0 ] && [ "$fa" = 0 ] && [ "$fd" = 0 ] && [ "${st:-0}" -le "$stIn" ]; then
    ok=$((ok+1))
    printf '%3d  OK    %-28s %-8s %s\n' "$n" "$base" "$espec" "$evalue"
  else
    fail=$((fail+1))
    printf '%3d  FAIL  %-28s %-8s %s  %s%s\n' "$n" "$base" "$espec" "$evalue" "$err" "$sample"
  fi
  printf '%s~%s~%s~%s~%s~%s~%s~%s~%s~%s~%s~%s~%s %s\n' \
    "$status" "$env" "$espec" "$ta" "$td" "$fa" "$fd" "$ni" "$no" "$el" "$sample" \
    "$(cygpath -w "$f")" "$evalue" "$err" >> "$OUT"
done

echo
echo "=== $ok clean, $fail failed, $skip skipped (of $TOTAL) ==="
echo "results: $OUT"
