// Package controller wires the MCP server, registers all tools, and starts
// the stdio transport. It is the only package that depends on both client and tools.
package controller

import (
	"context"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/markis/azure-devops-mcp/internal/client"
	"github.com/markis/azure-devops-mcp/internal/tools"
)

// Config holds the configuration required to run the MCP server.
type Config struct {
	OrgURL  string
	PAT     string
	Project string
}

// Run creates the ADO client, registers all tools, and starts the MCP server
// over stdio. It blocks until the client disconnects or ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	adoClient, err := client.NewClient(ctx, cfg.OrgURL, cfg.PAT)
	if err != nil {
		return fmt.Errorf("creating ADO client: %w", err)
	}

	h := tools.NewHandlers(adoClient, cfg.Project)
	srv := CreateServer()
	RegisterTools(srv, h)

	return srv.Run(ctx, &mcp.StdioTransport{})
}

// CreateServer creates and configures the MCP server instance.
func CreateServer() *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{
		Name:    "azure-devops-mcp",
		Version: "0.1.0",
	}, nil)
}

// RegisterTools registers all Azure DevOps work item tools with the MCP server.
func RegisterTools(srv *mcp.Server, h *tools.Handlers) {
	registerGetWorkItem(srv, h)
	registerListWorkItems(srv, h)
	registerListMyWorkItems(srv, h)
	registerCreateWorkItem(srv, h)
	registerUpdateWorkItem(srv, h)
	registerAddComment(srv, h)
}

// getWorkItemInput is the input structure for the get_work_item tool.
type getWorkItemInput struct {
	ID      FlexID `json:"id"                jsonschema:"work item ID or reference (required)"`
	Project string `json:"project,omitempty" jsonschema:"project name (optional, uses server default)"`
}

// registerGetWorkItem registers the get_work_item tool.
func registerGetWorkItem(srv *mcp.Server, h *tools.Handlers) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "get_work_item",
		Description: "Fetch a single Azure DevOps work item by numeric ID. " +
			"Returns title, state, type, assignee, description, story points, and other core fields.",
	}, func(
		ctx context.Context, _ *mcp.CallToolRequest, in getWorkItemInput,
	) (*mcp.CallToolResult, *client.WorkItem, error) {
		id := int(in.ID)

		if id <= 0 {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "id must be a positive integer",
				}},
			}, nil, nil
		}

		workItem, text, err := h.GetWorkItem(ctx, id, in.Project)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf(
						"could not retrieve work item %d: check the ID and project", id,
					),
				}},
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, workItem, nil
	})
}

// registerListWorkItems registers the list_work_items tool.
func registerListWorkItems(srv *mcp.Server, h *tools.Handlers) {
	type listWorkItemsInput struct {
		Query   string `json:"query"             jsonschema:"WIQL query to filter work items (required)"`
		Project string `json:"project,omitempty" jsonschema:"project name (optional, uses server default)"`
	}

	type listWorkItemsOutput struct {
		WorkItems []*client.WorkItemSummary `json:"work_items" jsonschema:"Work items matching query"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name: "list_work_items",
		Description: "Run a WIQL (Work Item Query Language) query and return matching Azure DevOps work items. " +
			"Returns a list with ID, title, state, type, and assignee for each work item.",
	}, func(
		ctx context.Context, _ *mcp.CallToolRequest, in listWorkItemsInput,
	) (*mcp.CallToolResult, *listWorkItemsOutput, error) {
		if in.Query == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "query parameter is required",
				}},
			}, nil, nil
		}

		workItems, text, err := h.ListWorkItems(ctx, in.Query, in.Project)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "could not execute WIQL query: check query syntax and project",
				}},
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, &listWorkItemsOutput{WorkItems: workItems}, nil
	})
}

// registerListMyWorkItems registers the list_my_work_items tool.
func registerListMyWorkItems(srv *mcp.Server, h *tools.Handlers) {
	type listMyWorkItemsInput struct {
		Project string `json:"project,omitempty" jsonschema:"project name (optional, uses server default)"`
	}

	type listMyWorkItemsOutput struct {
		WorkItems []*client.WorkItemSummary `json:"work_items" jsonschema:"Work items assigned to user"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name: "list_my_work_items",
		Description: "Return active work items assigned to the authenticated user. " +
			"Filters for work items in Active or New state. Returns ID, title, state, type, and assignee.",
	}, func(
		ctx context.Context, _ *mcp.CallToolRequest, in listMyWorkItemsInput,
	) (*mcp.CallToolResult, *listMyWorkItemsOutput, error) {
		workItems, text, err := h.ListMyWorkItems(ctx, in.Project)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "could not retrieve your work items: check authentication and project",
				}},
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, &listMyWorkItemsOutput{WorkItems: workItems}, nil
	})
}

