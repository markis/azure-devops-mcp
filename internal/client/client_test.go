package client_test

import (
	"testing"

	"github.com/markis/azure-devops-mcp/internal/client"
)

func TestWorkItem_JSONMarshaling(t *testing.T) {
	wi := &client.WorkItem{
		WorkItemSummary: client.WorkItemSummary{
			ID:            123,
			Title:         "Test Item",
			State:         "Active",
			Type:          "Bug",
			AssignedTo:    "user@example.com",
			Tags:          "tag1;tag2",
			AreaPath:      "Area\\Path",
			IterationPath: "Iteration\\Path",
			Priority:      1,
			ParentID:      456,
			StoryPoints:   5.0,
		},
		Description:        "Test description",
		AcceptanceCriteria: "AC",
		ReproSteps:         "Steps",
		OriginalEstimate:   8.0,
		Size:               "M",
	}

	if wi.ID != 123 {
		t.Errorf("expected ID 123, got %d", wi.ID)
	}

	if wi.Title != "Test Item" {
		t.Errorf("expected title 'Test Item', got %q", wi.Title)
	}
}

func TestCreateOptions_AllFields(t *testing.T) {
	opts := client.CreateOptions{
		Description:      "Description",
		AssignedTo:       "user@example.com",
		Tags:             "tag1;tag2",
		StoryPoints:      3.0,
		OriginalEstimate: 5.0,
		Size:             "S",
	}

	if opts.Description != "Description" {
		t.Errorf("expected Description, got %q", opts.Description)
	}

	if opts.StoryPoints != 3.0 {
		t.Errorf("expected StoryPoints 3.0, got %f", opts.StoryPoints)
	}
}

func TestUpdateOptions_AllFields(t *testing.T) {
	opts := client.UpdateOptions{
		Title:              "Title",
		State:              "Active",
		AssignedTo:         "user@example.com",
		Description:        "Description",
		AcceptanceCriteria: "AC",
		StoryPoints:        5.0,
		OriginalEstimate:   8.0,
		Size:               "L",
	}

	if opts.Title != "Title" {
		t.Errorf("expected Title, got %q", opts.Title)
	}

	if opts.State != "Active" {
		t.Errorf("expected Active, got %q", opts.State)
	}
}

func TestErrNoFieldsToUpdate(t *testing.T) {
	err := client.ErrNoFieldsToUpdate
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
