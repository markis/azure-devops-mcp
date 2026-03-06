// Package tools implements MCP tool handlers for Azure DevOps work items.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/markistaylor/azure-devops-mcp/internal/client"
)

// Handlers holds the ADO client and default project for all tool handlers.
type Handlers struct {
	ado            client.ADOClient
	defaultProject string
}

// NewHandlers creates a Handlers with the given client and default project.
func NewHandlers(ado client.ADOClient, defaultProject string) *Handlers {
	return &Handlers{ado: ado, defaultProject: defaultProject}
}

// GetWorkItem fetches a single work item by ID.
func (h *Handlers) GetWorkItem(ctx context.Context, id int, project string) (string, error) {
	wi, err := h.ado.GetWorkItem(ctx, h.project(project), id)
	if err != nil {
		return "", err
	}

	return marshal(wi)
}

// ListWorkItems runs a WIQL query and returns matching work items.
func (h *Handlers) ListWorkItems(ctx context.Context, wiql, project string) (string, error) {
	items, err := h.ado.ListWorkItems(ctx, h.project(project), wiql)
	if err != nil {
		return "", err
	}

	return marshal(items)
}

// ListMyWorkItems returns active work items assigned to the current user.
func (h *Handlers) ListMyWorkItems(ctx context.Context, project string) (string, error) {
	items, err := h.ado.ListMyWorkItems(ctx, h.project(project))
	if err != nil {
		return "", err
	}

	return marshal(items)
}

// CreateWorkItem creates a new work item of the given type with the given title.
func (h *Handlers) CreateWorkItem(ctx context.Context, workItemType, title string, opts client.CreateOptions, project string) (string, error) {
	wi, err := h.ado.CreateWorkItem(ctx, h.project(project), workItemType, title, opts)
	if err != nil {
		return "", err
	}

	return marshal(wi)
}

// UpdateWorkItem patches fields on an existing work item.
func (h *Handlers) UpdateWorkItem(ctx context.Context, id int, opts client.UpdateOptions, project string) (string, error) {
	wi, err := h.ado.UpdateWorkItem(ctx, h.project(project), id, opts)
	if err != nil {
		return "", err
	}

	return marshal(wi)
}

// AddComment adds a comment to a work item and returns a confirmation message.
func (h *Handlers) AddComment(ctx context.Context, id int, text, project string) (string, error) {
	if err := h.ado.AddComment(ctx, h.project(project), id, text); err != nil {
		return "", err
	}

	return fmt.Sprintf(`{"message":"Comment added to work item %d"}`, id), nil
}

// project returns the override project if set, otherwise the default.
func (h *Handlers) project(override string) string {
	if override != "" {
		return override
	}

	return h.defaultProject
}

// marshal serializes v to a JSON string, returning an error on failure.
func marshal(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("serializing result: %w", err)
	}

	return string(b), nil
}