// createBugInput is the input structure for creating Bug work items.
type createBugInput struct {
	Type        string `json:"type"`
	Title       string `json:"title"                 jsonschema:"work item title (required)"`
	Description string `json:"description,omitempty" jsonschema:"detailed description in plain text or HTML"`
	AssignedTo  string `json:"assigned_to,omitempty" jsonschema:"email or display name to assign to"`
	Tags        string `json:"tags,omitempty"        jsonschema:"semicolon-separated tags"`
	Project     string `json:"project,omitempty"     jsonschema:"project name (optional)"`

	// Core fields
	IterationPath    string    `json:"iteration_path,omitempty"    jsonschema:"iteration/sprint path"`
	AreaPath         string    `json:"area_path,omitempty"         jsonschema:"area path in the project"`
	Priority         FlexInt   `json:"priority,omitempty"          jsonschema:"priority level (1-4)"`
	StoryPoints      FlexFloat `json:"story_points,omitempty"      jsonschema:"story points estimate"`
	OriginalEstimate FlexFloat `json:"original_estimate,omitempty" jsonschema:"time estimate in hours"`
	CompletedWork    FlexFloat `json:"completed_work,omitempty"    jsonschema:"completed work in hours"`
	RemainingWork    FlexFloat `json:"remaining_work,omitempty"    jsonschema:"remaining work in hours"`
	Effort           FlexFloat `json:"effort,omitempty"            jsonschema:"effort in hours"`
	Size             string    `json:"size,omitempty"              jsonschema:"t-shirt size (S, M, L, XL)"`
	Severity         string    `json:"severity,omitempty"          jsonschema:"severity for Bug"`
	Activity         string    `json:"activity,omitempty"          jsonschema:"activity type (Development/Testing/etc)"`
	ValueArea        string    `json:"value_area,omitempty"        jsonschema:"value area (Business/Architectural)"`

	// Date fields
	StartDate          *FlexDateTime `json:"start_date,omitempty"           jsonschema:"planned start date (ISO 8601)"`
	FinishDate         *FlexDateTime `json:"finish_date,omitempty"          jsonschema:"planned finish date (ISO 8601)"`
	TargetDate         *FlexDateTime `json:"target_date,omitempty"          jsonschema:"target completion date (ISO 8601)"`
	DueDate            *FlexDateTime `json:"due_date,omitempty"             jsonschema:"hard deadline (ISO 8601)"`
	MarketDate         *FlexDateTime `json:"market_date,omitempty"          jsonschema:"market release date (ISO 8601)"`
	DevCompleteDate    *FlexDateTime `json:"dev_complete_date,omitempty"    jsonschema:"dev completion date (ISO 8601)"`
	QCStartDate        *FlexDateTime `json:"qc_start_date,omitempty"        jsonschema:"QC start date (ISO 8601)"`
	QCCompleteDate     *FlexDateTime `json:"qc_complete_date,omitempty"     jsonschema:"QC completion date (ISO 8601)"`
	OriginalTargetDate *FlexDateTime `json:"original_target_date,omitempty" jsonschema:"original target date (ISO 8601)"`

	// Planning fields
	BusinessValue   FlexInt   `json:"business_value,omitempty"   jsonschema:"business value score"`
	StackRank       FlexFloat `json:"stack_rank,omitempty"       jsonschema:"backlog ordering rank"`
	Risk            string    `json:"risk,omitempty"             jsonschema:"risk level"`
	TimeCriticality FlexFloat `json:"time_criticality,omitempty" jsonschema:"time sensitivity score"`

	// Build integration
	FoundIn          string `json:"found_in,omitempty"          jsonschema:"build where bug was found"`
	IntegrationBuild string `json:"integration_build,omitempty" jsonschema:"build with fix"`

	// Requirements
	FunctionalRequirements    string `json:"functional_requirements,omitempty"    jsonschema:"functional reqs (HTML)"`
	NonfunctionalRequirements string `json:"nonfunctional_requirements,omitempty" jsonschema:"non-func reqs (HTML)"`
	BusinessCase              string `json:"business_case,omitempty"              jsonschema:"business case (HTML)"`
	SuggestedTests            string `json:"suggested_tests,omitempty"            jsonschema:"suggested test cases (HTML)"`

	// Salesforce integration
	SalesforceCaseNumber         string `json:"salesforce_case_number,omitempty"         jsonschema:"Salesforce case no."`
	SalesforceURL                string `json:"salesforce_url,omitempty"                 jsonschema:"Salesforce case URL"`
	SalesforceEscalatingEngineer string `json:"salesforce_escalating_engineer,omitempty" jsonschema:"escalating engineer"`

	// Security
	CVENumber           string `json:"cve_number,omitempty"           jsonschema:"CVE identifier"`
	VulnerabilitySource string `json:"vulnerability_source,omitempty" jsonschema:"source of vulnerability"`

	// Quality
	ApprovedBy string `json:"approved_by,omitempty" jsonschema:"who approved"`
	RCAReason  string `json:"rca_reason,omitempty"  jsonschema:"root cause analysis reason"`

	// Metrics
	Confidence          FlexInt `json:"confidence,omitempty"           jsonschema:"confidence level"`
	PrioritizationScore FlexInt `json:"prioritization_score,omitempty" jsonschema:"prioritization score"`

	// Fields for Bug work items
	SystemInfo  string `json:"system_info,omitempty"  jsonschema:"system information for bug (HTML)"`
	Blocked     string `json:"blocked,omitempty"      jsonschema:"is work blocked"`
	ProposedFix string `json:"proposed_fix,omitempty" jsonschema:"proposed solution (HTML)"`
}

// createFeatureInput is the input structure for creating Feature work items.
type createFeatureInput struct {
	Type        string `json:"type"`
	Title       string `json:"title"                 jsonschema:"work item title (required)"`
	Description string `json:"description,omitempty" jsonschema:"detailed description in plain text or HTML"`
	AssignedTo  string `json:"assigned_to,omitempty" jsonschema:"email or display name to assign to"`
	Tags        string `json:"tags,omitempty"        jsonschema:"semicolon-separated tags"`
	Project     string `json:"project,omitempty"     jsonschema:"project name (optional)"`

	// Core fields (same as Bug)
	IterationPath    string    `json:"iteration_path,omitempty"    jsonschema:"iteration/sprint path"`
	AreaPath         string    `json:"area_path,omitempty"         jsonschema:"area path in the project"`
	Priority         FlexInt   `json:"priority,omitempty"          jsonschema:"priority level (1-4)"`
	StoryPoints      FlexFloat `json:"story_points,omitempty"      jsonschema:"story points estimate"`
	OriginalEstimate FlexFloat `json:"original_estimate,omitempty" jsonschema:"time estimate in hours"`
	CompletedWork    FlexFloat `json:"completed_work,omitempty"    jsonschema:"completed work in hours"`
	RemainingWork    FlexFloat `json:"remaining_work,omitempty"    jsonschema:"remaining work in hours"`
	Effort           FlexFloat `json:"effort,omitempty"            jsonschema:"effort in hours"`
	Size             string    `json:"size,omitempty"              jsonschema:"t-shirt size (S, M, L, XL)"`
	Severity         string    `json:"severity,omitempty"          jsonschema:"severity level"`
	Activity         string    `json:"activity,omitempty"          jsonschema:"activity type (Development/Testing/etc)"`
	ValueArea        string    `json:"value_area,omitempty"        jsonschema:"value area (Business/Architectural)"`

	// Date fields (same as Bug)
	StartDate          *FlexDateTime `json:"start_date,omitempty"           jsonschema:"planned start date (ISO 8601)"`
	FinishDate         *FlexDateTime `json:"finish_date,omitempty"          jsonschema:"planned finish date (ISO 8601)"`
	TargetDate         *FlexDateTime `json:"target_date,omitempty"          jsonschema:"target completion date (ISO 8601)"`
	DueDate            *FlexDateTime `json:"due_date,omitempty"             jsonschema:"hard deadline (ISO 8601)"`
	MarketDate         *FlexDateTime `json:"market_date,omitempty"          jsonschema:"market release date (ISO 8601)"`
	DevCompleteDate    *FlexDateTime `json:"dev_complete_date,omitempty"    jsonschema:"dev completion date (ISO 8601)"`
	QCStartDate        *FlexDateTime `json:"qc_start_date,omitempty"        jsonschema:"QC start date (ISO 8601)"`
	QCCompleteDate     *FlexDateTime `json:"qc_complete_date,omitempty"     jsonschema:"QC completion date (ISO 8601)"`
	OriginalTargetDate *FlexDateTime `json:"original_target_date,omitempty" jsonschema:"original target date (ISO 8601)"`

	// Planning fields (same as Bug)
	BusinessValue   FlexInt   `json:"business_value,omitempty"   jsonschema:"business value score"`
	StackRank       FlexFloat `json:"stack_rank,omitempty"       jsonschema:"backlog ordering rank"`
	Risk            string    `json:"risk,omitempty"             jsonschema:"risk level"`
	TimeCriticality FlexFloat `json:"time_criticality,omitempty" jsonschema:"time sensitivity score"`

	// Build integration (same as Bug)
	FoundIn          string `json:"found_in,omitempty"          jsonschema:"build where bug was found"`
	IntegrationBuild string `json:"integration_build,omitempty" jsonschema:"build with fix"`

	// Requirements (same as Bug)
	FunctionalRequirements    string `json:"functional_requirements,omitempty"    jsonschema:"functional reqs (HTML)"`
	NonfunctionalRequirements string `json:"nonfunctional_requirements,omitempty" jsonschema:"non-func reqs (HTML)"`
	BusinessCase              string `json:"business_case,omitempty"              jsonschema:"business case (HTML)"`
	SuggestedTests            string `json:"suggested_tests,omitempty"            jsonschema:"suggested test cases (HTML)"`

	// Salesforce integration (same as Bug)
	SalesforceCaseNumber         string `json:"salesforce_case_number,omitempty"         jsonschema:"Salesforce case no."`
	SalesforceURL                string `json:"salesforce_url,omitempty"                 jsonschema:"Salesforce case URL"`
	SalesforceEscalatingEngineer string `json:"salesforce_escalating_engineer,omitempty" jsonschema:"escalating engineer"`

	// Security (same as Bug)
	CVENumber           string `json:"cve_number,omitempty"           jsonschema:"CVE identifier"`
	VulnerabilitySource string `json:"vulnerability_source,omitempty" jsonschema:"source of vulnerability"`

	// Quality (same as Bug)
	ApprovedBy string `json:"approved_by,omitempty" jsonschema:"who approved"`
	RCAReason  string `json:"rca_reason,omitempty"  jsonschema:"root cause analysis reason"`

	// Metrics (same as Bug)
	Confidence          FlexInt `json:"confidence,omitempty"           jsonschema:"confidence level"`
	PrioritizationScore FlexInt `json:"prioritization_score,omitempty" jsonschema:"prioritization score"`

	// Feature-specific fields
	AtRisk         FlexBool `json:"at_risk"                   jsonschema:"is feature at risk (required for Feature)"`
	Documentation  string   `json:"documentation"             jsonschema:"documentation status (required for Feature)"`
	DeliveryRisk   string   `json:"delivery_risk,omitempty"   jsonschema:"delivery risk level"`
	RiskReason     string   `json:"risk_reason,omitempty"     jsonschema:"reason for risk"`
	MitigationPlan string   `json:"mitigation_plan,omitempty" jsonschema:"risk mitigation plan (HTML)"`
}

