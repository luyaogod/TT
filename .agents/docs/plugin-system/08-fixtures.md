# 测试夹具设计

本文件只描述夹具布局与验证点；实现属于后续阶段。

## 布局

```
testdata/plugins/
├─ reference/
│  ├─ manifest.json
│  ├─ cli/
│  ├─ service/
│  ├─ frontend/
│  └─ skills/
└─ variants/
   ├─ no-contribution/
   ├─ invalid-manifest/
   ├─ namespace-collision/
   ├─ frontend-without-service/
   ├─ service-only/
   ├─ crash-on-start/
   ├─ refuses-to-exit/
   ├─ slow-start/
   └─ skills-conflict/
```

`reference` 同时声明 `cli`、`service`、`frontend`、`skills` 四类贡献，作为“完整插件”的基线。各变体只改与验证点相关的最小部分，避免维护十份互不相干的完整副本。

## 变体与验证点

| 变体 | 构成 | 验证点 |
|---|---|---|
| `reference` | 四类贡献齐全 | 安装、启用、CLI 转发、Host API、四类注册均走通 |
| `no-contribution` | 无任何贡献 | 合法安装/启用，不产生空注册，也不因“没东西可跑”而崩溃 |
| `invalid-manifest` | manifest 损坏或缺必填字段 | 安装拒绝，registry 不出现半条记录 |
| `namespace-collision` | namespace 命中内置名、旧别名或 `plugin` | 安装/启用拒绝，内置命令不被遮住 |
| `frontend-without-service` | 只有 frontend | 页面隔离成立；不要求 service 就绪 |
| `service-only` | 只有 service | ready 行与 `GET /healthz` 生效；崩溃/停止路径可测 |
| `crash-on-start` | 启动即退出 | 有界退避，不热循环，进入 `degraded` |
| `refuses-to-exit` | 忽略停止信号 | 实例身份校验 + Windows 进程树清理，不留下孤儿进程 |
| `slow-start` | 超过启动超时才 ready | 启用失败可回滚，不出现半激活状态 |
| `skills-conflict` | 与内置或其他插件技能重名 | 启用时报告冲突，不做静默覆盖 |

## 构建规则

- 夹具由 Go helper 确定性生成，不以手工文件作为唯一来源。
- 不访问网络，不连接真实客户环境。
- 生成结果可重复：同一输入得到同一目录树与同一 manifest。
- 测试只验证编排、注册、错误码与清理，不验证真实 ERP 业务。

## 尚未实现

`testdata/plugins/` 及生成 helper 尚未实现；上表是后续阶段的验收清单。
