import assert from 'node:assert/strict';
import { mkdirSync, mkdtempSync, rmdirSync, readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { resolve, join, dirname } from 'node:path';
import { tmpdir } from 'node:os';

const binary = process.argv[2];
if (!binary) throw new Error('Usage: bun scripts/research-chart-strategies.mjs <server-exe> [saved-evaluation.json]');
const root = resolve(import.meta.dir, '..');
const saved = process.argv[3] ? await Bun.file(resolve(process.argv[3])).json() : null;
const runtime = mkdtempSync(join(tmpdir(), 'gokub-chart-research-'));
const server = Bun.spawn([resolve(binary)], {
  cwd: runtime,
  env: { ...process.env, BTK_APIKEY: '', BTK_SECRET: '', LOG_LEVEL: 'error' },
  stdin: 'pipe', stdout: 'pipe', stderr: 'pipe',
});
const pending = new Map();
let sequence = 0;
const stderr = new Response(server.stderr).text();
const output = (async () => {
  const decoder = new TextDecoder();
  let buffer = '';
  for await (const bytes of server.stdout) {
    buffer += decoder.decode(bytes, { stream: true });
    let newline;
    while ((newline = buffer.indexOf('\n')) >= 0) {
      const line = buffer.slice(0, newline).trim();
      buffer = buffer.slice(newline + 1);
      if (!line) continue;
      const message = JSON.parse(line);
      const waiter = pending.get(message.id);
      if (waiter) {
        pending.delete(message.id);
        clearTimeout(waiter.timeout);
        if (message.error) waiter.reject(new Error(JSON.stringify(message.error)));
        else waiter.resolve(message.result);
      }
    }
  }
  for (const waiter of pending.values()) { clearTimeout(waiter.timeout); waiter.reject(new Error('Server stdout closed before response')); }
  pending.clear();
})();
async function rpc(method, params) {
  const id = ++sequence;
  const response = new Promise((resolve, reject) => {
    const timeout = setTimeout(() => { pending.delete(id); reject(new Error(`Timed out: ${method}`)); }, 90000);
    pending.set(id, { resolve, reject, timeout });
  });
  server.stdin.write(JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n');
  await server.stdin.flush();
  return response;
}
async function call(name, args) {
  const result = await rpc('tools/call', { name, arguments: args });
  if (result.isError) throw new Error(result.content?.map(part => part.text ?? '').join('\n') || `${name} failed`);
  assert.ok(result.structuredContent, `${name} returned no structured data`);
  return result.structuredContent;
}
const strategies = ['breakout', 'pullback', 'macd_cross', 'supertrend_flip', 'bollinger_breakout', 'ema_crossover'];
const costs = [{ fee_percent: .25, slippage_percent: .1 }, { fee_percent: .25, slippage_percent: .2 }];
const report = {
  captured_at: Math.floor(Date.now() / 1000),
  as_of: saved?.as_of ?? Math.floor(Date.now() / 1000) - 60,
  mode: saved ? 'saved_snapshot_replay' : 'public_market_snapshot',
  market_source: 'Bitkub public get_historical_candles MCP tool; no account or order tools',
  scope: 'Descriptive fixed-rule comparison. Same saved candles and evaluation cutoffs; not parameter selection, portfolio validation, TradingView numerical parity or proof of a best indicator.',
  datasets: [], results: [], errors: [],
};
const fingerprint = createHash('sha256');
for (const file of ['main.go', 'tools/analyze_chart.go', 'tools/check_market_regime.go', 'tools/backtest_signals.go', 'tools/detect_breakout_signal.go', 'tools/detect_pullback_signal.go', 'tools/calculate_ema.go', 'tools/calculate_atr.go', 'tools/calculate_rsi.go', 'tools/candles.go', 'utils/args.go']) {
  fingerprint.update(file + '\n');
  fingerprint.update(readFileSync(join(root, file)));
}
report.source_sha256 = fingerprint.digest('hex');
try {
  const initialized = await rpc('initialize', { protocolVersion: '2025-06-18', capabilities: {}, clientInfo: { name: 'chart-research', version: '1.0.0' } });
  report.server_info = initialized.serverInfo;
  report.protocol_version = initialized.protocolVersion;
  server.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized' }) + '\n');
  await server.stdin.flush();
  const inventory = await rpc('tools/list', {});
  report.tool_count = inventory.tools.length;
  console.log(`MCP initialized: ${initialized.serverInfo.name}; ${inventory.tools.length} tools`);
  for (const name of ['get_historical_candles', 'analyze_chart', 'backtest_signals']) assert.ok(inventory.tools.some(tool => tool.name === name), `${name} is not registered`);
  const inputs = saved?.datasets ?? ['btc_thb', 'eth_thb'].flatMap(symbol => [15, 240].map(resolution => ({ symbol, resolution })));
  for (const input of inputs) {
    const { symbol, resolution } = input;
    try {
      const history = input.candles ? input : await call('get_historical_candles', { symbols: [symbol], resolution, limit: 1000, closed_only: true, as_of: report.as_of });
      const dataset = { symbol, resolution, candles: history.candles, history_metadata: history.history_metadata };
      assert.ok(dataset.candles.length >= 221, 'History is too short for evaluation');
      report.datasets.push(dataset);
      const chart = await call('analyze_chart', { candles: dataset.candles, resolution, as_of: report.as_of });
      for (const cost of costs) {
        for (const strategy of strategies) {
          const result = await call('backtest_signals', { candles: dataset.candles, resolution, as_of: report.as_of, strategy, folds: 3, reward_risk: 2, ...cost });
          report.results.push({ symbol, resolution, strategy, ...cost, chart, result });
        }
      }
      console.log(`${symbol} ${resolution}m: ${dataset.candles.length} candles, ${strategies.length * costs.length} evaluations`);
    } catch (error) {
      report.errors.push({ symbol, resolution, error: String(error) });
      console.error(`${symbol} ${resolution}m: ${error}`);
    }
  }
  const path = join(root, 'docs', 'research', 'tradingview-evaluation.json');
  mkdirSync(dirname(path), { recursive: true });
  await Bun.write(path, JSON.stringify(report));
  console.log(`Saved ${report.results.length} evaluations and ${report.errors.length} errors to ${path}`);
  if (report.errors.length || report.results.length === 0) process.exitCode = 1;
} finally {
  server.stdin.end();
  const timer = setTimeout(() => server.kill(), 5000);
  const code = await server.exited;
  clearTimeout(timer);
  await output;
  const diagnostic = await stderr;
  if (code !== 0 && diagnostic) console.error(diagnostic.slice(0, 2000));
  rmdirSync(runtime);
}
