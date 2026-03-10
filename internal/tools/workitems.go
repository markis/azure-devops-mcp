// Package tools implements MCP tool handlers for Azure DevOps work items.
package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/markis/azure-devops-mcp/internal/client"
)

// Handlers holds the ADO client and default project for all tool handlers.
type Handlers struct {
	client         client.ADOClient
	defaultProject string
}

// NewHandlers creates a Handlers with the given client and default project.
func NewHandlers(client client.ADOClient, defaultProject string) *Handlers {
	return &Handlers{client: client, defaultProject: defaultProject}
}

// GetWorkItem fetches a single work item by ID.
// Returns the work item, a human-readable markdown representation, and any error.
func (h *Handlers) GetWorkItem(ctx context.Context, id int, project string) (*client.WorkItem, string, error) {
	wi, err := h.client.GetWorkItem(ctx, h.project(project), id)
	if err != nil {
		return nil, "", err
	}

	return wi, formatWorkItem(wi), nil
}

// ListWorkItems runs a WIQL query and returns matching work items.
// Returns the work items, a human-readable markdown representation, and any error.
func (h *Handlers) ListWorkItems(
	ctx context.Context, wiql, project string,
) ([]*client.WorkItemSummary, string, error) {
	items, err := h.client.ListWorkItems(ctx, h.project(project), wiql)
	if err != nil {
		return nil, "", err
	}

	return items, formatWorkItemSummaries(items), nil
}

// ListMyWorkItems returns active work items assigned to the current user.
// Returns the work items, a human-readable markdown representation, and any error.
func (h *Handlers) ListMyWorkItems(ctx context.Context, project string) ([]*client.WorkItemSummary, string, error) {
	items, err := h.client.ListMyWorkItems(ctx, h.project(project))
	if err != nil {
		return nil, "", err
	}

	return items, formatWorkItemSummaries(items), nil
}

// CreateWorkItem creates a new work item of the given type with the given title.
// Returns the created work item, a human-readable confirmation message, and any error.
func (h *Handlers) CreateWorkItem(
	ctx context.Context, workItemType, title string, opts client.CreateOptions, project string,
) (*client.WorkItem, string, error) {
	wi, err := h.client.CreateWorkItem(ctx, h.project(project), workItemType, title, opts)
	if err != nil {
		return nil, "", err
	}

	return wi, formatWorkItemCreated(wi), nil
}

// UpdateWorkItem patches fields on an existing work item.
// Returns the updated work item, a human-readable confirmation message, and any error.
func (h *Handlers) UpdateWorkItem(
	ctx context.Context, id int, opts client.UpdateOptions, project string,
) (*client.WorkItem, string, error) {
	wi, err := h.client.UpdateWorkItem(ctx, h.project(project), id, opts)
	if err != nil {
		return nil, "", err
	}

	return wi, formatWorkItemUpdated(wi), nil
}

// AddComment adds a comment to a work item and returns a confirmation message.
func (h *Handlers) AddComment(ctx context.Context, id int, text, project string) (string, error) {
	if err := h.client.AddComment(ctx, h.project(project), id, text); err != nil {
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

// formatWorkItem converts a work item to human-readable text.
// Used for detailed display (get operations).
func formatWorkItem(wi *client.WorkItem) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Work Item #%d: %s\n", wi.ID, wi.Title)
	fmt.Fprintf(&b, "Type: %s | State: %s", wi.Type, wi.State)

	if wi.AssignedTo != "" {
		fmt.Fprintf(&b, " | Assigned: %s", wi.AssignedTo)
	}

	if wi.StoryPoints > 0 {
		fmt.Fprintf(&b, " | Story Points: %.0f", wi.StoryPoints)
	}

	if wi.Severity != "" {
		fmt.Fprintf(&b, " | Severity: %s", wi.Severity)
	}

	if wi.Reason != "" {
		fmt.Fprintf(&b, " | Reason: %s", wi.Reason)
	}

	b.WriteString("\n")

	if wi.Description != "" {
		fmt.Fprintf(&b, "\nDescription: %s\n", wi.Description)
	}

	if wi.AcceptanceCriteria != "" {
		fmt.Fprintf(&b, "\nAcceptance Criteria: %s\n", wi.AcceptanceCriteria)
	}

	if wi.Tags != "" {
		fmt.Fprintf(&b, "\nTags: %s\n", wi.Tags)
	}

	if wi.CompletedWork > 0 || wi.RemainingWork > 0 {
		fmt.Fprintf(&b, "\nTime Tracking: Completed: %.1fh, Remaining: %.1fh\n",
			wi.CompletedWork, wi.RemainingWork)
	}

	return b.String()
}

// formatWorkItemCreated returns a simple confirmation message for created work items.
func formatWorkItemCreated(wi *client.WorkItem) string {
	return fmt.Sprintf("Created work item #%d: %s (Type: %s, State: %s)", wi.ID, wi.Title, wi.Type, wi.State)
}

// formatWorkItemUpdated returns a simple confirmation message for updated work items.
func formatWorkItemUpdated(wi *client.WorkItem) string {
	return fmt.Sprintf("Updated work item #%d: %s (State: %s)", wi.ID, wi.Title, wi.State)
}

// formatWorkItemSummaries converts a list of work item summaries to human-readable text.
func formatWorkItemSummaries(items []*client.WorkItemSummary) string {
	if len(items) == 0 {
		return "No work items found."
	}

	var b strings.Builder

	fmt.Fprintf(&b, "Found %d work item(s):\n\n", len(items))

	for _, wi := range items {
		fmt.Fprintf(&b, "#%d: %s (Type: %s, State: %s", wi.ID, wi.Title, wi.Type, wi.State)

		if wi.AssignedTo != "" {
			fmt.Fprintf(&b, ", Assigned: %s", wi.AssignedTo)
		}

		if wi.StoryPoints > 0 {
			fmt.Fprintf(&b, ", Points: %.0f", wi.StoryPoints)
		}

		b.WriteString(")\n")
	}

	return b.String()
}
