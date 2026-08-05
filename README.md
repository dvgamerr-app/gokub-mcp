<div align="center">

# 🚀 Bitkub MCP Server

[![CodeQL](https://github.com/dvgamerr-app/gokub-mcp/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/dvgamerr-app/gokub-mcp/actions/workflows/codeql-analysis.yml)
[![Go Version](https://img.shields.io/badge/Go-1.25.5+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Discord](https://img.shields.io/badge/Discord-Join%20Us-7289DA?style=flat&logo=discord)](https://discord.gg/QDccF497Mw)

**Model Context Protocol server for Bitkub Cryptocurrency Exchange API**

*เชื่อมต่อ Claude Desktop กับ Bitkub Exchange ผ่าน MCP Protocol*

![logo](./docs/logo-ai.png)

[Features](#-features) • [Installation](#-installation) • [API Tools](#-available-tools) • [Configuration](#-configuration) • [Community](#-community)

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🎯 Core Features
- ✅ **MCP 2026-07-28** - Built with the official Go SDK
- 🌐 **Streamable HTTP** - Stateless `/mcp` endpoint
- 🔐 **Secure** - HMAC SHA256 signature
- 💰 **Wallet** - View balances & transactions

</td>
<td width="50%">

### 🚀 Developer Experience
- 💎 **Go-Bitkub SDK** - Full API v3 support
- 🧠 **Integration** - with Claude Desktop
- 📊 **Market Data** - Real-time ticker & depth
- 📖 **Order** - Track open orders

</td>
</tr>
</table>

## 🔧 Installation

> Fastest path: **install the server → register it as an MCP server in Claude Code / Codex →
> install the `bitkub-trade` playbook plugin.** Steps 3–5 are the important part.

### 1️⃣ Install the MCP server

Requires Go 1.25.5 or newer.

```bash
go install github.com/dvgamerr-app/gokub-mcp@latest
```

This installs `gokub-mcp` to `$(go env GOPATH)/bin` by default.

Windows default path:

```powershell
# PowerShell
Write-Output "$env:USERPROFILE\go\bin\gokub-mcp.exe"
```

```cmd
:: cmd.exe
echo %USERPROFILE%\go\bin\gokub-mcp.exe
```

```nu
# NuShell
$env.USERPROFILE | path join go bin gokub-mcp.exe
```

```bash
# macOS/Linux default
$(go env GOPATH)/bin/gokub-mcp
```

For local development from a clone:

```bash
git clone https://github.com/dvgamerr-app/gokub-mcp.git
cd gokub-mcp
go install .
```

> Prefer containers? Use [🐳 Docker](#-docker) (Streamable HTTP mode).

### 2️⃣ Credentials

The server reads `BTK_APIKEY` / `BTK_SECRET` from its **environment** — pass them when you
register the server (steps 3–4), so no `.env` file is required. For a plain local
`go run .`, a `.env` in the repo root still works:

```bash
BTK_APIKEY=your_api_key
BTK_SECRET=your_secret_key
```

### 3️⃣ Register the MCP server in Claude Code

The current way is the `claude mcp add` CLI (replaces hand-editing JSON). Pass the env
inline with `-e`:

```bash
claude mcp add bitkub \
  -e BTK_APIKEY=your_api_key \
  -e BTK_SECRET=your_secret_key \
  -- "$(go env GOPATH)/bin/gokub-mcp"
```

Windows:

```powershell
# PowerShell
claude mcp add bitkub `
  -e BTK_APIKEY=your_api_key `
  -e BTK_SECRET=your_secret_key `
  -- "$env:USERPROFILE\go\bin\gokub-mcp.exe"
```

```cmd
:: cmd.exe
claude mcp add bitkub -e BTK_APIKEY=your_api_key -e BTK_SECRET=your_secret_key -- "%USERPROFILE%\go\bin\gokub-mcp.exe"
```

```nu
# NuShell
let gokub = ($env.USERPROFILE | path join go bin gokub-mcp.exe)
claude mcp add bitkub -e BTK_APIKEY=your_api_key -e BTK_SECRET=your_secret_key -- $gokub
```

- Add `-s user` to make it global across all projects (default scope is project-local `.mcp.json`).
- It runs over **stdio** by default — no extra flags. Verify with `/mcp` in a session.

### 4️⃣ Register the MCP server in Codex

```bash
codex mcp add gokub \
  --env BTK_APIKEY=your_api_key \
  --env BTK_SECRET=your_secret_key \
  -- "$(go env GOPATH)/bin/gokub-mcp"
```

Windows:

```powershell
# PowerShell
codex mcp add gokub `
  --env BTK_APIKEY=your_api_key `
  --env BTK_SECRET=your_secret_key `
  -- "$env:USERPROFILE\go\bin\gokub-mcp.exe"
```

```cmd
:: cmd.exe
codex mcp add gokub ^
  --env BTK_APIKEY=your_api_key ^
  --env BTK_SECRET=your_secret_key ^
  -- "%USERPROFILE%\go\bin\gokub-mcp.exe"
```

```nu
# NuShell
let gokub = ($env.USERPROFILE | path join go bin gokub-mcp.exe)
codex mcp add gokub --env BTK_APIKEY=your_api_key --env BTK_SECRET=your_secret_key -- $gokub
```

…or edit `~/.codex/config.toml` directly:

```toml
[mcp_servers.gokub]
command = "/absolute/path/to/gokub-mcp"
args = []

[mcp_servers.gokub.env]
BTK_APIKEY = "your_api_key"
BTK_SECRET = "your_secret_key"
```

Verify with `/mcp` inside Codex.

### 5️⃣ Install the trading playbook plugin (skill)

The tools are the *hands*; the **`bitkub-trade`** plugin is the *brain* — the strategy
playbook + guardrails. The repo doubles as a plugin marketplace for both clients:

```bash
# Claude Code
/plugin marketplace add dvgamerr-app/gokub-mcp
/plugin install bitkub-trade@gokub-mcp

# Codex
codex plugin marketplace add dvgamerr-app/gokub-mcp
codex plugin add bitkub-trade@gokub-mcp
```

See [`plugins/bitkub-trade`](plugins/bitkub-trade/README.md) for details.

## 🐳 Docker

Runs the server in **stateless Streamable HTTP mode** (good for a shared/remote server; for a local
client that spawns the binary over stdio, use the native build above):

```bash
docker build -t gokub-mcp .
docker run --rm -p 8080:8080 \
  -e BTK_APIKEY=your_api_key \
  -e BTK_SECRET=your_secret_key \
  gokub-mcp
```

Then point a Streamable HTTP client at `http://localhost:8080/mcp`. Override the port with
`-e PORT=9000 -p 9000:9000`.

> The build fetches Go deps from the module proxy (needs internet). For an offline /
> hermetic build, run `go mod vendor` first and build with `-mod=vendor`.

## 🎮 Usage

### Standard input/output mode

This is the default transport for locally registered Claude Code and Codex clients:

```bash
go run .
```

### Streamable HTTP Server Mode

```bash
# Default port 3000
go run . -serv

# Custom port
PORT=9000 go run . -serv
```

<details>
<summary>📡 Server Endpoints</summary>

> Default port is **3000** (override with `PORT`). Docker image defaults to 8080.

| Endpoint | Purpose | Method |
|----------|---------|--------|
| `http://localhost:3000/mcp` | Stateless MCP requests | POST |

</details>

## 🛠️ Available Tools

**39 tools** covering the full long-only swing-trading workflow (screen → analyze →
size → place → manage → log).

<details open>
<summary><b>Foundation</b> (7)</summary>

`get_wallet_balance` · `get_ticker` · `get_market_depth` · `get_my_open_orders` ·
`get_symbols` · `get_symbol_rules` · `get_fee_schedule`
</details>

<details>
<summary><b>Market & Liquidity</b> (5)</summary>

`calculate_spread` · `calculate_liquidity_depth` · `get_market_screener` ·
`get_historical_candles` · `extract_close_prices`
</details>

<details>
<summary><b>Indicators</b> (6)</summary>

`calculate_ema` · `calculate_roc` · `calculate_atr` · `calculate_rsi` ·
`check_market_regime` · `calculate_capm`
</details>

<details>
<summary><b>Entry Signals</b> (3)</summary>

`detect_breakout_signal` · `detect_pullback_signal` · `calculate_relative_strength_rank`
</details>

<details>
<summary><b>Position & Risk</b> (3)</summary>

`calculate_position_size` · `validate_trade_setup` · `round_to_exchange_rules`
</details>

<details>
<summary><b>Order Management</b> (4)</summary>

`place_limit_order` · `place_stop_limit_order` · `cancel_order` · `get_order_status`
</details>

<details>
<summary><b>Trade Management</b> (4)</summary>

`check_trade_pnl` · `calculate_trailing_stop` · `check_exit_signals` ·
`client_side_stop_worker` *(Bitkub has no native stop/OCO — client-side trigger)*
</details>

<details>
<summary><b>Logging & Performance</b> (4)</summary>

`log_trade_entry` · `log_trade_exit` · `calculate_expectancy` · `get_trade_history`
*(flat-file journal `trades.json`, override with `TRADES_FILE`)*
</details>

<details>
<summary><b>Helpers</b> (3)</summary>

`get_market_overview` · `simulate_trade` · `pnl_with_fees`
</details>

## 🧠 Trading Playbook (Plugin / Skill)

The tools are the *hands*; the **`bitkub-trade`** plugin is the *brain* — a playbook
skill that encodes the strategy (screen → regime → relative-strength → ATR → signal →
size → **validate gate** → round → place → manage → log) and the hard guardrails
(≤2% risk/trade, long-only, no entry unless `can_trade=true`, client-side stop, TP ≥2R).

→ Install it in [step 5 of Installation](#-installation). Details:
[`plugins/bitkub-trade`](plugins/bitkub-trade/README.md).


## ⚙️ Configuration

### 🔐 Environment variables

| Variable | Required | Purpose |
|----------|----------|---------|
| `BTK_APIKEY` | for trading/wallet | Bitkub API key |
| `BTK_SECRET` | for trading/wallet | Bitkub API secret |
| `PORT` | no (default `3000`) | Streamable HTTP port (`-serv` mode) |
| `TRADES_FILE` | no (default `trades.json`) | trade-journal file path |
| `LOG_FORMAT` | no (default JSON) | Set to `text` for console-friendly logs |
| `LOG_LEVEL` | no (default `info`) | `trace`, `debug`, `info`, `warn`, `error`, `fatal`, or `panic` |

Pass keys when you register the MCP server (see [Installation](#-installation) steps 3–4) —
that's preferred over a committed `.env`. Public market-data tools work without keys;
wallet/order tools need them.

> ⚠️ Keep secrets in env vars / your client config, never commit `.env` (it's gitignored).

<details>
<summary><b>🤖 Remote Streamable HTTP config</b></summary>

If your MCP client accepts a remote server URL, point it at the Streamable HTTP endpoint:

```json
{
  "mcpServers": {
    "gokub": { "type": "http", "url": "http://localhost:3000/mcp" }
  }
}
```

For **Claude Code / Codex**, prefer `claude mcp add` / `codex mcp add` (stdio) above.

</details>

## 🧪 Development and verification

```bash
go mod verify
go test ./...
go vet ./...
go build ./...
```

The default test suite is deterministic and excludes `tools/get_historical_candles_live_test.go`.
The opt-in `live` build tag calls public Bitkub market-data endpoints and should only be
used intentionally. Do not run authenticated wallet or order flows as automated tests
against real credentials.

## 📁 Project Structure

```
gokub-mcp/
├── 📄 main.go                      # MCP Server entry point (Streamable HTTP) + tool registration
├── 🐳 Dockerfile                   # Streamable HTTP server image
├── 📂 tools/                       # 39 MCP tools (+ unit tests)
├── 📂 prompts/                     # trading_strategy, market_analysis prompts
├── 📂 resources/                   # bitkub://symbols, bitkub://ticker/{symbol}
├── 📂 utils/                       # Utility functions
├── 📂 docs/                        # Project images and supporting documentation
├── 📂 plugins/bitkub-trade/        # bitkub-trade plugin (playbook skill)
│   ├── .claude-plugin/plugin.json  # Claude manifest
│   ├── .codex-plugin/plugin.json   # Codex manifest
│   └── skills/playbook/SKILL.md    # shared trading playbook
├── 📂 .claude-plugin/              # Claude plugin marketplace
└── 📂 .agents/plugins/             # Codex plugin marketplace
```

## 📊 API Rate Limits

| Category | Rate Limit | Note |
|----------|------------|------|
| 📈 Market Data | 100 req/sec | Public endpoints |
| 💱 Trading Operations | 150-200 req/sec | Authenticated endpoints |

> 📚 [Bitkub API Docs](https://github.com/bitkub/bitkub-official-api-docs) สำหรับข้อมูลเพิ่มเติม

## 🚀 Roadmap

### ✅ Completed
- [x] Bitkub API golang library + MCP Server (Streamable HTTP + stdio)
- [x] **39 trading tools** — foundation, indicators, entry signals, risk, orders, management, logging
- [x] Risk-based position sizing + pre-trade `validate_trade_setup` gate
- [x] Client-side stop worker (Bitkub has no native stop/OCO)
- [x] Trade journal + expectancy (flat-file `trades.json`)
- [x] **`bitkub-trade` playbook plugin** — installable in Claude Code & Codex

### 🚧 In Progress / Next
- [ ] `git push` the plugin so the marketplaces resolve from GitHub
- [ ] Optional MCP prompts/resources for non-Claude/Codex clients
- [ ] Backtesting over OHLCV history

### 🎯 Planned Features
- [ ] Rebalancing / Grid strategy presets
- [ ] Docker image + WebSocket real-time data

## 📚 References

🔧 [**Official MCP Go SDK**](https://github.com/modelcontextprotocol/go-sdk)

💎 [**Go-Bitkub SDK**](https://github.com/dvgamerr-app/go-bitkub)

📖 [**Bitkub Official API Docs**](https://github.com/bitkub/bitkub-official-api-docs)

🤖 [**MCP 2026-07-28 Specification**](https://modelcontextprotocol.io/specification/2026-07-28)



## 👥 Community

<div align="center">

[![Discord](https://img.shields.io/badge/Discord-Join%20Our%20Server-7289DA?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/QDccF497Mw)

**Join our community to discuss, get help, and share your trading strategies!**

**Made with ❤️ by [dvgamerr-app](https://github.com/dvgamerr-app)**

⭐ Star this repo if you find it helpful!

[Report Bug](https://github.com/dvgamerr-app/gokub-mcp/issues) • [Request Feature](https://github.com/dvgamerr-app/gokub-mcp/issues) • [Contribute](https://github.com/dvgamerr-app/gokub-mcp/pulls)

</div>
