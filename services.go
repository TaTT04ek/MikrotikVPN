package main

import (
	_ "embed"
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
}

func findService(id string) (servicePreset, bool) {
	i := slices.IndexFunc(servicePresets, func(p servicePreset) bool { return p.ID == id })
	if i < 0 {
		return servicePreset{}, false
	}
	return servicePresets[i], true
}
