# CloudNecromancer

## Project Overview
CLI tool that reconstructs point-in-time AWS infrastructure snapshots by replaying CloudTrail events. Written in Go 1.25+.

## Build & Test
```bash
make build    # builds to bin/cloudnecromancer
make test     # go test ./...
make lint     # golangci-lint run
```

## Architecture
- `cmd/` — Cobra CLI commands (fetch, resurrect, diff, export, info)
- `internal/aws/` — AWS SDK client interface + CloudTrail fetcher
- `internal/splunk/` — Splunk REST API client + CloudTrail event fetcher
- `internal/parser/` — Event parser interface, registry, 23 per-service parsers (133 CloudTrail events)
- `internal/engine/` — Resurrection replay engine, snapshot model, diff logic
- `internal/store/` — DuckDB event cache
- `internal/export/` — Exporters (JSON, Terraform/HCL, CloudFormation, CDK, Pulumi, OCSF, CSV)
- `testdata/` — CloudTrail event JSON fixtures for parser tests

## Data Sources
The `fetch` command supports two sources (via `--source` flag):
- `cloudtrail` (default) — CloudTrail `LookupEvents` API (90-day limit)
- `splunk` — Splunk REST API `/services/search/jobs/export` (no time limit)

Both sources produce `[]aws.RawEvent` which flows into the same store/parse/replay pipeline.

Splunk auth: bearer token via `--splunk-token` or `SPLUNK_TOKEN` env var.
Splunk URL: via `--splunk-url` or `SPLUNK_URL` env var.

## Key Patterns
- All AWS calls go through the `CloudTrailAPI` interface (in `internal/aws/client.go`) for testability
- Splunk client uses `/services/search/jobs/export` for streaming results (no pagination)
- Parsers self-register via `init()` → `parser.Register()`
- `ResourceDelta` is the shared contract between parser and engine
- Exporters implement `export.Exporter` interface
- Table-driven tests with `t.Run()` subtests throughout
- Errors wrapped with `fmt.Errorf("context: %w", err)`

## Dependencies
- CLI: `github.com/spf13/cobra`
- AWS: `github.com/aws/aws-sdk-go-v2`
- DB: `github.com/marcboeker/go-duckdb` (CGO required)
- Output: `github.com/charmbracelet/lipgloss`, `github.com/schollz/progressbar/v3`
- Concurrency: `golang.org/x/sync/errgroup`
- Testing: `github.com/stretchr/testify`
- Splunk: net/http (stdlib only, no external SDK)

## Export Formats
`GetExporter(format)` in `internal/export/exporter.go` supports:
- `json` — indented JSON snapshot
- `terraform` / `hcl` / `tf` — Terraform HCL with import blocks
- `cloudformation` / `cfn` — CloudFormation JSON template
- `cdk` — CDK TypeScript stack
- `pulumi` — Pulumi TypeScript program
- `ocsf` — OCSF Inventory Info (NDJSON)
- `csv` — Splunk lookup table

## Security
- Code generation exporters (HCL, CDK, Pulumi) sanitize attribute keys and values
- CSV exporter sanitizes cells against formula injection
- SQL queries are fully parameterized (no string interpolation)
- Output files created with 0600 permissions
- DuckDB files get 0600 on creation
- File paths cleaned with `filepath.Clean()`
- Account ID validated as 12-digit number

## Release
- Tag `vX.Y.Z` → triggers `.github/workflows/release.yml`
- Builds native binaries: linux/amd64, darwin/amd64, darwin/arm64
- Auto-publishes Homebrew formula to `pfrederiksen/homebrew-tap`
- Secret `HOMEBREW_TAP_GITHUB_TOKEN` required for tap push

## Code Quality
- No `panic` in library code
- `go vet` and `staticcheck` clean
- 80%+ test coverage target on `internal/` packages
- No real AWS calls in tests — use mock implementing `CloudTrailAPI`
- Splunk tests use `httptest.NewServer` for HTTP-level testing
