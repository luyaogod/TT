// scripts/fgltokens.test.mjs
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { compile } from "monaco-editor/esm/vs/editor/standalone/common/monarch/monarchCompile.js";
import { MonarchTokenizer } from "monaco-editor/esm/vs/editor/standalone/common/monarch/monarchLexer.js";

// src/fglTokens.ts
var FGL_KEYWORD_RE = /(?<![A-Za-z0-9_])(?i:ABSOLUTE|ACCEPT|ACTION|AFTER|ALL|ALTER|AND|ANY|APPEND|ARRAY|AS|ASC|ATTRIBUTE|ATTRIBUTES|AUTHORIZATION|AVG|BEFORE|BEGIN|BETWEEN|BIGINT|BIGSERIAL|BOOLEAN|BOTTOM|BREAKPOINT|BUFFER|BY|BYTE|CALL|CANCEL|CASE|CATCH|CENTER|CHANGE|CHAR|CIRCUIT|CLEAR|CLIPPED|CLOSE|COLLAPSE|COMMAND|COMMIT|CONNECT|CONSTANT|CONSTRAINT|CONSTRUCT|CONTINUE|COUNT|CREATE|CROSS|CURRENT|CURSOR|DATABASE|DATE|DATETIME|DAY|DECIMAL|DECLARE|DEFAULT|DEFAULTS|DEFER|DEFINE|DELETE|DESC|DIALOG|DISCONNECT|DISPLAY|DISTINCT|DO|DOUBLE|DROP|DYNAMIC|ELSE|ELSIF|END|ERROR|ESCAPE|EVERY|EXCLUSIVE|EXECUTE|EXISTS|EXIT|EXPAND|EXPLAIN|EXTERNAL|FALSE|FETCH|FGL|FIELD|FILE|FINISH|FIRST|FLOAT|FLUSH|FOR|FOREACH|FOREIGN|FORM|FORMAT|FORMONLY|FOUND|FRACTION|FREE|FROM|FULL|FUNCTION|GLOBALS|GOTO|GRANT|GROUP|HAVING|HEADER|HELP|HIDE|HOLD|HOUR|IF|IMMEDIATE|IMPORT|IN|INDEX|INFIELD|INITIALIZE|INNER|INOUT|INPUT|INSERT|INT|INTEGER|INTERFACE|INTERRUPT|INTERSECT|INTERVAL|INTO|IS|ISOLATION|JAVA|JOIN|KEY|LABEL|LAST|LEFT|LENGTH|LET|LEVEL|LIKE|LINE|LINENO|LOAD|LOCK|MAIN|MARGIN|MATCHES|MAX|MENU|MESSAGE|MIN|MINUS|MINUTE|MOD|MODE|MODULE|MONEY|MONLY|MONTH|NAME|NATURAL|NEED|NEXT|NO|NOT|NOTFOUND|NOWAIT|NULL|NUMERIC|OF|ON|OPEN|OPTION|OPTIONS|OR|ORDER|OTHERWISE|OUT|OUTER|PACKAGE|PAGE|PAGENO|PAUSE|PERCENT|PIPE|PREPARE|PRIMARY|PRINT|PRINTER|PRINTX|PRIOR|PRIVATE|PROMPT|PUBLIC|PUT|QUIT|RAISE|REAL|RECORD|RECOVER|REFERENCES|RELATIVE|RELEASE|RENAME|REPORT|RETURN|RETURNING|REVOKE|RIGHT|ROLLBACK|ROW|SAVEPOINT|SCHEMA|SCREEN|SCROLL|SECOND|SELECT|SELECTION|SEQUENCE|SERIAL|SERIAL8|SESSION|SET|SHARE|SHORT|SHOW|SKIP|SLEEP|SMALLFLOAT|SMALLINT|SPACES|SQL|SQLERROR|START|STATIC|STATISTICS|STEP|STOP|STRING|STYLE|STYLES|SUBDIALOG|SUM|SYNONYM|TABLE|TEMP|TEMPORARY|TERMINATE|TEXT|THEN|THROUGH|THRU|TIMER|TINYINT|TO|TOP|TRAILER|TRIGGER|TRUE|TRY|TYPE|UNBUFFERED|UNION|UNIQUE|UNITS|UNLOAD|UPDATE|USING|VALIDATE|VALUES|VARCHAR|VIEW|WARNING|WEEKDAY|WHEN|WHENEVER|WHERE|WHILE|WINDOW|WITH|WITHOUT|WORDWRAP|WORK|WRAP|YEAR)(?![A-Za-z0-9_])/;
var FGL_MONARCH = {
  // 对齐 BDL 关键字规则的 (?i)(大小写不敏感);注释/字符串/中和规则无字母,不受影响
  ignoreCase: true,
  tokenizer: {
    root: [
      // ---- 注释(必须排在字符串之前:实测注释里含引号远多于字符串里含 #)----
      // --#{ 必须排在 --# 与 -- 之前
      [/--#\{/, { token: "comment", next: "@cmt" }],
      [/--#.*$/, "comment"],
      // `--` 不能行首锚定:真实源码有行尾 -- 注释
      [/--.*$/, "comment"],
      [/#.*$/, "comment"],
      // 花括号注释按语言规定不能嵌套:遇第一个 } 即结束(故内容两份一起抹掉)
      [/\{/, { token: "comment", next: "@brace" }],
      // ---- 字符串(在注释之后;%" 本地化串必须排在 " 之前)----
      [/%"/, { token: "string", next: "@loc" }],
      [/'/, { token: "string", next: "@sq" }],
      [/"/, { token: "string", next: "@dq" }],
      [/`/, { token: "string", next: "@bt" }],
      // ---- 中和规则(无名 = 保持默认色;必须排在关键字之前)----
      // 点号后的成员/方法访问:g_browser.clear()、g_qryparam.where
      [/\.[ \t]*[A-Za-z_][A-Za-z0-9_]*/, ""],
      // ON ACTION 后的动作名是用户自定义的(ON ACTION next/close/delete),只给 ON/ACTION 上色。
      // 捕获组必须无缝铺满整条匹配,故把空白也各列一组(见文件头第 3 点)。
      [
        /(ON)([ \t]+)(ACTION)([ \t]+)([A-Za-z_][A-Za-z0-9_]*)/,
        ["keyword", "", "keyword", "", ""]
      ],
      // 预处理器指令不是关键字:&ifdef / &else / &endif / &include
      [/&[A-Za-z_][A-Za-z0-9_]*/, ""],
      // ---- 全部关键字共用一个 keyword token(类型词也在表内 → 统一蓝)----
      [FGL_KEYWORD_RE, "keyword"],
      // ---- 中性词符:整段吃掉标识符/数字词,防词尾镶嵌的关键字被上一条匹配 ----
      [/[0-9A-Za-z_]+/, ""],
      // 其余一律不上色(数字、运算符、括号、中文……)
      [/./, ""]
    ],
    // --#{ … --#}
    cmt: [
      [/--#\}/, { token: "comment", next: "@pop" }],
      [/./, "comment"]
    ],
    // { … }(不可嵌套:遇第一个 } 结束)
    brace: [
      [/\}/, { token: "comment", next: "@pop" }],
      [/./, "comment"]
    ],
    // %"…" 本地化字符串:内部规则必须显式重申 string(Monarch 无 scope 栈)
    loc: [
      [/\\../, "string"],
      [/""/, "string"],
      [/"/, { token: "string", next: "@pop" }],
      [/./, "string"]
    ],
    // '…':内部 '' 是字面引号,故闭合引号判定为「后面不再跟一个 '」
    sq: [
      [/\\../, "string"],
      [/''/, "string"],
      [/'/, { token: "string", next: "@pop" }],
      [/./, "string"]
    ],
    // "…":同上,内部 "" 是字面引号
    dq: [
      [/\\../, "string"],
      [/""/, "string"],
      [/"/, { token: "string", next: "@pop" }],
      [/./, "string"]
    ],
    // 反引号是原始字符串:不处理转义
    bt: [
      [/`/, { token: "string", next: "@pop" }],
      [/./, "string"]
    ]
  }
};

// scripts/fgltokens.test.mjs
var HERE = path.dirname(fileURLToPath(import.meta.url));
var DIR = path.join(HERE, "fgl-fixtures");
function makeTokenizer() {
  const languageService = {
    languageIdCodec: { encodeLanguageId: (v) => v, decodeLanguageId: (v) => v },
    getLanguageId: () => "4gl",
    getLanguages: () => [],
    onDidChange: () => ({ dispose() {
    } })
  };
  const themeService = {
    getColorTheme: () => ({ tokenColors: [] }),
    onColorThemeChange: () => ({ dispose() {
    } })
  };
  const configService = {
    getValue: () => 2e4,
    // editor.maxTokenizationLineLength
    onDidChangeConfiguration: () => ({ dispose() {
    } })
  };
  return new MonarchTokenizer(languageService, themeService, "4gl", compile("4gl", FGL_MONARCH), configService);
}
var ALLOWED = /* @__PURE__ */ new Set(["", "comment", "string", "keyword"]);
var kindOf = (t2) => String(t2.type).replace(/\.4gl$/, "");
function spans(tokens, line) {
  return tokens.map((t2, i) => ({
    kind: kindOf(t2),
    start: t2.offset,
    end: i + 1 < tokens.length ? tokens[i + 1].offset : line.length
  }));
}
var kwTexts = (sp, line) => sp.filter((s) => s.kind === "keyword").map((s) => line.slice(s.start, s.end));
function coverCheck(sp, kind, start, end) {
  const inside = sp.filter((s) => s.start >= start && s.end <= end && s.kind === kind);
  const covered = inside.reduce((n, s) => n + (s.end - s.start), 0);
  if (covered !== end - start) return `[${start},${end}) \u672A\u88AB ${kind} \u5B8C\u6574\u8986\u76D6(\u5B9E\u9645 ${covered} \u5B57\u7B26)`;
  const outside = sp.filter((s) => s.kind === kind && (s.start < start || s.end > end));
  if (outside.length) return `\u533A\u95F4\u5916\u8FD8\u6709 ${kind}: ${JSON.stringify(outside)}`;
  return null;
}
var BATTERY = [
  { line: "g_append = 1", kw: [], note: "\u8BCD\u5C3E\u9576\u5D4C\u7684\u5173\u952E\u5B57\u4E0D\u67D3\u8272(Monarch \u65E0 lookbehind,\u9760\u4E2D\u6027\u8BCD\u7B26\u89C4\u5219)" },
  { line: "xEND", kw: [], note: "\u540C\u4E0A:\u524D\u7F00\u7C98\u8FDE" },
  { line: "2END", kw: [], note: "\u540C\u4E0A:\u6570\u5B57\u524D\u7F00\u7C98\u8FDE" },
  { line: "_END", kw: [], note: "\u540C\u4E0A:\u4E0B\u5212\u7EBF\u7C98\u8FDE" },
  { line: "END2", kw: [], note: "lookahead \u6709\u6548:\u540E\u7F00\u7C98\u8FDE\u4E0D\u67D3\u8272" },
  { line: "END_", kw: [], note: "\u540C\u4E0A" },
  { line: "END\u5B9A\u4E49", kw: ["END"], note: "\u7D27\u90BB\u4E2D\u6587\u4ECD\u67D3\u8272(ASCII \u524D\u540E\u77BB\u800C\u975E \\b \u7684\u7528\u610F)" },
  { line: "DEFINE r RECORD", kw: ["DEFINE", "RECORD"], note: "\u666E\u901A\u5173\u952E\u5B57" },
  { line: "DISPLAY ARRAY sa TO s.*", kw: ["DISPLAY", "ARRAY", "TO"], note: "\u6570\u7EC4\u5F62\u6001" },
  { line: "END REPORT", kw: ["END", "REPORT"], note: "\u8BED\u53E5\u7EA7\u5173\u952E\u5B57" },
  { line: "g_browser.clear()", kw: [], note: "\u70B9\u53F7\u540E\u6210\u5458/\u65B9\u6CD5\u540D\u4E0D\u67D3\u8272" },
  { line: "g_qryparam.where = 1", kw: [], note: "\u540C\u4E0A(where \u662F\u5173\u952E\u5B57,\u4F46\u524D\u9762\u6709\u70B9\u53F7)" },
  { line: "ON ACTION next", kw: ["ON", "ACTION"], note: "ON ACTION \u52A8\u4F5C\u540D\u4E0D\u67D3\u8272(\u5373\u4FBF\u52A8\u4F5C\u540D\u672C\u8EAB\u662F\u5173\u952E\u5B57)" },
  { line: "on action Delete", kw: ["on", "action"], note: "\u540C\u4E0A;\u5927\u5C0F\u5199\u4E0D\u654F\u611F" },
  { line: "ON ACTION controlp INFIELD pmdldocno", kw: ["ON", "ACTION", "INFIELD"], note: "\u5C3E\u90E8 INFIELD \u7167\u5E38\u67D3\u8272" },
  { line: "ON t1.col_a = t2.col_a", kw: ["ON"], note: "SQL JOIN \u6761\u4EF6\u4E0D\u662F dialog \u5B50\u5757" },
  { line: "&ifdef DEBUG", kw: [], note: "\u9884\u5904\u7406\u5668\u6307\u4EE4\u4E0D\u67D3\u8272" },
  { line: '&include "x.4gl"', kw: [], note: "\u540C\u4E0A" },
  { line: "LET s = 'it''s ok'", kw: ["LET"], note: "\u6210\u5BF9\u5F15\u53F7\u662F\u5B57\u9762\u5F15\u53F7:\u6574\u5BF9\u662F\u4E00\u4E2A\u5B57\u7B26\u4E32", str: [8, 18] },
  { line: "cl_replace_str(x, '\\'', '')", kw: [], note: "\u8F6C\u4E49\u5F15\u53F7 '\\'' \u4E0D\u80FD\u8BA9\u5F15\u53F7\u72B6\u6001\u8DD1\u98DE" },
  { line: '%"localized"', kw: [], note: '%" \u672C\u5730\u5316\u5B57\u7B26\u4E32', str: [0, 12] },
  { line: "`raw \\n text`", kw: [], note: "\u53CD\u5F15\u53F7\u539F\u59CB\u5B57\u7B26\u4E32(\u4E0D\u5904\u7406\u8F6C\u4E49)", str: [0, 13] },
  { line: "-- it's a comment", kw: [], note: "-- \u6CE8\u91CA\u91CC\u7684\u6487\u53F7\u4E0D\u5F00\u5B57\u7B26\u4E32", cmt: [0, 17] },
  { line: "{ ON ACTION z }", kw: [], note: "\u82B1\u62EC\u53F7\u5757\u6CE8\u91CA\u5185\u5173\u952E\u5B57\u4E0D\u67D3\u8272", cmt: [0, 15] },
  { line: "--#{ ON ACTION z --#}", kw: [], note: "--#{ --#} \u5757\u6CE8\u91CA", cmt: [0, 21] },
  { line: "LET a = 1 -- \u884C\u5C3E\u6CE8\u91CA", kw: ["LET"], note: "\u884C\u5C3E -- \u6CE8\u91CA(\u8BED\u6CD5\u6587\u4EF6\u4E0D\u951A\u884C\u9996)", cmt: [10, 17] },
  { line: "LET b = 2 # \u4E95\u53F7\u6CE8\u91CA", kw: ["LET"], note: "# \u884C\u6CE8\u91CA", cmt: [10, 16] }
];
var t = makeTokenizer();
var state0 = await t.getInitialState();
var fail = 0;
var bad = (msg) => {
  console.error("  \u274C " + msg);
  fail++;
};
console.log("\u2460 \u5BF9\u6297\u7528\u4F8B");
for (const c of BATTERY) {
  const r = await t.tokenize(c.line, false, state0);
  const sp = spans(r.tokens, c.line);
  const kinds = [...new Set(sp.map((s) => s.kind))];
  const stray = kinds.filter((k) => !ALLOWED.has(k));
  let problem = null;
  if (stray.length) problem = `\u51FA\u73B0\u4E09\u7C7B\u4E4B\u5916\u7684 token: ${stray.join(",")}`;
  else {
    const got = kwTexts(sp, c.line);
    if (got.join("|") !== c.kw.join("|")) problem = `\u5173\u952E\u5B57\u671F\u671B [${c.kw.join(",")}]\uFF0C\u5B9E\u9645 [${got.join(",")}]`;
    else if (c.str) problem = coverCheck(sp, "string", c.str[0], c.str[1]);
    else if (c.cmt) problem = coverCheck(sp, "comment", c.cmt[0], c.cmt[1]);
  }
  if (problem) bad(`${JSON.stringify(c.line)} \u2014 ${problem}  (${c.note})`);
}
if (!fail) console.log(`  \u2705 ${BATTERY.length} \u4E2A\u7528\u4F8B\u5168\u8FC7(\u5173\u952E\u5B57\u8FB9\u754C / \u4E2D\u548C\u89C4\u5219 / \u4E09\u7C7B\u6CE8\u91CA / \u8F6C\u4E49)`);
var files = fs.readdirSync(DIR).filter((f) => f.endsWith(".4gl")).sort();
var strayFiles = 0;
var unbalanced = [];
for (const f of files) {
  const text = fs.readFileSync(path.join(DIR, f), "utf8");
  const lines = text.split(/\r\n|\r|\n/);
  let st = state0;
  for (let i = 0; i < lines.length; i++) {
    const r = await t.tokenize(lines[i], false, st);
    st = r.endState;
    for (const s of spans(r.tokens, lines[i])) {
      if (!ALLOWED.has(s.kind)) {
        if (strayFiles++ < 5) bad(`${f}:${i + 1} \u51FA\u73B0\u7B2C\u56DB\u7C7B token ${JSON.stringify(s.kind)}`);
      }
    }
  }
  const sent = await t.tokenize("XXSENTINELXX", false, st);
  const sentKinds = spans(sent.tokens, "XXSENTINELXX").map((s) => s.kind).filter((k) => k !== "");
  if (sentKinds.length) unbalanced.push(`${f}(${sentKinds.join(",")})`);
}
console.log(`
\u2461 \u5939\u5177\u8BED\u6599 ${files.length} \u4E2A\u6587\u4EF6\u9010\u884C\u8DD1`);
if (strayFiles) console.log(`  \u274C ${strayFiles} \u5904\u51FA\u73B0\u7B2C\u56DB\u7C7B token`);
else console.log("  \u2705 \u53EA\u51FA\u73B0 comment / string / keyword / \u9ED8\u8BA4\u8272 \u56DB\u7C7B(\u524D\u4E09 + \u65E0\u8272)");
console.log("\n\u2462 \u72B6\u6001\u5E73\u8861(\u6587\u4EF6\u672B\u5C3E\u54E8\u5175\u4E0D\u5F97\u88AB\u67D3\u8272)");
var UNBALANCED_BASELINE = [];
var cur = unbalanced.join(" ");
var base = UNBALANCED_BASELINE.join(" ");
if (cur !== base) {
  bad(`\u672A\u95ED\u5408\u6587\u4EF6\u96C6\u5408\u4E0E\u57FA\u7EBF\u4E0D\u7B26:
      \u671F\u671B [${base}]
      \u5B9E\u9645 [${cur}]`);
} else {
  console.log(`  \u2705 \u4E0E\u57FA\u7EBF\u4E00\u81F4(${UNBALANCED_BASELINE.length} \u4E2A\u672A\u95ED\u5408\u7247\u6BB5)`);
}
console.log(fail === 0 ? "\n\u2705 \u5168\u90E8\u901A\u8FC7" : `
${fail} \u5904\u5931\u8D25`);
process.exit(fail === 0 ? 0 : 1);
