# MCP Client Integration Guide

This guide explains how to configure MCP clients (Claude Code and Continue.dev) to connect to the go-mcp-atlassian server running in HTTP mode, including authentication configuration.

## Authentication Overview

When running in HTTP mode with authentication enabled (via `MCP_AUTH_TOKEN` environment variable), all requests must include the `X-MCP-Auth-Token` header with the configured token value.

## Claude Code Integration

### Configuration Location

Claude Code configuration is stored in:
- **macOS/Linux**: `~/.claude/claude_code_config.json`
- **Windows**: `%USERPROFILE%\.claude\claude_code_config.json`

Alternatively, project-level configuration in `.mcp.json` in your project root.

### HTTP Mode Configuration

Add the following to your Claude Code configuration:

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

### Configuration with Environment Variable for Token

For better security, you can reference environment variables:

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

### Local Development (stdio mode)

For local development without HTTP:

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

## Continue.dev Integration

### Configuration Location

Continue.dev configuration is stored in:
- **macOS/Linux**: `~/.continue/config.json`
- **Windows**: `%USERPROFILE%\.continue\config.json`

### HTTP Mode Configuration

Add the MCP server to your Continue.dev configuration:

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

### Local Development (stdio mode)

For local development without HTTP:

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

## Testing the Connection

### Using curl

```bash
# Test health endpoint (no auth required)
curl http://your-alb-url:3000/health

# Test MCP endpoint with auth
curl -X POST http://your-alb-url:3000/ \
    -H "Content-Type: application/json" \
    -H "X-MCP-Auth-Token: your-secure-auth-token" \
    -d '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}},"id":1}'

# List available tools
curl -X POST http://your-alb-url:3000/ \
    -H "Content-Type: application/json" \
    -H "X-MCP-Auth-Token: your-secure-auth-token" \
    -d '{"jsonrpc":"2.0","method":"tools/list","id":2}'
```

### Expected Responses

Health check:
```json
{"status": "healthy", "server": "go-mcp-atlassian"}
```

Unauthorized (missing/invalid token):
```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {"code": -32001, "message": "Unauthorized: invalid or missing authentication token"}
}
```

## Security Best Practices

1. **Use HTTPS in production**: Always use HTTPS (via ALB/CloudFront) for production deployments
2. **Rotate tokens regularly**: Implement token rotation policies
3. **Use environment variables**: Never hardcode tokens in configuration files committed to version control
4. **Restrict token scope**: Use different tokens for different environments/users
5. **Enable Read-Only Mode**: Use `READ_ONLY_MODE=true` to prevent write operations

## Troubleshooting

### Connection Refused
- Verify the server is running and accessible
- Check security group rules allow traffic on port 3000

### 401 Unauthorized
- Verify the `X-MCP-Auth-Token` header is set correctly
- Ensure the token matches the `MCP_AUTH_TOKEN` environment variable on the server

### Timeout Errors
- Increase client timeout settings
- Check for network latency issues
