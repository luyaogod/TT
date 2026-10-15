// 迁移台账:现有前端每一个功能点 → 目标 shadcn 组件 + 风险级别 + 迁移要点。
//
// 事实来源:对 web/app/src/*.tsx 与 web/shared/*.tsx 的只读盘点(逐块给到 file:line)。
// 台账只写"改什么、为什么要小心",不写实现过程;预览由各 section 里的真实组件承担。
//
// 风险级别的口径:
//   low  —— 纯展示或纯受控表单,换组件不动行为
//   mid  —— 受控状态/派生禁用/写死色语义要逐个定性,换错会静默改行为
//   high —— 有滚动祖先、键盘流、Monaco、异步确认、性能上限等硬约束,必须结构照搬

export type Risk = 'low' | 'mid' | 'high'

export interface Block {
  /** 锚点 id,左导航与台账跳转共用 */
  id: string
  title: string
  /** 现状落点(仓库内相对路径:行) */
  source: string
  /** 目标:官方 shadcn 组件名 */
  target: string[]
  risk: Risk
  /** 迁移要点:写"必须保住的那件事",不是描述现状 */
  note: string
}

export interface CatalogSection {
  id: string
  title: string
  /** 这一屏回答的决策问题 */
  question: string
  blocks: Block[]
}

export const RISK_LABEL: Record<Risk, string> = { low: '低', mid: '中', high: '高' }

