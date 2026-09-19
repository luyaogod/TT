using System;
using System.Collections;
using System.Collections.Generic;
using System.Reflection;
using System.Runtime.ExceptionServices;
using System.Text.RegularExpressions;
using System.Xml.Linq;
using Newtonsoft.Json.Linq;

namespace TzsCli.Designer
{
    // Deliberately NOT TzsCli.Designer.Fns like the other three modules in this
    // directory. Fns/Session.cs declares a static module class also called Session, and
    // a type in the same namespace outranks an enclosing one -- so inside .Fns the name
    // `Session` would bind to that module and every `Session s` parameter here would be
    // CS0721 (static types cannot be used as parameters). W2-B hit this and chose the
    // enclosing namespace; the alias route fails because a using-alias does not outrank
    // a same-namespace type. Rpc.BuildMap dispatches by class name, so the split costs
    // nothing. Do not "tidy" it without renaming the module class first.

    /// <summary>
    /// The attribute-write functions. All four are Mutating: the first call on a handle
    /// registers its UndoRedoManager (SPEC §11.24 (b)), which is a one-time step of the session
    /// and cannot be undone for the life of the handle.
    ///
    /// The two rules of SPEC §11.24 (d) are the reason this file looks the way it does:
    ///
    ///   1. A layout attribute is written ONLY through the XmlElement indexer
    ///      (`el[attr] = value`), never by constructing FormAttributesUndoRedoCommand. The
    ///      indexer is what enforces whitelist-by-presence, same-value short-circuit, the
    ///      repeat/stepX/rowCount gating and the MinGrid* clamping, and it is what finally
    ///      invokes that command (XmlElement.cs:1888-1987). Calling the command directly skips
    ///      all of it -- P0 proved it writes attributes the element does not have.
    ///   2. A spec attribute is checked against the node's OWN materialised attribute set
    ///      before it is written. Edit.cs's `set` had no such check, which is how `case="upper"`
    ///      ended up in a .tsd where `case` means nothing.
    ///
    /// Both rules exist because the failure mode is identical and silent: a write that "works"
    /// and puts a meaningless attribute in the file, which the designer then either ignores or
    /// round-trips forever.
    ///
    /// Each function writes both sides of the change: the designer's own command for the model,
    /// and the per-handle FormWriter (Fns/Session.cs, the one `save` renders) for the .4fd text --
    /// only the attributes that actually moved, exactly as Edit.cs does. The `layoutDelta` in the
    /// result is that same change reported back, for the caller's benefit; it is already applied.
    /// </summary>
    public static class Attr
    {
        public static void Register(IDictionary<string, Fn> into) {
            into["set_spec_attr"]    = SetSpecAttr;
            into["set_layout_attr"]  = SetLayoutAttr;
            into["rename_component"] = RenameComponent;
            into["set_tree_source"]  = SetTreeSource;
        }

        // ================================================================ errors

        /// <summary>
        /// The errors carry a machine-readable `detail` beside the message, because SPEC §11.24
        /// (a) makes it load-bearing ("未知属性，detail 里给可用值"): an AI that cannot see the
        /// legal set cannot self-correct, which is the whole point of the kind/detail split.
        ///
        /// The type is the namespace-level DetailedError (Rpc.cs) rather than one of our own, so
        /// that the transport can classify these by type and put the payload on the wire; a
        /// second private subclass would be caught only by the plain TzsError handler and lose
        /// the detail. The message also spells the list out in prose, so even a transport that
        /// maps nothing but Code and Message still shows the caller what was legal.
        ///
        /// The Code is the E_* name the task and §11.24 (h) use (E_FATAL_LOAD_TIMEOUT is a code
        /// there, not a kind); Rpc.Map takes an E_* code at face value.
        /// </summary>
        static DetailedError NotWhitelisted(string what, string attr, IEnumerable<string> legal, string path, string kind) {
            var list = new List<string>();
            foreach (string s in legal) list.Add(s);
            var d = new JObject();
            if (path != null) d["path"] = path;
            if (kind != null) d["kind"] = kind;
            d["attr"] = attr;
            d["legal"] = Read.StrArray(list);
            string msg = what + " 没有属性 \"" + attr + "\"；它有的是: " + string.Join(", ", list.ToArray());
            return new DetailedError("E_ATTR_NOT_WHITELIST", msg, d);
        }

