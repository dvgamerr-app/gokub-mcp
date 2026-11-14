---
name: playbook
description: >-
  Playbook for swing-trading the long side on Bitkub (THB pairs) using the gokub-mcp tools. Use when asked to find a trade, screen coins, evaluate an entry, size a position, place/manage an order, or review trading performance on Bitkub. Encodes the risk rules, the screen→regime→signal→size→validate→place→manage→log flow, and Bitkub-specific constraints (no native stop/OCO).
license: MIT
author: Kananek T.
version: 0.7.1
category: trading
---

# Bitkub Trade Playbook

Long-only swing trading on Bitkub THB pairs via the gokub-mcp tools. The tools are
the hands; this is the brain. Follow the flow, respect the guardrails, never skip the
validation gate.

## Native MCP workflow

Use `analyze_market_trades(trend_resolution=240, entry_resolution=15,
target_net_percent=2, risk_percent=2)` for a full scan. It analyzes every liquid
exchange-only THB pair, rejects broker/frozen pairs, checks completed fresh candles,
relative strength, EMA, ATR, signals, open orders and journal positions, then returns
fee-aware plans and validation results. User-specified timeframes override these
defaults. Do not write JS, shell scripts, or custom API calls to duplicate this flow.

Use the existing MCP tools directly for focused follow-ups:
- `get_market_screener(exchange_only=true, limit=0)` returns ALL passing pairs.
- `get_historical_candles(..., closed_only=true)` excludes forming candles, keeps
  the last update for duplicate timestamps, and fills no-trade gaps with the previous
  close and zero volume. These bars are marked `synthetic`; inspect `history_metadata`
  for the last actual candle. Never treat a synthetic bar as a new trade or enter on
  a zero-volume latest entry bar. Missing/out-of-window source history still errors.
  Pass numeric prices directly.
- Read `complete`, `analysis_error_count`, `screen_stats`, and each result's `error`;
  report unavailable data separately from a failed trading condition.
- `can_trade=true` is an analysis result, not an executed or guaranteed trade.
  Inspect the plan and current state before placing an authorized order using the
  existing order tools. Keep the client-side stop requirement.

The manual flow below documents the same individual steps. Prefer the native
workflow for scans; it never submits orders or starts a background stop worker.

## Execution discipline (never violate)

- **Run the full flow autonomously.** Never pause mid-flow to ask whether to continue. Screen → regime → signal → size → validate → decide — all in one pass.
- **Report results, not options.** End with a decision table (pass/fail per coin, with entry/stop/TP for any that pass) and the reason. Never end with "ถ้าต้องการจะรันต่อ…" or any conditional offer.
- **No raw data dumps.** Never return raw candle arrays or large result blobs to the user. Summarize each intermediate step in ≤ 1 line.

## Hard guardrails (never violate)

- **Risk ≤ 2% of equity per trade.** Size from the stop, not from a fixed amount.
- **Long only, uptrend only.** No shorting, no counter-trend entries.
- **No entry unless `validate_trade_setup` returns `can_trade=true`.** This is the gate.
- **Target 1–2% net profit on position value**, including both-leg fees and a spread allowance. Require net reward at least as large as net stop loss. No trade has a guaranteed profit.
- **Bitkub has NO native stop/OCO.** Protect positions with `client_side_stop_worker`
  (poll it on a loop until `triggered=true`). Never assume a resting stop exists.
- **Never lower a trailing stop.** Move it up only.
- **No averaging down** on a short-term plan.

## Key mechanic: indicators need candles

The analysis tools (`check_market_regime`, `calculate_atr`, `calculate_rsi`,
`detect_breakout_signal`, `detect_pullback_signal`, `calculate_ema`, `calculate_roc`)
take **arrays** (prices or OHLCV candles), not a symbol. So always:

1. `get_historical_candles(symbols=[symbol], format="close", closed_only=true)` returns `result.prices[]` for a single symbol; multi-symbol output uses `result.symbols[symbol]`.
2. Pass numeric price arrays directly into price-based tools (EMA, RSI, ROC, regime), or use `result.candles` for OHLCV tools. Do not stringify prices or candle objects.

Use 240 minutes (4H) for regime, relative strength and volatility fit, and 15 minutes
for entry signals and stops, unless the user specifies other timeframes.

## The flow

### Step 0 — Pre-session portfolio review (always run first)

Before screening for new trades, check what you already hold.

1. `get_trade_history(status_filter=open)` — list all open positions.
2. For each open position:
   - `get_historical_candles(symbols=[symbol], resolution=15, closed_only=true)` → `check_exit_signals(candles)`.
   - If result has `should_exit=true` or ≥ 2 exit reasons → **exit immediately**:
     `place_limit_order(side=sell)` → `log_trade_exit(trade_id, exit_price, exit_reason)`.
   - `check_trade_pnl(symbol, entry, qty, stop)` — note current R to inform next steps.
3. `get_my_open_orders` — collect all symbols that already have a pending limit order on
   the exchange. **Skip the entry step for any symbol already on this list.** This prevents
   duplicate orders when the skill runs on a recurring schedule (e.g. every 2H).
