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
            into["field_add"] = FieldAdd;
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
            CheckOutPath(outPath);

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
        /// The two gates every write passes through: `out` must not be an already-open package,
        /// and it must be inside the workspace.
        ///
        /// <para>
        /// WHY THIS HAS TO BE HERE. <c>Save.Run</c> ends in <c>File.WriteAllBytes(outPath, ...)</c>,
        /// so pointing `out` at the package we are editing replaces the original material on
        /// disk. Before this, the only thing saying "never do that" was a red line in
        /// skills/tt-dev-tzs/SKILL.md -- zero code behind it. The `.tzc` pipeline has had the
        /// opposite shape all along: three gates, an atomic write, a `prev.tzc` backup and a
        /// source sha256 check. This is that discipline arriving on the `.tzs` side, at the one
        /// place where losing data is possible.
        /// </para>
        ///
        /// <para>
        /// The workspace half moves an EXISTING verdict earlier rather than inventing one: the
        /// designer already refuses a package outside the workspace at LOAD time
        /// (<c>TzpManager.InCurrentWorkspace</c>, same string-prefix rule), but writing was never
        /// checked -- so a write outside the workspace used to return success and the package
        /// then could not be opened (SKILL §4.4 records that trap). We mirror the rule; we do not
        /// call it, because it lives in the designer's assemblies and this file talks to the
        /// designer by reflection only.
        /// </para>
        ///
        /// <para>
        /// Compared separator-insensitively (both sides run through <see cref="Normalize"/>).
        /// That direction matters: being permissive here costs nothing -- the designer still
        /// refuses the package later -- while being strict would reject a call that works.
        /// </para>
        ///
        /// <para>
        /// KNOWN LIMIT, left in on purpose: gate (a) compares the strings as given (separators
        /// normalised), it does not resolve relative paths against the process CWD. A relative
        /// `out` therefore cannot be caught pointing at an open package. That case is not worth
        /// the resolution: the daemon's CWD is the engine's own directory, so a relative `out`
        /// lands next to the engine, where no source package ever lives -- and the dangerous
        /// version of this mistake (an absolute path back into the workspace) is exactly what
        /// gate (a) does catch. Gate (b) does resolve, because there the comparison is against
        /// the workspace directory rather than against another caller-supplied string.
        /// </para>
        /// </summary>
        static void CheckOutPath(string outPath) {
            string outFull = Normalize(outPath);

            // (a) `out` == ANY open package. Not just the one being saved: `save --form A --out
            //     <B's path>` would destroy B, and B never entered this call.
            lock (_gate) {
                foreach (KeyValuePair<string, DS> kv in _handles) {
                    if (kv.Value == null || string.IsNullOrEmpty(kv.Value.Path)) continue;
                    if (string.Equals(Normalize(kv.Value.Path), outFull, StringComparison.OrdinalIgnoreCase))
                        throw new DetailedError("validation",
                            "out 指向一个已打开的包（" + kv.Value.Path + "）：那会覆盖原始素材。"
                            + "另给一个 out 路径（写新包），或先 close 它。",
                            new JObject {
                                { "param",    "out" },
                                { "reason",   "out-is-open-package" },
                                { "value",    outPath },
                                { "conflict", kv.Value.Path },
                            });
                }
            }

            // (b) `out` inside the workspace. Same resolution order as Rpc.PipeName: the booted
            //     value first (a daemon is always Booted before it serves anything), the hint
            //     second (`--pipe-name` runs pre-Boot). Reading only WorkspaceHint here was the
            //     first cut and it silently skipped the check, because the daemon branch of
            //     tzs-server.cs never sets the hint -- measured, not deduced.
            string ws = Designer.Workspace;
            if (string.IsNullOrEmpty(ws)) ws = Rpc.WorkspaceHint;
            if (string.IsNullOrEmpty(ws)) return;
            string outDir;
            try {
                outDir = Path.GetDirectoryName(Path.GetFullPath(outPath));
            } catch (Exception) {
                // Unresolvable path: the write itself will fail with a clearer IO error than
                // anything invented here.
                return;
            }
            if (string.IsNullOrEmpty(outDir)) return;
            string wsPrefix = Normalize(ws) + "\\";
            string dirPrefix = Normalize(outDir) + "\\";
            if (!dirPrefix.StartsWith(wsPrefix, StringComparison.OrdinalIgnoreCase))
                throw new DetailedError("validation",
                    "out 落在工作区之外（" + outPath + "）：现在写会退 0 成功，但那个包之后打不开"
                    + "（设计器在**加载**时才查工作区）。把它写到工作区内，或改工作区设置。",
                    new JObject {
                        { "param",     "out" },
                        { "reason",    "out-outside-workspace" },
                        { "value",     outPath },
                        { "workspace", ws },
                    });
        }

        /// <summary>
        /// Separators to `\` so that two spellings of one path compare equal. Trailing
        /// separators are NOT trimmed: the caller compares directory prefixes with one added,
        /// and trimming here would make `C:\ws\` and `C:\ws` collide in a way the designer's own
        /// rule does not (internal/dev/tzs/server.go normalises the workspace on the Go side).
        /// </summary>
        static string Normalize(string p) {
            return p.Replace('/', '\\');
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

        // ------------------------------------------------------------------ field_add（任务级动词）

        /// <summary>.4fd 里"能当容器"的标签集 —— 与 Manifest.CONTAINERS 同源，去掉其中的 "None"
        /// （它不是标签）。**不复用** Struct.CONTAINER_TYPES：那一份只列 add_field 接受的*容器模式*，
        /// 比标签集窄（没有 HBox/VBox/Folder/Page），两回事。</summary>
        static readonly string[] CONTAINER_TAGS =
            { "Grid", "Group", "ScrollGrid", "Table", "Tree", "Page", "HBox", "VBox", "Folder" };

        /// <summary>
        /// `field_add` —— 任务级动词：一次请求做完「按列建字段 → 报校验增量 →（可选）存新包」。
        ///
        /// 为什么要有它（实测语料 apmt500_wf(c).tzs，加 3 列）：同一条链手工做是 **9 次调用**
        /// （list_columns / open / list_spec_nodes / form_tree / validate / add_field / validate /
        /// save / close），输出 62,873 字节，其中两次 validate 占掉 85% 的时间；而调用方真正想说的
        /// 只有一句"把这几列加到这张表单上"。这里把**动作之间的那段**搬进引擎：
        ///
        ///   · 容器自己挑（worksheet → *layout* → 根下唯一容器，见 PickContainer），也可 into 指定；
        ///   · file 已经开着就**复用**，不再撞 E_KEY_IN_USE（手工链最容易踩的一脚，实测踩过）；
        ///   · 校验基线在改动**之前**建立，所以 newErrors/newWarnings 是真增量（§7.4）。
        ///
        /// 它不重复实现任何东西：建字段走 Struct.AddFieldFn（设计器自己的 UICreator 一次构造 N 列），
        /// 校验走 Validate.Run，存包走本模块的 SaveFn。返回体刻意**瘦**：不回表单全量、不回校验的
        /// baseline/after 全表，只回答"加了什么、有没有新增问题、存哪了"。
        /// </summary>
        public static object FieldAdd(DS s0, JObject a) {
            string file = Str(a, "file");
            string handle = Str(a, "handle");
            DS s = s0;
            bool opened = false;

            if (s == null) {
                if (!string.IsNullOrEmpty(file)) {
                    s = Rpc.FindByPath(file);          // 已经开着同一个文件 → 复用，不撞 E_KEY_IN_USE
                    if (s == null) {
                        var o = new JObject();
                        o["path"] = file;
                        Open(null, o);                 // 走模块自己的 Open：登记进 _handles，也进核心 _open
                        s = Rpc.FindByPath(file);
                        opened = true;
                    }
                } else if (!string.IsNullOrEmpty(handle)) {
                    s = Rpc.FindByHandle(handle);
                    if (s == null) {
                        // 逻辑键（程序名 / ProgramKey）也认，与其它动词一致（Rpc.FindByKey）。
                        List<DS> byKey = Rpc.FindByKey(handle);
                        if (byKey.Count == 1) s = byKey[0];
                        else if (byKey.Count > 1)
                            throw TzsError.Validation(
                                "程序名 " + handle + " 对应多个开着的会话：请用 ProgramKey 形式指定");
                    }
                } else {
                    throw TzsError.Validation(
                        "field_add 需要 file（.tzs 路径，没开就顺手开）或 handle（已在开的表单）之一");
                }
                if (s == null) throw TzsError.NotFound("会话", file ?? handle);
            }

            string container = Str(a, "into");
            if (string.IsNullOrEmpty(container)) container = PickContainer(s);

            // 基线必须在**改动之前**建立：否则第一次 validate 会把"改完的样子"当成基线，
            // 增量按构造恒为空（Validate.Run 的注释与 §7.4）。已经有了就直接用，省一次校验。
            bool hadBaseline = Validate.HasBaseline(s.Handle);
            if (!hadBaseline) Validate.Run(s, new JObject());

            var addArgs = new JObject();
            addArgs["path"] = container;
            addArgs["table"] = Str(a, "table");
            addArgs["columns"] = a["columns"];
            string cmode = Str(a, "container");
            if (!string.IsNullOrEmpty(cmode)) addArgs["container"] = cmode;
            object added = Struct.AddFieldFn(s, addArgs);

            JObject v = Lean(Validate.Run(s, new JObject()));

            object saved = null;
            string outPath = Str(a, "out");
            if (!string.IsNullOrEmpty(outPath)) {
                var saveArgs = new JObject();
                saveArgs["out"] = outPath;
                saved = SaveFn(s, saveArgs);
            }

            var rep = new Dictionary<string, object> {
                { "form",      s.Program },
                { "key",       KeyString(s.Key) },
                { "handle",    s.Handle },
                { "opened",    opened },
                { "container", container },
                { "added",     added },
                { "validate",  v },
                { "baselineCached", hadBaseline },
            };
            if (saved != null) rep["saved"] = saved;
            return rep;
        }

        /// <summary>挑默认父容器。规则只认设计器模板的命名习惯，**不靠猜**：
        ///   ① 叫 worksheet 的容器 —— 设计器模板里放字段的那块（实测 6 个真实包里 5 个有）；
        ///   ② 名字含 layout 的容器（mainlayout 是外层布局框；capp002 这类没有 worksheet 的有它）；
        ///   ③ 根（&lt;Form&gt;）下**唯一**的容器 —— 没有歧义才用；
        ///   ④ 都不成立就报错要求显式 into：宁可不做，也不把字段塞进一个猜出来的盒子里。
        /// </summary>
        static string PickContainer(DS s) {
            byte[] zip = Read.Zip(s);
            string fdText = Reflect.EntryText(zip, Read.Entry(zip, ".4fd"));
            ElementIndex idx = ElementIndex.Build(fdText);

            string byLayout = null, onlyAtRoot = null;
            int rootContainers = 0;
            foreach (ElementSpan el in idx.All) {
                if (el.Parent == null) continue;                    // 跳过 .4fd 的文档根
                if (Array.IndexOf(CONTAINER_TAGS, el.Tag) < 0) continue;
                if (string.Equals(el.Name, "worksheet", StringComparison.Ordinal)) return el.Path;
                if (byLayout == null && el.Name != null
                    && el.Name.IndexOf("layout", StringComparison.OrdinalIgnoreCase) >= 0) byLayout = el.Path;
                if (el.Depth == 2) { rootContainers++; onlyAtRoot = el.Path; }  // 根 <Form> 的直接子容器
            }
            if (byLayout != null) return byLayout;
            if (rootContainers == 1) return onlyAtRoot;
            throw TzsError.Validation(
                "挑不出要加进哪个容器（没有 worksheet、也没有唯一的 layout）：用 into 显式指定父容器的 name-path；"
                + "可先 `form_tree` 看一眼结构");
        }

        /// <summary>把 validate 的返回体瘦成"增量 + 计数"。
        ///
        /// 任务动词不该把 baseline/after 两张全表塞进调用方的 context（真实表单各 ≈2 KB，
        /// 整表更大）—— 要看全量就单独调一次 validate。这里只回答"这次改动新增了什么"。
        /// </summary>
        static JObject Lean(object validateResult) {
            var d = validateResult as IDictionary<string, object>;
            var o = new JObject();
            if (d == null) { o["raw"] = JToken.FromObject(validateResult); return o; }
            foreach (string k in new[] { "newErrors", "newWarnings", "elapsedMs" })
                if (d.ContainsKey(k)) o[k] = JToken.FromObject(d[k]);
            o["newErrorCount"] = Count(d, "newErrors");
            o["newWarningCount"] = Count(d, "newWarnings");
            return o;
        }

        static int Count(IDictionary<string, object> d, string key) {
            object v;
            if (!d.TryGetValue(key, out v) || v == null) return 0;
            var list = v as System.Collections.IList;
            return list == null ? 0 : list.Count;
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