        /// <summary>
        /// The success frame for "you asked for a state the form is already in".
        ///
        /// SPEC §11.24 (a) makes E_NO_OP an OUTCOME, not a failure: the caller asked for a value
        /// the model already holds, the designer's indexer short-circuits on old==new, and nothing
        /// changed. Reporting it as an error (it used to arrive as ok:false / kind "internal") told
        /// the caller "something went wrong, do not retry" -- the exact opposite of the truth, and
        /// the one answer that makes an AI abandon a request that had in fact already succeeded.
        ///
        /// The three outcomes a caller must be able to tell apart are applied / applied-but-clamped
        /// / no-op. Each carries exactly one positive marker -- `applied` / `clamped` / `noop` --
        /// beside the §11.24 (a) code in `code`, so the distinction survives even a caller that
        /// only forwards the payload. `written` is what the model actually holds; for a no-op that
        /// is the value asked for, which is precisely why nothing was done.
        /// </summary>
        static JObject Noop(string what, string attr, string value, string path) {
            var res = new JObject();
            if (path != null) res["path"] = path;
            res["attr"] = attr;
            res["old"] = value;
            res["value"] = value;
            res["written"] = value;
            res["changed"] = false;
            res["noop"] = true;
            res["code"] = "E_NO_OP";
            res["note"] = what + "." + attr + " 已经是 \"" + value + "\"，设计器同值短路，未修改任何值";
            return res;
        }

        static DetailedError Refused(string message, JObject detail) {
            if (detail == null) detail = new JObject();
            return new DetailedError("E_DESIGNER", message, detail);
        }

        /// <summary>The designer's own refusals arrive as ordinary exceptions whose message is a
        /// resource string ("名称重复" etc.), so the classification is by message. Anything that
        /// is not one of those keeps travelling outward -- reporting an unexpected NRE as
        /// "designer" would tell the caller not to retry a bug.</summary>
        static bool IsDesignerRefusal(Exception ex) {
            for (Exception e = ex; e != null; e = e.InnerException) {
                string m = e.Message ?? "";
                if (m.IndexOf("重复", StringComparison.Ordinal) >= 0
                    || m.IndexOf("AlreadyExist", StringComparison.OrdinalIgnoreCase) >= 0
                    || m.IndexOf("field is required", StringComparison.OrdinalIgnoreCase) >= 0
                    || m.IndexOf("Message_", StringComparison.Ordinal) >= 0
                    || m.IndexOf("Incorrent", StringComparison.OrdinalIgnoreCase) >= 0) return true;
            }
            return false;
        }

        // ================================================================ session plumbing

        internal static object Urm(Session s) {
            s.RegisterUndoRedo();
            return ((IDictionary)Reflect.Prop(Designer.SettingManager, "undoRedoManagerMap"))[s.Key];
        }

        /// <summary>Runs a designer command. AddThenExecute is what Edit.cs uses (its `set`), and
        /// it matters: the command's own Execute() resolves its manager through
        /// SettingManager.GetUndoRedoManager(key), which throws "No UndoRedoManager" unless the
        /// session has registered one -- so the order here is not cosmetic.</summary>
        internal static void Run(Session s, object cmd) {
            object urm = Urm(s);
            if (urm != null) { Reflect.Call(urm, "AddThenExecute", cmd); return; }
            Reflect.Call(cmd, "Execute");
        }

        internal static object NewCommand(string typeName, params object[] ctorArgs) {
            Type t = Reflect.Find(Designer.A, typeName);
            if (t == null) throw new Exception("找不到命令类 " + typeName);
            return Activator.CreateInstance(t, ctorArgs);
        }

        /// <summary>Attributes of a form element before/after a command, so only what really
        /// moved reaches the layout text. Writing an unchanged attribute would still splice the
        /// same bytes, but only by luck of escaping -- and the point of the delta is that it
        /// cannot be wrong.</summary>
        internal static JObject Diff(Dictionary<string, string> before, Dictionary<string, string> after) {
            var d = new JObject();
            foreach (var kv in after) {
                string was;
                if (before.TryGetValue(kv.Key, out was) && was == kv.Value) continue;
                d[kv.Key] = kv.Value;
            }
            return d;
        }

