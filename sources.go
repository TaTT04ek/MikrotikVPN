package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	legacySourceID = "legacy"
	maxSources     = 64
	maxSourceSize  = 8192
)

type nodeSource struct {
	ID        string
	URL       string `json:",omitempty"`
	Nodes     []node
	UpdatedAt time.Time
}

type sourceStatus struct {
	ID        string    `json:"id"`
	Remote    bool      `json:"remote"`
	Label     string    `json:"label"`
	NodeCount int       `json:"nodeCount"`
	UpdatedAt time.Time `json:"updatedAt"`
	Error     string    `json:"error,omitempty"`
}

func (m *manager) sourceStatuses() ([]sourceStatus, time.Time, string, bool) {
	result := make([]sourceStatus, 0, len(m.current.Sources))
	var lastUpdate time.Time
	var refreshError string
	var subscribed bool
	for _, source := range m.current.Sources {
		item := sourceStatus{ID: source.ID, Remote: source.URL != "", NodeCount: len(source.Nodes),
			UpdatedAt: source.UpdatedAt, Error: m.sourceErrors[source.ID]}
		if len(source.Nodes) > 0 {
			item.Label = source.Nodes[0].Label
		}
		if item.Remote {
			subscribed = true
			if item.UpdatedAt.After(lastUpdate) {
				lastUpdate = item.UpdatedAt
			}
			if refreshError == "" {
				refreshError = item.Error
			}
		}
		result = append(result, item)
	}
	return result, lastUpdate, refreshError, subscribed
}

func mergeSourceNodes(sources []nodeSource) ([]node, error) {
	var nodes []node
	seen := make(map[string]bool)
	for _, source := range sources {
		for _, n := range source.Nodes {
			if seen[n.ID] {
				continue
			}
			if len(nodes) == maxNodes {
				return nil, fmt.Errorf("все источники вместе могут содержать до %d уникальных узлов", maxNodes)
			}
			seen[n.ID] = true
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func sourceIndex(sources []nodeSource, id string) int {
	return slices.IndexFunc(sources, func(source nodeSource) bool { return source.ID == id })
}

// applySources is called with mu held, after any network request has completed.
func (m *manager) applySources(ctx context.Context, sources []nodeSource) error {
	nodes, err := mergeSourceNodes(sources)
	if err != nil {
		return err
	}
	next := m.current
	next.Sources, next.Nodes = sources, nodes
	if !slices.EqualFunc(nodes, m.current.Nodes, func(a, b node) bool {
		return a.ID == b.ID && a.Label == b.Label
	}) {
		return m.apply(ctx, next)
	}
	payload, err := json.Marshal(next)
	if err != nil {
		return errors.New("не удалось подготовить источники")
	}
	if err = writeAtomic(m.path, payload); err != nil {
		return errors.New("не удалось сохранить источники на диск")
	}
	m.current = next
	return nil
}

func (m *manager) addSource(ctx context.Context, id, input string) error {
	m.importMu.Lock()
	defer m.importMu.Unlock()
	input = strings.TrimSpace(input)
	if input == "" || len(input) > maxSourceSize || len(strings.Fields(input)) != 1 {
		return errors.New("вставьте одну ссылку подписки или один ключ подключения")
	}
	scheme, _, _ := strings.Cut(input, "://")
	scheme = strings.ToLower(scheme)
	urlSource := scheme == "https" || scheme == "http"
	if id == legacySourceID && !urlSource {
		return errors.New("укажите ссылку HTTP или HTTPS")
	}
	m.mu.Lock()
	sources := slices.Clone(m.current.Sources)
	m.mu.Unlock()
	index := sourceIndex(sources, id)
	if index < 0 && len(sources) >= maxSources {
		return fmt.Errorf("можно добавить до %d источников", maxSources)
	}

	source := nodeSource{ID: id}
	if source.ID == "" {
		source.ID = rand.Text()
	}
	var err error
	if urlSource {
		for _, existing := range sources {
			if existing.ID != id && existing.URL == input {
				return errors.New("эта подписка уже добавлена")
			}
		}
		source.URL = input
		source.Nodes, err = fetchSubscription(ctx, m.subClient, input)
	} else {
		switch scheme {
		case "vless", "vmess", "trojan", "ss", "hysteria2", "hy2", "tuic":
			source.Nodes, err = parseSubscription([]byte(input))
		default:
			return errors.New("формат ключа подключения не поддерживается")
		}
		if err != nil || len(source.Nodes) != 1 {
			return errors.New("не удалось прочитать ключ подключения")
		}
		for _, existing := range sources {
			if existing.URL == "" && len(existing.Nodes) == 1 && existing.Nodes[0].ID == source.Nodes[0].ID {
				return errors.New("этот ключ уже добавлен")
			}
		}
	}
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	source.UpdatedAt = time.Now()
	if index < 0 {
		sources = append(sources, source)
	} else {
		sources[index] = source
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err = m.applySources(ctx, sources); err != nil {
		return err
	}
	delete(m.sourceErrors, source.ID)
	notify(m.refreshNow)
	return nil
}

func (m *manager) deleteSource(ctx context.Context, id string) error {
	m.importMu.Lock()
	defer m.importMu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	index := sourceIndex(m.current.Sources, id)
	if index < 0 {
		if id == legacySourceID {
			return nil
		}
		return errors.New("источник не найден")
	}
	sources := slices.Delete(slices.Clone(m.current.Sources), index, index+1)
	if err := m.applySources(ctx, sources); err != nil {
		return err
	}
	delete(m.sourceErrors, id)
	notify(m.refreshNow)
	return nil
}

func (m *manager) refreshSource(ctx context.Context, id string) error {
	m.importMu.Lock()
	defer m.importMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	sources := slices.Clone(m.current.Sources)
	m.mu.Unlock()
	index := sourceIndex(sources, id)
	if index < 0 || sources[index].URL == "" {
		return errors.New("выберите URL-подписку для обновления")
	}
	nodes, err := fetchSubscription(ctx, m.subClient, sources[index].URL)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err == nil {
		sources[index].Nodes, sources[index].UpdatedAt = nodes, time.Now()
		err = m.applySources(ctx, sources)
	}
	if err != nil {
		if m.sourceErrors == nil {
			m.sourceErrors = make(map[string]string)
		}
		m.sourceErrors[id] = err.Error()
		return err
	}
	delete(m.sourceErrors, id)
	return nil
}

func (m *manager) refreshLoop(ctx context.Context) {
	for ctx.Err() == nil {
		m.mu.Lock()
		hours := m.current.Settings.RefreshHours
		var ids []string
		for _, source := range m.current.Sources {
			if source.URL != "" {
				ids = append(ids, source.ID)
			}
		}
		m.mu.Unlock()
		var tick <-chan time.Time
		var timer *time.Timer
		if hours > 0 && len(ids) > 0 {
			timer = time.NewTimer(time.Duration(hours) * time.Hour)
			tick = timer.C
		}
		select {
		case <-ctx.Done():
		case <-m.refreshNow:
		case <-tick:
			for _, id := range ids {
				if ctx.Err() != nil {
					break
				}
				_ = m.refreshSource(ctx, id)
			}
		}
		if timer != nil {
			timer.Stop()
		}
	}
}
