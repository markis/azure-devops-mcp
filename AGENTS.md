# Agent Guidelines for Azure DevOps MCP

This document provides essential information for AI coding agents working in this repository.

## Project Overview

**Language**: Go 1.25
**Architecture**: MCP (Model Context Protocol) server for Azure DevOps work item management
**Module**: `github.com/markistaylor/azure-devops-mcp`

**Key Technologies**:
- MCP SDK: `github.com/modelcontextprotocol/go-sdk v1.4.0`
- Azure DevOps SDK: `github.com/microsoft/azure-devops-go-api/azuredevops/v7 v7.1.0`
- HTML to Markdown: `github.com/JohannesKaufmann/html-to-markdown v1.6.0`

## Build, Lint, and Test Commands

### Building
```bash
# Build to bin/ directory
go build -o ./bin/azure-devops-mcp ./cmd/azure-devops-mcp/...

# Build for current platform
go build ./cmd/azure-devops-mcp/...

# Build all packages
go build ./...
```

### Testing
```bash
# Run all tests
go test ./...

# Run all tests with verbose output
go test ./... -v

# Run a single test
go test ./internal/tools/... -run TestGetWorkItem_ReturnsWorkItem -v

# Run tests in a specific package
go test ./internal/tools/...

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Linting and Formatting
```bash
# Run all linters (uses .golangci.yml)
golangci-lint run

# Run on specific path
golangci-lint run ./internal/...

# Show all issues (no truncation)
golangci-lint run --max-issues-per-linter=0 --max-same-issues=0

# Auto-fix issues where possible
golangci-lint run --fix

# Format code with gofumpt (stricter than gofmt)
gofumpt -l -w .
```

### Dependencies
```bash
# Install/update dependencies
go mod download

# Clean up dependencies
go mod tidy
```

## Code Style Guidelines

### Import Organization
**Always organize imports in this exact order**:
1. Standard library imports (grouped)
2. Blank line
3. External dependencies (grouped)
4. Blank line (if internal imports exist)
5. Internal project imports (grouped)

**Example**:
```go
import (
    "context"
    "fmt"

    "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
    "github.com/modelcontextprotocol/go-sdk/mcp"

    "github.com/markistaylor/azure-devops-mcp/internal/client"
    "github.com/markistaylor/azure-devops-mcp/internal/tools"
)
```

**Import aliases**: Only when needed for clarity (e.g., `md "github.com/JohannesKaufmann/html-to-markdown"`)

### Formatting
- Use **gofumpt** (stricter than gofmt)
- Let golangci-lint enforce wsl_v5 whitespace rules
- No manual line wrapping needed; tools handle it

### Types and Naming

**Variables**:
- Short names in limited scope: `f`, `wi`, `h`, `m`, `v`, `id`
- Descriptive names in broader scope: `defaultProject`, `orgURL`, `htmlToMD`
- Receiver names: Single letter (`c`, `h`, `m`) or short word

**Functions/Methods**:
- Exported: `PascalCase` (`GetWorkItem`, `NewHandlers`)
- Unexported: `camelCase` (`marshal`, `toWorkItem`, `project`)
- Test functions: `Test<Function>_<Scenario>` (underscores allowed)

**Types**:
- Exported structs: `PascalCase` (`WorkItem`, `Client`)
- Options structs: `<Action>Options` (`CreateOptions`, `UpdateOptions`)
- Input structs: `<tool>Input` (`getWorkItemInput`, `listWorkItemsInput`)

**Interfaces**:
- No "Interface" suffix: `ADOClient` (not `ADOClientInterface`)

**Constants/Errors**:
- Error sentinels: `Err<Description>` (`ErrNoFieldsToUpdate`)
- Package-level vars: `camelCase` (`htmlToMD`)

### Error Handling

**Wrap errors with context at boundaries**:
```go
if err != nil {
    return nil, fmt.Errorf("creating work item tracking client: %w", err)
}

if err != nil {
    return nil, fmt.Errorf("get work item %d: %w", id, err)
}
```

**Use early returns**:
```go
if cfg.OrgURL == "" {
    log.Fatal("AZURE_DEVOPS_ORG_URL is required")
}
```

**Define sentinel errors for known conditions**:
```go
var ErrNoFieldsToUpdate = errors.New("no fields to update: provide at least one of...")

if len(ops) == 0 {
    return nil, ErrNoFieldsToUpdate
}
```

**Propagate errors without wrapping when context is clear**:
```go
if err := h.client.AddComment(ctx, h.project(project), id, text); err != nil {
    return "", err
}
```

**Graceful fallbacks where appropriate**:
```go
converted, err := htmlToMD.ConvertString(raw)
if err != nil {
    return raw  // Fall back to raw HTML
}
```

### Comments

**Package comments**: Describe purpose concisely
```go
// Package client provides the ADOClient interface, shared types, and implementations
// for interacting with the Azure DevOps work item tracking API.
package client
```

**Function/method comments**: Start with function name, complete sentences
```go
// GetWorkItem fetches a single work item by ID.
func (c *Client) GetWorkItem(...) (*WorkItem, error)
```

**Struct/type comments**: Explain purpose and key details
```go
// WorkItem is a slim representation of an Azure DevOps work item.
// Only fields Claude needs are included — not the full API response.
type WorkItem struct { ... }
```

**Inline comments**: Use sparingly for non-obvious code only
```go
// ADO returns numeric fields as float64 in the interface{} map.
func fieldInt(f *map[string]any, key string) int { ... }
```

## Architecture Patterns

### Project Structure
```
cmd/azure-devops-mcp/     # Entry point (env validation, bootstrap)
internal/
  controller/             # MCP server wiring, tool registration
  tools/                  # Business logic, tool handlers
  client/                 # Azure DevOps API abstraction
    client.go             # Real implementation
    mock.go               # Mock for testing
