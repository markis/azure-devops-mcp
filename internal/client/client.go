// Package client provides the ADOClient interface, shared types, and implementations
// for interacting with the Azure DevOps work item tracking API.
package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/microcosm-cc/bluemonday"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var mdOpts = &md.Options{
	CodeBlockStyle: "fenced",
}

// htmlToMD converts HTML strings returned by the ADO API to Markdown.
// Created once at package init; safe for concurrent use.
var htmlToMD = md.NewConverter("", true, mdOpts)

// mdConverter converts Markdown to HTML using GitHub Flavored Markdown.
// Created once at package init; safe for concurrent use.
var mdConverter = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
		html.WithUnsafe(), // Allow raw HTML (will be sanitized)
	),
)

// htmlSanitizer strips dangerous HTML tags and attributes.
// Created once at package init; safe for concurrent use.
var htmlSanitizer = bluemonday.UGCPolicy()

// ErrNoFieldsToUpdate is returned when UpdateWorkItem is called with no fields set.
var ErrNoFieldsToUpdate = errors.New(
	"no fields to update: provide at least one of title, state, assigned_to, " +
		"description, acceptance_criteria, story_points, original_estimate, " +
		"completed_work, remaining_work, size, severity, reason, tags, priority, " +
		"iteration_path, area_path, effort, activity, or value_area",
)

