package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type telegramTransport func(*http.Request) (*http.Response, error)

func (f telegramTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestTelegramRefreshReplacesAndKeepsLastGoodList(t *testing.T) {
	dir := t.TempDir()
	current := defaultState()
	current.TelegramCIDRs = []string{"91.108.4.0/22"}
	body := "149.154.160.0/20\n91.105.192.0/23\n2001:b28:f23d::/48\n91.105.192.0/23\n"
	var proxyRequests int
	m := manager{current: current, path: filepath.Join(dir, "state.json"), subClient: &http.Client{
		Transport: telegramTransport(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("direct connection timed out")
		}),
	}, telegramProxy: &http.Client{
		Transport: telegramTransport(func(req *http.Request) (*http.Response, error) {
			proxyRequests++
			if req.URL.String() != telegramURL {
				t.Fatalf("unexpected source %s", req.URL)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
		}),
	}}
	if err := m.refreshTelegram(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{"149.154.160.0/20", "91.105.192.0/23"}
	updated := m.snapshot().Telegram
	if proxyRequests != 1 || !slices.Equal(updated.CIDRs, want) || updated.UpdatedAt.IsZero() {
		t.Fatalf("unexpected refreshed list: %+v", updated)
	}
	body = "2001:b28:f23d::/48"
	if err := m.refreshTelegram(context.Background()); err == nil {
		t.Fatal("IPv6-only feed accepted")
	}
	if failed := m.snapshot().Telegram; !slices.Equal(failed.CIDRs, want) || !failed.UpdatedAt.Equal(updated.UpdatedAt) || failed.Error == "" {
		t.Fatalf("failed refresh changed the last good list: %+v", failed)
	}
	saved, _, err := loadState(dir)
	if err != nil || !slices.Equal(saved.TelegramCIDRs, want) || !saved.TelegramUpdatedAt.Equal(updated.UpdatedAt) {
		t.Fatalf("refresh not retained on restart: %+v, %v", saved, err)
	}
}

func TestTelegramRuleUsesSavedCIDRs(t *testing.T) {
	s := defaultState()
	s.TelegramCIDRs = []string{"149.154.160.0/20"}
	s.Settings.Rules = []rule{{Match: "SERVICE", Value: "telegram", Action: "VPN"}}
	if err := normalizeSettings(&s.Settings); err != nil {
		t.Fatal(err)
	}
	payload, err := coreConfig(s, true, "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Rules     []string
		Providers map[string]struct {
			Behavior string
			Payload  []string
		} `json:"rule-providers"`
	}
	if err := json.Unmarshal(payload, &config); err != nil {
		t.Fatal(err)
	}
	provider := config.Providers["service-telegram"]
	if provider.Behavior != "ipcidr" || !slices.Equal(provider.Payload, s.TelegramCIDRs) ||
		!slices.Contains(config.Rules, "SUB-RULE,(RULE-SET,service-telegram,no-resolve),vpn-by-network") {
		t.Fatalf("unexpected Telegram routing rule: %+v", config)
	}
}
