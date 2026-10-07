# skills — AI 技能（也是给人看的操作手册）

六套技能，每套一个目录、目录里一份 `SKILL.md`：

| 技能 | 讲什么 |
|---|---|
| `tt-debug` | 调试：命令、人机交接的话术、接口报文日志、只读 SQL 的读法 |
| `tt-dev-tzc` | 代码包 `.tzc`：流程、红线、退出码、报错怎么读、动词速查 |
| `tt-dev-tzs` | 表单包 `.tzs`：一次调用改一处的形状、寻址、校验与"增量"、动词全表 |
| `tt-dict` | 数据字典：能查什么、每个命令的返回怎么读、数据维护、类型码 |
| `tt-drawio` | 原型图：组件清单、排版规格怎么写、`tt drawio compose` 出图 |
| `tt-erp-read` | 读 ERP 源码：源码获取顺序、命名规范（程序/函数/变量/表格） |

## 三条约定

1. **目录名必须等于 `SKILL.md` 开头元数据里的 `name`** —— 这是技能规范的硬要求。
   所以**本目录下不再放 `README.md` 之类的文件**，那份 `SKILL.md` 就是该技能目录的说明。
   `tt install skills` 装之前会连 `description`（非空、≤1024 字符）一起逐个校验，不合规当场拒绝。
2. **它们是对外文档**：改了工具面（动词、参数、退出码、错误文案）就要当作改 API 文档对待。
   评测装置里，执行者**只能读这一份** —— 所以每一份都必须自足，不能出现"详见别处"式的唯一出处。
3. **它们是发行物**：便携包与安装包都带着整个 `skills/` 目录，用户可以直接编辑。

## 安装到别处

```bash
tt install skills --to auto             # 装到 agent 会读的位置：已有 .claude/ 之类就用它，否则建 .agents/skills
tt install skills --agent claude-code   # 按名字落到某家 agent 读的位置
tt install skills --to .agents/skills   # 跨客户端公约数
tt install skills                       # 复制到 <当前目录>/skills（与 exe 旁边同形的那一份，给人读）
```

`--agent` 接受的名字：`agents` / `universal`、`claude-code` / `claude`、`copilot` / `github-copilot` /
`vscode`、`cursor`。`.agents/skills` 是跨客户端公约数（pi、Codex、Cursor、Gemini CLI、
OpenCode、Copilot 都扫它），所以 `--to auto` 在没有任何已初始化 agent 目录时新建它。

安装前逐个校验：`SKILL.md` 里 frontmatter 的 `name` 必须等于目录名、`description` 非空且
≤1024 字符。不合规的技能装出去不会报错，失败全在我们看不到的地方（description 缺失时客户端
直接跳过，name 不匹配则各家宽容度不一），所以这里当场拒绝。

## 判据

```bash
go test ./internal/cli -run 'TestInstall|TestListSkills|TestResolve' -count=1
```

覆盖安装面（复制、冲突拒绝、源=目标拒绝）、frontmatter 校验（目录名≠name、description
空/超长）、`--to auto` 探测与 `--agent` 名字表，以及仓库自带这棵树本身合规。

## 细节去哪

- 各技能怎么用：直接读它的 `SKILL.md`
- `SKILL.md` 的字段规范 → <https://agentskills.io/specification>
- 契约与不变量（技能里不讲的那部分） → [../internal/README.md](../internal/README.md)
