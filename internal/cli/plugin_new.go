package cli

// tt plugin new：生成一个能立刻构建/打包/安装的插件骨架。
//
// 骨架刻意最小：一个 CLI 子命令 + 一个常驻服务（healthz/ready）+ 一个页面，
// 全部走 TT 的公开 SDK，不复制 TT 源码。

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/plugin"
)

func newPluginNewCmd() *cobra.Command {
	var id, name, ns, route, replacePath string
	var noCLI, noService, noUI bool
	cmd := &cobra.Command{
		Use:   "new <目录>",
		Short: "生成插件骨架（可构建、可打包）",
		Long: `在空目录里生成一个插件骨架：

  plugin.json          清单（CLI/服务/页面按需取舍）
  go.mod               依赖 github.com/luyaogod/TT 的公开 SDK
  cmd/<ns>/main.go     CLI + 常驻服务
  client/client.js     页面（page 槽位）
  scripts/build.ps1    构建到 bin/windows-amd64/
  README.md

生成后用 tt plugin pack 打包、tt plugin install 安装。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			if err := plugin.ValidateID(id); err != nil {
				return err
			}
			if ns == "" {
				ns = strings.TrimPrefix(id, "com.")
				ns = strings.ReplaceAll(ns, ".", "-")
			}
			if err := plugin.ValidateNamespace(ns); err != nil {
				return err
			}
			if noUI && noCLI && noService {
				return fmt.Errorf("CLI/服务/页面至少要保留一类")
			}
			if route == "" {
				route = ns
			}
			if err := plugin.ValidateRoute(route); err != nil {
				return err
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			if entries, _ := os.ReadDir(dir); len(entries) > 0 {
				return fmt.Errorf("目录 %s 不是空的", dir)
			}
			if err := writePluginScaffold(dir, scaffold{
				ID: id, Name: derefName(name, ns), Namespace: ns, Route: route,
				CLI: !noCLI, Service: !noService, UI: !noUI, Replace: replacePath,
			}); err != nil {
				return err
			}
			b, _ := json.Marshal(map[string]any{"ok": true, "path": dir, "id": id, "namespace": ns, "route": route})
			fmt.Println(string(b))
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "插件 id（必填，反向域名，如 com.example.hello）")
	cmd.Flags().StringVar(&name, "name", "", "展示名（缺省用命名空间）")
	cmd.Flags().StringVar(&ns, "namespace", "", "CLI 命名空间（缺省由 id 推导）")
	cmd.Flags().StringVar(&route, "route", "", "页面路由（缺省同命名空间）")
	cmd.Flags().BoolVar(&noCLI, "no-cli", false, "不生成 CLI")
	cmd.Flags().BoolVar(&noService, "no-service", false, "不生成常驻服务")
	cmd.Flags().BoolVar(&noUI, "no-ui", false, "不生成页面")
	cmd.Flags().StringVar(&replacePath, "replace", "", "在 go.mod 里 replace 到本地 TT 源码目录（开发态）")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func derefName(name, fallback string) string {
	if strings.TrimSpace(name) == "" {
		return fallback
	}
	return name
}

type scaffold struct {
	ID, Name, Namespace, Route string
	CLI, Service, UI           bool
	Replace                    string
}

func writePluginScaffold(dir string, s scaffold) error {
	files := map[string]string{
		"plugin.json": scaffoldManifest(s),
		"README.md":   scaffoldReadme(s),
		".gitignore":  "bin/\ndist/\n",
	}
	if s.UI {
		files["client.js"] = scaffoldClient(s)
	}
	if s.CLI || s.Service {
		files[filepath.ToSlash(filepath.Join("cmd", s.Namespace, "main.go"))] = scaffoldMain(s)
		files["scripts/build.ps1"] = scaffoldBuild(s)
		files["go.mod"] = scaffoldGoMod(s)
	}
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func scaffoldManifest(s scaffold) string {
	var contribs []string
	if s.CLI {
		contribs = append(contribs, fmt.Sprintf(`    "cli": {
      "namespace": %q,
      "entry": "bin/windows-amd64/%s.exe",
      "summary": %q
    }`, s.Namespace, s.Namespace, s.Name))
	}
	if s.Service {
		contribs = append(contribs, fmt.Sprintf(`    "service": {
      "entry": "bin/windows-amd64/%s.exe",
      "args": ["serve"],
      "healthPath": "/healthz"
    }`, s.Namespace))
	}
	if s.UI {
		contribs = append(contribs, fmt.Sprintf(`    "ui": [
      { "slot": "page", "route": %q, "title": %q, "entry": "client.js" }
    ]`, s.Route, s.Name))
	}
	return fmt.Sprintf(`{
  "schemaVersion": 2,
  "id": %q,
  "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "tt": ">=0.7.0 <1.0.0",
  "name": %q,
  "description": "",
  "contributions": {
%s
  }
}
`, s.ID, s.Name, strings.Join(contribs, ",\n"))
}

func scaffoldGoMod(s scaffold) string {
	if s.Replace != "" {
		p := filepath.ToSlash(s.Replace)
		return fmt.Sprintf(`module example.com/%s

go 1.26.5

require github.com/luyaogod/TT v0.0.0

replace github.com/luyaogod/TT => %s
`, s.Namespace, p)
	}
	return fmt.Sprintf(`module example.com/%s

go 1.26.5

require github.com/luyaogod/TT v0.7.0
`, s.Namespace)
}

func scaffoldMain(s scaffold) string {
	return fmt.Sprintf(`// %s 的插件入口：CLI 子命令 + 常驻服务。
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/luyaogod/TT/sdk"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "hello":
			fmt.Println("hello from %s")
			return
		case "serve":
			serve()
			return
		}
	}
	fmt.Println("用法: %s hello | %s serve")
}

