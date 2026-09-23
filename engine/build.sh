#!/usr/bin/env bash
# Builds every program in .tzs-cli.
#
# The csc invocation is long because .NET Framework 4.0 keeps WPF in the GAC, and
# PresentationCore lives in GAC_64 while everything else is in GAC_MSIL. Doing this by
# hand is easy to get wrong, hence the script.
#
#   ./build.sh              build all
#   ./build.sh AddField     build just one (name = test/<name>.cs)
#
# Output goes to $OUT (default: the temp probe dir). Override with OUT=... ./build.sh
#
# ---------------------------------------------------------------------------
# Two-library layout (Wave 1)
# ---------------------------------------------------------------------------
#   src/*.cs          -> TzsCli.dll          pure text/zip layer: System.Xml.Linq +
#                                            System.IO.Compression, nothing else.
#   src/Designer/**   -> TzsCli.Designer.dll reflection pipeline + bootstrap + session,
#                                            compiled against -r:TzsCli.dll and Newtonsoft.
#                                            One-way: see the CAVEAT below -- recompiling
#                                            src/*.cs INTO it gives two copies of the TzsCli
#                                            types and every `ref` program gets CS0433.
#   test/*.cs         -> one .exe per program: Main + that program's own logic only.
#
# Both libraries are built into $OUT before any program. A program in `ref` mode gets
#   -r:$OUT\TzsCli.dll  -r:$OUT\TzsCli.Designer.dll
# and finds them next to its own .exe at run time, so no AssemblyResolve rule is needed.
# There is deliberately no .csproj: MSBuild is not part of this toolchain.
#
# TzsCli.dll is checked after every build: if a designer assembly name turns up in its
# reference set the build FAILS and the DLL is deleted. That check is the mechanical
# enforcement of "the pure layer stays pure" -- without it a designer type leaks in
# silently, TzsCli.dll stops being buildable/testable without the designer, and the
# runtime AssemblyResolve handler starts pulling designer DLLs into the default load
# context. csc drops a -r: no type is taken from, so both the requested and the emitted
# reference set are checked.
#
# ---------------------------------------------------------------------------
# AssemblyVersion rule  (violating it is a hard csc error: CS0579)
# ---------------------------------------------------------------------------
#   [assembly: AssemblyVersion("1.0.0.251")]
# belongs in EVERY exe and in NO library. SettingManager.Version reads
# Assembly.GetEntryAssembly(), so an exe is the only place the attribute is meaningful,
# and two copies compiled into one assembly is a duplicate-attribute error. The programs
# that run in `link` mode today compile src/*.cs into themselves, so a program must carry
# the attribute in its own test/*.cs before it is migrated to `ref`.

set -u
HERE_POSIX="$(cd "$(dirname "$0")" && pwd)"
HERE="$(cygpath -w "$HERE_POSIX")"          # csc is a Windows binary: it wants a Windows path
CSC="/c/Windows/Microsoft.NET/Framework64/v4.0.30319/csc.exe"

# Default to engine/out, beside the sources. The old default was a shared temp probe directory,
# which made sense when this tree WAS the project and several agents built into it at once. Here
# the artifacts get collected into tt's package, so they belong somewhere stable and per-project.
# Still overridable with OUT=...
if [ -z "${OUT:-}" ]; then
  OUT="$HERE_POSIX/out"
  mkdir -p "$OUT"
fi
OUTW="$(cygpath -w "$OUT")"

GACMS='C:\Windows\Microsoft.NET\assembly\GAC_MSIL'
GAC64='C:\Windows\Microsoft.NET\assembly\GAC_64'

# The designer's install directory -- for BUILD TIME only.
#
# TzsCli.Designer.dll needs Newtonsoft.Json at compile time (SpecFn.cs parses the request
# payload with JObject) and this is where it lives -- v9.0.0.0. That is a real third-party
# library, not a designer assembly, so referencing it does not compromise TzsCli.dll's purity;
# only the designer library gets this -r:.
#
# At RUN time nothing reads this. The engine resolves its designer from TZSCLI_INSTALL if that
# is set, and otherwise from <its own directory>\designer -- the copy the package ships with it
# (Bootstrap.cs; engine/BUILD.md explains why the version is pinned by the package). Here it is
# purely the compile-time -r: above, and the probe programs under test/ also read it at run time
# when they are run out of engine/out/ -- which is not a package and has no designer/.
INSTALL="${TZSCLI_INSTALL:-}"

