# Refactor Integration Tests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor integration tests to mock `workitemtracking.Client` instead of `ADOClient` to test real client code and increase coverage from 45% to ~70%.

**Architecture:** Create mock Azure DevOps SDK client, inject it into RealADOClient via new test constructor, update all 13 integration tests to mock SDK types instead of application types.

**Tech Stack:** Go 1.25, Azure DevOps Go API v7, MCP SDK, testify

---

## Task 1: Create Mock WIT Client Structure

**Files:**
- Create: `internal/client/mock_wit.go`

**Step 1: Create mock file with basic structure**

```go
package client

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

// MockWITClient implements workitemtracking.Client for testing.
// Only implements methods used by RealADOClient.
type MockWITClient struct {
	GetWorkItemFn    func(context.Context, workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error)
	QueryByWiqlFn    func(context.Context, workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error)
	CreateWorkItemFn func(context.Context, workitemtracking.CreateWorkItemArgs) (*workitemtracking.WorkItem, error)
	UpdateWorkItemFn func(context.Context, workitemtracking.UpdateWorkItemArgs) (*workitemtracking.WorkItem, error)
	AddCommentFn     func(context.Context, workitemtracking.AddCommentArgs) (*workitemtracking.Comment, error)
}

// GetWorkItem delegates to GetWorkItemFn.
func (m *MockWITClient) GetWorkItem(ctx context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
	if m.GetWorkItemFn != nil {
		return m.GetWorkItemFn(ctx, args)
	}
	panic("GetWorkItemFn not set")
}

// QueryByWiql delegates to QueryByWiqlFn.
func (m *MockWITClient) QueryByWiql(ctx context.Context, args workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error) {
	if m.QueryByWiqlFn != nil {
		return m.QueryByWiqlFn(ctx, args)
	}
	panic("QueryByWiqlFn not set")
}

// CreateWorkItem delegates to CreateWorkItemFn.
func (m *MockWITClient) CreateWorkItem(ctx context.Context, args workitemtracking.CreateWorkItemArgs) (*workitemtracking.WorkItem, error) {
	if m.CreateWorkItemFn != nil {
		return m.CreateWorkItemFn(ctx, args)
	}
	panic("CreateWorkItemFn not set")
}

// UpdateWorkItem delegates to UpdateWorkItemFn.
func (m *MockWITClient) UpdateWorkItem(ctx context.Context, args workitemtracking.UpdateWorkItemArgs) (*workitemtracking.WorkItem, error) {
	if m.UpdateWorkItemFn != nil {
		return m.UpdateWorkItemFn(ctx, args)
	}
	panic("UpdateWorkItemFn not set")
}

// AddComment delegates to AddCommentFn.
func (m *MockWITClient) AddComment(ctx context.Context, args workitemtracking.AddCommentArgs) (*workitemtracking.Comment, error) {
	if m.AddCommentFn != nil {
		return m.AddCommentFn(ctx, args)
	}
	panic("AddCommentFn not set")
}

// All other workitemtracking.Client methods panic with "not implemented"
// (Add stubs for all other interface methods here)
```

**Step 2: Add stub methods for all other interface methods**

Add panic stubs for all remaining `workitemtracking.Client` methods. Use this command to see all methods:

Run: `go doc github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking Client`

For each method not in the list above, add:
```go
func (m *MockWITClient) MethodName(ctx context.Context, args workitemtracking.MethodArgs) (*workitemtracking.Result, error) {
	panic("MethodName not implemented in mock")
}
```

**Step 3: Verify compilation**

Run: `go build ./internal/client/...`
Expected: Compiles successfully

**Step 4: Commit**

```bash
git add internal/client/mock_wit.go
git commit -m "test: add mock workitemtracking.Client for integration tests"
```

---

## Task 2: Add Test Constructor to RealADOClient

**Files:**
- Modify: `internal/client/ado.go` (after line 122)

**Step 1: Add test constructor**

Add after `NewRealADOClient`:

```go
// NewRealADOClientWithWIT creates a client with an injected WIT client for testing.
func NewRealADOClientWithWIT(wit workitemtracking.Client) *RealADOClient {
	return &RealADOClient{wit: wit}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/client/...`
Expected: Compiles successfully

