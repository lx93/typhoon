package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"typhoon/internal/relay"
)

type SingBoxConfigInput struct {
	Relay             relay.Descriptor
	TunnelIPv4Address string
	TunnelIPv6Address string
	DNSServers        []string
	MTU               int
}

func BuildSingBoxConfig(input SingBoxConfigInput) ([]byte, error) {
	if err := validateRelayForConfig(input.Relay); err != nil {
		return nil, err
	}

	tunnelIPv4Address := input.TunnelIPv4Address
	if tunnelIPv4Address == "" {
		tunnelIPv4Address = "172.19.0.1/30"
	}
	tunnelIPv6Address := input.TunnelIPv6Address
	if tunnelIPv6Address == "" {
		tunnelIPv6Address = "fdfe:dcba:9876::1/126"
	}
	dnsServers := input.DNSServers
	if len(dnsServers) == 0 {
		dnsServers = []string{"1.1.1.1", "8.8.8.8"}
	}
	mtu := input.MTU
	if mtu == 0 {
		mtu = 1500
	}
	if mtu < 0 {
		return nil, errors.New("mtu must be positive")
	}

	cfg := map[string]any{
		"log": map[string]any{
			"level":     "info",
			"timestamp": true,
		},
		"dns": map[string]any{
			"servers": dnsServerObjects(dnsServers),
			"final":   "dns-0",
		},
		"inbounds": []any{
			map[string]any{
				"type":                     "tun",
				"tag":                      "tun-in",
				"address":                  []string{tunnelIPv4Address, tunnelIPv6Address},
				"mtu":                      mtu,
				"auto_route":               true,
				"strict_route":             true,
				"stack":                    "system",
				"dns_mode":                 "hijack",
				"endpoint_independent_nat": true,
			},
		},
		"outbounds": []any{
			map[string]any{
				"type":            "vless",
				"tag":             "proxy",
				"server":          input.Relay.PublicHost,
				"server_port":     input.Relay.PublicPort,
				"uuid":            input.Relay.ClientID,
				"flow":            input.Relay.Flow,
				"network":         "tcp",
				"packet_encoding": "xudp",
				"tls": map[string]any{
					"enabled":     true,
					"server_name": input.Relay.ServerName,
					"utls": map[string]any{
						"enabled":     true,
						"fingerprint": "chrome",
					},
					"reality": map[string]any{
						"enabled":    true,
						"public_key": input.Relay.RealityPublicKey,
						"short_id":   input.Relay.ShortID,
					},
				},
			},
			map[string]any{
				"type": "direct",
				"tag":  "direct",
			},
			map[string]any{
				"type": "block",
				"tag":  "block",
			},
		},
		"route": map[string]any{
			"auto_detect_interface":   true,
			"default_domain_resolver": "dns-0",
			"rules": []any{
				map[string]any{
					"protocol": "dns",
					"action":   "hijack-dns",
				},
			},
			"final":                   "proxy",
		},
	}

	return json.MarshalIndent(cfg, "", "  ")
}

func dnsServerObjects(servers []string) []any {
	out := make([]any, 0, len(servers))
	for i, server := range servers {
		out = append(out, map[string]any{
			"tag":    fmt.Sprintf("dns-%d", i),
			"type":   "tcp",
			"server": server,
			"detour": "proxy",
		})
	}
	return out
}

func validateRelayForConfig(candidate relay.Descriptor) error {
	if candidate.Protocol != relay.ProtocolVLESSRealityVision {
		return errors.New("relay protocol is not vless-reality-vision")
	}
	if candidate.Flow != relay.FlowVision {
		return errors.New("relay flow is not xtls-rprx-vision")
	}
	if candidate.ExitMode != relay.ExitModeDirect {
		return errors.New("relay exit mode is not direct")
	}
	if !hasRequiredConnectionFields(candidate) {
		return errors.New("relay is missing required connection fields")
	}
	return nil
}