// Field path constants for Azure DevOps work item fields.
// These are used as pointers in JsonPatchOperation.Path.
var (
	// System fields.
	fieldPathTitle         = "/fields/System.Title"
	fieldPathState         = "/fields/System.State"
	fieldPathDescription   = "/fields/System.Description"
	fieldPathAssignedTo    = "/fields/System.AssignedTo"
	fieldPathTags          = "/fields/System.Tags"
	fieldPathIterationPath = "/fields/System.IterationPath"
	fieldPathAreaPath      = "/fields/System.AreaPath"
	fieldPathReason        = "/fields/System.Reason"

	// Common fields.
	fieldPathAcceptanceCriteria = "/fields/Microsoft.VSTS.Common.AcceptanceCriteria"
	fieldPathPriority           = "/fields/Microsoft.VSTS.Common.Priority"
	fieldPathSeverity           = "/fields/Microsoft.VSTS.Common.Severity"
	fieldPathActivity           = "/fields/Microsoft.VSTS.Common.Activity"
	fieldPathValueArea          = "/fields/Microsoft.VSTS.Common.ValueArea"

	// Scheduling fields.
	fieldPathStoryPoints      = "/fields/Microsoft.VSTS.Scheduling.StoryPoints"
	fieldPathOriginalEstimate = "/fields/Microsoft.VSTS.Scheduling.OriginalEstimate"
	fieldPathCompletedWork    = "/fields/Microsoft.VSTS.Scheduling.CompletedWork"
	fieldPathRemainingWork    = "/fields/Microsoft.VSTS.Scheduling.RemainingWork"
	fieldPathEffort           = "/fields/Microsoft.VSTS.Scheduling.Effort"
	fieldPathStartDate        = "/fields/Microsoft.VSTS.Scheduling.StartDate"
	fieldPathFinishDate       = "/fields/Microsoft.VSTS.Scheduling.FinishDate"
	fieldPathTargetDate       = "/fields/Microsoft.VSTS.Scheduling.TargetDate"
	fieldPathDueDate          = "/fields/Microsoft.VSTS.Scheduling.DueDate"

	// Status tracking fields.
	fieldPathActivatedBy     = "/fields/Microsoft.VSTS.Common.ActivatedBy"
	fieldPathActivatedDate   = "/fields/Microsoft.VSTS.Common.ActivatedDate"
	fieldPathResolvedBy      = "/fields/Microsoft.VSTS.Common.ResolvedBy"
	fieldPathResolvedDate    = "/fields/Microsoft.VSTS.Common.ResolvedDate"
	fieldPathResolvedReason  = "/fields/Microsoft.VSTS.Common.ResolvedReason"
	fieldPathClosedBy        = "/fields/Microsoft.VSTS.Common.ClosedBy"
	fieldPathClosedDate      = "/fields/Microsoft.VSTS.Common.ClosedDate"
	fieldPathStateChangeDate = "/fields/Microsoft.VSTS.Common.StateChangeDate"

	// Planning fields.
	fieldPathBusinessValue   = "/fields/Microsoft.VSTS.Common.BusinessValue"
	fieldPathStackRank       = "/fields/Microsoft.VSTS.Common.StackRank"
	fieldPathRisk            = "/fields/Microsoft.VSTS.Common.Risk"
	fieldPathTimeCriticality = "/fields/Microsoft.VSTS.Common.TimeCriticality"
	fieldPathRating          = "/fields/Microsoft.VSTS.Common.Rating"
	fieldPathTriage          = "/fields/Microsoft.VSTS.Common.Triage"
	fieldPathReviewedBy      = "/fields/Microsoft.VSTS.Common.ReviewedBy"

	// Build integration fields.
	fieldPathFoundIn          = "/fields/Microsoft.VSTS.Build.FoundIn"
	fieldPathIntegrationBuild = "/fields/Microsoft.VSTS.Build.IntegrationBuild"

	// Fields for Bug work items.
	fieldPathReproSteps  = "/fields/Microsoft.VSTS.TCM.ReproSteps"
	fieldPathSystemInfo  = "/fields/Microsoft.VSTS.TCM.SystemInfo"
	fieldPathBlocked     = "/fields/Microsoft.VSTS.CMMI.Blocked"
	fieldPathProposedFix = "/fields/Microsoft.VSTS.CMMI.ProposedFix"

	// Feature-specific fields.
	fieldPathMitigationPlan = "/fields/Microsoft.VSTS.CMMI.MitigationPlan"

	// Test Case-specific fields.
	fieldPathSteps                = "/fields/Microsoft.VSTS.TCM.Steps"
	fieldPathAutomatedTestName    = "/fields/Microsoft.VSTS.TCM.AutomatedTestName"
	fieldPathAutomatedTestStorage = "/fields/Microsoft.VSTS.TCM.AutomatedTestStorage"
	fieldPathAutomatedTestType    = "/fields/Microsoft.VSTS.TCM.AutomatedTestType"
	fieldPathAutomatedTestID      = "/fields/Microsoft.VSTS.TCM.AutomatedTestId"
	fieldPathAutomationStatus     = "/fields/Microsoft.VSTS.TCM.AutomationStatus"
	fieldPathParameters           = "/fields/Microsoft.VSTS.TCM.Parameters"
	fieldPathLocalDataSource      = "/fields/Microsoft.VSTS.TCM.LocalDataSource"

	// Code Review fields.
	fieldPathContext        = "/fields/Microsoft.VSTS.CodeReview.Context"
	fieldPathContextCode    = "/fields/Microsoft.VSTS.CodeReview.ContextCode"
	fieldPathContextOwner   = "/fields/Microsoft.VSTS.CodeReview.ContextOwner"
	fieldPathContextType    = "/fields/Microsoft.VSTS.CodeReview.ContextType"
	fieldPathAcceptedBy     = "/fields/Microsoft.VSTS.CodeReview.AcceptedBy"
	fieldPathAcceptedDate   = "/fields/Microsoft.VSTS.CodeReview.AcceptedDate"
	fieldPathClosedStatus   = "/fields/Microsoft.VSTS.CodeReview.ClosedStatus"
	fieldPathClosingComment = "/fields/Microsoft.VSTS.CodeReview.ClosingComment"

	// Custom fields - Dates.
	fieldPathMarketDate         = "/fields/Custom.MarketDate"
	fieldPathDevCompleteDate    = "/fields/Custom.DevCompleteDate"
	fieldPathQCStartDate        = "/fields/Custom.QCStartDate"
	fieldPathQCCompleteDate     = "/fields/Custom.QCCompleteDate"
	fieldPathOriginalTargetDate = "/fields/Custom.OriginalTargetDate"

	// Custom fields - Salesforce integration.
	fieldPathSalesforceCaseNumber         = "/fields/Custom.SalesforceCaseNumber"
	fieldPathSalesforceCaseStatus         = "/fields/Custom.SalesforceCaseStatus"
	fieldPathSalesforceCaseClosed         = "/fields/Custom.SalesforceCaseClosed"
	fieldPathSalesforceURL                = "/fields/Custom.SalesforceURL"
	fieldPathSalesforceEscalatingEngineer = "/fields/Custom.SalesforceEscalatingEngineer"
	fieldPathSalesforceEscalationReason   = "/fields/Custom.SalesforceEscalationReason"
	fieldPathEscalationAttachmentsFolder  = "/fields/Custom.EscalationAttachmentsFolder"

	// Custom fields - Requirements.
	fieldPathFunctionalRequirements    = "/fields/Custom.FunctionalRequirements"
	fieldPathNonfunctionalRequirements = "/fields/Custom.NonfunctionalRequirements"
	fieldPathBusinessCase              = "/fields/Custom.BusinessCase"
	fieldPathSuggestedTests            = "/fields/Custom.SuggestedTests"
	fieldPathRejectedIdeas             = "/fields/Custom.RejectedIdeas"
	fieldPathResources                 = "/fields/Custom.Resources"

	// Custom fields - Quality/Review.
	fieldPathApprovedBy                  = "/fields/Custom.ApprovedBy"
	fieldPathInitialDetailQuality        = "/fields/Custom.InitialDetailQuality"
	fieldPathInitialDetailQualityComment = "/fields/Custom.InitialDetailQualityComment"
	fieldPathDocumentation               = "/fields/Custom.Documentation"
	fieldPathRCAReason                   = "/fields/Custom.RCAReason"

	// Custom fields - Metrics.
	fieldPathDaysAwaitingDev          = "/fields/Custom.DaysAwaitingDev"
	fieldPathDaysAwaitingSupport      = "/fields/Custom.DaysAwaitingSupport"
	fieldPathDaysSinceLastDevUpdate   = "/fields/Custom.DaysSinceLastDevUpdate"
	fieldPathTimeSpent                = "/fields/Custom.TimeSpent"
	fieldPathPrioritizationScore      = "/fields/Custom.PrioritizationScore"
	fieldPathConfidence               = "/fields/Custom.Confidence"
	fieldPathRemainingWorkChangedDate = "/fields/Custom.RemainingWorkChangedDate"

	// Custom fields - Security.
	fieldPathCVENumber           = "/fields/Custom.CVENumber"
	fieldPathVulnerabilitySource = "/fields/Custom.VulnerabilitySource"

	// Custom fields - Feature-specific.
	fieldPathAtRisk       = "/fields/Custom.AtRisk"
	fieldPathDeliveryRisk = "/fields/Custom.DeliveryRisk"
	fieldPathRiskReason   = "/fields/Custom.RiskReason"

	// Custom fields - User Story-specific.
	fieldPathDevOwner = "/fields/Custom.DevOwner"
	fieldPathPoker    = "/fields/Custom.Poker"

	// Custom fields - Other.
	fieldPathSize          = "/fields/Custom.Teeshirtsizing"
	fieldPathClosedInBuild = "/fields/Custom.ClosedinBuild"
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
	Effort             float64 `json:"effort,omitempty"              jsonschema:"Effort in hours"`
	Size               string  `json:"size,omitempty"                jsonschema:"T-shirt size estimate"`
	Severity           string  `json:"severity,omitempty"            jsonschema:"Severity (Critical/High/Medium/Low)"`
	Activity           string  `json:"activity,omitempty"            jsonschema:"Activity type (Development/Testing/etc)"`
	ValueArea          string  `json:"value_area,omitempty"          jsonschema:"Value area (Business/Architectural)"`
	Reason             string  `json:"reason,omitempty"              jsonschema:"Reason for current state"`
	URL                string  `json:"url"                           jsonschema:"Work item URL"`

	// Date fields
	StartDate       *time.Time `json:"start_date,omitempty"        jsonschema:"Planned start date"`
	FinishDate      *time.Time `json:"finish_date,omitempty"       jsonschema:"Planned finish date"`
	TargetDate      *time.Time `json:"target_date,omitempty"       jsonschema:"Target completion date"`
	DueDate         *time.Time `json:"due_date,omitempty"          jsonschema:"Hard deadline"`
	ActivatedDate   *time.Time `json:"activated_date,omitempty"    jsonschema:"When work item was activated"`
	ResolvedDate    *time.Time `json:"resolved_date,omitempty"     jsonschema:"When work item was resolved"`
	ClosedDate      *time.Time `json:"closed_date,omitempty"       jsonschema:"When work item was closed"`
	StateChangeDate *time.Time `json:"state_change_date,omitempty" jsonschema:"When state last changed"`

	// Planning fields
	BusinessValue   *int     `json:"business_value,omitempty"   jsonschema:"Business value score"`
	StackRank       *float64 `json:"stack_rank,omitempty"       jsonschema:"Backlog ordering rank"`
	Risk            string   `json:"risk,omitempty"             jsonschema:"Risk level"`
	TimeCriticality *float64 `json:"time_criticality,omitempty" jsonschema:"Time sensitivity"`

	// Status tracking
	ActivatedBy string `json:"activated_by,omitempty" jsonschema:"Who activated"`
	ResolvedBy  string `json:"resolved_by,omitempty"  jsonschema:"Who resolved"`
	ClosedBy    string `json:"closed_by,omitempty"    jsonschema:"Who closed"`

	// Build integration
	FoundIn          string `json:"found_in,omitempty"          jsonschema:"Build where bug was found"`
	IntegrationBuild string `json:"integration_build,omitempty" jsonschema:"Build with fix"`

	// Custom fields (high priority)
	AtRisk     *bool      `json:"at_risk,omitempty"     jsonschema:"Is work item at risk"`
	CVENumber  string     `json:"cve_number,omitempty"  jsonschema:"CVE identifier"`
	DevOwner   string     `json:"dev_owner,omitempty"   jsonschema:"Development owner"`
	MarketDate *time.Time `json:"market_date,omitempty" jsonschema:"Market release date"`
}

