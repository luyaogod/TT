using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using System.Xml.Linq;
using TzsCli;

namespace TzsCli.Designer
{
    /// <summary>
    /// The save tail: Edit.cs:716-741 extracted, with the entry names derived from the input
    /// bytes rather than read out of Main's locals.
    ///
    /// The order is the designer's own -- SaveToTSD() and SaveBinding() rebuild TSDElement and
    /// ElementBindings from the live model, and those two strings, not the ones in the original
    /// archive, are what has to be written back. Only the entries that actually changed are
    /// replaced, so every untouched entry keeps its original compressed bytes and timestamp
    /// (SPEC §11.7); the designer's own PackageManager.Packing cannot do that, which is the
    /// whole reason TzsRepacker exists.
    ///
    /// Returns the repacked bytes as well as writing them, so the caller can report the sizes
    /// ([6] repacked: N -> M bytes) without stat-ing the file it just wrote.
    /// </summary>
    public static class Save
    {
        public static byte[] Run(byte[] original, string outPath, FormWriter w4, Session s) {
            string fdEntry  = Reflect.EntryName(original, ".4fd");
            string tsdEntry = Reflect.EntryName(original, ".tsd");
            string bdxEntry = Reflect.EntryName(original, ".bdx");

            // A spec-only change is legitimate: can_edit maps to noEntry, and if that attribute
            // already holds the target value the layout text is untouched while the .tsd still
            // has to be rewritten. So this must not be gated on FormWriter.Dirty -- Render()
            // returns the input byte-for-byte when nothing was applied, so the .4fd entry is
            // simply rewritten with identical bytes.
            string fd4New = w4.Render();
            Reflect.Call(s.Si, "SaveToTSD");
            Reflect.Call(s.Si, "SaveBinding");
            string tsdNew = ((XElement)Reflect.Prop(s.Si, "TSDElement")).ToString(SaveOptions.DisableFormatting);
            string bdxNew = Reflect.Prop(s.Tzp, "ElementBindings") as string;

            var repl = new Dictionary<string, byte[]> {
                { fdEntry, Encoding.UTF8.GetBytes(fd4New) },
                { tsdEntry, Encoding.UTF8.GetBytes(tsdNew) }
            };
            // A null or empty ElementBindings means the package has no .bdx to write; the
            // original entry (if any) is then left untouched rather than blanked.
            if (bdxEntry != null && !string.IsNullOrEmpty(bdxNew)) repl[bdxEntry] = Encoding.UTF8.GetBytes(bdxNew);
            byte[] result = TzsRepacker.Repack(original, repl);
            File.WriteAllBytes(outPath, result);
            return result;
        }
    }
}
