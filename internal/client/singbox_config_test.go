package client

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuildSingBoxConfig(t *testing.T) {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	cfg, err := BuildSingBoxConfig(SingBoxConfigInput{Relay: validRelay(now)})
	if err != nil {
		t.Fatalf("build sing-box config: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(cfg, &decoded); err != nil {
		t.Fatalf("config should be valid JSON: %v", err)
	}

	inbounds := decoded["inbounds"].([]any)
	tun := inbounds[0].(map[string]any)
	if tun["type"] != "tun" || tun["auto_route"] != true || tun["strict_route"] != true {
		t.Fatalf("unexpected tun inbound: %+v", tun)
	}
	if _, ok := tun["dns_mode"]; ok {
		t.Fatalf("dns_mode requires sing-box 1.14+, but desktop MVP targets 1.13.x: %+v", tun)
	}

	outbounds := decoded["outbounds"].([]any)
	proxy := outbounds[0].(map[string]any)
	if proxy["type"] != "vless" || proxy["server"] != "volunteer.example.com" {
		t.Fatalf("unexpected proxy outbound: %+v", proxy)
	}
	if proxy["server_port"].(float64) != 443 {
		t.Fatalf("unexpected server port: %+v", proxy["server_port"])
	}

	tls := proxy["tls"].(map[string]any)
	reality := tls["reality"].(map[string]any)
	if tls["server_name"] != "www.microsoft.com" ||
		reality["public_key"] != "public-key" ||
		reality["short_id"] != "5f7a8d9c01ab23cd" {
		t.Fatalf("unexpected reality TLS config: %+v", tls)
	}

	route := decoded["route"].(map[string]any)
	rules := route["rules"].([]any)
	dnsRule := rules[0].(map[string]any)
	if dnsRule["protocol"] != "dns" || dnsRule["action"] != "hijack-dns" {
		t.Fatalf("expected DNS hijack rule, got %+v", dnsRule)
	}
	if route["final"] != "proxy" {
		t.Fatalf("expected proxy final route, got %+v", route["final"])
	}
	if route["default_domain_resolver"] != "dns-0" {
		t.Fatalf("expected dns-0 default domain resolver, got %+v", route["default_domain_resolver"])
	}

	dns := decoded["dns"].(map[string]any)
	servers := dns["servers"].([]any)
	dns0 := servers[0].(map[string]any)
	if dns0["type"] != "tcp" || dns0["detour"] != "proxy" {
		t.Fatalf("expected TCP DNS through proxy, got %+v", dns0)
	}
}