// CommonFields holds fields shared between CreateOptions and UpdateOptions.
type CommonFields struct {
	// Existing fields
	AssignedTo       string
	Description      string
	IterationPath    string
	AreaPath         string
	Priority         int
	StoryPoints      float64
	OriginalEstimate float64
	CompletedWork    float64
	RemainingWork    float64
	Effort           float64
	Size             string
	Severity         string
	Activity         string
	ValueArea        string

	// High-priority shared fields
	StartDate        *time.Time
	FinishDate       *time.Time
	TargetDate       *time.Time
	BusinessValue    *int
	StackRank        *float64
	Risk             string
	FoundIn          string
	IntegrationBuild string
}

// CreateOptions holds optional fields for creating a work item.
type CreateOptions struct {
	// Embedded common fields (includes high-priority shared fields)
	CommonFields

	// Optional field groups (use pointers for true optionality)
	DateFields        *DateFields
	StatusFields      *StatusFields
	PlanningFields    *PlanningFields
	BuildFields       *BuildFields
	RequirementFields *RequirementFields
	QualityFields     *QualityFields
	MetricsFields     *MetricsFields
	SecurityFields    *SecurityFields
	SalesforceFields  *SalesforceFields

	// Type-specific fields (will be ignored if not applicable to work item type)
	FeatureFields    *FeatureSpecificFields
	BugFields        *BugSpecificFields
	UserStoryFields  *UserStorySpecificFields
	TestCaseFields   *TestCaseSpecificFields
	CodeReviewFields *CodeReviewFields

	// Existing
	Tags string
}

