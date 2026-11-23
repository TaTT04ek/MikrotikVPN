package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type nodeStatus struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Kind  string `json:"kind"`
	Delay int    `json:"delay"`
	UDP   bool   `json:"udp"`
}

type status struct {
	Settings     settings       `json:"settings"`
	Subscribed   bool           `json:"subscribed"`
	Nodes        []nodeStatus   `json:"nodes"`
	Selected     string         `json:"selected"`
	SelectedUDP  string         `json:"selectedUdp"`
	CheckedAt    time.Time      `json:"checkedAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	RefreshError string         `json:"refreshError"`
	CoreError    bool           `json:"coreError"`
	Tun          bool           `json:"tun"`
	Traffic      *traffic       `json:"traffic,omitempty"`
	Telegram     telegramStatus `json:"telegram"`
}

type manager struct {
	importMu      sync.Mutex
	telegramMu    sync.Mutex
	mu            sync.Mutex
	current       state
	revision      uint64
	selected      string
	selectedUDP   string
	checks        []nodeCheck
	checkedAt     time.Time
	updatedAt     time.Time
	refreshError  string
	telegramError string
	coreError     bool
	path          string
	core          coreClient
	tun           bool
	subClient     *http.Client
	telegramProxy *http.Client
	probeNow      chan struct{}
	refreshNow    chan struct{}
}

func (m *manager) snapshot() status {
	m.mu.Lock()
	defer m.mu.Unlock()
	nodes := make([]nodeStatus, 0, len(m.current.Nodes))
	for i, n := range m.current.Nodes {
		entry := nodeStatus{ID: n.ID, Label: n.Label, Kind: n.Kind}
		if i < len(m.checks) {
			entry.Delay, entry.UDP = m.checks[i].Delay, m.checks[i].UDP
		}
		nodes = append(nodes, entry)
	}
	return status{Settings: m.current.Settings, Subscribed: m.current.Subscription != "", Nodes: nodes,
		Selected: m.selected, SelectedUDP: m.selectedUDP, CheckedAt: m.checkedAt, UpdatedAt: m.updatedAt,
		RefreshError: m.refreshError, CoreError: m.coreError, Tun: m.tun,
		Telegram: telegramStatus{CIDRs: telegramCIDRs(m.current), UpdatedAt: m.current.TelegramUpdatedAt, Error: m.telegramError}}
}

// apply is called with mu held so persistence, runtime configuration and status advance together.
func (m *manager) apply(ctx context.Context, next state) error {
	saved, err := json.Marshal(next)
	if err != nil {
		return errors.New("не удалось подготовить настройки")
	}
	payload, err := coreConfig(next, m.tun, m.core.secret)
	if err != nil {
		return errors.New("не удалось подготовить параметры ядра")
	}
	if err := m.core.reload(ctx, payload); err != nil {
		m.restoreCore()
		return errors.New("ядро не применило настройки, проверьте параметры узлов")
	}
	if err := writeAtomic(m.path, saved); err != nil {
		m.restoreCore()
		return errors.New("не удалось сохранить настройки на диск")
	}
	m.current = next
	m.revision++
	m.checks = nil
	m.checkedAt = time.Time{}
	m.selected, err = m.core.selected(ctx, groupName)
	var udpErr error
	m.selectedUDP, udpErr = m.core.selected(ctx, udpGroupName)
	m.coreError = err != nil || udpErr != nil
	notify(m.probeNow)
	return nil
}

func (m *manager) restoreCore() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	previous, err := coreConfig(m.current, m.tun, m.core.secret)
	if err == nil {
		err = m.core.reload(ctx, previous)
	}
	if err == nil {
		for group, id := range map[string]string{groupName: m.selected, udpGroupName: m.selectedUDP} {
			if id != "" {
				err = errors.Join(err, m.core.selectNode(ctx, group, id))
			}
		}
	}
	m.coreError = err != nil
}

func notify(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

func (m *manager) saveSettings(ctx context.Context, s settings) error {
	if err := normalizeSettings(&s); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	next := m.current
	next.Settings = s
	err := m.apply(ctx, next)
	if err == nil {
		notify(m.refreshNow)
	}
	return err
}

func (m *manager) importFeed(ctx context.Context, replacement *string) error {
	m.importMu.Lock()
	defer m.importMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	original := m.current.Subscription
	m.mu.Unlock()
	raw := original
	if replacement != nil {
		raw = *replacement
	}
	if replacement == nil && raw == "" {
		return errors.New("добавьте URL подписки")
	}
	var nodes []node
	var err error
	if raw != "" {
		nodes, err = fetchSubscription(ctx, m.subClient, raw)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err == nil {
		next := m.current
		next.Subscription, next.Nodes = raw, nodes
		before, _ := json.Marshal(m.current.Nodes)
		after, _ := json.Marshal(nodes)
		if raw != original || string(before) != string(after) {
			err = m.apply(ctx, next)
		}
	}
	if err != nil {
		if replacement == nil {
			m.refreshError = err.Error()
		}
		return err
	}
	m.refreshError = ""
	m.updatedAt = time.Time{}
	if raw != "" {
		m.updatedAt = time.Now()
	}
	if replacement != nil {
		notify(m.refreshNow)
	}
	return nil
}

func (m *manager) refreshLoop(ctx context.Context) {
	for ctx.Err() == nil {
		m.mu.Lock()
		hours := m.current.Settings.RefreshHours
		hasFeed := m.current.Subscription != ""
		m.mu.Unlock()
		var tick <-chan time.Time
		var timer *time.Timer
		if hours > 0 && hasFeed {
			timer = time.NewTimer(time.Duration(hours) * time.Hour)
			tick = timer.C
		}
		select {
		case <-ctx.Done():
		case <-m.refreshNow:
		case <-tick:
			_ = m.importFeed(ctx, nil)
		}
		if timer != nil {
			timer.Stop()
		}
	}
}

func bestNode(nodes []node, checks []nodeCheck, selected string, tolerance int, requireUDP bool) string {
	best, current := -1, -1
	for i, check := range checks {
		if check.Delay <= 0 || (requireUDP && !check.UDP) {
			continue
		}
		if best < 0 || check.Delay < checks[best].Delay {
			best = i
		}
		if nodes[i].ID == selected {
			current = i
		}
	}
	if best < 0 {
		return direct
	}
	if current >= 0 && checks[current].Delay-checks[best].Delay <= tolerance {
		return selected
	}
	return nodes[best].ID
}

func (c coreClient) updateSelection(ctx context.Context, s state, checks []nodeCheck, group string, requireUDP bool) (string, error) {
	selected, err := c.selected(ctx, group)
	if err != nil {
		return selected, err
	}
	target := bestNode(s.Nodes, checks, selected, s.Settings.Tolerance, requireUDP)
	if target != selected {
		if err := c.selectNode(ctx, group, target); err != nil {
			return selected, err
		}
	}
	return target, nil
}

func (m *manager) probe(ctx context.Context) time.Duration {
	m.mu.Lock()
	s, revision := m.current, m.revision
	m.mu.Unlock()
	results := make([]nodeCheck, len(s.Nodes))
	var probes sync.WaitGroup
	jobs := make(chan int, len(s.Nodes))
	for i := range s.Nodes {
		jobs <- i
	}
	close(jobs)
	for i := 0; i < min(8, len(s.Nodes)); i++ {
		probes.Go(func() {
			for index := range jobs {
				if ctx.Err() != nil {
					return
				}
				results[index] = m.core.probe(ctx, s.Nodes[index].ID, s.Settings)
			}
		})
	}
	probes.Wait()
	m.mu.Lock()
	defer m.mu.Unlock()
	if revision == m.revision && ctx.Err() == nil {
		selected, err := m.core.updateSelection(ctx, s, results, groupName, false)
		selectedUDP, udpErr := m.core.updateSelection(ctx, s, results, udpGroupName, true)
		m.selected, m.selectedUDP = selected, selectedUDP
		m.coreError = err != nil || udpErr != nil
		m.checks = results
		m.checkedAt = time.Now()
	}
	return time.Duration(s.Settings.ProbeInterval) * time.Second
}

func (m *manager) probeLoop(ctx context.Context) {
	for ctx.Err() == nil {
		timer := time.NewTimer(m.probe(ctx))
		select {
		case <-ctx.Done():
		case <-timer.C:
		case <-m.probeNow:
		}
		timer.Stop()
	}
}

func loadState(dir string) (state, string, error) {
	path := filepath.Join(dir, "state.json")
	s := defaultState()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, path, nil
	}
	if err != nil {
		return s, path, err
	}
	if err = json.Unmarshal(data, &s); err != nil {
		return s, path, err
	}
	return s, path, normalizeSettings(&s.Settings)
}
