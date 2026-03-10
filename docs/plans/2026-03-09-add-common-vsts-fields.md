# Add Common VSTS Fields Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add support for Severity, CompletedWork, RemainingWork, and Reason fields to enable Vulnerability work item creation and improve Task time tracking.

**Architecture:** Add four new optional fields as universal fields across all work item types, following the existing pattern used for Priority, StoryPoints, and OriginalEstimate. Let Azure DevOps handle field validation per work item type.

**Tech Stack:** Go 1.25, Azure DevOps REST API v7, MCP SDK v1.4.0

---

## Task 1: Add Field Path Constants

**Files:**
- Modify: `internal/client/client.go:30-42`

**Step 1: Add field path constants**

Add these constants after line 42 (after `fieldPathSize`):

```go
	fieldPathSeverity       = "/fields/Microsoft.VSTS.Common.Severity"
	fieldPathCompletedWork  = "/fields/Microsoft.VSTS.Scheduling.CompletedWork"
	fieldPathRemainingWork  = "/fields/Microsoft.VSTS.Scheduling.RemainingWork"
	fieldPathReason         = "/fields/System.Reason"
```

**Step 2: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/client/client.go
git commit -m "feat(client): add field path constants for new VSTS fields"
```

---

## Task 2: Update WorkItem Struct

**Files:**
- Modify: `internal/client/client.go:60-72`

**Step 1: Add fields to WorkItem struct**

Update the WorkItem struct (around line 63) to add the new fields:

```go
type WorkItem struct {
	WorkItemSummary

	Description        string  `json:"description"                   jsonschema:"Work item description"`
	AcceptanceCriteria string  `json:"acceptance_criteria,omitempty" jsonschema:"Acceptance criteria"`
	ReproSteps         string  `json:"repro_steps,omitempty"         jsonschema:"Reproduction steps"`
	OriginalEstimate   float64 `json:"original_estimate,omitempty"   jsonschema:"Time estimate in hours"`
	CompletedWork      float64 `json:"completed_work,omitempty"      jsonschema:"Completed work in hours"`
	RemainingWork      float64 `json:"remaining_work,omitempty"      jsonschema:"Remaining work in hours"`
	Size               string  `json:"size,omitempty"                jsonschema:"T-shirt size estimate"`
	Severity           string  `json:"severity,omitempty"            jsonschema:"Severity level (Critical/High/Medium/Low)"`
	Reason             string  `json:"reason,omitempty"              jsonschema:"Reason for current state"`
	URL                string  `json:"url"                           jsonschema:"Work item URL"`
}
```

**Step 2: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/client/client.go
git commit -m "feat(client): add Severity, CompletedWork, RemainingWork, Reason to WorkItem"
```

---

## Task 3: Update CreateOptions Struct

**Files:**
- Modify: `internal/client/client.go:74-82`

**Step 1: Add fields to CreateOptions struct**

Update the CreateOptions struct (around line 74):

```go
type CreateOptions struct {
	Description      string
	AssignedTo       string
	Tags             string
	StoryPoints      float64
	OriginalEstimate float64
	CompletedWork    float64
	RemainingWork    float64
	Size             string
	Severity         string
}
```

**Step 2: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/client/client.go
git commit -m "feat(client): add new fields to CreateOptions"
```

---

## Task 4: Update UpdateOptions Struct

**Files:**
- Modify: `internal/client/client.go:84-95`

**Step 1: Add fields to UpdateOptions struct**

Update the UpdateOptions struct (around line 84):

```go
type UpdateOptions struct {
	Title              string
	State              string
	AssignedTo         string
	Description        string
	AcceptanceCriteria string
	StoryPoints        float64
	OriginalEstimate   float64
	CompletedWork      float64
	RemainingWork      float64
	Size               string
	Severity           string
	Reason             string
}
```

**Step 2: Update ErrNoFieldsToUpdate message**

Find ErrNoFieldsToUpdate (around line 25) and update it:

```go
var ErrNoFieldsToUpdate = errors.New(
	"no fields to update: provide at least one of title, state, assigned_to, " +
		"description, acceptance_criteria, story_points, original_estimate, " +
		"completed_work, remaining_work, size, severity, or reason",
)
```

**Step 3: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/client/client.go
git commit -m "feat(client): add new fields to UpdateOptions and update error message"
```