4. If available THB < **200 THB** (the highest `min_quote_size` across all symbols) after
   exits → do **not** open new positions this session. Review only. Do **not** ask the user
   — decide silently based on `get_wallet_balance` result.

Only after all urgent exits are handled, proceed to step 1 for new entries.

---

1. **Shortlist** — `get_market_screener` (volume, spread, depth) or `get_market_overview`
   for context. Scan **all THB pairs**; keep 5–10 most liquid. Never restrict to a single
   symbol unless the user explicitly names one.
2. **Regime filter** — `get_historical_candles` (4H, closed only) → `check_market_regime(prices)`.
   Trade only if UPTREND. (Confirm with `calculate_ema`: close > EMA200, EMA50 > EMA200.)
3. **Relative strength** — `calculate_relative_strength_rank(symbols={sym: prices, ...},
   benchmark=btc_thb)`. `symbols` is an **object** of `symbol: prices[]` (1D close prices
   from `get_historical_candles(..., resolution=240, format="close", closed_only=true)` per symbol), not a plain array of
   symbol names. Keep the top names (`top3`).
4. **Volatility fit** — `calculate_atr(candles)` → keep ATR% in a sane zone (~2–6%/day).
5. **Entry signal** (entry timeframe) — `detect_breakout_signal` (close above 20-bar high
   + volume ≥1.5×) or `detect_pullback_signal` (pullback to EMA20 + RSI bounce 40–50; also
   reports `volume_ratio` — use it for step 7's `volume_ok`, pullback has no built-in
   volume gate the way breakout does).
6. **Size** — `get_wallet_balance` (use `available`, not `total_thb` — that includes funds
   reserved in open orders) → `calculate_position_size(balance, risk_percent=2, entry,
   stop, maker_fee, taker_fee)`. Stop from signal's `suggested_stop` (swing low or
   1.5×ATR). Fees from `get_fee_schedule` chain straight through: its `maker_fee`/
   `taker_fee` are already percentage points (e.g. `0.25` = 0.25%), same unit
   `calculate_position_size` expects.
7. **Gate** — `validate_trade_setup` with the metrics above (its input is
   `position_value_thb`, matching `calculate_position_size`'s output field of the same
   name — not `position_value`). Stop here if `can_trade=false`.
8. **Round** — `get_symbol_rules` → `round_to_exchange_rules(symbol, price, qty)`. Reject
   if `valid=false` (below min notional).
9. **Profit check** — `simulate_trade` with an explicit target for 1–2% net on position value, including both-leg fees and spread. Reject if net reward is smaller than net stop loss.
10. **Place + log** — `place_limit_order(side=buy)` then `log_trade_entry` → keep the
    `trade_id`.
11. **Protect** — arm `client_side_stop_worker(symbol, side=sell, trigger=stop,
    limit_price, qty)` and poll it (e.g. via `/loop`) until triggered.

## Managing an open trade

- **PnL check** — `check_trade_pnl(symbol, entry, qty, stop)` (fetches price; nets fees).
- **At +1R** — move stop to break-even (re-arm `client_side_stop_worker` at entry).
- **At the net target** — take profit; alternatively take partial and switch to `calculate_trailing_stop(candles,
  atr_multiplier=2)`; re-arm the worker at the new level.
- **Emergency exit** — `check_exit_signals(candles)`: STRUCTURE_BREAK or 2+ reasons →
  exit now (`place_limit_order(side=sell)` or let the worker fire).
- **On exit** — `log_trade_exit(trade_id, exit_price, exit_reason)` (computes net PnL + R).

## Review (weekly)

- `get_trade_history(status_filter=closed)` and `calculate_expectancy`. Expectancy =
  win% × avgWinR − (1−win%) × avgLossR. If negative, change ONE variable and re-test.

## Minimum order size per symbol (`min_quote_size`)

The minimum THB amount per buy order (`place_limit_order side=buy`). Source: `GET /api/v3/market/symbols`.

| Min (THB) | Symbols |
|-----------|---------|
| **200 THB** | NEXO |
| **10 THB** | ALL SYMBOL _(all `_THB` pairs)_ |

> Step 8 (`round_to_exchange_rules`) rejects orders below this threshold (`valid=false`). Never place an order smaller than the symbol's `min_quote_size`.

## Notes

- **All tool calls** (including `round_to_exchange_rules`, `get_historical_candles`, `check_trade_pnl`, etc.) must receive symbols in lowercase `coin_thb` format (e.g. `btc_thb`, `eth_thb`). **Never pass `THB_BTC` or uppercase** — the MCP library normalizes the format for the Bitkub API internally; your job is only to pass `coin_thb`.
- `place_limit_order` amount: **buy = THB to spend**, **sell = coin quantity**.
- Fees are charged both legs — the PnL/sizing tools already net them; don't double-count.
- The trade journal is a flat file (`trades.json`, override with env `TRADES_FILE`).
