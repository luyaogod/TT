# AGENTS.md

TT 是用于开发鼎捷数智旗下的大型ERP系统（T100）所构建的CLI工具。修改代码前请阅读[docs/DESIGN.md](docs/DESIGN.md)；编写文档请遵循[docs/AGENTS.md](docs/AGENTS.md)

## 项目结构

```
TT/
├─ AGENTS.md            仓库说明（AI/协作者必读）
├─ Makefile             必跑命令的快捷方式（make = build + test；深档目标自带副本闸门）
├─ config.example.json  配置文件样例（字段说明入口之一）
│
├─ internal/            Go 后端全部实现
│  ├─ cli/              cobra 命令树装配处：五条线 + 共用命令；唯一的接线处
│  │  ├─ common/        命令组间共享上下文（全局开关、MetaProvider、WebFS；叶子包防成环）
│  │  ├─ debug/         调试命令组门面（fgldb 那条线）
│  │  ├─ dev/           tt dev 命令组（三层，目录即分层）
│  │  │  ├─ register.go/run.go   cobra 树：路由、组级帮助、未知命令退 2、install 墓碑
│  │  │  │                       + 根开关桥 takeRootFlags + 退出码透传 exitCode
│  │  │  ├─ common/     两条线共用的脚手架（契约件只此一份）：错误信封 Fail、
│  │  │  │              稳定 JSON、ParseArgs、StripIdentitySuffix、
│  │  │  │              config `tdev` 节接缝、总帮助 Usage
│  │  │  ├─ tzc/        .tzc 代码包线：八个动词核心（export/status/verify/apply/
│  │  │  │              unlock/rename/newfn/selftest）+ 31 项对抗自检 + 接线清单 Verbs()
│  │  │  └─ tzs/        .tzs 表单包线：cmdTzs 分发、引擎定位/manifest、
│  │  │                 detail 渲染；整线 DisableFlagParsing 透传（动词面由引擎定义）
│  │  └─ dict/          字典命令组（查询 / db sync / mirror / spill）
│  ├─ config/           统一配置层：位置解析、唯一写入口 Edit、原子写、打码、缓存清单
│  │                    + 统一路径管理器 Locations（数据目录/缓存/服务状态/字典库/引擎 exe 落点）
│  ├─ debug/            调试子系统：会话管理、断点存档、源码镜像 srccache、执行日志、企业快照
│  ├─ dev/              .tzc 设计器包管线（纯管线，无命令层）
│  │  ├─ fence/         围栏协议：渲染/解析/字节级写回编辑
│  │  ├─ verify/        三道闸门 + 不变量 I1–I15
│  │  ├─ store/         工作区落盘：manifest、基线、原子写、锁
│  │  ├─ tzs/           .tzs 引擎客户端（管道协议、manifest、doctor；不 import cli/dev）
│  │  └─ split/ tapfile/ tglfile/ fgl/ model/ pkgfile/ synth/ testutil/
│  │                    写回计划、TAP/TGL 格式、FGL 处理、包模型、包格式、合成、语料夹具
│  ├─ dict/             字典子系统：db(本地 SQLite 查询) / dbsync(同步) / live(远程直查) / server(web 面)
│  ├─ host/             唯一的 SSH/远程服务器能力层（连接、登录、服务器侧只读探测）
│  ├─ erpdb/            客户端直连库封装（Oracle / 金仓，一个接口两个实现）
│  ├─ dbconfig/         配置里"库"那一节的连接模型
│  ├─ sshtun/           SSH 端口转发隧道（本地监听 → 远端 host:port）
│  ├─ entdir/           企业目录（ENT→账号）共享件：快照、指纹、新鲜期（debug 与 dict 共用）
│  ├─ web/              统一 Web 服务端：REST + WebSocket，嵌入前端 SPA，挂载调试/字典子系统
│  ├─ output/           唯一的输出出口（表格 / JSON / CSV 一个入口）
│  ├─ safesql/          只读 SQL 的文本层防线（远端查询入口用）
│  ├─ pathinstall/      用户 PATH（HKCU）增删，永不碰系统 PATH
│  ├─ winproc/          起/判/杀不随本进程死的子进程（Windows）
│  ├─ testenv/          本机测试环境解析（config.local.json + 环境变量，不带 testing）
│  └─ testkit/          测试助手（stdout 捕获、吞输出 Silent、跳过台账、单源检查）
│
├─ web/                 前端（npm workspaces，根）
│  ├─ app/              唯一 workspace 成员：React + vite 调试工作台/设置页
│  ├─ shared/           前端共享层（主题、设计系统；被 @source 扫描，动它要小心）
│  └─ dist/             构建产物，被 Go 嵌入（.gitkeep 是承重墙，删了 go build 失败）
│
├─ engine/              .tzs 表单引擎（C#，csc.exe 单独构建，只在改 engine/ 时重编）
│  ├─ src/              引擎 C# 源码（tzs-server）
│  ├─ designer/         设计器 28 个程序集（≈10.5MB，有意入库，保证全网同版设计器）
│  ├─ out/              构建产物 + 十几个探测程序（不进发行包、不能落 dist/ 下）
│  └─ test/ + build.sh / SPEC.md / HANDOFF.md  引擎测试、构建脚本、冻结契约、交接
│
├─ designer-src/        设计器反编译源码（只读）：16 工程 + 12 子包代码地图，
│                       引擎注释里的 file:line 核对落点
├─ installer/           MSI 安装包定义（WiX v3；perUser 装到 %LOCALAPPDATA%\Programs\TT）
├─ tools/               Go 小工具：zip(打包) / wixremovefolders(卸载清目录) / tzsmini(语料构建)
├─ testdata/            测试语料：tzs-mini(3.2MB 筛小钉住的标准件) / fgl-fixtures
├─ skills/              对外技能手册（tt-debug / tt-dict / tt-dev-tzc / tt-dev-tzs / tt-drawio / tt-erp-read），
│                       评测中执行者唯一能读的东西，与源代码同级重要
└─ docs/                AGENTS（文档规范：住哪、写什么、什么不许写）+ DESIGN（整体结构分层）
```

## 开发命令

所有开发过程中用到的命令由[Makefile](Makefile)统一管理