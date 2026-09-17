// 主题机制已移到 web/shared/theme.ts —— 两套页面共用同一份,用户的选择在页面之间共享。
// 这里保留一层再导出,是为了让本目录下的既有调用点(store.ts / main.tsx 等)不用改路径;
// 清掉这层 shim 直接引 ../../shared/theme 是后续清理,不影响行为。
export * from '../../shared/theme'