```

### Dependency Injection
- Use interfaces for testability (`ADOClient` interface)
- Inject dependencies through constructors
- Avoid global state (exception: package-level HTML converter is safe)

**Example**:
```go
type Handlers struct {
    client         client.ADOClient  // Interface, not concrete
    defaultProject string
}

func NewHandlers(client client.ADOClient, defaultProject string) *Handlers {
    return &Handlers{client: client, defaultProject: defaultProject}
}
```

### Testing
- Use external test packages (`package tools_test`)
- Custom mocks with function fields (no external frameworks)
- Table-driven tests where appropriate

**Mock pattern**:
```go
mock := &client.MockADOClient{
    GetWorkItemFn: func(_ context.Context, _ string, id int) (*client.WorkItem, error) {
        if id != 42 {
            t.Fatalf("expected id 42, got %d", id)
        }
        return &client.WorkItem{ID: 42, Title: "Test"}, nil
    },
}
```

### Common Patterns

**Options pattern for optional parameters**:
```go
type CreateOptions struct {
    Description string
    AssignedTo  string
    StoryPoints float64
}
```

**Constructor pattern (return pointers)**:
```go
func NewHandlers(client client.ADOClient, defaultProject string) *Handlers {
    return &Handlers{client: client, defaultProject: defaultProject}
}
```

**Helper functions for field extraction**:
```go
func fieldString(f *map[string]any, key string) string { ... }
func fieldInt(f *map[string]any, key string) int { ... }
func fieldDateTime(f *map[string]any, key string) *time.Time { ... }
func fieldBoolPtr(f *map[string]any, key string) *bool { ... }
func fieldIntPtr(f *map[string]any, key string) *int { ... }
func fieldFloatPtr(f *map[string]any, key string) *float64 { ... }
```

## Field Types and Field Groups

### Flexible Input Types

The controller layer uses flexible types to accept various input formats:

**FlexDateTime** (`internal/controller/flextypes.go`):
- Accepts: ISO 8601, RFC3339, RFC3339Nano, date-only (`2024-03-15`), empty string
- Converts to: `*time.Time` (nil for zero/empty)
- Example: `"2024-03-15T10:30:00Z"` or `"2024-03-15"`

**FlexBool** (`internal/controller/flextypes.go`):
- Accepts: `true/false`, `"yes"/"no"`, `"1"/"0"`, `1/0` (numbers), case-insensitive
- Converts to: `*bool`
- Example: `true`, `"yes"`, `1` all convert to `true`

**FlexInt** and **FlexFloat** (existing):
- Accept: numbers or string representations
- Converts to: `int` or `float64`

### Field Organization

Fields are organized into reusable field group structs (`internal/client/fields.go`):

**Date Fields** (9 fields):
```go
type DateFields struct {
    StartDate, FinishDate, TargetDate, DueDate *time.Time
    MarketDate, DevCompleteDate, QCStartDate, QCCompleteDate *time.Time
    OriginalTargetDate *time.Time
}
```

**Planning Fields** (6 fields):
```go
type PlanningFields struct {
    BusinessValue   *int     // Business value score
    StackRank       *float64 // CRITICAL for backlog ordering!
    Risk            string
    TimeCriticality *float64
    Rating, Triage  string
}
```

**Type-Specific Fields**:
- `FeatureSpecificFields` - `AtRisk` (required for Feature), `DeliveryRisk`, `RiskReason`, `MitigationPlan`
- `BugSpecificFields` - `SystemInfo`, `Blocked`, `ProposedFix`
- `UserStorySpecificFields` - `DevOwner`, `Poker`
- `TestCaseSpecificFields` - `Steps`, `AutomatedTestName`, `AutomationStatus`, etc.

**Other Field Groups**:
- `BuildFields` - `FoundIn`, `IntegrationBuild`, `ClosedInBuild`
- `SalesforceFields` - Integration with Salesforce (7 fields)
- `RequirementFields` - Requirements documentation (6 HTML fields)
- `QualityFields` - Quality and review tracking (6 fields)
- `MetricsFields` - Metrics and tracking (7 fields)
- `SecurityFields` - Security tracking (`CVENumber`, `VulnerabilitySource`)
- `StatusFields` - Status tracking (8 fields with dates)
- `CodeReviewFields` - Code review fields (8 fields)

### Field Group Pattern

**Creating work items with field groups**:
```go
opts := client.CreateOptions{
    CommonFields: client.CommonFields{
        AssignedTo: "john@example.com",
        StartDate:  convertFlexDateTime(in.StartDate),  // From CommonFields
        // ...
    },
    DateFields: &client.DateFields{
        DueDate:     convertFlexDateTime(in.DueDate),
        MarketDate:  convertFlexDateTime(in.MarketDate),
    },
    FeatureFields: &client.FeatureSpecificFields{
        AtRisk:         convertFlexBoolToPtr(in.AtRisk),
        RiskReason:     in.RiskReason,
        MitigationPlan: in.MitigationPlan,
    },
    Tags: in.Tags,
}
```

**Field group builders** (`buildDateFieldOps`, `buildFeatureFieldOps`, etc.):
```go
func buildDateFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *DateFields) {
    if fields == nil {
        return
    }
    addDateField(ops, operation, &fieldPathDueDate, fields.DueDate)
    // ... other date fields
}
```

### Conversion Helpers

**Controller → Client** (`internal/controller/helpers.go`):
```go
convertFlexDateTime(FlexDateTime) → *time.Time
convertFlexBool(FlexBool) → bool
convertFlexBoolToPtr(FlexBool) → *bool
convertFlexIntToPtr(FlexInt) → *int
convertFlexFloatToPtr(FlexFloat) → *float64
```

### Adding New Fields

To add a new field:
1. Add to appropriate `*Fields` struct in `internal/client/fields.go`
2. Add field path constant in `internal/client/client.go`
3. Add to `buildXxxFieldOps` function
4. Add to input struct in `internal/controller/controller.go`
5. Map in `registerCreateWorkItem` / `registerUpdateWorkItem`
6. Add extraction in `toWorkItem` if needed for responses
7. Add tests

## Type-Specific Schemas with oneOf

### Overview
The `create_work_item` tool uses discriminated union schemas (oneOf) to provide type-specific field visibility. When creating a Bug, only Bug-relevant fields are shown; when creating a Feature, only Feature-relevant fields are shown.

### Schema Structure
- **Bug schema** - Common fields + SystemInfo, Blocked, ProposedFix
- **Feature schema** - Common fields + AtRisk (required), Documentation (required), DeliveryRisk, RiskReason, MitigationPlan
- **User Story schema** - Common fields + DevOwner, Poker
- **Task schema** - Common fields (emphasis on OriginalEstimate, CompletedWork, RemainingWork)
- **Other schema** - All fields (fallback for Epic, Test Case, Issue, etc.)

### Implementation
Located in `internal/controller/controller.go`:
- Type-specific input structs: `createBugInput`, `createFeatureInput`, `createUserStoryInput`, `createTaskInput`, `createOtherInput`
- Schema generation: `buildCreateWorkItemSchema()` uses `jsonschema.For[T]()` to generate schemas from structs
- Tool registration: `registerCreateWorkItem()` uses `InputSchema` to override automatic schema generation

### Type Discriminator
Each schema enforces a `const` constraint on the `type` field programmatically:
- Bug: `type.Const = "Bug"`
- Feature: `type.Const = "Feature"`  
- User Story: `type.Const = "User Story"`
- Task: `type.Const = "Task"`
- Other: No const (accepts any type value)

### Testing
- Unit tests: `internal/controller/schema_test.go` verifies oneOf structure and field presence
- Integration tests: Automated tests verify schema generation without panic

## Linter Configuration

**Config**: `.golangci.yml` with `default: all` linters enabled

**Key disabled linters** (with rationale in config):
- `depguard`, `exhaustruct`, `funlen`, `gochecknoglobals`, `nlreturn`, `noinlineerr`, `paralleltest`, `tagliatelle`, `varnamelen`, `wrapcheck`, `wsl`

**Key enabled linters with settings**:
- `lll` - Line length enforced at 120 characters
- `gofumpt` - Stricter formatter than gofmt

## JSON and API Conventions

**JSON tags**: Use `snake_case` for Azure DevOps API compatibility
```go
type WorkItem struct {
    ID         int    `json:"id"`
    Title      string `json:"title"`
    AssignedTo string `json:"assigned_to"`
}
```

**Optional fields**: Use `omitempty` tag
```go
StoryPoints float64 `json:"story_points,omitempty"`
```

## When Adding New Features

1. **Define interface first** (in `internal/client/`) if adding client methods
2. **Add mock implementation** in `client/mock.go`
3. **Write tests** in `internal/tools/*_test.go` using external test package
4. **Implement business logic** in `internal/tools/`
5. **Register tool** in `internal/controller/controller.go`
6. **Run linter and tests** before committing
7. **Update this file** if introducing new patterns

## Common Commands Reference

```bash
# Full check before commit
golangci-lint run && go test ./... -v

# Build and run locally
go build -o ./bin/azure-devops-mcp ./cmd/azure-devops-mcp/...
./bin/azure-devops-mcp

# Debug single test with verbose output
go test ./internal/tools/... -run TestGetWorkItem_ReturnsWorkItem -v -count=1
```

---

**Last updated**: 2026-03-05
**Generated by**: OpenCode AI Agent
