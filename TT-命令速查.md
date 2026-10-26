# tt 命令速查（0.5.0）

> 全部命令按「你实际要干的事」分组。示例里的环境用本机的 `正式区`，企业用 `99` / `710`，照套即可。

---

## 0. 先记住这几条通用约定

| 约定 | 说明 |
|---|---|
| **环境必须点名** | `--at <环境>` 或 `--at <环境>@<企业>`，也可以 `--env <环境>`。**没有默认环境**。不知道填什么就加 `--at-list`（纯本地，不联网） |
| **输出默认 JSON** | 想给人看加 `--human`（= `--format table`）；要 Excel 加 `--csv` |
| **企业用 @ 带** | `正式区@99`。字典/查库类命令大多需要企业（决定用哪个账号连库）；不带就用账号清单首项 |
| **远端只读** | 服务器上只有 `dev tzc` 三道闸门和 `dev tzs` 引擎动词能写；`file` 没有 put/rm |
| **退出码按命令线定义** | interface：0 成功 / 1 表单不合法 / 2 用法或歧义 / 4 不存在；outline：0/2/4；tzc：0/2/3/4/5 |
| **别名** | `tt dict r.t` 也可写 `rt` / `table`；`r.q`=`rq`/`win`；`r.v`=`rv`/`check` |

全局 flag：`--at` `--at-list` `--env` `--config` `--format` `--json` `--csv` `--human` `-v/--verbose`

---

## 1. 环境 / 配置 / 安装 / 自更新

| 目的 | 命令 |
|---|---|
| 看配了哪些环境 | `tt env list` |
| 看环境详情（含账号） | `tt env show 正式区` |
| 列每个环境里有哪些企业 | `tt env ents` / `tt env ents 正式区 --ent 710` / 现查同步 `--refresh` |
| 切默认环境 | `tt env use 正式区` |
| 配环境的企业 | `tt env topent 正式区 710` |
| 配置文件在哪 | `tt config path` |
| 校验配置 | `tt config validate` |
| 看/改配置项 | `tt config show`、`tt config get query.limit`、`tt config set query.limit 50`（`--raw` 才显明文口令） |
| 清缓存 | `tt cache clear --dry-run` / `tt cache clear --older-than 24h` |
| 装技能 / 加 PATH | `tt install skills --to auto`、`tt install skills --force`、`tt install path`（`tt install path --dry-run` 只看不写） |
| 查新版 / 更新 | `tt update check`、`tt update --yes`、`tt update log --tail 50` |
| 起本地 Web 服务 | `tt serve`（后台常驻）、`tt serve --foreground`、`tt serve --stop` |
| 版本 | `tt version` |

---

## 2. 查字典（dict）——用得最多的一组

公共参数：`--at 正式区@99`（或 `--env`）、`--limit N`、`--lang zh_CN`、`-d <sqlite路径>`

### 表 / 字段
```powershell
tt dict r.t --at 正式区                      # 列出全部表（上百张，配合 --kw）
tt dict r.t --kw 采购 --at 正式区            # 按表名/中文表说明搜表
tt dict r.t pmdl_t --at 正式区               # 单表：表说明/模块/类型/字段数
tt dict r.t pmdl_t --fields --at 正式区      # 加字段骨架（字段名/说明/类型/必填）
tt dict r.t pmdl_t --full --at 正式区        # 字段+键值+索引（最全）
tt dict r.t "pmda_t,pmdb_t" --brief --at 正式区   # 多表只看表级，避免刷屏
tt dict r.t pmdn_t --col pmdn001,pmdn007 --at 正式区   # 只看指定字段
tt dict r.t pmdl_t --kw 备注 --at 正式区     # 在表里按字段名/说明筛字段
tt dict r.t pmdl_t --who --at 正式区         # 反查哪些程序在用这张表（S/I/U/D）
```
### 其他字典
```powershell
tt dict desc pmdl_t pmdl004 --at 正式区      # 字段对应的前端规格
tt dict scc 11 --at 正式区@99                # 系统分类码（下拉框值）
tt dict scc --kw 交货状态 --at 正式区@99     # 按名称找分类码
tt dict scc 2035,2058 --brief --at 正式区@99
tt dict sysp E-CIR-0759 --at 正式区@99       # 企业级/据点级系统参数
tt dict docp D-BWF-0004 --at 正式区@99       # 单据别参数
tt dict msg std-00003 --at 正式区@99         # 系统消息/报错内容
tt dict msg --text 驳回 --type 1 --at 正式区@99   # 按文字/类型(0警告1错误2资讯)搜
tt dict prog axmt500_wf --at 正式区          # 程序/作业字典（作业名、类别、用到的表）
tt dict prog --kw 采购单 --at 正式区         # 按中文找作业
tt dict prog --sub --kw tzc --at 正式区      # 搜子程序/元件
tt dict r.q q_apca001 --at 正式区            # 开窗（CALL q_xxx）的 SQL/参数/回传列
tt dict r.v v_ooba002_07 --at 正式区         # 校验带值模板的 SQL/参数
tt dict acc show 3211 --at 正式区            # 应用分类码 ACC 元模型
tt dict acc query 3211 --template --at 正式区 > 3211.json   # 出模板→填→再 query
tt dict bdldoc dir                           # BDL 官方文档目录
tt dict db list                              # 字典库里的账号（--show-secrets 显口令）
tt dict db status --at 正式区                # 本地字典镜像状态
tt dict db sync --at 正式区                  # 同步字典到本地 SQLite（离线查更快）
tt dict spill list / tt dict spill show latest   # 结果被截断时看落盘的完整版
```

