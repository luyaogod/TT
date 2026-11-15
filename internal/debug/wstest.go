package debug

// 集成服务检测：按 T100 awsq990 page_1 的口径，从 T100 服务器向接口发一次请求。
//
// 为什么在服务器上发（而不是本机）：
//   - 原生 awsq990 跑在服务器上，com.HTTPRequest 与网关（awsp920）在同一网络侧；
//   - 用户机到 T100 Web 端口的连通性不可假设，配置里也没有 WS 地址；
//   - 服务器侧执行所需的原语（SSHOutput/OutputStdin/SFTP）本仓已有。
//
// 请求体走 stdin（curl --data-binary @-），不落服务器临时文件；URL 用单引号转义。
// 与原生一致：仅 HTTP 200 判成功；非 200 的正文**保留**（原生丢弃，看不到 Proxy error 细节）。
// 不写 wsfa_t —— 日志由 T100 网关/服务程序在真实调用时自己写。

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tt/internal/host"
)

const (
	wsTestMaxRequest     = 256 * 1024
	wsTestMaxResponse    = 2 * 1024 * 1024
	wsTestDefaultTimeout = 60
	wsTestMaxTimeout     = 120
)

// WSTestMethod 一种接口方式（与 awsq990.42f 的 wsfc001 下拉一致：1/3/4/5）。
type WSTestMethod struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Endpoint string `json:"endpoint"`
	SOAP     bool   `json:"soap"`
}

// wsTestMethodList 顺序即界面下拉顺序；2（JAVA API）原生下拉里没有，不提供。
var wsTestMethodList = []WSTestMethod{
	{ID: "1", Label: "Web Service", Endpoint: "awsp900", SOAP: true},
	{ID: "3", Label: "Restful", Endpoint: "awsp920"},
	{ID: "4", Label: "OpenApi Restful", Endpoint: "awsp940"},
	{ID: "5", Label: "OpenApi Web Service", Endpoint: "awsp930", SOAP: true},
}

// wsTestMethod 按 id 取方法；未知 id 返回 false。
func wsTestMethod(id string) (WSTestMethod, bool) {
	for _, m := range wsTestMethodList {
		if m.ID == id {
			return m, true
		}
	}
	return WSTestMethod{}, false
}

// WSTestRequest /api/wslogs/test 的请求体。
type WSTestRequest struct {
	Env        string `json:"env"`
	Method     string `json:"method"`
	URL        string `json:"url"`
	Request    string `json:"request"`
	TimeoutSec int    `json:"timeoutSec"`
}

// WSTestResult 一次检测的结果（前端结果小表 + Response 区共用）。
type WSTestResult struct {
	At             string  `json:"at"` // yyyy-mm-dd hh:mm:ss.SSS
	HTTPCode       int     `json:"httpCode"`
	StatusText     string  `json:"statusText"`
	DurationSec    float64 `json:"durationSec"`
	Body           string  `json:"body"`
	TransportError string  `json:"transportError,omitempty"`
	Truncated      bool    `json:"truncated,omitempty"`
	RequestBytes   int     `json:"requestBytes"`
	ResponseBytes  int     `json:"responseBytes"`
}

// WSTestMethodMeta 前端下拉项（带默认地址）。
type WSTestMethodMeta struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Endpoint   string `json:"endpoint"`
	SOAP       bool   `json:"soap"`
	DefaultURL string `json:"defaultUrl"`
}

// WSTestLimits 与前端共享的上限（前端据此禁用/提示，后端仍会再校验）。
type WSTestLimits struct {
	RequestBytes  int `json:"requestBytes"`
	ResponseBytes int `json:"responseBytes"`
	TimeoutSec    int `json:"timeoutSec"`
}

// WSTestMeta GET /api/wslogs/meta 的应答。
type WSTestMeta struct {
	Env        string             `json:"env"`
	Host       string             `json:"host"`
	Zone       string             `json:"zone"`
	ZoneName   string             `json:"zoneName"`
	ZoneSource string             `json:"zoneSource"`
	Methods    []WSTestMethodMeta `json:"methods"`
	Limits     WSTestLimits       `json:"limits"`
}