// createUserStoryInput is the input structure for creating User Story work items.
type createUserStoryInput struct {
	Type        string `json:"type"`
	Title       string `json:"title"                 jsonschema:"work item title (required)"`
	Description string `json:"description,omitempty" jsonschema:"detailed description in plain text or HTML"`
	AssignedTo  string `json:"assigned_to,omitempty" jsonschema:"email or display name to assign to"`
	Tags        string `json:"tags,omitempty"        jsonschema:"semicolon-separated tags"`
	Project     string `json:"project,omitempty"     jsonschema:"project name (optional)"`

	// Core fields (same as Bug)
	IterationPath    string    `json:"iteration_path,omitempty"    jsonschema:"iteration/sprint path"`
	AreaPath         string    `json:"area_path,omitempty"         jsonschema:"area path in the project"`
	Priority         FlexInt   `json:"priority,omitempty"          jsonschema:"priority level (1-4)"`
	StoryPoints      FlexFloat `json:"story_points,omitempty"      jsonschema:"story points estimate"`
	OriginalEstimate FlexFloat `json:"original_estimate,omitempty" jsonschema:"time estimate in hours"`
	CompletedWork    FlexFloat `json:"completed_work,omitempty"    jsonschema:"completed work in hours"`
	RemainingWork    FlexFloat `json:"remaining_work,omitempty"    jsonschema:"remaining work in hours"`
	Effort           FlexFloat `json:"effort,omitempty"            jsonschema:"effort in hours"`
	Size             string    `json:"size,omitempty"              jsonschema:"t-shirt size (S, M, L, XL)"`
	Severity         string    `json:"severity,omitempty"          jsonschema:"severity level"`
	Activity         string    `json:"activity,omitempty"          jsonschema:"activity type (Development/Testing/etc)"`
	ValueArea        string    `json:"value_area,omitempty"        jsonschema:"value area (Business/Architectural)"`

	// Date fields (same as Bug)
	StartDate          *FlexDateTime `json:"start_date,omitempty"           jsonschema:"planned start date (ISO 8601)"`
	FinishDate         *FlexDateTime `json:"finish_date,omitempty"          jsonschema:"planned finish date (ISO 8601)"`
	TargetDate         *FlexDateTime `json:"target_date,omitempty"          jsonschema:"target completion date (ISO 8601)"`
	DueDate            *FlexDateTime `json:"due_date,omitempty"             jsonschema:"hard deadline (ISO 8601)"`
	MarketDate         *FlexDateTime `json:"market_date,omitempty"          jsonschema:"market release date (ISO 8601)"`
	DevCompleteDate    *FlexDateTime `json:"dev_complete_date,omitempty"    jsonschema:"dev completion date (ISO 8601)"`
	QCStartDate        *FlexDateTime `json:"qc_start_date,omitempty"        jsonschema:"QC start date (ISO 8601)"`
	QCCompleteDate     *FlexDateTime `json:"qc_complete_date,omitempty"     jsonschema:"QC completion date (ISO 8601)"`
	OriginalTargetDate *FlexDateTime `json:"original_target_date,omitempty" jsonschema:"original target date (ISO 8601)"`

	// Planning fields (same as Bug)
	BusinessValue   FlexInt   `json:"business_value,omitempty"   jsonschema:"business value score"`
	StackRank       FlexFloat `json:"stack_rank,omitempty"       jsonschema:"backlog ordering rank"`
	Risk            string    `json:"risk,omitempty"             jsonschema:"risk level"`
	TimeCriticality FlexFloat `json:"time_criticality,omitempty" jsonschema:"time sensitivity score"`

	// Build integration (same as Bug)
	FoundIn          string `json:"found_in,omitempty"          jsonschema:"build where bug was found"`
	IntegrationBuild string `json:"integration_build,omitempty" jsonschema:"build with fix"`

	// Requirements (same as Bug)
	FunctionalRequirements    string `json:"functional_requirements,omitempty"    jsonschema:"functional reqs (HTML)"`
	NonfunctionalRequirements string `json:"nonfunctional_requirements,omitempty" jsonschema:"non-func reqs (HTML)"`
	BusinessCase              string `json:"business_case,omitempty"              jsonschema:"business case (HTML)"`
	SuggestedTests            string `json:"suggested_tests,omitempty"            jsonschema:"suggested test cases (HTML)"`

	// Salesforce integration (same as Bug)
	SalesforceCaseNumber         string `json:"salesforce_case_number,omitempty"         jsonschema:"Salesforce case no."`
	SalesforceURL                string `json:"salesforce_url,omitempty"                 jsonschema:"Salesforce case URL"`
	SalesforceEscalatingEngineer string `json:"salesforce_escalating_engineer,omitempty" jsonschema:"escalating engineer"`

	// Security (same as Bug)
	CVENumber           string `json:"cve_number,omitempty"           jsonschema:"CVE identifier"`
	VulnerabilitySource string `json:"vulnerability_source,omitempty" jsonschema:"source of vulnerability"`

	// Quality (same as Bug)
	ApprovedBy string `json:"approved_by,omitempty" jsonschema:"who approved"`
	RCAReason  string `json:"rca_reason,omitempty"  jsonschema:"root cause analysis reason"`

	// Metrics (same as Bug)
	Confidence          FlexInt `json:"confidence,omitempty"           jsonschema:"confidence level"`
	PrioritizationScore FlexInt `json:"prioritization_score,omitempty" jsonschema:"prioritization score"`

	// User Story-specific fields
	DevOwner string `json:"dev_owner,omitempty" jsonschema:"development owner"`
	Poker    string `json:"poker,omitempty"     jsonschema:"planning poker estimate"`
}

