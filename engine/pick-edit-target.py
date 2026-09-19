#!/usr/bin/env python
"""Pick an edit target in a .tzs for batch testing.

    pick-edit-target.py <file.tzs>      -> prints "<namePath> field:can_edit <flipped>"

Chooses the first data-bound layout element whose spec <field> carries a can_edit value,
and returns the flipped value so the change is guaranteed to differ from the current one.
That matters: SpecAttributeUndoRedoCommand drops an old==new change, so a same-value edit
would be a silent no-op and a batch full of those would prove nothing.

can_edit is the attribute of choice because it maps to a layout attribute too
(TransformCanEdit writes noEntry), so the run exercises both sides of the write.
"""
import re, sys, zipfile
from xml.etree import ElementTree as ET


def paths_of(root, prefix):
    """All elements under `root` as name-paths (name attribute, else tag), starting at
    `prefix`. Mirrors ElementIndex: the path root is the ManagedForm element, so callers
    pass the <Form> subtree with the document root's name as the prefix -- walking from the
    document root instead picks up the <Record> sections, whose name is "Undefined"."""
    out = []

    def walk(e, here):
        out.append((here, e))
        for c in e:
            walk(c, here + "/" + (c.get("name") or c.tag))

    walk(root, prefix + "/" + (root.get("name") or root.tag))
    return out


def main():
    if len(sys.argv) < 2:
        sys.exit("usage: pick-edit-target.py <file.tzs> [set|add]")
    mode = sys.argv[2] if len(sys.argv) > 2 else "set"
    z = zipfile.ZipFile(sys.argv[1])
    fd = next((n for n in z.namelist() if n.endswith(".4fd")), None)
    ts = next((n for n in z.namelist() if n.endswith(".tsd")), None)
    if not fd or not ts:
        return
    root = ET.fromstring(z.read(fd).decode("utf-8", errors="replace"))
    form = root.find("Form")
    if form is None:
        return
    tsd = z.read(ts).decode("utf-8", errors="replace")

    if mode == "add":
        # A Button's SpecNodeType is ACTION, and Grid/Group accept it. Pick the first
        # populated Grid so the button lands somewhere the mime rule allows.
        for p, e in paths_of(form, root.get("name") or root.tag):
            if e.tag == "Grid" and len(e):
                print(p, "Button")
                return
        return

    can_edit = {}
    # <field> is often paired (it carries a CDATA body) rather than self-closing, so match
    # the opening tag only.
    for m in re.finditer(r"<field\b[^>]*>", tsd):
        g = m.group(0)
        n = re.search(r' name="([^"]*)"', g)
        c = re.search(r' can_edit="([^"]*)"', g)
        if n and c:
            can_edit[n.group(1)] = c.group(1)

    for p, e in paths_of(form, root.get("name") or root.tag):
        name = e.get("name")
        if not name or not e.get("colName"):
            continue
        cur = can_edit.get(name)
        if cur not in ("Y", "N"):
            continue
        print(p, "field:can_edit", "N" if cur == "Y" else "Y")
        return


if __name__ == "__main__":
    main()
