// Package controller wires the MCP server, registers all tools, and starts
// the stdio transport. It is the only package that depends on both client and tools.
package controller

import (
	"context"
	"fmt"

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

// createWorkItemInput is the input structure for the create_work_item tool.
type createWorkItemInput struct {
	Type             string    `json:"type"                        jsonschema:"work item type (required)"`
	Title            string    `json:"title"                       jsonschema:"work item title (required)"`
	Description      string    `json:"description,omitempty"       jsonschema:"detailed description in plain text or HTML"`
	AssignedTo       string    `json:"assigned_to,omitempty"       jsonschema:"email or display name to assign to"`
	Tags             string    `json:"tags,omitempty"              jsonschema:"semicolon-separated tags"`
	StoryPoints      FlexFloat `json:"story_points,omitempty"      jsonschema:"story points estimate (for User Stories)"`
	OriginalEstimate FlexFloat `json:"original_estimate,omitempty" jsonschema:"time estimate in hours (for Tasks)"`
	CompletedWork    FlexFloat `json:"completed_work,omitempty"    jsonschema:"completed work in hours (for Tasks)"`
	RemainingWork    FlexFloat `json:"remaining_work,omitempty"    jsonschema:"remaining work in hours (for Tasks)"`
	Size             string    `json:"size,omitempty"              jsonschema:"t-shirt size (S, M, L, XL)"`
	Severity         string    `json:"severity,omitempty"          jsonschema:"severity for Bug/Vulnerability"`
	Project          string    `json:"project,omitempty"           jsonschema:"project name (optional)"`
}

// registerCreateWorkItem registers the create_work_item tool.
func registerCreateWorkItem(srv *mcp.Server, h *tools.Handlers) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "create_work_item",
		Description: "Create a new Azure DevOps work item. " +
			"Requires work item type (User Story, Bug, Task, etc.) and title. " +
			"Returns the newly created work item's ID and details.",
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
				StoryPoints:      float64(in.StoryPoints),
				OriginalEstimate: float64(in.OriginalEstimate),
				CompletedWork:    float64(in.CompletedWork),
				RemainingWork:    float64(in.RemainingWork),
				Size:             in.Size,
				Severity:         in.Severity,
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
	StoryPoints        FlexFloat `json:"story_points,omitempty"        jsonschema:"story points estimate"`
	OriginalEstimate   FlexFloat `json:"original_estimate,omitempty"   jsonschema:"time estimate in hours"`
	CompletedWork      FlexFloat `json:"completed_work,omitempty"      jsonschema:"new completed work in hours"`
	RemainingWork      FlexFloat `json:"remaining_work,omitempty"      jsonschema:"new remaining work in hours"`
	Size               string    `json:"size,omitempty"                jsonschema:"t-shirt size (S, M, L, XL)"`
	Severity           string    `json:"severity,omitempty"            jsonschema:"new severity level"`
	Reason             string    `json:"reason,omitempty"              jsonschema:"reason for state change"`
	Project            string    `json:"project,omitempty"             jsonschema:"project name (optional)"`
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
		storyPoints := float64(in.StoryPoints)
		originalEstimate := float64(in.OriginalEstimate)

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
				StoryPoints:      storyPoints,
				OriginalEstimate: originalEstimate,
				CompletedWork:    float64(in.CompletedWork),
				RemainingWork:    float64(in.RemainingWork),
				Size:             in.Size,
				Severity:         in.Severity,
			},
			Title:              in.Title,
			State:              in.State,
			AcceptanceCriteria: in.AcceptanceCriteria,
			Reason:             in.Reason,
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
