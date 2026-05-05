package proxy

import (
	"ant-chrome/backend/internal/config"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestValidateProxyConfigInvalidRawString(t *testing.T) {
	ok, msg := ValidateProxyConfig("not-a-proxy-config", nil, "")
	if ok {
		t.Fatalf("expected invalid raw string to fail validation")
	}
	if !strings.Contains(msg, "解析失败") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestValidateProxyConfigMissingProxyId(t *testing.T) {
	ok, msg := ValidateProxyConfig("", []config.BrowserProxy{
		{ProxyId: "p1", ProxyConfig: "http://127.0.0.1:7890"},
	}, "missing-proxy")
	if ok {
		t.Fatalf("expected missing proxyId to fail validation")
	}
	if !strings.Contains(msg, "不存在") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestValidateProxyConfigMissingProxyIdFallbackToRawConfig(t *testing.T) {
	ok, msg := ValidateProxyConfig("socks5://127.0.0.1:1080", []config.BrowserProxy{
		{ProxyId: "p1", ProxyConfig: "http://127.0.0.1:7890"},
	}, "missing-proxy")
	if !ok {
		t.Fatalf("expected fallback proxyConfig to pass, msg=%s", msg)
	}
}

func TestValidateProxyConfigStandardProxy(t *testing.T) {
	ok, msg := ValidateProxyConfig("socks5://127.0.0.1:1080", nil, "")
	if !ok {
		t.Fatalf("expected standard proxy to pass: %s", msg)
	}
}

func TestValidateProxyConfigChainSocks5Proxy(t *testing.T) {
	chainConfig := buildTestChainSocks5Config(t, 0)
	ok, msg := ValidateProxyConfig(chainConfig, nil, "")
	if !ok {
		t.Fatalf("expected chain proxy to pass: %s", msg)
	}
	if !RequiresBridge(chainConfig, nil, "") {
		t.Fatalf("expected chain proxy to require bridge")
	}
}

func buildTestChainSocks5Config(t *testing.T, localPort int) string {
	t.Helper()
	localPortField := ""
	if localPort > 0 {
		localPortField = fmt.Sprintf(`,"localPort":%d`, localPort)
	}
	raw := fmt.Sprintf(`{"first":{"protocol":"socks5","server":"127.0.0.1","port":1081,"username":"u1","password":"p1"},"second":{"protocol":"socks5","server":"127.0.0.2","port":1082}%s}`, localPortField)
	return "chain+socks5://" + url.QueryEscape(raw)
}
