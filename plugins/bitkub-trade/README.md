# bitkub-trade

A Claude Code plugin: the **Bitkub swing-trading playbook** skill for the
[gokub-mcp](https://github.com/dvgamerr-app/gokub-mcp) tools.

It encodes the procedural knowledge an agent needs to trade with the gokub-mcp
tools — the screen → regime → relative-strength → ATR → signal → size →
**validate gate** → round → place → manage → log flow, plus the hard risk
guardrails (≤2% risk per trade, long-only, no entry unless `can_trade=true`,
Bitkub client-side stop, TP ≥2R).

> The tools (the *hands*) live in the gokub-mcp MCP server. This plugin is the
> *brain* — the playbook for using them. Install the MCP server separately.

## Install (Claude Code)

```
/plugin marketplace add dvgamerr-app/gokub-mcp
/plugin install bitkub-trade@gokub-mcp
```

Then the skill is available as `/bitkub-trade:playbook`, and the model will also
invoke it automatically when you ask it to find/evaluate/manage a Bitkub trade.

## Install (Codex)

Codex reads skills from `AGENTS.md` / its skills directory rather than the Claude
plugin marketplace. Point Codex at the same file:

```
# from the repo root, link or copy the skill into Codex's skills path
mkdir -p ~/.codex/skills/bitkub-trade
cp plugins/bitkub-trade/skills/playbook/SKILL.md ~/.codex/skills/bitkub-trade/
```

(The `SKILL.md` body is plain, tool-agnostic Markdown, so it works as a skill in
both Claude Code and Codex.)

## Structure

```
plugins/bitkub-trade/
├── .claude-plugin/plugin.json
├── README.md
└── skills/playbook/SKILL.md
```
