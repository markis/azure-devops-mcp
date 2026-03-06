package client

import (
	"context"
	"fmt"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

// htmlToMD converts HTML strings returned by the ADO API to Markdown.
// Created once at package init; safe for concurrent use.
var htmlToMD = md.NewConverter("", true, nil)

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

// RealADOClient calls the Azure DevOps REST API using a PAT connection.
type RealADOClient struct {
	wit workitemtracking.Client
}

// NewRealADOClient creates a PAT-authenticated ADO client.
func NewRealADOClient(ctx context.Context, orgURL, pat string) (*RealADOClient, error) {
	conn := azuredevops.NewPatConnection(orgURL, pat)
	wit, err := workitemtracking.NewClient(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("creating work item tracking client: %w", err)
	}
	return &RealADOClient{wit: wit}, nil
}

func (c *RealADOClient) GetWorkItem(ctx context.Context, project string, id int) (*WorkItem, error) {
	fields := []string{
		"System.Id", "System.Title", "System.State",
		"System.WorkItemType", "System.AssignedTo",
		"System.Description", "System.Tags",
	}
	item, err := c.wit.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{
		Id:      &id,
		Project: &project,
		Fields:  &fields,
	})
	if err != nil {
		return nil, fmt.Errorf("get work item %d: %w", id, err)
	}
	return toWorkItem(item), nil
}

func (c *RealADOClient) ListWorkItems(ctx context.Context, project, wiql string) ([]*WorkItem, error) {
	result, err := c.wit.QueryByWiql(ctx, workitemtracking.QueryByWiqlArgs{
		Wiql:    &workitemtracking.Wiql{Query: &wiql},
		Project: &project,
	})
	if err != nil {
		return nil, fmt.Errorf("WIQL query: %w", err)
	}
	return c.fetchByRefs(ctx, project, result.WorkItems)
}

func (c *RealADOClient) ListMyWorkItems(ctx context.Context, project string) ([]*WorkItem, error) {
	wiql := "SELECT [System.Id] FROM WorkItems WHERE [System.AssignedTo] = @Me AND [System.State] NOT IN ('Done','Closed','Resolved') ORDER BY [System.ChangedDate] DESC"
	return c.ListWorkItems(ctx, project, wiql)
}

func (c *RealADOClient) CreateWorkItem(ctx context.Context, project, workItemType, title string, opts CreateOptions) (*WorkItem, error) {
	add := webapi.OperationValues.Add
	ops := []webapi.JsonPatchOperation{
		{Op: &add, Path: strPtr("/fields/System.Title"), Value: title},
	}
	if opts.Description != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &add, Path: strPtr("/fields/System.Description"), Value: opts.Description})
	}
	if opts.AssignedTo != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &add, Path: strPtr("/fields/System.AssignedTo"), Value: opts.AssignedTo})
	}
	if opts.Tags != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &add, Path: strPtr("/fields/System.Tags"), Value: opts.Tags})
	}
	item, err := c.wit.CreateWorkItem(ctx, workitemtracking.CreateWorkItemArgs{
		Document: &ops,
		Project:  &project,
		Type:     &workItemType,
	})
	if err != nil {
		return nil, fmt.Errorf("create work item: %w", err)
	}
	return toWorkItem(item), nil
}

func (c *RealADOClient) UpdateWorkItem(ctx context.Context, project string, id int, opts UpdateOptions) (*WorkItem, error) {
	replace := webapi.OperationValues.Replace
	var ops []webapi.JsonPatchOperation
	if opts.Title != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &replace, Path: strPtr("/fields/System.Title"), Value: opts.Title})
	}
	if opts.State != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &replace, Path: strPtr("/fields/System.State"), Value: opts.State})
	}
	if opts.AssignedTo != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &replace, Path: strPtr("/fields/System.AssignedTo"), Value: opts.AssignedTo})
	}
	if opts.Description != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &replace, Path: strPtr("/fields/System.Description"), Value: opts.Description})
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("no fields to update: provide at least one of title, state, assigned_to, or description")
	}
	item, err := c.wit.UpdateWorkItem(ctx, workitemtracking.UpdateWorkItemArgs{
		Document: &ops,
		Id:       &id,
		Project:  &project,
	})
	if err != nil {
		return nil, fmt.Errorf("update work item %d: %w", id, err)
	}
	return toWorkItem(item), nil
}

func (c *RealADOClient) AddComment(ctx context.Context, project string, id int, text string) error {
	_, err := c.wit.AddComment(ctx, workitemtracking.AddCommentArgs{
		Request:    &workitemtracking.CommentCreate{Text: &text},
		Project:    &project,
		WorkItemId: &id,
	})
	return err
}

// fetchByRefs fetches full work item details for a list of WorkItemReference.
func (c *RealADOClient) fetchByRefs(ctx context.Context, project string, refs *[]workitemtracking.WorkItemReference) ([]*WorkItem, error) {
	if refs == nil || len(*refs) == 0 {
		return nil, nil
	}
	ids := make([]int, len(*refs))
	for i, ref := range *refs {
		ids[i] = *ref.Id
	}
	fields := []string{"System.Id", "System.Title", "System.State", "System.WorkItemType", "System.AssignedTo", "System.Tags", "System.Description"}
	items, err := c.wit.GetWorkItemsBatch(ctx, workitemtracking.GetWorkItemsBatchArgs{
		WorkItemGetRequest: &workitemtracking.WorkItemBatchGetRequest{
			Ids:    &ids,
			Fields: &fields,
		},
		Project: &project,
	})
	if err != nil {
		return nil, fmt.Errorf("batch fetch work items: %w", err)
	}
	result := make([]*WorkItem, len(*items))
	for i, item := range *items {
		result[i] = toWorkItem(&item)
	}
	return result, nil
}

// toWorkItem maps an ADO API WorkItem to our slim WorkItem type.
func toWorkItem(item *workitemtracking.WorkItem) *WorkItem {
	if item == nil || item.Fields == nil {
		return &WorkItem{}
	}
	f := item.Fields
	get := func(key string) string {
		if v, ok := (*f)[key]; ok && v != nil {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
	wi := &WorkItem{
		Title: get("System.Title"),
		State: get("System.State"),
		Type:  get("System.WorkItemType"),
		Tags:  get("System.Tags"),
	}
	if raw := get("System.Description"); raw != "" {
		if converted, err := htmlToMD.ConvertString(raw); err == nil {
			wi.Description = converted
		} else {
			wi.Description = raw // fall back to raw HTML on conversion failure
		}
	}
	if item.Id != nil {
		wi.ID = *item.Id
	}
	// AssignedTo is an IdentityRef object; extract displayName.
	if v, ok := (*f)["System.AssignedTo"]; ok && v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			if dn, ok := m["displayName"].(string); ok {
				wi.AssignedTo = dn
			}
		}
	}
	if item.Url != nil {
		wi.URL = *item.Url
	}
	return wi
}

func strPtr(s string) *string { return &s }
