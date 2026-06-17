# AGENTS.md

@C:\Users\dvgamerr\.codex\RTK.md

Bitkub MCP Server — a Go MCP server for Bitkub Exchange: market data, indicators,
risk sizing, orders, trade logging, plus the `bitkub-trade` playbook plugin.

## Stack

Go `1.25.5` · module `github.com/dvgamerr-app/gokub-mcp` · `mark3labs/mcp-go` ·
`dvgamerr-app/go-bitkub` · `rs/zerolog`

## Layout

```text
main.go                  entrypoint + tool registration
tools/                   MCP tools (+ tests)
prompts/ resources/      MCP prompts & resources
utils/                   args, helpers, logging
plugins/bitkub-trade/    trading playbook plugin (skill)
.agents/ .claude-plugin/ Codex & Claude marketplace metadata
docs/                    images & docs
```

## Conventions

- Reuse or modify existing code before adding files; prefer small, focused changes.
- Self-explanatory code — no comments, no inline docs.
- New tool → put in `tools/`, match nearby naming, register in `main.go`, add a focused
  test when the logic is non-trivial.
- Tool shape: `New…Tool() mcp.Tool` + a handler that validates args via `utils/`,
  returns `utils.ErrorResult` / `utils.ArtifactsResult`, and logs with zerolog.

## Trading safety

- Order and wallet tools are **live** — validate inputs before calling Bitkub.
- Never bypass `validate_trade_setup`; preserve risk controls (max risk per trade,
  exchange rounding, stop handling, fee-aware math). Long-only assumptions follow
  `plugins/bitkub-trade/skills/playbook/SKILL.md`.
- Keep market-data tools usable without API keys.

## Shell (rtk)

- Prefix executables with `rtk` (e.g. `rtk rg` to search); if `rtk` can't run a builtin
  or cmdlet, use the nearest executable.
- Stage only the files for the current change — no `git add .` / `-A` / `-u` unless asked.

## Env

| var | default | purpose |
|-----|---------|---------|
| `BTK_APIKEY` / `BTK_SECRET` | — | Bitkub auth (wallet/orders) |
| `PORT` | `3000` | HTTP/SSE port (`-serv`) |
| `TRADES_FILE` | `trades.json` | trade journal path |
| `LOG_FORMAT` | JSON | `text` for console |
| `LOG_LEVEL` | `info` | `trace`…`panic` |

## Commands

```bash
go run main.go            # stdio
go run main.go -serv      # HTTP/SSE
go test ./...
go install .
docker build -t gokub-mcp .
```

Give the user the test command instead of running tests; no summary after finishing.
