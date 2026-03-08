package controller_test

import (
	"context"
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
