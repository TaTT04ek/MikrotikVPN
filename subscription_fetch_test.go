package main

import (
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestFetchSubscriptionRemnawave(t *testing.T) {
	const urlLink = "vless://00000000-0000-4000-8000-000000000001@example.com:443#Sample"
	for _, test := range []struct {
		name         string
		firstHTML    bool
		wantRequests []string
	}{
		{"ordinary URL returns Base64", false, []string{"MikroTikVPN/1.0"}},
		{"HTML response falls back to Mihomo", true, []string{"MikroTikVPN/1.0", "mihomo/1.19.32"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			requests := make(chan string, 3)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r.UserAgent()
				if r.URL.Path != "/sub/fake-token" {
					http.Error(w, "unexpected path", http.StatusBadRequest)
					return
				}
				if test.firstHTML && r.UserAgent() == "MikroTikVPN/1.0" {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					_, _ = w.Write([]byte("<html>Subscription page</html>"))
					return
				}
				if test.firstHTML {
					_, _ = w.Write([]byte("proxies:\n  - {name: Sample, type: vless, server: example.com, port: 443, uuid: 00000000-0000-4000-8000-000000000001}\n"))
					return
				}
				_, _ = w.Write([]byte(base64.StdEncoding.EncodeToString([]byte(urlLink))))
			}))
			defer server.Close()
			nodes, err := fetchSubscription(context.Background(), server.Client(), server.URL+"/sub/fake-token")
			if err != nil || len(nodes) != 1 || nodes[0].Kind != "vless" {
				t.Fatalf("fetchSubscription() returned %d nodes and error %v", len(nodes), err)
			}
			var got []string
			for len(requests) > 0 {
				got = append(got, <-requests)
			}
			if !slices.Equal(got, test.wantRequests) {
				t.Fatalf("User-Agents = %v, want %v", got, test.wantRequests)
			}
		})
	}
}

func TestSubscriptionRequestErrorHidesURL(t *testing.T) {
	wrapped := &url.Error{Op: "Get", URL: "https://example.com/sub/secret-token", Err: &net.DNSError{Err: "not found", Name: "example.com"}}
	message := subscriptionRequestError(wrapped).Error()
	if !strings.Contains(message, "DNS") || strings.Contains(message, "secret-token") {
		t.Fatalf("unexpected public diagnostic: %q", message)
	}
}
