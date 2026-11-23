# Graph Report - mikrot  (2026-11-23)

## Corpus Check
- Corpus is ~6,457 words - fits in a single context window. You may not need a graph.

## Summary
- 216 nodes · 404 edges · 12 communities (11 shown, 1 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 72 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Core process and networking
- Mihomo and manager controls
- Device compatibility and testing
- GitHub Actions image validation
- Configuration validation and state
- RouterOS installer and network
- Service catalog and presets
- Web panel implementation
- TCP UDP and DNS routing
- Application and HTTP API
- Community 10
- Go module

## God Nodes (most connected - your core abstractions)
1. `manager` - 16 edges
2. `webHandler()` - 13 edges
3. `state` - 10 edges
4. `coreClient` - 10 edges
5. `coreConfig()` - 9 edges
6. `startCore()` - 9 edges
7. `run()` - 9 edges
8. `Image build validation and on-device verification` - 9 edges
9. `with` - 8 edges
10. `normalizeSettings()` - 8 edges

## Surprising Connections (you probably didn't know these)
- `Optional Mihomo DNS redir-host on port 53 maps real IPs to domains` --references--> `coreConfig()`  [INFERRED]
  docs/architecture.md → config.go
- `Inline local Mihomo rule provider for exact domains and suffixes` --references--> `coreConfig()`  [EXTRACTED]
  docs/architecture.md → config.go
- `Built-in service catalog exposed by GET /api/services` --references--> `webHandler()`  [INFERRED]
  docs/architecture.md → main.go
- `Management JSON API and local Mihomo API` --references--> `webHandler()`  [INFERRED]
  docs/architecture.md → main.go
- `Built-in service catalog exposed by GET /api/services` --references--> `servicePreset`  [INFERRED]
  docs/architecture.md → services.go

## Import Cycles
- None detected.

## Communities (12 total, 1 thin omitted)

### Community 0 - "Core process and networking"
Cohesion: 0.08
Nodes (7): enableIPv4Forwarding(), env(), main(), run(), publicAddress(), remoteURL(), subscriptionClient()

### Community 1 - "Mihomo and manager controls"
Cohesion: 0.19
Nodes (12): coreConfig(), writeAtomic(), startCore(), bestNode(), notify(), coreClient, manager, node (+4 more)

### Community 2 - "Device compatibility and testing"
Cohesion: 0.08
Nodes (29): ARMv6 build without a tested device model, RouterOS CHR x86-64 AMD64 tested on RouterOS 7.24.6, CHR installation verified on local HTTPS server with redirects and custom subnet, Совместимость MikroTik VPN, On-device checklist for installation, reboot, subscription, TCP/UDP, fallback, and memory, hAP ac² ARMv7 candidate requiring startup and memory checks, hAP ac³ ARMv7 candidate with USB ext4, hAP ax³ ARM64 candidate (+21 more)

### Community 3 - "GitHub Actions image validation"
Cohesion: 0.08
Nodes (24): runs-on, steps, timeout-minutes, jobs, build, name, on, pull_request (+16 more)

### Community 4 - "Configuration validation and state"
Cohesion: 0.18
Nodes (10): defaultState(), normalizeSettings(), validAction(), loadState(), rule, settings, TestBestNode(), TestCoreConfigPreservesRuleOrder() (+2 more)

### Community 5 - "RouterOS installer and network"
Cohesion: 0.17
Nodes (15): Архитектура Mikrot VPN, RouterOS IPv4 client selection (mikrot-vpn-clients), Dedicated container IPv4 subnet 172.30.55.0/30, Installer-created VETH, persistent /data volume, environment profile, access rules and routing table, Main routing table fallback when VETH is unavailable, Interactive RouterOS installer, Post-start routing activation for chosen IPv4 clients, Mounted storage discovery and image archive selection (+7 more)

### Community 6 - "Service catalog and presets"
Cohesion: 0.19
Nodes (10): Built-in service catalog exposed by GET /api/services, Inline local Mihomo rule provider for exact domains and suffixes, Ordered SERVICE rule carrying the selected preset ID, servicePreset, findService(), newPreset(), Expandable service domain preview, Service, domain, suffix, destination IP and source IP conditions (+2 more)

### Community 7 - "Web panel implementation"
Cohesion: 0.30
Nodes (12): addRule(), api(), el(), fields, loadSettings(), loadStatus(), mutate(), notice() (+4 more)

### Community 8 - "TCP UDP and DNS routing"
Cohesion: 0.24
Nodes (12): DIRECT outbound and fail-open fallback for either transport, Optional Mihomo DNS redir-host on port 53 maps real IPs to domains, Per-node HTTP URL health and latency probe, Mihomo ordered split-routing rules, Independent TCP and UDP node selection with latency threshold, Configuration revision invalidates stale health results, UDP support reported by the Mihomo API, MIKROT-VPN-UDP UDP outbound selector group (+4 more)

### Community 9 - "Application and HTTP API"
Cohesion: 0.25
Nodes (6): readJSON(), result(), webHandler(), writeJSON(), nodeStatus, status

### Community 10 - "Community 10"
Cohesion: 0.29
Nodes (8): Management JSON API and local Mihomo API, Persistent last-known-good /data/state.json, Status fields omit subscription URL and node credentials, Subscription fetch, parsing and applying Mihomo config, Active TCP and UDP exits, health and on-demand checks, Server list and count, Subscription controls with masked URL, load, refresh and remove, Mikrot VPN web panel

## Knowledge Gaps
- **20 isolated node(s):** `mikrot`, `Test the web panel, first VPN rule and direct route from the same client`, `RouterOS fetch with DNS and verified TLS`, `hAP ac³ ARMv7 candidate with USB ext4`, `hAP ac² ARMv7 candidate requiring startup and memory checks` (+15 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 64 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Mihomo ordered split-routing rules` connect `TCP UDP and DNS routing` to `RouterOS installer and network`, `Service catalog and presets`?**
  _High betweenness centrality (0.288) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `webHandler()` (e.g. with `Built-in service catalog exposed by GET /api/services` and `Management JSON API and local Mihomo API`) actually correct?**
  _`webHandler()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `mikrot`, `Test the web panel, first VPN rule and direct route from the same client`, `RouterOS fetch with DNS and verified TLS` to the rest of the system?**
  _20 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Core process and networking` be split into smaller, more focused modules?**
  _Cohesion score 0.08170731707317073 - nodes in this community are weakly interconnected._
- **Why does `TUN IPv4 TCP and UDP traffic intake` connect `RouterOS installer and network` to `TCP UDP and DNS routing`?**
  _High betweenness centrality (0.273) - this node is a cross-community bridge._
- **Should `Device compatibility and testing` be split into smaller, more focused modules?**
  _Cohesion score 0.0812807881773399 - nodes in this community are weakly interconnected._
- **Should `GitHub Actions image validation` be split into smaller, more focused modules?**
  _Cohesion score 0.08 - nodes in this community are weakly interconnected._
Graphify does not receive token usage from this Codex subagent interface.
