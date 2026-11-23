package main

import (
	"encoding/base64"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestParseSubscription(t *testing.T) {
	links := "vless://00000000-0000-4000-8000-000000000001@example.com:443?security=reality&type=tcp&pbk=test&sid=abcd#VLESS\n" +
		"trojan://sample@example.org:443#Trojan"
	yamlFeed := "proxies:\n  - {name: Sample, type: ss, server: example.com, port: 443, cipher: aes-128-gcm, password: fake}\n" +
		"  - {name: Duplicate, type: ss, server: example.com, port: 443, cipher: aes-128-gcm, password: fake}"
	cases := []struct {
		name, body string
		kinds      []string
	}{
		{"text", links, []string{"vless", "trojan"}},
		{"base64", base64.RawStdEncoding.EncodeToString([]byte(links)), []string{"vless", "trojan"}},
		{"clash and duplicate", yamlFeed, []string{"ss"}},
		{"invalid", "<html>error</html>", nil},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			nodes, err := parseSubscription([]byte(test.body))
			if test.kinds == nil {
				if err == nil {
					t.Fatal("invalid subscription accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var kinds []string
			for _, node := range nodes {
				kinds = append(kinds, node.Kind)
				if node.Options["name"] != node.ID {
					t.Fatal("node alias missing")
				}
			}
			if !slices.Equal(kinds, test.kinds) {
				t.Fatalf("protocols = %v, want %v", kinds, test.kinds)
			}
		})
	}
}

func TestBestNode(t *testing.T) {
	nodes := []node{{ID: "a"}, {ID: "b"}}
	for _, test := range []struct {
		name      string
		checks    []nodeCheck
		selected  string
		tolerance int
		udp       bool
		want      string
	}{
		{"fastest", []nodeCheck{{Delay: 50}, {Delay: 30}}, direct, 0, false, "b"},
		{"hysteresis", []nodeCheck{{Delay: 50}, {Delay: 30}}, "a", 20, false, "a"},
		{"active failed", []nodeCheck{{}, {Delay: 120}}, "a", 100, false, "b"},
		{"all failed", []nodeCheck{{}, {}}, "a", 20, false, direct},
		{"no subscription", nil, direct, 20, false, direct},
		{"skip TCP-only node", []nodeCheck{{Delay: 50}, {Delay: 80, UDP: true}}, "a", 100, true, "b"},
		{"no available UDP node", []nodeCheck{{Delay: 20}, {UDP: true}}, "b", 0, true, direct},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := bestNode(nodes, test.checks, test.selected, test.tolerance, test.udp); got != test.want {
				t.Fatalf("chosen %q, want %q", got, test.want)
			}
		})
	}
}

func TestCoreConfigPreservesRuleOrder(t *testing.T) {
	s := defaultState()
	s.Settings.Rules = []rule{{"DOMAIN-SUFFIX", " example.com ", "VPN"}, {"SRC-IP-CIDR", "192.168.88.50", direct}, {"SERVICE", " YouTube ", "VPN"}, {"IP-CIDR", "198.51.100.15/24", "VPN"}}
	if err := normalizeSettings(&s.Settings); err != nil {
		t.Fatal(err)
	}
	data, err := coreConfig(s, true, "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Rules     []string
		SubRules  map[string][]string                   `json:"sub-rules"`
		Providers map[string]struct{ Payload []string } `json:"rule-providers"`
	}
	if err = json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	want := []string{"SUB-RULE,(DOMAIN-SUFFIX,example.com),vpn-by-network", "SRC-IP-CIDR,192.168.88.50/32,DIRECT", "SUB-RULE,(RULE-SET,service-youtube),vpn-by-network", "SUB-RULE,(IP-CIDR,198.51.100.0/24,no-resolve),vpn-by-network", "MATCH,DIRECT"}
	if !slices.Equal(config.Rules, want) {
		t.Fatalf("rules = %v, want %v", config.Rules, want)
	}
	if !slices.Equal(config.SubRules[vpnNetworkRules], []string{"NETWORK,UDP,MIKROT-VPN-UDP", "MATCH,MIKROT-VPN"}) {
		t.Fatalf("unexpected network routing: %v", config.SubRules)
	}
	payload := config.Providers["service-youtube"].Payload
	for _, domain := range []string{"youtube.com", "googlevideo.com", "ytimg.com", "youtubei.googleapis.com"} {
		if !slices.Contains(payload, "DOMAIN-SUFFIX,"+domain) {
			t.Errorf("missing %s", domain)
		}
	}
	if slices.Contains(payload, "DOMAIN-SUFFIX,googleapis.com") {
		t.Fatal("shared Google API suffix matched")
	}
}

func TestServiceCatalogRules(t *testing.T) {
	for _, preset := range servicePresets {
		if preset.ID == telegramService {
			continue // Its IPv4 rule provider is covered in telegram_test.go.
		}
		if len(preset.rules) == 0 {
			t.Fatalf("empty preset: %s", preset.ID)
		}
		for _, line := range preset.rules {
			kind, domain, ok := strings.Cut(line, ",")
			if !ok || (kind != "DOMAIN" && kind != "DOMAIN-SUFFIX") || (domain != "youtube" && !domainPattern.MatchString(domain)) {
				t.Errorf("invalid %s rule: %s", preset.ID, line)
			}
		}
	}
}
