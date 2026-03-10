// Package client provides the ADOClient interface, shared types, and implementations
// for interacting with the Azure DevOps work item tracking API.
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
var ErrNoFieldsToUpdate = errors.New(
	"no fields to update: provide at least one of title, state, assigned_to, " +
		"description, acceptance_criteria, story_points, original_estimate, " +
		"completed_work, remaining_work, size, severity, or reason",
)

// Field path constants for Azure DevOps work item fields.
// These are used as pointers in JsonPatchOperation.Path.
var (
	fieldPathTitle              = "/fields/System.Title"
	fieldPathState              = "/fields/System.State"
	fieldPathDescription        = "/fields/System.Description"
	fieldPathAssignedTo         = "/fields/System.AssignedTo"
	fieldPathTags               = "/fields/System.Tags"
	fieldPathAcceptanceCriteria = "/fields/Microsoft.VSTS.Common.AcceptanceCriteria"
	fieldPathStoryPoints        = "/fields/Microsoft.VSTS.Scheduling.StoryPoints"
	fieldPathOriginalEstimate   = "/fields/Microsoft.VSTS.Scheduling.OriginalEstimate"
	fieldPathSize               = "/fields/Custom.Teeshirtsizing"
	fieldPathSeverity           = "/fields/Microsoft.VSTS.Common.Severity"
	fieldPathCompletedWork      = "/fields/Microsoft.VSTS.Scheduling.CompletedWork"
	fieldPathRemainingWork      = "/fields/Microsoft.VSTS.Scheduling.RemainingWork"
	fieldPathReason             = "/fields/System.Reason"
)

// WorkItemSummary is a lightweight representation for list operations.
// Excludes large text fields (description, acceptance criteria, repro steps).
type WorkItemSummary struct {
	ID            int     `json:"id"                       jsonschema:"Unique work item ID"`
	Title         string  `json:"title"                    jsonschema:"Work item title"`
	State         string  `json:"state"                    jsonschema:"Work item state"`
	Type          string  `json:"type"                     jsonschema:"Work item type"`
	AssignedTo    string  `json:"assigned_to"              jsonschema:"Email or display name of assignee"`
	Tags          string  `json:"tags"                     jsonschema:"Semicolon-separated tags"`
	Priority      int     `json:"priority,omitempty"       jsonschema:"Priority level (1-4)"`
	StoryPoints   float64 `json:"story_points,omitempty"   jsonschema:"Story points estimate"`
	AreaPath      string  `json:"area_path,omitempty"      jsonschema:"Area path in the project"`
	IterationPath string  `json:"iteration_path,omitempty" jsonschema:"Iteration/sprint path"`
	ParentID      int     `json:"parent_id,omitempty"      jsonschema:"ID of parent work item"`
}

// WorkItem is a slim representation of an Azure DevOps work item.
// Only fields Claude needs are included — not the full API response.
// Embeds WorkItemSummary and adds large text fields.
type WorkItem struct {
	WorkItemSummary

	Description        string  `json:"description"                   jsonschema:"Work item description"`
	AcceptanceCriteria string  `json:"acceptance_criteria,omitempty" jsonschema:"Acceptance criteria"`
	ReproSteps         string  `json:"repro_steps,omitempty"         jsonschema:"Reproduction steps"`
	OriginalEstimate   float64 `json:"original_estimate,omitempty"   jsonschema:"Time estimate in hours"`
	CompletedWork      float64 `json:"completed_work,omitempty"      jsonschema:"Completed work in hours"`
	RemainingWork      float64 `json:"remaining_work,omitempty"      jsonschema:"Remaining work in hours"`
	Size               string  `json:"size,omitempty"                jsonschema:"T-shirt size estimate"`
	Severity           string  `json:"severity,omitempty"            jsonschema:"Severity (Critical/High/Medium/Low)"`
	Reason             string  `json:"reason,omitempty"              jsonschema:"Reason for current state"`
	URL                string  `json:"url"                           jsonschema:"Work item URL"`
}

// CommonFields holds fields shared between CreateOptions and UpdateOptions.
type CommonFields struct {
	AssignedTo       string
	Description      string
	StoryPoints      float64
	OriginalEstimate float64
	CompletedWork    float64
	RemainingWork    float64
	Size             string
	Severity         string
}

// CreateOptions holds optional fields for creating a work item.
type CreateOptions struct {
	CommonFields

	Tags string
}

// UpdateOptions holds fields that can be patched on a work item.
// Only non-zero/non-empty values are applied.
type UpdateOptions struct {
	CommonFields

	Title              string
	State              string
	AcceptanceCriteria string
	Reason             string
}

// ADOClient is the interface tool handlers depend on.
// The real implementation calls the Azure DevOps REST API;
// the mock implementation is used in unit tests.
type ADOClient interface {
	GetWorkItem(ctx context.Context, project string, id int) (*WorkItem, error)
	ListWorkItems(ctx context.Context, project string, wiql string) ([]*WorkItemSummary, error)
	ListMyWorkItems(ctx context.Context, project string) ([]*WorkItemSummary, error)
	CreateWorkItem(ctx context.Context, project, workItemType, title string, opts CreateOptions) (*WorkItem, error)
	UpdateWorkItem(ctx context.Context, project string, id int, opts UpdateOptions) (*WorkItem, error)
	AddComment(ctx context.Context, project string, id int, text string) error
}

