import { CATALOG, type Block, type CatalogSection } from './catalog'

/** 全部功能点,按台账顺序 */
export const ALL_BLOCKS: Block[] = CATALOG.flatMap((s) => s.blocks)

const BLOCK_INDEX = new Map(ALL_BLOCKS.map((b) => [b.id, b]))

export function blockOf(id: string): Block {
  const b = BLOCK_INDEX.get(id)
  if (!b) throw new Error(`台账里没有这个功能点:${id}`)
  return b
}

export function sectionOf(id: string): CatalogSection {
  const s = CATALOG.find((x) => x.id === id)
  if (!s) throw new Error(`台账里没有这一节:${id}`)
  return s
}

/** 某一节里"功能点 id → 台账条目"的查表,section 里逐块取用 */
export function blocksOf(sectionId: string): Record<string, Block> {
  const out: Record<string, Block> = {}
  for (const b of sectionOf(sectionId).blocks) out[b.id] = b
  return out
}

/** 风险统计:决策总览用 */
export function riskCounts(): { total: number; low: number; mid: number; high: number } {
  const c = { total: ALL_BLOCKS.length, low: 0, mid: 0, high: 0 }
  for (const b of ALL_BLOCKS) c[b.risk] += 1
  return c
}
