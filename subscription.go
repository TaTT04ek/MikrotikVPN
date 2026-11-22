package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/metacubex/mihomo/common/convert"
	"gopkg.in/yaml.v3"
)

const maxNodes = 64
const maxFeed = 2 << 20

type node struct {
	ID      string
	Label   string
	Kind    string
	Options map[string]any
}

func parseSubscription(data []byte) ([]node, error) {
	data = bytes.TrimSpace(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}))
	var feed struct{ Proxies []map[string]any }
	if err := yaml.Unmarshal(data, &feed); err != nil || len(feed.Proxies) == 0 {
		feed.Proxies, _ = convert.ConvertsV2Ray(data)
	}
	seen := make(map[string]bool)
	nodes := make([]node, 0, min(maxNodes, len(feed.Proxies)))
	for _, options := range feed.Proxies {
		kind, _ := options["type"].(string)
		switch kind {
		case "vless", "vmess", "trojan", "ss", "hysteria2", "tuic":
		default:
			continue
		}
		host, _ := options["server"].(string)
		port, err := strconv.Atoi(fmt.Sprint(options["port"]))
		if host == "" || err != nil || port < 1 || port > 65535 {
			continue
		}
		label, _ := options["name"].(string)
		label = strings.TrimSpace(strings.Map(func(r rune) rune {
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
				return -1
			}
			return r
		}, label))
		labelRunes := []rune(label)
		label = string(labelRunes[:min(100, len(labelRunes))])
		delete(options, "name")
		delete(options, "dialer-proxy")
		delete(options, "interface-name")
		delete(options, "routing-mark")
		normalized, err := json.Marshal(options)
		if err != nil {
			continue
		}
		hash := sha256.Sum256(normalized)
		id := fmt.Sprintf("node-%x", hash[:12])
		if seen[id] {
			continue
		}
		seen[id] = true
		if len(nodes) == maxNodes {
			return nil, fmt.Errorf("подписка содержит больше %d уникальных узлов", maxNodes)
		}
		if label == "" {
			label = host
		}
		options["name"] = id
		nodes = append(nodes, node{ID: id, Label: label, Kind: kind, Options: options})
	}
	if len(nodes) == 0 {
		return nil, errors.New("подписка не содержит поддерживаемых узлов")
	}
	return nodes, nil
}

func remoteURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.Fragment != "" {
		return nil, errors.New("укажите ссылку HTTP или HTTPS без логина и фрагмента")
	}
	return u, nil
}

func publicAddress(a netip.Addr) bool {
	a = a.Unmap()
	return a.IsGlobalUnicast() && !a.IsPrivate() &&
		!netip.MustParsePrefix("100.64.0.0/10").Contains(a) &&
		!netip.MustParsePrefix("198.18.0.0/15").Contains(a)
}

func subscriptionClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return nil, err
		}
		dialer := net.Dialer{Timeout: 6 * time.Second}
		for _, candidate := range addresses {
			if !publicAddress(candidate) {
				continue
			}
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(candidate.String(), port))
			if err == nil {
				return conn, nil
			}
		}
		return nil, errors.New("сервер подписки недоступен по публичному адресу")
	}
	return &http.Client{
		Transport: transport, Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if _, err := remoteURL(req.URL.String()); err != nil {
				return err
			}
			if len(via) >= 5 || (via[0].URL.Scheme == "https" && req.URL.Scheme != "https") {
				return errors.New("недопустимое перенаправление подписки")
			}
			return nil
		},
	}
}

func fetchSubscription(ctx context.Context, client *http.Client, raw string) ([]node, error) {
	u, err := remoteURL(raw)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.New("не удалось подготовить запрос подписки")
	}
	req.Header.Set("User-Agent", "mihomo/1.19.32")
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("не удалось скачать подписку")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервер подписки ответил HTTP %d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxFeed+1))
	if err != nil {
		return nil, errors.New("не удалось прочитать подписку")
	}
	if len(data) > maxFeed {
		return nil, errors.New("ответ подписки больше 2 МиБ")
	}
	return parseSubscription(data)
}