// Client calls the Azure DevOps REST API using a PAT connection.
type Client struct {
	wit workitemtracking.Client
}

// NewClient creates a PAT-authenticated ADO client.
func NewClient(ctx context.Context, orgURL, pat string) (*Client, error) {
	conn := azuredevops.NewPatConnection(orgURL, pat)

	wit, err := workitemtracking.NewClient(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("creating work item tracking client: %w", err)
	}

	return &Client{wit: wit}, nil
}

// NewClientWithWIT creates a client with an injected WIT client for testing.
func NewClientWithWIT(wit workitemtracking.Client) *Client {
	return &Client{wit: wit}
}

// GetWorkItem fetches a single work item by ID.
func (c *Client) GetWorkItem(ctx context.Context, project string, id int) (*WorkItem, error) {
	fields := []string{
		"System.Id", "System.Title", "System.State",
		"System.WorkItemType", "System.AssignedTo",
		"System.Description", "System.Tags", "System.Reason",
		"System.AreaPath", "System.IterationPath", "System.Parent",
		"Microsoft.VSTS.Common.AcceptanceCriteria",
		"Microsoft.VSTS.Common.Priority",
		"Microsoft.VSTS.Common.Severity",
		"Custom.Teeshirtsizing",
		"Microsoft.VSTS.Scheduling.StoryPoints",
		"Microsoft.VSTS.Scheduling.OriginalEstimate",
		"Microsoft.VSTS.Scheduling.CompletedWork",
		"Microsoft.VSTS.Scheduling.RemainingWork",
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
func (c *Client) ListWorkItems(ctx context.Context, project, wiql string) ([]*WorkItemSummary, error) {
	result, err := c.wit.QueryByWiql(ctx, workitemtracking.QueryByWiqlArgs{
		Wiql:    &workitemtracking.Wiql{Query: &wiql},
		Project: &project,
	})
	if err != nil {
		return nil, fmt.Errorf("WIQL query: %w", err)
	}

	return c.fetchSummariesByRefs(ctx, project, result.WorkItems)
}

// ListMyWorkItems returns active work items assigned to the authenticated user.
func (c *Client) ListMyWorkItems(ctx context.Context, project string) ([]*WorkItemSummary, error) {
	wiql := "SELECT [System.Id] FROM WorkItems WHERE [System.AssignedTo] = @Me " +
		"AND [System.State] NOT IN ('Done','Closed','Resolved') " +
		"ORDER BY [System.ChangedDate] DESC"

	return c.ListWorkItems(ctx, project, wiql)
}

// addStringField appends a string field operation if the value is non-empty.
func addStringField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value string) {
	if value != "" {
		*ops = append(*ops, webapi.JsonPatchOperation{Op: op, Path: path, Value: value})
	}
}

// addFloatField appends a float field operation if the value is non-zero.
func addFloatField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value float64) {
	if value != 0 {
		*ops = append(*ops, webapi.JsonPatchOperation{Op: op, Path: path, Value: value})
	}
}

