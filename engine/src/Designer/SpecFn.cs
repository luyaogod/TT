using System;
using System.Collections.Generic;
using Newtonsoft.Json.Linq;

namespace TzsCli.Designer {

    /// <summary>
    /// A server function. Returns anything JSON-serialisable; throwing is how it reports an
    /// error, and the transport maps TzsError to the wire codes of SPEC §11.24 (a).
    ///
    /// Frozen in Wave 0 (§11.24 (e)) so that the Fns/*.cs modules can be written in parallel
    /// without any of them reading another's code. This file is the shared declaration; nobody
    /// edits it during a wave. Manifest.cs holds the *table* of descriptors; the Fns modules
    /// hold the *bodies*. Three writers, one contract.
    /// </summary>
    public delegate object Fn(Session s, JObject args);

    /// <summary>Parameter type, used for both CLI positional parsing and JSON validation.</summary>
    public enum PType {
        Str, Int, Bool,
        Path,        // a .tzs path or a name-path inside the form
        Handle,      // a server-minted handle ("h1")
        Kind,        // one of the seven spec node kinds
        AttrName,    // describe_from decides the legal set at runtime
        Enum,        // Values decides it statically
        StrList, PathList
    }

    public sealed class Param {
        public string Name;
        public PType Type;
        public bool Required;
        public string Default;
        public string Desc;
        public string[] Values;      // for PType.Enum
        /// <summary>"spec:&lt;kind&gt;" for spec-node attributes, "layout" for element attributes.
        /// The legal set is materialised from the live model, never hardcoded -- that is what
        /// makes describe_kind authoritative and stops a caller writing an attribute the node
        /// does not have (SPEC §11.24 (d)).</summary>
        public string DescribeFrom;
    }

    /// <summary>One declared function. Frozen in Wave 0 (§11.24 (c)).</summary>
    public sealed class SpecFn {
        public string Name;      // "set_layout_attr"
        public string Group;     // "attr" -- drives CLI help grouping
        public string Summary;   // one line
        public Param[] Params;   // ordered: drives BOTH CLI positionals and JSON validation
        public bool NeedsHandle = true;
        /// <summary>True means the first call transitions the handle Loaded -> Mutable by
        /// registering an UndoRedoManager (SPEC §11.24 (b)). Irreversible for the handle.</summary>
        public bool Mutating;
        /// <summary>Marked on functions that take seconds, so a caller does not put them in a
        /// loop. validate is 1.6 s on a 114-element form and 10.4 s on a 670-element one.</summary>
        public bool Slow;
        public string[] Errors;  // the E_* codes this fn can return; drives help AND tests
        public string Returns;   // closed vocabulary: void|handle|el|tree|list<el>|delta|kindmap
    }

    /// <summary>Small builders so the manifest table reads as a table rather than as object
    /// initialisers. Kept here, with the types, so every Fns module writes them the same way.</summary>
    public static class P {
        public static Param Str(string n, bool req = true, string desc = null)
            { return new Param { Name = n, Type = PType.Str, Required = req, Desc = desc }; }
        public static Param Int(string n, bool req = true, string desc = null)
            { return new Param { Name = n, Type = PType.Int, Required = req, Desc = desc }; }
        public static Param Bool(string n, bool req = false, string desc = null)
            { return new Param { Name = n, Type = PType.Bool, Required = req, Desc = desc }; }
        public static Param Path(string n, bool req = true, string desc = null)
            { return new Param { Name = n, Type = PType.Path, Required = req, Desc = desc }; }
        public static Param Handle(string n = "handle")
            { return new Param { Name = n, Type = PType.Handle, Required = true }; }
        public static Param Kind(string n = "kind")
            { return new Param { Name = n, Type = PType.Kind, Required = true }; }
        public static Param Attr(string n, string from, bool req = true)
            { return new Param { Name = n, Type = PType.AttrName, Required = req, DescribeFrom = from }; }
        public static Param Enum(string n, string[] values, bool req = true)
            { return new Param { Name = n, Type = PType.Enum, Required = req, Values = values }; }
        public static Param List(string n, bool req = true)
            { return new Param { Name = n, Type = PType.PathList, Required = req }; }
        /// <summary>A list of plain strings -- NOT paths. `List` above is PathList, and using it
        /// for a value list publishes the wrong type to every reader of --manifest (set_items'
        /// `items` carried `path[]` for a list of "name|text|description" strings).</summary>
        public static Param StrList(string n, bool req = false, string desc = null)
            { return new Param { Name = n, Type = PType.StrList, Required = req, Desc = desc }; }
    }

    /// <summary>The seven spec-node slots a FormSpecModel actually holds.
    ///
    /// NOT eight: "SpecItem" appeared in AddField.Promote's array but exists nowhere in the
    /// designer -- Prop() returned null for it on every call, so it was a harmless dead entry.
    /// See SPEC §11.24 0.3. FormSpecModel.cs is the authority: SpecField :267, SpecAction :272,
    /// SpecHelpCode :277, SpecMultiLang :282, SpecProgRel :287, SpecReference :292, SpecTree :297.
    ///
    /// The kind strings are the XML element names the .tsd uses, which is also what the CLI and
    /// the JSON surface use.</summary>
    public static class SpecSlots {
        public static readonly string[] Kinds =
            { "field", "hfield", "pfield", "rfield", "mlfield", "tree", "act" };

        /// <summary>kind name -> the FormSpecModel property that holds it.</summary>
        public static readonly Dictionary<string, string> ByKind = new Dictionary<string, string> {
            { "field",   "SpecField"     },
            { "hfield",  "SpecHelpCode"  },
            { "pfield",  "SpecProgRel"   },
            { "rfield",  "SpecReference" },
            { "mlfield", "SpecMultiLang" },
            { "tree",    "SpecTree"      },
            { "act",     "SpecAction"    },
        };
    }
}