        internal static JObject Delta(string path, string rename, JObject attrs) {
            var d = new JObject();
            d["path"] = path;
            if (rename != null) d["rename"] = rename;
            if (attrs != null && attrs.Count > 0) d["attrs"] = attrs;
            return d;
        }

        /// <summary>Look up a form element, or the caller's error. Used by all four.</summary>
        internal static object El(Session s, string path) {
            object el = s.FindByPath(path);
            if (el == null) throw TzsError.NotFound("路径", path);
            return el;
        }

        /// <summary>
        /// The per-handle .4fd writer. It lives in Fns/Session.cs (which `save` renders) and its
        /// contract is explicit: mutating functions must edit through THIS instance, because
        /// FormWriter keeps its edits inside itself -- a second writer is not a second view of
        /// the same document, it is a second document, and `save` would never see it.
        ///
        /// So the .4fd side of every write is applied here, immediately after the model side,
        /// which is also what Edit.cs does (it snapshots the element before/after the command and
        /// writes only what moved). A spec-only change legitimately writes nothing to the .4fd:
        /// can_query, for instance, maps to no layout attribute at all, and Render() then returns
        /// the original text byte-for-byte.
        /// </summary>
        internal static FormWriter Layout(Session s) { return Fns.Session.Layout(s); }

        /// <summary>Applies a {attr: value} diff to the .4fd at one path.
        ///
        /// `name` goes last on purpose: it is the attribute the name-path is built from, so
        /// writing it re-indexes the writer and moves the path out from under anything still
        /// addressed by the old one. That combination is reachable -- set_spec_attr with
        /// attr="name" renames through SpecificationInfo.Rename, which also pushes the element's
        /// name, and a Transform* can have moved another attribute in the same call.</summary>
        internal static void Patch(FormWriter w, string path, JObject attrs) {
            if (w == null || attrs == null) return;
            JToken rename = attrs["name"];
            foreach (JProperty p in attrs.Properties()) {
                if (p.Name == "name") continue;
                w.SetAttribute(path, p.Name, (string)p.Value);
            }
            if (rename != null) w.SetAttribute(path, "name", (string)rename);
        }

        internal static object Fsm(Session s, object el) {
            string name = Read.Str(Session.Raw(el, "name"));
            object fsm = Reflect.Call(s.Si, "FindNodeByName", name);
            if (fsm == null)
                throw Refused("元素 \"" + name + "\" 在 FormSpeDictionary 里没有规格模型"
                    + "（可能是纯布局元素，例如容器/标签；规格属性只存在于有规格节点的元素上）", null);
            return fsm;
        }

        // ================================================================ set_spec_attr

        /// <summary>
        /// Writes one attribute of one spec node, through the designer's own
        /// SpecAttributeUndoRedoCommand -- which also pushes the value onto the paired form
        /// element (TransformRequired / TransformCanEdit / TransformTableColumn / ...) and sets
        /// Status |= MODIFY. Re-deriving that mapping here would be guessing at tables the
        /// designer already owns (SPEC §11.16).
        ///
        /// What this adds over Edit.cs's `set` is the whitelist check of rule (d) 2: `attr` is
        /// first matched against the node's own materialised attribute set, and a name that is
        /// not in it is refused with the legal list rather than written. That is exactly the
        /// `case="upper"` bug: `case` is a layout attribute, the node never had it, and the
        /// command happily wrote it into the .tsd.
        /// </summary>
        public static object SetSpecAttr(Session s, JObject a) {
            string path = Read.Need(a, "path");
            string kind = Read.Need(a, "kind");
            string attr = Read.Need(a, "attr");
            string value = Read.NeedKey(a, "value");

            if (Array.IndexOf(SpecSlots.Kinds, kind) < 0)
                throw TzsError.Validation("未知节点类型 kind=" + kind + "；合法值: " + string.Join(",", SpecSlots.Kinds));

            object el = El(s, path);
            object fsm = Fsm(s, el);
            object node = Reflect.Prop(fsm, SpecSlots.ByKind[kind]);
            if (node == null)
                throw Refused("元素 \"" + Read.Str(Session.Raw(el, "name")) + "\" 没有 " + kind + " 节点；"
                    + "先用 get_component 看它有哪些 kind，或用 describe_kind 看该 kind 的合法属性", null);

