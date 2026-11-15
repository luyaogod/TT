# CLI 分发与错误边界

## 命名空间

第三方命令走 `tt <namespace> ...`，namespace 规则见 [07-namespaces.md](07-namespaces.md)。

- 内置命令与旧别名始终优先：插件不得覆盖 `db`、`tdebug` 等名字。
- 冲突在安装/启用时失败，不能拖到首次调用才暴露。
- 根帮助只从 registry 的静态元数据列出插件 namespace，**绝不启动插件**。
- `tt <namespace> --help` 转发给插件，TT 不伪造插件帮助。

## 转发契约

| 项 | 规则 |
|---|---|
| 业务参数 | 原样转发，保持顺序与参数边界 |
| TT 全局旗标 | `--config`、`--env`、`--at`、`--json`、`--csv`、`--human`、`--format`、`--verbose` 由 TT 消费，不传给插件 |
| 标准流 | stdin/stdout/stderr 继承 |
| 工作目录 | 继承调用者的 cwd |
| 环境变量 | 只注入 `TT_HOST_API_URL`、`TT_PLUGIN_TOKEN`、`TT_PLUGIN_ID`、`TT_PLUGIN_DATA_DIR`、`TT_ENV` |
| 禁止注入 | 完整 config、其他插件的 token |

只注入最小环境是防止插件顺手读到无关环境的凭据。插件需要环境信息时走 `ProfilesService`，不要依赖父进程环境。

## 退出码边界

| 情形 | 返回值与输出 |
|---|---|
| 插件业务退出码 | **原样返回**；TT 不为它打印任何信封，也不重映射 |
| TT 基础设施失败 | TT 自己的错误信封 + 稳定 code；不得伪装成插件业务退出码 |

TT 基础设施失败包括 manifest 损坏、插件已禁用、spawn 失败、启动超时。插件返回 `2` 不会被解释为 TT 的用法错误；反过来，TT 的 spawn 失败也不会借用插件定义的业务码。

TT 全局退出码：

| 码 | 含义 |
|---|---|
| 0 | 成功 |
| 1 | 用法 |
| 2 | 数据源 |
| 3 | 缺失 |

各命令线自己的码表保持不变，不并入这张全局表。

## 错误信封

JSON 模式下错误走 TT 既有信封（stdout），含稳定 `code`、人读 `message`、`hint` 与 `env`/`source`；非 JSON 模式在 stderr 输出散文与提示。插件进程的退出码不产生信封，避免“TT 信封”和“插件自己的错误输出”叠成两个 JSON 对象。

## 尚未实现

插件 namespace 的注册/冲突检测、CLI 分发、supervisor 与 token 注入尚未实现。当前 `tt <namespace>` 只会走内置命令树，没有插件命令可转发。