# Fail before compiling, with a sentence that names the fix. Without this the csc for
# TzsCli.Designer.dll reports `error CS0006: Metadata file '...\Newtonsoft.Json.dll' could not be
# found` against a path the reader has never seen -- and the other twelve units still build, so
# the run ends with "some builds failed" and a success-looking list above it.
#
# There is deliberately NO default: a designer lives wherever its owner installed it, so any
# built-in answer would be right on exactly one machine.
if [ -z "$INSTALL" ]; then
  echo "TZSCLI_INSTALL is not set -- the designer build needs an installed T100 designer." >&2
  echo "  export TZSCLI_INSTALL='D:\\APPS\\T100设计器_<版本>_免安装'" >&2
  echo "  engine/BUILD.md explains why; a checkout cannot carry the designer itself." >&2
  exit 1
fi
if [ ! -d "$INSTALL" ]; then
  echo "TZSCLI_INSTALL is not a directory: $INSTALL" >&2
  exit 1
fi
if [ ! -f "$INSTALL/Newtonsoft.Json.dll" ]; then
  echo "Not a designer directory: $INSTALL" >&2
  echo "  expected Newtonsoft.Json.dll in it (the designer ships it; see engine/BUILD.md)." >&2
  exit 1
fi

REFS=(
  "$GACMS\\PresentationFramework\\v4.0_4.0.0.0__31bf3856ad364e35\\PresentationFramework.dll"
  "$GAC64\\PresentationCore\\v4.0_4.0.0.0__31bf3856ad364e35\\PresentationCore.dll"
  "$GACMS\\WindowsBase\\v4.0_4.0.0.0__31bf3856ad364e35\\WindowsBase.dll"
  "$GACMS\\System.Xaml\\v4.0_4.0.0.0__b77a5c561934e089\\System.Xaml.dll"
  "$GACMS\\System.IO.Compression\\v4.0_4.0.0.0__b77a5c561934e089\\System.IO.Compression.dll"
  "$GACMS\\System.IO.Compression.FileSystem\\v4.0_4.0.0.0__b77a5c561934e089\\System.IO.Compression.FileSystem.dll"
)
RFLAGS=()
for r in "${REFS[@]}"; do RFLAGS+=(-r:"$r"); done

# TzsCli.dll gets the zip half of that list and no WPF. Filtered out of REFS rather than
# spelled again, so the GAC paths stay declared in exactly one place.
LIBRFLAGS=()
for r in "${REFS[@]}"; do case "$r" in *System.IO.Compression*) LIBRFLAGS+=(-r:"$r");; esac; done

LIB=( "$HERE\\src\\ElementIndex.cs" "$HERE\\src\\RecordRebuilder.cs"
      "$HERE\\src\\FormWriter.cs"    "$HERE\\src\\TzsRepacker.cs" )

# Assemblies TzsCli.dll must never reference. Names, not paths: that is what shows up in
# an assembly reference list, and it is what AssemblyResolve would resolve at run time.
DESIGNER_ASMS=( SpecDesignerCommon SpecDesigner.FormEditor SpecDesigner.Controls UndoRedoFramework )