// wsZoneName 推导 T100 的 ZONE 字符串（URL 里的区段）。
//
// 优先级（宁可空着让用户手填，也不猜一个会静默打到别的区的值）：
//  1. host.ZoneTNSName(zone)：旧式登录菜单码 1/2/3/4 → topprd/toptst/topstd/toppth；
//  2. cfg.DB.Svc()：Oracle 服务名 / 金仓库名（本环境 t35prd / t35tst）；
//  3. cfg.Zone：本身已经是 zone 字符串（如 topprd / t35prd）时；
//  4. 空。
//
// 注意：不要反过来依赖 ZoneTNSName 覆盖 35/36 —— 它已被收窄为只认 1/2/3/4，
// 对 35/36 返回空串（见 internal/host/dbprobe.go 的注释）。
func wsZoneName(cfg *Config) (name, source string) {
	if cfg == nil {
		return "", ""
	}
	if n := host.ZoneTNSName(cfg.Zone); n != "" {
		return n, "zone-tns"
	}
	// 本机出货区的菜单码是 t（ZONE=topprd）；ZoneTNSName 只收旧式 1/2/3/4。
	if strings.EqualFold(strings.TrimSpace(cfg.Zone), "t") {
		return "topprd", "zone"
	}
	if cfg.DB != nil {
		if s := strings.TrimSpace(cfg.DB.Svc()); s != "" {
			return s, "db-service"
		}
	}
	if z := strings.TrimSpace(cfg.Zone); z != "" && !allDigits(z) {
		return z, "zone"
	}
	return "", ""
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// wsTestURL 按接口方式与环境生成默认地址；缺 host 或 zone 时返回空串（不硬拼）。
//
// 口径与 awsq990.4gl:526-531 一致：
//
//	ZONE=topprd：http://<host>/ws<ZONE>/ws/r/<endpoint>
//	其余区      ：http://<host>/w<ZONE>/ws/r/<endpoint>
func wsTestURL(cfg *Config, method string) string {
	m, ok := wsTestMethod(method)
	if !ok {
		return ""
	}
	if cfg == nil {
		return ""
	}
	hostIP := strings.TrimSpace(cfg.SSH.Host)
	zone, _ := wsZoneName(cfg)
	if hostIP == "" || zone == "" {
		return ""
	}
	return "http://" + hostIP + wsURLPrefix(cfg.Zone) + zone + "/ws/r/" + m.Endpoint
}

// wsTestMetaFor 组装前端元信息。纯函数（不碰 SSH/DB），便于单测。
func wsTestMetaFor(cfg *Config, envName string) WSTestMeta {
	zone, source := wsZoneName(cfg)
	m := WSTestMeta{
		Env:        envName,
		Zone:       cfg.Zone,
		ZoneName:   zone,
		ZoneSource: source,
		Methods:    make([]WSTestMethodMeta, 0, len(wsTestMethodList)),
		Limits: WSTestLimits{
			RequestBytes:  wsTestMaxRequest,
			ResponseBytes: wsTestMaxResponse,
			TimeoutSec:    wsTestDefaultTimeout,
		},
	}
	if cfg != nil {
		m.Host = strings.TrimSpace(cfg.SSH.Host)
	}
	for _, it := range wsTestMethodList {
		m.Methods = append(m.Methods, WSTestMethodMeta{
			ID: it.ID, Label: it.Label, Endpoint: it.Endpoint, SOAP: it.SOAP,
			DefaultURL: wsTestURL(cfg, it.ID),
		})
	}
	return m
}

// validateWSTestURL URL 的形状校验（不判断可达性）。
func validateWSTestURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("集成方式(URL)不能为空")
	}
	if len(raw) > 2048 {
		return fmt.Errorf("URL 超过 2048 字符")
	}
	if strings.ContainsFunc(raw, func(r rune) bool { return r < 0x20 || r == 0x7f }) {
		return fmt.Errorf("URL 含控制字符")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("URL 解析失败: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL 必须以 http:// 或 https:// 开头")
	}
	if u.Host == "" {
		return fmt.Errorf("URL 缺少主机名")
	}
	if u.User != nil {
		return fmt.Errorf("URL 不允许携带用户名/口令")
	}
	return nil
}

// clampWSTestTimeout 把超时夹到 [1,120]，0 → 默认 60。
func clampWSTestTimeout(n int) int {
	if n <= 0 {
		return wsTestDefaultTimeout
	}
	if n > wsTestMaxTimeout {
		return wsTestMaxTimeout
	}
	return n
}

