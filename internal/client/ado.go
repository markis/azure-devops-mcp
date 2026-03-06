package client

import (
	"context"
	"errors"
	"fmt"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

var mdOpts = &md.Options{
	CodeBlockStyle: "fenced",
}

// htmlToMD converts HTML strings returned by the ADO API to Markdown.
// Created once at package init; safe for concurrent use.
var htmlToMD = md.NewConverter("", true, mdOpts)

// ErrNoFieldsToUpdate is returned when UpdateWorkItem is called with no fields set.
var ErrNoFieldsToUpdate = errors.New("no fields to update: provide at least one of title, state, assigned_to, or description")

// WorkItem is a slim representation of an Azure DevOps work item.
// Only fields Claude needs are included — not the full API response.
type WorkItem struct {
	ID                 int    `json:"id"`
	Title              string `json:"title"`
	State              string `json:"state"`
	Type               string `json:"type"`
	AssignedTo         string `json:"assigned_to"`
	Description        string `json:"description"`
	AcceptanceCriteria string `json:"acceptance_criteria,omitempty"`
	ReproSteps         string `json:"repro_steps,omitempty"`
	Tags               string `json:"tags"`
	Priority           int    `json:"priority,omitempty"`
	AreaPath           string `json:"area_path,omitempty"`
	IterationPath      string `json:"iteration_path,omitempty"`
	ParentID           int    `json:"parent_id,omitempty"`
	URL                string `json:"url"`
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

// GetWorkItem fetches a single work item by ID.
func (c *RealADOClient) GetWorkItem(ctx context.Context, project string, id int) (*WorkItem, error) {
	fields := []string{
		"System.Id", "System.Title", "System.State",
		"System.WorkItemType", "System.AssignedTo",
		"System.Description", "System.Tags",
		"System.AreaPath", "System.IterationPath", "System.Parent",
		"Microsoft.VSTS.Common.AcceptanceCriteria",
		"Microsoft.VSTS.Common.Priority",
		"Microsoft.VSTS.TCM.ReproSteps",
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

// ListWorkItems runs a WIQL query and returns matching work items.
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

// ListMyWorkItems returns active work items assigned to the authenticated user.
func (c *RealADOClient) ListMyWorkItems(ctx context.Context, project string) ([]*WorkItem, error) {
	wiql := "SELECT [System.Id] FROM WorkItems WHERE [System.AssignedTo] = @Me AND [System.State] NOT IN ('Done','Closed','Resolved') ORDER BY [System.ChangedDate] DESC"

	return c.ListWorkItems(ctx, project, wiql)
}

// CreateWorkItem creates a new work item of the given type.
func (c *RealADOClient) CreateWorkItem(ctx context.Context, project, workItemType, title string, opts CreateOptions) (*WorkItem, error) {
	add := webapi.OperationValues.Add

	ops := []webapi.JsonPatchOperation{
		{Op: &add, Path: ptr("/fields/System.Title"), Value: title},
	}
	if opts.Description != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &add, Path: ptr("/fields/System.Description"), Value: opts.Description})
	}

	if opts.AssignedTo != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &add, Path: ptr("/fields/System.AssignedTo"), Value: opts.AssignedTo})
	}

	if opts.Tags != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &add, Path: ptr("/fields/System.Tags"), Value: opts.Tags})
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

// UpdateWorkItem patches fields on an existing work item.
// Returns ErrNoFieldsToUpdate if no fields are provided.
func (c *RealADOClient) UpdateWorkItem(ctx context.Context, project string, id int, opts UpdateOptions) (*WorkItem, error) {
	replace := webapi.OperationValues.Replace

	var ops []webapi.JsonPatchOperation
	if opts.Title != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &replace, Path: ptr("/fields/System.Title"), Value: opts.Title})
	}

	if opts.State != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &replace, Path: ptr("/fields/System.State"), Value: opts.State})
	}

	if opts.AssignedTo != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &replace, Path: ptr("/fields/System.AssignedTo"), Value: opts.AssignedTo})
	}

	if opts.Description != "" {
		ops = append(ops, webapi.JsonPatchOperation{Op: &replace, Path: ptr("/fields/System.Description"), Value: opts.Description})
	}

	if len(ops) == 0 {
		return nil, ErrNoFieldsToUpdate
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

// AddComment posts a comment on a work item.
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

	fields := []string{
		"System.Id", "System.Title", "System.State", "System.WorkItemType",
		"System.AssignedTo", "System.Tags", "System.Description",
		"System.AreaPath", "System.IterationPath", "System.Parent",
		"Microsoft.VSTS.Common.AcceptanceCriteria",
		"Microsoft.VSTS.Common.Priority",
		"Microsoft.VSTS.TCM.ReproSteps",
	}

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

	wi := &WorkItem{
		Title:              fieldString(f, "System.Title"),
		State:              fieldString(f, "System.State"),
		Type:               fieldString(f, "System.WorkItemType"),
		Tags:               fieldString(f, "System.Tags"),
		Description:        convertToMarkdown(fieldString(f, "System.Description")),
		AcceptanceCriteria: convertToMarkdown(fieldString(f, "Microsoft.VSTS.Common.AcceptanceCriteria")),
		ReproSteps:         convertToMarkdown(fieldString(f, "Microsoft.VSTS.TCM.ReproSteps")),
		AreaPath:           fieldString(f, "System.AreaPath"),
		IterationPath:      fieldString(f, "System.IterationPath"),
		Priority:           fieldInt(f, "Microsoft.VSTS.Common.Priority"),
		ParentID:           extractParentID(f),
		AssignedTo:         extractAssignedTo(f),
	}
	if item.Id != nil {
		wi.ID = *item.Id
	}

	if item.Url != nil {
		wi.URL = *item.Url
	}

	return wi
}

// fieldString extracts a string value from the ADO fields map.
func fieldString(f *map[string]any, key string) string {
	if v, ok := (*f)[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}

	return ""
}

// fieldInt extracts an int value from the ADO fields map.
// ADO returns numeric fields as float64 in the interface{} map.
func fieldInt(f *map[string]any, key string) int {
	v, ok := (*f)[key]
	if !ok || v == nil {
		return 0
	}

	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

// extractParentID pulls the numeric ID from the System.Parent relation field.
// ADO returns this as a float64 (the parent work item ID).
func extractParentID(f *map[string]any) int {
	return fieldInt(f, "System.Parent")
}

// extractAssignedTo pulls the display name from the IdentityRef object
// that ADO returns for the System.AssignedTo field.
func extractAssignedTo(f *map[string]any) string {
	v, ok := (*f)["System.AssignedTo"]
	if !ok || v == nil {
		return ""
	}

	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}

	dn, _ := m["displayName"].(string)

	return dn
}

// convertToMarkdown converts an HTML description to Markdown.
// Falls back to the raw HTML string if conversion fails.
func convertToMarkdown(raw string) string {
	if raw == "" {
		return ""
	}

	converted, err := htmlToMD.ConvertString(raw)
	if err != nil {
		return raw
	}

	return converted
}

func ptr[T any](s T) *T { return &s }
