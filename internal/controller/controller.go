// Package controller wires the MCP server, registers all tools, and starts
// the stdio transport. It is the only package that depends on both client and tools.
package controller

import (
	"context"
	"fmt"

	"github.com/markistaylor/azure-devops-mcp/internal/client"
	"github.com/markistaylor/azure-devops-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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
	ado, err := client.NewRealADOClient(ctx, cfg.OrgURL, cfg.PAT)
	if err != nil {
		return fmt.Errorf("creating ADO client: %w", err)
	}

	h := tools.NewHandlers(ado, cfg.Project)
	srv := createServer()
	registerGetWorkItem(srv, h)
	registerListWorkItems(srv, h)
	registerListMyWorkItems(srv, h)
	registerCreateWorkItem(srv, h)
	registerUpdateWorkItem(srv, h)
	registerAddComment(srv, h)

	return srv.Run(ctx, &mcp.StdioTransport{})
}

// createServer creates and configures the MCP server instance.
func createServer() *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{
		Name:    "azure-devops-mcp",
		Version: "0.1.0",
	}, nil)
}

// registerGetWorkItem registers the get_work_item tool.
func registerGetWorkItem(srv *mcp.Server, h *tools.Handlers) {
	type getWorkItemInput struct {
		ID      int    `json:"id"`
		Project string `json:"project,omitempty"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_work_item",
		Description: "Fetch a single Azure DevOps work item by numeric ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getWorkItemInput) (*mcp.CallToolResult, any, error) {
		text, err := h.GetWorkItem(ctx, in.ID, in.Project)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})
}

// registerListWorkItems registers the list_work_items tool.
func registerListWorkItems(srv *mcp.Server, h *tools.Handlers) {
	type listWorkItemsInput struct {
		Query   string `json:"query"`
		Project string `json:"project,omitempty"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_work_items",
		Description: "Run a WIQL query and return matching Azure DevOps work items.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listWorkItemsInput) (*mcp.CallToolResult, any, error) {
		text, err := h.ListWorkItems(ctx, in.Query, in.Project)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})
}

// registerListMyWorkItems registers the list_my_work_items tool.
func registerListMyWorkItems(srv *mcp.Server, h *tools.Handlers) {
	type listMyWorkItemsInput struct {
		Project string `json:"project,omitempty"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_my_work_items",
		Description: "Return active work items assigned to the authenticated user.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listMyWorkItemsInput) (*mcp.CallToolResult, any, error) {
		text, err := h.ListMyWorkItems(ctx, in.Project)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})
}

// registerCreateWorkItem registers the create_work_item tool.
func registerCreateWorkItem(srv *mcp.Server, h *tools.Handlers) {
	type createWorkItemInput struct {
		Type             string  `json:"type"`
		Title            string  `json:"title"`
		Description      string  `json:"description,omitempty"`
		AssignedTo       string  `json:"assigned_to,omitempty"`
		Tags             string  `json:"tags,omitempty"`
		StoryPoints      float64 `json:"story_points,omitempty"`
		OriginalEstimate float64 `json:"original_estimate,omitempty"`
		Size             string  `json:"size,omitempty"`
		Project          string  `json:"project,omitempty"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_work_item",
		Description: "Create a new Azure DevOps work item.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createWorkItemInput) (*mcp.CallToolResult, any, error) {
		opts := client.CreateOptions{
			Description:      in.Description,
			AssignedTo:       in.AssignedTo,
			Tags:             in.Tags,
			StoryPoints:      in.StoryPoints,
			OriginalEstimate: in.OriginalEstimate,
			Size:             in.Size,
		}

		text, err := h.CreateWorkItem(ctx, in.Type, in.Title, opts, in.Project)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})
}

// registerUpdateWorkItem registers the update_work_item tool.
func registerUpdateWorkItem(srv *mcp.Server, h *tools.Handlers) {
	type updateWorkItemInput struct {
		ID                 int     `json:"id"`
		Title              string  `json:"title,omitempty"`
		State              string  `json:"state,omitempty"`
		AssignedTo         string  `json:"assigned_to,omitempty"`
		Description        string  `json:"description,omitempty"`
		AcceptanceCriteria string  `json:"acceptance_criteria,omitempty"`
		StoryPoints        float64 `json:"story_points,omitempty"`
		OriginalEstimate   float64 `json:"original_estimate,omitempty"`
		Size               string  `json:"size,omitempty"`
		Project            string  `json:"project,omitempty"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_work_item",
		Description: "Update fields on an existing Azure DevOps work item.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in updateWorkItemInput) (*mcp.CallToolResult, any, error) {
		opts := client.UpdateOptions{
			Title:              in.Title,
			State:              in.State,
			AssignedTo:         in.AssignedTo,
			Description:        in.Description,
			AcceptanceCriteria: in.AcceptanceCriteria,
			StoryPoints:        in.StoryPoints,
			OriginalEstimate:   in.OriginalEstimate,
			Size:               in.Size,
		}

		text, err := h.UpdateWorkItem(ctx, in.ID, opts, in.Project)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})
}

// registerAddComment registers the add_comment tool.
func registerAddComment(srv *mcp.Server, h *tools.Handlers) {
	type addCommentInput struct {
		ID      int    `json:"id"`
		Text    string `json:"text"`
		Project string `json:"project,omitempty"`
	}

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "add_comment",
		Description: "Add a comment to an Azure DevOps work item.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in addCommentInput) (*mcp.CallToolResult, any, error) {
		text, err := h.AddComment(ctx, in.ID, in.Text, in.Project)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})
}