// UpdateOptions holds fields that can be patched on a work item.
// Only non-zero/non-empty values are applied.
type UpdateOptions struct {
	CommonFields

	// Optional field groups (same as CreateOptions)
	DateFields        *DateFields
	StatusFields      *StatusFields
	PlanningFields    *PlanningFields
	BuildFields       *BuildFields
	RequirementFields *RequirementFields
	QualityFields     *QualityFields
	MetricsFields     *MetricsFields
	SecurityFields    *SecurityFields
	SalesforceFields  *SalesforceFields

	// Type-specific fields
	FeatureFields    *FeatureSpecificFields
	BugFields        *BugSpecificFields
	UserStoryFields  *UserStorySpecificFields
	TestCaseFields   *TestCaseSpecificFields
	CodeReviewFields *CodeReviewFields

	// Update-specific fields
	Title              string
	State              string
	AcceptanceCriteria string
	Reason             string
	Tags               string
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
		"Microsoft.VSTS.Common.Activity",
		"Microsoft.VSTS.Common.ValueArea",
		"Custom.Teeshirtsizing",
		"Microsoft.VSTS.Scheduling.StoryPoints",
		"Microsoft.VSTS.Scheduling.OriginalEstimate",
		"Microsoft.VSTS.Scheduling.CompletedWork",
		"Microsoft.VSTS.Scheduling.RemainingWork",
		"Microsoft.VSTS.Scheduling.Effort",
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

// addHTMLField appends an HTML field operation after converting Markdown to HTML and sanitizing.
// If the value is empty, no operation is added.
func addHTMLField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value string) {
	prepared := prepareHTMLField(value)
	addStringField(ops, op, path, prepared)
}

// addFloatField appends a float field operation if the value is non-zero.
func addFloatField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value float64) {
	if value != 0 {
		*ops = append(*ops, webapi.JsonPatchOperation{Op: op, Path: path, Value: value})
	}
}

// addIntField appends an int field operation if the value is non-zero.
func addIntField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value int) {
	if value != 0 {
		*ops = append(*ops, webapi.JsonPatchOperation{Op: op, Path: path, Value: value})
	}
}

// addDateField appends a date field operation if the value is non-nil.
func addDateField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value *time.Time) {
	if value != nil && !value.IsZero() {
		// Format as ISO8601 for Azure DevOps
		*ops = append(*ops, webapi.JsonPatchOperation{
			Op:    op,
			Path:  path,
			Value: value.Format(time.RFC3339),
		})
	}
}

// addBoolField appends a bool field operation if the value is non-nil.
func addBoolField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value *bool) {
	if value != nil {
		*ops = append(*ops, webapi.JsonPatchOperation{Op: op, Path: path, Value: *value})
	}
}

// addOptionalIntField appends an int field operation if the value is non-nil.
func addOptionalIntField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value *int) {
	if value != nil {
		*ops = append(*ops, webapi.JsonPatchOperation{Op: op, Path: path, Value: *value})
	}
}

