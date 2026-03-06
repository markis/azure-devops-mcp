package client_test

import (
	"context"
	"testing"

	"github.com/markistaylor/azure-devops-mcp/internal/client"
)

func TestMockADOClient_GetWorkItem(t *testing.T) {
	called := false
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, project string, _ int) (*client.WorkItem, error) {
			called = true

			if project != "TestProject" {
				t.Errorf("expected project 'TestProject', got %q", project)
			}

			return &client.WorkItem{ID: 42, Title: "Test"}, nil
		},
	}

	wi, err := mock.GetWorkItem(context.Background(), "TestProject", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("GetWorkItemFn was not called")
	}

	if wi.ID != 42 {
		t.Errorf("expected ID 42, got %d", wi.ID)
	}
}

func TestMockADOClient_ListWorkItems(t *testing.T) {
	called := false
	mock := &client.MockADOClient{
		ListWorkItemsFn: func(_ context.Context, project, _ string) ([]*client.WorkItem, error) {
			called = true

			if project != "TestProject" {
				t.Errorf("expected project 'TestProject', got %q", project)
			}

			return []*client.WorkItem{{ID: 1}, {ID: 2}}, nil
		},
	}

	items, err := mock.ListWorkItems(context.Background(), "TestProject", "query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("ListWorkItemsFn was not called")
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestMockADOClient_ListMyWorkItems(t *testing.T) {
	called := false
	mock := &client.MockADOClient{
		ListMyWorkItemsFn: func(_ context.Context, _ string) ([]*client.WorkItem, error) {
			called = true

			return []*client.WorkItem{{ID: 1}}, nil
		},
	}

	items, err := mock.ListMyWorkItems(context.Background(), "TestProject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("ListMyWorkItemsFn was not called")
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestMockADOClient_CreateWorkItem(t *testing.T) {
	called := false
	mock := &client.MockADOClient{
		CreateWorkItemFn: func(
			_ context.Context, _, _, title string, _ client.CreateOptions,
		) (*client.WorkItem, error) {
			called = true

			if title != "New Item" {
				t.Errorf("expected title 'New Item', got %q", title)
			}

			return &client.WorkItem{ID: 100, Title: title}, nil
		},
	}

	wi, err := mock.CreateWorkItem(context.Background(), "TestProject", "Bug", "New Item", client.CreateOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("CreateWorkItemFn was not called")
	}

	if wi.ID != 100 {
		t.Errorf("expected ID 100, got %d", wi.ID)
	}
}

func TestMockADOClient_UpdateWorkItem(t *testing.T) {
	called := false
	mock := &client.MockADOClient{
		UpdateWorkItemFn: func(
			_ context.Context, _ string, id int, _ client.UpdateOptions,
		) (*client.WorkItem, error) {
			called = true

			if id != 42 {
				t.Errorf("expected id 42, got %d", id)
			}

			return &client.WorkItem{ID: 42, Title: "Updated"}, nil
		},
	}

	wi, err := mock.UpdateWorkItem(context.Background(), "TestProject", 42, client.UpdateOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("UpdateWorkItemFn was not called")
	}

	if wi.Title != "Updated" {
		t.Errorf("expected title 'Updated', got %q", wi.Title)
	}
}

func TestMockADOClient_AddComment(t *testing.T) {
	called := false
	mock := &client.MockADOClient{
		AddCommentFn: func(_ context.Context, _ string, _ int, text string) error {
			called = true

			if text != "Test comment" {
				t.Errorf("expected text 'Test comment', got %q", text)
			}

			return nil
		},
	}

	err := mock.AddComment(context.Background(), "TestProject", 42, "Test comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("AddCommentFn was not called")
	}
}
