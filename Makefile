# TT 的构建/测试快捷入口。每个目标的语义、前提与代价写在它自己的 ## 注释里；
# 这里只做串接 —— 改命令先改这里。
#
# 前提：Git Bash 环境里的 GNU make（配方按 sh 语义写）。

SHELL := sh
.SHELLFLAGS := -c

.DEFAULT_GOAL := all
.PHONY: all help build test check-web web-build engine selftest doctor deep-tzc deep-tzs fns e2e skips drawio-lib package msi

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

drawio-lib: build ## drawio 形状库 → dist/ 下两个 mxlibrary .xml（File → Open Library from 加载）
	./tt.exe drawio lib

package: ## 便携包 → dist/tt-portable/ 与 dist/tt-portable.zip
	cmd //c build_portable.bat

msi: ## MSI 安装包（需 WiX v3；复用便携包载荷）
	cmd //c build_msi.bat
