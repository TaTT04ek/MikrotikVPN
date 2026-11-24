package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"
)

func TestServicePackImportAndRouting(t *testing.T) {
	dir := t.TempDir()
	m := manager{current: defaultState(), path: filepath.Join(dir, "state.json")}
	pack := []serviceDefinition{{ID: "custom", Label: "Custom", Domains: []string{"exact.example.org", "+.example.net"}}}
	bad := append(slices.Clone(pack), serviceDefinition{ID: "telegram", Label: "Collision", Domains: []string{"example.com"}})
	if err := m.importServices(context.Background(), bad); err == nil || len(m.current.Services) != 0 {
		t.Fatal("conflicting pack must be rejected in full")
	}
	if err := m.importServices(context.Background(), pack); err != nil {
		t.Fatal(err)
	}
	if len(m.current.Settings.Rules) != 0 {
		t.Fatal("import activated a rule")
	}
	s, _, err := loadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	s.Settings.Rules = []rule{{Match: "SERVICE", Value: "custom", Action: "VPN"}}
	if err := normalizeSettings(&s.Settings, s.Services...); err != nil {
		t.Fatal(err)
	}
	config, err := coreConfig(s, true, "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	var rendered struct {
		Providers map[string]struct{ Payload []string } `json:"rule-providers"`
		Rules     []string
	}
	if err := json.Unmarshal(config, &rendered); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(rendered.Providers["service-custom"].Payload, []string{"DOMAIN,exact.example.org", "DOMAIN-SUFFIX,example.net"}) ||
		!slices.Contains(rendered.Rules, "SUB-RULE,(RULE-SET,service-custom),vpn-by-network") {
		t.Fatal("wrong service routing rule")
	}
}
