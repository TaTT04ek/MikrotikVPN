# Graph Report - mikrot  (2026-11-23)

## Corpus Check
- Corpus is ~7,397 words - fits in a single context window. You may not need a graph.

## Summary
- 242 nodes · 450 edges · 11 communities (9 shown, 2 thin omitted)
- Extraction: 81% EXTRACTED · 19% INFERRED · 0% AMBIGUOUS · INFERRED: 86 edges (avg confidence: 0.9)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Core process and networking
- RouterOS installer and network
- TCP UDP and DNS routing
- Mihomo and manager controls
- Device compatibility and testing
- GitHub Actions image validation
- Configuration validation and state
- Web panel implementation
- Application and HTTP API
- Community 9
- Go module

## God Nodes (most connected - your core abstractions)
1. `manager` - 16 edges
2. `webHandler()` - 13 edges
3. `state` - 10 edges
4. `coreClient` - 10 edges
5. `MikroTik VPN` - 10 edges
6. `coreConfig()` - 9 edges
7. `startCore()` - 9 edges
8. `run()` - 9 edges
9. `Independent TCP and UDP node selection with latency threshold` - 9 edges
10. `Image build validation and on-device verification` - 9 edges

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

## Communities (11 total, 2 thin omitted)

### Community 1 - "RouterOS installer and network"
Cohesion: 0.08
Nodes (35): Архитектура Mikrot VPN, RouterOS IPv4 client selection (mikrot-vpn-clients), Dedicated container IPv4 subnet 172.30.55.0/30, Installer-created VETH, persistent /data volume, environment profile, access rules and routing table, Main routing table fallback when VETH is unavailable, Interactive RouterOS installer, Post-start routing activation for chosen IPv4 clients, Mounted storage discovery and image archive selection (+27 more)

### Community 2 - "TCP UDP and DNS routing"
Cohesion: 0.09
Nodes (28): Built-in service catalog exposed by GET /api/services, DIRECT outbound and fail-open fallback for either transport, Optional Mihomo DNS redir-host on port 53 maps real IPs to domains, Per-node HTTP URL health and latency probe, Inline local Mihomo rule provider for exact domains and suffixes, Mihomo ordered split-routing rules, Independent TCP and UDP node selection with latency threshold, Ordered SERVICE rule carrying the selected preset ID (+20 more)

### Community 3 - "Mihomo and manager controls"
Cohesion: 0.17
Nodes (11): coreConfig(), writeAtomic(), startCore(), notify(), coreClient, manager, nodeCheck, nodeStatus (+3 more)

### Community 4 - "Device compatibility and testing"
Cohesion: 0.08
Nodes (28): ARMv6 build without a tested device model, RouterOS CHR x86-64 AMD64 tested on RouterOS 7.24.6, CHR installation verified on local HTTPS server with redirects and custom subnet, Совместимость MikroTik VPN, On-device checklist for installation, reboot, subscription, TCP/UDP, fallback, and memory, hAP ac² ARMv7 candidate requiring startup and memory checks, hAP ac³ ARMv7 candidate with USB ext4, hAP ax³ ARM64 candidate (+20 more)

### Community 5 - "GitHub Actions image validation"
Cohesion: 0.08
Nodes (24): runs-on, steps, timeout-minutes, jobs, build, name, on, pull_request (+16 more)

### Community 6 - "Configuration validation and state"
Cohesion: 0.16
Nodes (13): defaultState(), normalizeSettings(), validAction(), bestNode(), loadState(), node, TestBestNode(), TestCoreConfigPreservesRuleOrder() (+5 more)

### Community 7 - "Web panel implementation"
Cohesion: 0.30
Nodes (12): addRule(), api(), el(), fields, loadSettings(), loadStatus(), mutate(), notice() (+4 more)

### Community 8 - "Application and HTTP API"
Cohesion: 0.21
Nodes (9): env(), main(), readJSON(), result(), run(), webHandler(), writeJSON(), publicAddress() (+1 more)

### Community 9 - "Community 9"
Cohesion: 0.25
Nodes (9): Management JSON API and local Mihomo API, Persistent last-known-good /data/state.json, Status fields omit subscription URL and node credentials, Subscription fetch, parsing and applying Mihomo config, Protect subscription and node data in the mounted state.json, Active TCP and UDP exits, health and on-demand checks, Server list and count, Subscription controls with masked URL, load, refresh and remove (+1 more)

## Knowledge Gaps
- **23 isolated node(s):** `mikrot`, `Main routing table fallback when VETH is unavailable`, `Test the web panel, first VPN rule and direct route from the same client`, `RouterOS fetch with DNS and verified TLS`, `GitHub Actions runs Go tests, go vet and JavaScript checks` (+18 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 67 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MikroTik VPN` connect `RouterOS installer and network` to `Community 9`, `Device compatibility and testing`?**
  _High betweenness centrality (0.254) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `webHandler()` (e.g. with `Built-in service catalog exposed by GET /api/services` and `Management JSON API and local Mihomo API`) actually correct?**
  _`webHandler()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `mikrot`, `Main routing table fallback when VETH is unavailable`, `Test the web panel, first VPN rule and direct route from the same client` to the rest of the system?**
  _23 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Core process and networking` be split into smaller, more focused modules?**
  _Cohesion score 0.09957325746799431 - nodes in this community are weakly interconnected._
- **Why does `Subscription fetch, parsing and applying Mihomo config` connect `Community 9` to `Mihomo and manager controls`?**
  _High betweenness centrality (0.194) - this node is a cross-community bridge._
- **Should `RouterOS installer and network` be split into smaller, more focused modules?**
  _Cohesion score 0.08403361344537816 - nodes in this community are weakly interconnected._
- **Should `TCP UDP and DNS routing` be split into smaller, more focused modules?**
  _Cohesion score 0.09247311827956989 - nodes in this community are weakly interconnected._
Graphify does not receive token usage from this Codex subagent interface.
