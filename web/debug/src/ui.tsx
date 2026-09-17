// UI 基元已移到 web/shared/ —— 两套页面共用,只此一份。
//   shared/ui.tsx        不依赖 Radix 的那部分(字典页只引它)
//   shared/ui-radix.tsx  需要 Radix 的那部分(下拉/弹层/日历/确认弹窗/手风琴)
// 这里再导出两者,让本目录下的既有调用点不用改 import。
export * from '../../shared/ui'
export * from '../../shared/ui-radix'
