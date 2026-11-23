package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestNodeSourcesMigrateAndManageIndependently(t *testing.T) {
	const (
		legacyKey = "trojan://legacy-secret@example.com:443#Old"
		manualKey = "trojan://manual-secret@example.net:443#Manual"
		nextKey   = "trojan://next-secret@example.org:443#Next"
	)
	var fail atomic.Bool
	var updated atomic.Bool
	feed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/old-url-token" {
			if fail.Load() {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_, _ = io.WriteString(w, legacyKey)
			return
		}
		body := manualKey
		if updated.Load() {
			body = nextKey
		}
		_, _ = io.WriteString(w, body)
	}))
	defer feed.Close()
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"now":"DIRECT"}`)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer core.Close()

	dir := t.TempDir()
	legacy := defaultState()
	legacy.Subscription = feed.URL + "/old-url-token"
	var err error
	legacy.Nodes, err = parseSubscription([]byte(legacyKey))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "state.json"), payload, 0600); err != nil {
		t.Fatal(err)
	}
	s, path, err := loadState(dir)
	if err != nil || s.Subscription != "" || len(s.Sources) != 1 || s.Sources[0].ID != legacySourceID {
		t.Fatalf("legacy subscription did not migrate: %v", err)
	}
	m := manager{current: s, path: path, subClient: feed.Client(), core: coreClient{base: core.URL, client: core.Client()}}
	ctx := context.Background()
	if err := m.addSource(ctx, "", manualKey); err != nil {
		t.Fatal(err)
	}
	if err := m.addSource(ctx, "", feed.URL+"/new-url-token"); err != nil {
		t.Fatal(err)
	}
	status := m.snapshot()
	if len(status.Sources) != 3 || len(status.Nodes) != 2 {
		t.Fatal("sources were not combined and deduplicated")
	}
	visible, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"old-url-token", "new-url-token", "manual-secret", "legacy-secret"} {
		if strings.Contains(string(visible), secret) {
			t.Fatal("source credentials appeared in status")
		}
	}
	fail.Store(true)
	updated.Store(true)
	if err := m.refreshSource(ctx, legacySourceID); err == nil {
		t.Fatal("failed refresh was reported as successful")
	}
	if err := m.refreshSource(ctx, status.Sources[2].ID); err != nil {
		t.Fatal(err)
	}
	if result := m.snapshot(); len(result.Nodes) != 3 || result.Sources[0].Error == "" || result.Sources[2].Error != "" {
		t.Fatal("refresh did not isolate the failed source")
	}
	if err := m.deleteSource(ctx, status.Sources[1].ID); err != nil {
		t.Fatal(err)
	}
	saved, _, err := loadState(dir)
	if err != nil || len(saved.Nodes) != 2 || len(saved.Sources) != 2 || saved.Sources[1].Nodes[0].Label != "Next" {
		t.Fatalf("deletion and refresh were not persisted: %v", err)
	}
}
