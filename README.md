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

### Native trade analysis (no client scripts)

Call `analyze_market_trades` with `trend_resolution: 240`, `entry_resolution: 15`,
`target_net_percent: 2`, and `risk_percent: 2` to scan every liquid exchange-only THB
pair. It checks closed fresh candles, regime/EMA, relative strength, ATR, entry
signals, available THB, fees, pending orders and journal positions. Results include
all pairs, explicit rejection/error reasons and fee/spread-aware plans. It never
places orders. Check `complete` and `analysis_error_count` before concluding that
all pairs were analyzed successfully.

For individual steps, use `get_market_screener(exchange_only=true, limit=0)` to
return every passing pair, and `get_historical_candles(..., closed_only=true)` to
exclude forming candles and fill no-trade gaps with the previous close and zero
volume. Filled bars carry `synthetic: true` (also a CSV column); `history_metadata`
records the last actual candle and filled count. Duplicate timestamps keep the last
API update without adding volumes. The analyzer blocks entry on zero-volume bars;
unavailable/out-of-window source data still errors. Completed candles are now the
default; request `closed_only=false` explicitly for forming candles. Supply the same
`as_of` Unix timestamp to align separate requests. Multi-symbol history fails explicitly
if any requested symbol is unavailable.

**42 tools** covering the full long-only swing-trading workflow (screen → analyze →
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
`check_market_regime` · `analyze_chart`
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
<summary><b>Helpers</b> (5)</summary>

`get_market_overview` · `simulate_trade` · `backtest_signals` · `pnl_with_fees` · `calculate_capm`

`calculate_capm` is an assumption-based asset-pricing calculator, not a chart forecast.
</details>

### Analysis semantics and historical evaluation

- Prefer completed OHLCV candles for `check_market_regime`. Wilder ADX requires
  at least 28 candles; price-only and close-only inputs remain supported with
  `adx_available=false`. OHLC regimes require ADX ≥25, matching +DI/-DI direction
  and net price direction; up-bar counts cannot bypass weak ADX. The legacy
  close-only assessment is explicitly marked `method=close_only_fallback`. Regime
  scores are heuristic scores, not win probabilities.
- Indicator and signal outputs preserve price precision. EMA output preserves its
  input-aligned array; ignore warmup zeros before `valid_from_index`. It exposes
  `trend_available`; exactly one seeded EMA has no trend.
  Flat-price RSI is defined as 50 (neutral).
- Malformed OHLC ranges, non-positive prices and invalid timestamps fail explicitly.
  Timestamps are optional for existing indicator callers; when supplied they must
  be present on every row, increasing and on a regular grid. Signal callers must
  supply completed candles; breakout requires volume, pullback/exit require open
  and volume. Synthetic latest bars cannot confirm entries or exit decisions.
- Pullback requires a preceding decline, bullish price recovery and rising RSI in
  the configured bounce zone. A candle merely closing near its high is insufficient.
- Relative-strength ranking requires benchmark data. `top3` remains the absolute
  ROC ranking; `relative_roc` and `outperforms_benchmark` are separate. The native
  workflow requires both top-three rank and outperformance (the benchmark itself
  is allowed), and uses one cutoff for all candle requests. Direct price arrays
  must be aligned by the caller before ranking.
- The native 2–6% volatility gate always uses **daily ATR(14)**, independent of the
  selected trend timeframe. `daily_atr` is the gate input; `trend_atr` and
  `entry_atr` retain their own timeframe meanings. Stale actual source candles
  block analysis even if synthetic bars fill the grid.
- `check_exit_signals` returns an error when data cannot support an exit decision,
  including insufficient history and a latest bar with no actual trades.

Call `backtest_signals(candles, resolution, as_of, strategy="breakout")` with
221–10000 timestamped completed OHLCV candles. It uses 220 candles for warmup,
then evaluates fixed rules in up to three default chronological
folds with expanding history. Optional inputs: `folds` (1–10), `fee_percent`
(per side, default 0.25), `slippage_percent` (per side including spread, default
0.1), and `reward_risk` (gross distance ratio, default 2).

Signals use prior candles only. Simulated entries fill at the next actual open;
stops gap to the open, and a candle touching both stop and target counts as a stop.
Folds close remaining positions only on an actual final bar. A no-trade tail
leaves `unresolved_positions` and `complete=false`, without looking ahead to choose
an earlier exit. Outputs include
trade records, per-fold and total win rates, approximate Wilson intervals, average
net returns/R, equal-notional profit factor, and explicitly hypothetical
full-notional closed-trade compounded returns/drawdown. They do not measure
intratrade or live-portfolio drawdown.
No trades produces null rates, not evidence of accuracy. This evaluates entry
signals only: it does not reproduce the live multi-symbol regime/ranking/ATR gate,
portfolio sizing, exchange rounding or limit-order fills. It performs no parameter
optimization and does not establish calibrated prediction probabilities.

Supported fixed-rule experiments: `breakout`, `pullback`, `macd_cross` (12/26/9),
`supertrend_flip` (10/3), `bollinger_breakout` (20/2), and `ema_crossover` (20/50).
Chart crossover experiments use 1.5×ATR14 initial stops and the same costs/fill/target
model. They do not become live-entry conditions automatically.

Call `analyze_chart(candles, resolution, as_of)` with 220–10000 completed timestamped
OHLCV bars for EMA/RSI/ATR/DMI, MACD, Bollinger Bands, Supertrend and anchored VWAP
readings. Default VWAP anchor is the last candle's UTC midnight; optional
`vwap_anchor` must be a covered candle-open timestamp. Uncovered anchors or no
volume return an unavailable VWAP. A daily anchor on daily bars is not meaningful.
Outputs disclose SMA-seeded EMA and ATR initialization differences from TradingView;
these are standard formulas, not a claim of exact platform parity.

Native trade analysis includes these entry-timeframe diagnostics as `chart`, and
requests 1000 trend candles to reduce EMA200 seed sensitivity. Existing risk gates
and exchange rounding remain required. The new indicators do not form an extra
majority-vote confidence score.

See [TradingView research and recorded comparison](docs/tradingview-research.md) for
primary sources, keep/add/defer decisions, numerical conventions and a reproducible
48-case public-market snapshot. It does not establish any universally best indicator.

## 🧠 Trading Playbook (Plugin / Skill)

The tools are the *hands*; the **`bitkub-trade`** plugin is the *brain* — a playbook
skill that encodes the strategy (screen → regime → relative-strength → ATR → signal →
size → **validate gate** → round → place → manage → log) and the hard guardrails
(≤2% risk/trade, long-only, no entry unless `can_trade=true`, client-side stop, fee-aware net-profit targets).

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
├── 📂 tools/                       # 40 MCP tools (+ unit tests)
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
- [x] **40 trading tools** — foundation, indicators, entry signals, risk, orders, management, logging, native market analysis
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
