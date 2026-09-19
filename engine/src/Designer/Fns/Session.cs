using System;
using System.Collections.Generic;
using System.IO;
using Newtonsoft.Json.Linq;
using TzsCli;
using DS = TzsCli.Designer.Session;

namespace TzsCli.Designer.Fns
{
    /// <summary>
    /// The session group: `open` / `save` / `close` / `list_open` (SPEC §11.24 (b), (g), (g-1)).
    ///
    /// WHAT THIS FILE ALSO OWNS, AND WHY IT IS NOT IN Session.cs
    ///
    /// <para>
    /// Two pieces of per-handle state have no home: SpecDesignerCommon's <c>Session</c>
    /// (src/Designer/Session.cs, frozen by W1-B) has no field for either, and it is not this
    /// agent's file. Both live here instead, keyed by handle -- never by ProgramKey, because
    /// four files in the corpus share one ProgramKey (SPEC §11.24 (g)).
    /// </para>
    ///
    /// <list type="number">
    /// <item><b>The layout writer.</b> <see cref="Layout"/> returns the per-handle
    /// <c>FormWriter</c> built from the package's own <c>.4fd</c> text. <c>save</c> renders it,
    /// and every mutating function (Fns/Attr.cs) must edit through it rather than making its own
    /// -- a second writer would accumulate edits that `save` never sees. Creating it lazily from
    /// the file is what makes `save` work on a handle that was never mutated: with no edits
    /// applied, <c>FormWriter.Render()</c> returns the input text byte-for-byte, which is the
    /// RoundTrip property.</item>
    ///
    /// <item><b>The register-live-writers step, i.e. the Loaded/Mutable split.</b> See
    /// <see cref="Touch"/> below; the reasoning is spelled out there because the contract calls
    /// the placement question out explicitly.</item>
    /// </list>
    ///
    /// <para>Everything else is a thin transliteration of the frozen contract: <c>open</c> is
    /// <c>Session.Open</c> (which already owns the duplicate-key refusal and the failed-load
    /// rollback -- re-implementing either here would be a second, drifting copy),
    /// <c>save</c> is <c>Save.Run</c>, <c>close</c> is <c>Session.Close</c>, and
    /// <c>list_open</c> is the registry rendered back out.</para>
    /// </summary>
    public static class Session
    {
        /// <summary>
        /// Every handle this process has opened and not yet closed, keyed by handle.
        ///
        /// Kept here rather than read out of Session's private <c>_open</c> because that map is
        /// keyed by PackageKey: in a collision cycle it says which *key* is live, not which
        /// handle the caller holds, and §11.24 (g-1) warns that close(A) may in fact have
        /// removed B. The handle -> session map is the one view that cannot lie about that.
        /// </summary>
        static readonly Dictionary<string, DS> _handles = new Dictionary<string, DS>(StringComparer.Ordinal);

        /// <summary>Per-handle .4fd writers, created lazily by <see cref="Layout"/>.</summary>
        static readonly Dictionary<string, FormWriter> _layout = new Dictionary<string, FormWriter>(StringComparer.Ordinal);

        static readonly object _gate = new object();

        public static void Register(IDictionary<string, Fn> into) {
            into["open"]      = Open;
            into["save"]      = SaveFn;
            into["close"]     = CloseFn;
            into["list_open"] = ListOpen;
        }

        // ------------------------------------------------------------------ the mutation gate

        /// <summary>
        /// Loaded -> Mutable. Every mutating function calls this as its first statement.
        ///
        /// WHERE THIS LANDED, AND WHY
        ///
        /// Measured against the two other files that were written in parallel (neither of which
        /// this agent touched):
        ///
        /// * <b>Rpc.cs does NOT do it.</b> Its dispatch has no branch on <c>SpecFn.Mutating</c>;
        ///   it resolves the handle and calls the body. So the one-line dispatcher placement the
        ///   flag exists for (`if (fn.Mutating) Touch(s);`) is available and unused.
        /// * <b>Fns/Attr.cs does it itself.</b> Its <c>Urm(Session)</c> helper calls
        ///   <c>s.RegisterUndoRedo()</c> and then fetches the manager out of undoRedoManagerMap,
        ///   and every write path in that file goes through <c>Urm</c>.
        ///
        /// So the transition already happens on the mutating side, and nothing needed to change in
        /// another agent's file to get it. This method is the same hook, kept here for the modules
        /// that have no helper of their own -- <c>set_excluded</c> in Fns/Validate.cs is its only
        /// caller today -- and it is deliberately never called by anything read-only.
        ///
        /// The dispatcher form is still strictly better if W2-A can spare the line, because a
        /// forgotten call is silent: mutating without an UndoRedoManager does not throw at the
        /// mutation, it throws later inside a designer command that resolves its manager by key
        /// ("No UndoRedoManager"), or not at all.
        ///
        /// Idempotent: Session.RegisterUndoRedo is guarded by the Mutable flag, and a second
        /// manager would orphan the first while every undo command still held a reference to it.
        /// </summary>
        public static void Touch(DS s) {
            if (s == null) throw TzsError.Validation("缺少 handle");
            if (s.Closed) throw TzsError.NotFound("句柄（已关闭）", s.Handle);
            s.RegisterUndoRedo();
        }

