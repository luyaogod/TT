# 信任模型

## 进程信任：安装即信任

插件是用户显式安装的本地可执行代码。TT 不对它做沙箱。同一 OS 用户下的其他进程本来就能读该用户的文件，插件进程也一样；不要声称 TT 能阻止这件事。

- capability 与环境授权限制的是 **Host API 愿意替插件做什么**，不是 OS 权限。
- 插件能直接访问的文件、网络、进程，TT 不拦截。
- manifest 和 token 约束调用路径，不是进程沙箱。

因此对插件的信任在安装时建立，而不是在运行时由 TT 证明。安装前看清来源，是这套模型里唯一真正的准入判断。

## 页面隔离

第三方前端页面只做 **page isolation**，不是进程隔离：

- 页面跑在 `iframe sandbox="allow-scripts"` 中，**不带** `allow-same-origin`。
- 页面不能触碰宿主 DOM、Zustand store、desktop preload。
- 页面不接收管理 API token。
- 页面处于不透明源，`Origin` 为字面量 `null`；Host API 只在沙箱 iframe 场景接受这个值，其他来源返回 `BAD_ORIGIN`。

页面隔离防止插件页面借宿主页面的身份调用管理面。它不防止插件进程读写用户文件，也不把插件变成不可信代码的安全容器。

## Electron 约束

- 第三方 iframe 不继承 `ttDesktopShell`。
- 桌面壳保持 `nodeIntegration:false`、`contextIsolation:true`、`sandbox:true`。
- 第三方页面拿不到 Node/Electron 能力，也拿不到 preload 暴露的桌面接口。

## 密钥

- 口令与私钥只在 Core 内解析，只用于建立连接，绝不跨 API 返回。
- `ProfilesService.List` 只返回脱敏元数据：环境、zone、库类型、库地址、**账号名清单**、SSH host/port/user、是否经 SSH、内置只读 SQL 开关。不返回口令、私钥、完整 config JSON。
- 审计只记 action、caller、env、kind、ok/code、duration，不记 SQL 文本、SSH 命令文本或密钥。
- 插件只拿到自己的实例 token；不注入完整 config，也不注入其他插件的 token。

## 非目标

- OS 级沙箱、同用户文件隔离：TT 不提供，文档也不应暗示已有。
- 对插件 SQL/命令的业务审查：TT 不做，只做连接边界上的只读检查与权限判断。

## 尚未实现

前端插件页面、页面与宿主之间的 bridge、插件实例 token 尚未实现；上述 iframe 与 Electron 规则是目标契约。当前没有第三方页面可加载。
