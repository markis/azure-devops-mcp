// Package client provides the ADOClient interface, shared types, and implementations
// for interacting with the Azure DevOps work item tracking API.
package client

import "context"

// MockADOClient implements ADOClient for use in unit tests.
// Set each function field to control the response for that method.
type MockADOClient struct {
	GetWorkItemFn     func(ctx context.Context, project string, id int) (*WorkItem, error)
	ListWorkItemsFn   func(ctx context.Context, project string, wiql string) ([]*WorkItem, error)
	ListMyWorkItemsFn func(ctx context.Context, project string) ([]*WorkItem, error)
	CreateWorkItemFn  func(ctx context.Context, project, workItemType, title string, opts CreateOptions) (*WorkItem, error)
	UpdateWorkItemFn  func(ctx context.Context, project string, id int, opts UpdateOptions) (*WorkItem, error)
	AddCommentFn      func(ctx context.Context, project string, id int, text string) error
}

// GetWorkItem delegates to GetWorkItemFn.
func (m *MockADOClient) GetWorkItem(ctx context.Context, project string, id int) (*WorkItem, error) {
	return m.GetWorkItemFn(ctx, project, id)
}

// ListWorkItems delegates to ListWorkItemsFn.
func (m *MockADOClient) ListWorkItems(ctx context.Context, project string, wiql string) ([]*WorkItem, error) {
	return m.ListWorkItemsFn(ctx, project, wiql)
}

// ListMyWorkItems delegates to ListMyWorkItemsFn.
func (m *MockADOClient) ListMyWorkItems(ctx context.Context, project string) ([]*WorkItem, error) {
	return m.ListMyWorkItemsFn(ctx, project)
}

// CreateWorkItem delegates to CreateWorkItemFn.
func (m *MockADOClient) CreateWorkItem(
	ctx context.Context, project, workItemType, title string, opts CreateOptions,
) (*WorkItem, error) {
	return m.CreateWorkItemFn(ctx, project, workItemType, title, opts)
}

// UpdateWorkItem delegates to UpdateWorkItemFn.
func (m *MockADOClient) UpdateWorkItem(
	ctx context.Context, project string, id int, opts UpdateOptions,
) (*WorkItem, error) {
	return m.UpdateWorkItemFn(ctx, project, id, opts)
}

// AddComment delegates to AddCommentFn.
func (m *MockADOClient) AddComment(ctx context.Context, project string, id int, text string) error {
	return m.AddCommentFn(ctx, project, id, text)
}
