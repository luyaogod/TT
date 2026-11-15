package debug

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/internal/host"
)

// wsTestServerWithEnv 起一个不拨真机的调试服务，并给它一个具名环境（可带 DB service）。
func wsTestServerWithEnv(t *testing.T, zone, service string) (*httptest.Server, *Server) {
	t.Helper()
	hs, s := actionTestDebugServer(t)
	e := host.NamedSsh{
		Name:      "h1",
		SSHConfig: host.SSHConfig{Host: "127.0.0.1", Port: 22, User: "u"},
		Zone:      zone,
	}
	if service != "" {
		e.DB = &dbconfig.Connection{Type: "oracle", Host: "127.0.0.1", Port: 1521, Service: service}
	}
	s.cfg.SSHs = append(s.cfg.SSHs, e)
	s.cfg.fillDefaults()
	s.cfg.SSH, s.cfg.Zone, s.cfg.DB = e.SSHConfig, zone, e.DB
	return hs, s
}

func TestWSServicesHandler(t *testing.T) {
	hs, _ := wsTestServerWithEnv(t, "36", "t35prd")

	// 缺 env → 400 USAGE；不拨号
	resp := doReq(t, "GET", hs.URL+"/api/wslogs/services", "", "")
	if resp.StatusCode != 400 {
		t.Fatalf("缺 env 该 400，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	if msg := respError(t, resp); !strings.Contains(msg, "需要 env") {
		t.Errorf("文案该说清缺 env，得 %q", msg)
	}

	// 点名不存在的环境 → 404 NOT_FOUND（新端点不静默回落）
	resp = doReq(t, "GET", hs.URL+"/api/wslogs/services?env=nope", "", "")
	if resp.StatusCode != 404 {
		t.Fatalf("未知环境该 404，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}

	// 存在的环境 → 拨号（被注入的 dial 拒绝）→ 500
	resp = doReq(t, "GET", hs.URL+"/api/wslogs/services?env=h1", "", "")
	if resp.StatusCode != 500 {
		t.Fatalf("拨号失败该 500，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	if msg := respError(t, resp); !strings.Contains(msg, errTestNoDial.Error()) {
		t.Errorf("该是注入的 dial 报错，得 %q", msg)
	}
}

func TestWSMetaHandler(t *testing.T) {
	hs, _ := wsTestServerWithEnv(t, "36", "t35prd")
	resp := doReq(t, "GET", hs.URL+"/api/wslogs/meta?env=h1", "", "")
	if resp.StatusCode != 200 {
		t.Fatalf("meta 该 200，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	var out struct {
		OK       bool   `json:"ok"`
		Env      string `json:"env"`
		Host     string `json:"host"`
		ZoneName string `json:"zoneName"`
		Methods  []struct {
			ID         string `json:"id"`
			Label      string `json:"label"`
			SOAP       bool   `json:"soap"`
			DefaultURL string `json:"defaultUrl"`
		} `json:"methods"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if !out.OK || out.Env != "h1" || out.Host != "127.0.0.1" || out.ZoneName != "t35prd" {
		t.Fatalf("meta 基本字段错: %+v", out)
	}
	if len(out.Methods) != 4 {
		t.Fatalf("方法数 = %d, want 4", len(out.Methods))
	}
	if out.Methods[1].ID != "3" || out.Methods[1].Label != "Restful" ||
		out.Methods[1].DefaultURL != "http://127.0.0.1/wt35prd/ws/r/awsp920" {
		t.Errorf("Restful 项错: %+v", out.Methods[1])
	}
}

func TestWSMetaWithoutDB(t *testing.T) {
	hs, _ := wsTestServerWithEnv(t, "36", "")
	resp := doReq(t, "GET", hs.URL+"/api/wslogs/meta?env=h1", "", "")
	if resp.StatusCode != 200 {
		t.Fatalf("meta 该 200，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	var out struct {
		ZoneName string `json:"zoneName"`
		Methods  []struct {
			DefaultURL string `json:"defaultUrl"`
		} `json:"methods"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.ZoneName != "" {
		t.Errorf("缺 DB service 且 zone 是数字时 zoneName 该空，得 %q", out.ZoneName)
	}
	for _, m := range out.Methods {
		if m.DefaultURL != "" {
			t.Errorf("zoneName 为空时不该硬拼默认地址，得 %q", m.DefaultURL)
		}
	}
}

func TestWSTestValidatesBeforeDialing(t *testing.T) {
	hs, _ := wsTestServerWithEnv(t, "36", "t35prd")
	cases := []struct {
		name   string
		body   string
		status int
		want   string
	}{
		{"缺 env", `{"method":"3"}`, 400, "需要 env"},
		{"未知环境", `{"env":"nope","method":"3"}`, 404, "不存在"},
		{"非法方法", `{"env":"h1","method":"2"}`, 400, "接口方式非法"},
		{"非法 URL", `{"env":"h1","method":"3","url":"ftp://h/a"}`, 400, "http"},
		{"请求超限", `{"env":"h1","method":"3","request":"` + strings.Repeat("x", wsTestMaxRequest+1) + `"}`, 400, "上限"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := doReq(t, "POST", hs.URL+"/api/wslogs/test", "", c.body)
			if resp.StatusCode != c.status {
				t.Fatalf("该退 %d，得 %d（%s）", c.status, resp.StatusCode, respError(t, resp))
			}
			if msg := respError(t, resp); !strings.Contains(msg, c.want) {
				t.Errorf("文案该含 %q，得 %q", c.want, msg)
			}
		})
	}

	// 通过校验后才拨号；拨号失败 → 500
	resp := doReq(t, "POST", hs.URL+"/api/wslogs/test", "", `{"env":"h1","method":"3","request":"{}"}`)
	if resp.StatusCode != 500 {
		t.Fatalf("拨号失败该 500，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	if msg := respError(t, resp); !strings.Contains(msg, errTestNoDial.Error()) {
		t.Errorf("该是注入的 dial 报错，得 %q", msg)
	}
}

func TestWSRoutesAreRegistered(t *testing.T) {
	hs, _ := wsTestServerWithEnv(t, "36", "t35prd")
	for _, c := range []struct{ method, path string }{
		{"GET", "/api/wslogs/services?env=h1"},
		{"GET", "/api/wslogs/meta?env=h1"},
		{"POST", "/api/wslogs/test"},
	} {
		resp := doReq(t, c.method, hs.URL+c.path, "", "")
		if resp.StatusCode == http.StatusNotFound {
			t.Errorf("%s %s 未注册（404）", c.method, c.path)
		}
	}
}