            XElement src = Reflect.Prop(node, "Source") as XElement;
            List<string> legal = Read.AttrNames(src);
            if (legal.IndexOf(attr) < 0)
                throw NotWhitelisted(kind + " 节点", attr, legal, path, kind);

            string old = src.Attribute(attr).Value;
            if (old == value) {
                // A success, not an error (§11.24 (a)): the spec value is already what was asked
                // for, so the designer's AddAttributeChanged would return on old==new and only
                // flip Status to 'u'. Same shape as the real result, plus the no-op marker.
                JObject noop = Noop(kind, attr, value, path);
                noop["kind"] = kind;
                string nst = Json.StatusLetter(node);
                if (nst != null) noop["specStatus"] = nst;
                noop["layoutDelta"] = Delta(path, null, null);
                return noop;
            }

            var before = Session.Attrs(el);
            object cmd = NewCommand("SpecAttributeUndoRedoCommand", new object[] { node });
            Reflect.Call(cmd, "AddAttributeChanged", attr, old, value);
            Run(s, cmd);
            var after = Session.Attrs(el);

            // The command pushes the spec value onto the paired form element (TransformRequired /
            // TransformCanEdit / TransformTableColumn / ...), so the .4fd has to follow -- but only
            // for the attributes that actually moved. A spec attribute with no layout counterpart
            // (can_query, cite_std, i_zoom, ...) leaves this empty, which is correct, not a miss.
            JObject moved = Diff(before, after);
            Patch(Layout(s), path, moved);

            var res = new JObject();
            res["path"] = path;
            res["kind"] = kind;
            res["attr"] = attr;
            res["old"] = old;
            res["value"] = value;
            string st = Json.StatusLetter(node);
            if (st != null) res["specStatus"] = st;
            res["layoutDelta"] = Delta(path, null, moved);
            return res;
        }

        // ================================================================ set_layout_attr

        /// <summary>
        /// Writes one layout attribute, or one attribute across several elements. Two forms:
        /// `path` (one element) and `paths` (a batch of them, same attribute, same value).
        ///
        /// The single form goes through the XmlElement indexer, which is the only path that
        /// applies the rules (SPEC §11.24 (d) 1). Three outcomes are distinguishable afterwards
        /// by re-reading the attribute, and all three are reported as SUCCESSES (§11.24 (a)),
        /// each with exactly one positive marker so the three cannot be confused:
        ///
        ///   value written as asked        -> applied:true, `changed`:true
        ///   value written but different   -> clamped:true, code E_ATTR_CLAMPED (a MinGrid* floor
        ///                                    moved it), and `written` carries what the model holds
        ///   value already equals `old`    -> noop:true,   code E_NO_OP (`changed`:false); the
        ///                                    indexer short-circuits and nothing moved
        ///
        /// A fourth, genuinely non-success case is still an error: the indexer silently ignores
        /// stepX, stepY, rowCount and columnCount outside a ScrollGrid, so `written` comes back
        /// equal to `old` for a value the caller did NOT ask for. That is neither a clamp nor a
        /// no-op -- it is the designer refusing -- and reporting it as success would be the
        /// silent-corruption failure mode this project exists to remove. It maps to E_DESIGNER.
        /// </summary>
        public static object SetLayoutAttr(Session s, JObject a) {
            string attr = Read.Need(a, "attr");
            string value = Read.NeedKey(a, "value");

            string one = null;
            string[] paths = Read.ListArg(a, "paths");
            JToken pj = a == null ? null : a["path"];
            if (pj != null && pj.Type != JTokenType.Null) {
                if (pj.Type == JTokenType.Array) { if (paths == null) paths = Read.ListArg(a, "path"); }
                else one = Read.Arg(a, "path");
            }
            // `paths` wins when both are present (the manifest documents it as taking over);
            // rejecting the pair outright would be stricter than the declared contract.
            if (paths != null && paths.Length > 0) return Batch(s, paths, attr, value);
            if (one != null) return One(s, one, attr, value);
            throw TzsError.Validation("set_layout_attr 需要 path（单个元素）或 paths（批量）");
        }