---

## 3. 连库查数（db）——只读

```powershell
tt db ping --at 正式区                       # 连通性（账号清单首项）
tt db probe --at 正式区@99                   # 更细的探测
tt db sql --at 正式区@710 "select count(*) from pmdl_t"
tt db sql --at 正式区@710 --file .\q.sql     # 从文件读语句
tt db sql --at 正式区@710 "select ..." --timeout 60
```
- 只允许单条 `SELECT`/`WITH`，只读事务，**最多 200 行**（超出看 `spill`）。
- 必须带 `@企业`：企业决定用哪个账号连库。
- 中文可能显示成 `?`（sqlplus 出口编码）：要中文先用 `rawtohex(...)` 取回再本地解码，或改用字典命令。

---

## 4. 看服务器文件 / 拉镜像（file）

```powershell
tt file ls --at 正式区                      # 默认 $TOP
tt file ls -la erp/apm/4gl --at 正式区
tt file find erp -name '*axmt500*' --at 正式区
tt file find erp -name '*.4gl' --type f --maxdepth 3 --at 正式区
tt file grep -rn --include '*.4gl' 'xmda071' erp/axm/4gl --at 正式区
tt file head -n 80 erp/apm/4gl/apmt500.4gl --at 正式区
tt file tail -n 50 <文件> ; tt file cat <文件> ; tt file stat <文件> ; tt file wc -l <文件>
tt file get erp/apm/4gl/apmt500.4gl -o .\apmt500.4gl --at 正式区   # 下载（-o - 打到屏幕）
tt file mirror path --at 正式区             # 本地镜像在哪（本机：F:\t100\prod）
tt file mirror pull --at 正式区             # 拉源码镜像（--profile source|interface，--full 全量重建）
tt file mirror dir --at 正式区              # 看/改某环境的镜像目录（--clear 回退全局）
```
> **有镜像就先读镜像**，比远程快一个数量级；`file` 系列只在镜像没覆盖时用。

---

## 5. 看前端画面（interface，.42f + .42s 合成）

```powershell
tt interface pages axmt500_wf --at 正式区                 # 有哪些页
tt interface show  axmt500_wf --depth 2 --at 正式区        # 页面树
tt interface show  axmt500_wf --fields --at 正式区         # 树里带字段
tt interface fields axmt500_wf --page page_1 --limit 30 --brief --at 正式区
tt interface fields axmt500_wf --col xmda071 --at 正式区    # 精确找字段
tt interface fields axmt500_wf --kw 备注 --at 正式区        # 模糊找字段
# 直接指定本地文件（不走镜像）
tt interface fields bsft001_wf --42f .\bsft001_wf.42f --42s .\bsft001_wf.42s --brief
```

---

## 6. 读代码大纲（outline）

```powershell
tt outline F:\t100\prod\erp\axm\4gl\axmt500_wf.4gl --human
tt outline <file.4gl> --brief --flat            # 扁平 + 省略行号
tt outline <file.4gl> --kind FUNCTION,INPUT     # 只看函数/INPUT 块
tt outline <file.4gl> --depth 2 --dir F:\t100\prod
```

---

## 7. 调试（debug / session）

```powershell
tt debug serve                     # 起后台调试服务（--stop 停；--foreground 前台）
tt debug ents --at 正式区          # 列企业（--cached 只读快照 / --refresh 现查）
tt debug resolve axmt500_wf        # 作业→源码定位
tt debug start axmt500_wf --ent 710 --env 正式区 --name A --timeout 120
tt debug status ; tt debug wait --for stopped --timeout 300
tt debug exec "break axmt500_wf.4gl:1234"           # 下断点（--file 批量）
tt debug exec "c" / "n" / "s" / "p g_xmda_m.xmdadocno"
tt debug bp add axmt500_wf.4gl:1234 --ignore 2      # 断点管理
tt debug bp add "g_xmda_m.xmdadocno" --kind watch   # 监视点
tt debug bp list ; tt debug bp del 1
tt debug locate axmt500_remaks                      # 按函数名找位置
tt debug source --path-only --from 100 --to 160     # 看源码片段
tt debug logs --tail 50
tt debug interrupt ; tt debug stop ; tt debug quit
tt debug wslogs --job 'wssp*' --fail --from 2026-10-01    # 接口日志
tt debug wsdebug <rowid> --set digi-body.std_data.parameter.x=1   # 报文重放
```
会话（多环境/多作业并存时用）：
```powershell
tt session new --env 正式区 --ent 710 --name A --prog axmt500_wf
tt session ls [-a] ; tt session show --session A
tt session ent 925 --session A        # 改会话企业（--clear 回默认）
tt session exec --session A -- "break 812"
tt session job --session A --prog apmt500
tt session quit --session A ; tt session rm --session A
```