        /// <summary>
        /// Handle -> Session. The dispatcher resolves the session it hands to the Fn delegate;
        /// this is for the modules that only have the string, and for the fns here that are also
        /// usable with no dispatcher at all (s == null).
        /// </summary>
        public static DS Get(string handle) {
            if (string.IsNullOrEmpty(handle)) throw TzsError.Validation("缺少 handle");
            DS s;
            lock (_gate) { _handles.TryGetValue(handle, out s); }
            if (s == null) throw TzsError.NotFound("句柄", handle);
            if (s.Closed) throw TzsError.NotFound("句柄（已关闭）", handle);
            return s;
        }

        /// <summary>
        /// The per-handle .4fd writer, built on first use from the bytes of the package this
        /// handle was opened on.
        ///
        /// Mutating functions must edit through this instance. `save` renders this instance.
        /// Anything else and the edits and the save are looking at two different documents;
        /// FormWriter keeps its edits inside itself, so a second writer is not a second view of
        /// the same document, it is a second document.
        ///
        /// Reading the file again per call is deliberate: it is the package as opened, and the
        /// handle's own Path never changes even when `save` writes elsewhere.
        /// </summary>
        public static FormWriter Layout(DS s) {
            if (s == null) throw TzsError.Validation("缺少 handle");
            lock (_gate) {
                FormWriter w;
                if (_layout.TryGetValue(s.Handle, out w)) return w;
                byte[] zip = File.ReadAllBytes(s.Path);
                string fdEntry = Reflect.EntryName(zip, ".4fd");
                if (fdEntry == null)
                    throw new TzsError("internal", "包里没有 .4fd 条目: " + s.Path);
                w = FormWriter.Load(Reflect.EntryText(zip, fdEntry));
                _layout[s.Handle] = w;
                return w;
            }
        }

        // ------------------------------------------------------------------ fns

        /// <summary>
        /// `open path [timeout]` -> handle info. Never validates: the designer's validators cost
        /// 1.6 s on a 114-element form and 10.4 s on a 670-element one, and open is on the
        /// critical path of every conversation (SPEC §11.24 (c) `slow`, §11.24 (i)).
        ///
        /// The duplicate-key refusal lives in Session.Open and is deliberately not repeated: it
        /// refuses when the key is already held by a live handle *or* by a stale tzpMap entry
        /// left by an earlier failed load, and it rolls a failed load back. `force` is not
        /// offered -- §11.24 (g) allows it only for Loaded handles, and a caller can already
        /// achieve exactly that by closing first.
        /// </summary>
        static object Open(DS ignored, JObject args) {
            string path = Str(args, "path");
            if (string.IsNullOrEmpty(path)) throw TzsError.Validation("open 需要一个 path");

            // Manifest.cs declares `force`, and §11.24 (g) gives it a meaning: take over an
            // occupant that is still only Loaded, refuse a Mutable one. It is NOT implemented
            // here, and the reason is the same reason the refusal itself is not implemented here:
            // Session.Open owns the occupancy check, and force means evicting that occupant --
            // exactly the "silent eviction" the contract forbids and this file must not
            // reintroduce. Saying so is the only honest option; ignoring the flag would hand the
            // caller an E_KEY_IN_USE with no hint that the flag they used does nothing.
            if (Bool(args, "force", false))
                throw new TzsError("not_implemented",
                    "force 尚未实现。占用者只能被显式关掉：先 close 占用它的那个句柄，再 open —— "
                    + "关掉之后同一个 key 可以重新加载（§11.24 (g-1) 已在真实包上实测）。");

            int timeout = Int(args, "timeout", 0);
            if (timeout <= 0) timeout = DS.TimeoutFromEnv();
            DS s = DS.Open(path, timeout);
            lock (_gate) _handles[s.Handle] = s;
            return Info(s);
        }

        /// <summary>
        /// `save handle outPath` -> Save.Run(original, outPath, w4, s).
        ///
        /// On a Loaded handle the writer has never been dirtied, so Render() returns the input
        /// .4fd text byte-for-byte and the entry is rewritten with identical bytes -- that is
        /// the fixed-point property (§11.24 (b)), not a special case coded here.
        ///
        /// Nothing about the handle changes: no UndoRedoManager, no state transition, which is
        /// why this is not a mutating fn.
        /// </summary>
        static object SaveFn(DS s0, JObject args) {
            DS s = Resolve(s0, args);
            // `out` is the name Manifest.cs declares and tzs-cli maps --out to. `outPath` is the
            // name SPEC §11.24 (b) uses in prose ("save → handle + outPath"). Accepting both is
            // what keeps the two documents from disagreeing at the one place they meet; the
            // canonical one is the manifest's.
            string outPath = Str(args, "out");
            if (string.IsNullOrEmpty(outPath)) outPath = Str(args, "outPath");
            if (string.IsNullOrEmpty(outPath)) throw TzsError.Validation("save 需要一个 out（输出 .tzs 路径）");

