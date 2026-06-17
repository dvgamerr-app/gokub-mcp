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

The repo doubles as a Codex marketplace (`.agents/plugins/marketplace.json`):

```
codex plugin marketplace add dvgamerr-app/gokub-mcp
codex plugin add bitkub-trade@gokub-mcp
```

Or install just the skill (no plugin manifest needed) via the `skill-installer`:

```
# install-skill-from-github.py --repo dvgamerr-app/gokub-mcp \
#   --path plugins/bitkub-trade/skills/playbook --name bitkub-trade
```

Start a new thread afterwards so Codex picks up the new skill.

## Structure (cross-tool)

```
gokub-mcp/
├── .claude-plugin/marketplace.json     # Claude marketplace
├── .agents/plugins/marketplace.json    # Codex marketplace
└── plugins/bitkub-trade/
    ├── .claude-plugin/plugin.json      # Claude plugin manifest
    ├── .codex-plugin/plugin.json       # Codex plugin manifest
    ├── README.md
    └── skills/playbook/SKILL.md        # shared skill (both tools)
```
