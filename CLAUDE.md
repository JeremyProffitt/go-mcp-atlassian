# MCP Server LLM Usability Checklist

This checklist ensures all tools in this MCP server are optimized for LLM consumption. Review all items before merging changes.

## Quick Verification Commands

```bash
# Verify compilation
go build ./...

# Run all tests
go test ./...

# Check formatting
go fmt ./...
```

## Checklist Categories

### 1. Tool Definitions

| Requirement | Status | Description |
|-------------|--------|-------------|
| Clear Purpose | [ ] | Each tool description explains what it does and when to use it |
| No Redundant Names | [ ] | Descriptions omit unnecessary platform prefixes |
| Parameter Hints | [ ] | Descriptions mention key parameters or capabilities |
| Use Case Guidance | [ ] | Complex tools include guidance vs similar tools |
| Consistent Naming | [ ] | All tools use `snake_case` naming convention |
| Action Verbs | [ ] | Tool names start with verbs: `get_`, `list_`, `create_`, `update_`, `delete_`, `search_` |

### 2. Parameter Documentation

| Requirement | Status | Description |
|-------------|--------|-------------|
| Examples Provided | [ ] | String parameters include format examples |
| Format Hints | [ ] | Date/time, ID, structured params document formats |
| Valid Values | [ ] | Fixed-option params list valid values |
| Default Location | [ ] | Defaults in Default field, not description |
| Array Format | [ ] | Array params explain expected item format |
| Object Structure | [ ] | Object params describe expected properties |

### 3. Schema Constraints

| Requirement | Status | Description |
|-------------|--------|-------------|
| Numeric Bounds | [ ] | Limit/offset/count params have Min/Max constraints |
| Integer Types | [ ] | Pagination/count params use "integer" not "number" |
| Enum Values | [ ] | Categorical params have Enum arrays defined |
| Array Items Typed | [ ] | All array params have Items property with type |
| Object Properties | [ ] | Complex objects have Properties defined |
| Pattern Validation | [ ] | ID fields have Pattern regex (optional) |

### 4. Tool Annotations

| Requirement | Status | Description |
|-------------|--------|-------------|
| Title Set | [ ] | All tools have human-readable Title annotation |
| ReadOnlyHint | [ ] | `get_*`, `list_*`, `search_*`, `describe_*` have `ReadOnlyHint: true` |
| DestructiveHint | [ ] | `delete_*` tools have `DestructiveHint: true` |
| IdempotentHint | [ ] | Safe-to-retry operations have `IdempotentHint: true` |
| OpenWorldHint | [ ] | External system tools have `OpenWorldHint: true` (optional) |

### 5. Token Efficiency

| Requirement | Status | Description |
|-------------|--------|-------------|
| Concise Descriptions | [ ] | Tool descriptions under 200 characters |
| No Duplicate Info | [ ] | No repetition between tool and parameter descriptions |
| Abbreviated Terms | [ ] | Use "Max results" not "Maximum number of results to return" |
| Consistent Params | [ ] | Common params (limit, offset, page) use identical descriptions |

### 6. Documentation

| Requirement | Status | Description |
|-------------|--------|-------------|
| README Tool Reference | [ ] | README includes all tool descriptions |
| Workflow Examples | [ ] | Common multi-tool workflows documented |
| Error Handling Guide | [ ] | Common errors and recovery strategies |
| Parameter Patterns | [ ] | Common formats (IDs, dates, queries) documented once |

### 7. Code Quality

| Requirement | Status | Description |
|-------------|--------|-------------|
| Compiles Successfully | [ ] | `go build ./...` completes without errors |
| Tests Pass | [ ] | `go test ./...` completes without failures |
| No Unused Code | [ ] | No commented-out code or unused variables |
| Consistent Formatting | [ ] | Code follows Go standards (`go fmt`) |

## Issue Resolution Process

When a checklist item fails:

1. **Document**: Note which item failed and in which file
2. **Fix**: Make necessary code changes
3. **Verify**: Re-run relevant checks
4. **Test**: Add tests for new functionality
5. **Re-verify**: Ensure fix didn't break other items

## Tool Naming Convention

| Prefix | Use For | Example |
|--------|---------|---------|
| `get_` | Retrieve single item by ID | `jira_get_issue` |
| `list_` | Retrieve multiple items | `jira_list_projects` |
| `search_` | Query-based retrieval | `jira_search` |
| `create_` | Create new item | `jira_create_issue` |
| `update_` | Modify existing item | `jira_update_issue` |
| `delete_` | Remove item | `jira_delete_issue` |
| `add_` | Add to collection | `jira_add_comment` |

## Parameter Naming Convention

| Parameter | Type | Description Format |
|-----------|------|-------------------|
| `limit` | integer | Max results (1-100, default: 50) |
| `offset` | integer | Skip first N results (default: 0) |
| `page` | integer | Page number (1-based) |
| `issue_key` | string | Issue identifier (e.g., "PROJ-123") |
| `project_key` | string | Project identifier (e.g., "PROJ") |
| `space_key` | string | Confluence space key (e.g., "DEV") |
| `page_id` | string | Confluence page ID |

## Common Description Patterns

Use these templates for consistency:

**Read tools:**
```
Retrieve [item] by [identifier]. Returns [fields].
```

**Search tools:**
```
Search [items] using [query language]. Supports [key features].
```

**Create tools:**
```
Create new [item] with specified [required fields]. Returns [result].
```

**Update tools:**
```
Update [item] [identifier]. Modifies [changeable fields].
```

**Delete tools:**
```
Permanently delete [item] by [identifier]. This action cannot be undone.
```

## Related Documentation

- [README.md](./README.md) - Tool reference and configuration
- [INTEGRATION.md](./INTEGRATION.md) - Client configuration guide
- [ECS.md](./ECS.md) - AWS ECS deployment guide
