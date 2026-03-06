package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/markistaylor/azure-devops-mcp/internal/client"
	"github.com/markistaylor/azure-devops-mcp/internal/tools"
)

func TestGetWorkItem_ReturnsWorkItem(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(ctx context.Context, project string, id int) (*client.WorkItem, error) {
			if id != 42 {
				t.Fatalf("expected id 42, got %d", id)
			}
			return &client.WorkItem{
				ID:    42,
				Title: "Fix the thing",
				State: "Active",
				Type:  "Bug",
			}, nil
		},
	}

	h := tools.NewHandlers(mock, "MyProject")
	result, err := h.GetWorkItem(context.Background(), 42, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wi client.WorkItem
	if err := json.Unmarshal([]byte(result), &wi); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if wi.ID != 42 {
		t.Errorf("expected ID 42, got %d", wi.ID)
	}
	if wi.Title != "Fix the thing" {
		t.Errorf("expected title 'Fix the thing', got %q", wi.Title)
	}
}

func TestListWorkItems_ReturnsItems(t *testing.T) {
	mock := &client.MockADOClient{
		ListWorkItemsFn: func(ctx context.Context, project, wiql string) ([]*client.WorkItem, error) {
			return []*client.WorkItem{
				{ID: 1, Title: "Item one", State: "Active"},
				{ID: 2, Title: "Item two", State: "Resolved"},
			}, nil
		},
	}

	h := tools.NewHandlers(mock, "MyProject")
	result, err := h.ListWorkItems(context.Background(), "SELECT [Id] FROM WorkItems", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []*client.WorkItem
	if err := json.Unmarshal([]byte(result), &items); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestListMyWorkItems_ReturnsAssignedItems(t *testing.T) {
	mock := &client.MockADOClient{
		ListMyWorkItemsFn: func(ctx context.Context, project string) ([]*client.WorkItem, error) {
			return []*client.WorkItem{
				{ID: 5, Title: "My task", State: "Active", AssignedTo: "me@example.com"},
			}, nil
		},
	}

	h := tools.NewHandlers(mock, "MyProject")
	result, err := h.ListMyWorkItems(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []*client.WorkItem
	if err := json.Unmarshal([]byte(result), &items); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if len(items) != 1 || items[0].ID != 5 {
		t.Errorf("unexpected items: %+v", items)
	}
}

func TestCreateWorkItem_CreatesAndReturnsItem(t *testing.T) {
	mock := &client.MockADOClient{
		CreateWorkItemFn: func(ctx context.Context, project, workItemType, title string, opts client.CreateOptions) (*client.WorkItem, error) {
			if title != "New bug" {
				t.Fatalf("expected title 'New bug', got %q", title)
			}
			if workItemType != "Bug" {
				t.Fatalf("expected type 'Bug', got %q", workItemType)
			}
			return &client.WorkItem{ID: 99, Title: title, Type: workItemType, State: "New"}, nil
		},
	}

	h := tools.NewHandlers(mock, "MyProject")
	result, err := h.CreateWorkItem(context.Background(), "Bug", "New bug", client.CreateOptions{}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wi client.WorkItem
	if err := json.Unmarshal([]byte(result), &wi); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if wi.ID != 99 {
		t.Errorf("expected ID 99, got %d", wi.ID)
	}
}

func TestUpdateWorkItem_UpdatesAndReturnsItem(t *testing.T) {
	mock := &client.MockADOClient{
		UpdateWorkItemFn: func(ctx context.Context, project string, id int, opts client.UpdateOptions) (*client.WorkItem, error) {
			if id != 42 {
				t.Fatalf("expected id 42, got %d", id)
			}
			return &client.WorkItem{ID: 42, Title: opts.Title, State: opts.State}, nil
		},
	}

	h := tools.NewHandlers(mock, "MyProject")
	opts := client.UpdateOptions{Title: "Updated title", State: "Resolved"}
	result, err := h.UpdateWorkItem(context.Background(), 42, opts, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wi client.WorkItem
	if err := json.Unmarshal([]byte(result), &wi); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if wi.Title != "Updated title" {
		t.Errorf("expected 'Updated title', got %q", wi.Title)
	}
}

func TestAddComment_PostsComment(t *testing.T) {
	called := false
	mock := &client.MockADOClient{
		AddCommentFn: func(ctx context.Context, project string, id int, text string) error {
			if id != 42 || text != "hello" {
				t.Fatalf("unexpected args: id=%d text=%q", id, text)
			}
			called = true
			return nil
		},
	}

	h := tools.NewHandlers(mock, "MyProject")
	result, err := h.AddComment(context.Background(), 42, "hello", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("AddCommentFn was not called")
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}