// createTaskInput is the input structure for creating Task work items.
type createTaskInput struct {
	Type        string `json:"type"`
	Title       string `json:"title"                 jsonschema:"work item title (required)"`
	Description string `json:"description,omitempty" jsonschema:"detailed description in plain text or HTML"`
	AssignedTo  string `json:"assigned_to,omitempty" jsonschema:"email or display name to assign to"`
	Tags        string `json:"tags,omitempty"        jsonschema:"semicolon-separated tags"`
	Project     string `json:"project,omitempty"     jsonschema:"project name (optional)"`

	// Core fields (same as Bug)
	IterationPath    string    `json:"iteration_path,omitempty"    jsonschema:"iteration/sprint path"`
	AreaPath         string    `json:"area_path,omitempty"         jsonschema:"area path in the project"`
	Priority         FlexInt   `json:"priority,omitempty"          jsonschema:"priority level (1-4)"`
	StoryPoints      FlexFloat `json:"story_points,omitempty"      jsonschema:"story points estimate"`
	OriginalEstimate FlexFloat `json:"original_estimate,omitempty" jsonschema:"time estimate in hours (for Tasks)"`
	CompletedWork    FlexFloat `json:"completed_work,omitempty"    jsonschema:"completed work in hours (for Tasks)"`
	RemainingWork    FlexFloat `json:"remaining_work,omitempty"    jsonschema:"remaining work in hours (for Tasks)"`
	Effort           FlexFloat `json:"effort,omitempty"            jsonschema:"effort in hours"`
	Size             string    `json:"size,omitempty"              jsonschema:"t-shirt size (S, M, L, XL)"`
	Severity         string    `json:"severity,omitempty"          jsonschema:"severity level"`
	Activity         string    `json:"activity,omitempty"          jsonschema:"activity type (Development/Testing/etc)"`
	ValueArea        string    `json:"value_area,omitempty"        jsonschema:"value area (Business/Architectural)"`

	// Date fields (same as Bug)
	StartDate          *FlexDateTime `json:"start_date,omitempty"           jsonschema:"planned start date (ISO 8601)"`
	FinishDate         *FlexDateTime `json:"finish_date,omitempty"          jsonschema:"planned finish date (ISO 8601)"`
	TargetDate         *FlexDateTime `json:"target_date,omitempty"          jsonschema:"target completion date (ISO 8601)"`
	DueDate            *FlexDateTime `json:"due_date,omitempty"             jsonschema:"hard deadline (ISO 8601)"`
	MarketDate         *FlexDateTime `json:"market_date,omitempty"          jsonschema:"market release date (ISO 8601)"`
	DevCompleteDate    *FlexDateTime `json:"dev_complete_date,omitempty"    jsonschema:"dev completion date (ISO 8601)"`
	QCStartDate        *FlexDateTime `json:"qc_start_date,omitempty"        jsonschema:"QC start date (ISO 8601)"`
	QCCompleteDate     *FlexDateTime `json:"qc_complete_date,omitempty"     jsonschema:"QC completion date (ISO 8601)"`
	OriginalTargetDate *FlexDateTime `json:"original_target_date,omitempty" jsonschema:"original target date (ISO 8601)"`

	// Planning fields (same as Bug)
	BusinessValue   FlexInt   `json:"business_value,omitempty"   jsonschema:"business value score"`
	StackRank       FlexFloat `json:"stack_rank,omitempty"       jsonschema:"backlog ordering rank"`
	Risk            string    `json:"risk,omitempty"             jsonschema:"risk level"`
	TimeCriticality FlexFloat `json:"time_criticality,omitempty" jsonschema:"time sensitivity score"`

	// Build integration (same as Bug)
	FoundIn          string `json:"found_in,omitempty"          jsonschema:"build where bug was found"`
	IntegrationBuild string `json:"integration_build,omitempty" jsonschema:"build with fix"`

	// Requirements (same as Bug)
	FunctionalRequirements    string `json:"functional_requirements,omitempty"    jsonschema:"functional reqs (HTML)"`
	NonfunctionalRequirements string `json:"nonfunctional_requirements,omitempty" jsonschema:"non-func reqs (HTML)"`
	BusinessCase              string `json:"business_case,omitempty"              jsonschema:"business case (HTML)"`
	SuggestedTests            string `json:"suggested_tests,omitempty"            jsonschema:"suggested test cases (HTML)"`

	// Salesforce integration (same as Bug)
	SalesforceCaseNumber         string `json:"salesforce_case_number,omitempty"         jsonschema:"Salesforce case no."`
	SalesforceURL                string `json:"salesforce_url,omitempty"                 jsonschema:"Salesforce case URL"`
	SalesforceEscalatingEngineer string `json:"salesforce_escalating_engineer,omitempty" jsonschema:"escalating engineer"`

	// Security (same as Bug)
	CVENumber           string `json:"cve_number,omitempty"           jsonschema:"CVE identifier"`
	VulnerabilitySource string `json:"vulnerability_source,omitempty" jsonschema:"source of vulnerability"`

	// Quality (same as Bug)
	ApprovedBy string `json:"approved_by,omitempty" jsonschema:"who approved"`
	RCAReason  string `json:"rca_reason,omitempty"  jsonschema:"root cause analysis reason"`

	// Metrics (same as Bug)
	Confidence          FlexInt `json:"confidence,omitempty"           jsonschema:"confidence level"`
	PrioritizationScore FlexInt `json:"prioritization_score,omitempty" jsonschema:"prioritization score"`
}

