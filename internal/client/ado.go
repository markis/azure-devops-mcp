// internal/client/ado.go
package client

import "context"

// WorkItem is a slim representation of an Azure DevOps work item.
// Only fields Claude needs are included — not the full API response.
type WorkItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	State       string `json:"state"`
	Type        string `json:"type"`
	AssignedTo  string `json:"assigned_to"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	URL         string `json:"url"`
}

// CreateOptions holds optional fields for creating a work item.
type CreateOptions struct {
	Description string
	AssignedTo  string
	Tags        string
}

// UpdateOptions holds fields that can be patched on a work item.
// Only non-empty strings are applied.
type UpdateOptions struct {
	Title       string
	State       string
	AssignedTo  string
	Description string
}

// ADOClient is the interface tool handlers depend on.
// The real implementation calls the Azure DevOps REST API;
// the mock implementation is used in unit tests.
type ADOClient interface {
	GetWorkItem(ctx context.Context, project string, id int) (*WorkItem, error)
	ListWorkItems(ctx context.Context, project string, wiql string) ([]*WorkItem, error)
	ListMyWorkItems(ctx context.Context, project string) ([]*WorkItem, error)
	CreateWorkItem(ctx context.Context, project, workItemType, title string, opts CreateOptions) (*WorkItem, error)
	UpdateWorkItem(ctx context.Context, project string, id int, opts UpdateOptions) (*WorkItem, error)
	AddComment(ctx context.Context, project string, id int, text string) error
}
