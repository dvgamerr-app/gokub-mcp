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

### Quick Start

```bash
# 1️⃣ Clone repository
git clone https://github.com/dvgamerr-app/gokub-mcp.git
cd gokub-mcp

# 2️⃣ Install dependencies
go mod download

# 3️⃣ Create .env file
echo "BTK_APIKEY=your_api_key_here" > .env
echo "BTK_SECRET=your_secret_key_here" >> .env

# 4️⃣ Run server
go run main.go
```

### 🏗️ Build Executable

```bash
# Windows
go build -o bitkub-mcp.exe
./bitkub-mcp.exe

# Linux/Mac
go build -o bitkub-mcp
./bitkub-mcp
```

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

| Endpoint | Purpose | Method |
|----------|---------|--------|
| `http://localhost:8080` | Main URL | GET |
| `http://localhost:8080/sse` | SSE Connection | GET |
| `http://localhost:8080/message` | Send Message | POST |

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

The repo doubles as a plugin marketplace for both Claude Code and Codex.

```bash
# Claude Code
/plugin marketplace add dvgamerr-app/gokub-mcp
/plugin install bitkub-trade@gokub-mcp

# Codex
codex plugin marketplace add dvgamerr-app/gokub-mcp
codex plugin add bitkub-trade@gokub-mcp
```

See [`plugins/bitkub-trade`](plugins/bitkub-trade/README.md) for details.


## ⚙️ Configuration

### 🔐 API Keys Setup

สร้างไฟล์ `.env` ใน root directory:

```bash
BTK_APIKEY=your_api_key
BTK_SECRET=your_secret_key
```


### 🤖 Claude Desktop Integration

<details open>
<summary><b>HTTP/SSE Mode (แนะนำ)</b></summary>

เพิ่มการตั้งค่าใน Claude Desktop config:

**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`  
**Mac:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "bitkub": {
      "url": "http://localhost:8080/sse",
      "transport": "sse"
    }
  }
}
```

</details>

<details>
<summary><b>Stdio Mode (Legacy)</b></summary>

```json
{
  "mcpServers": {
    "bitkub": {
      "command": "e:\\.dvgamerr\\gokub-mcp\\bitkub-mcp.exe",
      "env": {
        "BTK_APIKEY": "your_api_key",
        "BTK_SECRET": "your_secret_key"
      }
    }
  }
}
```

> ⚠️ **หมายเหตุ:** ควรตั้งค่า API keys ผ่าน environment variables แทนการใส่ใน config file

</details>

## 📁 Project Structure

```
gokub-mcp/
├── 📄 main.go                      # MCP Server entry point (HTTP/SSE) + tool registration
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