# name : how the program gets the shared code
#   link -> compile src/*.cs into it    (transitional; programs not yet migrated)
#   ref  -> -r: TzsCli.dll + TzsCli.Designer.dll out of $OUT
#   none -> neither (programs that only need the designer DLLs from the install dir)
#
# CAVEAT, now resolved (architecture call, 2026-09-19). TzsCli.Designer.dll was originally
# compiled from src/*.cs + src/Designer/*.cs, which gave it its own copy of the four public
# TzsCli types. A `ref` program using one of them and referencing BOTH libraries was then a
# duplicate-type error:
#   CS0433: The type 'TzsCli.ElementIndex' exists in both 'TzsCli.dll' and 'TzsCli.Designer.dll'
# The dependency now runs one way only -- TzsCli.Designer is built from src/Designer/** against
# -r:TzsCli.dll, and TzsCli references nothing. Two agents found this independently (W1-A by
# reproducing CS0433, W1-B by noting the combined artifact carries both type sets), which is
# what settled it. Do not reintroduce ${LIB[@]} into the designer library's compile line.
declare -A LINK_SRC=(
  [TestRebuild]=link [TestWriter]=link [TestRepack]=link [VerifyRepack]=link
  [E2E]=link [MakeTestFile]=link [AddField]=link [Edit]=link
  [CheckOut]=none [RoundTrip]=none [Probe]=none
  [tzs-server]=ref [tzs-cli]=ref
)

mkdir -p "$OUT"
only="${1:-}"
fail=0
LAST_CSC_ARGS=()      # what the last compile() handed csc -- the purity check reads this

# One machine-readable line per unit, so a report can quote which bytes it ran.
built_line() {                              # built_line <name> <file>
  printf 'BUILT   %-14s %s %s\n' "$1" "$(sha256sum "$2" | cut -c1-16)" "$(stat -c%s "$2")"
}

# Compile <name> <out-file> <csc args...>; returns non-zero when the unit did not build.
# BUILT_QUIET=1 suppresses the BUILT line -- TzsCli.dll uses it so that no BUILT line is
# printed for an artifact the purity guard is about to delete.
compile() {
  local name="$1" exe="$2"; shift 2
  local raw rc out
  LAST_CSC_ARGS=( "$@" )

  # Delete the target first. csc leaves the previous binary untouched when it fails, and a
  # stale exe emitting plausible output is the easiest way to "verify" a change that never
  # built. "$OUT" is shared by every program and accumulates across sessions, so this is not
  # hypothetical -- there was a case-colliding addfield.exe sitting next to AddField.exe.
  rm -f "$exe"

  raw="$("$CSC" -nologo -out:"$(cygpath -w "$exe")" "$@" 2>&1)"
  rc=$?
  out="$(printf '%s\n' "$raw" | grep -v 'warning CS' | grep . || true)"

  if [ "$rc" != 0 ] || [ ! -f "$exe" ]; then
    echo "FAILED  $name  (csc rc=$rc)"
    [ -n "$out" ] && printf '%s\n' "$out" | head -5
    fail=1
    return 1
  fi
  [ "${BUILT_QUIET:-0}" = 1 ] || built_line "$name" "$exe"
  return 0
}

# Purity probe for TzsCli.dll. Prints "" when clean, else one reason line prefixed
# `leak:` (a designer assembly is reachable from TzsCli.dll) or `probe:` (we could not get
# a reference list at all, so a leak could hide behind the blind spot).
pure_leak() {                               # pure_leak <dll> <requested -r: args...>
  local dll="$1"; shift
  local a base n names

  # 1. What the invocation asks for. csc discards a -r: that no type is used from, so a
  #    designer reference added to the line below would leave no trace in the manifest.
  for a in "$@"; do
    base="${a##*\\}"; base="${base##*/}"; base="${base%.dll}"
    for n in "${DESIGNER_ASMS[@]}"; do
      [ "$base" = "$n" ] && { echo "leak:-r: $n"; return; }
    done
  done

  # 2. What the built assembly actually carries.
  names="$(powershell -NoProfile -Command \
    "[System.Reflection.Assembly]::ReflectionOnlyLoadFrom('$(cygpath -w "$dll")').GetReferencedAssemblies() | % Name" 2>&1 \
    | tr -d '\r')"
  for n in "${DESIGNER_ASMS[@]}"; do
    printf '%s\n' "$names" | grep -qxF "$n" && { echo "leak:referenced $n"; return; }
  done
  printf '%s\n' "$names" | grep -qxF mscorlib || { echo "probe:$(printf '%s\n' "$names" | grep . | head -1)"; return; }

  echo ""
}

