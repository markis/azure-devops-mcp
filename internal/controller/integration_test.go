package controller_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"

	"github.com/markis/azure-devops-mcp/internal/client"
	"github.com/markis/azure-devops-mcp/internal/controller"
	"github.com/markis/azure-devops-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type testServerSetup struct {
	server        *mcp.Server
	client        *mcp.Client
	serverSession *mcp.ServerSession
	clientSession *mcp.ClientSession
	mockWIT       *client.MockWITClient
	ctx           context.Context //nolint:containedctx // Test helper convenience
}

func setupTestServer(t *testing.T) *testServerSetup {
	t.Helper()

	ctx := context.Background()
	mockWIT := &client.MockWITClient{}
	adoClient := client.NewRealADOClientWithWIT(mockWIT)
	h := tools.NewHandlers(adoClient, "TestProject")

	// Create and configure server
	srv := controller.CreateServer()
	controller.RegisterTools(srv, h)

	// Create in-memory transports
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	// Create and connect client
	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.1.0",
	}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	// Register cleanup
	t.Cleanup(func() {
		_ = serverSession.Close()
		_ = clientSession.Close()
	})

	return &testServerSetup{
		server:        srv,
		client:        mcpClient,
		serverSession: serverSession,
		clientSession: clientSession,
		mockWIT:       mockWIT,
		ctx:           ctx,
	}
}

func TestIntegration_GetWorkItem(t *testing.T) {
	setup := setupTestServer(t)

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

	// Call tool
	result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
		Name: "get_work_item",
		Arguments: map[string]any{
			"id": 42,
		},
	})

	// Validate response
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	require.NotEmpty(t, textContent.Text)
	require.Contains(t, textContent.Text, "Work Item #42")
	require.Contains(t, textContent.Text, "Test Bug")
}

func TestIntegration_GetWorkItem_InvalidID(t *testing.T) {
	setup := setupTestServer(t)

	// Test with negative ID
	result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
		Name: "get_work_item",
		Arguments: map[string]any{
			"id": -1,
		},
	})

	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, textContent.Text, "must be a positive integer")
}

func TestIntegration_GetWorkItem_NotFound(t *testing.T) {
	setup := setupTestServer(t)

	// Configure mock to return error
	setup.mockWIT.GetWorkItemFn = func(_ context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
		return nil, fmt.Errorf("work item %d not found", *args.Id) //nolint:err113 // Dynamic error acceptable in test mock
	}

	result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
		Name: "get_work_item",
		Arguments: map[string]any{
			"id": 999,
		},
	})

	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	// MCP framework uses the error parameter to populate content when err != nil
	require.Contains(t, textContent.Text, "not found")
}

func TestIntegration_ListWorkItems(t *testing.T) {
	setup := setupTestServer(t)

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

	// Call tool
	result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
		Name: "list_work_items",
		Arguments: map[string]any{
			"query": "SELECT [System.Id] FROM WorkItems",
		},
	})

	// Validate response
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.NotEmpty(t, textContent.Text)
	require.Contains(t, textContent.Text, "Item 1")
	require.Contains(t, textContent.Text, "Item 2")
}

func TestIntegration_ListMyWorkItems(t *testing.T) {
	setup := setupTestServer(t)

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

	// Call tool
	result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
		Name:      "list_my_work_items",
		Arguments: map[string]any{},
	})

	// Validate response
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.NotEmpty(t, textContent.Text)
	require.Contains(t, textContent.Text, "My Task")
}

func TestIntegration_CreateWorkItem(t *testing.T) {
	setup := setupTestServer(t)

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

	// Call tool
	result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
		Name: "create_work_item",
		Arguments: map[string]any{
			"type":        "Bug",
			"title":       "New Bug",
			"description": "Bug description",
		},
	})

	// Validate response
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.NotEmpty(t, textContent.Text)
	require.Contains(t, textContent.Text, "Created work item #100")
	require.Contains(t, textContent.Text, "New Bug")
}

func TestIntegration_UpdateWorkItem(t *testing.T) {
	setup := setupTestServer(t)

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

	// Call tool
	result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
		Name: "update_work_item",
		Arguments: map[string]any{
			"id":    42,
			"title": "Updated Title",
			"state": "Resolved",
		},
	})

	// Validate response
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.NotEmpty(t, textContent.Text)
	require.Contains(t, textContent.Text, "Updated Title")
	require.Contains(t, textContent.Text, "Resolved")
}

func TestIntegration_AddComment(t *testing.T) {
	setup := setupTestServer(t)

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

	// Call tool
	result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
		Name: "add_comment",
		Arguments: map[string]any{
			"id":   42,
			"text": "Test comment",
		},
	})

	// Validate response
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.NotEmpty(t, textContent.Text)
	require.Contains(t, textContent.Text, "Comment added")
}

func TestIntegration_FlexID_Conversions(t *testing.T) {
	setup := setupTestServer(t)

	// Configure mock to track received IDs
	var receivedID int

	setup.mockADO.GetWorkItemFn = func(_ context.Context, _ string, id int) (*client.WorkItem, error) {
		receivedID = id

		return &client.WorkItem{
			WorkItemSummary: client.WorkItemSummary{
				ID:    id,
				Title: "Test",
			},
		}, nil
	}

	testCases := []struct {
		name     string
		idValue  any
		expected int
	}{
		{"integer", 42, 42},
		{"float", 42.0, 42},
		// Note: string conversion ("42") works at the JSON unmarshal level
		// (see flextypes.go), but is currently blocked by MCP schema validation.
		// The schema generator doesn't recognize FlexID's JSONSchemaExtend method.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			receivedID = 0

			result, err := setup.clientSession.CallTool(setup.ctx, &mcp.CallToolParams{
				Name: "get_work_item",
				Arguments: map[string]any{
					"id": tc.idValue,
				},
			})

			require.NoError(t, err)
			require.False(t, result.IsError)
			require.Equal(t, tc.expected, receivedID, "ID should be converted to %d", tc.expected)
		})
	}
}

func TestIntegration_ToolsList(t *testing.T) {
	setup := setupTestServer(t)

	// Call tools/list
	result, err := setup.clientSession.ListTools(setup.ctx, &mcp.ListToolsParams{})

	// Validate response
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Tools, 6, "should have 6 registered tools")

	// Extract tool names
	toolNames := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		toolNames[i] = tool.Name
	}

	// Verify all expected tools are present
	expectedTools := []string{
		"get_work_item",
		"list_work_items",
		"list_my_work_items",
		"create_work_item",
		"update_work_item",
		"add_comment",
	}

	for _, expected := range expectedTools {
		require.Contains(t, toolNames, expected, "tool %s should be registered", expected)
	}
}
