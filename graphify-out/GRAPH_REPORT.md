# Graph Report - mikrot  (2026-11-22)

## Corpus Check
- Corpus is ~6,344 words - fits in a single context window. You may not need a graph.

## Summary
- 171 nodes · 353 edges · 10 communities (9 shown, 1 thin omitted)
- Extraction: 80% EXTRACTED · 20% INFERRED · 0% AMBIGUOUS · INFERRED: 71 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Core configuration and dependencies
- Mihomo and manager controls
- TCP UDP and DNS routing
- RouterOS deployment
- Application and HTTP API
- Service catalog and presets
- Web panel implementation
- Regression tests
- Subscription access
- Go module

## God Nodes (most connected - your core abstractions)
1. `manager` - 16 edges
2. `webHandler()` - 13 edges
3. `state` - 10 edges
4. `coreClient` - 10 edges
5. `coreConfig()` - 9 edges
6. `startCore()` - 9 edges
7. `run()` - 9 edges
8. `normalizeSettings()` - 8 edges
9. `Independent TCP and UDP node selection with latency threshold` - 8 edges
10. `settings` - 7 edges

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

## Communities (10 total, 1 thin omitted)

### Community 0 - "Core configuration and dependencies"
Cohesion: 0.11
Nodes (9): defaultState(), normalizeSettings(), validAction(), loadState(), nodeStatus, rule, settings, state (+1 more)

### Community 1 - "Mihomo and manager controls"
Cohesion: 0.18
Nodes (10): coreConfig(), writeAtomic(), enableIPv4Forwarding(), startCore(), webHandler(), bestNode(), notify(), coreClient (+2 more)

### Community 2 - "TCP UDP and DNS routing"
Cohesion: 0.11
Nodes (25): DIRECT outbound and fail-open fallback for either transport, Optional Mihomo DNS redir-host on port 53 maps real IPs to domains, Per-node HTTP URL health and latency probe, Management JSON API and local Mihomo API, Mihomo ordered split-routing rules, Independent TCP and UDP node selection with latency threshold, Persistent last-known-good /data/state.json, Status fields omit subscription URL and node credentials (+17 more)

### Community 3 - "RouterOS deployment"
Cohesion: 0.12
Nodes (18): Архитектура Mikrot VPN, RouterOS IPv4 client selection (mikrot-vpn-clients), Main routing table fallback when VETH is unavailable, TUN IPv4 TCP and UDP traffic intake, VETH gateway and policy table with ping monitoring, Web panel in a trusted LAN, Docker Buildx for ARMv5, ARMv6, ARMv7 and ARM64, FastTrack exclusions for both directions (+10 more)

### Community 4 - "Application and HTTP API"
Cohesion: 0.15
Nodes (6): env(), main(), readJSON(), result(), run(), writeJSON()

### Community 5 - "Service catalog and presets"
Cohesion: 0.18
Nodes (11): Built-in service catalog exposed by GET /api/services, Inline local Mihomo rule provider for exact domains and suffixes, Ordered SERVICE rule carrying the selected preset ID, servicePreset, Bundled V2Fly domain-list-community snapshot, YouTube, GitHub and Discord routing presets, YouTube website, video CDN, images, APIs and embedded player domains, findService() (+3 more)

### Community 6 - "Web panel implementation"
Cohesion: 0.30
Nodes (12): addRule(), api(), el(), fields, loadSettings(), loadStatus(), mutate(), notice() (+4 more)

### Community 7 - "Regression tests"
Cohesion: 0.31
Nodes (4): TestBestNode(), TestCoreConfigPreservesRuleOrder(), TestParseSubscription(), TestServiceCatalogRules()

### Community 8 - "Subscription access"
Cohesion: 0.28
Nodes (6): node, fetchSubscription(), parseSubscription(), publicAddress(), remoteURL(), subscriptionClient()

## Knowledge Gaps
- **17 isolated node(s):** `mikrot`, `Pinned Mihomo v1.19.32 release`, `Web panel in a trusted LAN`, `VLESS/REALITY, VMess, Trojan, Shadowsocks, Hysteria2 and TUIC`, `FastTrack exclusions for both directions` (+12 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 44 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Mihomo ordered split-routing rules` connect `TCP UDP and DNS routing` to `RouterOS deployment`?**
  _High betweenness centrality (0.211) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `webHandler()` (e.g. with `Built-in service catalog exposed by GET /api/services` and `Management JSON API and local Mihomo API`) actually correct?**
  _`webHandler()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `mikrot`, `Pinned Mihomo v1.19.32 release`, `Web panel in a trusted LAN` to the rest of the system?**
  _17 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Core configuration and dependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.10873440285204991 - nodes in this community are weakly interconnected._
- **Why does `webHandler()` connect `Mihomo and manager controls` to `TCP UDP and DNS routing`, `Application and HTTP API`, `Service catalog and presets`?**
  _High betweenness centrality (0.197) - this node is a cross-community bridge._
- **Should `TCP UDP and DNS routing` be split into smaller, more focused modules?**
  _Cohesion score 0.10666666666666667 - nodes in this community are weakly interconnected._
- **Why does `TUN IPv4 TCP and UDP traffic intake` connect `RouterOS deployment` to `TCP UDP and DNS routing`?**
  _High betweenness centrality (0.180) - this node is a cross-community bridge._
Graphify does not receive token usage from this Codex subagent interface.
