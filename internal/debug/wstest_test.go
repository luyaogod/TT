package debug

import (
	"strings"
	"testing"

	"tt/internal/dbconfig"
	"tt/internal/host"
)

func testWSCfg(zone, service string) *Config {
	c := &Config{SSH: host.SSHConfig{Host: "172.16.1.109"}, Zone: zone}
	if service != "" {
		c.DB = &dbconfig.Connection{Type: "oracle", Host: "172.16.1.109", Service: service}
	}
	return c
}

func TestWSZoneName(t *testing.T) {
	cases := []struct {
		cfg    *Config
		want   string
		source string
	}{
		{testWSCfg("36", "t35prd"), "t35prd", "db-service"},
		{testWSCfg("35", "t35tst"), "t35tst", "db-service"},
		{testWSCfg("1", ""), "topprd", "zone-tns"},
		{testWSCfg("2", ""), "toptst", "zone-tns"},
		{testWSCfg("t", ""), "topprd", "zone"},
		{testWSCfg("topprd", ""), "topprd", "zone"},
		{testWSCfg("36", ""), "", ""},
	}
	for _, c := range cases {
		got, src := wsZoneName(c.cfg)
		if got != c.want || src != c.source {
			t.Errorf("zone=%q service=%v → (%q,%q), want (%q,%q)",
				c.cfg.Zone, c.cfg.DB, got, src, c.want, c.source)
		}
	}
}

func TestWSTestURL(t *testing.T) {
	cases := []struct {
		cfg    *Config
		method string
		want   string
	}{
		{testWSCfg("36", "t35prd"), "3", "http://172.16.1.109/wt35prd/ws/r/awsp920"},
		{testWSCfg("36", "t35prd"), "1", "http://172.16.1.109/wt35prd/ws/r/awsp900"},
		{testWSCfg("36", "t35prd"), "4", "http://172.16.1.109/wt35prd/ws/r/awsp940"},
		{testWSCfg("36", "t35prd"), "5", "http://172.16.1.109/wt35prd/ws/r/awsp930"},
		{testWSCfg("1", ""), "3", "http://172.16.1.109/wstopprd/ws/r/awsp920"},
		{testWSCfg("t", ""), "3", "http://172.16.1.109/wstopprd/ws/r/awsp920"},
		{testWSCfg("36", ""), "3", ""},
		{testWSCfg("36", "t35prd"), "2", ""},
	}
	for _, c := range cases {
		if got := wsTestURL(c.cfg, c.method); got != c.want {
			t.Errorf("zone=%q method=%s: got %q, want %q", c.cfg.Zone, c.method, got, c.want)
		}
	}
}

func TestWSTestMeta(t *testing.T) {
	m := wsTestMetaFor(testWSCfg("36", "t35prd"), "109prd")
	if m.Env != "109prd" || m.Host != "172.16.1.109" || m.ZoneName != "t35prd" {
		t.Fatalf("meta 基本字段错: %#v", m)
	}
	if len(m.Methods) != 4 {
		t.Fatalf("方法数 = %d, want 4", len(m.Methods))
	}
	if m.Methods[1].ID != "3" || m.Methods[1].DefaultURL != "http://172.16.1.109/wt35prd/ws/r/awsp920" {
		t.Errorf("Restful 默认地址错: %#v", m.Methods[1])
	}
	if m.Methods[0].SOAP != true || m.Methods[1].SOAP != false {
		t.Error("SOAP 标记错")
	}
	if m.Limits.RequestBytes != wsTestMaxRequest || m.Limits.ResponseBytes != wsTestMaxResponse {
		t.Error("limits 错")
	}
}

func TestValidateWSTestURL(t *testing.T) {
	ok := []string{
		"http://172.16.1.109/wt35prd/ws/r/awsp920",
		"https://example.com/a?b=1",
	}
	for _, u := range ok {
		if err := validateWSTestURL(u); err != nil {
			t.Errorf("%q 应通过，得到 %v", u, err)
		}
	}
	bad := []string{
		"",
		"ftp://h/a",
		"http:///a",
		"http://user:pass@h/a",
		"http://h/a\nX",
		"http://" + strings.Repeat("a", 2100),
	}
	for _, u := range bad {
		if err := validateWSTestURL(u); err == nil {
			t.Errorf("%q 应被拒", u)
		}
	}
}

func TestResolveWSTest(t *testing.T) {
	cfg := testWSCfg("36", "t35prd")
	// 默认 URL
	m, u, timeout, err := resolveWSTest(cfg, WSTestRequest{Method: "3", Request: "{}"})
	if err != nil || m.ID != "3" || u != "http://172.16.1.109/wt35prd/ws/r/awsp920" || timeout != 60 {
		t.Fatalf("默认解析错: m=%#v u=%q t=%d err=%v", m, u, timeout, err)
	}
	// ?wsdl 剥离
	_, u, _, err = resolveWSTest(cfg, WSTestRequest{Method: "1", URL: "http://h/a?WSDL", Request: ""})
	if err != nil || u != "http://h/a" {
		t.Errorf("?wsdl 未剥离: u=%q err=%v", u, err)
	}
	// 非法方法 / 超限 / 无法生成默认地址 / 超时钳位
	if _, _, _, err := resolveWSTest(cfg, WSTestRequest{Method: "2"}); err == nil {
		t.Error("method=2 应被拒")
	}
	if _, _, _, err := resolveWSTest(cfg, WSTestRequest{Method: "3", Request: strings.Repeat("x", wsTestMaxRequest+1)}); err == nil {
		t.Error("超长请求应被拒")
	}
	if _, _, _, err := resolveWSTest(testWSCfg("36", ""), WSTestRequest{Method: "3"}); err == nil {
		t.Error("缺 zone 且无 URL 应被拒")
	}
	_, _, timeout, _ = resolveWSTest(cfg, WSTestRequest{Method: "3", TimeoutSec: 999})
	if timeout != wsTestMaxTimeout {
		t.Errorf("超时上限 = %d, want %d", timeout, wsTestMaxTimeout)
	}
	if clampWSTestTimeout(0) != 60 || clampWSTestTimeout(-1) != 60 {
		t.Error("timeout 0/-1 应回默认 60")
	}
}