// createOtherInput is the input structure for creating work items of any type (fallback for Epic, Test Case, etc.).
// This struct contains all possible fields to support any Azure DevOps work item type.
type createOtherInput struct {
	Type             string    `json:"type"                        jsonschema:"work item type (required)"`
	Title            string    `json:"title"                       jsonschema:"work item title (required)"`
	Description      string    `json:"description,omitempty"       jsonschema:"detailed description in plain text or HTML"`
	AssignedTo       string    `json:"assigned_to,omitempty"       jsonschema:"email or display name to assign to"`
	Tags             string    `json:"tags,omitempty"              jsonschema:"semicolon-separated tags"`
	IterationPath    string    `json:"iteration_path,omitempty"    jsonschema:"iteration/sprint path"`
	AreaPath         string    `json:"area_path,omitempty"         jsonschema:"area path in the project"`
	Priority         FlexInt   `json:"priority,omitempty"          jsonschema:"priority level (1-4)"`
	StoryPoints      FlexFloat `json:"story_points,omitempty"      jsonschema:"story points estimate (for User Stories)"`
	OriginalEstimate FlexFloat `json:"original_estimate,omitempty" jsonschema:"time estimate in hours (for Tasks)"`
	CompletedWork    FlexFloat `json:"completed_work,omitempty"    jsonschema:"completed work in hours (for Tasks)"`
	RemainingWork    FlexFloat `json:"remaining_work,omitempty"    jsonschema:"remaining work in hours (for Tasks)"`
	Effort           FlexFloat `json:"effort,omitempty"            jsonschema:"effort in hours"`
	Size             string    `json:"size,omitempty"              jsonschema:"t-shirt size (S, M, L, XL)"`
	Severity         string    `json:"severity,omitempty"          jsonschema:"severity for Bug/Vulnerability"`
	Activity         string    `json:"activity,omitempty"          jsonschema:"activity type (Development/Testing/etc)"`
	ValueArea        string    `json:"value_area,omitempty"        jsonschema:"value area (Business/Architectural)"`

	// Date fields
	StartDate          *FlexDateTime `json:"start_date,omitempty"           jsonschema:"planned start date (ISO 8601)"`
	FinishDate         *FlexDateTime `json:"finish_date,omitempty"          jsonschema:"planned finish date (ISO 8601)"`
	TargetDate         *FlexDateTime `json:"target_date,omitempty"          jsonschema:"target completion date (ISO 8601)"`
	DueDate            *FlexDateTime `json:"due_date,omitempty"             jsonschema:"hard deadline (ISO 8601)"`
	MarketDate         *FlexDateTime `json:"market_date,omitempty"          jsonschema:"market release date (ISO 8601)"`
	DevCompleteDate    *FlexDateTime `json:"dev_complete_date,omitempty"    jsonschema:"dev completion date (ISO 8601)"`
	QCStartDate        *FlexDateTime `json:"qc_start_date,omitempty"        jsonschema:"QC start date (ISO 8601)"`
	QCCompleteDate     *FlexDateTime `json:"qc_complete_date,omitempty"     jsonschema:"QC completion date (ISO 8601)"`
	OriginalTargetDate *FlexDateTime `json:"original_target_date,omitempty" jsonschema:"original target date (ISO 8601)"`

	// Planning fields
	BusinessValue   FlexInt   `json:"business_value,omitempty"   jsonschema:"business value score"`
	StackRank       FlexFloat `json:"stack_rank,omitempty"       jsonschema:"backlog ordering rank"`
	Risk            string    `json:"risk,omitempty"             jsonschema:"risk level"`
	TimeCriticality FlexFloat `json:"time_criticality,omitempty" jsonschema:"time sensitivity score"`

	// Build integration
	FoundIn          string `json:"found_in,omitempty"          jsonschema:"build where bug was found"`
	IntegrationBuild string `json:"integration_build,omitempty" jsonschema:"build with fix"`

	// Feature-specific
	AtRisk         FlexBool `json:"at_risk,omitempty"         jsonschema:"is feature at risk (required for Feature)"`
	Documentation  string   `json:"documentation,omitempty"   jsonschema:"documentation status (required for Feature)"`
	DeliveryRisk   string   `json:"delivery_risk,omitempty"   jsonschema:"delivery risk level"`
	RiskReason     string   `json:"risk_reason,omitempty"     jsonschema:"reason for risk"`
	MitigationPlan string   `json:"mitigation_plan,omitempty" jsonschema:"risk mitigation plan (HTML)"`

	// Fields for Bug work items
	SystemInfo  string `json:"system_info,omitempty"  jsonschema:"system information for bug (HTML)"`
	Blocked     string `json:"blocked,omitempty"      jsonschema:"is work blocked"`
	ProposedFix string `json:"proposed_fix,omitempty" jsonschema:"proposed solution (HTML)"`

	// User Story-specific
	DevOwner string `json:"dev_owner,omitempty" jsonschema:"development owner"`
	Poker    string `json:"poker,omitempty"     jsonschema:"planning poker estimate"`

	// Requirements
	FunctionalRequirements    string `json:"functional_requirements,omitempty"    jsonschema:"functional reqs (HTML)"`
	NonfunctionalRequirements string `json:"nonfunctional_requirements,omitempty" jsonschema:"non-func reqs (HTML)"`
	BusinessCase              string `json:"business_case,omitempty"              jsonschema:"business case (HTML)"`
	SuggestedTests            string `json:"suggested_tests,omitempty"            jsonschema:"suggested test cases (HTML)"`

	// Salesforce integration
	SalesforceCaseNumber         string `json:"salesforce_case_number,omitempty"         jsonschema:"Salesforce case no."`
	SalesforceURL                string `json:"salesforce_url,omitempty"                 jsonschema:"Salesforce case URL"`
	SalesforceEscalatingEngineer string `json:"salesforce_escalating_engineer,omitempty" jsonschema:"escalating engineer"`

	// Security
	CVENumber           string `json:"cve_number,omitempty"           jsonschema:"CVE identifier"`
	VulnerabilitySource string `json:"vulnerability_source,omitempty" jsonschema:"source of vulnerability"`

	// Quality
	ApprovedBy string `json:"approved_by,omitempty" jsonschema:"who approved"`
	RCAReason  string `json:"rca_reason,omitempty"  jsonschema:"root cause analysis reason"`

	// Metrics
	Confidence          FlexInt `json:"confidence,omitempty"           jsonschema:"confidence level"`
	PrioritizationScore FlexInt `json:"prioritization_score,omitempty" jsonschema:"prioritization score"`

	Project string `json:"project,omitempty" jsonschema:"project name (optional)"`
}

