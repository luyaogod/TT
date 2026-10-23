---
name: tt-dev-tzc
description: 有关`.tzc`文件操作时，请激活此技能
---

# tt dev tzc：安全地改 `.tzc` 代码包

`.tzc` 是 T100 设计器的代码包（一个 zip，装着 FGL 程序、`.tap` 定义、编译好的 `.4gl` 等条目）。
二进制包不能用文本编辑器直接改，`tt dev tzc` 把它渲染成一个带围栏的 4GL 工作区，
让人或 AI 只改该改的那几行，再由工具校验后写回包。

记住一句话：`export` 读包、改工作区、`apply` 写包。

表单包（`.tzs`/`.tzv`）是另一条互不相通的管线，见 `tt-dev-tzs`。拿错入口会被挡住：
`.tzc` 跑 `tzs export` 会以退出码 2 失败。

## tzc文件基本处理操作

```powershell
tt dev tzc export "D:\pkg\capt110(c).tzc"   # 包 → 工作区 D:\pkg\capt110-ws（-o 可改）
cd /d D:\pkg\capt110-ws                     # 进工作区后所有命令都能省 <dir>

# 只改 prog.full.4gl 里标着 EDITABLE 的围栏正文
tt dev tzc status                           # 改了什么、在第几行、会不会被拒
tt dev tzc verify                           # 不写盘预检
tt dev tzc apply                            # 校验通过才写回包（先备份旧包）
```

工作区长这样：`prog.full.4gl`（可编辑的渲染结果）、`manifest.json`（源包身份与 sha256）、
`snapshot/`（逐条目快照）、`.tdev/`（比对基线 `base.full.4gl`、上一版包）、`.git/`（每次 apply 一次提交）。

## 在注释区添加摘要注释并说明修改人

首先需要想向用户索要本次修改的标记`<单号>#<项次>`、工号，用于作为本次修改的唯一标识，如果用户拒绝提供，就自行生成一个不带分隔符号的16位uuid即可

1. 在global.memo 这个 add-point 内添加一行本次修改的一个标题（摘要）格式如下：

```
#<单号>#<项次> <作业日> By <工号> <一句说明>
```

如果用户未提供，示例如下：

```
#F4F95D14 20261021 By agent XXXXXX
```

修改摘要的添加位置默认在客制区，除非用户有特别要求：

```
#add-point:填寫註解說明(客製用) name="global.memo_customerization"
...
#F4F95D14 20261021 By agent XXXXXX
#end add-point
```

根据摘要的内容，涉及到的代码修改用下面示例的块包裹：

```
#<单号>#<项次> -s
#<单号>#<项次> -s
```

同理如果用户没有提供的，就使用标题（摘要）的uuid

```
#F4F95D14 s
#F4F95D14 e
```

## 关于“围栏”

`prog.full.4gl` 不是孤立文件，它是一份从包渲染出来、还要原样写回去的映射。
所以文档用 `{//@tdev:begin point …}` / `{//@tdev:end …}` 这类行把「哪些字节属于哪个定义点」标死：

- 围栏行本身是坐标，改了就整份映射作废，gate1 直接拒。
- 围栏外的字节（文件头尾、区段之间的空行、缩进）必须逐字节不变，所以不要用格式化工具重排整个文件。
- 结构行（`END FUNCTION` / `END DIALOG` / `END REPORT`）逐字节比对，同样不能改。

## 不能做的事

1. 不要改围栏行、结构行、围栏外的任何字节。只改围栏正文。
2. 不要手改 `.tap`、不要绕开工具直接改 zip：写回包的唯一路径是 `apply`。
3. 不要编辑 `.tdev/base.full.4gl`：那是 gate1 的比对基线，改它等于伪造验证。
4. 结构变化走工作区动词：改函数签名用 `tt dev tzc rename`（同步围栏 `fn`、签名行、描述块），
   加新函数用 `tt dev tzc newfn`（带设计器模板，它没有 `--desc`）。
5. 不要硬改只读区段：先 `tt dev tzc unlock`。解锁单向不可逆（解开后规格调整不再自动生成程序代码），
   `--yes` 必须由人确认，AI 不得自主加；锚点区段即使解锁也依然只读。
6. 不要动真实语料或客户包：`export` 只读源包，实验一律用副本或临时目录。
7. 源包在 export 之后被别人动过，就重新 `export`：`apply` 会发现 sha256 不符并拒写（退出码 5）。
8. 如果用户的需求必须要改代码框架，请如实告知用户，询问用户是否需要解开框架，如果用户明确批准，你可以执行命令解开代码框架

## 报错怎么读

`verify` / `apply` 拒绝时会打位置，照着改就行（删除类问题给基线文件的行号）：

```
apply 被拒绝（gate1 error=1 warn=0 info=1）：
  [error] gate1.readonly-region  只读 Region 内容被改动（写入被拒）
            位置：prog.full.4gl:4        ← 你编辑的那份文件 + 行号
            该行：   # 我加的字           ← 该行内容
            deny=section-locked          ← 原因：框架未解开，先 unlock
```

退出码：`0` 成功 / `2` 包格式或用法错 / `3` 验证失败 / `4` 写入被拒（权限、缺 `--yes`）/
`5` IO 与环境失败。（`.tzs` 那条线没有 `3`、多一个 `1`。）
加 `--json` 时 stdout 恰好一个 JSON 对象，每条发现带 `file`/`line`/`snippet`。

## 其余细节看帮助

八个动词、每个开关与默认值都在命令行里：`tt dev tzc --help`、`tt dev tzc apply --help`……
（`export` 渲染工作区、`status` 看改动、`verify` 预检、`apply` 写包、`unlock` 解锁、
`rename` 改名、`newfn` 新增、`selftest` 31 项内置对抗用例，不需要真实语料。）
