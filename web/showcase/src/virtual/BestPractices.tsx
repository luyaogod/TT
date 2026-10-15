// 这台虚拟应用遵循的"最佳实践"清单:每一条都是对某个具体问题的取舍,不是风格口号。
// 它同时是这张屏的图例 —— 看预览时对着它能看出"这里为什么这么做"。
export const PRACTICES: { title: string; why: string }[] = [
  {
    title: '依赖底座统一到官方通道',
    why: '组件来自 registry 的 new-york-v4:统一 radix-ui 包 + 官方 cn 包。不混用逐个 @radix-ui/react-*,否则同一套 Radix 会装两份。',
  },
  {
    title: '默认 token 全量保留,只加三个语义色',
    why: 'shadcn 默认没有 success/warning/info。补这三个(--success / --warning / --info),状态栏、Badge、Alert 共用;其余一律默认取值,不改圆角、不改色板。',
  },
  {
    title: '密度用官方尺寸变体,不做全站覆盖',
    why: '按钮用 sm/xs/icon,表格行用 text-xs + h-8 —— 紧凑来自"选小号",不是把 h-9 硬压成 h-7。这样升级 registry 不会和你打架。',
  },
  {
    title: '危险确认 Promise 化,调用点形状不变',
    why: '三处 window.confirm 依赖同步布尔返回值。用 useConfirm() 收进 Provider,调用点只是多一个 await —— 行为等价,不必改写控制流。',
  },
  {
    title: '全局结果走 Sonner,按钮反馈留在按钮上',
    why: '保存成功/同步完成/已复制标识这类"动作完成"用 toast;「已复制」这种"就在你手边"的反馈留在按钮文案上,不抢焦点。',
  },
  {
    title: '表格:借官方单元格,滚动祖先自己拿',
    why: '官方 Table 自带 overflow-x-auto 容器,会让 sticky 表头失效。所以用官方 TableHead/TableRow/TableCell 组自己的 table 与滚动容器,表头背景不透明。',
  },
  {
    title: 'Monaco 与它的浮层保持手写',
    why: '编辑器主题、悬停卡片(Monaco content widget,不在 React 树里)、gutter 装饰都不能换组件;唯一的对齐要求是 editor.background = --card。',
  },
  {
    title: '交互硬约束原样保留',
    why: '账号表 Enter 提交、筛选框回车才生效、"仅失败"即时生效、Esc 关详情、协作模式下写操作一律禁用(服务端同样 403)。',
  },
]