// createWorkItemInput is the unified input structure used internally by the handler.
// The MCP SDK validates against type-specific schemas (oneOf) but unmarshals into this struct.
type createWorkItemInput = createOtherInput

// buildCreateWorkItemSchema constructs a oneOf schema for create_work_item tool.
// Returns a schema with 5 variants: Bug, Feature, User Story, Task, Other.
func buildCreateWorkItemSchema() *jsonschema.Schema {
	bugSchema, err := jsonschema.For[createBugInput](nil)
	if err != nil {
		panic(fmt.Sprintf("failed to generate Bug schema: %v", err))
	}

	bugConst := any("Bug")
	bugSchema.Properties["type"].Const = &bugConst
	bugSchema.Required = append(bugSchema.Required, "type")

	featureSchema, err := jsonschema.For[createFeatureInput](nil)
	if err != nil {
		panic(fmt.Sprintf("failed to generate Feature schema: %v", err))
	}

	featureConst := any("Feature")
	featureSchema.Properties["type"].Const = &featureConst
	featureSchema.Required = append(featureSchema.Required, "type")

	userStorySchema, err := jsonschema.For[createUserStoryInput](nil)
	if err != nil {
		panic(fmt.Sprintf("failed to generate User Story schema: %v", err))
	}

	userStoryConst := any("User Story")
	userStorySchema.Properties["type"].Const = &userStoryConst
	userStorySchema.Required = append(userStorySchema.Required, "type")

	taskSchema, err := jsonschema.For[createTaskInput](nil)
	if err != nil {
		panic(fmt.Sprintf("failed to generate Task schema: %v", err))
	}

	taskConst := any("Task")
	taskSchema.Properties["type"].Const = &taskConst
	taskSchema.Required = append(taskSchema.Required, "type")

	otherSchema, err := jsonschema.For[createOtherInput](nil)
	if err != nil {
		panic(fmt.Sprintf("failed to generate Other schema: %v", err))
	}
	// Note: otherSchema.Properties["type"] does NOT get a const value - it accepts any type
	// But we need to exclude the specific types to make oneOf work correctly
	otherSchema.Properties["type"].Not = &jsonschema.Schema{
		Enum: []any{"Bug", "Feature", "User Story", "Task"},
	}
	otherSchema.Required = append(otherSchema.Required, "type")

	return &jsonschema.Schema{
		Type: "object",
		OneOf: []*jsonschema.Schema{
			bugSchema,
			featureSchema,
			userStorySchema,
			taskSchema,
			otherSchema,
		},
	}
}

// registerCreateWorkItem registers the create_work_item tool.
func registerCreateWorkItem(srv *mcp.Server, h *tools.Handlers) {
	schema := buildCreateWorkItemSchema()

	mcp.AddTool(srv, &mcp.Tool{
		Name: "create_work_item",
		Description: "Create a new Azure DevOps work item. " +
			"Requires work item type (User Story, Bug, Task, etc.) and title. " +
			"Returns the newly created work item's ID and details. " +
			"Type-specific fields are shown based on the work item type.",
		InputSchema: schema,
	}, func(
		ctx context.Context, _ *mcp.CallToolRequest, in createWorkItemInput,
	) (*mcp.CallToolResult, *client.WorkItem, error) {
		if in.Type == "" || in.Title == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "type and title are required fields",
				}},
			}, nil, nil
		}

		opts := client.CreateOptions{
			CommonFields: client.CommonFields{
				AssignedTo:       in.AssignedTo,
				Description:      in.Description,
				IterationPath:    in.IterationPath,
				AreaPath:         in.AreaPath,
				Priority:         convertFlexInt(in.Priority),
				StoryPoints:      convertFlexFloat(in.StoryPoints),
				OriginalEstimate: convertFlexFloat(in.OriginalEstimate),
				CompletedWork:    convertFlexFloat(in.CompletedWork),
				RemainingWork:    convertFlexFloat(in.RemainingWork),
				Effort:           convertFlexFloat(in.Effort),
				Size:             in.Size,
				Severity:         in.Severity,
				Activity:         in.Activity,
				ValueArea:        in.ValueArea,
				StartDate:        convertFlexDateTime(in.StartDate),
				FinishDate:       convertFlexDateTime(in.FinishDate),
				TargetDate:       convertFlexDateTime(in.TargetDate),
				BusinessValue:    convertFlexIntToPtr(in.BusinessValue),
				StackRank:        convertFlexFloatToPtr(in.StackRank),
				Risk:             in.Risk,
				FoundIn:          in.FoundIn,
				IntegrationBuild: in.IntegrationBuild,
			},
			DateFields: &client.DateFields{
				DueDate:            convertFlexDateTime(in.DueDate),
				MarketDate:         convertFlexDateTime(in.MarketDate),
				DevCompleteDate:    convertFlexDateTime(in.DevCompleteDate),
				QCStartDate:        convertFlexDateTime(in.QCStartDate),
				QCCompleteDate:     convertFlexDateTime(in.QCCompleteDate),
				OriginalTargetDate: convertFlexDateTime(in.OriginalTargetDate),
			},
			PlanningFields: &client.PlanningFields{
				TimeCriticality: convertFlexFloatToPtr(in.TimeCriticality),
			},
			RequirementFields: &client.RequirementFields{
				FunctionalRequirements:    in.FunctionalRequirements,
				NonfunctionalRequirements: in.NonfunctionalRequirements,
				BusinessCase:              in.BusinessCase,
				SuggestedTests:            in.SuggestedTests,
			},
			QualityFields: &client.QualityFields{
				ApprovedBy:    in.ApprovedBy,
				Documentation: in.Documentation,
				RCAReason:     in.RCAReason,
			},
			MetricsFields: &client.MetricsFields{
				Confidence:          convertFlexIntToPtr(in.Confidence),
				PrioritizationScore: convertFlexIntToPtr(in.PrioritizationScore),
			},
			SecurityFields: &client.SecurityFields{
				CVENumber:           in.CVENumber,
				VulnerabilitySource: in.VulnerabilitySource,
			},
			SalesforceFields: &client.SalesforceFields{
				CaseNumber:         in.SalesforceCaseNumber,
				URL:                in.SalesforceURL,
				EscalatingEngineer: in.SalesforceEscalatingEngineer,
			},
			FeatureFields: &client.FeatureSpecificFields{
				AtRisk:         convertFlexBoolToPtr(in.AtRisk),
				DeliveryRisk:   in.DeliveryRisk,
				RiskReason:     in.RiskReason,
				MitigationPlan: in.MitigationPlan,
			},
			BugFields: &client.BugSpecificFields{
				SystemInfo:  in.SystemInfo,
				Blocked:     in.Blocked,
				ProposedFix: in.ProposedFix,
			},
			UserStoryFields: &client.UserStorySpecificFields{
				DevOwner: in.DevOwner,
				Poker:    in.Poker,
			},
			Tags: in.Tags,
		}

		workItem, text, err := h.CreateWorkItem(ctx, in.Type, in.Title, opts, in.Project)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "could not create work item: check work item type and project permissions",
				}},
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, workItem, nil
	})
}

