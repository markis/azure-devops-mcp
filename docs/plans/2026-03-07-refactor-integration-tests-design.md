# Refactor Integration Tests to Mock Azure DevOps SDK

**Date**: 2026-03-07  
**Status**: Approved  
**Goal**: Refactor integration tests to mock `workitemtracking.Client` instead of `ADOClient` interface

## Problem

Current integration tests mock at the wrong layer, bypassing all client code:

```
Integration Tests → MockADOClient → Tools → Controller → MCP
                         ↓
                   [RealADOClient never executed]
```

**Issues:**
- Client package shows 43.3% coverage but real client code has 0% coverage
- Tools package shows 0% coverage despite being tested via integration
- Not testing field extraction, error wrapping, or type conversions
- Mock is at application boundary, not external dependency boundary

## Solution

Mock the Azure DevOps SDK (`workitemtracking.Client`) instead:

```
Integration Tests → MockWITClient → RealADOClient → Tools → Controller → MCP
                                           ↓
                                    [All code tested]
```

**Benefits:**
- Tests real client transformation code
- Higher coverage (~70-75% total vs 45%)
- Finds real bugs in field extraction and error handling
- Follows best practice: mock external dependencies, not internal interfaces

## Architecture

### Component Changes

**1. New: `internal/client/mock_wit.go`**

Mock implementation of `workitemtracking.Client` interface using function field pattern:

```go
type MockWITClient struct {
    GetWorkItemFn   func(context.Context, workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error)
    QueryByWiqlFn   func(context.Context, workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error)
    CreateWorkItemFn func(context.Context, workitemtracking.CreateWorkItemArgs) (*workitemtracking.WorkItem, error)
    UpdateWorkItemFn func(context.Context, workitemtracking.UpdateWorkItemArgs) (*workitemtracking.WorkItem, error)
    AddCommentFn    func(context.Context, workitemtracking.AddCommentArgs) (*workitemtracking.Comment, error)
}
```

Only implements the 5 methods actually used by `RealADOClient`. All other methods panic with "not implemented".

**2. Modified: `internal/client/ado.go`**

Add test constructor that accepts `workitemtracking.Client`:

```go
// NewRealADOClientWithWIT creates a client with injected WIT client (for testing)
func NewRealADOClientWithWIT(wit workitemtracking.Client) *RealADOClient {
    return &RealADOClient{wit: wit}
}
```

Keep existing `NewRealADOClient` for production PAT authentication.

**3. Modified: `internal/controller/integration_test.go`**

Update test setup:

```go
func setupTestServer(t *testing.T) *testServerSetup {
    ctx := context.Background()
    mockWIT := &client.MockWITClient{}
    adoClient := client.NewRealADOClientWithWIT(mockWIT)
    h := tools.NewHandlers(adoClient, "TestProject")
    
    // ... rest of MCP server setup
    
    return &testServerSetup{
        mockWIT: mockWIT,
        // ...
    }
}
```

Update test mocks to return ADO SDK types:

```go
setup.mockWIT.GetWorkItemFn = func(ctx context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
    id := *args.Id
    return &workitemtracking.WorkItem{
        Id: &id,
        Fields: &map[string]interface{}{
            "System.Title": "Test Bug",
            "System.State": "Active",
            "System.WorkItemType": "Bug",
            "System.AssignedTo": map[string]interface{}{
                "displayName": "test@example.com",
            },
        },
    }, nil
}
```

## Implementation Details

### Mock Structure

The mock returns Azure DevOps SDK types with proper nested structure:

**Fields map format:**
- Simple fields: `"System.Title": "value"`
- Identity fields: `"System.AssignedTo": map[string]interface{}{"displayName": "name"}`
- Numeric fields: `"Microsoft.VSTS.Scheduling.StoryPoints": 5.0`
- Parent field: `"System.Parent": map[string]interface{}{"id": 100}`

**WIQL Query results:**
```go
&workitemtracking.WorkItemQueryResult{
    WorkItems: &[]workitemtracking.WorkItemReference{
        {Id: intPtr(1)},
        {Id: intPtr(2)},
    },
}
```

### Testing Strategy

Tests remain functionally identical but mock at lower level:

**Before:**
```go
setup.mockADO.GetWorkItemFn = func(...) (*client.WorkItem, error) {
    return &client.WorkItem{ID: 42, Title: "Test"}, nil
}
```

**After:**
```go
setup.mockWIT.GetWorkItemFn = func(ctx, args) (*workitemtracking.WorkItem, error) {
    return &workitemtracking.WorkItem{
        Id: args.Id,
        Fields: &map[string]interface{}{
            "System.Title": "Test",
        },
    }, nil
}
```

More verbose, but tests the real transformation from SDK types to our types.

## Expected Outcomes

### Coverage Improvement

| Package | Before | After | Change |
|---------|--------|-------|--------|
| Total | 45.2% | **~70-75%** | +25-30% |
| Client | 43.3% | **~85-90%** | +42-47% |
| Controller | 74.8% | 74.8% | 0% |
| Tools | 0%* | **~75%** | +75% |

*Tools showed 0% because no tests in that package, but was actually tested

### Code Quality

- ✅ Tests real field extraction logic (`fieldString`, `fieldInt`, etc.)
- ✅ Tests error wrapping and propagation
- ✅ Tests type conversions (ADO types → our types)
- ✅ Tests HTML to Markdown conversion
- ✅ Validates parent ID extraction, identity field parsing
- ✅ Catches bugs in `toWorkItem` and `buildUpdateOps`

### Test Speed

Minimal impact - mocks are still in-memory, just slightly more complex object construction.

## Migration Strategy

1. Create `mock_wit.go` with mock implementation
2. Add test constructor to `RealADOClient`
3. Update integration test setup one test at a time
4. Verify coverage improvements after each change
5. Remove old `MockADOClient` and `mock.go` when done

## Risks and Mitigations

**Risk:** Mock complexity - ADO SDK types are nested and verbose  
**Mitigation:** Create helper functions for common field structures

**Risk:** Tests become harder to read  
**Mitigation:** Add comments explaining field structure, extract setup to helper functions

**Risk:** Breaking changes if ADO SDK updates  
**Mitigation:** Mock interface, not concrete types. Tests will catch API changes.

## Success Criteria

- [ ] All 13 integration tests pass
- [ ] Total coverage increases to 70%+
- [ ] Client coverage increases to 85%+
- [ ] Tools coverage shows actual ~75%
- [ ] Test execution time remains under 1 second
- [ ] All tests validate real client code paths

## References

- MCP Go SDK testing best practices
- Azure DevOps Go API documentation
- Original protocol integration tests design (2026-03-07)