// resolveWSTest 校验请求并补齐默认 URL，返回 (方法, URL, 超时, error)。
// 全部在拨号之前完成 —— 注定被拒的请求不白拨一次 SSH。
func resolveWSTest(cfg *Config, req WSTestRequest) (WSTestMethod, string, int, error) {
	m, ok := wsTestMethod(strings.TrimSpace(req.Method))
	if !ok {
		return WSTestMethod{}, "", 0, fmt.Errorf("接口方式非法: %q（只支持 1/3/4/5）", req.Method)
	}
	if len(req.Request) > wsTestMaxRequest {
		return WSTestMethod{}, "", 0, fmt.Errorf("请求报文 %d 字节，超过 %dKB 上限", len(req.Request), wsTestMaxRequest/1024)
	}
	raw := strings.TrimSpace(req.URL)
	if raw == "" {
		raw = wsTestURL(cfg, m.ID)
		if raw == "" {
			return WSTestMethod{}, "", 0, fmt.Errorf("无法生成默认地址（环境缺少登录区域或服务名），请手动填写集成方式")
		}
	}
	raw = stripWSDLSuffix(raw)
	if err := validateWSTestURL(raw); err != nil {
		return WSTestMethod{}, "", 0, err
	}
	return m, raw, clampWSTestTimeout(req.TimeoutSec), nil
}

// stripWSDLSuffix 去掉 ?wsdl（不区分大小写）。原生 req_test 同样处理，
// 否则 POST 会打到 WSDL 描述页而不是服务端点。
func stripWSDLSuffix(raw string) string {
	const suffix = "?wsdl"
	if len(raw) >= len(suffix) && strings.EqualFold(raw[len(raw)-len(suffix):], suffix) {
		return raw[:len(raw)-len(suffix)]
	}
	return raw
}

// soapEnvelope 把请求载荷包进 TIPTOP SOAP 信封（与 awsq990 btn_test 的 1/5 分支一致）。
// 原生只转义 < >；这里加 &（单趟替换，先 & 防二次转义），否则载荷里的 & 会产出非法 XML。
func soapEnvelope(payload string) string {
	esc := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(payload)
	return `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tip="http://www.digiwin.com.cn/tiptop/TIPTOPServiceGateWay">` + "\n" +
		"   <soapenv:Header/>\n" +
		"   <soapenv:Body>\n" +
		"      <tip:invokeSrv>\n" +
		"         <request>\n" + esc + "\n" +
		"         </request>\n" +
		"      </tip:invokeSrv>\n" +
		"   </soapenv:Body>\n" +
		"</soapenv:Envelope>"
}

// decodeXMLEntities 反转义 SOAP 响应里的实体。
// 原生用「」替换 &quot;（疑似笔误），这里按 XML 标准用双引号；其余一致。
func decodeXMLEntities(s string) string {
	return strings.NewReplacer(
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&apos;", "'",
		"&amp;", "&",
	).Replace(s)
}

// extractSOAPResponse 反转义后截取 <response>…</response>（含标签）。
// 截不到就返回原文 —— 不丢内容，也不假装成功。
func extractSOAPResponse(s string) string {
	d := decodeXMLEntities(s)
	start := strings.Index(d, "<response>")
	end := strings.Index(d, "</response>")
	if start < 0 || end < start {
		return s
	}
	return d[start : end+len("</response>")]
}

// reHTTPStatusLine 解析 curl -D - 输出的状态行：HTTP/1.1 400 Proxy error / HTTP/2 200。
var reHTTPStatusLine = regexp.MustCompile(`^HTTP/\d(?:\.\d)?\s+(\d{3})\s*(.*)$`)