# ---------------------------------------------------------------------------
# Libraries first: nothing in `ref` mode can compile before these exist.
# ---------------------------------------------------------------------------
# TzsCli.dll is compiled quietly: it has not "built" until the purity guard says so, and a
# BUILT line for an artifact the guard is about to delete is exactly the kind of plausible
# evidence this script exists to refuse.
BUILT_QUIET=1
compile TzsCli.dll "$OUT/TzsCli.dll" -target:library "${LIBRFLAGS[@]}" "${LIB[@]}"
librc=$?
BUILT_QUIET=0

if [ "$librc" = 0 ]; then
  # LAST_CSC_ARGS, not LIBRFLAGS: the guard must see every -r: csc was actually given.
  reason="$(pure_leak "$OUT/TzsCli.dll" "${LAST_CSC_ARGS[@]}")"
  case "$reason" in
    leak:*)
      echo "FAILED  TzsCli.dll  (designer reference leaked: ${reason#leak:})"
      # An impure TzsCli.dll must not survive: a `ref` program would happily link it.
      rm -f "$OUT/TzsCli.dll"
      fail=1 ;;
    probe:*)
      echo "FAILED  TzsCli.dll  (purity probe failed: ${reason#probe:})"
      fail=1 ;;
    *)  built_line TzsCli.dll "$OUT/TzsCli.dll" ;;
  esac
fi
# src/Designer/ (and any Fns/ below it) is written by a parallel agent -- absent or empty
# means "not ready", which must not block anything else.
DESIGNER_SRC=()
if [ -d "$HERE_POSIX/src/Designer" ]; then
  while IFS= read -r f; do
    [ -n "$f" ] && DESIGNER_SRC+=( "$(cygpath -w "$f")" )
  done < <(find "$HERE_POSIX/src/Designer" -type f -name '*.cs' | LC_ALL=C sort)
fi

if [ "${#DESIGNER_SRC[@]}" -eq 0 ]; then
  echo "skip    TzsCli.Designer.dll (src/Designer not ready)"
else
  # References TzsCli.dll, does NOT recompile src/*.cs into itself -- see the CAVEAT above.
  # Newtonsoft is only needed if SpecFn.cs (JObject) is present; the -r: is harmless otherwise.
  compile TzsCli.Designer.dll "$OUT/TzsCli.Designer.dll" -target:library \
    "${RFLAGS[@]}" -r:"$OUTW\\TzsCli.dll" -r:"$INSTALL\\Newtonsoft.Json.dll" "${DESIGNER_SRC[@]}"
fi

# ---------------------------------------------------------------------------
# Programs.
# ---------------------------------------------------------------------------
for name in "${!LINK_SRC[@]}"; do
  if [ -n "$only" ] && [ "$only" != "$name" ]; then continue; fi
  src="$HERE\\test\\$name.cs"
  if [ ! -f "$(cygpath -u "$src")" ]; then echo "skip    $name (no test/$name.cs)"; continue; fi

  files=()
  case "${LINK_SRC[$name]}" in
    link) files=( "${LIB[@]}" ) ;;
    ref)  files=( -r:"$OUTW\\TzsCli.dll" -r:"$OUTW\\TzsCli.Designer.dll" ) ;;
    none) ;;
    *)    echo "FAILED  $name  (unknown LINK_SRC mode '${LINK_SRC[$name]}')"; fail=1; continue ;;
  esac

  compile "$name" "$OUT/$name.exe" "${RFLAGS[@]}" "${files[@]}" "$src"
done

echo
if [ "$fail" = 0 ]; then echo "all built -> $OUT"; else echo "some builds failed"; fi
echo
echo "runtime note: the shipped engine loads its designer from <its own dir>\\designer -- the copy"
echo "  the package carries with it. Out of engine/out/ (which is not a package) the probe"
echo "  programs fall back to TZSCLI_INSTALL instead:"
echo "  INSTALL = $INSTALL"
echo "  AddField / E2E / MakeTestFile also load SpecDesigner.FormEditor.dll from there."
