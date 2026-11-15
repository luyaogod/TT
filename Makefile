# TT 的构建/测试快捷入口。每个目标的语义、前提与代价写在它自己的 ## 注释里；
# 这里只做串接 —— 改命令先改这里。
#
# 前提：Git Bash 环境里的 GNU make（配方按 sh 语义写）。

SHELL := sh
.SHELLFLAGS := -c

.DEFAULT_GOAL := all
.PHONY: all help build test check-web web-build engine selftest doctor deep-tzc deep-tzs fns e2e skips version-check update-live-check package msi desktop-check desktop-dist proto-tools proto-gen proto-lint proto-breaking proto-check

all: build test

help: ## 列出全部目标
	@grep -E '^[a-z][a-z0-9-]*:' $(MAKEFILE_LIST) | grep '## ' | sed 's/:[^#]*## /  -- /'

build: ## go build -o tt.exe .（前端未构建也能过）
	go build -o tt.exe .

test: ## 默认档：全量 go test（约 1 分钟，含文档新鲜度检查）
	go test ./... -count=1

check-web: ## 前端三项检查：fgltokens / fgloutline / store
	cd web && npm run check:app

web-build: ## 前端构建（含 tsc 类型检查；check:app 不做类型检查）
	cd web && npm run build

# ---------- Host API（proto/tt/host/v1） ----------
#
# 公共契约只有 proto/tt/host/v1 一处，生成物入库（internal/gen）。规矩：
#   1) 改 .proto 必须重跑 proto-gen，proto-check 会拿 git diff 抓"改了没重生成"。
#   2) Buf 的缓存目录固定到仓库内：某些受限环境（企业 EDR / 只读 %LOCALAPPDATA%）
#      连 mkdir 都不许，buf 默认落到用户目录会当场失败。
#   3) 插件（protoc-gen-go / protoc-gen-connect-go）必须由 proto-tools 装，别下 exe 包：
#      下载来的可执行文件在某些终端防护策略下被禁止写盘，buf/protoc 会报
#      "Access is denied" 而看不出原因。
BUF_CACHE_DIR := $(CURDIR)/.tools/buf-cache
BUF_ENV := BUF_CACHE_DIR=$(BUF_CACHE_DIR)
# 兼容基线 —— **已发布契约的快照**，不是 git 分支：
#   - 用 git 分支没意义：本项目的 Host API 是新增的，main 上没有任何 .proto；
#   - 快照 = `buf build` 出的 FileDescriptorSet（JSON，可 diff 可评审），随契约一起入库；
#   - 只在**发布并有意扩充契约**（纯增量）后刷新基线：先发版，再
#     `buf build -o proto/baseline/v1.json --as-file-descriptor-set`，
#     这样下一个版本才拿"上一个已发布契约"作比较。
BUF_BREAKING_AGAINST ?= proto/baseline/v1.json

proto-tools: ## 装 Host API 的生成工具链（buf + 两个 protoc 插件，进 $(go env GOPATH)/bin）
	go install github.com/bufbuild/buf/cmd/buf@v1.75.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.7
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.19.1

proto-gen: ## 由 .proto 生成 internal/gen（protoc-gen-go + protoc-gen-connect-go）
	@command -v buf >/dev/null || { echo '缺 buf：先跑 make proto-tools（它把 buf 装进 $$(go env GOPATH)/bin，并确保该目录在 PATH 里）'; exit 1; }
	$(BUF_ENV) buf generate

proto-lint: ## buf lint（STANDARD，只有 buf.yaml 里写明的豁免）
	@command -v buf >/dev/null || { echo '缺 buf：先跑 make proto-tools'; exit 1; }
	$(BUF_ENV) buf lint

proto-breaking: ## buf breaking：对照已发布的契约快照（BUF_BREAKING_AGAINST 可覆盖）
	@command -v buf >/dev/null || { echo '缺 buf：先跑 make proto-tools'; exit 1; }
	$(BUF_ENV) buf breaking --against '$(BUF_BREAKING_AGAINST)'

proto-check: proto-lint proto-gen ## lint + 重生成 + 「生成物无 diff」闸门（改 .proto 忘了重生成会在这里红）
	@git diff --quiet -- internal/gen || { echo 'internal/gen 与 .proto 不一致：跑 make proto-gen 并把生成物一起提交'; git --no-pager diff --stat -- internal/gen; exit 1; }

engine: ## 只在真的改了 engine/ 时才跑！重编会孤儿化在跑的守护进程
	cd engine && ./build.sh

selftest: ## .tzc 的 31 项对抗用例（全部合成包，不需要真实语料）
	./tt.exe dev tzc selftest

doctor: ## .tzs 引擎环境自检（引擎 exe / 设计器 / 工作区 / 语料根）
	./tt.exe dev tzs doctor