// parseCurlOutput 解析 curl 的输出（响应头 + 正文 + 尾部 META）。
//
//	curl ... -D - -w '\n---TTMETA---%{http_code}|%{time_total}|%{size_download}'
//
// 返回 code/status/body/duration/transportErr。纯函数，便于单测。
func parseCurlOutput(out string) (code int, status, body string, dur float64, transportErr string) {
	const metaMark = "\n---TTMETA---"
	metaOK := false
	meta := ""
	if i := strings.LastIndex(out, metaMark); i >= 0 {
		meta = strings.TrimSpace(out[i+len(metaMark):])
		out = out[:i]
		metaOK = true
		if parts := strings.Split(meta, "|"); len(parts) >= 2 {
			code, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
			dur, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		}
	}
	// 头/正文分隔：优先 CRLF，其次 LF。
	headerPart, bodyPart := out, ""
	if i := strings.Index(out, "\r\n\r\n"); i >= 0 {
		headerPart, bodyPart = out[:i], out[i+4:]
	} else if i := strings.Index(out, "\n\n"); i >= 0 {
		headerPart, bodyPart = out[:i], out[i+2:]
	}
	statusLine := ""
	for _, ln := range strings.Split(headerPart, "\n") {
		ln = strings.TrimRight(ln, "\r")
		if strings.TrimSpace(ln) != "" {
			statusLine = ln
			break
		}
	}
	body = bodyPart
	if m := reHTTPStatusLine.FindStringSubmatch(statusLine); m != nil {
		c, _ := strconv.Atoi(m[1])
		code = c
		status = strings.TrimSpace(m[2])
		return code, status, body, dur, ""
	}
	// 没有状态行：curl 没拿到 HTTP 响应（连接被拒/解析失败/超时），或输出被裁掉。
	if metaOK && code == 0 {
		return 0, "", body, dur, "没有拿到 HTTP 状态码（连接被拒 / 域名解析失败 / 超时；服务器侧 curl 已返回）"
	}
	msg := strings.TrimSpace(out)
	if msg == "" {
		msg = "无 HTTP 响应（连接失败或超时）"
	}
	if len(msg) > 300 {
		msg = msg[:300]
	}
	return 0, "", body, dur, "无 HTTP 响应: " + msg
}

// wsCurlCmd 拼 curl 命令串：命令里只有常量与已转义的 URL，报文走 stdin。
func wsCurlCmd(m WSTestMethod, url string, timeoutSec int) string {
	headers := `-H 'Content-Type: application/json; charset=UTF-8'`
	if m.SOAP {
		headers = `-H 'Content-Type: text/xml; charset=UTF-8' -H 'SOAPAction: ""' -H 'User-Agent: Jakarta Commons-HttpClient/3.0.1'`
	}
	return fmt.Sprintf(
		`curl -sS --max-redirs 0 --max-time %d -H 'Expect:' %s --data-binary @- -D - -w '\n---TTMETA---%%{http_code}|%%{time_total}|%%{size_download}' %s`,
		timeoutSec, headers, shQuote(url))
}

// shQuote 单引号包裹并转义内部单引号（POSIX 惯例）。
func shQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// WSTest 执行一次检测。返回的 *WSTestResult 即使传输失败也非 nil：
// 「连不上」是这个工具要回答的问题，不应该变成 500。
func WSTest(conn *host.SSHConn, m WSTestMethod, rawURL, request string, timeoutSec int) *WSTestResult {
	timeoutSec = clampWSTestTimeout(timeoutSec)
	res := &WSTestResult{At: time.Now().Format("2006-01-02 15:04:05.000")}
	body := request
	if m.SOAP {
		body = soapEnvelope(request)
	}
	res.RequestBytes = len(body)
	// curl 不存在时用一个哨兵先返回；与请求共用一次 SSH（不额外拨一次）。
	cmd := `command -v curl >/dev/null 2>&1 || { echo TT_NO_CURL; exit 127; }; ` + wsCurlCmd(m, rawURL, timeoutSec)
	out, err := conn.OutputStdin(cmd, []byte(body), time.Duration(timeoutSec+15)*time.Second)
	if strings.Contains(out, "TT_NO_CURL") {
		res.TransportError = "服务器上没有 curl，无法执行检测（请在该服务器安装 curl，或改从本机手工请求）"
		return res
	}
	code, status, respBody, dur, transportErr := parseCurlOutput(out)
	if transportErr != "" {
		// 把 curl 的 stderr 摘要也带上；out 可能含响应头，截断即可。
		detail := strings.TrimSpace(out)
		if len(detail) > 400 {
			detail = detail[:400]
		}
		if err != nil && detail == "" {
			detail = err.Error()
		}
		res.TransportError = transportErr
		if detail != "" {
			res.TransportError += "（" + detail + "）"
		}
		res.DurationSec = dur
		return res
	}
	res.HTTPCode = code
	res.StatusText = status
	res.DurationSec = dur
	res.ResponseBytes = len(respBody)
	if m.SOAP && code == 200 {
		respBody = extractSOAPResponse(respBody)
	}
	if len(respBody) > wsTestMaxResponse {
		respBody = respBody[:wsTestMaxResponse]
		res.Truncated = true
	}
	res.Body = respBody
	return res
}