        static PropertyInfo IndexerProp(object el) {
            foreach (var p in el.GetType().GetProperties(BindingFlags.Public | BindingFlags.Instance)) {
                if (p.Name != "Item") continue;
                ParameterInfo[] ix = p.GetIndexParameters();
                if (ix.Length == 1 && ix[0].ParameterType == typeof(string) && p.CanWrite) return p;
            }
            return null;
        }

        /// <summary>Writes through the indexer. Unwrapping TargetInvocationException matters:
        /// the indexer's `repeat` gate throws the designer's own InvalidOperationException, and
        /// that is a fact about the form, not a reflection failure.</summary>
        static void IndexerSet(object el, string attr, string value) {
            PropertyInfo pi = IndexerProp(el);
            if (pi == null) throw new Exception("XmlElement 上没有单字符串索引器（设计器版本变了？）");
            try {
                pi.SetValue(el, value, new object[] { attr });
            } catch (TargetInvocationException ex) {
                ExceptionDispatchInfo.Capture(ex.InnerException ?? ex).Throw();
            }
        }

        static object One(Session s, string path, string attr, string value) {
            Urm(s);                                  // the indexer ends in a command that needs one
            object el = El(s, path);

            string old = Read.Str(Session.Raw(el, attr));
            if (old == null) throw NotWhitelisted("元素", attr, Session.Attrs(el).Keys, path, null);
            if (old == value) {
                // §11.24 (a): E_NO_OP is a success. The caller asked for the state the element is
                // already in; `Patch` is deliberately skipped because nothing moved.
                JObject noop = Noop("布局属性", attr, value, path);
                noop["layoutDelta"] = Delta(path, null, null);
                return noop;
            }

            IndexerSet(el, attr, value);
            string written = Read.Str(Session.Raw(el, attr));

            var res = new JObject();
            res["path"] = path;
            res["attr"] = attr;
            res["old"] = old;
            if (written == value) {
                res["value"] = value;
                res["written"] = written;
                res["applied"] = true;
                res["changed"] = true;
            } else if (written == old && !Clampable(el, attr)) {
                // The indexer's gating branch returned without writing at all. Only stepX/stepY/
                // rowCount/columnCount outside a ScrollGrid can do that, and reporting it as
                // success would be the silent-corruption failure mode this project exists to
                // remove.
                throw Refused("属性 \"" + attr + "\" 在这个元素上被索引器静默忽略（仍然是 \"" + old
                    + "\"）：stepX/stepY/rowCount/columnCount 只在 ScrollGrid 的子元素上有效", res);
            } else {
                // Clamped. `written` is what the model holds -- which can equal `old` when the
                // floor happens to be the value that is already there (writing gridWidth=1 on an
                // element whose MinGridWidth is already its width). That is still a clamp, and it
                // still has to say so rather than pretending the request went through.
                res["value"] = value;
                res["written"] = written;
                res["clamped"] = true;
                res["code"] = "E_ATTR_CLAMPED";
                res["changed"] = written != old;
                res["note"] = "被 MinGrid*/ScrollGrid 下限夹到 " + written;
            }
            // The delta carries what the model holds, not what was asked for -- otherwise a save
            // would write the un-clamped value straight back into the .4fd.
            Patch(Layout(s), path, Read.AttrsToJ(new Dictionary<string, string> { { attr, written } }));
            res["layoutDelta"] = Delta(path, null, Read.AttrsToJ(new Dictionary<string, string> { { attr, written } }));
            return res;
        }

        /// <summary>
        /// Whether "nothing changed" can mean a clamp rather than a refusal. The indexer clamps
        /// gridWidth/gridHeight/posX/posY against the MinGrid* of the element, and
        /// rowCount/columnCount against the ScrollGrid it belongs to; for the other gated
        /// attributes (stepX/stepY, and rowCount/columnCount outside a ScrollGrid) the indexer
        /// returns without writing anything, and that is a designer refusal, not a clamp.
        /// </summary>
        static bool Clampable(object el, string attr) {
            if (attr == "gridWidth" || attr == "gridHeight" || attr == "posX" || attr == "posY") return true;
            if (attr == "rowCount" || attr == "columnCount") return ParentIsScrollGrid(el);
            return false;
        }