type service struct{ srv *http.Server }

func (s *service) Start(_ sdk.Context) error {
	addr := os.Getenv("TT_PLUGIN_LISTEN")
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`+"`"+`{"ok":true}`+"`"+`))
	})
	s.srv = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	fmt.Printf("{\"ready\":true,\"addr\":%%q}\n", ln.Addr().String())
	go func() { _ = s.srv.Serve(ln) }()
	return nil
}

func (s *service) Stop(_ sdk.Context) error {
	if s.srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func serve() { sdk.Run(&service{}) }
`, s.Name, s.Name, s.Namespace, s.Namespace)
}

func scaffoldClient(s scaffold) string {
	return fmt.Sprintf(`// %s 的同页页面：用宿主注入的 ctx.React，不依赖构建。
export function register(ctx) {
  const h = ctx.React.createElement
  function Page() {
    return h('div', { className: 'p-4 text-sm text-foreground' },
      h('div', { className: 'text-base font-medium' }, %q),
      h('div', { className: 'mt-1 text-muted-foreground' }, '骨架已跑通，去 cmd/%s/main.go 和这里继续写。')
    )
  }
  ctx.slot('page', { route: %q, title: %q, component: Page })
}
`, s.Name, s.Name, s.Namespace, s.Route, s.Name)
}

func scaffoldBuild(s scaffold) string {
	return fmt.Sprintf(`$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
  New-Item -ItemType Directory -Force "bin\windows-amd64" | Out-Null
  go build -trimpath -o "bin\windows-amd64\%s.exe" ./cmd/%s
  if ($LASTEXITCODE -ne 0) { throw "go build 失败" }
} finally { Pop-Location }
`, s.Namespace, s.Namespace)
}

func scaffoldReadme(s scaffold) string {
	cli := fmt.Sprintf("tt %s hello", s.Namespace)
	if !s.CLI {
		cli = "（本插件未声明 CLI）"
	}
	return fmt.Sprintf(`# %s

由 tt plugin new 生成。

## 开发

~~~powershell
go mod tidy
pwsh -File scripts/build.ps1
tt plugin pack . --out dist/%s-0.1.0.zip
tt plugin install dist/%s-0.1.0.zip
tt plugin enable %s
~~~

~~~powershell
%s
~~~

页面：TT 左侧「插件」里点 **%s**（路由 #plugin/%s）。

## 结构

- cmd/%s/main.go：CLI 与常驻服务（healthz/ready）
- client.js：页面（page 槽位，手写 ESM）
- plugin.json：清单（schemaVersion 2）
`, s.Name, s.ID, s.ID, s.ID, cli, s.Name, s.Route, s.Namespace)
}
