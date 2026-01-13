# ECS Deployment Guide for go-mcp-atlassian

Deploy go-mcp-atlassian as an HTTP service on AWS ECS (Elastic Container Service) using Fargate or EC2 launch types.

## Quick Reference

| Item | Value |
|------|-------|
| Default Port | 3000 |
| Health Endpoint | `/health` |
| Auth Header | `X-MCP-Auth-Token` |
| Log Location | CloudWatch `/ecs/go-mcp-atlassian` |

## Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   MCP Client    │────▶│  Load Balancer  │────▶│   ECS Service   │
│ (Claude Code/   │     │     (ALB)       │     │   (Fargate/EC2) │
│  Continue.dev)  │     └─────────────────┘     └─────────────────┘
└─────────────────┘              │                       │
                                 │                       ▼
                                 │              ┌─────────────────┐
                                 │              │   Jira /        │
                                 │              │   Confluence    │
                                 │              └─────────────────┘
                                 ▼
                        ┌─────────────────┐
                        │ Secrets Manager │
                        └─────────────────┘
```

## Prerequisites

| Requirement | Description |
|-------------|-------------|
| AWS CLI | Configured with appropriate IAM permissions |
| Docker | Installed locally for building images |
| ECR Repository | Created for storing the Docker image |
| VPC | Subnets configured for ECS deployment |
| Credentials | Atlassian API tokens (Jira and/or Confluence) |

## Deployment Steps

### Step 1: Build and Push Docker Image

```bash
# Set variables
export AWS_REGION="us-east-1"
export AWS_ACCOUNT_ID="123456789012"
export ECR_REPO="go-mcp-atlassian"

# Authenticate to ECR
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin \
  $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Build the image
docker build -t go-mcp-atlassian .

# Tag for ECR
docker tag go-mcp-atlassian:latest \
  $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO:latest

# Push to ECR
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO:latest
```

### Step 2: Create Secrets in AWS Secrets Manager

```bash
aws secretsmanager create-secret \
    --name mcp/atlassian \
    --secret-string '{
        "JIRA_URL": "https://your-domain.atlassian.net",
        "JIRA_USERNAME": "your-email@example.com",
        "JIRA_API_TOKEN": "your-jira-api-token",
        "CONFLUENCE_URL": "https://your-domain.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "your-email@example.com",
        "CONFLUENCE_API_TOKEN": "your-confluence-api-token",
        "MCP_AUTH_TOKEN": "your-secure-auth-token"
    }'
```

### Step 3: Create IAM Task Execution Role

Create a role with this policy to allow ECS to pull images and retrieve secrets:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:BatchGetImage"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "arn:aws:logs:*:*:*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "secretsmanager:GetSecretValue"
            ],
            "Resource": "arn:aws:secretsmanager:YOUR_REGION:YOUR_ACCOUNT_ID:secret:mcp/atlassian*"
        }
    ]
}
```

### Step 4: Create ECS Resources

```bash
# Create CloudWatch Log Group
aws logs create-log-group --log-group-name /ecs/go-mcp-atlassian

# Register Task Definition
aws ecs register-task-definition --cli-input-json file://ecs-task-definition.json

# Create ECS Cluster (if not exists)
aws ecs create-cluster --cluster-name mcp-servers

# Create Service
aws ecs create-service \
    --cluster mcp-servers \
    --service-name go-mcp-atlassian \
    --task-definition go-mcp-atlassian \
    --desired-count 1 \
    --launch-type FARGATE \
    --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}"
```

## Environment Variables Reference

### Required Variables (at least one product)

| Variable | Cloud | Server/DC | Description |
|----------|-------|-----------|-------------|
| `JIRA_URL` | Yes | Yes | Jira instance URL (e.g., `https://company.atlassian.net`) |
| `JIRA_USERNAME` | Yes | No | Email address for Cloud authentication |
| `JIRA_API_TOKEN` | Yes | No | API token for Cloud authentication |
| `JIRA_PERSONAL_TOKEN` | No | Yes | Personal Access Token for Server/DC |
| `CONFLUENCE_URL` | Yes | Yes | Confluence instance URL |
| `CONFLUENCE_USERNAME` | Yes | No | Email address for Cloud authentication |
| `CONFLUENCE_API_TOKEN` | Yes | No | API token for Cloud authentication |
| `CONFLUENCE_PERSONAL_TOKEN` | No | Yes | Personal Access Token for Server/DC |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_AUTH_TOKEN` | None | HTTP authentication token (recommended for production) |
| `MCP_LOG_LEVEL` | `info` | Log level: `off`, `error`, `warn`, `info`, `debug` |
| `READ_ONLY_MODE` | `false` | Disable all write operations |
| `JIRA_SSL_VERIFY` | `true` | Verify SSL certificates for Jira |
| `CONFLUENCE_SSL_VERIFY` | `true` | Verify SSL certificates for Confluence |

## HTTP Authentication

When `MCP_AUTH_TOKEN` is set, all requests must include the authentication header:

```bash
# Example authenticated request
curl -X POST http://your-alb-url:3000/ \
    -H "Content-Type: application/json" \
    -H "X-MCP-Auth-Token: your-secure-auth-token" \
    -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

## Security Checklist

| Requirement | Implementation |
|-------------|----------------|
| HTTPS | Use ALB with HTTPS termination |
| Private Subnets | Deploy ECS tasks in private subnets with NAT Gateway |
| Security Groups | Restrict inbound traffic to ALB security group only |
| Secrets | Use AWS Secrets Manager (never hardcode credentials) |
| Authentication | Enable `MCP_AUTH_TOKEN` in production |
| Read-Only Mode | Consider `READ_ONLY_MODE=true` for safety |

## Monitoring

### CloudWatch Logs

Logs are automatically sent to CloudWatch at `/ecs/go-mcp-atlassian`.

### Health Check

The service exposes a health endpoint:

```bash
curl http://your-alb-url:3000/health
```

Response:
```json
{"status": "healthy", "server": "go-mcp-atlassian"}
```

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| Task fails to start | Configuration error | Check CloudWatch logs for startup errors |
| Health check failures | Network/Security | Verify security group allows inbound on port 3000 |
| Secrets not loading | IAM permissions | Verify task execution role has Secrets Manager access |
| Connection to Atlassian fails | Network | Check VPC has outbound internet access (NAT Gateway) |
| 401 Unauthorized | Auth token mismatch | Verify `X-MCP-Auth-Token` header matches `MCP_AUTH_TOKEN` |

## Related Documentation

- [INTEGRATION.md](./INTEGRATION.md) - Configure Claude Code and Continue.dev clients
- [README.md](./README.md) - Full tool reference and configuration options
