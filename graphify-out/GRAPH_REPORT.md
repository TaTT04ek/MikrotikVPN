# Graph Report - mikrot  (2026-11-23)

## Corpus Check
- Corpus is ~7,471 words - fits in a single context window. You may not need a graph.

## Summary
- 243 nodes · 455 edges · 11 communities (9 shown, 2 thin omitted)
- Extraction: 81% EXTRACTED · 19% INFERRED · 0% AMBIGUOUS · INFERRED: 87 edges (avg confidence: 0.9)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- RouterOS installer and network
- Core process and networking
- Mihomo and manager controls
- GitHub Actions image validation
- Configuration validation and state
- TCP UDP and DNS routing
- Application and HTTP API
- Device compatibility and testing
- Web panel implementation
- Service catalog and presets
- Go module

## God Nodes (most connected - your core abstractions)
1. `manager` - 16 edges
2. `webHandler()` - 13 edges
3. `MikroTik VPN` - 13 edges
4. `state` - 10 edges
5. `coreClient` - 10 edges
6. `coreConfig()` - 9 edges
7. `startCore()` - 9 edges
8. `run()` - 9 edges
9. `Image build validation and on-device verification` - 9 edges
10. `Independent TCP and UDP node selection with latency threshold` - 9 edges

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

### Community 0 - "RouterOS installer and network"
Cohesion: 0.05
Nodes (53): Архитектура Mikrot VPN, RouterOS IPv4 client selection (mikrot-vpn-clients), Dedicated container IPv4 subnet 172.30.55.0/30, Installer-created VETH, persistent /data volume, environment profile, access rules and routing table, Main routing table fallback when VETH is unavailable, Persistent last-known-good /data/state.json, Interactive RouterOS installer, Post-start routing activation for chosen IPv4 clients (+45 more)

### Community 2 - "Mihomo and manager controls"
Cohesion: 0.17
Nodes (11): coreConfig(), writeAtomic(), startCore(), notify(), coreClient, manager, nodeCheck, nodeStatus (+3 more)

### Community 3 - "GitHub Actions image validation"
Cohesion: 0.08
Nodes (24): runs-on, steps, timeout-minutes, jobs, build, name, on, pull_request (+16 more)

### Community 4 - "Configuration validation and state"
Cohesion: 0.11
Nodes (18): defaultState(), normalizeSettings(), validAction(), env(), main(), run(), bestNode(), loadState() (+10 more)

### Community 5 - "TCP UDP and DNS routing"
Cohesion: 0.14
Nodes (21): DIRECT outbound and fail-open fallback for either transport, Optional Mihomo DNS redir-host on port 53 maps real IPs to domains, Per-node HTTP URL health and latency probe, Mihomo ordered split-routing rules, Independent TCP and UDP node selection with latency threshold, Configuration revision invalidates stale health results, UDP support reported by the Mihomo API, MIKROT-VPN-UDP UDP outbound selector group (+13 more)

### Community 6 - "Application and HTTP API"
Cohesion: 0.20
Nodes (11): Management JSON API and local Mihomo API, Status fields omit subscription URL and node credentials, Subscription fetch, parsing and applying Mihomo config, readJSON(), result(), webHandler(), writeJSON(), Active TCP and UDP exits, health and on-demand checks (+3 more)

### Community 7 - "Device compatibility and testing"
Cohesion: 0.16
Nodes (14): ARMv6 build without a tested device model, RouterOS CHR x86-64 AMD64 tested on RouterOS 7.24.6, CHR installation verified on local HTTPS server with redirects and custom subnet, Совместимость MikroTik VPN, On-device checklist for installation, reboot, subscription, TCP/UDP, fallback, and memory, hAP ac² ARMv7 candidate requiring startup and memory checks, hAP ac³ ARMv7 candidate with USB ext4, hAP ax³ ARM64 candidate (+6 more)

### Community 8 - "Web panel implementation"
Cohesion: 0.30
Nodes (12): addRule(), api(), el(), fields, loadSettings(), loadStatus(), mutate(), notice() (+4 more)

### Community 9 - "Service catalog and presets"
Cohesion: 0.31
Nodes (6): Built-in service catalog exposed by GET /api/services, Inline local Mihomo rule provider for exact domains and suffixes, Ordered SERVICE rule carrying the selected preset ID, servicePreset, findService(), newPreset()

## Knowledge Gaps
- **23 isolated node(s):** `mikrot`, `RouterOS 7 container and Mihomo IPv4 split routing`, `Enable the installer mangle rules after adding the first client`, `Inspect container state and container logs if the panel fails to open`, `Check URL reachability and format; retain the last valid list on refresh failure` (+18 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 67 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MikroTik VPN` connect `RouterOS installer and network` to `Device compatibility and testing`?**
  _High betweenness centrality (0.265) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `webHandler()` (e.g. with `Built-in service catalog exposed by GET /api/services` and `Management JSON API and local Mihomo API`) actually correct?**
  _`webHandler()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `mikrot`, `RouterOS 7 container and Mihomo IPv4 split routing`, `Enable the installer mangle rules after adding the first client` to the rest of the system?**
  _23 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `RouterOS installer and network` be split into smaller, more focused modules?**
  _Cohesion score 0.05370101596516691 - nodes in this community are weakly interconnected._
- **Why does `Subscription fetch, parsing and applying Mihomo config` connect `Application and HTTP API` to `RouterOS installer and network`, `Mihomo and manager controls`?**
  _High betweenness centrality (0.197) - this node is a cross-community bridge._
- **Should `Core process and networking` be split into smaller, more focused modules?**
  _Cohesion score 0.09957325746799431 - nodes in this community are weakly interconnected._
- **Why does `Persistent last-known-good /data/state.json` connect `RouterOS installer and network` to `Application and HTTP API`?**
  _High betweenness centrality (0.186) - this node is a cross-community bridge._
Graphify does not receive token usage from this Codex subagent interface.
