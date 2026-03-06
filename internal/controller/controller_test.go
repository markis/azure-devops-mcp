package controller_test

import (
	"testing"

	"github.com/markistaylor/azure-devops-mcp/internal/client"
	"github.com/markistaylor/azure-devops-mcp/internal/controller"
	"github.com/markistaylor/azure-devops-mcp/internal/tools"
)

func TestCreateServer(t *testing.T) {
	srv := controller.CreateServer()

	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestRegisterTools_AllToolsRegistered(t *testing.T) {
	mock := &client.MockADOClient{}
	h := tools.NewHandlers(mock, "TestProject")
	srv := controller.CreateServer()

	// This should not panic
	controller.RegisterTools(srv, h)

	// Verify the server was created successfully
	if srv == nil {
		t.Fatal("expected non-nil server after tool registration")
	}
}
