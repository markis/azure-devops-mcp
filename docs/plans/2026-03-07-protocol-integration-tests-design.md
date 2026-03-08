# Protocol-Level Integration Tests Design

**Date**: 2026-03-07  
**Status**: Approved  
**Approach**: Protocol-First Testing Strategy

## Overview

Add protocol-level integration tests using MCP SDK's in-memory transports to validate the full request/response cycle. This fills the testing gap between unit tests (business logic with mocks) and end-to-end tests (stdio transport), providing fast and deterministic validation of MCP protocol integration.

## Background

The current test suite has excellent coverage of business logic (`internal/tools/workitems_test.go`) using mocks, but lacks validation of:
- MCP protocol wiring and tool registration
- JSON argument serialization/deserialization
- FlexID type conversion
- MCP response formatting
- Error response structure

Protocol-level tests address these gaps by exercising the complete MCP stack through in-memory transports.

## Design Decisions

### Test Location

**File**: `internal/controller/integration_test.go`

**Rationale**:
- Controller package owns tool registration and server wiring
- Natural boundary between business logic and MCP protocol
- Co-located with existing `controller_test.go` registration tests

### Test Architecture

#### Shared Test Helper

```go
type testServerSetup struct {
    server        *mcp.Server
    client        *mcp.Client
    serverSession *mcp.Session
    clientSession *mcp.Session
    mockADO       *client.MockADOClient
    ctx           context.Context
}

func setupTestServer(t *testing.T) *testServerSetup
```

**Responsibilities**:
1. Create mock ADO client
2. Create handlers with mock
3. Create and configure MCP server
4. Register all tools via `controller.RegisterTools()`
5. Create bidirectional in-memory transports
6. Connect both server and client sessions
7. Register cleanup handlers via `t.Cleanup()`

**Benefits**:
- Fresh isolated environment per test
- Exposes mock for test-specific configuration
- Automatic resource cleanup
- Reduces boilerplate in individual tests

#### Mock Configuration Pattern

Each test configures only the mock functions needed:

```go
setup := setupTestServer(t)
setup.mockADO.GetWorkItemFn = func(ctx context.Context, project string, id int) (*client.WorkItem, error) {
    return &client.WorkItem{
        WorkItemSummary: client.WorkItemSummary{
            ID: 42,
            Title: "Test Item",
            State: "Active",
        },
    }, nil
}

result, err := setup.clientSession.CallTool(ctx, &mcp.CallToolParams{
    Name: "get_work_item",
    Arguments: map[string]any{"id": 42},
})
```

## Test Coverage

### Happy Path Tests (6 tests)

1. **`TestIntegration_GetWorkItem`**
   - Call `get_work_item` with valid ID
   - Verify complete work item data in response
   - Validate markdown text formatting

2. **`TestIntegration_ListWorkItems`**
   - Call `list_work_items` with WIQL query
   - Verify list of items returned
   - Check markdown table formatting

3. **`TestIntegration_ListMyWorkItems`**
   - Call `list_my_work_items`
   - Verify assigned items returned

4. **`TestIntegration_CreateWorkItem`**
   - Call `create_work_item` with type/title/options
   - Verify creation success and returned work item

5. **`TestIntegration_UpdateWorkItem`**
   - Call `update_work_item` with ID and fields
   - Verify update success

6. **`TestIntegration_AddComment`**
   - Call `add_comment` with ID and text
   - Verify comment posted

### Error and Edge Case Tests (4 tests)

7. **`TestIntegration_GetWorkItem_InvalidID`**
   - Test FlexID validation: negative, zero, malformed strings
   - Verify proper MCP error responses with `IsError: true`

8. **`TestIntegration_GetWorkItem_NotFound`**
   - Mock returns error from ADO client
   - Verify MCP error response structure

9. **`TestIntegration_ToolsList`**
   - Call `tools/list` protocol method
   - Verify all 6 tools registered with correct names
   - Validate tool schemas are present

10. **`TestIntegration_ArgumentValidation`**
    - Test missing required fields
    - Test wrong argument types
    - Verify helpful error messages

### Validation Points