// CreateWorkItem creates a new work item of the given type.
func (c *Client) CreateWorkItem(
	ctx context.Context, project, workItemType, title string, opts CreateOptions,
) (*WorkItem, error) {
	add := webapi.OperationValues.Add

	ops := []webapi.JsonPatchOperation{
		{Op: &add, Path: &fieldPathTitle, Value: title},
	}

	addStringField(&ops, &add, &fieldPathTags, opts.Tags)
	buildCommonOps(&ops, &add, opts.CommonFields)

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
func (c *Client) UpdateWorkItem(
	ctx context.Context, project string, id int, opts UpdateOptions,
) (*WorkItem, error) {
	ops := buildUpdateOps(opts)
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
func (c *Client) AddComment(ctx context.Context, project string, id int, text string) error {
	_, err := c.wit.AddComment(ctx, workitemtracking.AddCommentArgs{
		Request:    &workitemtracking.CommentCreate{Text: &text},
		Project:    &project,
		WorkItemId: &id,
	})

	return err
}

// fetchSummariesByRefs retrieves work item summaries (without large text fields) by batch.
func (c *Client) fetchSummariesByRefs(
	ctx context.Context, project string, refs *[]workitemtracking.WorkItemReference,
) ([]*WorkItemSummary, error) {
	if refs == nil || len(*refs) == 0 {
		return nil, nil
	}

	ids := make([]int, len(*refs))
	for i, ref := range *refs {
		ids[i] = *ref.Id
	}

	// Only fetch essential fields for summaries (no description, acceptance criteria, repro steps)
	fields := []string{
		"System.Id", "System.Title", "System.State", "System.WorkItemType",
		"System.AssignedTo", "System.Tags",
		"System.AreaPath", "System.IterationPath", "System.Parent",
		"Microsoft.VSTS.Common.Priority",
		"Microsoft.VSTS.Scheduling.StoryPoints",
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

	result := make([]*WorkItemSummary, len(*items))
	for i, item := range *items {
		result[i] = toWorkItemSummary(&item)
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
		WorkItemSummary: WorkItemSummary{
			Title:         fieldString(f, "System.Title"),
			State:         fieldString(f, "System.State"),
			Type:          fieldString(f, "System.WorkItemType"),
			Tags:          fieldString(f, "System.Tags"),
			AreaPath:      fieldString(f, "System.AreaPath"),
			IterationPath: fieldString(f, "System.IterationPath"),
			Priority:      fieldInt(f, "Microsoft.VSTS.Common.Priority"),
			StoryPoints:   fieldFloat(f, "Microsoft.VSTS.Scheduling.StoryPoints"),
			ParentID:      extractParentID(f),
			AssignedTo:    extractAssignedTo(f),
		},
		Description:        convertToMarkdown(fieldString(f, "System.Description")),
		AcceptanceCriteria: convertToMarkdown(fieldString(f, "Microsoft.VSTS.Common.AcceptanceCriteria")),
		ReproSteps:         convertToMarkdown(fieldString(f, "Microsoft.VSTS.TCM.ReproSteps")),
		OriginalEstimate:   fieldFloat(f, "Microsoft.VSTS.Scheduling.OriginalEstimate"),
		CompletedWork:      fieldFloat(f, "Microsoft.VSTS.Scheduling.CompletedWork"),
		RemainingWork:      fieldFloat(f, "Microsoft.VSTS.Scheduling.RemainingWork"),
		Size:               fieldString(f, "Custom.Teeshirtsizing"),
		Severity:           fieldString(f, "Microsoft.VSTS.Common.Severity"),
		Reason:             fieldString(f, "System.Reason"),
	}
	if item.Id != nil {
		wi.ID = *item.Id
	}

	if item.Url != nil {
		wi.URL = *item.Url
	}

	return wi
}

// toWorkItemSummary maps an ADO API WorkItem to our lightweight WorkItemSummary type.
// Excludes large text fields like description, acceptance criteria, and repro steps.
func toWorkItemSummary(item *workitemtracking.WorkItem) *WorkItemSummary {
	if item == nil || item.Fields == nil {
		return &WorkItemSummary{}
	}

	f := item.Fields

	summary := &WorkItemSummary{
		Title:         fieldString(f, "System.Title"),
		State:         fieldString(f, "System.State"),
		Type:          fieldString(f, "System.WorkItemType"),
		Tags:          fieldString(f, "System.Tags"),
		AreaPath:      fieldString(f, "System.AreaPath"),
		IterationPath: fieldString(f, "System.IterationPath"),
		Priority:      fieldInt(f, "Microsoft.VSTS.Common.Priority"),
		StoryPoints:   fieldFloat(f, "Microsoft.VSTS.Scheduling.StoryPoints"),
		ParentID:      extractParentID(f),
		AssignedTo:    extractAssignedTo(f),
	}
	if item.Id != nil {
		summary.ID = *item.Id
	}

	return summary
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

// buildCommonOps adds patch operations for CommonFields.
func buildCommonOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields CommonFields) {
	addStringField(ops, operation, &fieldPathAssignedTo, fields.AssignedTo)
	addStringField(ops, operation, &fieldPathDescription, fields.Description)
	addStringField(ops, operation, &fieldPathSize, fields.Size)
	addStringField(ops, operation, &fieldPathSeverity, fields.Severity)

	addFloatField(ops, operation, &fieldPathStoryPoints, fields.StoryPoints)
	addFloatField(ops, operation, &fieldPathOriginalEstimate, fields.OriginalEstimate)
	addFloatField(ops, operation, &fieldPathCompletedWork, fields.CompletedWork)
	addFloatField(ops, operation, &fieldPathRemainingWork, fields.RemainingWork)
}

// buildUpdateOps converts UpdateOptions into a JSON patch operation slice.
func buildUpdateOps(opts UpdateOptions) []webapi.JsonPatchOperation {
	replace := webapi.OperationValues.Replace

	var ops []webapi.JsonPatchOperation

	addStringField(&ops, &replace, &fieldPathTitle, opts.Title)
	addStringField(&ops, &replace, &fieldPathState, opts.State)
	addStringField(&ops, &replace, &fieldPathAcceptanceCriteria, opts.AcceptanceCriteria)
	addStringField(&ops, &replace, &fieldPathReason, opts.Reason)

	buildCommonOps(&ops, &replace, opts.CommonFields)

	return ops
}

// fieldFloat extracts a float64 value from the ADO fields map.
func fieldFloat(f *map[string]any, key string) float64 {
	v, ok := (*f)[key]
	if !ok || v == nil {
		return 0
	}

	n, _ := v.(float64)

	return n
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