        static bool ParentIsScrollGrid(object el) {
            object p = Reflect.Prop(el, "Parent");
            if (p == null) return false;
            string t = Read.Str(Reflect.Prop(p, "Type"));
            return t == "ScrollGrid";
        }

        /// <summary>
        /// Batch form: several elements, one attribute, one value.
        ///
        /// Goes through MultiFormAttributesUndoRedoCommand (SPEC §11.22's 批量改属性), but the
        /// checks the single form gets from the indexer are done here first, because that
        /// command's Execute() writes attributes with a bare SetAttribute and so does not
        /// enforce them itself:
        ///
        ///   * every path must resolve, and every element must already HAVE the attribute --
        ///     otherwise E_ATTR_NOT_WHITELIST (this is rule (d) 1's whitelist-by-presence,
        ///     applied by hand for the one command that cannot apply it);
        ///   * if every element already holds the value, that is a E_NO_OP success (§11.24 (a)),
        ///     not an error: the whole batch is already in the asked-for state;
        ///   * gridWidth/gridHeight ARE clamped even on this path, because the command assigns
        ///     them through the XmlElement.GridWidth/GridHeight properties rather than
        ///     SetAttribute, and those clamp against MinGridWidth/MinGridHeight -- so the
        ///     read-back below can still report E_ATTR_CLAMPED.
        ///
        /// Every element of `results` carries exactly one marker -- applied / clamped / noop (or
        /// `ignored` for the silently-dropped stepX/rowCount the command cannot honour) -- so the
        /// caller reads the same three-way distinction as the single form, one element at a time.
        /// </summary>
        static object Batch(Session s, string[] paths, string attr, string value) {
            Urm(s);

            var els = new List<object>();
            var olds = new List<string>();
            bool anyDifferent = false;
            for (int i = 0; i < paths.Length; i++) {
                object el = El(s, paths[i]);
                string old = Read.Str(Session.Raw(el, attr));
                if (old == null) throw NotWhitelisted("元素", attr, Session.Attrs(el).Keys, paths[i], null);
                if (old != value) anyDifferent = true;
                els.Add(el);
                olds.Add(old);
            }
            if (!anyDifferent) {
                // §11.24 (a) again: all elements already hold the value. Reported as a success
                // with the same shape as a real batch, so a caller reads one vocabulary.
                var noopArr = new JArray();
                for (int i = 0; i < paths.Length; i++) {
                    var o = new JObject();
                    o["path"] = paths[i];
                    o["old"] = olds[i];
                    o["written"] = olds[i];
                    o["changed"] = false;
                    o["noop"] = true;
                    noopArr.Add(o);
                }
                var noopRes = new JObject();
                noopRes["attr"] = attr;
                noopRes["value"] = value;
                noopRes["count"] = paths.Length;
                noopRes["results"] = noopArr;
                noopRes["changed"] = false;
                noopRes["noop"] = true;
                noopRes["code"] = "E_NO_OP";
                noopRes["note"] = "所有元素都已经是 \"" + value + "\"，未修改任何值";
                return noopRes;
            }

            // Activator cannot bind List<object> to the ctor's IEnumerable<XmlElement>, so the
            // list is built at the element type the model actually uses.
            Type listType = typeof(List<>).MakeGenericType(els[0].GetType());
            IList typed = (IList)Activator.CreateInstance(listType);
            foreach (object e in els) typed.Add(e);
            object cmd = NewCommand("MultiFormAttributesUndoRedoCommand", new object[] { typed, attr, value });
            Run(s, cmd);