**Step 3: Commit**

```bash
git add internal/client/ado.go
git commit -m "feat: add test constructor for RealADOClient with injected WIT client"
```

---

## Task 3: Update Integration Test Setup

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 14-64)

**Step 1: Update testServerSetup struct**

Replace `mockADO *client.MockADOClient` with:

```go
type testServerSetup struct {
	server        *mcp.Server
	client        *mcp.Client
	serverSession *mcp.ServerSession
	clientSession *mcp.ClientSession
	mockWIT       *client.MockWITClient  // Changed from mockADO
	ctx           context.Context //nolint:containedctx
}
```

**Step 2: Update setupTestServer function**

Replace lines 27-29:

```go
ctx := context.Background()
mockWIT := &client.MockWITClient{}  // Changed from MockADOClient
adoClient := client.NewRealADOClientWithWIT(mockWIT)
h := tools.NewHandlers(adoClient, "TestProject")
```

And update the return statement:

```go
return &testServerSetup{
	server:        srv,
	client:        mcpClient,
	serverSession: serverSession,
	clientSession: clientSession,
	mockWIT:       mockWIT,  // Changed from mockADO
	ctx:           ctx,
}
```

**Step 3: Verify compilation fails (expected)**

Run: `go build ./internal/controller/...`
Expected: FAIL - all tests now reference `setup.mockADO` which doesn't exist

**Step 4: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: update integration test setup to use mock WIT client"
```

---

## Task 4: Refactor TestIntegration_GetWorkItem

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 66-103)

**Step 1: Update mock configuration**

Replace the mock setup (lines 70-83) with:

```go
// Configure mock
id := 42
title := "Test Bug"
state := "Active"
wiType := "Bug"
assignedTo := "test@example.com"

setup.mockWIT.GetWorkItemFn = func(_ context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
	require.NotNil(t, args.Id)
	require.NotNil(t, args.Project)
	require.Equal(t, 42, *args.Id)
	require.Equal(t, "TestProject", *args.Project)
	
	return &workitemtracking.WorkItem{
		Id: &id,
		Fields: &map[string]interface{}{
			"System.Title":        title,
			"System.State":        state,
			"System.WorkItemType": wiType,
			"System.AssignedTo": map[string]interface{}{
				"displayName": assignedTo,
			},
			"System.Description": "Test description",
		},
	}, nil
}
```

**Step 2: Run the test**

Run: `go test ./internal/controller/... -run TestIntegration_GetWorkItem -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: TestIntegration_GetWorkItem to use mock WIT client"
```

---

## Task 5: Refactor TestIntegration_GetWorkItem Error Tests

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 105-148)

**Step 1: TestIntegration_GetWorkItem_InvalidID stays the same**

No changes needed - it doesn't configure mock, just tests invalid input

**Step 2: Update TestIntegration_GetWorkItem_NotFound**

Replace mock setup (lines 128-131) with:

```go
// Configure mock to return error
setup.mockWIT.GetWorkItemFn = func(_ context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
	return nil, fmt.Errorf("work item %d not found", *args.Id)
}
```

**Step 3: Run tests**

Run: `go test ./internal/controller/... -run TestIntegration_GetWorkItem -v`
Expected: All 3 GetWorkItem tests PASS

**Step 4: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: GetWorkItem error tests to use mock WIT client"
```

---

## Task 6: Refactor TestIntegration_ListWorkItems

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 150-183)

**Step 1: Update mock configuration**

Replace mock setup with:

```go
// Configure mock
setup.mockWIT.QueryByWiqlFn = func(_ context.Context, args workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error) {
	require.NotNil(t, args.Project)
	require.NotNil(t, args.Wiql)
	require.Equal(t, "TestProject", *args.Project)
	require.Contains(t, *args.Wiql.Query, "SELECT")
	
	id1, id2 := 1, 2
	return &workitemtracking.WorkItemQueryResult{
		WorkItems: &[]workitemtracking.WorkItemReference{
			{Id: &id1},
			{Id: &id2},
		},
	}, nil
}

// Mock GetWorkItem calls for the IDs returned by WIQL query
setup.mockWIT.GetWorkItemFn = func(_ context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
	id := *args.Id
	title := fmt.Sprintf("Item %d", id)
	state := "Active"
	wiType := "Bug"
	if id == 2 {
		state = "Resolved"
		wiType = "Task"
	}
	
	return &workitemtracking.WorkItem{
		Id: &id,
		Fields: &map[string]interface{}{
			"System.Title":        title,
			"System.State":        state,
			"System.WorkItemType": wiType,
		},
	}, nil
}
```

