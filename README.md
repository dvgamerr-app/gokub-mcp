# gokub-bot

[![CodeQL](https://github.com/touno-io/gokub-bot/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/touno-io/gokub-bot/actions/workflows/codeql-analysis.yml)

Bitkub MCP Server - Model Context Protocol server for Bitkub Cryptocurrency Exchange API

![Foo](./docs/gokub.png)

## Features

- ✅ **MCP Server** - Built with mcp-go framework
- 🔐 **Secure Authentication** - HMAC SHA256 signature
- 💰 **Get Wallet Balance** - View your Bitkub wallet balances
- 🚀 **Easy Integration** - Works with Claude Desktop and other MCP clients
- 💎 **Go-Bitkub SDK** - Full Bitkub API v3 support

## Prerequisites

- Go 1.21 or higher
- Bitkub API Key และ Secret Key

## Installation

1. Install dependencies:
```bash
go mod download
```

2. สร้างไฟล์ `.env` และใส่ API keys:
```bash
BTK_APIKEY=your_api_key_here
BTK_SECRETKEY=your_secret_key_here
```

## การใช้งาน

### Run MCP Server

```bash
go run main.go
```

### Build

```bash
go build -o bitkub-mcp.exe
./bitkub-mcp.exe
```

## Available Tools

### 1. get_wallet_balance

ดึงข้อมูลยอดเงินในกระเป๋า Bitkub ทั้งหมด

**Parameters:** ไม่มี

**Response Example:**
```
📊 Wallet Balance:

💰 THB:
   Available: 10000.00000000
   Reserved:  0.00000000
   Total:     10000.00000000

💰 BTC:
   Available: 0.00150000
   Reserved:  0.00000000
   Total:     0.00150000

💵 Total THB: 10000.00 THB
```

### 2. get_ticker

ดูราคาปัจจุบันและข้อมูล market ticker

**Parameters:**
- `symbol` (required): Trading pair เช่น `btc_thb`, `eth_thb`, `ada_thb`

**Response Example:**
```
📈 BTC_THB Market Ticker:

💰 Last Price:   2500000.00 THB
📊 24h Volume:   12.3456
📈 24h High:     2550000.00 THB
📉 24h Low:      2480000.00 THB
🔄 24h Change:   1.25%
💵 Best Bid:     2499500.00 THB
💸 Best Ask:     2500500.00 THB
```

### 3. get_market_depth

ดู order book (คำสั่งซื้อขายที่รออยู่)

**Parameters:**
- `symbol` (required): Trading pair
- `limit` (optional): จำนวน orders ที่ต้องการดู (default: 10, max: 100)

**Response Example:**
```
📊 Market Depth for BTC_THB:

📉 ASKS (Sell Orders):
   2505000.00 THB | 0.00120000
   2504000.00 THB | 0.00150000
   ...

━━━━━━━━━━━━━━━━━━━━━━

📈 BIDS (Buy Orders):
   2499000.00 THB | 0.00200000
   2498000.00 THB | 0.00180000
   ...
```

### 4. get_my_open_orders

ดูคำสั่งซื้อขายที่เปิดอยู่ของคุณ

**Parameters:**
- `symbol` (required): Trading pair

**Response Example:**
```
📋 Open Orders for BTC_THB:

1. Order ID: 12345678
   Side: BUY
   Type: limit
   Rate: 2500000.00 THB
   Amount: 0.00100000
   Timestamp: 1730717234567
```

### 5. get_symbols

ดูรายการ trading pairs ทั้งหมดที่พร้อมใช้งาน

**Parameters:** ไม่มี

**Response Example:**
```
📋 Available Trading Pairs:

• BTC_THB
• ETH_THB
• ADA_THB
• XRP_THB
...

Total: 150 active trading pairs
```

## การตั้งค่า API Keys

### วิธีที่ 1: ใช้ไฟล์ .env

สร้างไฟล์ `.env` ใน root directory:
```
BTK_APIKEY=your_api_key
BTK_SECRETKEY=your_secret_key
```

### วิธีที่ 2: Environment Variables

**Windows (PowerShell):**
```powershell
$env:BTK_APIKEY="your_api_key"
$env:BTK_SECRETKEY="your_secret_key"
```

**Linux/Mac:**
```bash
export BTK_APIKEY="your_api_key"
export BTK_SECRETKEY="your_secret_key"
```

## การเชื่อมต่อกับ Claude Desktop

เพิ่มการตั้งค่าใน Claude Desktop config:

**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

**Mac:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "bitkub": {
      "command": "e:\\.dvgamerr\\gokub-mcp\\bitkub-mcp.exe",
      "env": {
        "BTK_APIKEY": "your_api_key",
        "BTK_SECRETKEY": "your_secret_key"
      }
    }
  }
}
```

## Project Structure

```
gokub-mcp/
├── main.go              # MCP Server entry point
├── go.mod               # Go module dependencies
├── go.sum               # Go module checksums
├── README.md            # This file
├── .env                 # API keys (create this file)
└── go-bitkub/          # Bitkub API client
    ├── client.go        # HTTP client
    ├── endpoint.go      # API endpoints
    ├── error.go         # Error handling
    ├── main.go          # Core functions
    └── market.go        # Market API functions
```

## Security Notes

- 🔒 ไม่ควร commit ไฟล์ `.env` ลง git
- 🔐 ใช้ IP whitelist ใน Bitkub API settings
- 🛡️ เก็บ API keys ให้ปลอดภัย
- ⚠️ ไม่แชร์ API keys กับผู้อื่น

## API Rate Limits

- Market Data: 100 req/sec
- Trading Operations: 150-200 req/sec
- ดูข้อมูลเพิ่มเติมที่ [Bitkub API Docs](https://github.com/bitkub/bitkub-official-api-docs)

## References

- [MCP-Go Documentation](https://github.com/mark3labs/mcp-go)
- [Go-Bitkub SDK](https://github.com/dvgamerr-app/go-bitkub)
- [Bitkub Official API](https://github.com/bitkub/bitkub-official-api-docs)
- [Model Context Protocol](https://modelcontextprotocol.io/)

## Community
- [Discord](https://discord.gg/9WSA7mMuGm)

## License

MIT License

## Disclaimer

⚠️ This is an unofficial MCP server. Use at your own risk. Always test thoroughly before using in production.


## TODO
- [x] Bitkub API golang library
- [ ] เริ่มด้วย Rebalancing Bot ก่อนละกันดูจะ ง่ายสุด (In Progress)
- [ ] Grid Trading ยังไม่รู้ทำไง ใครรู้สอนหน่อยสิ

## Features
- Application GUI (`Windows`, `Linux`, `Mac`)
- Support Docker Image
- Support K8s Multiple Deploy

## Ref
- [Official Documentation for Bitkub APIs](https://github.com/bitkub/bitkub-official-api-docs)