// updateWorkItemInput is the input structure for the update_work_item tool.
type updateWorkItemInput struct {
	ID                 FlexID    `json:"id"                            jsonschema:"work item ID or reference (required)"`
	Title              string    `json:"title,omitempty"               jsonschema:"new title for the work item"`
	State              string    `json:"state,omitempty"               jsonschema:"new state"`
	AssignedTo         string    `json:"assigned_to,omitempty"         jsonschema:"email or display name to reassign to"`
	Description        string    `json:"description,omitempty"         jsonschema:"new description in plain text or HTML"`
	AcceptanceCriteria string    `json:"acceptance_criteria,omitempty" jsonschema:"acceptance criteria for work item"`
	Tags               string    `json:"tags,omitempty"                jsonschema:"semicolon-separated tags"`
	IterationPath      string    `json:"iteration_path,omitempty"      jsonschema:"iteration/sprint path"`
	AreaPath           string    `json:"area_path,omitempty"           jsonschema:"area path in the project"`
	Priority           FlexInt   `json:"priority,omitempty"            jsonschema:"priority level (1-4)"`
	StoryPoints        FlexFloat `json:"story_points,omitempty"        jsonschema:"story points estimate"`
	OriginalEstimate   FlexFloat `json:"original_estimate,omitempty"   jsonschema:"time estimate in hours"`
	CompletedWork      FlexFloat `json:"completed_work,omitempty"      jsonschema:"new completed work in hours"`
	RemainingWork      FlexFloat `json:"remaining_work,omitempty"      jsonschema:"new remaining work in hours"`
	Effort             FlexFloat `json:"effort,omitempty"              jsonschema:"effort in hours"`
	Size               string    `json:"size,omitempty"                jsonschema:"t-shirt size (S, M, L, XL)"`
	Severity           string    `json:"severity,omitempty"            jsonschema:"new severity level"`
	Activity           string    `json:"activity,omitempty"            jsonschema:"activity type"`
	ValueArea          string    `json:"value_area,omitempty"          jsonschema:"value area (Business/Architectural)"`
	Reason             string    `json:"reason,omitempty"              jsonschema:"reason for state change"`

	// Date fields (same as create)
	StartDate          *FlexDateTime `json:"start_date,omitempty"           jsonschema:"planned start date (ISO 8601)"`
	FinishDate         *FlexDateTime `json:"finish_date,omitempty"          jsonschema:"planned finish date (ISO 8601)"`
	TargetDate         *FlexDateTime `json:"target_date,omitempty"          jsonschema:"target completion date (ISO 8601)"`
	DueDate            *FlexDateTime `json:"due_date,omitempty"             jsonschema:"hard deadline (ISO 8601)"`
	MarketDate         *FlexDateTime `json:"market_date,omitempty"          jsonschema:"market release date (ISO 8601)"`
	DevCompleteDate    *FlexDateTime `json:"dev_complete_date,omitempty"    jsonschema:"dev completion date (ISO 8601)"`
	QCStartDate        *FlexDateTime `json:"qc_start_date,omitempty"        jsonschema:"QC start date (ISO 8601)"`
	QCCompleteDate     *FlexDateTime `json:"qc_complete_date,omitempty"     jsonschema:"QC completion date (ISO 8601)"`
	OriginalTargetDate *FlexDateTime `json:"original_target_date,omitempty" jsonschema:"original target date (ISO 8601)"`

	// Planning fields
	BusinessValue   FlexInt   `json:"business_value,omitempty"   jsonschema:"business value score"`
	StackRank       FlexFloat `json:"stack_rank,omitempty"       jsonschema:"backlog ordering rank"`
	Risk            string    `json:"risk,omitempty"             jsonschema:"risk level"`
	TimeCriticality FlexFloat `json:"time_criticality,omitempty" jsonschema:"time sensitivity score"`

	// Build integration
	FoundIn          string `json:"found_in,omitempty"          jsonschema:"build where bug was found"`
	IntegrationBuild string `json:"integration_build,omitempty" jsonschema:"build with fix"`

	// Feature-specific
	AtRisk         FlexBool `json:"at_risk,omitempty"         jsonschema:"is feature at risk"`
	Documentation  string   `json:"documentation,omitempty"   jsonschema:"documentation status"`
	DeliveryRisk   string   `json:"delivery_risk,omitempty"   jsonschema:"delivery risk level"`
	RiskReason     string   `json:"risk_reason,omitempty"     jsonschema:"reason for risk"`
	MitigationPlan string   `json:"mitigation_plan,omitempty" jsonschema:"risk mitigation plan (HTML)"`

	// Fields for Bug work items
	SystemInfo  string `json:"system_info,omitempty"  jsonschema:"system information (HTML)"`
	Blocked     string `json:"blocked,omitempty"      jsonschema:"is work blocked"`
	ProposedFix string `json:"proposed_fix,omitempty" jsonschema:"proposed solution (HTML)"`

	// User Story-specific
	DevOwner string `json:"dev_owner,omitempty" jsonschema:"development owner"`
	Poker    string `json:"poker,omitempty"     jsonschema:"planning poker estimate"`

	// Requirements
	FunctionalRequirements    string `json:"functional_requirements,omitempty"    jsonschema:"functional reqs (HTML)"`
	NonfunctionalRequirements string `json:"nonfunctional_requirements,omitempty" jsonschema:"non-func reqs (HTML)"`
	BusinessCase              string `json:"business_case,omitempty"              jsonschema:"business case (HTML)"`
	SuggestedTests            string `json:"suggested_tests,omitempty"            jsonschema:"suggested test cases (HTML)"`

	// Salesforce integration
	SalesforceCaseNumber         string `json:"salesforce_case_number,omitempty"         jsonschema:"Salesforce case no."`
	SalesforceURL                string `json:"salesforce_url,omitempty"                 jsonschema:"Salesforce case URL"`
	SalesforceEscalatingEngineer string `json:"salesforce_escalating_engineer,omitempty" jsonschema:"escalating engineer"`

	// Security
	CVENumber           string `json:"cve_number,omitempty"           jsonschema:"CVE identifier"`
	VulnerabilitySource string `json:"vulnerability_source,omitempty" jsonschema:"source of vulnerability"`

	// Quality
	ApprovedBy string `json:"approved_by,omitempty" jsonschema:"who approved"`
	RCAReason  string `json:"rca_reason,omitempty"  jsonschema:"root cause analysis reason"`

	// Metrics
	Confidence          FlexInt `json:"confidence,omitempty"           jsonschema:"confidence level"`
	PrioritizationScore FlexInt `json:"prioritization_score,omitempty" jsonschema:"prioritization score"`

	Project string `json:"project,omitempty" jsonschema:"project name (optional)"`
}