**Step 2: Run test**

Run: `go test ./internal/controller/... -run TestIntegration_ListWorkItems -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: TestIntegration_ListWorkItems to use mock WIT client"
```

---

## Task 7: Refactor TestIntegration_ListMyWorkItems

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 185-212)

**Step 1: Update mock configuration**

Replace mock setup with:

```go
// Configure mock
setup.mockWIT.QueryByWiqlFn = func(_ context.Context, args workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error) {
	require.NotNil(t, args.Project)
	require.Equal(t, "TestProject", *args.Project)
	
	id := 5
	return &workitemtracking.WorkItemQueryResult{
		WorkItems: &[]workitemtracking.WorkItemReference{
			{Id: &id},
		},
	}, nil
}

setup.mockWIT.GetWorkItemFn = func(_ context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
	id := *args.Id
	return &workitemtracking.WorkItem{
		Id: &id,
		Fields: &map[string]interface{}{
			"System.Title":        "My Task",
			"System.State":        "Active",
			"System.WorkItemType": "Task",
			"System.AssignedTo": map[string]interface{}{
				"displayName": "me@example.com",
			},
		},
	}, nil
}
```

**Step 2: Run test**

Run: `go test ./internal/controller/... -run TestIntegration_ListMyWorkItems -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: TestIntegration_ListMyWorkItems to use mock WIT client"
```

---

## Task 8: Refactor TestIntegration_CreateWorkItem

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 214-257)

**Step 1: Update mock configuration**

Replace mock setup with:

```go
// Configure mock
setup.mockWIT.CreateWorkItemFn = func(_ context.Context, args workitemtracking.CreateWorkItemArgs) (*workitemtracking.WorkItem, error) {
	require.NotNil(t, args.Project)
	require.NotNil(t, args.Type)
	require.Equal(t, "TestProject", *args.Project)
	require.Equal(t, "Bug", *args.Type)
	
	// Extract title from document
	var title string
	if args.Document != nil {
		for _, op := range *args.Document {
			if op.Path != nil && *op.Path == "/fields/System.Title" {
				title = op.Value.(string)
				break
			}
		}
	}
	require.Equal(t, "New Bug", title)
	
	id := 100
	return &workitemtracking.WorkItem{
		Id: &id,
		Fields: &map[string]interface{}{
			"System.Title":        title,
			"System.WorkItemType": *args.Type,
			"System.State":        "New",
			"System.Description":  "Bug description",
		},
	}, nil
}
```

**Step 2: Run test**

Run: `go test ./internal/controller/... -run TestIntegration_CreateWorkItem -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: TestIntegration_CreateWorkItem to use mock WIT client"
```

---

## Task 9: Refactor TestIntegration_UpdateWorkItem

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 259-301)

**Step 1: Update mock configuration**

Replace mock setup with:

```go
// Configure mock
setup.mockWIT.UpdateWorkItemFn = func(_ context.Context, args workitemtracking.UpdateWorkItemArgs) (*workitemtracking.WorkItem, error) {
	require.NotNil(t, args.Project)
	require.NotNil(t, args.Id)
	require.Equal(t, "TestProject", *args.Project)
	require.Equal(t, 42, *args.Id)
	
	// Extract updated fields from document
	var title, state string
	if args.Document != nil {
		for _, op := range *args.Document {
			if op.Path == nil {
				continue
			}
			switch *op.Path {
			case "/fields/System.Title":
				title = op.Value.(string)
			case "/fields/System.State":
				state = op.Value.(string)
			}
		}
	}
	
	id := *args.Id
	return &workitemtracking.WorkItem{
		Id: &id,
		Fields: &map[string]interface{}{
			"System.Title":        title,
			"System.State":        state,
			"System.WorkItemType": "Bug",
		},
	}, nil
}
```

**Step 2: Run test**

