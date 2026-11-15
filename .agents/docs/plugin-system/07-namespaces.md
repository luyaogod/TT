# 命名空间

## 顶层内置命令

| 命令 | 说明 |
|---|---|
| `cache` | 缓存管理 |
| `completion` | shell 补全 |
| `config` | 配置读写 |
| `db` | 数据库连通性与只读 SQL |
| `debug` | 作业调试 |
| `dev` | 设计器包 |
| `dict` | 数据字典 |
| `env` | 环境管理 |
| `file` | 服务器文件只读浏览 |
| `help` | 帮助 |
| `install` | 安装 |
| `interface` | 画面/字段/标签 |
| `outline` | 4GL 大纲 |
| `serve` | 本地 Web 服务 |
| `session` | 调试会话 |
| `skills` | 技能分发 |
| `update` | TT 自更新 |
| `version` | 版本 |

## 旧别名

旧二进制名在解析前重写，因此与内置命令同等优先：

| 旧名 | 重写为 |
|---|---|
| `tdebug` | `debug` |
| `tdev` | `dev` |
| `tdict` | `dict` |

## 保留名

以下全部保留给 TT，插件 namespace 不得使用：

- `plugin`：新的插件管理组。
- 上表所有顶层内置命令。
- 上表所有旧别名。

插件 namespace 必须由小写字母、数字、连字符组成；当前校验要求以小写字母开头、最长 32 字符。它不得与保留名冲突，冲突在安装/启用时失败，不允许“先装上、调用时再遮住内置命令”。

## 分发规则

- 内置命令和旧别名始终赢；插件不能覆盖或别名到保留名。
- 根帮助只从 registry 的静态元数据列出插件 namespace，不启动插件。
- `tt <namespace> --help` 转发给插件；根帮助不推断插件的子命令。
- 插件命令的业务参数原样转发，全局 TT 旗标由 TT 消费（见 [05-cli-and-errors.md](05-cli-and-errors.md)）。

## 尚未实现

`plugin` 管理组、插件 namespace 注册、冲突检测与 CLI 分发尚未实现。当前根命令树只有内置命令，没有第三方 namespace。
