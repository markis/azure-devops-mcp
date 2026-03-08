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