Each test validates:
- ✅ Arguments serialize from `map[string]any` to input structs
- ✅ FlexID conversion (float64/string → int)
- ✅ Handler execution through full MCP stack
- ✅ Response contains expected `Content` (TextContent)
- ✅ Error responses set `IsError: true` with messages
- ✅ JSON marshaling succeeds

## Assertion Strategy

### Successful Tool Calls

```go
result, err := setup.clientSession.CallTool(ctx, params)
require.NoError(t, err)
require.False(t, result.IsError)
require.Len(t, result.Content, 1)

textContent, ok := result.Content[0].(*mcp.TextContent)
require.True(t, ok, "expected TextContent")
require.NotEmpty(t, textContent.Text)
require.Contains(t, textContent.Text, "Work Item #42")
```

### Error Cases

```go
result, err := setup.clientSession.CallTool(ctx, params)
require.True(t, result.IsError)
require.Len(t, result.Content, 1)

textContent, ok := result.Content[0].(*mcp.TextContent)
require.True(t, ok)
require.Contains(t, textContent.Text, "could not retrieve")
```

**Rationale**:
- Use `require.*` (fail fast) for structural assertions
- Use `assert.*` (continue) for content validation where appropriate
- Validate MCP protocol layer, not just business logic
- Keep assertions focused and readable

## Implementation Details

### Dependencies

Already available in project:
- `github.com/stretchr/testify/require` - Assertion library
- `github.com/modelcontextprotocol/go-sdk/mcp` - Provides `NewInMemoryTransports()`

### MCP SDK Usage

```go
// Create bidirectional in-memory connection
serverTransport, clientTransport := mcp.NewInMemoryTransports()

// Server connects to its transport
serverSession, err := server.Connect(ctx, serverTransport, nil)
require.NoError(t, err)

// Client connects to its transport
mcpClient := mcp.NewClient(&mcp.Implementation{
    Name: "test-client",
    Version: "0.1.0",
}, nil)
clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
require.NoError(t, err)
```

### FlexID Testing

The existing `FlexID` type handles multiple input formats. Tests will verify:
- Integer: `"id": 42` → works
- Float: `"id": 42.0` (from JSON) → works
- String: `"id": "42"` → works
- Invalid: `"id": -1`, `"id": 0`, `"id": "abc"` → proper errors

### Context Management

- Use `context.Background()` in setup
- Individual tests can add timeouts if needed
- Sessions closed via `t.Cleanup()` for proper resource cleanup

### Mock Isolation

- Each test configures only needed mock functions
- Unconfigured functions remain nil
- Accidental calls to unconfigured mocks will panic (catches test bugs early)

## Testing Strategy Alignment

This design follows best practices from the MCP Go SDK documentation:

1. **Separation of concerns**: Business logic already has unit tests; these test the protocol layer
2. **Fast and deterministic**: In-memory transports avoid flaky stdio issues
3. **Full stack validation**: Exercises the complete MCP request/response cycle
4. **Foundation for future**: Can add stdio tests later if needed

## Estimated Scope

- **Lines of code**: ~300-400 lines
- **Test functions**: 10 tests
- **Implementation time**: 2-3 hours
- **Maintenance burden**: Low (stable API surface)

## Success Criteria

Tests pass when:
1. All 6 tools can be called successfully through the MCP protocol
2. Error cases produce proper MCP error responses
3. `tools/list` returns all registered tools
4. FlexID conversion works for all valid input types
5. Tests run in <1 second total
6. No test flakiness

## Future Enhancements

After protocol tests are stable, consider:
- Stdio end-to-end tests (2-3 smoke tests)
- Conformance testing against MCP spec
- Performance benchmarks for tool calls
- Concurrent request testing

## References

- [MCP Go SDK Examples](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/examples/http)
- [Microsoft CosmosDB MCP Testing Blog](https://devblogs.microsoft.com/cosmosdb/build-ai-tooling-in-go-with-the-mcp-sdk-connecting-ai-apps-to-databases/)
- [MCP Testing Best Practices](https://www.jlowin.dev/blog/stop-vibe-testing-mcp-servers)
