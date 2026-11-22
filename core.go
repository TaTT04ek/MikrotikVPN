package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type coreClient struct {
	base, secret string
	client       *http.Client
}

func (c coreClient) request(ctx context.Context, method, path string, body, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.secret)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("Mihomo API: HTTP %d", res.StatusCode)
	}
	if result == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(result)
}

func (c coreClient) reload(ctx context.Context, data []byte) error {
	return c.request(ctx, http.MethodPut, "/configs", map[string]string{"payload": string(data)}, nil)
}

func (c coreClient) selectNode(ctx context.Context, group, id string) error {
	return c.request(ctx, http.MethodPut, "/proxies/"+group, map[string]string{"name": id}, nil)
}

func (c coreClient) selected(ctx context.Context, group string) (string, error) {
	var result struct{ Now string }
	err := c.request(ctx, http.MethodGet, "/proxies/"+group, nil, &result)
	return result.Now, err
}

type nodeCheck struct {
	Delay int
	UDP   bool
}

func (c coreClient) probe(ctx context.Context, id string, s settings) nodeCheck {
	var result struct{ Delay int }
	query := url.Values{"url": {s.ProbeURL}, "timeout": {strconv.Itoa(s.Timeout)}, "expected": {"200-399"}}
	path := "/proxies/" + id
	started := time.Now()
	// Mihomo records HTTP status validation in the per-URL history. The
	// delay endpoint can return 200 for an unexpected HTTP status and 503 for a sub-ms success.
	_ = c.request(ctx, http.MethodGet, path+"/delay?"+query.Encode(), nil, &result)
	var health struct {
		UDP   bool
		Extra map[string]struct {
			Alive   bool
			History []struct {
				Time  time.Time
				Delay int
			}
		}
	}
	if err := c.request(ctx, http.MethodGet, path, nil, &health); err != nil {
		return nodeCheck{}
	}
	resultCheck := nodeCheck{UDP: health.UDP}
	check := health.Extra[s.ProbeURL]
	if !check.Alive || len(check.History) == 0 {
		return resultCheck
	}
	last := check.History[len(check.History)-1]
	if last.Time.Before(started) {
		return resultCheck
	}
	resultCheck.Delay = max(1, last.Delay)
	return resultCheck
}

func startCore(ctx context.Context, binary, dir string, s state, tun bool, c coreClient) (<-chan error, error) {
	if tun {
		if err := enableIPv4Forwarding(); err != nil {
			return nil, fmt.Errorf("не удалось включить net.ipv4.ip_forward для TUN: %w", err)
		}
	}
	data, err := coreConfig(s, tun, c.secret)
	if err != nil {
		return nil, err
	}
	config := filepath.Join(dir, "config.json")
	if err = writeAtomic(config, data); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, binary, "-d", dir, "-f", config)
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	ready, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ready.Done():
			return nil, errors.New("Mihomo не запустился за 15 секунд")
		case err := <-done:
			return nil, fmt.Errorf("Mihomo завершился при запуске: %v", err)
		case <-ticker.C:
			if tun {
				iface, err := net.InterfaceByName(tunDevice)
				if err != nil || iface.Flags&net.FlagUp == 0 {
					continue
				}
			}
			if _, err := c.selected(ready, groupName); err == nil {
				return done, nil
			}
		}
	}
}

func enableIPv4Forwarding() error {
	const path = "/proc/sys/net/ipv4/ip_forward"
	value, err := os.ReadFile(path)
	if err != nil || strings.TrimSpace(string(value)) == "1" {
		return err
	}
	return os.WriteFile(path, []byte("1\n"), 0)
}
