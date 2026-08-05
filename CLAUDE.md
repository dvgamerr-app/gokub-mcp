Use `AGENTS.md` as the primary repository instructions.

## Project boundaries

- Go 1.25.5 MCP server with stdio as the default transport and HTTP/SSE behind `-serv`.
- Public market-data tools work without credentials. Wallet and order tools are authenticated and may affect a live Bitkub account.
- Preserve the long-only risk controls, `validate_trade_setup` gate, exchange rounding, stop handling, and fee-aware calculations.
- Default automated tests must not use the `live` build tag or call authenticated wallet/order APIs.
- Skill edits require a matching frontmatter version bump. This optimization pass does not change skill files.

## Validation

```powershell
go mod verify
go test -count=1 ./...
go vet ./...
go build -o "$env:TEMP\gokub-mcp-check.exe" .
git diff --check
```

When Go cannot read VCS metadata because the repository is outside Git's trusted directories, add `-buildvcs=false` to Go commands rather than changing global Git configuration.

The race detector is optional and requires a CGO-enabled Go installation plus a C compiler. The default test suite intentionally excludes the `live` build tag.

## Optimization log

### 2026-08-02

The repository started from a clean `main` working tree.

- Reuse `utils.GetFloat64ArrayArg` across EMA, RSI, and ROC handlers, preserving their existing validation messages while removing repeated array conversion logic.
- Calculate Wilder RSI gains and losses as running values instead of growing two temporary slices.
- Handle the valid one-price EMA(1) case without indexing before the start of the EMA slice; current and previous EMA are equal and the trend is neutral.
- Remove the malformed, unused indirect module path `github.com/wk8/gov0.46.0d-map/v2`, restoring module verification and documentation tooling without retaining an unnecessary dependency.
- Correct README requirements, stdio/HTTP commands, logging variables, test boundaries, and the project structure description.

## Evidence

For a 512-price RSI input, five-run benchmark medians improved from approximately 7,310 ns/op, 16,256 B/op, and 14 allocations/op to 2,493 ns/op, 0 B/op, and 0 allocations/op. Existing formula and handler tests plus focused array-parser and EMA(1) regression tests preserve output behavior.

## Final verification

- `go mod verify`: passed; the baseline malformed module path previously made this command fail.
- `gofmt -d` on every touched Go file: clean.
- `go test -buildvcs=false -count=1 ./...`: passed across all five packages.
- `go vet -buildvcs=false ./...` and a temporary `go build -buildvcs=false`: passed.
- `git diff --check`: passed with only the repository's LF-to-CRLF working-tree warnings.
- No external linter is configured in the repository, and neither `golangci-lint` nor `staticcheck` is installed; `go vet` is the available lint check.
- `go test -race` was unavailable because this environment has `CGO_ENABLED=0` and no C compiler. Live-tagged network tests were not run by design.
- A read-only `go mod tidy -diff` still reports broader pre-existing stale indirect requirements and checksums. They were left unchanged to avoid unrelated dependency churn.