// addOptionalFloatField appends a float field operation if the value is non-nil.
func addOptionalFloatField(ops *[]webapi.JsonPatchOperation, op *webapi.Operation, path *string, value *float64) {
	if value != nil {
		*ops = append(*ops, webapi.JsonPatchOperation{Op: op, Path: path, Value: *value})
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

	// Common fields
	buildCommonOps(&ops, &add, opts.CommonFields)

	// Optional field groups
	buildDateFieldOps(&ops, &add, opts.DateFields)
	buildStatusFieldOps(&ops, &add, opts.StatusFields)
	buildPlanningFieldOps(&ops, &add, opts.PlanningFields)
	buildBuildFieldOps(&ops, &add, opts.BuildFields)
	buildSalesforceFieldOps(&ops, &add, opts.SalesforceFields)
	buildRequirementFieldOps(&ops, &add, opts.RequirementFields)
	buildQualityFieldOps(&ops, &add, opts.QualityFields)
	buildMetricsFieldOps(&ops, &add, opts.MetricsFields)
	buildSecurityFieldOps(&ops, &add, opts.SecurityFields)

	// Type-specific fields
	buildFeatureFieldOps(&ops, &add, opts.FeatureFields)
	buildBugFieldOps(&ops, &add, opts.BugFields)
	buildUserStoryFieldOps(&ops, &add, opts.UserStoryFields)
	buildTestCaseFieldOps(&ops, &add, opts.TestCaseFields)
	buildCodeReviewFieldOps(&ops, &add, opts.CodeReviewFields)

	// Tags
	addStringField(&ops, &add, &fieldPathTags, opts.Tags)

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
	prepared := prepareHTMLField(text)
	_, err := c.wit.AddComment(ctx, workitemtracking.AddCommentArgs{
		Request:    &workitemtracking.CommentCreate{Text: &prepared},
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
		Effort:             fieldFloat(f, "Microsoft.VSTS.Scheduling.Effort"),
		Size:               fieldString(f, "Custom.Teeshirtsizing"),
		Severity:           fieldString(f, "Microsoft.VSTS.Common.Severity"),
		Activity:           fieldString(f, "Microsoft.VSTS.Common.Activity"),
		ValueArea:          fieldString(f, "Microsoft.VSTS.Common.ValueArea"),
		Reason:             fieldString(f, "System.Reason"),

		// Date fields
		StartDate:       fieldDateTime(f, "Microsoft.VSTS.Scheduling.StartDate"),
		FinishDate:      fieldDateTime(f, "Microsoft.VSTS.Scheduling.FinishDate"),
		TargetDate:      fieldDateTime(f, "Microsoft.VSTS.Scheduling.TargetDate"),
		DueDate:         fieldDateTime(f, "Microsoft.VSTS.Scheduling.DueDate"),
		ActivatedDate:   fieldDateTime(f, "Microsoft.VSTS.Common.ActivatedDate"),
		ResolvedDate:    fieldDateTime(f, "Microsoft.VSTS.Common.ResolvedDate"),
		ClosedDate:      fieldDateTime(f, "Microsoft.VSTS.Common.ClosedDate"),
		StateChangeDate: fieldDateTime(f, "Microsoft.VSTS.Common.StateChangeDate"),

		// Planning fields
		BusinessValue:   fieldIntPtr(f, "Microsoft.VSTS.Common.BusinessValue"),
		StackRank:       fieldFloatPtr(f, "Microsoft.VSTS.Common.StackRank"),
		Risk:            fieldString(f, "Microsoft.VSTS.Common.Risk"),
		TimeCriticality: fieldFloatPtr(f, "Microsoft.VSTS.Common.TimeCriticality"),

		// Status tracking
		ActivatedBy: fieldString(f, "Microsoft.VSTS.Common.ActivatedBy"),
		ResolvedBy:  fieldString(f, "Microsoft.VSTS.Common.ResolvedBy"),
		ClosedBy:    fieldString(f, "Microsoft.VSTS.Common.ClosedBy"),

		// Build integration
		FoundIn:          fieldString(f, "Microsoft.VSTS.Build.FoundIn"),
		IntegrationBuild: fieldString(f, "Microsoft.VSTS.Build.IntegrationBuild"),

		// Custom fields
		AtRisk:     fieldBoolPtr(f, "Custom.AtRisk"),
		CVENumber:  fieldString(f, "Custom.CVENumber"),
		DevOwner:   fieldString(f, "Custom.DevOwner"),
		MarketDate: fieldDateTime(f, "Custom.MarketDate"),
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

// buildDateFieldOps constructs JSON patch operations from DateFields.
func buildDateFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *DateFields) {
	if fields == nil {
		return
	}

	addDateField(ops, operation, &fieldPathStartDate, fields.StartDate)
	addDateField(ops, operation, &fieldPathFinishDate, fields.FinishDate)
	addDateField(ops, operation, &fieldPathTargetDate, fields.TargetDate)
	addDateField(ops, operation, &fieldPathDueDate, fields.DueDate)
	addDateField(ops, operation, &fieldPathMarketDate, fields.MarketDate)
	addDateField(ops, operation, &fieldPathDevCompleteDate, fields.DevCompleteDate)
	addDateField(ops, operation, &fieldPathQCStartDate, fields.QCStartDate)
	addDateField(ops, operation, &fieldPathQCCompleteDate, fields.QCCompleteDate)
	addDateField(ops, operation, &fieldPathOriginalTargetDate, fields.OriginalTargetDate)
}

// buildStatusFieldOps constructs JSON patch operations from StatusFields.
func buildStatusFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *StatusFields) {
	if fields == nil {
		return
	}

	addStringField(ops, operation, &fieldPathActivatedBy, fields.ActivatedBy)
	addDateField(ops, operation, &fieldPathActivatedDate, fields.ActivatedDate)
	addStringField(ops, operation, &fieldPathResolvedBy, fields.ResolvedBy)
	addDateField(ops, operation, &fieldPathResolvedDate, fields.ResolvedDate)
	addStringField(ops, operation, &fieldPathResolvedReason, fields.ResolvedReason)
	addStringField(ops, operation, &fieldPathClosedBy, fields.ClosedBy)
	addDateField(ops, operation, &fieldPathClosedDate, fields.ClosedDate)
	addDateField(ops, operation, &fieldPathStateChangeDate, fields.StateChangeDate)
}

// buildPlanningFieldOps constructs JSON patch operations from PlanningFields.
func buildPlanningFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *PlanningFields) {
	if fields == nil {
		return
	}

	addOptionalIntField(ops, operation, &fieldPathBusinessValue, fields.BusinessValue)
	addOptionalFloatField(ops, operation, &fieldPathStackRank, fields.StackRank)
	addStringField(ops, operation, &fieldPathRisk, fields.Risk)
	addOptionalFloatField(ops, operation, &fieldPathTimeCriticality, fields.TimeCriticality)
	addStringField(ops, operation, &fieldPathRating, fields.Rating)
	addStringField(ops, operation, &fieldPathTriage, fields.Triage)
}

// buildBuildFieldOps constructs JSON patch operations from BuildFields.
func buildBuildFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *BuildFields) {
	if fields == nil {
		return
	}

	addStringField(ops, operation, &fieldPathFoundIn, fields.FoundIn)
	addStringField(ops, operation, &fieldPathIntegrationBuild, fields.IntegrationBuild)
	addStringField(ops, operation, &fieldPathClosedInBuild, fields.ClosedInBuild)
}

