# ECS Deployment Guide for go-mcp-atlassian

This guide covers deploying go-mcp-atlassian as an HTTP service on AWS ECS (Elastic Container Service) using either Fargate or EC2 launch types.

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

1. AWS CLI configured with appropriate permissions
2. Docker installed locally for building images
3. An ECR repository created for the image
4. VPC with subnets configured for ECS
5. Atlassian credentials (Jira and/or Confluence)

## Quick Start

### 1. Build and Push Docker Image

```bash
# Authenticate to ECR
aws ecr get-login-password --region YOUR_REGION | docker login --username AWS --password-stdin YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com

# Build the image
docker build -t go-mcp-atlassian .

# Tag for ECR
docker tag go-mcp-atlassian:latest YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/go-mcp-atlassian:latest

# Push to ECR
docker push YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/go-mcp-atlassian:latest
```

### 2. Create Secrets in AWS Secrets Manager

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

### 3. Create IAM Roles

#### Task Execution Role

This role allows ECS to pull images and retrieve secrets:

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

### 4. Create ECS Resources

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

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `JIRA_URL` | Yes* | Jira instance URL |
| `JIRA_USERNAME` | Yes* | Jira username (for Cloud) |
| `JIRA_API_TOKEN` | Yes* | Jira API token (for Cloud) |
| `JIRA_PERSONAL_TOKEN` | Yes* | Jira PAT (for Server/DC) |
| `CONFLUENCE_URL` | Yes* | Confluence instance URL |
| `CONFLUENCE_USERNAME` | Yes* | Confluence username (for Cloud) |
| `CONFLUENCE_API_TOKEN` | Yes* | Confluence API token (for Cloud) |
| `CONFLUENCE_PERSONAL_TOKEN` | Yes* | Confluence PAT (for Server/DC) |
| `MCP_AUTH_TOKEN` | No | Token for HTTP authentication |
| `MCP_LOG_LEVEL` | No | Log level (default: info) |
| `READ_ONLY_MODE` | No | Enable read-only mode |

*At least one of Jira or Confluence credentials required.

### Authentication

When `MCP_AUTH_TOKEN` is set, all HTTP requests to the MCP server must include the `X-MCP-Auth-Token` header with the matching token value.

```bash
# Example request with authentication
curl -X POST http://your-alb-url:3000/ \
    -H "Content-Type: application/json" \
    -H "X-MCP-Auth-Token: your-secure-auth-token" \
    -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

## Security Considerations

1. **Use HTTPS**: Place an ALB with HTTPS termination in front of the ECS service
2. **Private Subnets**: Deploy ECS tasks in private subnets with NAT Gateway
3. **Security Groups**: Restrict inbound traffic to only the ALB security group
4. **Secrets Management**: Always use AWS Secrets Manager for credentials
5. **Authentication**: Enable `MCP_AUTH_TOKEN` for production deployments
6. **Read-Only Mode**: Consider enabling `READ_ONLY_MODE=true` for safety

## Monitoring

### CloudWatch Logs

Logs are automatically sent to CloudWatch Logs at `/ecs/go-mcp-atlassian`.

### Health Checks

The service exposes a `/health` endpoint that returns:
```json
{"status": "healthy", "server": "go-mcp-atlassian"}
```

## Troubleshooting

### Common Issues

1. **Task fails to start**: Check CloudWatch logs for startup errors
2. **Health check failures**: Ensure security group allows inbound on port 3000
3. **Secrets not loading**: Verify task execution role has Secrets Manager permissions
4. **Connection to Atlassian fails**: Check VPC has outbound internet access

See [INTEGRATION.md](./INTEGRATION.md) for configuring Claude Code and Continue.dev to connect to this service.
