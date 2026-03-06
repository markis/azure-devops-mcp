package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/markis/azure-devops-mcp/internal/client"
	"github.com/markis/azure-devops-mcp/internal/tools"
)

func TestGetWorkItem_ReturnsWorkItem(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, _ string, id int) (*client.WorkItem, error) {
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
		ListWorkItemsFn: func(_ context.Context, _, _ string) ([]*client.WorkItem, error) {
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
		ListMyWorkItemsFn: func(_ context.Context, _ string) ([]*client.WorkItem, error) {
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
		CreateWorkItemFn: func(
			_ context.Context, _, workItemType, title string, _ client.CreateOptions,
		) (*client.WorkItem, error) {
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
		UpdateWorkItemFn: func(_ context.Context, _ string, id int, opts client.UpdateOptions) (*client.WorkItem, error) {
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

func TestGetWorkItem_ReturnsEstimationFields(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, _ string, _ int) (*client.WorkItem, error) {
			return &client.WorkItem{
				ID:               42,
				Title:            "My story",
				StoryPoints:      5,
				OriginalEstimate: 8.5,
				Size:             "M",
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

	if wi.StoryPoints != 5 {
		t.Errorf("expected story points 5, got %v", wi.StoryPoints)
	}

	if wi.OriginalEstimate != 8.5 {
		t.Errorf("expected original estimate 8.5, got %v", wi.OriginalEstimate)
	}

	if wi.Size != "M" {
		t.Errorf("expected size 'M', got %q", wi.Size)
	}
}

func TestUpdateWorkItem_UpdatesEstimationFields(t *testing.T) {
	mock := &client.MockADOClient{
		UpdateWorkItemFn: func(_ context.Context, _ string, id int, opts client.UpdateOptions) (*client.WorkItem, error) {
			return &client.WorkItem{
				ID:               id,
				StoryPoints:      opts.StoryPoints,
				OriginalEstimate: opts.OriginalEstimate,
				Size:             opts.Size,
			}, nil
		},
	}

	h := tools.NewHandlers(mock, "MyProject")
	opts := client.UpdateOptions{StoryPoints: 3, OriginalEstimate: 4.0, Size: "S"}

	result, err := h.UpdateWorkItem(context.Background(), 42, opts, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wi client.WorkItem
	if err := json.Unmarshal([]byte(result), &wi); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}

	if wi.StoryPoints != 3 {
		t.Errorf("expected story points 3, got %v", wi.StoryPoints)
	}

	if wi.OriginalEstimate != 4.0 {
		t.Errorf("expected original estimate 4.0, got %v", wi.OriginalEstimate)
	}

	if wi.Size != "S" {
		t.Errorf("expected size 'S', got %q", wi.Size)
	}
}

func TestGetWorkItem_ReturnsAcceptanceCriteria(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, _ string, _ int) (*client.WorkItem, error) {
			return &client.WorkItem{
				ID:                 42,
				Title:              "My story",
				AcceptanceCriteria: "## AC\n- Does the thing",
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

	if wi.AcceptanceCriteria != "## AC\n- Does the thing" {
		t.Errorf("expected acceptance criteria, got %q", wi.AcceptanceCriteria)
	}
}

func TestGetWorkItem_ReturnsExtendedFields(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, _ string, _ int) (*client.WorkItem, error) {
			return &client.WorkItem{
				ID:            42,
				Title:         "My story",
				Priority:      2,
				AreaPath:      "Access Analyzer\\Team A",
				IterationPath: "Access Analyzer\\Sprint 10",
				ReproSteps:    "1. Do thing\n2. See bug",
				ParentID:      100,
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

	if wi.Priority != 2 {
		t.Errorf("expected priority 2, got %d", wi.Priority)
	}

	if wi.AreaPath != "Access Analyzer\\Team A" {
		t.Errorf("unexpected area path: %q", wi.AreaPath)
	}

	if wi.IterationPath != "Access Analyzer\\Sprint 10" {
		t.Errorf("unexpected iteration path: %q", wi.IterationPath)
	}

	if wi.ReproSteps != "1. Do thing\n2. See bug" {
		t.Errorf("unexpected repro steps: %q", wi.ReproSteps)
	}

	if wi.ParentID != 100 {
		t.Errorf("expected parent ID 100, got %d", wi.ParentID)
	}
}

func TestAddComment_PostsComment(t *testing.T) {
	called := false
	mock := &client.MockADOClient{
		AddCommentFn: func(_ context.Context, _ string, id int, text string) error {
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

// Error path tests

func TestGetWorkItem_Error(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, _ string, _ int) (*client.WorkItem, error) {
			return nil, client.ErrNoFieldsToUpdate
		},
	}

	h := tools.NewHandlers(mock, "MyProject")

	_, err := h.GetWorkItem(context.Background(), 42, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListWorkItems_Error(t *testing.T) {
	mock := &client.MockADOClient{
		ListWorkItemsFn: func(_ context.Context, _ string, _ string) ([]*client.WorkItem, error) {
			return nil, client.ErrNoFieldsToUpdate
		},
	}

	h := tools.NewHandlers(mock, "MyProject")

	_, err := h.ListWorkItems(context.Background(), "query", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListMyWorkItems_Error(t *testing.T) {
	mock := &client.MockADOClient{
		ListMyWorkItemsFn: func(_ context.Context, _ string) ([]*client.WorkItem, error) {
			return nil, client.ErrNoFieldsToUpdate
		},
	}

	h := tools.NewHandlers(mock, "MyProject")

	_, err := h.ListMyWorkItems(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateWorkItem_Error(t *testing.T) {
	mock := &client.MockADOClient{
		CreateWorkItemFn: func(
			_ context.Context, _, _, _ string, _ client.CreateOptions,
		) (*client.WorkItem, error) {
			return nil, client.ErrNoFieldsToUpdate
		},
	}

	h := tools.NewHandlers(mock, "MyProject")

	_, err := h.CreateWorkItem(context.Background(), "Bug", "Title", client.CreateOptions{}, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateWorkItem_Error(t *testing.T) {
	mock := &client.MockADOClient{
		UpdateWorkItemFn: func(
			_ context.Context, _ string, _ int, _ client.UpdateOptions,
		) (*client.WorkItem, error) {
			return nil, client.ErrNoFieldsToUpdate
		},
	}

	h := tools.NewHandlers(mock, "MyProject")

	_, err := h.UpdateWorkItem(context.Background(), 42, client.UpdateOptions{}, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAddComment_Error(t *testing.T) {
	mock := &client.MockADOClient{
		AddCommentFn: func(_ context.Context, _ string, _ int, _ string) error {
			return client.ErrNoFieldsToUpdate
		},
	}

	h := tools.NewHandlers(mock, "MyProject")

	_, err := h.AddComment(context.Background(), 42, "comment", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProject_UsesOverride(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, project string, _ int) (*client.WorkItem, error) {
			if project != "OverrideProject" {
				t.Errorf("expected project 'OverrideProject', got %q", project)
			}

			return &client.WorkItem{ID: 1}, nil
		},
	}

	h := tools.NewHandlers(mock, "DefaultProject")

	_, err := h.GetWorkItem(context.Background(), 1, "OverrideProject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProject_UsesDefault(t *testing.T) {
	mock := &client.MockADOClient{
		GetWorkItemFn: func(_ context.Context, project string, _ int) (*client.WorkItem, error) {
			if project != "DefaultProject" {
				t.Errorf("expected project 'DefaultProject', got %q", project)
			}

			return &client.WorkItem{ID: 1}, nil
		},
	}

	h := tools.NewHandlers(mock, "DefaultProject")

	_, err := h.GetWorkItem(context.Background(), 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