/** 台账与左导航的唯一出处:分组、顺序、锚点都在这里 */
export const CATALOG: CatalogSection[] = [
  {
    id: 'foundation',
    title: '基元与语义色',
    question: 'shadcn 默认口径下,按钮/徽标/输入/下拉/浮层长什么样,现有写死色如何映射到语义 token?',
    blocks: [
      {
        id: 'f-button',
        title: '按钮(六变体 × 四尺寸)',
        source: 'web/shared/ui.tsx:10-40',
        target: ['Button'],
        risk: 'low',
        note: '现有 cva 变体与官方基本同构,但尺寸更矮(h-8/h-7)、且硬写 rounded-[6px] 绕过零圆角 token。改用官方 Button 后高度回到 h-9/h-8、圆角跟 --radius,全站密度与圆角口径同时改变 —— 这是"是否保留编辑器式紧凑"的决策点本身。',
      },
      {
        id: 'f-badge',
        title: '状态徽标(成功/警告/危险/信息/中性)',
        source: 'web/shared/ui.tsx:42-62',
        target: ['Badge'],
        risk: 'mid',
        note: '现有 tone 用 emerald/yellow/red/sky 写死色板,官方 Badge 只有 default/secondary/destructive/outline。状态码、HTTP 码、忙碌提示合起来有 5 种语义,需要先定"语义 → variant + 少量自定义类"的映射,否则会退化成只有红/灰两档。',
      },
      {
        id: 'f-input',
        title: '输入框',
        source: 'web/shared/ui.tsx:78-92',
        target: ['Input'],
        risk: 'mid',
        note: '现有高度 h-8、带 focus ring;但日志/服务测试的查询框刻意用 QUERY_INPUT 压掉了 ring 且只有 h-7。换官方 Input 会把搜索框变高并恢复聚焦环 —— 需要决定查询框是否保留紧凑覆盖。',
      },
      {
        id: 'f-select',
        title: '下拉选择',
        source: 'web/shared/ui-radix.tsx:16-83',
        target: ['Select'],
        risk: 'low',
        note: '已经是 Radix 实现、与官方同构;差别在官方多了滚动按钮、分组标签与尺寸参数。可直接换,注意 SelectItem 的选中勾由官方渲染。',
      },
      {
        id: 'f-checkbox',
        title: '勾选框',
        source: 'web/shared/ui-radix.tsx:109-126',
        target: ['Checkbox'],
        risk: 'low',
        note: '设置页唯一那处"允许 AI 执行只读 SQL"用的是原生 input[type=checkbox],不是这个基元;迁移时要把它一并收编,而不是只换已有的两处。',
      },
      {
        id: 'f-switch',
        title: '开关(替代原生复选)',
        source: 'web/app/src/SettingsView.tsx:926',
        target: ['Switch'],
        risk: 'mid',
        note: '该处语义是"默认开启的布尔权限",适合 Switch;但同一张卡里的"断点持久化"是三态(跟随默认/开/关),Switch 表达不了三态 —— 官方也没有三态开关,那处应保留 Select。',
      },
      {
        id: 'f-tooltip',
        title: '提示(约 20 处原生 title)',
        source: 'web/app/src/SettingsView.tsx:792,822,863,932… / Panels.tsx:49,262 / Toolbar.tsx:24',
        target: ['Tooltip'],
        risk: 'mid',
        note: '全站靠原生 title 承载解释性文案(含"协作模式下由 AI 主导"这类行为说明)。换 Tooltip 要逐处改,且要保留 disabled 元素上的提示 —— 原生 title 在 disabled 按钮上仍显示,Radix Tooltip 需要包一层容器才能做到。',
      },
      {
        id: 'f-progress',
        title: '长跑任务进度',
        source: 'web/app/src/SettingsView.tsx:158-186(轨道 + 填充 + bar-indeterminate)',
        target: ['Progress'],
        risk: 'mid',
        note: '现有"不确定进度"类 bar-indeterminate 在仓库里没有任何 CSS 定义,动画实际是失效的(渲染成静止的 1/3 宽色条)。换 Progress 顺手修掉这个静默缺陷,但要确认它是有意留白还是回归。',
      },
      {
        id: 'f-dialog',
        title: '危险操作确认',
        source: 'web/shared/ui-radix.tsx:263-319(AlertDialog) + SettingsView.tsx:589,665,678 三处 window.confirm',
        target: ['AlertDialog', 'Dialog'],
        risk: 'high',
        note: '三处 window.confirm 依赖同步布尔返回值(全量重建镜像/字典同步/升级 tt),Radix 确认框是异步的,必须重写成 Promise 或状态机 —— 这是最容易"顺手换成 Dialog"而改错行为的一处。',
      },
      {
        id: 'f-alert',
        title: '内联结果条(成功/失败/信息/警告)',
        source: 'web/app/src/SettingsView.tsx:942-949,1276-1279 / WsLogView.tsx:406-408 / WsTestView.tsx:93',
        target: ['Alert'],
        risk: 'mid',
        note: '四处风格各写一遍(border-l-2 竖条、圆角块、整条横幅),颜色都是写死的 emerald/red/sky/amber。官方 Alert 只有 default/destructive 两态,四色语义需要保留自定义类或收窄语义 —— 决策点:是否接受"警告"并入 destructive。',
      },
      {
        id: 'f-tabs',
        title: '页签(三处手写)',
        source: 'web/app/src/SettingsView.tsx:828-836 / WsLogView.tsx:483-489 / Panels.tsx:464-475',
        target: ['Tabs'],
        risk: 'mid',
        note: '设置页切页签不卸载状态(showPwd/addAcct/accProbe 靠这一点存活)、日志详情页签要能塞进脏标记 •、时间线页签是"下划线压住分割线"的编辑器样式。换 Tabs 必须确认 forceMount 与样式覆盖都能表达这三件事。',
      },
      {
        id: 'f-scroll',
        title: '滚动区',
        source: 'web/shared/tokens.css:100-128(全局 webkit 滚动条)',
        target: ['ScrollArea'],
        risk: 'high',
        note: 'ScrollArea 内部多一层 overflow:hidden 的视口,会改变滚动祖先 —— 任何 sticky 表头(日志列表、账号表)都会失效。建议只在无 sticky 的普通面板用,列表类保持原生 overflow-auto。',
      },
    ],
  },
  {
    id: 'workbench',
    title: '调试工作台外壳',
    question: '编辑器式布局(活动栏 / 页签 / 浮动工具条 / 状态栏 / 手风琴面板 / 树)在 shadcn 默认风格下如何成立?',
    blocks: [
      {
        id: 'w-activity',
        title: '活动栏(左右两条图标栏)',
        source: 'web/app/src/App.tsx:26-37,267-272,328-334',
        target: ['Button', 'Tooltip'],
        risk: 'low',
        note: '现为手写 button + 硬写 rounded-[6px] 与 accent 选中态。官方等价物是 ghost/icon 按钮 + Tooltip 标签;若要更强的"选中即按钮"观感可用 Toggle variant="outline"。',
      },
      {
        id: 'w-statusbar',
        title: '状态栏(整条底色表达调试状态)',
        source: 'web/app/src/StatusBar.tsx:35-56,64-93',
        target: ['Badge', 'Button'],
        risk: 'mid',
        note: '底色是写死 hex(停站 #cc6633 / 运行 #0078d4),而 shadcn 默认 token 里 primary 在亮色是近黑、暗色是近白 —— 直接把状态栏铺成 primary 会得到"黑条/白条"。需要为"停站/运行"定义语义色(建议 chart-* 或少量自定义 token),这是本次改造最明显的默认主题缺口。',
      },
      {
        id: 'w-tabs',
        title: '源码页签栏(左滚右固定 + 关闭钮)',
        source: 'web/app/src/SourceView.tsx:239-272',
        target: ['Button', 'Badge'],
        risk: 'high',
        note: '结构硬约束:overflow-x-auto 只包页签,右端两个面板收展按钮 shrink-0 不跟着滚走;活动页签靠 .tab-active 的 inset 主色细线与编辑区同底。Radix Tabs 难保这个结构,建议保留手写布局、只把按钮换成官方 Button。',
      },
      {
        id: 'w-toolbar',
        title: '浮动调试工具条(可拖拽 + 位置记忆)',
        source: 'web/app/src/Toolbar.tsx:35-90',
        target: ['Button', 'Tooltip'],
        risk: 'mid',
        note: '工具条 six 个动作靠颜色区分语义(继续/步过/步入/步出=sky,重启=green,结束=red)。官方 Button 只有变体色,green 主行动需要自定义或改用 primary/destructive 两档 —— 决策点:是否收窄到官方色阶。',
      },
      {
        id: 'w-accordion',
        title: '右侧面板手风琴(调用栈/自动变量/监视/断点)',
        source: 'web/app/src/Panels.tsx:90-218',
        target: ['Accordion', 'Badge'],
        risk: 'mid',
        note: '共享 Accordion 刻意做成"中性的壳",把撑满高度与分割线留给调用方 —— 因为设置页那套是轻量树形、需求相反。换官方 Accordion 后要重新确认:数据区 flex-1 自滚、条目间 1px 线、开合受控(自动开关联动)。',
      },
      {
        id: 'w-tree',
        title: '树形(变量树 / 大纲)',
        source: 'web/app/src/VarTreeUi.tsx:6-48 / OutlinePanel.tsx:145-173',
        target: ['Collapsible', 'ScrollArea'],
        risk: 'mid',
        note: '两处都是手写递归 div + ▸/▾ 文本箭头,且大纲按深度着色(顶层=foreground,一层=sky,更深=muted)。官方 Collapsible 可替换单个节点,但"缩进用 paddingLeft 计算 + 高亮跟随编辑器光标"要保留。',
      },
      {
        id: 'w-timeline',
        title: '时间线 / 原始协议流 + 命令输入(↑↓ 历史)',
        source: 'web/app/src/Panels.tsx:417-516',
        target: ['Tabs', 'Badge', 'Input'],
        risk: 'mid',
        note: '命令输入是"无边框 + > 提示符 + ↑↓ 翻历史"的控制台语义(README 里认这条)。换 Input 需要覆盖输入框外观,历史键处理必须自己保留(官方无此组件)。',
      },
      {
        id: 'w-empty',
        title: '空状态与遮罩(Mónaco 无源码 / 忙碌)',
        source: 'web/app/src/SourceView.tsx:290-316 / WsLogView.tsx:305-313 / Panels.tsx:480',
        target: ['Alert', 'Skeleton'],
        risk: 'low',
        note: '三处空状态文案与形态各不相同(居中块、居中卡片、左对齐一行)。可统一到官方 Empty/Alert 风格,但"该模块无源码(仅 42m)"这类信息必须逐字保留。',
      },
    ],
  },
  {
    id: 'interfaces',
    title: '接口日志与服务测试',
    question: '数据密集页(筛选工具条 / sticky 表头长表格 / 详情页签 / 报文编辑器)在 shadcn 下哪些能换、哪些必须手写?',
    blocks: [
      {
        id: 'i-filter',
        title: '筛选工具条(5 个查询框 + 日期范围 + 复选)',
        source: 'web/app/src/WsLogView.tsx:322-403',
        target: ['Input', 'Checkbox', 'Calendar', 'Popover', 'Button'],
        risk: 'mid',
        note: '文本条件"回车才生效"、日期与"仅失败"即时生效 —— 这个不对称是有意的(注释里记着历史 bug)。换 Input 只改外观,不要把 onChange 改成即时提交。',
      },
      {
        id: 'i-table',
        title: '日志表(11 列 + sticky 表头 + 行选中)',
        source: 'web/app/src/WsLogView.tsx:414-471',
        target: ['Table', 'Badge'],
        risk: 'high',
        note: '四条硬约束缺一不可:container={false} 自供滚动祖先、table-fixed + border-separate、单元格只留右下边框、表头背景不透明且与列表同色。官方 Table 默认带 overflow-x-auto 容器与 bg-muted/50 表头,直接替换会让 sticky 失效或让行从表头透出。',
      },
      {
        id: 'i-paging',
        title: '分页与排序(仅当前页)',
        source: 'web/app/src/WsLogView.tsx:391-402,226-252',
        target: ['Button', 'DropdownMenu'],
        risk: 'mid',
        note: '无虚拟滚动是有意为之(一页上限 200/500);排序循环 升序→降序→默认,且空值与 NaN 恒排最后。引入官方 DataTable(服务端排序)会变更这两条语义。',
      },
      {
        id: 'i-detail',
        title: '报文详情(页签 + 工具栏 + 15 行键值)',
        source: 'web/app/src/WsLogView.tsx:474-594',
        target: ['Tabs', 'Button', 'Badge', 'Item'],
        risk: 'high',
        note: 'Esc 关闭目前由 window keydown + preventDefault 自己管;换成 Radix Dialog/Sheet 后 Esc 会被上层消费,必须明确谁负责关闭,否则会出现"关两次"或"关不掉"。request/response 页必须保持 overflow-hidden(滚动交给 Monaco),否则出双滚动条。',
      },
      {
        id: 'i-copy',
        title: '复制反馈(1.5s 复位,失败态变红)',
        source: 'web/app/src/WsLogView.tsx:512-539,267-294',
        target: ['Button'],
        risk: 'low',
        note: '三态(复制/已复制/复制失败)是按钮自身的文案与颜色切换,不弹 toast —— 迁移时如果引入 Toast,要保留按钮上的即时反馈,否则用户在按钮上得不到确认。',
      },
      {
        id: 'i-wstest',
        title: '服务测试(方式下拉 + 动态 URL + 双栏报文)',
        source: 'web/app/src/WsTestView.tsx:48-125',
        target: ['Select', 'Input', 'Checkbox', 'Button', 'Card', 'Badge'],
        risk: 'mid',
        note: '两个面板头与官方 Card 语义冲突:官方 Card 自带 gap-6/py-6,而这里必须 min-h-40 + flex-1 且零多余内边距(Monaco 的 automaticLayout 要拿到确定高度)。用 Card 必须覆盖默认间距,或改用 Separator + 手写头。',
      },
      {
        id: 'i-payload',
        title: '报文编辑器三态(加载中/空/编辑器)',
        source: 'web/app/src/PayloadEditor.tsx:29-36',
        target: ['Skeleton', 'Empty'],
        risk: 'mid',
        note: '占位态底色必须等于 Monaco 的 editor.background(现有注释写死了这条对应关系)。换官方 Empty/Skeleton 时别引入 bg-muted,否则占位块与编辑器深一块浅一块。',
      },
    ],
  },
  {
    id: 'settings',
    title: '统一设置页',
    question: '设置页骨架(左导航 / 卡片 / 字段行 / 可编辑表格 / 长跑任务 / 只读信息)如何映射到 shadcn 默认组件?',
    blocks: [
      {
        id: 's-shell',
        title: '设置页骨架(左导航 176px + 右滚动列)',
        source: 'web/app/src/SettingsView.tsx:714-762',
        target: ['Sidebar 或 Tabs', 'Card', 'Separator'],
        risk: 'mid',
        note: '左导航与深链接(#settings/分区)共用一份分区清单。若用官方 Sidebar,注意它是"应用级导航"组件(含 cookie 持久化与移动端 Sheet),对设置页可能过重;更轻的选择是 Tabs 竖排或保留手写按钮。',
      },
      {
        id: 's-card',
        title: '设置卡片(标题 + 说明 + 右侧操作位)',
        source: 'web/shared/settings.tsx:18-50',
        target: ['Card'],
        risk: 'mid',
        note: '现卡片刻意"卡内不打分割线",且 panel 模式自带 flex-1 min-h-0 overflow-auto(站点管理两栏靠它定高)。官方 Card 的 CardContent 默认 py-6 + gap-6,套进定高面板会顶出边框 —— 必须覆盖默认间距。',
      },
      {
        id: 's-row',
        title: '设置行(标签在上 / 说明其次 / 控件在下)',
        source: 'web/shared/settings.tsx:62-79,82-91',
        target: ['Field', 'Label'],
        risk: 'low',
        note: '官方 Field 族的 FieldLabel/FieldDescription 与这套版式几乎一比一对应,是本次改造收益最高的一处。注意 SettingRow 的控件宽度统一收到 max-w-md 这条约束要保留,否则长文本输入会被拉成细线。',
      },
      {
        id: 's-accounts',
        title: '账号可编辑表格(格内输入 + Enter 提交)',
        source: 'web/app/src/SettingsView.tsx:951-1018',
        target: ['Table', 'Input'],
        risk: 'high',
        note: '三重耦合:自供滚动容器(container={false})、表头必须不透明、单元格用无边框填满格的 .cell-input + 聚焦环。且 Enter 即提交,任何包装组件吞掉 keydown 都会静默失效(无测试覆盖)。',
      },
      {
        id: 's-confirm',
        title: '删除环境确认(确认即落盘)',
        source: 'web/app/src/SettingsView.tsx:1427-1441,469-477',
        target: ['AlertDialog'],
        risk: 'high',
        note: '已是官方 AlertDialog 形态。真正要保住的是:确认后"删除 + 立即保存",且被删的是默认环境时默认环境要顺延到首条(否则后端 400)。换成"先改状态再统一保存"会踩这个坑。',
      },
      {
        id: 's-dirty',
        title: '脏标记与保存按钮(按配置节粒度)',
        source: 'web/app/src/SettingsView.tsx:701-708,318-331,408-431',
        target: ['Button', 'Badge'],
        risk: 'high',
        note: '脏标记按 7 个配置节分别跟踪,且保存按钮"不脏时根本不渲染"(按钮存在性本身是状态表达)。换成常驻 disabled 按钮或全局脏标记,会回归"用陈旧快照覆盖别张卡刚改好的节"这个已修复的 bug。',
      },
      {
        id: 's-progress',
        title: '长跑任务(镜像拉取 / 字典同步 + 800ms 轮询)',
        source: 'web/app/src/SettingsView.tsx:1055-1192,652-661',
        target: ['Progress', 'Alert', 'Button'],
        risk: 'mid',
        note: '两个任务的阶段语义与百分比算法不同,轮询只在 running 时挂载。换 Progress 时不要把轮询挪到组件常驻 effect,否则空闲时也在打后端。',
      },
      {
        id: 's-theme',
        title: '外观主题卡(亮/暗/跟随系统)',
        source: 'web/app/src/SettingsView.tsx:105-155,1287-1290',
        target: ['ToggleGroup 或 Card + Button'],
        risk: 'low',
        note: '三张预览卡里的 hex 是"预览内容"而非主题取值 —— 不要被"禁止写死色板"的规矩扫掉。换官方组件时保留 aria-pressed 与整卡可点。',
      },
      {
        id: 's-readonly',
        title: '只读信息行与缓存清单',
        source: 'web/shared/settings.tsx:110-121 / SettingsView.tsx:1338-1420',
        target: ['Item', 'Separator', 'Table'],
        risk: 'low',
        note: '纯展示,收益高、风险低:官方 Item/描述列表能替掉 w-28 + font-mono 的手写两列。注意长路径要保留 break-all,否则长路径会撑破容器。',
      },
      {
        id: 's-info',
        title: '错误/状态文案(顶部横幅 + 每卡就近 note)',
        source: 'web/app/src/SettingsView.tsx:304,312,336,746-753',
        target: ['Alert', 'Badge'],
        risk: 'mid',
        note: '"每张卡各自的 note"是刻意就近显示的(错误跟着出错的卡片),不要统一收编到页面顶部或 Toast,否则用户要来回找。',
      },
    ],
  },
  {
    id: 'patterns',
    title: '必须手写、不要换的部分',
    question: '哪些地方不能换成 shadcn 组件,只能保持手写(最多把写死色 token 化)?',
    blocks: [
      {
        id: 'p-monaco',
        title: '源码编辑器与报文编辑器(Monaco)',
        source: 'web/app/src/SourceView.tsx:274-288 / PayloadEditor.tsx:36-66',
        target: [],
        risk: 'high',
        note: 'Monaco 不是 shadcn 的地盘:主题必须与页面明暗联动(现在对齐的是 --card,换 shadcn 默认 token 后要重新对齐),editorOptions 引用必须稳定(否则重渲染复位滚动),automaticLayout 要求父链高度确定。',
      },
      {
        id: 'p-hover',
        title: '变量悬浮卡片(Monaco content widget)',
        source: 'web/app/src/fglHover.tsx:118-189',
        target: [],
        risk: 'high',
        note: '卡片渲染进 Monaco 挂载节点、不在 React 树里,Radix 的 Tooltip/Popover/HoverCard 全都用不了(没有锚点 DOM、Portal 会被 minimap 盖住)。只能把 .fgl-hover* 里的写死值 token 化。',
      },
      {
        id: 'p-gutter',
        title: 'gutter 装饰(断点红点 / 当前行黄箭头)',
        source: 'web/app/src/index.css:12-46 / SourceView.tsx:143-170',
        target: [],
        risk: 'mid',
        note: '命令式 CSS 类,且颜色有两处来源(overviewRuler 写死 #ef4444/#eab308 与 CSS 里的同值)。要么全部 token 化,要么原样保留 —— 只改一半会出现"尺子上黄、行上一半黄"。',
      },
      {
        id: 'p-split',
        title: '分栏拖拽(比例持久化 + 反馈环锁)',
        source: 'web/app/src/WsLogView.tsx:169-210,410-413 / App.tsx:179-225',
        target: [],
        risk: 'high',
        note: '换官方 Resizable 会丢掉:比例持久化键、比例夹取区间、5px 占位、hover 主色加粗、以及"左栏 min-w-0 + 面板 shrink-0"这条防宽度反馈环的锁。收益低、风险高,建议保留。',
      },
      {
        id: 'p-collab',
        title: '协作模式闸门(界面不装作能做)',
        source: 'web/app/src/SourceView.tsx:66-74 / Toolbar.tsx:106 / Panels.tsx:251',
        target: [],
        risk: 'high',
        note: '协作模式下断点、选帧、命令、TOPENT 一律禁用,服务端同样 403。这是行为契约而非样式:任何"顺手修好"的改造都会让界面撒谎。',
      },
    ],
  },
]

