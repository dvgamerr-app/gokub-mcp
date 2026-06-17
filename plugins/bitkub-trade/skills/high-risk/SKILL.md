---
name: high-risk
description: High-risk overlay for Bitkub swing trading. Up to 3% risk per trade, BTC + ETH + top-10 altcoins allowed, smaller cash reserve, shorter entry timeframe. Use when asked to trade aggressively, maximize returns, or "high risk". Applies looser filters on top of the playbook — read this first, then follow the playbook. Always confirm with the user before placing a real order, unless the user has explicitly waived confirmation for this session.
---

# High-Risk Trading Rules (Bitkub)

Applies on top of `playbook`. These rules **override** the playbook where they conflict.
When they are silent, the playbook rules apply unchanged.

## Confirmation gate (mandatory unless waived)

**Before every `place_limit_order` call**, show the user a trade summary and wait for
explicit approval. Do not call the order tool until the user says yes.

Summary format:

```
📋 Trade Summary — [SYMBOL]
  Side       : BUY
  Entry      : [price] THB
  Stop Loss  : [price] THB  (−[R] THB / −[%]%)
  Take Profit: [price] THB  (+[2R] THB / +[%]%)
  Qty        : [qty] coins
  Risk       : [THB] THB ([%]% of equity)

ยืนยันซื้อจริงไหมครับ? (ตอบ ใช่ / ยืนยัน / go / yes เพื่อดำเนินการ)
```

**Waiver** — skip the gate and execute immediately if the user has said any of:
- "ไม่ต้อง confirm", "ยืนยันได้เลย", "ทำงานได้เลย", "auto", "no confirm", "just do it", "go ahead"

The waiver lasts for the current session only. Re-ask on next session.

## Why high-risk is still managed risk

High-risk does not mean reckless. It means accepting more volatility for higher
potential reward. Risk per trade is capped, correlation is known, and the validation
gate is never skipped. The difference from low-risk is looser filters and larger
position sizing — not abandoning the system.

## Portfolio allocation

| Bucket | Allocation | Purpose |
|--------|-----------|---------|
| Cash reserve | **≥ 30% of equity** | Minimum floor; never go all-in |
| Active capital | ≤ 70% of equity | Tradeable pool |
| Risk per trade | **≤ 3% of equity** | Hard ceiling |
| Max open positions | 3 concurrent | Avoid overexposure |

> Example: 100,000 THB → 30,000 cash → 70,000 active → risk ≤ 3,000 THB per trade.

Check `get_wallet_balance` before every session. If cash < 30%, reduce or close a
position before opening a new one.

## Symbol filter

Allowed (in priority order):

| Priority | Symbol | Rationale |
|----------|--------|-----------|
| 1st | `btc_thb` | Highest liquidity, lowest spread |
| 2nd | `eth_thb` | High liquidity; accepted correlation risk |
| 3rd | Top-10 by Bitkub 24h THB volume | Screened per session |

Excluded regardless:
- 24h volume < 5M THB on Bitkub
- Any coin that moved > 20% in the last 24h (chasing momentum after the fact)
- Meme coins, newly listed coins (< 30 days)

## Entry checklist

Evaluate on the **1H candle** for entry signal and **4H candle** for regime. Fewer
conditions than low-risk, but `validate_trade_setup` gate is still mandatory.

| # | Check | Tool |
|---|-------|------|
| 1 | Regime = UPTREND **or** NEUTRAL | `check_market_regime` on 4H closes |
| 2 | Close > EMA50 (4H) | `calculate_ema(closes_4h, 50)` |
| 3 | RSI(14) between 45–75 on 1H | `calculate_rsi(closes_1h, 14)` — avoid overbought chase |
| 4 | `validate_trade_setup` returns `can_trade=true` | gate — never skip |

DOWNTREND regime → hold cash regardless of other signals.

## Take-profit & stop rules

- **Minimum RR = 2:1.** Same floor as low-risk and playbook.
- **Stop = 1.5× ATR below entry** (from 1H ATR) or below 1H swing low, whichever is tighter.
- At +1R: move stop to break-even.
- At +2R: take 50%, trail the rest with `calculate_trailing_stop(candles, atr_multiplier=1.5)`.
- At +3R: close remaining or continue trailing at atr_multiplier=1.
- Maximum loss per session: **2 losing trades** → stop for the day, review journal.
  (Tighter than low-risk because single losses are larger.)

## Execution flow

Follow the full **playbook** flow. No shortcuts.

Steps 1–3 (Shortlist, Regime, Relative Strength) use this skill's looser filters.
Steps 4–11 follow the playbook exactly, with one addition:

**Between step 9 (simulate) and step 10 (place + log):**
→ Show the confirmation summary above and wait for user approval (unless waived).

→ Continue with: `plugins/bitkub-trade/skills/playbook/SKILL.md` step 4 onward.