// buildSalesforceFieldOps constructs JSON patch operations from SalesforceFields.
func buildSalesforceFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *SalesforceFields) {
	if fields == nil {
		return
	}

	addStringField(ops, operation, &fieldPathSalesforceCaseNumber, fields.CaseNumber)
	addStringField(ops, operation, &fieldPathSalesforceCaseStatus, fields.CaseStatus)
	addBoolField(ops, operation, &fieldPathSalesforceCaseClosed, fields.CaseClosed)
	addStringField(ops, operation, &fieldPathSalesforceURL, fields.URL)
	addStringField(ops, operation, &fieldPathSalesforceEscalatingEngineer, fields.EscalatingEngineer)
	addStringField(ops, operation, &fieldPathSalesforceEscalationReason, fields.EscalationReason)
	addStringField(ops, operation, &fieldPathEscalationAttachmentsFolder, fields.AttachmentsFolder)
}

// buildRequirementFieldOps constructs JSON patch operations from RequirementFields.
func buildRequirementFieldOps(
	ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *RequirementFields,
) {
	if fields == nil {
		return
	}

	addHTMLField(ops, operation, &fieldPathFunctionalRequirements, fields.FunctionalRequirements)
	addHTMLField(ops, operation, &fieldPathNonfunctionalRequirements, fields.NonfunctionalRequirements)
	addHTMLField(ops, operation, &fieldPathBusinessCase, fields.BusinessCase)
	addHTMLField(ops, operation, &fieldPathSuggestedTests, fields.SuggestedTests)
	addHTMLField(ops, operation, &fieldPathRejectedIdeas, fields.RejectedIdeas)
	addHTMLField(ops, operation, &fieldPathResources, fields.Resources)
}

// buildQualityFieldOps constructs JSON patch operations from QualityFields.
func buildQualityFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *QualityFields) {
	if fields == nil {
		return
	}

	addStringField(ops, operation, &fieldPathApprovedBy, fields.ApprovedBy)
	addStringField(ops, operation, &fieldPathReviewedBy, fields.ReviewedBy)
	addStringField(ops, operation, &fieldPathInitialDetailQuality, fields.InitialDetailQuality)
	addStringField(ops, operation, &fieldPathInitialDetailQualityComment, fields.InitialDetailQualityComment)
	addStringField(ops, operation, &fieldPathDocumentation, fields.Documentation)
	addStringField(ops, operation, &fieldPathRCAReason, fields.RCAReason)
}

// buildMetricsFieldOps constructs JSON patch operations from MetricsFields.
func buildMetricsFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *MetricsFields) {
	if fields == nil {
		return
	}

	addOptionalIntField(ops, operation, &fieldPathDaysAwaitingDev, fields.DaysAwaitingDev)
	addOptionalIntField(ops, operation, &fieldPathDaysAwaitingSupport, fields.DaysAwaitingSupport)
	addOptionalIntField(ops, operation, &fieldPathDaysSinceLastDevUpdate, fields.DaysSinceLastDevUpdate)
	addOptionalFloatField(ops, operation, &fieldPathTimeSpent, fields.TimeSpent)
	addOptionalIntField(ops, operation, &fieldPathPrioritizationScore, fields.PrioritizationScore)
	addOptionalIntField(ops, operation, &fieldPathConfidence, fields.Confidence)
	addDateField(ops, operation, &fieldPathRemainingWorkChangedDate, fields.RemainingWorkChangedDate)
}

// buildSecurityFieldOps constructs JSON patch operations from SecurityFields.
func buildSecurityFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *SecurityFields) {
	if fields == nil {
		return
	}

	addStringField(ops, operation, &fieldPathCVENumber, fields.CVENumber)
	addStringField(ops, operation, &fieldPathVulnerabilitySource, fields.VulnerabilitySource)
}

// buildFeatureFieldOps constructs JSON patch operations from FeatureSpecificFields.
func buildFeatureFieldOps(
	ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *FeatureSpecificFields,
) {
	if fields == nil {
		return
	}

	addBoolField(ops, operation, &fieldPathAtRisk, fields.AtRisk)
	addStringField(ops, operation, &fieldPathDeliveryRisk, fields.DeliveryRisk)
	addStringField(ops, operation, &fieldPathRiskReason, fields.RiskReason)
	addHTMLField(ops, operation, &fieldPathMitigationPlan, fields.MitigationPlan)
}