// registerUpdateWorkItem registers the update_work_item tool.
func registerUpdateWorkItem(srv *mcp.Server, h *tools.Handlers) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "update_work_item",
		Description: "Update fields on an existing Azure DevOps work item. " +
			"Provide the work item ID and any fields to update. " +
			"At least one field must be provided. Returns the updated work item details.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in updateWorkItemInput) (*mcp.CallToolResult, any, error) {
		id := int(in.ID)

		if id <= 0 {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "id must be a positive integer",
				}},
			}, nil, nil
		}

		opts := client.UpdateOptions{
			CommonFields: client.CommonFields{
				AssignedTo:       in.AssignedTo,
				Description:      in.Description,
				IterationPath:    in.IterationPath,
				AreaPath:         in.AreaPath,
				Priority:         convertFlexInt(in.Priority),
				StoryPoints:      convertFlexFloat(in.StoryPoints),
				OriginalEstimate: convertFlexFloat(in.OriginalEstimate),
				CompletedWork:    convertFlexFloat(in.CompletedWork),
				RemainingWork:    convertFlexFloat(in.RemainingWork),
				Effort:           convertFlexFloat(in.Effort),
				Size:             in.Size,
				Severity:         in.Severity,
				Activity:         in.Activity,
				ValueArea:        in.ValueArea,
				StartDate:        convertFlexDateTime(in.StartDate),
				FinishDate:       convertFlexDateTime(in.FinishDate),
				TargetDate:       convertFlexDateTime(in.TargetDate),
				BusinessValue:    convertFlexIntToPtr(in.BusinessValue),
				StackRank:        convertFlexFloatToPtr(in.StackRank),
				Risk:             in.Risk,
				FoundIn:          in.FoundIn,
				IntegrationBuild: in.IntegrationBuild,
			},
			DateFields: &client.DateFields{
				DueDate:            convertFlexDateTime(in.DueDate),
				MarketDate:         convertFlexDateTime(in.MarketDate),
				DevCompleteDate:    convertFlexDateTime(in.DevCompleteDate),
				QCStartDate:        convertFlexDateTime(in.QCStartDate),
				QCCompleteDate:     convertFlexDateTime(in.QCCompleteDate),
				OriginalTargetDate: convertFlexDateTime(in.OriginalTargetDate),
			},
			PlanningFields: &client.PlanningFields{
				TimeCriticality: convertFlexFloatToPtr(in.TimeCriticality),
			},
			RequirementFields: &client.RequirementFields{
				FunctionalRequirements:    in.FunctionalRequirements,
				NonfunctionalRequirements: in.NonfunctionalRequirements,
				BusinessCase:              in.BusinessCase,
				SuggestedTests:            in.SuggestedTests,
			},
			QualityFields: &client.QualityFields{
				ApprovedBy:    in.ApprovedBy,
				Documentation: in.Documentation,
				RCAReason:     in.RCAReason,
			},
			MetricsFields: &client.MetricsFields{
				Confidence:          convertFlexIntToPtr(in.Confidence),
				PrioritizationScore: convertFlexIntToPtr(in.PrioritizationScore),
			},
			SecurityFields: &client.SecurityFields{
				CVENumber:           in.CVENumber,
				VulnerabilitySource: in.VulnerabilitySource,
			},
			SalesforceFields: &client.SalesforceFields{
				CaseNumber:         in.SalesforceCaseNumber,
				URL:                in.SalesforceURL,
				EscalatingEngineer: in.SalesforceEscalatingEngineer,
			},
			FeatureFields: &client.FeatureSpecificFields{
				AtRisk:         convertFlexBoolToPtr(in.AtRisk),
				DeliveryRisk:   in.DeliveryRisk,
				RiskReason:     in.RiskReason,
				MitigationPlan: in.MitigationPlan,
			},
			BugFields: &client.BugSpecificFields{
				SystemInfo:  in.SystemInfo,
				Blocked:     in.Blocked,
				ProposedFix: in.ProposedFix,
			},
			UserStoryFields: &client.UserStorySpecificFields{
				DevOwner: in.DevOwner,
				Poker:    in.Poker,
			},
			Title:              in.Title,
			State:              in.State,
			AcceptanceCriteria: in.AcceptanceCriteria,
			Reason:             in.Reason,
			Tags:               in.Tags,
		}

		workItem, text, err := h.UpdateWorkItem(ctx, id, opts, in.Project)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf(
						"could not update work item %d: check the ID and fields", id,
					),
				}},
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, workItem, nil
	})
}

// addCommentInput is the input structure for the add_comment tool.
type addCommentInput struct {
	ID      FlexID `json:"id"                jsonschema:"work item ID or reference (required)"`
	Text    string `json:"text"              jsonschema:"comment text in plain text or HTML (required)"`
	Project string `json:"project,omitempty" jsonschema:"project name (optional, uses server default)"`
}

// registerAddComment registers the add_comment tool.
func registerAddComment(srv *mcp.Server, h *tools.Handlers) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "add_comment",
		Description: "Add a comment to an Azure DevOps work item. " +
			"Comments are visible in the work item's discussion section. Returns confirmation of the added comment.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in addCommentInput) (*mcp.CallToolResult, any, error) {
		id := int(in.ID)

		if id <= 0 {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "id must be a positive integer",
				}},
			}, nil, nil
		}

		if in.Text == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: "text parameter is required",
				}},
			}, nil, nil
		}

		text, err := h.AddComment(ctx, id, in.Text, in.Project)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf(
						"could not add comment to work item %d: check the ID and permissions", id,
					),
				}},
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})
}
