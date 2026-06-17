<div align="center">

# 🚀 Bitkub MCP Server

[![CodeQL](https://github.com/dvgamerr-app/gokub-mcp/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/dvgamerr-app/gokub-mcp/actions/workflows/codeql-analysis.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
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
- ✅ **MCP Server** - Built with mcp-go
- 🌐 **HTTP/SSE** - Real-time communication
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

> Fastest path: **build the server → register it as an MCP server in Claude Code / Codex →
> install the `bitkub-trade` playbook plugin.** Steps 3–5 are the important part.

### 1️⃣ Build the MCP server

```bash
git clone https://github.com/dvgamerr-app/gokub-mcp.git
cd gokub-mcp
go build -o bitkub-mcp .          # Windows: go build -o bitkub-mcp.exe .
```

> Prefer containers? Skip the build and use [🐳 Docker](#-docker) (HTTP/SSE mode).

### 2️⃣ Credentials

The server reads `BTK_APIKEY` / `BTK_SECRET` from its **environment** — pass them when you
register the server (steps 3–4), so no `.env` file is required. For a plain local
`go run main.go`, a `.env` in the repo root still works:

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
  -- /absolute/path/to/bitkub-mcp
```

- Add `-s user` to make it global across all projects (default scope is project-local `.mcp.json`).
- It runs over **stdio** by default — no extra flags. Verify with `/mcp` in a session.

### 4️⃣ Register the MCP server in Codex

```bash
codex mcp add bitkub \
  --env BTK_APIKEY=your_api_key \
  --env BTK_SECRET=your_secret_key \
  -- /absolute/path/to/bitkub-mcp
```

…or edit `~/.codex/config.toml` directly:

```toml
[mcp_servers.bitkub]
command = "/absolute/path/to/bitkub-mcp"
args = []

[mcp_servers.bitkub.env]
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

Runs the server in **HTTP/SSE mode** (good for a shared/remote server; for a local
client that spawns the binary over stdio, use the native build above):

```bash
docker build -t bitkub-mcp .
docker run --rm -p 8080:8080 \
  -e BTK_APIKEY=your_api_key \
  -e BTK_SECRET=your_secret_key \
  bitkub-mcp
```

Then point an SSE-capable client at `http://localhost:8080/sse`. Override the port with
`-e PORT=9000 -p 9000:9000`.

## 🎮 Usage

### HTTP/SSE Server Mode

```bash
# Default port 8080
go run main.go

# Custom port
PORT=3000 go run main.go
```

<details>
<summary>📡 Server Endpoints</summary>

> Default port is **3000** (override with `PORT`). Docker image defaults to 8080.

| Endpoint | Purpose | Method |
|----------|---------|--------|
| `http://localhost:3000/sse` | SSE Connection | GET |
| `http://localhost:3000/msg` | Send Message | POST |

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
| `PORT` | no (default `3000`) | HTTP/SSE port (`-serv` mode) |
| `TRADES_FILE` | no (default `trades.json`) | trade-journal file path |

Pass keys when you register the MCP server (see [Installation](#-installation) steps 3–4) —
that's preferred over a committed `.env`. Public market-data tools work without keys;
wallet/order tools need them.

> ⚠️ Keep secrets in env vars / your client config, never commit `.env` (it's gitignored).

<details>
<summary><b>🤖 Legacy: Claude Desktop (SSE) config</b></summary>

If you run the HTTP server (or the Docker image) and use the Claude **Desktop** app,
point it at the SSE endpoint:

```json
{
  "mcpServers": {
    "bitkub": { "url": "http://localhost:3000/sse", "transport": "sse" }
  }
}
```

**Windows:** `%APPDATA%\Claude\claude_desktop_config.json` ·
**Mac:** `~/Library/Application Support/Claude/claude_desktop_config.json`

For **Claude Code / Codex**, prefer `claude mcp add` / `codex mcp add` (stdio) above.

</details>

## 📁 Project Structure

```
gokub-mcp/
├── 📄 main.go                      # MCP Server entry point (HTTP/SSE) + tool registration
├── 🐳 Dockerfile                   # HTTP/SSE server image
├── 📂 tools/                       # 39 MCP tools (+ unit tests)
├── 📂 prompts/                     # trading_strategy, market_analysis prompts
├── 📂 resources/                   # bitkub://symbols, bitkub://ticker/{symbol}
├── 📂 utils/                       # Utility functions
├── 📂 docs/                        # ASSIGNMENT.md (spec & build progress)
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
- [x] Bitkub API golang library + MCP Server (HTTP/SSE + stdio)
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

🔧 [**MCP-Go Framework**](https://github.com/mark3labs/mcp-go)

💎 [**Go-Bitkub SDK**](https://github.com/dvgamerr-app/go-bitkub)

📖 [**Bitkub Official API Docs**](https://github.com/bitkub/bitkub-official-api-docs)

🤖 [**Protocol MCP Spec**](https://modelcontextprotocol.io/)



## 👥 Community

<div align="center">

[![Discord](https://img.shields.io/badge/Discord-Join%20Our%20Server-7289DA?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/QDccF497Mw)

**Join our community to discuss, get help, and share your trading strategies!**

**Made with ❤️ by [dvgamerr-app](https://github.com/dvgamerr-app)**

⭐ Star this repo if you find it helpful!

[Report Bug](https://github.com/dvgamerr-app/gokub-mcp/issues) • [Request Feature](https://github.com/dvgamerr-app/gokub-mcp/issues) • [Contribute](https://github.com/dvgamerr-app/gokub-mcp/pulls)

</div>
