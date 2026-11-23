package main

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"encoding/json"
)

const groupName = "MIKROT-VPN"
const udpGroupName = "MIKROT-VPN-UDP"
const vpnNetworkRules = "vpn-by-network"
const direct = "DIRECT"
const tunDevice = "mikrot-tun"

type rule struct {
	Match  string `json:"match"`
	Value  string `json:"value"`
	Action string `json:"action"`
}

type settings struct {
	ProbeURL      string `json:"probeUrl"`
	ProbeInterval int    `json:"probeInterval"`
	Timeout       int    `json:"timeout"`
	Tolerance     int    `json:"tolerance"`
	RefreshHours  int    `json:"refreshHours"`
	Rules         []rule `json:"rules"`
	DefaultAction string `json:"defaultAction"`
	DNSMapping    bool   `json:"dnsMapping"`
}

type state struct {
	Settings          settings
	Subscription      string
	Nodes             []node
	TelegramCIDRs     []string
	TelegramUpdatedAt time.Time
}

func defaultState() state {
	return state{Settings: settings{
		ProbeURL:      "https://www.gstatic.com/generate_204",
		ProbeInterval: 45, Timeout: 3500, Tolerance: 35, RefreshHours: 12,
		DefaultAction: direct, Rules: []rule{},
	}}
}

var domainPattern = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func validAction(action string) bool { return action == direct || action == "VPN" }

func normalizeSettings(s *settings) error {
	if _, err := remoteURL(s.ProbeURL); err != nil {
		return errors.New("проверьте URL теста")
	}
	if s.ProbeInterval < 15 || s.ProbeInterval > 600 || s.Timeout < 500 || s.Timeout > 10000 ||
		s.Tolerance < 0 || s.Tolerance > 500 || s.RefreshHours < 0 || s.RefreshHours > 168 {
		return errors.New("параметры проверок или обновления вне допустимого диапазона")
	}
	if len(s.Rules) > 128 || !validAction(s.DefaultAction) {
		return errors.New("проверьте число правил и маршрут по умолчанию")
	}
	if s.Rules == nil {
		s.Rules = []rule{}
	}
	for i := range s.Rules {
		r := &s.Rules[i]
		r.Value = strings.ToLower(strings.TrimSpace(r.Value))
		ok := false
		switch r.Match {
		case "SERVICE":
			_, ok = findService(r.Value)
		case "DOMAIN", "DOMAIN-SUFFIX":
			ok = len(r.Value) <= 253 && domainPattern.MatchString(r.Value)
		case "IP-CIDR", "SRC-IP-CIDR":
			prefix, err := netip.ParsePrefix(r.Value)
			if addr, addrErr := netip.ParseAddr(r.Value); addrErr == nil && addr.Is4() {
				prefix, err = netip.PrefixFrom(addr, 32), nil
			}
			ok = err == nil && prefix.Addr().Is4()
			if ok {
				r.Value = prefix.Masked().String()
			}
		}
		if !ok || !validAction(r.Action) {
			return fmt.Errorf("проверьте правило №%d", i+1)
		}
	}
	return nil
}

func coreConfig(s state, tun bool, secret string) ([]byte, error) {
	rules := make([]string, 0, len(s.Settings.Rules)+2)
	providers := make(map[string]any)
	for _, r := range s.Settings.Rules {
		match := r.Match + "," + r.Value
		noResolve := r.Match == "IP-CIDR"
		if r.Match == "SERVICE" {
			preset, ok := findService(r.Value)
			if !ok {
				return nil, fmt.Errorf("неизвестный сервис %q", r.Value)
			}
			provider := "service-" + preset.ID
			providers[provider] = map[string]any{"type": "inline", "behavior": "classical", "payload": preset.rules}
			match = "RULE-SET," + provider
			if preset.ID == telegramService {
				providers[provider] = map[string]any{"type": "inline", "behavior": "ipcidr", "payload": telegramCIDRs(s)}
				noResolve = true
			}
		}
		if r.Action == "VPN" {
			if noResolve {
				match += ",no-resolve"
			}
			rules = append(rules, "SUB-RULE,("+match+"),"+vpnNetworkRules)
			continue
		}
		line := match + "," + direct
		if noResolve {
			line += ",no-resolve"
		}
		rules = append(rules, line)
	}
	if s.Settings.DefaultAction == "VPN" {
		rules = append(rules, "NETWORK,UDP,"+udpGroupName, "MATCH,"+groupName)
	} else {
		rules = append(rules, "MATCH,"+direct)
	}
	proxies := make([]map[string]any, 0, len(s.Nodes))
	members := []string{direct}
	for _, n := range s.Nodes {
		proxies = append(proxies, n.Options)
		members = append(members, n.ID)
	}
	config := map[string]any{
		"external-controller": "127.0.0.1:9090", "secret": secret, "log-level": "silent", "mode": "rule",
		"profile":    map[string]any{"store-selected": false},
		"mixed-port": map[bool]int{false: 7890, true: 0}[tun], "allow-lan": false,
		"ipv6": false, "proxies": proxies, "rules": rules,
		"rule-providers": providers,
		"sub-rules":      map[string][]string{vpnNetworkRules: {"NETWORK,UDP," + udpGroupName, "MATCH," + groupName}},
		"proxy-groups": []map[string]any{
			{"name": groupName, "type": "select", "proxies": members},
			{"name": udpGroupName, "type": "select", "proxies": members},
		},
		"tun": map[string]any{
			"enable": tun, "device": tunDevice, "stack": "system",
			"auto-route": true, "auto-detect-interface": true,
			"route-exclude-address": []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "169.254.0.0/16", "224.0.0.0/4"},
		},
		"sniffer": map[string]any{
			"enable": true, "force-dns-mapping": true, "parse-pure-ip": true, "override-destination": false,
			"sniff": map[string]any{
				"HTTP": map[string]any{"ports": []string{"80", "8080-8880"}},
				"TLS":  map[string]any{"ports": []int{443, 8443}},
				"QUIC": map[string]any{"ports": []int{443, 8443}},
			},
		},
	}
	if tun && s.Settings.DNSMapping {
		config["dns"] = map[string]any{
			"enable": true, "listen": "0.0.0.0:53", "enhanced-mode": "redir-host",
			"ipv6": false, "nameserver": []string{"system"}, "cache-max-size": 1024,
		}
	}
	return json.Marshal(config)
}

func writeAtomic(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".mikrot-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