---

## Task 5: Update GetWorkItem to Fetch New Fields

**Files:**
- Modify: `internal/client/client.go:132-144`

**Step 1: Add new fields to GetWorkItem fields list**

Update the fields slice in GetWorkItem (around line 133):

```go
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
```

**Step 2: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/client/client.go
git commit -m "feat(client): fetch new fields in GetWorkItem"
```

---

## Task 6: Update CreateWorkItem to Set New Fields

**Files:**
- Modify: `internal/client/client.go:180-231`

**Step 1: Add JSON patch operations for new fields**

After the Size field handling (around line 218), add:

```go
	if opts.Severity != "" {
		ops = append(ops, webapi.JsonPatchOperation{
			Op: &add, Path: &fieldPathSeverity, Value: opts.Severity,
		})
	}

	if opts.CompletedWork != 0 {
		ops = append(ops, webapi.JsonPatchOperation{
			Op: &add, Path: &fieldPathCompletedWork, Value: opts.CompletedWork,
		})
	}

	if opts.RemainingWork != 0 {
		ops = append(ops, webapi.JsonPatchOperation{
			Op: &add, Path: &fieldPathRemainingWork, Value: opts.RemainingWork,
		})
	}
```

**Step 2: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/client/client.go
git commit -m "feat(client): support new fields in CreateWorkItem"
```

---

## Task 7: Update buildUpdateOps to Handle New Fields

**Files:**
- Modify: `internal/client/client.go:400-456`

**Step 1: Add update operations for new fields**

After the Size field handling (around line 451), add:

```go
	if opts.Severity != "" {
		ops = append(ops, webapi.JsonPatchOperation{
			Op: &replace, Path: &fieldPathSeverity, Value: opts.Severity,
		})
	}

	if opts.CompletedWork != 0 {
		ops = append(ops, webapi.JsonPatchOperation{
			Op: &replace, Path: &fieldPathCompletedWork, Value: opts.CompletedWork,
		})
	}

	if opts.RemainingWork != 0 {
		ops = append(ops, webapi.JsonPatchOperation{
			Op: &replace, Path: &fieldPathRemainingWork, Value: opts.RemainingWork,
		})
	}

	if opts.Reason != "" {
		ops = append(ops, webapi.JsonPatchOperation{
			Op: &replace, Path: &fieldPathReason, Value: opts.Reason,
		})
	}
```

**Step 2: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/client/client.go
git commit -m "feat(client): support new fields in UpdateWorkItem"
```

---

## Task 8: Update toWorkItem to Extract New Fields

**Files:**
- Modify: `internal/client/client.go:307-343`

**Step 1: Extract new fields from ADO response**

Update the WorkItem initialization in toWorkItem (around line 315):

```go
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
```

**Step 2: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/client/client.go
git commit -m "feat(client): extract new fields in toWorkItem"
```

---

## Task 9: Update formatWorkItem to Display New Fields

**Files:**
- Modify: `internal/tools/workitems.go:102-133`

**Step 1: Add severity and reason to header line**

Update formatWorkItem (around line 108) to display Severity and Reason:

```go
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
```

**Step 2: Add time tracking section**

After the Tags section (around line 129), add:

```go
	if wi.CompletedWork > 0 || wi.RemainingWork > 0 {
		fmt.Fprintf(&b, "\nTime Tracking: Completed: %.1fh, Remaining: %.1fh\n",
			wi.CompletedWork, wi.RemainingWork)
	}
```