Run: `go test ./internal/controller/... -run TestIntegration_UpdateWorkItem -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: TestIntegration_UpdateWorkItem to use mock WIT client"
```

---

## Task 10: Refactor TestIntegration_AddComment

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 303-333)

**Step 1: Update mock configuration**

Replace mock setup with:

```go
// Configure mock
setup.mockWIT.AddCommentFn = func(_ context.Context, args workitemtracking.AddCommentArgs) (*workitemtracking.Comment, error) {
	require.NotNil(t, args.Project)
	require.NotNil(t, args.WorkItemId)
	require.NotNil(t, args.Request)
	require.Equal(t, "TestProject", *args.Project)
	require.Equal(t, 42, *args.WorkItemId)
	require.Equal(t, "Test comment", *args.Request.Text)
	
	id := 1
	return &workitemtracking.Comment{
		Id:   &id,
		Text: args.Request.Text,
	}, nil
}
```

**Step 2: Run test**

Run: `go test ./internal/controller/... -run TestIntegration_AddComment -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: TestIntegration_AddComment to use mock WIT client"
```

---

## Task 11: Refactor TestIntegration_FlexID_Conversions

**Files:**
- Modify: `internal/controller/integration_test.go` (lines 335-380)

**Step 1: Update mock configuration**

Replace mock setup with:

```go
// Configure mock to track received IDs
var receivedID int
setup.mockWIT.GetWorkItemFn = func(_ context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
	receivedID = *args.Id
	id := *args.Id
	return &workitemtracking.WorkItem{
		Id: &id,
		Fields: &map[string]interface{}{
			"System.Title": "Test",
		},
	}, nil
}
```

**Step 2: Run test**

Run: `go test ./internal/controller/... -run TestIntegration_FlexID -v`
Expected: PASS with all subtests

**Step 3: Commit**

```bash
git add internal/controller/integration_test.go
git commit -m "refactor: TestIntegration_FlexID_Conversions to use mock WIT client"
```

---

## Task 12: TestIntegration_ToolsList Needs No Changes

**Files:**
- None - this test doesn't use mocks

**Step 1: Verify test still passes**

Run: `go test ./internal/controller/... -run TestIntegration_ToolsList -v`
Expected: PASS

---

## Task 13: Remove Old Mock and Verify All Tests

**Files:**
- Delete: `internal/client/mock.go`

**Step 1: Delete old mock file**

Run: `rm internal/client/mock.go`

**Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All tests PASS

**Step 3: Check coverage**

Run: `go test ./... -coverprofile=coverage.out -coverpkg=./... && go tool cover -func=coverage.out | tail -1`
Expected: Total coverage ~70-75% (up from 45%)

**Step 4: Run linter**

Run: `golangci-lint run`
Expected: 0 issues

**Step 5: Commit**

```bash
git add internal/client/mock.go
git commit -m "refactor: remove old MockADOClient, now using MockWITClient

All integration tests now mock at the Azure DevOps SDK level,
testing real client code (field extraction, error handling, type conversions).

Coverage improvement:
- Total: 45.2% → ~70-75%
- Client: 43.3% → ~85-90%
- Tools: 0%* → ~75%

*Tools showed 0% because no tests in that package"
```

---

## Verification Checklist

After completing all tasks:

- ✅ All 13 integration tests pass
- ✅ MockWITClient implements workitemtracking.Client
- ✅ RealADOClient has test constructor
- ✅ All tests mock SDK types, not application types
- ✅ Total coverage increased to 70%+
- ✅ Client coverage increased to 85%+
- ✅ Tools coverage shows actual ~75%
- ✅ No linting errors
- ✅ Test execution time remains fast (<1s)
- ✅ Clean git history with descriptive commits

## Success Criteria

The refactor is complete when:
1. All tests pass using MockWITClient
2. Coverage metrics show improvement
3. Real client code is being tested (field extraction, error handling)
4. Old MockADOClient is removed
5. Code follows project guidelines (AGENTS.md)

## Notes

- Mock setup is more verbose but tests real transformation code
- Use helper variables for pointer values (e.g., `id := 42; Id: &id`)
- ADO SDK uses pointers extensively - be careful with nil checks
- WIQL queries require two-phase mocking: QueryByWiql + GetWorkItem for each result
