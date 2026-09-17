// scripts/fgloutline.fixtures.mjs
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

// src/fgloutline.ts
var W_S = "[ \\t]*";
var W_P = "[ \\t]+";
var ID = "[A-Za-z_][A-Za-z0-9_.]*";
var NB = "(?![A-Za-z0-9_])";
var RE_END = new RegExp(
  "^" + W_S + "END" + W_P + "(FUNCTION|MAIN|REPORT|DIALOG|INPUT|DISPLAY|CONSTRUCT|MENU)" + NB,
  "i"
);
var CH_LF = 10;
var CH_HASH = 35;
var CH_SQ = 39;
var CH_DQ = 34;
var CH_BT = 96;
var CH_BS = 92;
var CH_CR = 13;
var CH_DASH = 45;
var CH_LB = 123;
var CH_RB = 125;
var TYPE_KEYWORDS = "INTEGER|INT|SMALLINT|BIGINT|TINYINT|SERIAL8|SERIAL|BIGSERIAL|CHAR|VARCHAR|NVARCHAR|STRING|TEXT|DECIMAL|NUMERIC|MONEY|FLOAT|SMALLFLOAT|DOUBLE|DATE|DATETIME|INTERVAL|BOOLEAN|BYTE|DYNAMIC|STATIC|RECORD|LIKE|ARRAY";
var SUB_EVENT = "ACTION|CHANGE|KEY|IDLE|TIMER|ROW" + W_P + "CHANGE|EVERY" + W_P + "ROW|LAST" + W_P + "ROW|FILL" + W_P + "BUFFER|SORT|APPEND|INSERT|UPDATE|DELETE|EXPAND|COLLAPSE|SELECTION" + W_P + "CHANGE|DRAG_START|DRAG_FINISHED|DRAG_ENTER|DRAG_OVER|DROP";
var SUB_CONTROL = "INPUT|CONSTRUCT|DISPLAY|DIALOG|MENU|ROW|FIELD|INSERT|DELETE|UPDATE|GROUP" + W_P + "OF";
var SUB_REPORT = "(?:FIRST" + W_P + ")?PAGE" + W_P + "(?:HEADER|TRAILER)";
var SUB_COMMAND = "(?:COMMAND)(?:" + W_P + "KEY" + W_S + "\\([^)]*\\))?" + NB;
var SUB_HEAD = "(?:BEFORE|AFTER)" + W_P + "(?:" + SUB_CONTROL + ")" + NB + "|(?:ON)" + W_P + "(?:" + SUB_EVENT + ")" + NB + "|" + SUB_REPORT + NB + "|" + SUB_COMMAND;
var RE_SUB = new RegExp("^" + W_S + "(?<head>" + SUB_HEAD + ")(?<rest>.*)", "i");
var RE_INPUT_OPTION = new RegExp("^" + W_S + "INPUT" + W_P + "(?:NO" + W_P + ")?WRAP" + W_S + "$", "i");
var RE_OPEN = new RegExp(
  "^" + W_S + "(?:(?:PUBLIC|PRIVATE)" + W_P + ")?(?:(?<fn>FUNCTION)" + W_S + "(?:[(][^)]*[)]" + W_S + ")?(?<fnName>" + ID + ")" + W_S + "[(]|(?<main>MAIN)" + NB + // REPORT 必须带 name( —— 否则 RECORD 里名叫 `report` 的字段会被当成块开头:
  // 声明 `report DYNAMIC ARRAY OF …` 的字段会造出幻影 Module 节点;
  // 又因 REPORT 属 MODULE_KINDS,还会连带截断宿主函数(实测有一处让所在函数少掉 200 余行)。
  // 语料里 2,368 处 END REPORT 对应的声明 100% 是 `REPORT name(...)`。
  "|(?<report>REPORT)" + NB + W_P + "(?<reportName>" + ID + ")" + W_S + "[(]|(?<dialog>DIALOG)(?![A-Za-z0-9_.])(?:" + W_P + "ATTRIBUTES?" + NB + "|" + W_P + "(?<dialogName>" + ID + ")" + W_S + "[(]|" + W_S + "$)|(?<input>INPUT)" + NB + "(?!" + W_P + "(?:NO" + W_P + ")?WRAP" + NB + ")" + W_P + "(?:(?<inputArray>ARRAY" + W_P + ID + ")|(?<inputByName>BY" + W_P + "NAME" + W_P + ID + ")|(?<inputRecord>" + ID + "))" + NB + "|(?<display>DISPLAY)" + NB + W_P + "ARRAY" + NB + W_P + "(?<displayArray>" + ID + ")|(?<construct>CONSTRUCT)" + NB + W_P + "(?<constructBy>BY" + W_P + "NAME" + W_P + ")?(?!(?:" + TYPE_KEYWORDS + ")" + NB + ")(?<constructName>" + ID + ")(?=" + W_S + "(?:(?:ON|FROM|ATTRIBUTE|ATTRIBUTES)" + NB + "|$))|(?<menu>MENU)(?![A-Za-z0-9_.])" + W_P + "(?<menuText>\"[^\"]*\"|'[^']*'|`[^`]*`|(?!(?:ATTRIBUTES?)" + NB + ")" + ID + ")|(?<menuAttr>MENU)(?![A-Za-z0-9_.])" + W_P + "ATTRIBUTES?" + NB + "|(?<inputBare>INPUT)" + NB + W_S + "$|(?<menuBare>MENU)" + NB + W_S + "$)",
  "i"
);
var TERMINATOR = {
  FUNCTION: "FUNCTION",
  MAIN: "MAIN",
  REPORT: "REPORT",
  DIALOG: "DIALOG",
  INPUT: "INPUT",
  DISPLAY: "DISPLAY",
  CONSTRUCT: "CONSTRUCT",
  MENU: "MENU"
};
var MODULE_KINDS = /* @__PURE__ */ new Set(["FUNCTION", "MAIN", "REPORT"]);
var RE_SAME_LINE_END = {};
for (const k of Object.keys(TERMINATOR)) {
  RE_SAME_LINE_END[k] = new RegExp("(?:^|[^A-Za-z0-9_])END" + W_P + TERMINATOR[k] + NB, "i");
}
var MAX_UNCLOSED_SPAN = 1e4;
function maskLines(text) {
  const code = [];
  const disp = [];
  let c = "";
  let d = "";
  let quote = 0;
  let inBrace = false;
  const both = (s) => {
    c += s;
    d += s;
  };
  for (let i = 0, n = text.length; i < n; i++) {
    const ch = text[i];
    const cc = text.charCodeAt(i);
    if (cc === CH_LF) {
      code.push(c);
      disp.push(d);
      c = "";
      d = "";
      continue;
    }
    if (cc === CH_CR) {
      if (text.charCodeAt(i + 1) === CH_LF) continue;
      code.push(c);
      disp.push(d);
      c = "";
      d = "";
      continue;
    }
    if (inBrace) {
      if (cc === CH_RB) inBrace = false;
      both(" ");
      continue;
    }
    if (quote) {
      if (cc === CH_BS) {
        const nx = text.charCodeAt(i + 1);
        if (quote === CH_BT || nx === quote || nx === CH_BS) {
          c += "  ";
          d += text[i] + text[i + 1];
          i++;
          continue;
        }
      }
      if (quote === CH_BT && cc === CH_BT) {
        quote = 0;
        both(ch);
        continue;
      }
      if (cc === quote) {
        if (quote !== CH_BT && text.charCodeAt(i + 1) === quote) {
          c += "  ";
          d += ch + ch;
          i++;
          continue;
        }
        quote = 0;
        both(ch);
        continue;
      }
      c += " ";
      d += ch;
      continue;
    }
    if (cc === CH_SQ || cc === CH_DQ || cc === CH_BT) {
      quote = cc;
      both(ch);
      continue;
    }
    if (cc === CH_LB) {
      inBrace = true;
      both(" ");
      continue;
    }
    if (cc === CH_HASH || cc === CH_DASH && text.charCodeAt(i + 1) === CH_DASH) {
      both(" ");
      while (i + 1 < n && text.charCodeAt(i + 1) !== CH_LF) {
        both(" ");
        i++;
      }
      continue;
    }
    both(ch);
  }
  code.push(c);
  disp.push(d);
  return { code, disp };
}
var normalize = (s) => s.replace(/[ \t]+/g, " ").trim().replace(/[.,]+$/, "");
function describe(mo) {
  const g = mo.groups;
  if (g.fn) {
    return {
      kind: "FUNCTION",
      label: g.fnName,
      // 函数名在参数表的 '(' 之前,必须从整个匹配里定位,否则会得到负偏移。
      nameCol: mo.index + Math.max(0, mo[0].lastIndexOf(g.fnName))
    };
  }
  if (g.main) return { kind: "MAIN", label: "MAIN" };
  if (g.report) return { kind: "REPORT", label: g.reportName ? "REPORT " + g.reportName : "REPORT" };
  if (g.dialog) return { kind: "DIALOG", label: g.dialogName ? "DIALOG " + g.dialogName : "DIALOG" };
  if (g.input || g.inputBare) {
    const arg = g.inputByName ?? g.inputArray ?? g.inputRecord ?? "";
    return { kind: "INPUT", label: normalize("INPUT " + arg) };
  }
  if (g.display) return { kind: "DISPLAY", label: normalize("DISPLAY ARRAY " + g.displayArray) };
  if (g.construct) {
    const by = g.constructBy ? "BY NAME " : "";
    return { kind: "CONSTRUCT", label: normalize("CONSTRUCT " + by + g.constructName) };
  }
  if (g.menu || g.menuBare || g.menuAttr) {
    return { kind: "MENU", label: normalize("MENU " + (g.menuText ?? "")) };
  }
  return null;
}
function scanBlocks(code, disp) {
  const root = {
    kind: "FUNCTION",
    label: "",
    startLine: 0,
    endLine: code.length - 1,
    selStart: 0,
    selEnd: 0,
    children: []
  };
  const stack = [{ kind: "ROOT", node: root }];
  const closeTop = (endLine) => {
    const f = stack.pop();
    f.node.endLine = Math.max(endLine, f.node.startLine);
  };
  const closeTopRepair = (endLine) => {
    const f = stack.pop();
    const isModule = MODULE_KINDS.has(f.kind);
    const bare = f.kind !== "SUB" && !isModule && f.node.children.length === 0;
    const end = bare ? f.node.startLine : isModule ? endLine : Math.min(endLine, f.node.startLine + MAX_UNCLOSED_SPAN);
    f.node.endLine = Math.max(end, f.node.startLine);
  };
  for (let i = 0, n = code.length; i < n; i++) {
    const line = code[i];
    if (RE_INPUT_OPTION.test(line)) continue;
    const me = RE_END.exec(line);
    if (me) {
      const type = me[1].toUpperCase();
      let k = stack.length - 1;
      if (MODULE_KINDS.has(type)) {
        while (k > 0 && stack[k].kind !== type) k--;
      } else {
        while (k > 0 && stack[k].kind === "SUB") k--;
      }
      if (k > 0 && stack[k].kind === type) {
        while (stack.length - 1 > k) closeTopRepair(i - 1);
        closeTop(i);
      }
      continue;
    }
    const ms = RE_SUB.exec(line);
    if (ms) {
      while (stack.length > 1 && stack[stack.length - 1].kind === "SUB") closeTop(i - 1);
      const msd = RE_SUB.exec(disp[i]) ?? ms;
      let label = normalize(msd.groups.head.toUpperCase() + msd.groups.rest);
      const cmd = /^(COMMAND(?:\s+KEY\s*\([^)]*\))?)\s+("[^"]*"|'[^']*'|`[^`]*`|[A-Za-z_]\w*(?:\[[^\]]*\])*)/i.exec(
        msd.groups.head + msd.groups.rest
      );
      if (cmd) label = normalize(cmd[1].toUpperCase() + " " + cmd[2]);
      const node2 = {
        kind: "SUB",
        label,
        startLine: i,
        endLine: i,
        selStart: ms.index,
        selEnd: ms.index + ms[0].length,
        children: []
      };
      stack[stack.length - 1].node.children.push(node2);
      stack.push({ kind: "SUB", node: node2 });
      continue;
    }
    const mo = RE_OPEN.exec(line);
    if (!mo) continue;
    const d = describe(RE_OPEN.exec(disp[i]) ?? mo);
    if (!d) continue;
    if (MODULE_KINDS.has(d.kind)) {
      while (stack.length > 1) closeTop(i - 1);
    }
    const node = {
      kind: d.kind,
      label: d.label,
      startLine: i,
      endLine: i,
      selStart: d.nameCol ?? mo.index,
      selEnd: d.nameCol !== void 0 ? d.nameCol + d.label.length : mo.index + mo[0].length,
      children: []
    };
    stack[stack.length - 1].node.children.push(node);
    if (RE_SAME_LINE_END[d.kind].test(line.slice(mo.index + mo[0].length))) node.endLine = i;
    else stack.push({ kind: d.kind, node });
  }
  while (stack.length > 1) closeTopRepair(code.length - 1);
  fixup(root);
  return root.children;
}
function fixup(root) {
  const order = [];
  const stack = [root];
  while (stack.length > 0) {
    const n = stack.pop();
    if (n.endLine < n.startLine) n.endLine = n.startLine;
    order.push(n);
    for (const c of n.children) stack.push(c);
  }
  for (let i = order.length - 1; i >= 0; i--) {
    const n = order[i];
    for (const c of n.children) if (c.endLine > n.endLine) n.endLine = c.endLine;
  }
}
function makeNode(nd, raw) {
  const endLine0 = Math.min(nd.endLine, raw.length - 1);
  const lineLen = raw[nd.startLine]?.length ?? 0;
  const selEnd = Math.min(nd.selEnd, lineLen);
  const selStart = Math.min(nd.selStart, selEnd);
  return {
    label: nd.label,
    kind: nd.kind,
    line: nd.startLine + 1,
    endLine: endLine0 + 1,
    selStart,
    selEnd
  };
}
function toOutlineNodes(roots, raw) {
  const map = /* @__PURE__ */ new Map();
  const order = [];
  const stack = [...roots].reverse();
  while (stack.length > 0) {
    const n = stack.pop();
    order.push(n);
    map.set(n, makeNode(n, raw));
    for (let i = n.children.length - 1; i >= 0; i--) stack.push(n.children[i]);
  }
  for (const n of order) {
    const kids = n.children.map((c) => map.get(c));
    if (kids.length > 0) map.get(n).children = kids;
  }
  return roots.map((n) => map.get(n));
}
function parseOutline(text) {
  const raw = text.split(/\r\n|\r|\n/);
  const { code, disp } = maskLines(text);
  return toOutlineNodes(scanBlocks(code, disp), raw);
}

