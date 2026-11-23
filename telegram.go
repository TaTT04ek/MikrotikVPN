package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"time"
)

const telegramService = "telegram"
const telegramURL = "https://core.telegram.org/resources/cidr.txt"
const telegramInterval = 3 * time.Hour
const maxTelegramList = 64 << 10

//go:embed services/telegram.txt
var telegramBootstrap string

var telegramSeed = strings.Fields(telegramBootstrap)

type telegramStatus struct {
	CIDRs     []string  `json:"cidrs"`
	UpdatedAt time.Time `json:"updatedAt"`
	Error     string    `json:"error,omitempty"`
}

func telegramCIDRs(s state) []string {
	if len(s.TelegramCIDRs) == 0 {
		return telegramSeed
	}
	return s.TelegramCIDRs
}

func parseTelegramCIDRs(body string) ([]string, error) {
	var list []string
	for _, item := range strings.Fields(body) {
		prefix, err := netip.ParsePrefix(item)
		if err != nil || prefix != prefix.Masked() {
			return nil, errors.New("Telegram вернул некорректный список CIDR")
		}
		if prefix.Addr().Is4() {
			if prefix.Bits() < 8 || !publicAddress(prefix.Addr()) {
				return nil, errors.New("Telegram вернул недопустимый IPv4-диапазон")
			}
			list = append(list, prefix.String())
		}
	}
	if len(list) == 0 {
		return nil, errors.New("в списке Telegram нет диапазонов IPv4")
	}
	slices.Sort(list)
	return slices.Compact(list), nil
}

func fetchTelegramCIDRs(ctx context.Context, client *http.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, telegramURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("не удалось скачать список CIDR Telegram")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервер списка Telegram ответил HTTP %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxTelegramList+1))
	if err != nil || len(body) > maxTelegramList {
		return nil, errors.New("не удалось прочитать список CIDR Telegram")
	}
	return parseTelegramCIDRs(string(body))
}

func (m *manager) refreshTelegram(ctx context.Context) error {
	if !m.telegramMu.TryLock() {
		return errors.New("обновление списка Telegram уже выполняется")
	}
	defer m.telegramMu.Unlock()
	list, err := fetchTelegramCIDRs(ctx, m.subClient)
	m.mu.Lock()
	defer m.mu.Unlock()
	if err == nil {
		next := m.current
		next.TelegramCIDRs, next.TelegramUpdatedAt = list, time.Now().UTC()
		used := slices.ContainsFunc(next.Settings.Rules, func(r rule) bool {
			return r.Match == "SERVICE" && r.Value == telegramService
		})
		if used && !slices.Equal(telegramCIDRs(m.current), list) {
			err = m.apply(ctx, next)
		} else {
			var payload []byte
			if payload, err = json.Marshal(next); err == nil {
				err = writeAtomic(m.path, payload)
			}
			if err == nil {
				m.current = next
			} else {
				err = errors.New("не удалось сохранить список CIDR Telegram")
			}
		}
	}
	m.telegramError = ""
	if err != nil {
		m.telegramError = err.Error()
	}
	return err
}

func (m *manager) telegramLoop(ctx context.Context) {
	m.mu.Lock()
	last := m.current.TelegramUpdatedAt
	m.mu.Unlock()
	wait := telegramInterval - time.Since(last)
	if wait < 0 || wait > telegramInterval {
		wait = 0
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			_ = m.refreshTelegram(ctx)
			timer.Reset(telegramInterval)
		}
	}
}