            var arr = new JArray();
            int clamped = 0, unchanged = 0, applied = 0;
            FormWriter w4 = Layout(s);
            for (int i = 0; i < els.Count; i++) {
                string written = Read.Str(Session.Raw(els[i], attr));
                var o = new JObject();
                o["path"] = paths[i];
                o["old"] = olds[i];
                o["written"] = written;
                o["changed"] = written != olds[i];
                if (written != value) {
                    if (written == olds[i]) {
                        // The value did not move. Two very different reasons, and the single form
                        // (One) already tells them apart, so batch must too:
                        //   * a MinGrid*/ScrollGrid floor happens to sit exactly on `old`, so this
                        //     IS a clamp (E_ATTR_CLAMPED) -- the caller's request was understood and
                        //     answered, just not by moving the number;
                        //   * the command silently drops stepX/stepY/rowCount/columnCount outside a
                        //     ScrollGrid, so nothing happened and nothing will: `ignored`, a refusal.
                        if (Clampable(els[i], attr)) { o["clamped"] = true; o["clampToOld"] = true; clamped++; }
                        else { o["ignored"] = true; unchanged++; }
                    } else { o["clamped"] = true; clamped++; }
                } else if (written != olds[i]) {
                    o["applied"] = true; applied++;
                } else {
                    o["noop"] = true;
                }
                // Only what actually moved reaches the .4fd text, so an element the command left
                // alone is not rewritten with a value it already had.
                if (written != olds[i]) w4.SetAttribute(paths[i], attr, written);
                arr.Add(o);
            }

            var res = new JObject();
            res["attr"] = attr;
            res["value"] = value;
            res["count"] = els.Count;
            res["results"] = arr;
            res["applied"] = applied;
            res["clamped"] = clamped;
            if (clamped > 0 && unchanged == 0) res["code"] = "E_ATTR_CLAMPED";
            if (unchanged > 0) {
                res["ignored"] = unchanged;
                res["note"] = unchanged + " 个元素的值没变（该属性在这些元素上被静默忽略）";
            }
            return res;
        }

        // ================================================================ rename_component

        /// <summary>
        /// Renames a component through RenameUndoRedoCommand -- i.e. through
        /// SpecificationInfo.Rename, which is where the designer enforces uniqueness
        /// (FormSpeDictionary-scoped, SPEC §11.20) and where it renames the spec nodes and the
        /// layout element together.
        ///
        /// The designer signals a collision by throwing Message_NameAlreadyExist ("名称重复");
        /// that is a fact about the form, so it maps to E_DESIGNER (do not retry). It is
        /// pre-checked here as well, because a throw from inside a command is a poor error
        /// message to hand an AI when the check is one dictionary lookup.
        ///
        /// NOTE the layout consequence, which is why the delta uses `rename` and not `attrs`:
        /// the element's `name` is what its name-path is built from, so the write is not an
        /// attribute value at a still-valid path -- the path itself moves. A session applying
        /// this at save time must rename, not address the old path afterwards.
        /// </summary>
        public static object RenameComponent(Session s, JObject a) {
            string path = Read.Need(a, "path");
            string name = Read.Need(a, "name");

            Urm(s);
            object el = El(s, path);
            string old = Read.Str(Session.Raw(el, "name"));
            if (old == name) {
                // §11.24 (a): a rename onto the name it already has is a no-op success. The path
                // does not move (no rename reached the .4fd), so `layoutDelta` carries no `rename`.
                var noop = new JObject();
                noop["path"] = path;
                noop["old"] = old;
                noop["name"] = name;
                noop["changed"] = false;
                noop["noop"] = true;
                noop["code"] = "E_NO_OP";
                noop["note"] = "控件代号已经是 \"" + name + "\"，未做任何修改";
                noop["layoutDelta"] = Delta(path, null, null);
                return noop;
            }

            if (Regex.IsMatch(name, "[^a-zA-Z0-9_\\.]"))
                throw Refused("控件代号只能用字母、数字、下划线和点：\"" + name + "\"", null);
            if ((bool)Reflect.Call(s.Si, "IsExists", name))
                throw Refused("名称重复：\"" + name + "\" 已经是这张表单上的另一个控件（Rename 查的是整张表单的 FormSpeDictionary）", null);

            object cmd = NewCommand("RenameUndoRedoCommand", new object[] { s.Key, old, name });
            try {
                Run(s, cmd);
            } catch (Exception ex) {
                if (IsDesignerRefusal(ex)) throw Refused(Deepest(ex).Message, null);
                throw;
            }

            // The .4fd side is a value change on the `name` attribute, not a new path: the writer
            // indexes the text, so patching the value re-indexes it and every later path agrees
            // with the model again. Done last, because until the model rename lands this path is
            // still the one the text carries.
            Layout(s).SetAttribute(path, "name", name);