deep-tzc: ## .tzc 全语料回归（9–11 分钟）。闸门：必须先设 TDEV_CORPUS 或 TTZS_CORPUS 指向副本
	@test -n "$$TDEV_CORPUS$$TTZS_CORPUS" || { echo '拒绝运行：先设 TDEV_CORPUS 或 TTZS_CORPUS 指向**副本**（如 TTZS_CORPUS=$$TEMP/ttws）——缺省语料根是真实客户目录'; exit 1; }
	TDEV_DEEP=1 go test ./internal/cli/dev -count=1 -timeout 30m

deep-tzs: ## .tzs 全语料回归（17 分钟）。闸门：必须先设 TTZS_CORPUS 指向副本
	@test -n "$$TTZS_CORPUS" || { echo '拒绝运行：先设 TTZS_CORPUS 指向**副本**（如 TTZS_CORPUS=$$TEMP/ttws）——缺省语料根是真实客户目录'; exit 1; }
	TTZS_DEEP=1 go test ./internal/dev/tzs -count=1 -timeout 30m

fns: ## 引擎函数面关卡（19 分钟）。闸门同 deep-tzs
	@test -n "$$TTZS_CORPUS" || { echo '拒绝运行：先设 TTZS_CORPUS 指向**副本**'; exit 1; }
	TTZS_FNS=1 go test ./internal/dev/tzs -run TestFnsGate -count=1 -timeout 30m -v

e2e: ## 真引擎 E2E。闸门：需要 TTZS_EXE 与 TTZS_WS
	@test -n "$$TTZS_EXE" -a -n "$$TTZS_WS" || { echo '拒绝运行：需要 TTZS_EXE 与 TTZS_WS'; exit 1; }
	TTZS_E2E=1 go test ./internal/dev/tzs -run TestE2E -count=1 -timeout 30m

skips: ## 回读：跑了多少、跳了多少（会真跑一遍全量）
	go test ./... -count=1 -v 2>&1 | grep -E '^--- (PASS|SKIP)' | sort | uniq -c

update-live-check: ## tt update 的真联网自检（要能访问 GitHub；代理用 TT_PROXY）
	@go build -ldflags "-X tt/internal/cli.Version=0.0.0" -o "$$TEMP/tt-updcheck.exe" .
	@$$TEMP/tt-updcheck.exe update check --format table --proxy "$$TT_PROXY" ; code=$$? ; rm -f "$$TEMP/tt-updcheck.exe" ; if [ "$$code" = "0" ] || [ "$$code" = "10" ] ; then echo "live-check OK (exit $$code)" ; else echo "live-check FAILED (exit $$code)" ; exit 1 ; fi


# 调 .bat 为什么长这样（两处都别退回去）：
#   1) 配方不是从交互式 shell 而是从 make 经 `sh -c` 下来，`cmd //c X` 在这种走法下
#      **什么也不做**：cmd 收到认不出的 `//c` 就进交互式，读到的 stdin 是空的于是
#      立刻退出 0 —— 目标报成功，产物一个没动。这比报错坏得多。
#   2) 换成 `cmd /c X` 也不行：从 Git Bash 直接敲时 MSYS 会把 `/c` 当路径转成 `C:\`。
#   `MSYS2_ARG_CONV_EXCL="*"` + 显式 `cmd.exe /c` 在两种走法下都对（已实测）。
version-check: ## 版本号只有 VERSION 一个出处：对照 web 的两份 package.json、lock 与 desktop/package.json
	@node -e 'const fs=require("fs");const want=fs.readFileSync("VERSION","utf8").trim();if(!/^[0-9]+\.[0-9]+\.[0-9]+$$/.test(want)){console.error("VERSION is not MAJOR.MINOR.PATCH: "+want);process.exit(1)}const lock=JSON.parse(fs.readFileSync("web/package-lock.json","utf8"));const rows=[["web/package.json",JSON.parse(fs.readFileSync("web/package.json","utf8")).version],["web/app/package.json",JSON.parse(fs.readFileSync("web/app/package.json","utf8")).version],["web/package-lock.json",lock.version],["web/package-lock.json packages[\"\"]",lock.packages[""].version],["desktop/package.json",JSON.parse(fs.readFileSync("desktop/package.json","utf8")).version]];const bad=rows.filter(r=>r[1]!==want);if(bad.length){console.error("version mismatch, VERSION = "+want+"\n  "+bad.map(r=>r[0]+" = "+r[1]).join("\n  "));process.exit(1)}console.log("version-check OK: "+want+" (VERSION is the single source)")'

package: version-check ## 便携包 → dist/tt-portable/ 与 dist/tt-portable.zip
	MSYS2_ARG_CONV_EXCL="*" cmd.exe /c "build_portable.bat"

msi: version-check ## MSI 安装包（需 WiX v3；复用便携包载荷）
	MSYS2_ARG_CONV_EXCL="*" cmd.exe /c "build_msi.bat"

desktop-check: ## 桌面壳的语法闸门 + 单元测试（纯 Node，不启 Electron、不需要载荷）
	cd desktop && npm test

desktop-dist: version-check ## 桌面版产物（Electron；复用便携包载荷，需能下 electron 运行时）
	MSYS2_ARG_CONV_EXCL="*" cmd.exe /c "build_desktop.bat"
