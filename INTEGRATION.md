# MCP Client Integration Guide

Configure MCP clients (Claude Code, Claude Desktop, and Continue.dev) to connect to go-mcp-atlassian.

## Quick Reference

| Transport Mode | Use Case | Auth Required |
|----------------|----------|---------------|
| stdio | Local development | No |
| HTTP | Remote/ECS deployment | Recommended (`X-MCP-Auth-Token`) |

## Configuration Locations

| Client | Platform | Config File |
|--------|----------|-------------|
| Claude Code | macOS/Linux | `~/.claude/claude_code_config.json` |
| Claude Code | Windows | `%USERPROFILE%\.claude\claude_code_config.json` |
| Claude Code | Project | `.mcp.json` in project root |
| Claude Desktop | macOS/Linux | `~/.config/claude/claude_desktop_config.json` |
| Claude Desktop | Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Continue.dev | macOS/Linux | `~/.continue/config.json` |
| Continue.dev | Windows | `%USERPROFILE%\.continue\config.json` |

## Claude Code / Claude Desktop Configuration

### Local Mode (stdio) - Recommended for Development

```json
{
  "mcpServers": {
    "atlassian": {
      "command": "/path/to/go-mcp-atlassian",
      "args": [],
      "env": {
        "JIRA_URL": "https://your-domain.atlassian.net",
        "JIRA_USERNAME": "your-email@example.com",
        "JIRA_API_TOKEN": "your-api-token",
        "CONFLUENCE_URL": "https://your-domain.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "your-email@example.com",
        "CONFLUENCE_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

### HTTP Mode - For Remote/ECS Deployments

```json
{
  "mcpServers": {
    "atlassian": {
      "type": "http",
      "url": "http://your-alb-url:3000",
      "headers": {
        "X-MCP-Auth-Token": "your-secure-auth-token"
      }
    }
  }
}
```

### HTTP Mode with Environment Variable (More Secure)

```json
{
  "mcpServers": {
    "atlassian": {
      "type": "http",
      "url": "http://your-alb-url:3000",
      "headers": {
        "X-MCP-Auth-Token": "${MCP_ATLASSIAN_TOKEN}"
      }
    }
  }
}
```

Then set the environment variable:
```bash
export MCP_ATLASSIAN_TOKEN="your-secure-auth-token"
```

## Continue.dev Configuration

### Local Mode (stdio)

```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "name": "atlassian",
        "transport": {
          "type": "stdio",
          "command": "/path/to/go-mcp-atlassian",
          "args": []
        },
        "env": {
          "JIRA_URL": "https://your-domain.atlassian.net",
          "JIRA_USERNAME": "your-email@example.com",
          "JIRA_API_TOKEN": "your-api-token"
        }
      }
    ]
  }
}
```

### HTTP Mode

```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "name": "atlassian",
        "transport": {
          "type": "http",
          "url": "http://your-alb-url:3000",
          "headers": {
            "X-MCP-Auth-Token": "your-secure-auth-token"
          }
        }
      }
    ]
  }
}
```

## Testing the Connection

### Health Check (No Auth Required)

```bash
curl http://your-alb-url:3000/health
```

Expected response:
```json
{"status": "healthy", "server": "go-mcp-atlassian"}
```

### Initialize MCP Session

```bash
curl -X POST http://your-alb-url:3000/ \
    -H "Content-Type: application/json" \
    -H "X-MCP-Auth-Token: your-secure-auth-token" \
    -d '{
      "jsonrpc": "2.0",
      "method": "initialize",
      "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {},
        "clientInfo": {"name": "test", "version": "1.0.0"}
      },
      "id": 1
    }'
```

### List Available Tools

```bash
curl -X POST http://your-alb-url:3000/ \
    -H "Content-Type: application/json" \
    -H "X-MCP-Auth-Token: your-secure-auth-token" \
    -d '{"jsonrpc":"2.0","method":"tools/list","id":2}'
```

## Common Response Codes

| Response | Meaning | Action |
|----------|---------|--------|
| Health: `{"status": "healthy"}` | Server is running | Connection successful |
| Error: `-32001 Unauthorized` | Invalid/missing auth token | Check `X-MCP-Auth-Token` header |
| Error: `Connection refused` | Server not running | Verify server is running and accessible |
| Error: `Timeout` | Network latency | Increase client timeout settings |

## Security Best Practices

| Practice | Description |
|----------|-------------|
| Use HTTPS | Always use HTTPS (via ALB/CloudFront) in production |
| Environment Variables | Never hardcode tokens in version-controlled configs |
| Token Rotation | Implement regular token rotation policies |
| Separate Tokens | Use different tokens per environment/user |
| Read-Only Mode | Use `READ_ONLY_MODE=true` to prevent accidental writes |

## Troubleshooting

| Issue | Possible Cause | Solution |
|-------|----------------|----------|
| Connection Refused | Server not running | Verify server is running and port is accessible |
| 401 Unauthorized | Token mismatch | Verify `X-MCP-Auth-Token` matches server's `MCP_AUTH_TOKEN` |
| Timeout Errors | Network latency | Increase client timeout; check network routing |
| Tools not appearing | Config syntax error | Validate JSON syntax; check client logs |
| SSL/TLS errors | Certificate issues | For Server/DC, check `*_SSL_VERIFY` settings |

## Related Documentation

- [README.md](./README.md) - Full tool reference and environment variables
- [ECS.md](./ECS.md) - AWS ECS deployment guide
