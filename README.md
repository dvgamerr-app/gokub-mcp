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
- ✅ **MCP Server** - Built with mcp-go framework
- 🌐 **HTTP/SSE Transport** - Real-time communication
- 🔐 **Secure Authentication** - HMAC SHA256 signature
- 💰 **Wallet Management** - View balances & transactions

</td>
<td width="50%">

### 🚀 Developer Experience
- 💎 **Go-Bitkub SDK** - Full API v3 support
- � **Easy Integration** - Works with Claude Desktop
- 📊 **Market Data** - Real-time ticker & depth
- � **Order Management** - Track open orders

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


1. `get_wallet_balance`
2. `get_ticker`
3. `get_market_depth`
4. `get_my_open_orders`
5. `get_symbols`


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
├── 📄 main.go              # MCP Server entry point (HTTP/SSE)
├── 📂 prompts/             # Trading prompts
├── 📂 resources/           # Market resources
├── 📂 tools/               # MCP tools implementation
└── 📂 utils/               # Utility functions
```

## 📊 API Rate Limits

| Category | Rate Limit | Note |
|----------|------------|------|
| 📈 Market Data | 100 req/sec | Public endpoints |
| 💱 Trading Operations | 150-200 req/sec | Authenticated endpoints |

> 📚 [Bitkub API Docs](https://github.com/bitkub/bitkub-official-api-docs) สำหรับข้อมูลเพิ่มเติม

## 🚀 Roadmap

### ✅ Completed
- [x] Bitkub API golang library
- [x] MCP Server implementation
- [x] HTTP/SSE transport
- [x] Basic wallet & market tools

### 🚧 In Progress
- [ ] Rebalancing Bot
- [ ] Grid Trading strategy
- [ ] Advanced order management

### 🎯 Planned Features
- [ ] Docker Image support
- [ ] Kubernetes deployment
- [ ] WebSocket real-time data
- [ ] Trading bot framework

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
