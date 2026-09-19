#!/usr/bin/env python
"""Pick a (table, column) to add to a .tzs as a new field -- for batch write-path tests.

    pick-column.py <file.tzs> <workspace>      -> prints "table column", or nothing

Choosing a *new* column matters: reusing one that is already a field in the form would
exercise the naming-collision path instead of the ordinary one, and the two should not be
conflated in a test whose point is the ordinary one.

Columns without a <col_attr> element are skipped on purpose -- UICreator reads the widget
type and width from there, and a column that has none is a different (also real, but less
representative) case.
"""
import os, re, sys, zipfile


def table_of(tsd_text, fd_text):
    """Tables named by the layout's <Record> section, in first-seen order, plus the columns
    already used as fields. The first table is often a placeholder like type_t, so the
    caller walks the whole list rather than trusting the head of it."""
    tables, used = [], set()
    for m in re.finditer(r'<RecordField\b[^>]*?/>', fd_text):
        tag = m.group(0)
        t = re.search(r' sqlTabName="([^"]*)"', tag)
        c = re.search(r' colName="([^"]*)"', tag)
        if c and c.group(1):
            used.add(c.group(1))
        if t and t.group(1) and t.group(1) not in tables:
            tables.append(t.group(1))
    return tables, used


def columns_of(path):
    """Columns of a .tbl that can actually become a widget, in file order.

    The widget metadata lives in a single <col_attr> block, one <field> child per column --
    that is what TableColumnHelper.GetColField returns, and it is where UICreator reads the
    widget type and width from.

    A field with an empty widget="" is skipped: type_t (the "data type reference" table) is
    made entirely of those, and UICreator NREs on them inside ComponentFactory -- the
    designer cannot drag such a column either.
    """
    text = open(path, encoding="utf-8", errors="replace").read()
    block = re.search(r"<col_attr>(.*?)</col_attr>", text, re.S)
    if not block:
        return []
    out = []
    for m in re.finditer(r'<field\s+name="([^"]*)"\s+widget="([^"]*)"', block.group(1)):
        if m.group(2):
            out.append(m.group(1))
    return out


def main():
    if len(sys.argv) < 3:
        sys.exit("usage: pick-column.py <file.tzs> <workspace>")
    tzs, ws = sys.argv[1], sys.argv[2]
    z = zipfile.ZipFile(tzs)
    fd = next((n for n in z.namelist() if n.endswith(".4fd")), None)
    ts = next((n for n in z.namelist() if n.endswith(".tsd")), None)
    if not fd or not ts:
        return
    fd_text = z.read(fd).decode("utf-8", errors="replace")
    tables, used = table_of(z.read(ts).decode("utf-8", errors="replace"), fd_text)
    if not tables:
        return

    # Index every <module>/tbl/<table>.tbl in the workspace once.
    by_table = {}
    for root, _dirs, files in os.walk(ws):
        if os.path.basename(root) != "tbl":
            continue
        for name in files:
            if name.endswith(".tbl"):
                by_table.setdefault(name[:-4], os.path.join(root, name))

    for table in tables:
        p = by_table.get(table)
        if not p:
            continue
        for col in columns_of(p):
            if col not in used:
                print(table, col)
                return


if __name__ == "__main__":
    main()
