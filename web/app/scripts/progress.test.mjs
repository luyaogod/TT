// 进度条数值的校验（纯函数，不连后端、不需要浏览器）。
//
// 这三条是这次修的问题，钉住的是"以后别再退回去"：
//   1. 下载速度用**窗口差分**算，不是"累计字节 ÷ 任务总耗时"；
//   2. 服务器打包阶段不报速度，改报"已打包 / 预估归档大小"；
//   3. 总量未知时不假装 0%，而是把进度交给"不确定态"。
import assert from 'node:assert/strict'

import { pushSample, rateOf, fmtRate, fmtBytes, mirrorPct, mirrorCells } from '../src/progress'

// 1) 采样窗口：老点被丢掉，至少保留两个点
{
  let list = []
  for (let i = 0; i <= 10; i++) list = pushSample(list, i * 1000, i * 1000, 4000)
  assert.ok(list.length <= 6, `窗口应只留最近的点，实得 ${list.length}`)
  assert.ok(list.length >= 2)
  const only = pushSample([], 500, 1000)
  assert.equal(only.length, 1, '只有一个点时也算保留')
  console.log('OK 1) 采样窗口只留最近的几个点')
}

// 2) 速度：窗口内差分
{
  // 2 秒内从 1 MB 涨到 3 MB → 1 MB/s
  const list = [{ t: 0, bytes: 1 << 20 }, { t: 2000, bytes: 3 << 20 }]
  assert.equal(rateOf(list), 1 << 20)
  assert.equal(rateOf([{ t: 0, bytes: 100 }]), 0, '不足两点 → 0')
  assert.equal(rateOf([{ t: 1000, bytes: 100 }, { t: 1000, bytes: 200 }]), 0, '时间没走 → 0')
  assert.equal(rateOf([{ t: 0, bytes: 300 }, { t: 1000, bytes: 300 }]), 0, '字节没涨 → 0')
  console.log('OK 2) 速度 = 窗口内字节差 / 时间差')
}

// 3) 速度怎么显示
{
  assert.equal(fmtRate(0), '')
  assert.equal(fmtRate(512 * 1024), '512 KB/s')
  assert.equal(fmtRate(2.5 * 1024 * 1024), '2.5 MB/s')
  assert.equal(fmtBytes(265.5 * 1024 * 1024), '265.5 MB')
  console.log('OK 3) KB/s 与 MB/s 的分界正确')
}

// 4) 打包阶段：分母是"预估归档大小"，且不出现速度
{
  const pack = { phase: 'pack', bytes: 40 << 20, total: 200 << 20, elapsed: '12s', files: 0, running: true }
  assert.equal(mirrorPct(pack), 20)
  const cells = mirrorCells(pack, 9 * 1024 * 1024) // 就算算得出速度，也不该出现
  assert.deepEqual(cells, ['已打包 40.0 MB', '预估 200.0 MB', '12s'])
  assert.ok(!cells.some((c) => /\/s/.test(c)), '打包阶段不能出现速度')
  console.log('OK 4) 打包阶段报"已打包 / 预估"，不报速度')
}

// 5) 下载阶段：报真实总量与速度
{
  const dl = { phase: 'download', bytes: 100 << 20, total: 400 << 20, elapsed: '20s', files: 0, running: true }
  assert.equal(mirrorPct(dl), 25)
  assert.deepEqual(mirrorCells(dl, 4 * 1024 * 1024), ['100.0 MB / 400.0 MB', '4.0 MB/s', '20s'])
  console.log('OK 5) 下载阶段报"已下载 / 总量"与速度')
}

// 6) 总量未知 / 完成 / 超额
{
  assert.equal(mirrorPct({ phase: 'pack', bytes: 1 << 20, total: 0, elapsed: '', files: 0, running: true }), undefined)
  assert.equal(mirrorPct({ phase: 'done', bytes: 1, total: 1, elapsed: '', files: 0, running: false }), 100)
  // 预估偏低时归档会超过预估：夹在 100，不出现 137%
  assert.equal(mirrorPct({ phase: 'pack', bytes: 300 << 20, total: 200 << 20, elapsed: '', files: 0, running: true }), 100)
  console.log('OK 6) 未知总量 / 完成 / 超额都落在预期区间')
}

// 7) 完成后的数值行：保留大小、不报速度
{
  const done = { phase: 'done', bytes: 265 << 20, total: 265 << 20, elapsed: '4m2s', files: 4321, running: false }
  assert.deepEqual(mirrorCells(done, 0), ['265.0 MB / 265.0 MB', '4m2s', '4321 个文件'])
  console.log('OK 7) 完成后仍显示总量与文件数')
}

console.log('\n全部通过：进度与速度的 7 组判定')