---

## 8. 改设计器包（dev）

### .tzc 代码包（唯一会写包的是 apply）
```powershell
tt dev tzc export pkg.tzc -o .\ws          # 渲染带围栏的 4GL 工作区（只读包）
tt dev tzc status .\ws                     # 相对上次 export/apply 改了什么
tt dev tzc verify .\ws                     # 不写盘跑 gate1+gate2
tt dev tzc apply  .\ws -o pkg.tzc --dry-run   # 看 diff；去掉 --dry-run 才写包+commit
tt dev tzc unlock .\ws                     # 改框架区段前必须解锁（单向）
tt dev tzc newfn .\ws ; tt dev tzc rename .\ws old new
tt dev tzc selftest                        # 31 项内置对抗用例
cd D:\pkg\ws ; tt dev tzc status           # 进了工作区可省略 <dir>
```
### .tzs 表单包（读写都走引擎动词，55 个）
```powershell
tt dev tzs export pkg.tzs -o .\unzip [--force]
tt dev tzs formcheck --form42f a.42f --tzs pkg.tzs      # 应然↔实然静态比对
tt dev tzs open     --args '{"path":"D:\\pkg\\aapp320(c).tzs"}' --json
tt dev tzs form_tree --form aapp320 --args '{"depth":2}'
tt dev tzs set_spec_attr --form aapp320 --args '{"path":"...","kind":"field","attr":"can_edit","value":"Y"}'
tt dev tzs validate --form aapp320 ; tt dev tzs save --form aapp320 --args '{"out":"D:\\pkg\\_ai.tzs"}'
tt dev tzs doctor ; tt dev tzs stop ; tt dev tzs reap --yes
```
> 参数**一律写进 `--args` 的 JSON**（中文/长文本用 `--args-file`）；不要写 `--handle h9` 这种。

---

## 9. 画原型图（drawio）

```powershell
tt drawio lib --catalog                    # 先看有哪些形状
tt drawio lib -o .\dist                    # 导出形状库，drawio 里加载
tt drawio compose spec.json -o out.drawio  # 按 spec 生成 .drawio（省 -o 打到屏幕）
```

---

## 10. 五条常用组合拳

**A. 看懂一个字段的业务含义**
```powershell
tt dict r.t xmda_t --kw 备注 --at 正式区@710      # 字段说明
tt dict scc 11 --at 正式区@710                    # 相关分类码取值
tt file grep -rn 'xmda071' erp/axm/4gl --at 正式区 # 谁在用
tt dict r.t xmda_t --who --at 正式区@710           # 哪些程序动过
```

**B. 查一张单据（以采购单为例）**
```powershell
tt db sql --at 正式区@710 "select pmdlent,pmdldocno from pmdl_t where pmdldocno='WX-1501-25062701'"
tt db sql --at 正式区@710 "select * from pmdn_t where pmdndocno='WX-1501-25062701' and pmdnent=710"
```
> 先查单头确认**企业**（同单号可能多家都有），再按 ent 过滤明细。

**C. 追一个作业的代码**
```powershell
tt dict prog axmt500_wf --at 正式区               # 作业→程序→用到的表
tt outline F:\t100\prod\erp\axm\4gl\axmt500_wf.4gl --human
tt interface pages axmt500_wf --at 正式区          # 画面结构
```

**D. 改一版 .tzc 回到服务器**
```powershell
tt dev tzc export pkg.tzc -o .\ws
tt dev tzc verify .\ws
tt dev tzc apply .\ws -o pkg.tzc --dry-run     # 先看 diff
tt dev tzc apply .\ws -o pkg.tzc --yes
```

**E. 线上报错定位**
```powershell
tt dict msg --text 错误关键字 --at 正式区@710
tt debug wslogs --job 'wssp*' --fail --from 2026-10-20
tt debug wsdebug <rowid> --set <入参路径>=<值>
```

---

## 11. 容易踩的坑

1. **忘了 `--at`**：所有联网命令都要点名环境，没有默认值。
2. **忘了 `@企业`**：`db sql` 硬性要求；字典命令不带则用账号首项，可能查到的不是你要的那家。
3. **默认 JSON 看着累**：加 `--human`。
4. **`db sql` 只有 200 行**：写 `order by` + `fetch first`，或去 `tt dict spill list` 拿完整结果。
5. **中文变 `?`**：走 sqlplus 的通道编码问题，用 `rawtohex()` 取回本地解码。
6. **`.tzc` 手改工作区后忘 `verify` 直接 `apply`**：闸门会拦，但先 verify 省一轮。
7. **框架区段没 `unlock` 就想改**：会被拒；解锁是单向的，改了要重 export。
8. **`dev tzs` 拿 `--handle` 当开关**：参数全在 `--args` 的 JSON 里。
