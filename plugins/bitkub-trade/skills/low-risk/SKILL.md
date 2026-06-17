---
name: low-risk
description: Low-risk overlay for Bitkub swing trading. BTC-only, 1% risk per trade, cash reserve enforced. Use when asked to trade safely, trade conservatively, or "low risk". Applies stricter filters on top of the playbook flow — read this first, then follow the playbook.
---

# Low-Risk Trading Rules (Bitkub)

Applies on top of `playbook`. These rules **override** the playbook where they conflict.
When they are silent, the playbook rules apply unchanged.

## Why BTC-only

BTC and ETH have correlation 0.7–0.9. Holding both does not diversify — ETH just
amplifies the drawdown. In a 50% BTC crash, ETH typically falls 30–40% more.
On Bitkub (no short, no hedge), capital preservation = stay in BTC or cash.

## Portfolio allocation (hard rules)

| Bucket | Allocation | Purpose |
|--------|-----------|---------|
| Cash reserve | **≥ 60% of equity** | Dry powder; never touch for a single trade |
| Active capital | ≤ 40% of equity | Only this pool is risked |
| Risk per trade | **≤ 1% of equity** | Half the playbook's 2% |

> Example: 100,000 THB equity → 60,000 cash → 40,000 active → risk ≤ 1,000 THB per trade.

Check `get_wallet_balance` before every session. If cash < 60%, do not open new positions.

## Symbol filter

- **Trade only `btc_thb`.** No ETH, no altcoins.
- Rationale: BTC has the best liquidity on Bitkub, lowest spread, least manipulation.

## Entry checklist (stricter than playbook)

All five must be true before entering. Evaluate on the **4H candle** for entry signal
and **1D candle** for regime confirmation.

| # | Check | Tool |
|---|-------|------|
| 1 | Regime = UPTREND | `check_market_regime` on 1D closes |
| 2 | Close > EMA200 (1D) | `calculate_ema(closes, 200)` |
| 3 | EMA50 > EMA200 (1D) | `calculate_ema(closes, 50)` |
| 4 | RSI(14) > 50 on 4H | `calculate_rsi(closes_4h, 14)` |
| 5 | `validate_trade_setup` returns `can_trade=true` | gate — never skip |

If any check fails → hold cash. Cash is a position.

## Take-profit & stop rules

- **Minimum RR = 2:1.** Same as playbook — do not lower.
- **Stop = below the 4H swing low** that preceded the entry signal.
- At +1R: move stop to break-even.
- At +2R: close at least 50% and trail the rest.
- Maximum loss per session: 3 losing trades → stop for the day, review journal.

## What NOT to touch

- Altcoins below top-10 market cap.
- Any coin with 24h volume < 1M THB on Bitkub.
- Futures or leverage (not available on Bitkub spot, and not wanted here).
- Entries based on influencer calls or news alone.

## Execution flow

After passing the checklist above, hand off to the **playbook** starting at step 0
(Pre-session portfolio review), then skip steps 1–3 (Shortlist, Regime, Relative Strength)
since we are BTC-only and the checklist above replaces them.

→ Continue with: `plugins/bitkub-trade/skills/playbook/SKILL.md` **Step 0**, then step 4 onward.