// buildBugFieldOps constructs JSON patch operations from BugSpecificFields.
func buildBugFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *BugSpecificFields) {
	if fields == nil {
		return
	}

	addHTMLField(ops, operation, &fieldPathReproSteps, fields.ReproSteps)
	addHTMLField(ops, operation, &fieldPathSystemInfo, fields.SystemInfo)
	addStringField(ops, operation, &fieldPathBlocked, fields.Blocked)
	addHTMLField(ops, operation, &fieldPathProposedFix, fields.ProposedFix)
}

// buildUserStoryFieldOps constructs JSON patch operations from UserStorySpecificFields.
func buildUserStoryFieldOps(
	ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *UserStorySpecificFields,
) {
	if fields == nil {
		return
	}

	addStringField(ops, operation, &fieldPathDevOwner, fields.DevOwner)
	addStringField(ops, operation, &fieldPathPoker, fields.Poker)
}

// buildTestCaseFieldOps constructs JSON patch operations from TestCaseSpecificFields.
func buildTestCaseFieldOps(
	ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *TestCaseSpecificFields,
) {
	if fields == nil {
		return
	}

	addHTMLField(ops, operation, &fieldPathSteps, fields.Steps)
	addStringField(ops, operation, &fieldPathAutomatedTestName, fields.AutomatedTestName)
	addStringField(ops, operation, &fieldPathAutomatedTestStorage, fields.AutomatedTestStorage)
	addStringField(ops, operation, &fieldPathAutomatedTestType, fields.AutomatedTestType)
	addStringField(ops, operation, &fieldPathAutomatedTestID, fields.AutomatedTestID)
	addStringField(ops, operation, &fieldPathAutomationStatus, fields.AutomationStatus)
	addHTMLField(ops, operation, &fieldPathParameters, fields.Parameters)
	addHTMLField(ops, operation, &fieldPathLocalDataSource, fields.LocalDataSource)
}

// buildCodeReviewFieldOps constructs JSON patch operations from CodeReviewFields.
func buildCodeReviewFieldOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields *CodeReviewFields) {
	if fields == nil {
		return
	}

	addStringField(ops, operation, &fieldPathContext, fields.Context)
	addOptionalIntField(ops, operation, &fieldPathContextCode, fields.ContextCode)
	addStringField(ops, operation, &fieldPathContextOwner, fields.ContextOwner)
	addStringField(ops, operation, &fieldPathContextType, fields.ContextType)
	addStringField(ops, operation, &fieldPathAcceptedBy, fields.AcceptedBy)
	addDateField(ops, operation, &fieldPathAcceptedDate, fields.AcceptedDate)
	addStringField(ops, operation, &fieldPathClosedStatus, fields.ClosedStatus)
	addStringField(ops, operation, &fieldPathClosingComment, fields.ClosingComment)
}

// buildCommonOps adds patch operations for CommonFields.
func buildCommonOps(ops *[]webapi.JsonPatchOperation, operation *webapi.Operation, fields CommonFields) {
	// Existing fields
	addStringField(ops, operation, &fieldPathAssignedTo, fields.AssignedTo)
	addHTMLField(ops, operation, &fieldPathDescription, fields.Description)
	addStringField(ops, operation, &fieldPathIterationPath, fields.IterationPath)
	addStringField(ops, operation, &fieldPathAreaPath, fields.AreaPath)
	addStringField(ops, operation, &fieldPathSize, fields.Size)
	addStringField(ops, operation, &fieldPathSeverity, fields.Severity)
	addStringField(ops, operation, &fieldPathActivity, fields.Activity)
	addStringField(ops, operation, &fieldPathValueArea, fields.ValueArea)

	addIntField(ops, operation, &fieldPathPriority, fields.Priority)

	addFloatField(ops, operation, &fieldPathStoryPoints, fields.StoryPoints)
	addFloatField(ops, operation, &fieldPathOriginalEstimate, fields.OriginalEstimate)
	addFloatField(ops, operation, &fieldPathCompletedWork, fields.CompletedWork)
	addFloatField(ops, operation, &fieldPathRemainingWork, fields.RemainingWork)
	addFloatField(ops, operation, &fieldPathEffort, fields.Effort)

	// New: Date fields
	addDateField(ops, operation, &fieldPathStartDate, fields.StartDate)
	addDateField(ops, operation, &fieldPathFinishDate, fields.FinishDate)
	addDateField(ops, operation, &fieldPathTargetDate, fields.TargetDate)

	// New: Planning fields
	addOptionalIntField(ops, operation, &fieldPathBusinessValue, fields.BusinessValue)
	addOptionalFloatField(ops, operation, &fieldPathStackRank, fields.StackRank)
	addStringField(ops, operation, &fieldPathRisk, fields.Risk)

	// New: Build fields
	addStringField(ops, operation, &fieldPathFoundIn, fields.FoundIn)
	addStringField(ops, operation, &fieldPathIntegrationBuild, fields.IntegrationBuild)
}