func TestStripWSDLSuffix(t *testing.T) {
	cases := map[string]string{
		"http://h/a?wsdl": "http://h/a",
		"http://h/a?WSDL": "http://h/a",
		"http://h/a":      "http://h/a",
		"http://h/a?x=1":  "http://h/a?x=1",
	}
	for in, want := range cases {
		if got := stripWSDLSuffix(in); got != want {
			t.Errorf("stripWSDLSuffix(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestSOAPEnvelopeAndExtract(t *testing.T) {
	got := soapEnvelope(`{"a":"x&y","b":"<t>"}`)
	for _, w := range []string{
		`xmlns:tip="http://www.digiwin.com.cn/tiptop/TIPTOPServiceGateWay"`,
		"<tip:invokeSrv>", "<request>", "</request>", "</soapenv:Envelope>",
		`{"a":"x&amp;y","b":"&lt;t&gt;"}`,
	} {
		if !strings.Contains(got, w) {
			t.Errorf("信封缺少 %q\n%s", w, got)
		}
	}
	if strings.Contains(got, "&amp;lt;") || strings.Contains(got, "&amp;amp;") {
		t.Errorf("出现二次转义:\n%s", got)
	}

	resp := `&lt;response&gt;&lt;ok&gt;1&lt;/ok&gt;&lt;/response&gt;`
	if got := extractSOAPResponse(resp); got != `<response><ok>1</ok></response>` {
		t.Errorf("extract = %q", got)
	}
	raw := "not a soap response"
	if got := extractSOAPResponse(raw); got != raw {
		t.Errorf("截不到时应回原文，得到 %q", got)
	}
}

func TestParseCurlOutput(t *testing.T) {
	// 200 + JSON body
	code, status, body, dur, terr := parseCurlOutput("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"ok\":true}\n---TTMETA---200|0.123|11")
	if terr != "" || code != 200 || status != "OK" || body != `{"ok":true}` || dur < 0.12 {
		t.Errorf("200 解析错: code=%d status=%q body=%q dur=%v terr=%q", code, status, body, dur, terr)
	}
	// 400 + reason phrase（Proxy error 来自 Apache，不是 T100 代码）
	code, status, _, _, terr = parseCurlOutput("HTTP/1.1 400 Proxy error\r\n\r\n\n---TTMETA---400|0.05|0")
	if terr != "" || code != 400 || status != "Proxy error" {
		t.Errorf("400 解析错: code=%d status=%q terr=%q", code, status, terr)
	}
	// 没有状态行（连接失败）
	code, _, _, _, terr = parseCurlOutput("curl: (7) Failed to connect\n---TTMETA---000|0.010|0")
	if code != 0 || terr == "" {
		t.Errorf("连接失败应给 transportErr，得到 code=%d terr=%q", code, terr)
	}
	// 完全没有 META
	code, _, _, _, terr = parseCurlOutput("curl: (6) Could not resolve host")
	if code != 0 || terr == "" {
		t.Errorf("无 META 应给 transportErr，得到 code=%d terr=%q", code, terr)
	}
	// HTTP/2 无 reason
	code, status, _, _, terr = parseCurlOutput("HTTP/2 200\r\n\r\nbody\n---TTMETA---200|0.2|4")
	if terr != "" || code != 200 || status != "" {
		t.Errorf("HTTP/2 解析错: code=%d status=%q terr=%q", code, status, terr)
	}
	// body 里含空行，不能被二次切
	_, _, body, _, _ = parseCurlOutput("HTTP/1.1 200 OK\r\n\r\na\r\n\r\nb\n---TTMETA---200|0.1|6")
	if body != "a\r\n\r\nb" {
		t.Errorf("body 切错: %q", body)
	}
}

func TestWSCurlCmd(t *testing.T) {
	rest := wsCurlCmd(WSTestMethod{ID: "3", Endpoint: "awsp920"}, "http://h/a'b", 60)
	for _, w := range []string{"--data-binary @-", "-D -", "--max-redirs 0", "application/json", `'http://h/a'\''b'`} {
		if !strings.Contains(rest, w) {
			t.Errorf("REST 命令缺少 %q:\n%s", w, rest)
		}
	}
	soap := wsCurlCmd(WSTestMethod{ID: "1", Endpoint: "awsp900", SOAP: true}, "http://h/a", 60)
	for _, w := range []string{"text/xml", `SOAPAction: ""`, "Jakarta Commons-HttpClient"} {
		if !strings.Contains(soap, w) {
			t.Errorf("SOAP 命令缺少 %q:\n%s", w, soap)
		}
	}
}