**Step 3: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/tools/workitems.go
git commit -m "feat(tools): display new fields in formatWorkItem"
```

---

## Task 10: Write Unit Tests for CreateWorkItem with New Fields

**Files:**
- Modify: `internal/client/client_test.go`

**Step 1: Write test for creating work item with severity**

Add test case to client_test.go:

```go
func TestCreateWorkItem_WithSeverity(t *testing.T) {
	mockWIT := &mockWITClient{
		CreateWorkItemFn: func(_ context.Context, args workitemtracking.CreateWorkItemArgs) (*workitemtracking.WorkItem, error) {
			// Verify severity field is in patch operations
			ops := *args.Document
			var foundSeverity bool
			for _, op := range ops {
				if *op.Path == "/fields/Microsoft.VSTS.Common.Severity" {
					foundSeverity = true
					if op.Value != "High" {
						t.Errorf("expected severity High, got %v", op.Value)
					}
				}
			}
			if !foundSeverity {
				t.Error("severity field not found in patch operations")
			}

			id := 100
			title := "Test Vulnerability"
			wiType := "Vulnerability"
			state := "New"
			fields := map[string]any{
				"System.Id":                     float64(100),
				"System.Title":                  title,
				"System.WorkItemType":           wiType,
				"System.State":                  state,
				"Microsoft.VSTS.Common.Severity": "High",
			}

			return &workitemtracking.WorkItem{
				Id:     &id,
				Fields: &fields,
			}, nil
		},
	}

	c := NewClientWithWIT(mockWIT)
	opts := CreateOptions{Severity: "High"}

	wi, err := c.CreateWorkItem(context.Background(), "TestProject", "Vulnerability", "Test Vulnerability", opts)
	if err != nil {
		t.Fatalf("CreateWorkItem failed: %v", err)
	}

	if wi.Severity != "High" {
		t.Errorf("expected severity High, got %s", wi.Severity)
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./internal/client/... -run TestCreateWorkItem_WithSeverity -v`
Expected: PASS

**Step 3: Write test for creating work item with time tracking**

Add test case:

```go
func TestCreateWorkItem_WithTimeTracking(t *testing.T) {
	mockWIT := &mockWITClient{
		CreateWorkItemFn: func(_ context.Context, args workitemtracking.CreateWorkItemArgs) (*workitemtracking.WorkItem, error) {
			ops := *args.Document
			var foundCompleted, foundRemaining bool
			for _, op := range ops {
				if *op.Path == "/fields/Microsoft.VSTS.Scheduling.CompletedWork" {
					foundCompleted = true
					if op.Value != 5.0 {
						t.Errorf("expected completed_work 5.0, got %v", op.Value)
					}
				}
				if *op.Path == "/fields/Microsoft.VSTS.Scheduling.RemainingWork" {
					foundRemaining = true
					if op.Value != 3.0 {
						t.Errorf("expected remaining_work 3.0, got %v", op.Value)
					}
				}
			}
			if !foundCompleted {
				t.Error("completed_work field not found in patch operations")
			}
			if !foundRemaining {
				t.Error("remaining_work field not found in patch operations")
			}

			id := 101
			title := "Test Task"
			wiType := "Task"
			state := "Active"
			fields := map[string]any{
				"System.Id":                                 float64(101),
				"System.Title":                              title,
				"System.WorkItemType":                       wiType,
				"System.State":                              state,
				"Microsoft.VSTS.Scheduling.CompletedWork":   5.0,
				"Microsoft.VSTS.Scheduling.RemainingWork":   3.0,
			}

			return &workitemtracking.WorkItem{
				Id:     &id,
				Fields: &fields,
			}, nil
		},
	}

	c := NewClientWithWIT(mockWIT)
	opts := CreateOptions{
		CompletedWork: 5.0,
		RemainingWork: 3.0,
	}

	wi, err := c.CreateWorkItem(context.Background(), "TestProject", "Task", "Test Task", opts)
	if err != nil {
		t.Fatalf("CreateWorkItem failed: %v", err)
	}

	if wi.CompletedWork != 5.0 {
		t.Errorf("expected completed_work 5.0, got %.1f", wi.CompletedWork)
	}
	if wi.RemainingWork != 3.0 {
		t.Errorf("expected remaining_work 3.0, got %.1f", wi.RemainingWork)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/client/... -run TestCreateWorkItem_WithTimeTracking -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/client/client_test.go
git commit -m "test(client): add tests for CreateWorkItem with new fields"
```

---

## Task 11: Write Unit Tests for UpdateWorkItem with New Fields

**Files:**
- Modify: `internal/client/client_test.go`

**Step 1: Write test for updating work item with severity and reason**

Add test case:

```go
func TestUpdateWorkItem_WithSeverityAndReason(t *testing.T) {
	mockWIT := &mockWITClient{
		UpdateWorkItemFn: func(_ context.Context, args workitemtracking.UpdateWorkItemArgs) (*workitemtracking.WorkItem, error) {
			ops := *args.Document
			var foundSeverity, foundReason bool
			for _, op := range ops {
				if *op.Path == "/fields/Microsoft.VSTS.Common.Severity" {
					foundSeverity = true
					if op.Value != "Critical" {
						t.Errorf("expected severity Critical, got %v", op.Value)
					}
				}
				if *op.Path == "/fields/System.Reason" {
					foundReason = true
					if op.Value != "Security incident" {
						t.Errorf("expected reason 'Security incident', got %v", op.Value)
					}
				}
			}
			if !foundSeverity {
				t.Error("severity field not found in patch operations")
			}
			if !foundReason {
				t.Error("reason field not found in patch operations")
			}

			id := *args.Id
			title := "Updated Bug"
			wiType := "Bug"
			state := "Active"
			fields := map[string]any{
				"System.Id":                     float64(id),
				"System.Title":                  title,
				"System.WorkItemType":           wiType,
				"System.State":                  state,
				"Microsoft.VSTS.Common.Severity": "Critical",
				"System.Reason":                 "Security incident",
			}

			return &workitemtracking.WorkItem{
				Id:     args.Id,
				Fields: &fields,
			}, nil
		},
	}

	c := NewClientWithWIT(mockWIT)
	opts := UpdateOptions{
		Severity: "Critical",
		Reason:   "Security incident",
	}

	wi, err := c.UpdateWorkItem(context.Background(), "TestProject", 200, opts)
	if err != nil {
		t.Fatalf("UpdateWorkItem failed: %v", err)
	}

	if wi.Severity != "Critical" {
		t.Errorf("expected severity Critical, got %s", wi.Severity)
	}
	if wi.Reason != "Security incident" {
		t.Errorf("expected reason 'Security incident', got %s", wi.Reason)
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./internal/client/... -run TestUpdateWorkItem_WithSeverityAndReason -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/client/client_test.go
git commit -m "test(client): add tests for UpdateWorkItem with new fields"
```

---

## Task 12: Run Full Test Suite

**Files:**
- N/A

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass

**Step 2: Run linter**

Run: `golangci-lint run`
Expected: No errors

**Step 3: Format code**

Run: `gofumpt -l -w .`
Expected: No output (code already formatted)

**Step 4: Build project**

Run: `go build ./...`
Expected: No errors

**Step 5: Verify no uncommitted changes**

Run: `git status`
Expected: Working tree clean (or only formatting changes if any)

---

## Task 13: Update MCP Tool Schemas (Controller)

**Files:**
- Modify: `internal/controller/controller.go`

**Step 1: Locate ado_create_work_item tool definition**

Find the tool definition for `ado_create_work_item` in the controller.

**Step 2: Add new input properties to schema**

Add these properties to the tool's input schema (after the existing optional properties):

```go
"severity": map[string]any{
	"type":        "string",
	"description": "severity level (Critical/High/Medium/Low) for Bug/Vulnerability work items",
},
"completed_work": map[string]any{
	"type":        "number",
	"description": "completed work in hours (for Tasks)",
},
"remaining_work": map[string]any{
	"type":        "number",
	"description": "remaining work in hours (for Tasks)",
},
```

**Step 3: Update the handler to extract new fields**

Find where CreateOptions is constructed and add the new fields:

```go
opts := client.CreateOptions{
	Description:      getString(args, "description"),
	AssignedTo:       getString(args, "assigned_to"),
	Tags:             getString(args, "tags"),
	StoryPoints:      getFloat(args, "story_points"),
	OriginalEstimate: getFloat(args, "original_estimate"),
	CompletedWork:    getFloat(args, "completed_work"),
	RemainingWork:    getFloat(args, "remaining_work"),
	Size:             getString(args, "size"),
	Severity:         getString(args, "severity"),
}
```

**Step 4: Locate ado_update_work_item tool definition**

Find the tool definition for `ado_update_work_item`.

**Step 5: Add new input properties to update schema**

Add these properties:

```go
"severity": map[string]any{
	"type":        "string",
	"description": "new severity level",
},
"completed_work": map[string]any{
	"type":        "number",
	"description": "new completed work in hours",
},
"remaining_work": map[string]any{
	"type":        "number",
	"description": "new remaining work in hours",
},
"reason": map[string]any{
	"type":        "string",
	"description": "reason for state change",
},
```

**Step 6: Update the handler to extract new fields**

Find where UpdateOptions is constructed and add the new fields:

```go
opts := client.UpdateOptions{
	Title:              getString(args, "title"),
	State:              getString(args, "state"),
	AssignedTo:         getString(args, "assigned_to"),
	Description:        getString(args, "description"),
	AcceptanceCriteria: getString(args, "acceptance_criteria"),
	StoryPoints:        getFloat(args, "story_points"),
	OriginalEstimate:   getFloat(args, "original_estimate"),
	CompletedWork:      getFloat(args, "completed_work"),
	RemainingWork:      getFloat(args, "remaining_work"),
	Size:               getString(args, "size"),
	Severity:           getString(args, "severity"),
	Reason:             getString(args, "reason"),
}
```

**Step 7: Verify code compiles**

Run: `go build ./...`
Expected: No errors

**Step 8: Commit**

```bash
git add internal/controller/controller.go
git commit -m "feat(controller): expose new fields in MCP tool schemas"
```

---

## Task 14: Update README Documentation

**Files:**
- Modify: `README.md`

**Step 1: Update features list**

Find the features list (around line 11) and update:

```markdown
## Features

- ✅ Get work items by ID
- ✅ List work items with WIQL queries
- ✅ List work items assigned to authenticated user
- ✅ Create new work items
- ✅ Update existing work items
- ✅ Add comments to work items
- ✅ Full field support (severity, time tracking, story points, acceptance criteria, area/iteration paths, etc.)
- ✅ HTML to Markdown conversion for descriptions
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: update README with new field support"
```

---

## Task 15: Final Verification and Integration Test

**Files:**
- N/A

**Step 1: Build the binary**

Run: `go build -o ./bin/azure-devops-mcp ./cmd/azure-devops-mcp/...`
Expected: Binary created successfully

**Step 2: Set up test environment variables**

```bash
export AZURE_DEVOPS_ORG_URL="https://dev.azure.com/netwrix-dev"
export AZURE_DEVOPS_PAT="your-pat"
export AZURE_DEVOPS_PROJECT="AccessAnalyzer"
```

**Step 3: Test creating a Vulnerability with severity (manual)**

Use the MCP server to create a Vulnerability work item with severity.

**Step 4: Verify the work item was created correctly**

Check Azure DevOps to confirm the Vulnerability has the correct severity.

**Step 5: Run final verification**

Run all verification commands:
```bash
go test ./... -v
golangci-lint run
go build ./...
```

Expected: All pass

**Step 6: Review all commits**

Run: `git log --oneline`
Expected: See all incremental commits with clear messages

---

## Success Criteria

- ✅ All unit tests pass
- ✅ Linter passes with no errors
- ✅ Can create Vulnerability work items with severity via MCP
- ✅ Can create/update work items with time tracking fields
- ✅ Can update work items with reason field
- ✅ Documentation updated
- ✅ No breaking changes to existing API

## Notes

- Each task commits incrementally - keep commits small and focused
- Follow TDD where applicable (tests before implementation)
- Use existing patterns from AGENTS.md for code style
- All new fields are optional - Azure DevOps handles validation per work item type