/** 建议的推进顺序:先低风险高收益,再动硬约束 */
export const ROADMAP: { phase: string; scope: string; why: string }[] = [
  {
    phase: '第一步',
    scope: '主题层替换:shadcn 默认 token(neutral)+ 圆角口径统一',
    why: '所有后续工作都建立在同一套语义 token 上;此刻同时解决"tokens 清零圆角、Button 却硬写 6px"的口径裂缝。',
  },
  {
    phase: '第二步',
    scope: '基元层:Button / Input / Badge / Select / Checkbox / Switch / Tooltip / Alert / Progress',
    why: '低风险高收益,且能立刻消除写死色板;Badge 与 Alert 的语义映射要先定下来。',
  },
  {
    phase: '第三步',
    scope: '设置页(唯一整页可换新组件的视图)',
    why: '卡片/字段行/只读信息是标准表单语义,Field 族一比一对应;风险集中在账号表与配置节脏标记上,单独收口。',
  },
  {
    phase: '第四步',
    scope: '接口日志与服务测试:工具条、详情页签、结果条',
    why: '数据密集页,先换外围控件,表格与滚动祖先保持手写,逐块验证 sticky 与键盘流。',
  },
  {
    phase: '第五步',
    scope: '工作台外壳:活动栏、状态栏、面板手风琴、树',
    why: '状态栏需要为"停站/运行"定语义色(默认 token 里没有),这一步同时决定是否保留编辑器式密度。',
  },
]
