package web

import (
	"fmt"
	"strings"
	"testing"

	"tt/internal/host"
)

// 「SSH 测试连接」的对内契约:参数校验、端口缺省、成功信封、失败信封。
//
// 真实现要真登一台 SSH,所以默认档只能测注入的假实现 —— 与 dbprobe/dbaccverify
// 同一取舍(见 Options 里那几个函数字段的注释):能钉住的是**对外形状**。

func TestSSHTestRejectsEmptyHostAndUser(t *testing.T) {
	s := New(Options{Version: "test"})
	cases := []struct {
		name string
		body map[string]any
		want string
	}{
		{"empty host", map[string]any{"host": "   ", "user": "u", "password": "p"}, "IP 主机不能为空"},
		{"empty user", map[string]any{"host": "10.0.0.1", "user": " ", "password": "p"}, "账号不能为空"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec, body := doJSON(t, s, "POST", "/api/sshtest", c.body)
			if rec.Code != 400 {
				t.Fatalf("status = %d, want 400 (body %v)", rec.Code, body)
			}
			if ok, _ := body["ok"].(bool); ok {
				t.Fatalf("ok = true, want false: %v", body)
			}
			if msg, _ := body["error"].(string); !strings.Contains(msg, c.want) {
				t.Fatalf("error = %q, want contains %q", msg, c.want)
			}
		})
	}
}

func TestSSHTestOKDefaultsPortAndTakesHostname(t *testing.T) {
	var got host.SSHConfig
	s := New(Options{
		Version: "test",
		SSHTest: func(cfg host.SSHConfig) (*host.SSHTestOut, error) {
			got = cfg
			return &host.SSHTestOut{Hostname: "t100dev"}, nil
		},
	})
	rec, body := doJSON(t, s, "POST", "/api/sshtest", map[string]any{
		"host": " 10.0.0.1 ", "port": 0, "user": " u ", "password": "pw",
	})
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200 (body %v)", rec.Code, body)
	}
	// 空白要去掉:表单里粘进来的地址常带尾随空格,原样透传会让 "host:22" 带空格连不上。
	if got.Host != "10.0.0.1" || got.User != "u" {
		t.Fatalf("cfg = %+v, want trimmed host/user", got)
	}
	// 端口 0 必须补成 22,而不是把 "host:0" 发出去。
	if got.Port != 22 {
		t.Fatalf("port = %d, want 22", got.Port)
	}
	if body["ok"] != true || body["hostname"] != "t100dev" {
		t.Fatalf("body = %v", body)
	}
	if body["address"] != "10.0.0.1:22" {
		t.Fatalf("address = %v, want 10.0.0.1:22", body["address"])
	}
}

func TestSSHTestKeepsNoteWhenHostnameUnavailable(t *testing.T) {
	s := New(Options{
		Version: "test",
		SSHTest: func(cfg host.SSHConfig) (*host.SSHTestOut, error) {
			return &host.SSHTestOut{Note: "已登录，但读取主机名失败: x"}, nil
		},
	})
	rec, body := doJSON(t, s, "POST", "/api/sshtest", map[string]any{
		"host": "10.0.0.1", "port": 2222, "user": "u", "password": "pw",
	})
	// 取不到主机名不是失败:登录本身已经成功。
	if rec.Code != 200 || body["ok"] != true {
		t.Fatalf("status = %d body = %v, want 200 ok", rec.Code, body)
	}
	if _, ok := body["hostname"]; ok {
		t.Fatalf("hostname should be omitted when empty: %v", body)
	}
	if note, _ := body["note"].(string); !strings.Contains(note, "读取主机名失败") {
		t.Fatalf("note = %v", body["note"])
	}
	if body["address"] != "10.0.0.1:2222" {
		t.Fatalf("address = %v", body["address"])
	}
}

func TestSSHTestWrapsDialError(t *testing.T) {
	s := New(Options{
		Version: "test",
		SSHTest: func(cfg host.SSHConfig) (*host.SSHTestOut, error) {
			return nil, fmt.Errorf("dial tcp 10.0.0.1:22: i/o timeout")
		},
	})
	rec, body := doJSON(t, s, "POST", "/api/sshtest", map[string]any{
		"host": "10.0.0.1", "port": 22, "user": "u", "password": "pw",
	})
	if rec.Code != 502 {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
	msg, _ := body["error"].(string)
	// 前缀是接口契约的一部分:前端直接把这句话贴进 toast。
	if !strings.HasPrefix(msg, "SSH 连接失败: ") || !strings.Contains(msg, "i/o timeout") {
		t.Fatalf("error = %q", msg)
	}
}
