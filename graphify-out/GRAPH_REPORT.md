# Graph Report - mikrot  (2026-11-22)

## Corpus Check
- 10 files · ~5,372 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 6 file(s) not represented in the graph (top: (none) 4, .rsc 1, .css 1)

## Summary
- 152 nodes · 331 edges · 9 communities (8 shown, 1 thin omitted)
- Extraction: 80% EXTRACTED · 20% INFERRED · 0% AMBIGUOUS · INFERRED: 65 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7a77e229`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Core configuration and dependencies
- Mihomo and manager controls
- TCP UDP and DNS routing
- Application and HTTP API
- config.go
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
- `Built-in service catalog exposed by GET /api/services` --references--> `webHandler()`  [INFERRED]
  docs/architecture.md → main.go
- `Management JSON API and local Mihomo API` --references--> `webHandler()`  [INFERRED]
  docs/architecture.md → main.go
- `Inline local Mihomo rule provider for exact domains and suffixes` --references--> `coreConfig()`  [EXTRACTED]
  docs/architecture.md → config.go
- `Independent TCP and UDP node selection with latency threshold` --references--> `bestNode()`  [EXTRACTED]
  docs/architecture.md → manager.go

## Import Cycles
- None detected.

## Communities (9 total, 1 thin omitted)

### Community 0 - "Core configuration and dependencies"
Cohesion: 0.12
Nodes (3): enableIPv4Forwarding(), nodeStatus, status

### Community 1 - "Mihomo and manager controls"
Cohesion: 0.22
Nodes (6): webHandler(), bestNode(), notify(), coreClient, manager, nodeCheck

### Community 2 - "TCP UDP and DNS routing"
Cohesion: 0.10
Nodes (27): Архитектура Mikrot VPN, RouterOS IPv4 client selection (mikrot-vpn-clients), DIRECT outbound and fail-open fallback for either transport, Optional Mihomo DNS redir-host on port 53 maps real IPs to domains, Per-node HTTP URL health and latency probe, Main routing table fallback when VETH is unavailable, Management JSON API and local Mihomo API, Mihomo ordered split-routing rules (+19 more)

### Community 4 - "Application and HTTP API"
Cohesion: 0.15
Nodes (6): env(), main(), readJSON(), result(), run(), writeJSON()

### Community 5 - "config.go"
Cohesion: 0.14
Nodes (19): coreConfig(), defaultState(), normalizeSettings(), validAction(), writeAtomic(), startCore(), Built-in service catalog exposed by GET /api/services, Inline local Mihomo rule provider for exact domains and suffixes (+11 more)

### Community 6 - "Web panel implementation"
Cohesion: 0.30
Nodes (12): addRule(), api(), el(), fields, loadSettings(), loadStatus(), mutate(), notice() (+4 more)

### Community 7 - "Regression tests"
Cohesion: 0.32
Nodes (3): TestBestNode(), TestParseSubscription(), TestServiceCatalogRules()

### Community 8 - "Subscription access"
Cohesion: 0.28
Nodes (6): node, fetchSubscription(), parseSubscription(), publicAddress(), remoteURL(), subscriptionClient()

## Knowledge Gaps
- **8 isolated node(s):** `mikrot`, `Persistent last-known-good /data/state.json`, `Configuration revision invalidates stale health results`, `UDP support reported by the Mihomo API`, `Service, domain, suffix, destination IP and source IP conditions` (+3 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 35 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `webHandler()` connect `Mihomo and manager controls` to `TCP UDP and DNS routing`, `Application and HTTP API`, `config.go`?**
  _High betweenness centrality (0.227) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `webHandler()` (e.g. with `Built-in service catalog exposed by GET /api/services` and `Management JSON API and local Mihomo API`) actually correct?**
  _`webHandler()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `mikrot`, `Persistent last-known-good /data/state.json`, `Configuration revision invalidates stale health results` to the rest of the system?**
  _8 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Core configuration and dependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.12307692307692308 - nodes in this community are weakly interconnected._
- **Why does `Add-service button and catalog service selector` connect `config.go` to `TCP UDP and DNS routing`, `Web panel implementation`?**
  _High betweenness centrality (0.150) - this node is a cross-community bridge._
- **Should `TCP UDP and DNS routing` be split into smaller, more focused modules?**
  _Cohesion score 0.09686609686609686 - nodes in this community are weakly interconnected._
- **Why does `Built-in service catalog exposed by GET /api/services` connect `config.go` to `Mihomo and manager controls`?**
  _High betweenness centrality (0.146) - this node is a cross-community bridge._