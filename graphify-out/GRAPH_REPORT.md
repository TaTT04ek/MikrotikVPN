# Graph Report - mikrot  (2026-11-22)

## Corpus Check
- Corpus is ~5,411 words - fits in a single context window. You may not need a graph.

## Summary
- 157 nodes · 338 edges · 10 communities (9 shown, 1 thin omitted)
- Extraction: 81% EXTRACTED · 19% INFERRED · 0% AMBIGUOUS · INFERRED: 65 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Core process and network dependencies
- Mihomo and manager controls
- TCP UDP and DNS routing
- Application and HTTP API
- Web panel implementation
- Subscription and regression tests
- Service catalog and presets
- Configuration validation and state
- RouterOS installation and routing
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

### Community 0 - "Core process and network dependencies"
Cohesion: 0.11
Nodes (6): enableIPv4Forwarding(), nodeStatus, status, publicAddress(), remoteURL(), subscriptionClient()

### Community 1 - "Mihomo and manager controls"
Cohesion: 0.21
Nodes (9): coreConfig(), writeAtomic(), startCore(), notify(), coreClient, manager, nodeCheck, state (+1 more)

### Community 2 - "TCP UDP and DNS routing"
Cohesion: 0.12
Nodes (23): DIRECT outbound and fail-open fallback for either transport, Optional Mihomo DNS redir-host on port 53 maps real IPs to domains, Per-node HTTP URL health and latency probe, Management JSON API and local Mihomo API, Mihomo ordered split-routing rules, Independent TCP and UDP node selection with latency threshold, Persistent last-known-good /data/state.json, Status fields omit subscription URL and node credentials (+15 more)

### Community 3 - "Application and HTTP API"
Cohesion: 0.13
Nodes (7): env(), main(), readJSON(), result(), run(), webHandler(), writeJSON()

### Community 4 - "Web panel implementation"
Cohesion: 0.30
Nodes (12): addRule(), api(), el(), fields, loadSettings(), loadStatus(), mutate(), notice() (+4 more)

### Community 5 - "Subscription and regression tests"
Cohesion: 0.21
Nodes (7): bestNode(), node, TestBestNode(), TestCoreConfigPreservesRuleOrder(), TestParseSubscription(), TestServiceCatalogRules(), parseSubscription()

### Community 6 - "Service catalog and presets"
Cohesion: 0.25
Nodes (8): Built-in service catalog exposed by GET /api/services, Inline local Mihomo rule provider for exact domains and suffixes, Ordered SERVICE rule carrying the selected preset ID, servicePreset, findService(), newPreset(), Expandable service domain preview, Add-service button and catalog service selector

### Community 7 - "Configuration validation and state"
Cohesion: 0.33
Nodes (6): defaultState(), normalizeSettings(), validAction(), loadState(), rule, settings

### Community 8 - "RouterOS installation and routing"
Cohesion: 0.28
Nodes (9): Архитектура Mikrot VPN, RouterOS IPv4 client selection (mikrot-vpn-clients), Dedicated container IPv4 subnet 172.30.55.0/30, Installer-created VETH, persistent /data volume, environment profile, access rules and routing table, Main routing table fallback when VETH is unavailable, Interactive RouterOS installer, Post-start routing activation for chosen IPv4 clients, Mounted storage discovery and image archive selection (+1 more)

## Knowledge Gaps
- **9 isolated node(s):** `mikrot`, `UDP support reported by the Mihomo API`, `Configuration revision invalidates stale health results`, `Persistent last-known-good /data/state.json`, `Main routing table fallback when VETH is unavailable` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 36 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `webHandler()` connect `Application and HTTP API` to `Mihomo and manager controls`, `TCP UDP and DNS routing`, `Service catalog and presets`?**
  _High betweenness centrality (0.216) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `webHandler()` (e.g. with `Built-in service catalog exposed by GET /api/services` and `Management JSON API and local Mihomo API`) actually correct?**
  _`webHandler()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `mikrot`, `UDP support reported by the Mihomo API`, `Configuration revision invalidates stale health results` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Core process and network dependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.11083743842364532 - nodes in this community are weakly interconnected._
- **Why does `Add-service button and catalog service selector` connect `Service catalog and presets` to `TCP UDP and DNS routing`, `Web panel implementation`?**
  _High betweenness centrality (0.150) - this node is a cross-community bridge._
- **Should `TCP UDP and DNS routing` be split into smaller, more focused modules?**
  _Cohesion score 0.11857707509881422 - nodes in this community are weakly interconnected._
- **Why does `Built-in service catalog exposed by GET /api/services` connect `Service catalog and presets` to `Application and HTTP API`?**
  _High betweenness centrality (0.141) - this node is a cross-community bridge._
Graphify does not receive token usage from this Codex subagent interface.