// scripts/fgloutline.fixtures.mjs
var DIR = path.join(path.dirname(fileURLToPath(import.meta.url)), "fgl-fixtures");
var KIND_NAME = {
  FUNCTION: "Function",
  MAIN: "Module",
  REPORT: "Module",
  DIALOG: "Interface",
  INPUT: "Object",
  CONSTRUCT: "Object",
  DISPLAY: "Array",
  MENU: "Namespace",
  SUB: "Event"
};
var argv = process.argv.slice(2);
var opt = { strict: false, filter: null, list: false, tree: false, maxDiffs: 12 };
for (let i = 0; i < argv.length; i++) {
  const a = argv[i];
  if (a === "--strict") opt.strict = true;
  else if (a === "--filter") opt.filter = argv[++i];
  else if (a === "--list") opt.list = true;
  else if (a === "--tree") opt.tree = true;
  else if (a === "-h" || a === "--help") {
    console.log("\u7528\u6CD5: npm run check:outline -- [--strict] [--filter \u5B50\u4E32] [--list] [--tree]");
    process.exit(0);
  }
}
if (!fs.existsSync(DIR)) {
  console.error(`\u7F3A\u5C11\u7528\u4F8B\u76EE\u5F55: ${DIR}`);
  process.exit(2);
}
var cases = [];
for (const f of fs.readdirSync(DIR).sort()) {
  if (!f.endsWith(".expected.json")) continue;
  const name = f.slice(0, -".expected.json".length);
  const srcPath = path.join(DIR, name + ".4gl");
  if (!fs.existsSync(srcPath)) {
    cases.push({ name, error: `\u7F3A\u5C11\u6E90\u6587\u4EF6 ${name}.4gl` });
    continue;
  }
  let exp;
  try {
    exp = JSON.parse(fs.readFileSync(path.join(DIR, f), "utf8"));
  } catch (e) {
    cases.push({ name, error: `\u671F\u671B JSON \u89E3\u6790\u5931\u8D25: ${e.message}` });
    continue;
  }
  cases.push({ name, srcPath, exp });
}
var selected = cases.filter((c) => !opt.filter || c.name.includes(opt.filter));
if (selected.length === 0) {
  console.error(`\u6CA1\u6709\u5339\u914D\u7684\u7528\u4F8B${opt.filter ? `\uFF08--filter ${opt.filter}\uFF09` : ""}`);
  process.exit(2);
}
if (opt.list) {
  for (const c of selected) {
    console.log(`${c.exp && c.exp.deviation ? "\u26A0" : " "} ${c.name}`);
    if (c.exp) {
      console.log(`    doc:   ${c.exp.doc}`);
      console.log(`    about: ${c.exp.about ?? ""}`);
      if (c.exp.deviation) console.log(`    dev:   ${c.exp.deviation}`);
    }
  }
  process.exit(0);
}
var toPlain = (n) => ({
  name: n.label,
  kind: KIND_NAME[n.kind] ?? String(n.kind),
  startLine: n.line - 1,
  endLine: n.endLine - 1,
  selection: [n.selStart, n.selEnd],
  children: (n.children ?? []).map(toPlain)
});
var renderTree = (nodes, depth = 0) => {
  const out = [];
  for (const n of nodes) {
    out.push(`${"  ".repeat(depth)}${n.name}  [${n.kind} L${n.startLine + 1}-${n.endLine + 1}]`);
    out.push(...renderTree(n.children ?? [], depth + 1));
  }
  return out;
};
var diffs = [];
function compare(exp, act, at) {
  const n = Math.max(exp.length, act.length);
  for (let i = 0; i < n; i++) {
    const e = exp[i];
    const a = act[i];
    const where = `${at}[${i}]`;
    if (!e) {
      diffs.push(`${where} \u591A\u51FA\u8282\u70B9 ${a.name} [${a.kind} L${a.startLine + 1}]`);
      continue;
    }
    if (!a) {
      diffs.push(`${where} \u7F3A\u5C11\u8282\u70B9 ${e.name} [${e.kind} L${e.startLine + 1}]`);
      continue;
    }
    const label = `${where} ${e.name}@L${e.startLine + 1}`;
    if (e.name !== a.name) diffs.push(`${label}: name \u671F\u671B ${JSON.stringify(e.name)}\uFF0C\u5B9E\u9645 ${JSON.stringify(a.name)}`);
    if (e.kind !== a.kind) diffs.push(`${label}: kind \u671F\u671B ${e.kind}\uFF0C\u5B9E\u9645 ${a.kind}`);
    if (e.startLine !== a.startLine) diffs.push(`${label}: startLine \u671F\u671B ${e.startLine}\uFF0C\u5B9E\u9645 ${a.startLine}`);
    if (e.endLine !== a.endLine) diffs.push(`${label}: endLine \u671F\u671B ${e.endLine}\uFF0C\u5B9E\u9645 ${a.endLine}`);
    if (e.selection && (e.selection[0] !== a.selection[0] || e.selection[1] !== a.selection[1])) {
      diffs.push(`${label}: selectionRange \u671F\u671B [${e.selection.join(",")}]\uFF0C\u5B9E\u9645 [${a.selection.join(",")}]`);
    }
    compare(e.children ?? [], a.children ?? [], `${at}[${i}].children`);
  }
}
var fatal = 0;
var known = 0;
var pass = 0;
var crashed = 0;
var report = [];
for (const c of selected) {
  if (c.error) {
    fatal++;
    report.push(`
\u274C ${c.name}
    - ${c.error}`);
    continue;
  }
  const text = fs.readFileSync(c.srcPath, "utf8");
  let nodes;
  try {
    nodes = parseOutline(text);
  } catch (e) {
    crashed++;
    fatal++;
    report.push(`
\u274C ${c.name}
    - \u89E3\u6790\u5668\u629B\u5F02\u5E38: ${e.message}`);
    continue;
  }
  const actual = (nodes ?? []).map(toPlain);
  diffs.length = 0;
  compare(c.exp.nodes ?? [], actual, "root");
  const d = diffs.slice(0, opt.maxDiffs);
  const more = diffs.length - d.length;
  if (d.length === 0) {
    pass++;
    if (opt.tree) {
      report.push(`
\u2705 ${c.name}
${renderTree(actual).map((l) => "    " + l).join("\n")}`);
    }
    continue;
  }
  const isKnown = !!c.exp.deviation && !opt.strict;
  if (isKnown) known++;
  else fatal++;
  report.push(
    `
${isKnown ? "\u26A0\uFE0F " : "\u274C"} ${c.name}  (${c.exp.doc})
    \u8003\u4EC0\u4E48: ${c.exp.about ?? ""}
` + (isKnown ? `    \u5DF2\u77E5\u504F\u5DEE: ${c.exp.deviation}
` : "") + d.map((x) => "    - " + x).join("\n") + (more > 0 ? `
    \u2026 \u53E6\u6709 ${more} \u5904\u5DEE\u5F02` : "") + `
    \u5B9E\u9645\u6811:
${renderTree(actual).map((l) => "      " + l).join("\n")}`
  );
}
console.log(`\u7528\u4F8B ${selected.length} \u4E2A\uFF1A\u901A\u8FC7 ${pass}\uFF0C\u5DF2\u77E5\u504F\u5DEE ${known}\uFF0C\u4E0D\u7B26 ${fatal}${crashed ? `\uFF0C\u89E3\u6790\u5668\u629B\u5F02\u5E38 ${crashed}` : ""}`);
console.log(report.join("\n"));
if (fatal) {
  console.error(
    `
\u5931\u8D25: ${fatal} \u4E2A\u7528\u4F8B\u4E0E\u6587\u6863\u63A8\u5BFC\u7684\u671F\u671B\u6811\u4E0D\u7B26\u3002
\u5224\u636E\u662F\u6587\u6863\uFF0C\u4E0D\u662F\u5B9E\u73B0 \u2014\u2014 \u82E5\u786E\u8BA4\u662F\u5B9E\u73B0\u9519\uFF0C\u6539 src/fgloutline.ts\uFF1B\u82E5\u786E\u8BA4\u671F\u671B\u5199\u9519\uFF0C\u6539 fgl-fixtures/*.expected.json\u3002
\uFF08${known} \u4E2A\u5DF2\u77E5\u504F\u5DEE\u5DF2\u7531\u671F\u671B\u6587\u4EF6\u91CC\u7684 "deviation" \u5B57\u6BB5\u58F0\u660E\uFF0C\u7528 --strict \u53EF\u628A\u5B83\u4EEC\u4E5F\u53D8\u6210\u5931\u8D25\u3002\uFF09`
  );
  process.exit(1);
}
console.log(`
\u2705 \u5168\u90E8\u901A\u8FC7${known ? `\uFF08\u53E6\u6709 ${known} \u4E2A\u5DF2\u5728\u671F\u671B\u6587\u4EF6\u91CC\u58F0\u660E\u7684\u5DF2\u77E5\u504F\u5DEE\uFF09` : ""}\u3002`);
