package proxy

import (
	"encoding/json"
	"os"
	"testing"

	"ant-chrome/backend/internal/config"
)

func TestChainSocks5RuntimeConfigRoutesThroughSecondHop(t *testing.T) {
	chainConfig := buildTestChainSocks5Config(t, 19090)
	chainCfg, err := ParseChainSocks5Config(chainConfig)
	if err != nil {
		t.Fatalf("ParseChainSocks5Config returned error: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Browser.UserDataRoot = t.TempDir()
	manager := &XrayManager{
		Config:  cfg,
		AppRoot: t.TempDir(),
	}

	cfgPath, err := manager.buildRuntimeConfigWithRoute(
		"chain-test",
		[]interface{}{
			chainSocks5Outbound(chainCfg.First, "first-hop", ""),
			chainSocks5Outbound(chainCfg.Second, "second-hop", "first-hop"),
		},
		[]interface{}{
			map[string]interface{}{
				"type":        "field",
				"inboundTag":  []string{"socks-in"},
				"outboundTag": "second-hop",
			},
		},
		chainCfg.LocalPort,
		"",
	)
	if err != nil {
		t.Fatalf("buildRuntimeConfigWithRoute returned error: %v", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read runtime config failed: %v", err)
	}
	var runtimeConfig map[string]interface{}
	if err := json.Unmarshal(data, &runtimeConfig); err != nil {
		t.Fatalf("unmarshal runtime config failed: %v", err)
	}

	inbounds := runtimeConfig["inbounds"].([]interface{})
	inbound := inbounds[0].(map[string]interface{})
	if got := int(inbound["port"].(float64)); got != 19090 {
		t.Fatalf("inbound port = %d, want 19090", got)
	}

	outbounds := runtimeConfig["outbounds"].([]interface{})
	byTag := map[string]map[string]interface{}{}
	for _, item := range outbounds {
		outbound := item.(map[string]interface{})
		if tag, ok := outbound["tag"].(string); ok {
			byTag[tag] = outbound
		}
	}
	secondHop := byTag["second-hop"]
	if secondHop == nil {
		t.Fatalf("second-hop outbound is missing: %+v", byTag)
	}
	proxySettings, ok := secondHop["proxySettings"].(map[string]interface{})
	if !ok {
		t.Fatalf("second-hop proxySettings is missing: %+v", secondHop)
	}
	if got := proxySettings["tag"]; got != "first-hop" {
		t.Fatalf("second-hop proxy tag = %v, want first-hop", got)
	}

	routing := runtimeConfig["routing"].(map[string]interface{})
	rules := routing["rules"].([]interface{})
	rule := rules[0].(map[string]interface{})
	if got := rule["outboundTag"]; got != "second-hop" {
		t.Fatalf("route outboundTag = %v, want second-hop", got)
	}
}
