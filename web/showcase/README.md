# web/showcase — 前端 UI 决策演示台

用**官方 shadcn 组件**与 **shadcn 默认明暗双色主题**，把现有前端的全部功能点逐块重建的一个页面。
它服务的是"迁移前先看清长什么样、有哪些坑、按什么顺序做"，**不是产品的一部分**。

产品 SPA 在 [`../app`](../app/README.md)；本目录的产物与依赖都与它无关。

## 做什么

两种看法，左上角切换：

**目标态整机**（`src/virtual/`）—— 按"已定的最佳实践"组装的一台能点的虚拟 TT 前端：左图标栏四个视图（调试 / 接口日志 / 服务测试 / 设置）、源码页签、右侧手风琴面板、下方面板与 fgldb 命令输入、浮动调试工具条、整条语义色状态栏；会话状态与协作模式可在标题栏里切，看控件可用性与状态栏如何跟着变。它回答的是"拼起来成不成立"，不是"某个组件长什么样"。

**逐块台账**（`src/sections/`）—— 现有前端每个功能点一块：实时预览 + 现状落点 `file:line` + 目标组件 + 风险级别 + 迁移要点；另有决策总览（口径、风险分布、推进顺序、可跳转的完整台账）。

两块共用的东西：真实文案与真实字段（预览里的中文标签、提示语、空状态都取自现有源码的逐块盘点，而不是占位文本）。

## 不做什么

- **不连后端**：数据全是固定夹具，页面在没有 `tt serve` 的情况下也能打开与评估。
- **不加载编辑器**：源码编辑器与报文编辑器区是静态示意（Monaco 有自己的主题与 worker，
  它属于"哪些必须手写"的结论，不属于"换成什么 shadcn 组件"的演示）。
- **不产出产品产物**：不写 `../dist`，不被 Go 内嵌，不参与 `make build`。
- **不提供业务能力**：没有路由、没有鉴权、没有接口调用；唯一的交互是预览（主题切换、展开、弹层、排序示意）。

## 独有机制与不变量

**独立包，不进 npm workspace。** `web/package.json` 的 `workspaces` 只有 `app`。原因是本机（未开开发者模式）
给 workspace 建链接需要管理员权限，`npm install` 会直接失败。所以本目录有自己的一份
`package.json` / `package-lock.json` / `node_modules`。

**组件源码不手改。** 36 个组件全部由官方 CLI 从 registry 拉取：

```bash
cd web/showcase
npx shadcn@latest add button badge card … --yes --overwrite
```

拉的是 registry 的 `new-york-v4` 通道：组件从统一包 `radix-ui` 引入、`cn` 从官方 `cn` 包引入。
要升级组件就再跑一次 `--overwrite`，改动会以 git diff 呈现。

**主题有一套、只有一个来源。** `components.json` 指定 `baseColor: neutral`、`cssVariables: true`；
`src/index.css` 的 token 是 registry `/r/colors/neutral.json` 的逐字取值（含 `--radius: 0.625rem`）；
运行时主题走 `next-themes`（`attribute="class"`、`storageKey="tt.theme"`），与产品同一个键、同一组取值，
首屏防闪色由 `index.html` 的内联脚本承担。

**唯一对默认的扩展是三个语义色。** shadcn 默认没有 success/warning/info，而 TT 的信息设计有四类状态
（成功 / 警告 / 信息 / 危险）。`index.css` 里补了 `--success`、`--warning`、`--info` 三组取值（亮暗各一套），
状态栏、`src/virtual/status.tsx` 的 `StatusBadge` / `StatusNote`、Alert 共用同一份来源。
registry 组件源码一个字节都没改 —— 语义包装在应用层，升级时 `--overwrite` 直接覆盖。

**分享链接**：`?theme=light|dark` 写一份主题偏好（随后摘掉参数，切换照常可用）；
`?view=debug|logs|test|settings` 直接落到虚拟应用的某个视图；`?mode=ledger` 进逐块台账。

**台账是唯一出处。** `src/lib/catalog.ts` 定义分组、顺序、锚点、风险与迁移要点；
左导航、总览表、每块的注记、风险统计全部读它。加/减功能点只改这一处，别在 section 里另写一份标题。

## 判据

```bash
cd web/showcase && npm install        # 首次；独立依赖，不走 workspace
cd web/showcase && npm run check:render   # 渲染检查：6 屏真实渲染 + 结构不变量
cd web/showcase && npm run build      # tsc --noEmit + vite build
cd web/showcase && npm run dev        # 默认 http://localhost:5174
```

- **期望结果**：`check:render` 逐屏报 `ok`（每屏附字符数、关键内容项数与锚点数，共 12 屏：6 屏台账 + 虚拟应用四个视图 + 设置页两个分区），
  末行 `渲染检查通过:6 屏全部渲染成功`；`npm run build` 通过（类型检查在构建里，失败即退出），产出 `web/showcase/dist/`。
- **这一层负责证明哪件事**：① 每个功能点在"纯 shadcn 默认"下的呈现与迁移风险**已被逐块看见**，且整机拼装也成立（人可以据此拍板）；
  ② 12 屏都能用真实组件树渲染成功（Radix 的 Provider、组件用法、导出名出错会在这里炸，而不是在用户眼前炸）；
  ③ **结构不变量**：`src/lib/catalog.ts` 登记的每一个功能点都在对应屏里渲染出了锚点 —— 台账、左导航与预览三者的块数一致。
- **故意不覆盖什么**：不覆盖交互行为与视觉正确性（点击、键盘、拖拽、样式对不对都不验），
  也不覆盖组件源码与主题 token 的正确性 —— 那些由 registry 保证（要升级就 `--overwrite` 重拉，改动进 git diff）。
  演示台不连后端，所以任何需要真实会话的行为都不在这里证明。

## 不用了怎么删

删掉整个 `web/showcase/` 即可；除本目录外，仓库里只有两处提到它（`web/README.md` 的目录表、
根 `AGENTS.md` 的项目结构），一并去掉即可，产品侧无任何引用。

## 细节去哪

- 迁移台账（功能点 / 落点 / 目标组件 / 风险 / 要点）与推进顺序 → [`src/lib/catalog.ts`](./src/lib/catalog.ts)
- 目标态虚拟应用（外壳、四个视图、最佳实践清单、Promise 化确认器、语义状态件）→ [`src/virtual/`](./src/virtual)
- 真实文案与字段夹具（来源：`web/app/src` 的逐块盘点）→ [`src/lib/fixtures.ts`](./src/lib/fixtures.ts)
- 每个功能点的预览实现 → [`src/sections/`](./src/sections)
- 渲染检查（SSR 通道 + 锚点断言）→ [`scripts/smoke.mjs`](./scripts/smoke.mjs)、[`scripts/smoke-entry.tsx`](./scripts/smoke-entry.tsx)
- 主题取值与来源 → [`src/index.css`](./src/index.css) 顶部注释、[`components.json`](./components.json)
- 产品 SPA 与设计系统 → [`../app/src/README.md`](../app/src/README.md)、[`../shared/README.md`](../shared/README.md)
