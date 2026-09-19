#!/usr/bin/env bash
# Batch write-path check: for every package in the corpus, add a real field through
# UICreator and then ask the designer whether the result is still a fixed point.
#
# The read-side batch (batch.sh) proves the designer's model reproduces every package
# unchanged. This one proves the *write* side survives contact with every package's
# layout: FormWriter's text splice, UICreator's placement, the .tsd status promotion and
# the repack all have to hold on 10-element forms and 1297-element forms alike.
#
# Container type rotates through all six so each package exercises one branch.
# The (table, column) comes from the package's own table dictionary, picking a column that
# is not already a field -- see pick-column.py.
#
# The produced file has to live *inside* the workspace: TzpManager refuses packages
# outside the configured one, so scratch output goes next to the source and is removed.
#
#   ./batch-write.sh                  whole default corpus
#   ./batch-write.sh /d/t100_wrok_dir
#   ./batch-write.sh /d/t100_wrok_dir 3     # stop after the first N (smoke test)
#
# Results land in batch-write-results.tsv, tilde-separated. Tilde and not tab because
# Windows paths are full of backslashes, and every tab-escaping route here (echo -e,
# printf with a \t escape) has already mangled one in testing -- D:\t100_... silently
# became a tab, which cost 11 rows.

set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="${1:-/d/t100_wrok_dir}"
LIMIT="${2:-0}"
OUT="$HERE/batch-write-results.tsv"
BIN=/c/Users/18526/AppData/Local/Temp/dep_probe
TYPES=(None Grid Group ScrollGrid Table Tree)

for e in "$BIN/AddField.exe" "$BIN/RoundTrip.exe"; do
  [ -x "$e" ] || { echo "missing $e -- run ./build.sh first"; exit 1; }
done

# our own outputs and scratch files are not corpus
mapfile -d '' FILES < <(find "$ROOT" -iname '*.tzs' ! -name '_ai*' ! -name '_dw*' -print0 | sort -z)
TOTAL=${#FILES[@]}
echo "corpus: $ROOT"
echo "packages: $TOTAL   container types rotate: ${TYPES[*]}"
echo

: > "$OUT"
n=0 ok=0 fail=0 skip=0
for f in "${FILES[@]}"; do
  n=$((n+1))
  [ "$LIMIT" != 0 ] && [ "$n" -gt "$LIMIT" ] && break
  base="$(basename "$f")"

  # The workspace is whichever ancestor holds an mta/ directory -- that is what makes a
  # module directory a workspace, and TzpManager refuses packages from outside the one it
  # was configured with.
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

  # a column to add, from this package's own table dictionary
  pick="$(python "$HERE/pick-column.py" "$(cygpath -u "$wsw")\\$base" "$wsw" 2>/dev/null)"
  set -- $pick
  if [ $# -lt 2 ]; then
    skip=$((skip+1)); printf '%3d  SKIP  %-28s (no unused column)\n' "$n" "$base"
    printf 'skip~no unused column~%s\n' "$(cygpath -w "$f")" >> "$OUT"; continue
  fi
  tbl="$1"; col="$2"
  type="${TYPES[$(( (n-1) % 6 ))]}"

  tmp="$ws/_aiw_$$.tzs"
  addout="$(TZSCLI_RELOAD_TIMEOUT=45 TZSCLI_WS="$wsw" timeout 180 "$BIN/AddField.exe" "$(cygpath -w "$f")" "$(cygpath -w "$tmp")" "$type" "@auto" "$tbl" "$col" 2>&1)"
  rc=$?
  if [ "$rc" -ne 0 ] || [ ! -f "$tmp" ]; then
    # AddField prints an unhandled exception on failure; keep the first line, not the stack.
    emsg="$(printf '%s' "$addout" | grep -E 'Exception|error|not found|path not' | head -1 | cut -c1-110)"
    [ -z "$emsg" ] && emsg="$(printf '%s' "$addout" | tail -2 | head -1 | cut -c1-110)"
    fail=$((fail+1))
    printf '%3d  FAIL  %-28s %-11s %s.%s  (AddField) %s\n' "$n" "$base" "$type" "$tbl" "$col" "$emsg"
    printf 'addfail~%s~%s~%s.%s %s\n' "$type" "$(cygpath -w "$f")" "$tbl" "$col" "$emsg" >> "$OUT"
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
    printf '%3d  OK    %-28s %-11s %s.%s\n' "$n" "$base" "$type" "$tbl" "$col"
  else
    fail=$((fail+1))
    printf '%3d  FAIL  %-28s %-11s %s.%s  %s%s\n' "$n" "$base" "$type" "$tbl" "$col" "$err" "$sample"
  fi
  printf '%s~%s~%s~%s~%s~%s~%s~%s~%s~%s~%s~%s~%s~%s~%s\n' \
    "$status" "$env" "$type" "$ta" "$td" "$fa" "$fd" "$ni" "$no" "$el" "$sample" \
    "$(cygpath -w "$f")" "$tbl" "$col" "$err" >> "$OUT"
done

echo
echo "=== $ok clean, $fail failed, $skip skipped (of $TOTAL) ==="
echo "results: $OUT"
