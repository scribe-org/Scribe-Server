# Scribe Server API MCP server

Scribe Server API is available as a hosted [Model Context Protocol](https://modelcontextprotocol.io) (MCP)
server, so AI agents such as Claude, ChatGPT, Cursor and VS Code can call it as tools.

| | |
| --- | --- |
| Server URL | `https://app.getelva.ai/mcp/elvademo/scribe-server-api` |
| Install page | https://app.getelva.ai/mcp/elvademo/scribe-server-api/install |
| Tools | 6, one per operation in [`swagger.yaml`](docs/swagger.yaml) |
| Calls | `https://scribe-server.toolforge.org` |
| Authentication | None, the API is public. |

## Install

**[Open the install page](https://app.getelva.ai/mcp/elvademo/scribe-server-api/install)** for step-by-step setup in Claude, ChatGPT,
Cursor, VS Code and other clients, or use one of the options below.

[![Install in Cursor](https://cursor.com/deeplink/mcp-install-dark.svg)](https://cursor.com/en/install-mcp?name=Scribe%20Server%20API&config=eyJ1cmwiOiJodHRwczovL2FwcC5nZXRlbHZhLmFpL21jcC9lbHZhZGVtby9zY3JpYmUtc2VydmVyLWFwaSJ9)
[![Install in VS Code](https://img.shields.io/badge/VS_Code-Install_MCP-0098FF?logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=scribe-server-api&config=%7B%22type%22%3A%22http%22%2C%22url%22%3A%22https%3A%2F%2Fapp.getelva.ai%2Fmcp%2Felvademo%2Fscribe-server-api%22%7D)

### Claude Code

```bash
claude mcp add --transport http scribe-server-api https://app.getelva.ai/mcp/elvademo/scribe-server-api
```

### Claude (claude.ai and Claude Desktop)

Settings → Connectors → **Add custom connector**, then paste the server URL.

### ChatGPT

Settings → Apps & Connectors → Advanced settings → turn on **Developer mode**, then
**Create** a connector and paste the server URL.

### Any other MCP client

```json
{
  "mcpServers": {
    "scribe-server-api": {
      "url": "https://app.getelva.ai/mcp/elvademo/scribe-server-api"
    }
  }
}
```

---

Generated from this repository's OpenAPI spec and hosted by [Elva](https://getelva.ai?utm_source=github&utm_medium=scribe-org/Scribe-Server&utm_campaign=githubprs).
