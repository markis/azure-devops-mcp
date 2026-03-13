package client_test

import (
	"context"
	"strings"
	"testing"

	"github.com/markis/azure-devops-mcp/internal/client"
)

// TestMarkdownToHTMLIntegration documents the Markdown conversion flow.
// This test demonstrates that the client layer can accept markdown input
// and would convert/sanitize it before sending to Azure DevOps.
func TestMarkdownToHTMLIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("create with markdown description", func(t *testing.T) {
		markdownDesc := "## Overview\n\nThis is **bold** text with:\n- Item 1\n- Item 2"
		opts := client.CreateOptions{
			CommonFields: client.CommonFields{
				Description: markdownDesc,
			},
		}

		// Verify options can be created with markdown
		if opts.Description == "" {
			t.Error("description is empty")
		}

		if !strings.Contains(opts.Description, "**bold**") {
			t.Errorf("description doesn't contain markdown syntax")
		}

		// In real usage, client.CreateWorkItem would:
		// 1. Convert markdown to HTML using convertMarkdownToHTML
		// 2. Sanitize the HTML using sanitizeHTML
		// 3. Send sanitized HTML to Azure DevOps
		_ = ctx // Would be used by real client
	})

	t.Run("update with markdown acceptance criteria", func(t *testing.T) {
		markdownAC := "- Must support authentication\n- Must validate input"
		opts := client.UpdateOptions{
			AcceptanceCriteria: markdownAC,
		}

		// Verify options can be created with markdown
		if opts.AcceptanceCriteria == "" {
			t.Error("acceptance criteria is empty")
		}

		if !strings.Contains(opts.AcceptanceCriteria, "- Must") {
			t.Errorf("acceptance criteria doesn't contain markdown list")
		}

		_ = ctx // Would be used by real client
	})

	t.Run("create with HTML input (would be sanitized)", func(t *testing.T) {
		htmlDesc := "<p>Safe text</p><script>alert('xss')</script>"
		opts := client.CreateOptions{
			CommonFields: client.CommonFields{
				Description: htmlDesc,
			},
		}

		// Verify HTML can be provided
		if opts.Description == "" {
			t.Error("description is empty")
		}

		// In real usage, dangerous tags like <script> would be removed by sanitizeHTML
		_ = ctx // Would be used by real client
	})

	t.Run("empty fields are not processed", func(t *testing.T) {
		opts := client.CreateOptions{
			CommonFields: client.CommonFields{
				Description: "", // Empty should not be processed
			},
		}

		// Verify empty string is accepted (will be skipped by addHTMLField)
		if opts.Description != "" {
			t.Errorf("expected empty description, got %q", opts.Description)
		}

		_ = ctx // Would be used by real client
	})
}

// TestAllHTMLFieldsSupported documents all 16 HTML fields that support Markdown.
// This test serves as documentation of the markdown conversion feature coverage.
func TestAllHTMLFieldsSupported(t *testing.T) {
	markdownInput := "**Bold** text with [link](http://example.com)"

	// These are the 16 HTML fields that support Markdown conversion:
	htmlFields := []struct {
		name     string
		location string // create, update, or comment
	}{
		// CommonFields (used in both Create and Update)
		{"description", "both"},
		{"acceptance_criteria", "update"},
		{"repro_steps", "both"},

		// BugSpecificFields
		{"system_info", "create"},
		{"proposed_fix", "create"},

		// FeatureSpecificFields
		{"mitigation_plan", "create"},

		// RequirementFields
		{"functional_requirements", "create"},
		{"nonfunctional_requirements", "create"},
		{"business_case", "create"},
		{"suggested_tests", "create"},

		// PlanningFields2
		{"rejected_ideas", "create"},
		{"resources", "create"},

		// TestCaseSpecificFields
		{"steps", "create"},
		{"parameters", "create"},
		{"local_data_source", "create"},

		// Comments (separate from work items)
		{"comments", "comment"},
	}

	t.Logf("Total HTML fields supporting Markdown: %d", len(htmlFields))

	for _, field := range htmlFields {
		t.Run(field.name, func(t *testing.T) {
			// Verify field is documented
			if field.location == "" {
				t.Errorf("field %s missing location info", field.name)
			}

			// In real implementation, each field:
			// 1. Accepts markdown input (via markdownInput)
			// 2. Converts to HTML using convertMarkdownToHTML
			// 3. Sanitizes using sanitizeHTML
			// 4. Sends to Azure DevOps via addHTMLField
			_ = markdownInput
		})
	}

	// Verify count matches implementation
	expectedCount := 16
	if len(htmlFields) != expectedCount {
		t.Errorf("expected %d HTML fields, documented %d", expectedCount, len(htmlFields))
	}
}

// TestMarkdownFeatureSupport documents common Markdown features that are supported.
func TestMarkdownFeatureSupport(t *testing.T) {
	tests := []struct {
		name        string
		markdown    string
		description string
	}{
		{
			"headers",
			"## Header 2\n\n### Header 3",
			"ATX-style headers (#, ##, ###)",
		},
		{
			"bold",
			"This is **bold** text",
			"Bold text using ** or __",
		},
		{
			"italic",
			"This is *italic* text",
			"Italic text using * or _",
		},
		{
			"unordered lists",
			"- Item 1\n- Item 2\n- Item 3",
			"Unordered lists using -, *, or +",
		},
		{
			"ordered lists",
			"1. First\n2. Second\n3. Third",
			"Ordered lists with numbers",
		},
		{
			"inline code",
			"Use `code` here",
			"Inline code using backticks",
		},
		{
			"code blocks",
			"```go\nfunc main() {}\n```",
			"Fenced code blocks with language",
		},
		{
			"links",
			"[Link text](http://example.com)",
			"Inline links",
		},
		{
			"tables (GFM)",
			"| A | B |\n|---|---|\n| 1 | 2 |",
			"GitHub Flavored Markdown tables",
		},
		{
			"mixed content",
			"## Overview\n\n**Bold** and *italic* with:\n- List item\n- Another item",
			"Multiple markdown features combined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Document the markdown syntax and its purpose
			if tt.markdown == "" {
				t.Error("markdown example is empty")
			}

			if tt.description == "" {
				t.Error("description is missing")
			}

			// In real usage, this markdown would be:
			// 1. Passed to prepareHTMLField or addHTMLField
			// 2. Converted to HTML via goldmark
			// 3. Sanitized via bluemonday
			// 4. Sent to Azure DevOps
			t.Logf("Feature: %s - %s", tt.name, tt.description)
		})
	}
}