            var res = new JObject();
            res["path"] = path;
            res["old"] = old;
            res["name"] = name;
            res["layoutDelta"] = Delta(path, name, null);
            return res;
        }

        static Exception Deepest(Exception ex) {
            Exception e = ex;
            while (e.InnerException != null) e = e.InnerException;
            return e;
        }

        // ================================================================ set_tree_source

        /// <summary>
        /// One cell of a Tree's data source: the &lt;tree&gt; spec node holds thirteen child
        /// elements (type1..type6, id, pid, desc, speed, stype, sid, spid) and each carries
        /// table/col/src. This writes one attribute of one of them, through
        /// SpecTreeAttributeUndoRedoCommand -- which is also where the notification name comes
        /// from: the designer itself calls it as ("type{0}", "Type{0}"), ("id","Id"),
        /// ("pid","Pid") ... i.e. the element name with its first letter capitalised
        /// (SpecTreeNode.cs:197/232/268/304/327), so that convention is reproduced rather than
        /// invented.
        ///
        /// `element` and `attr` are both validated against the node's own materialised set, for
        /// the same reason as the spec whitelist: `table`/`col`/`src`/`no` are the attributes
        /// that exist and anything else is a meaningless write.
        /// </summary>
        public static object SetTreeSource(Session s, JObject a) {
            string path = Read.Need(a, "path");
            string element = Read.Need(a, "element");
            // The manifest names this parameter `property` (after the command's own
            // SpecTreeAttributeUndoRedoCommand(node, elementName, propertyName), where the third
            // argument is the notification name); `attr` is accepted as the alias the rest of
            // this file's vocabulary uses. The value is the attribute on the tree child either
            // way -- table/col/src/no.
            string attr = Read.Arg(a, "property");
            if (string.IsNullOrEmpty(attr)) attr = Read.Arg(a, "attr");
            if (string.IsNullOrEmpty(attr)) throw TzsError.Validation("缺少参数 property（tree 子元素的属性名，如 table/col/src）");
            string value = Read.NeedKey(a, "value");

            Urm(s);
            object el = El(s, path);
            object fsm = Fsm(s, el);
            object node = Reflect.Prop(fsm, SpecSlots.ByKind["tree"]);
            if (node == null)
                throw Refused("元素 \"" + Read.Str(Session.Raw(el, "name")) + "\" 没有 tree 节点"
                    + "（数据来源只存在于 Tree 组件上）", null);

            XElement src = Reflect.Prop(node, "Source") as XElement;
            XElement child = src == null ? null : src.Element(element);
            if (child == null) {
                var names = new List<string>();
                if (src != null) foreach (XElement c in src.Elements()) names.Add(c.Name.LocalName);
                throw TzsError.NotFound("tree 子元素", element + "；可用: " + string.Join(", ", names.ToArray()));
            }
            List<string> legal = Read.AttrNames(child);
            if (legal.IndexOf(attr) < 0)
                throw NotWhitelisted("tree/" + element, attr, legal, path, "tree");

            string old = child.Attribute(attr).Value;
            if (old == value) {
                // §11.24 (a): the tree cell already holds this value, so the write is a no-op
                // success rather than an error. Same shape as the real result, plus the marker.
                JObject noop = Noop("tree/" + element, attr, value, path);
                noop["element"] = element;
                string nst = Json.StatusLetter(node);
                if (nst != null) noop["specStatus"] = nst;
                return noop;
            }

            string prop = element.Substring(0, 1).ToUpperInvariant()
                + (element.Length > 1 ? element.Substring(1) : "");
            object cmd = NewCommand("SpecTreeAttributeUndoRedoCommand", new object[] { node, element, prop });
            Reflect.Call(cmd, "AddAttributeChanged", attr, value);
            Run(s, cmd);

            var res = new JObject();
            res["path"] = path;
            res["element"] = element;
            res["attr"] = attr;
            res["old"] = old;
            res["value"] = value;
            string st = Json.StatusLetter(node);
            if (st != null) res["specStatus"] = st;
            // No layoutDelta: the tree's data source lives in the .tsd only. The command does
            // select the paired component (ComponentHelper.AddSelection), which is UI state and
            // not part of any file.
            return res;
        }
    }
}