// buildUpdateOps converts UpdateOptions into a JSON patch operation slice.
func buildUpdateOps(opts UpdateOptions) []webapi.JsonPatchOperation {
	replace := webapi.OperationValues.Replace

	var ops []webapi.JsonPatchOperation

	// Update-specific fields
	addStringField(&ops, &replace, &fieldPathTitle, opts.Title)
	addStringField(&ops, &replace, &fieldPathState, opts.State)
	addHTMLField(&ops, &replace, &fieldPathAcceptanceCriteria, opts.AcceptanceCriteria)
	addStringField(&ops, &replace, &fieldPathReason, opts.Reason)
	addStringField(&ops, &replace, &fieldPathTags, opts.Tags)

	// Common fields
	buildCommonOps(&ops, &replace, opts.CommonFields)

	// Optional field groups
	buildDateFieldOps(&ops, &replace, opts.DateFields)
	buildStatusFieldOps(&ops, &replace, opts.StatusFields)
	buildPlanningFieldOps(&ops, &replace, opts.PlanningFields)
	buildBuildFieldOps(&ops, &replace, opts.BuildFields)
	buildSalesforceFieldOps(&ops, &replace, opts.SalesforceFields)
	buildRequirementFieldOps(&ops, &replace, opts.RequirementFields)
	buildQualityFieldOps(&ops, &replace, opts.QualityFields)
	buildMetricsFieldOps(&ops, &replace, opts.MetricsFields)
	buildSecurityFieldOps(&ops, &replace, opts.SecurityFields)

	// Type-specific fields
	buildFeatureFieldOps(&ops, &replace, opts.FeatureFields)
	buildBugFieldOps(&ops, &replace, opts.BugFields)
	buildUserStoryFieldOps(&ops, &replace, opts.UserStoryFields)
	buildTestCaseFieldOps(&ops, &replace, opts.TestCaseFields)
	buildCodeReviewFieldOps(&ops, &replace, opts.CodeReviewFields)

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

// fieldDateTime extracts a DateTime field from the work item fields map.
// Returns nil if the field is not present or cannot be parsed.
func fieldDateTime(f *map[string]any, key string) *time.Time {
	val, ok := (*f)[key]
	if !ok || val == nil {
		return nil
	}

	str, ok := val.(string)
	if !ok {
		return nil
	}

	// Try multiple formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return &t
		}
	}

	return nil
}

// fieldBoolPtr extracts a boolean field and returns a pointer to it.
// Returns nil if the field is not present.
func fieldBoolPtr(f *map[string]any, key string) *bool {
	val, ok := (*f)[key]
	if !ok || val == nil {
		return nil
	}

	b, ok := val.(bool)
	if !ok {
		return nil
	}

	return &b
}

// fieldIntPtr extracts an int field and returns a pointer to it.
// Returns nil if the field is not present or is zero.
func fieldIntPtr(f *map[string]any, key string) *int {
	i := fieldInt(f, key)
	if i == 0 {
		return nil
	}

	return &i
}

// fieldFloatPtr extracts a float field and returns a pointer to it.
// Returns nil if the field is not present or is zero.
func fieldFloatPtr(f *map[string]any, key string) *float64 {
	fl := fieldFloat(f, key)
	if fl == 0 {
		return nil
	}

	return &fl
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

// containsHTMLTags checks if a string contains common HTML block-level tags or closing tags.
// Used to distinguish HTML input from Markdown input.
func containsHTMLTags(s string) bool {
	tags := []string{"<p>", "<div>", "<h1>", "<h2>", "<h3>", "<h4>", "<h5>", "<h6>", "<ul>", "<ol>", "</"}
	for _, tag := range tags {
		if strings.Contains(s, tag) {
			return true
		}
	}

	return false
}

// convertMarkdownToHTML converts a Markdown string to HTML using GitHub Flavored Markdown.
// Falls back to the original string if conversion fails.
func convertMarkdownToHTML(md string) string {
	if md == "" {
		return ""
	}

	var buf bytes.Buffer
	if err := mdConverter.Convert([]byte(md), &buf); err != nil {
		// Fallback to original on error
		return md
	}

	return buf.String()
}

// sanitizeHTML removes dangerous HTML tags and attributes using bluemonday's UGC policy.
// This prevents XSS attacks while allowing safe HTML formatting.
func sanitizeHTML(html string) string {
	return htmlSanitizer.Sanitize(html)
}

// prepareHTMLField converts Markdown to HTML (if needed) and sanitizes the result.
// If the input already contains HTML tags, it skips Markdown conversion.
// Always sanitizes the output to prevent XSS attacks.
func prepareHTMLField(input string) string {
	if input == "" {
		return ""
	}

	var html string
	if containsHTMLTags(input) {
		// Already HTML, skip conversion
		html = input
	} else {
		// Treat as Markdown and convert
		html = convertMarkdownToHTML(input)
	}

	// Always sanitize before sending to Azure DevOps
	return sanitizeHTML(html)
}
