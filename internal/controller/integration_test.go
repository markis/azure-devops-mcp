package controller_test

import (
	"context"
	"fmt"
	"testing"

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
	mockADO       *client.MockADOClient
	ctx           context.Context //nolint:containedctx // Test helper convenience
}

func setupTestServer(t *testing.T) *testServerSetup {
	t.Helper()

	ctx := context.Background()
	mock := &client.MockADOClient{}
	h := tools.NewHandlers(mock, "TestProject")

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
		mockADO:       mock,
		ctx:           ctx,
	}
}

func TestIntegration_GetWorkItem(t *testing.T) {
	setup := setupTestServer(t)

	// Configure mock
	setup.mockADO.GetWorkItemFn = func(_ context.Context, project string, id int) (*client.WorkItem, error) {
		require.Equal(t, "TestProject", project)
		require.Equal(t, 42, id)

		return &client.WorkItem{
			WorkItemSummary: client.WorkItemSummary{
				ID:         42,
				Title:      "Test Bug",
				State:      "Active",
				Type:       "Bug",
				AssignedTo: "test@example.com",
			},
			Description: "Test description",
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
	setup.mockADO.GetWorkItemFn = func(_ context.Context, _ string, id int) (*client.WorkItem, error) {
		return nil, fmt.Errorf("work item %d not found", id) //nolint:err113 // Dynamic error acceptable in test mock
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
	setup.mockADO.ListWorkItemsFn = func(_ context.Context, project, query string) ([]*client.WorkItemSummary, error) {
		require.Equal(t, "TestProject", project)
		require.Contains(t, query, "SELECT")

		return []*client.WorkItemSummary{
			{ID: 1, Title: "Item 1", State: "Active", Type: "Bug"},
			{ID: 2, Title: "Item 2", State: "Resolved", Type: "Task"},
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
	setup.mockADO.ListMyWorkItemsFn = func(_ context.Context, project string) ([]*client.WorkItemSummary, error) {
		require.Equal(t, "TestProject", project)

		return []*client.WorkItemSummary{
			{ID: 5, Title: "My Task", State: "Active", Type: "Task", AssignedTo: "me@example.com"},
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
	setup.mockADO.CreateWorkItemFn = func(
		_ context.Context, project, workItemType, title string, opts client.CreateOptions,
	) (*client.WorkItem, error) {
		require.Equal(t, "TestProject", project)
		require.Equal(t, "Bug", workItemType)
		require.Equal(t, "New Bug", title)
		require.Equal(t, "Bug description", opts.Description)

		return &client.WorkItem{
			WorkItemSummary: client.WorkItemSummary{
				ID:    100,
				Title: title,
				Type:  workItemType,
				State: "New",
			},
			Description: opts.Description,
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
