# skills — AI 技能（也是给人看的操作手册）

五套技能，每套一个目录、目录里一份 `SKILL.md`：

| 技能 | 讲什么 |
|---|---|
| `tt-debug` | 调试：命令、人机交接的话术、接口报文日志、只读 SQL 的读法 |
| `tt-dev-tzc` | 代码包 `.tzc`：流程、红线、退出码、报错怎么读、动词速查 |
| `tt-dev-tzs` | 表单包 `.tzs`：一次调用改一处的形状、寻址、校验与"增量"、动词全表 |
| `tt-dict` | 数据字典：能查什么、每个命令的返回怎么读、数据维护、类型码 |
| `erp-read` | 读 ERP 源码：源码获取顺序、命名规范（程序/函数/变量/表格） |

## 三条约定

1. **目录名必须等于 `SKILL.md` 开头元数据里的 `name`** —— 这是技能规范的硬要求。
   所以**本目录下不再放 `README.md` 之类的文件**，那份 `SKILL.md` 就是该技能目录的说明。
2. **它们是对外文档**：改了工具面（动词、参数、退出码、错误文案）就要当作改 API 文档对待。
   评测装置里，执行者**只能读这一份** —— 所以每一份都必须自足，不能出现"详见别处"式的唯一出处。
3. **它们是发行物**：便携包与安装包都带着整个 `skills/` 目录，用户可以直接编辑。

## 安装到别处

```bash
tt install skills                       # 复制到 <当前目录>/skills
tt install skills --to .claude/skills   # 装到 Claude Code 直接读的位置
```

## 判据

```bash
go test ./internal/cli -run TestInstall      # 安装面：复制、冲突拒绝、源=目标拒绝
python tools/eval/grade.py --run <装置根>     # 工具面评测（回读产物，不接受自述）
```

## 细节去哪

- 各技能怎么用：直接读它的 `SKILL.md`
- 工具面评测装置 → [../tools/eval/README.md](../tools/eval/README.md)
- 契约与不变量（技能里不讲的那部分） → [../internal/README.md](../internal/README.md)
