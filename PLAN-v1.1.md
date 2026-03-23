# Rift v1.1 — Summary, Policy, Secrets, Layer Attribution

## Context

Rift v1.0 ships file-level diffs, security analysis, package awareness, and multiple output formats. But user research (Reddit, HN, container-diff GitHub issues) reveals 4 gaps:

1. **Output is too noisy** — users want a one-screen verdict, not 200 lines
2. **No policy enforcement** — CI needs pass/fail rules, not just diffs
3. **No secrets detection** — GitGuardian found 100K secrets in Docker Hub images
4. **No layer attribution** — "which Dockerfile instruction caused this 200MB?"

These are the 4 features to build, in order.

---

## Feature 1: `--summary` flag (Small)

One-screen verdict replacing per-file listing:
```
rift --summary myapp:v1 myapp:v2

  Image:    myapp:v1 → myapp:v2
  Size:     142 MB → 148 MB (+6 MB)
  Files:    12 added, 3 removed, 47 modified
  Packages: busybox 1.36→1.37, openssl 3.1→3.2, +2 new, -1 removed
  Security: 1 new executable, 0 SUID/SGID, 0 world-writable
  Verdict:  ⚠ 1 security finding
```

**Create:**
- `internal/output/summary.go` — `FormatSummaryReport(result, image1, image2, layerSummary, events, pkgChanges) string`
- `internal/output/summary_test.go`

**Modify:**
- `cmd/root.go` — add `summary bool` flag, auto-enable `--packages` when summary is active, early-return with summary output before format switch

**Key detail:** When `--summary` is set, force `flags.showPackages = true` right after `applyConfig()` so package data is always available for the summary line.

---

## Feature 2: Policy rules in `.rift.yml` (Medium)

```yaml
policy:
  max-size-growth: 50MB
  no-new-suid: true
  no-world-writable: true
  max-new-executables: 10
```

`rift --policy myapp:v1 myapp:v2` evaluates rules and exits 2 on failure.

**Create:**
- `internal/policy/policy.go` — `PolicyConfig` struct, `RuleResult` struct, `Evaluate(cfg, result, events) []RuleResult` pure function
- `internal/policy/policy_test.go`

**Modify:**
- `internal/config/config.go` — add `PolicyConfig` struct, add `Policy *PolicyConfig` to `Config`
- `internal/config/config_test.go` — test YAML parsing of policy section
- `cmd/root.go` — add `--policy` flag, evaluate after security analysis, print pass/fail per rule, exit 2 if any fail

**Rules:**
- `max-size-growth` — reuse `exitcode.ParseSizeThreshold()` for parsing, check `AddedBytes - RemovedBytes > threshold`
- `no-new-suid` — fail if any `KindNewSUID`, `KindSUIDAdded`, `KindNewSGID`, `KindSGIDAdded` events
- `no-world-writable` — fail if any `KindWorldWritable` events
- `max-new-executables` — count `KindNewExecutable` events, fail if > limit

---

## Feature 3: Secrets detection (Medium)

Scan added/modified files for secrets patterns.

**Create:**
- `internal/secrets/secrets.go` — `SecretFinding` struct, `AnalyzePaths(result) []SecretFinding` (pure, fast, always-on), `AnalyzeContent(path, content) []SecretFinding` (regex-based, gated behind `--secrets`), `ToSecurityEvents(findings) []SecurityEvent`
- `internal/secrets/secrets_test.go`

**Modify:**
- `internal/security/security.go` — add new `SecurityEventKind` constants: `KindSecretPrivateKey`, `KindSecretAWSKey`, `KindSecretAPIToken`, `KindSecretFilePath`
- `internal/output/sarif.go` — add SARIF rules for new secret kinds (all level `"error"`)
- `cmd/root.go` — add `--secrets` flag, run path-based detection always, content-based when `--secrets` set, merge into events slice

**Detection patterns:**
- Paths: `.env`, `*.pem`, `id_rsa`, `id_ed25519`, `.ssh/*`, `credentials.json`, `.aws/credentials`
- Content: `-----BEGIN (RSA|EC|OPENSSH|DSA) PRIVATE KEY-----`, `AKIA[0-9A-Z]{16}`, generic `api_key|api_secret|access_token` patterns

**Design:** Convert `SecretFinding` → `SecurityEvent` via `ToSecurityEvents()` so all existing formatters and `--fail-on-security` work without changes.

---

## Feature 4: Layer attribution with `--layers` (Large)

Group changes by the Dockerfile instruction that caused them:
```
Layer 3 (RUN apt-get install -y curl nginx)
  + usr/bin/curl          (+2.1 MB)
  + usr/lib/libcurl.so.4  (+512 KB)
  Layer total: +2.6 MB

Layer 5 (RUN npm install)
  ~ node_modules/express/index.js  [content] (+12 bytes)
  + node_modules/new-dep/          (+4.3 MB)
  Layer total: +4.3 MB
```

**Create:**
- `internal/output/layerattr.go` — `LayerGroup` struct, `GroupByLayer(entries) []LayerGroup`, `FormatLayerAttribution(groups, image1, image2) string`
- `internal/output/layerattr_test.go`

**Modify:**
- `internal/tree/tree.go`:
  - Add `LayerIndex int` to `FileNode`
  - Add `BuildTreeWithAttribution(layers, baseIndex)` that sets `node.LayerIndex = baseIndex + i`
  - `BuildTree` calls `BuildTreeWithAttribution(layers, 0)`
  - `BuildFromImageSkipFirst` calls `BuildTreeWithAttribution(layers[skip:], skip)`
- `internal/tree/tree_test.go` — verify LayerIndex is correct across multi-layer builds, overrides, and skipped layers
- `internal/output/json.go` — add optional `layer_index` and `layer_command` to `ChangeEntry` (omitempty)
- `cmd/root.go`:
  - Add `--layers` flag and `--dockerfile` flag
  - When `--layers`: access `img2.ConfigFile().History` for layer commands, group entries, render with `FormatLayerAttribution`
  - When `--dockerfile`: parse Dockerfile and map RUN/COPY/ADD instructions to layer indices (stretch goal)

**Layer commands from go-containerregistry:**
```go
cfg, _ := img.ConfigFile()
for i, h := range cfg.History {
    fmt.Println(i, h.CreatedBy)  // e.g., "/bin/sh -c apt-get install -y curl"
}
```

---

## Execution Order

| # | Feature | Scope | New flags |
|---|---------|-------|-----------|
| 1 | `--summary` | Small | `--summary` |
| 2 | Policy | Medium | `--policy` |
| 3 | Secrets | Medium | `--secrets` |
| 4 | Layer attribution | Large | `--layers`, `--dockerfile` |

No cross-dependencies. Each can be tested independently.

## Key Files

- `cmd/root.go` — all 4 features wire in here (flags, RunE pipeline)
- `internal/tree/tree.go` — Feature 4 adds LayerIndex to FileNode
- `internal/security/security.go` — Feature 3 adds new SecurityEventKind constants
- `internal/config/config.go` — Feature 2 adds PolicyConfig
- `internal/output/` — Features 1, 4 add new formatters; Feature 3 extends SARIF rules

## Verification

After each feature:
1. `make test` — all tests pass
2. `make lint` — no lint errors
3. `make build && ./rift version` — builds clean
4. Manual test: `./rift alpine:3.18 alpine:3.19` with the new flag
5. Verify existing flags still work unchanged
