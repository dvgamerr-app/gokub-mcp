---
name: playbook
description: >-
  Playbook for swing-trading the long side on Bitkub (THB pairs) using the gokub-mcp tools. Use when asked to find a trade, screen coins, evaluate an entry, size a position, place/manage an order, or review trading performance on Bitkub. Encodes the risk rules, the screen→regime→signal→size→validate→place→manage→log flow, and Bitkub-specific constraints (no native stop/OCO).
license: MIT
author: Kananek T.
version: 0.4
category: trading
---

# Bitkub Trade Playbook

Long-only swing trading on Bitkub THB pairs via the gokub-mcp tools. The tools are
the hands; this is the brain. Follow the flow, respect the guardrails, never skip the
validation gate.

## Execution discipline (never violate)

- **Run the full flow autonomously.** Never pause mid-flow to ask whether to continue. Screen → regime → signal → size → validate → decide — all in one pass.
- **Report results, not options.** End with a decision table (pass/fail per coin, with entry/stop/TP for any that pass) and the reason. Never end with "ถ้าต้องการจะรันต่อ…" or any conditional offer.
- **No raw data dumps.** Never return raw candle arrays or large result blobs to the user. Summarize each intermediate step in ≤ 1 line.

## Hard guardrails (never violate)

- **Risk ≤ 2% of equity per trade.** Size from the stop, not from a fixed amount.
- **Long only, uptrend only.** No shorting, no counter-trend entries.
- **No entry unless `validate_trade_setup` returns `can_trade=true`.** This is the gate.
- **Take-profit ≥ 2R.** If R:R < 2, skip the trade.
- **Bitkub has NO native stop/OCO.** Protect positions with `client_side_stop_worker`
  (poll it on a loop until `triggered=true`). Never assume a resting stop exists.
- **Never lower a trailing stop.** Move it up only.
- **No averaging down** on a short-term plan.

## Key mechanic: indicators need candles

The analysis tools (`check_market_regime`, `calculate_atr`, `calculate_rsi`,
`detect_breakout_signal`, `detect_pullback_signal`, `calculate_ema`, `calculate_roc`)
take **arrays** (prices or OHLCV candles), not a symbol. So always:

1. `get_historical_candles(symbols=[symbol], format="close")` — returns `result.prices[]` directly.
2. Pass `result.prices` into price-based tools (EMA, RSI, ROC, regime) or use `result.candles` (default format) for OHLCV tools (ATR, breakout, pullback, exit signals).

Use the big timeframe (1D / 240) for regime + relative strength, the entry timeframe
(60 / 15) for ATR + entry signals.

## The flow

### Step 0 — Pre-session portfolio review (always run first)

Before screening for new trades, check what you already hold.

1. `get_trade_history(status_filter=open)` — list all open positions.
2. For each open position:
   - `get_historical_candles(symbol, interval=60)` → `check_exit_signals(candles)`.
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
2. **Regime filter** — `get_historical_candles` (1D) → `check_market_regime(prices)`.
   Trade only if UPTREND. (Confirm with `calculate_ema`: close > EMA200, EMA50 > EMA200.)
3. **Relative strength** — `calculate_relative_strength_rank(symbols[], benchmark=btc_thb)`.
   Keep the top names.
4. **Volatility fit** — `calculate_atr(candles)` → keep ATR% in a sane zone (~2–6%/day).
5. **Entry signal** (entry timeframe) — `detect_breakout_signal` (close above 20-bar high
   + volume ≥1.5×) or `detect_pullback_signal` (pullback to EMA20 + RSI bounce 40–50).
6. **Size** — `get_wallet_balance` → `calculate_position_size(balance, risk_percent=2,
   entry, stop)`. Stop from signal's `suggested_stop` (swing low or 1.5×ATR).
7. **Gate** — `validate_trade_setup` with the metrics above. Stop here if `can_trade=false`.
8. **Round** — `get_symbol_rules` → `round_to_exchange_rules(symbol, price, qty)`. Reject
   if `valid=false` (below min notional).
9. **(Optional) sanity** — `simulate_trade` to confirm R:R ≥ 2 before committing.
10. **Place + log** — `place_limit_order(side=buy)` then `log_trade_entry` → keep the
    `trade_id`.
11. **Protect** — arm `client_side_stop_worker(symbol, side=sell, trigger=stop,
    limit_price, qty)` and poll it (e.g. via `/loop`) until triggered.

## Managing an open trade

- **PnL check** — `check_trade_pnl(symbol, entry, qty, stop)` (fetches price; nets fees).
- **At +1R** — move stop to break-even (re-arm `client_side_stop_worker` at entry).
- **At +2R** — take partial and/or switch to `calculate_trailing_stop(candles,
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
