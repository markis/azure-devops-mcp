package controller_test

import (
	"context"
	"testing"

	"github.com/markistaylor/azure-devops-mcp/internal/client"
	"github.com/markistaylor/azure-devops-mcp/internal/controller"
	"github.com/markistaylor/azure-devops-mcp/internal/tools"
)

func TestRegisterAllTools(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, _ string, id int) (*client.WorkItem, error) {
			return &client.WorkItem{ID: id, Title: "Test"}, nil
		},
		ListWorkItemsFn: func(_ context.Context, _ string, _ string) ([]*client.WorkItem, error) {
			return []*client.WorkItem{{ID: 1}}, nil
		},
		ListMyWorkItemsFn: func(_ context.Context, _ string) ([]*client.WorkItem, error) {
			return []*client.WorkItem{{ID: 1}}, nil
		},
		CreateWorkItemFn: func(
			_ context.Context, _, _, _ string, _ client.CreateOptions,
		) (*client.WorkItem, error) {
			return &client.WorkItem{ID: 1}, nil
		},
		UpdateWorkItemFn: func(
			_ context.Context, _ string, _ int, _ client.UpdateOptions,
		) (*client.WorkItem, error) {
			return &client.WorkItem{ID: 1}, nil
		},
		AddCommentFn: func(_ context.Context, _ string, _ int, _ string) error {
			return nil
		},
	}

	h := tools.NewHandlers(mock, "Project")
	srv := controller.CreateServer()

	// This should register all 6 tools without panicking
	controller.RegisterTools(srv, h)

	if srv == nil {
		t.Fatal("server should not be nil")
	}
}
