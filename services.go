package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

//go:embed services/youtube.txt
var youtubeDomains string

//go:embed services/github.txt
var githubDomains string

//go:embed services/discord.txt
var discordDomains string

type servicePreset struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Domains []string `json:"domains"`
	rules   []string
}

type serviceDefinition struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Domains []string `json:"domains"`
}

var serviceIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,47}$`)

func (d serviceDefinition) preset() servicePreset {
	p := servicePreset{ID: d.ID, Label: d.Label, Domains: d.Domains}
	for _, domain := range d.Domains {
		if suffix, ok := strings.CutPrefix(domain, "+."); ok {
			p.rules = append(p.rules, "DOMAIN-SUFFIX,"+suffix)
		} else {
			p.rules = append(p.rules, "DOMAIN,"+domain)
		}
	}
	return p
}

func newPreset(id, label, payload string) servicePreset {
	p := servicePreset{ID: id, Label: label}
	for _, line := range strings.Fields(payload) {
		kind, domain, _ := strings.Cut(line, ",")
		p.rules = append(p.rules, line)
		if kind == "DOMAIN-SUFFIX" {
			domain = "+." + domain
		}
		p.Domains = append(p.Domains, domain)
	}
	return p
}

var servicePresets = []servicePreset{
	newPreset("youtube", "YouTube", youtubeDomains),
	newPreset("github", "GitHub", githubDomains),
	newPreset("discord", "Discord", discordDomains),
	{ID: telegramService, Label: "Telegram", Domains: []string{}},
}

func findService(id string, custom ...serviceDefinition) (servicePreset, bool) {
	i := slices.IndexFunc(servicePresets, func(p servicePreset) bool { return p.ID == id })
	if i >= 0 {
		return servicePresets[i], true
	}
	for _, d := range custom {
		if d.ID == id {
			return d.preset(), true
		}
	}
	return servicePreset{}, false
}

func (m *manager) services() []servicePreset {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := slices.Clone(servicePresets)
	for _, d := range m.current.Services {
		result = append(result, d.preset())
	}
	return result
}

func (m *manager) importServices(ctx context.Context, definitions []serviceDefinition) error {
	if len(definitions) == 0 {
		return errors.New("добавьте сервисы в пакет")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(definitions) > 64 {
		return errors.New("можно сохранить до 64 пользовательских сервисов")
	}
	next := m.current
	next.Services = slices.Clone(next.Services)
	ids := make(map[string]bool, len(definitions))
	for i, d := range definitions {
		d.ID = strings.TrimSpace(d.ID)
		d.Label = strings.TrimSpace(d.Label)
		if !serviceIDPattern.MatchString(d.ID) || len([]rune(d.Label)) == 0 || len([]rune(d.Label)) > 80 ||
			len(d.Domains) == 0 || len(d.Domains) > 200 {
			return fmt.Errorf("проверьте идентификатор, название и список доменов сервиса №%d", i+1)
		}
		if _, builtin := findService(d.ID); builtin || ids[d.ID] {
			return fmt.Errorf("идентификатор %q уже используется", d.ID)
		}
		ids[d.ID] = true
		d.Domains = slices.Clone(d.Domains)
		seen := make(map[string]bool)
		for j, entry := range d.Domains {
			entry = strings.ToLower(strings.TrimSpace(entry))
			domain, _ := strings.CutPrefix(entry, "+.")
			if len(domain) > 253 || !domainPattern.MatchString(domain) || seen[entry] {
				return fmt.Errorf("проверьте домен №%d сервиса %q", j+1, d.ID)
			}
			seen[entry] = true
			d.Domains[j] = entry
		}
		index := slices.IndexFunc(m.current.Services, func(saved serviceDefinition) bool { return saved.ID == d.ID })
		if index >= 0 {
			saved := m.current.Services[index]
			if saved.Label != d.Label || !slices.Equal(saved.Domains, d.Domains) {
				return fmt.Errorf("сервис %q уже есть в каталоге с другим названием или доменами", d.ID)
			}
			continue
		}
		if len(next.Services) >= 64 {
			return errors.New("можно сохранить до 64 пользовательских сервисов")
		}
		next.Services = append(next.Services, d)
	}
	if len(next.Services) == len(m.current.Services) {
		return nil
	}
	payload, err := json.Marshal(next)
	if err != nil {
		return errors.New("не удалось подготовить пакет сервисов")
	}
	if err := writeAtomic(m.path, payload); err != nil {
		return errors.New("не удалось сохранить пакет сервисов")
	}
	m.current = next
	return nil
}