            byte[] original = File.ReadAllBytes(s.Path);
            FormWriter w4 = Layout(s);
            byte[] result = TzsCli.Designer.Save.Run(original, outPath, w4, s);
            return new Dictionary<string, object> {
                { "handle",     s.Handle },
                { "path",       s.Path },
                { "out",        outPath },
                { "bytesIn",    (long)original.Length },
                { "bytesOut",   (long)result.Length },
                { "layoutDirty", w4.Dirty },
                { "state",      State(s) },
            };
        }

        /// <summary>
        /// `close handle`. Tolerant by contract: §11.24 (g-1) measured that in a key-collision
        /// cycle close(A) actually removes B, because removal is per map key rather than per
        /// handle -- so "not found" is a legitimate outcome and is reported, not thrown. The
        /// handle string is never reused, so a caller that closes twice gets closed:false and
        /// can tell it was already gone.
        /// </summary>
        static object CloseFn(DS s0, JObject args) {
            string handle = Str(args, "handle");
            DS s = s0;
            lock (_gate) {
                if (s == null && !string.IsNullOrEmpty(handle)) _handles.TryGetValue(handle, out s);
            }
            if (s == null) {
                return new Dictionary<string, object> {
                    { "handle", handle },
                    { "closed", false },
                    { "state",  "Closed" },
                    { "note",   "该句柄不存在——可能已被一次 key 碰撞的 close 摘掉" },
                };
            }

            bool wasClosed = s.Closed;
            if (!wasClosed) s.Close();

            var dead = new List<string>();
            lock (_gate) {
                foreach (var kv in _handles)
                    if (kv.Value == s || (kv.Value.Key != null && kv.Value.Key.Equals(s.Key))) dead.Add(kv.Key);
                foreach (string h in dead) { _handles.Remove(h); _layout.Remove(h); }
                _layout.Remove(s.Handle);
            }
            // The cached validate baseline is the pristine state of a *file*. A reopened key is a
            // genuinely new load (§11.24 (g-1) item 3), so the cache must not survive it.
            Validate.Forget(s.Handle);

            return new Dictionary<string, object> {
                { "handle", s.Handle },
                { "closed", !wasClosed },
                { "key",    KeyString(s.Key) },
                { "state",  "Closed" },
            };
        }

        /// <summary>
        /// `list_open` -> one entry per live handle, `key` included so that a caller can see two
        /// files sharing one ProgramKey (which is precisely the collision §11.24 (g) exists for).
        /// </summary>
        static object ListOpen(DS ignored, JObject args) {
            List<DS> snapshot;
            lock (_gate) {
                snapshot = new List<DS>(_handles.Values);
            }
            snapshot.Sort(delegate(DS a, DS b) { return string.CompareOrdinal(a.Handle, b.Handle); });
            var list = new List<object>();
            foreach (DS s in snapshot) list.Add(Info(s));
            return list;
        }

        // ------------------------------------------------------------------ helpers

        /// <summary>Handle -> the session, preferring the one the dispatcher already resolved.</summary>
        internal static DS Resolve(DS s, JObject args) {
            DS got = s;
            if (got == null) got = Get(Str(args, "handle"));
            if (got.Closed) throw TzsError.NotFound("句柄（已关闭）", got.Handle);
            return got;
        }

        internal static Dictionary<string, object> Info(DS s) {
            return new Dictionary<string, object> {
                { "handle",  s.Handle },
                { "path",    s.Path },
                { "program", s.Program },
                { "key",     KeyString(s.Key) },
                { "state",   State(s) },
                { "mutated", s.Mutable },
                { "loadMs",  Math.Round(s.LoadMs, 1) },
            };
        }

        /// <summary>Loaded / Mutable / Closed (§11.24 (b)). Mutable is irreversible for the
        /// handle, so Close does not reset it -- `mutated` keeps reading true after a close,
        /// which is what a caller asking "did I change this?" wants.</summary>
        internal static string State(DS s) {
            if (s.Closed) return "Closed";
            return s.Mutable ? "Mutable" : "Loaded";
        }

        /// <summary>
        /// PackageKey as one readable string. Not the object itself: PackageKey carries only
        /// Program/PackType/Memo, none of which render as anything but a type name, and the
        /// point of exposing the key at all is that two colliding files produce the *same*
        /// string -- which a caller can see only if it is a string.
        /// </summary>
        internal static string KeyString(object key) {
            if (key == null) return null;
            return Reflect.S(Reflect.Prop(key, "Program")) + "|" + Reflect.S(Reflect.Prop(key, "PackType"));
        }

        internal static string Str(JObject a, string n) {
            if (a == null) return null;
            JToken t = a[n];
            if (t == null || t.Type == JTokenType.Null) return null;
            return t.ToString();
        }

        internal static int Int(JObject a, string n, int dflt) {
            string v = Str(a, n);
            int i;
            return v != null && int.TryParse(v, out i) ? i : dflt;
        }

        internal static bool Bool(JObject a, string n, bool dflt) {
            string v = Str(a, n);
            bool b;
            return v != null && bool.TryParse(v, out b) ? b : dflt;
        }
    }
}
