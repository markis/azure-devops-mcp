# Design: Add Common VSTS Fields to MCP Server

**Date:** 2026-03-09  
**Status:** Approved  
**Context:** Vulnerability work items require the Severity field, which is currently not supported by the MCP server. This design adds Severity and other commonly-used VSTS fields.

## Problem Statement

The Azure DevOps MCP server cannot create Vulnerability work items because it lacks support for the required `Microsoft.VSTS.Common.Severity` field. Additionally, the server is missing other commonly-used fields like time tracking (CompletedWork, RemainingWork) and state transition reasoning (Reason).

### Example Failure
```bash
az boards work-item create --type Vulnerability \
  --title "Security Issue" \
  --fields "Microsoft.VSTS.Common.Severity=Medium"
# ✅ Succeeds via Azure CLI

# ❌ Fails via MCP server - no severity parameter available
```

## Goals

1. Enable creation of Vulnerability work items with required Severity field
2. Add support for time tracking fields used with Task work items
3. Add support for Reason field (state transition context)
4. Keep the API simple and uniform across all work item types
5. Maintain backward compatibility

## Non-Goals

- Type-specific validation of fields (let Azure DevOps handle this)
- Custom work item type handling
- Field value validation beyond what ADO provides

## Design Decisions

### Approach: Universal Optional Fields

Add new fields as universal optional fields across all work item types. This matches the existing pattern used for Priority, StoryPoints, and OriginalEstimate.

**Rationale:**
- Simple, uniform API - users don't need type-specific knowledge
- Azure DevOps gracefully ignores fields not supported by a work item type
- Future-proof for custom work item types
- Minimal code changes required
- Consistent with existing field handling patterns

**Alternatives Considered:**
- Type-specific option structs (rejected: too complex, breaks uniform API)
- Only add Severity (rejected: incomplete, will need another PR soon)

## Implementation

### Fields to Add

| Field | ADO Path | JSON Key | Type | Used By |
|-------|----------|----------|------|---------|
| Severity | `Microsoft.VSTS.Common.Severity` | `severity` | string | Vulnerability (required), Bug (optional) |
| Completed Work | `Microsoft.VSTS.Scheduling.CompletedWork` | `completed_work` | float64 | Task |
| Remaining Work | `Microsoft.VSTS.Scheduling.RemainingWork` | `remaining_work` | float64 | Task |
| Reason | `System.Reason` | `reason` | string | All types |

**Severity Values:**
- "1 - Critical" / "Critical"
- "2 - High" / "High"
- "3 - Medium" / "Medium"
- "4 - Low" / "Low"

### Code Changes

#### 1. internal/client/client.go

**Add field path constants** (after line 42):
```go
fieldPathSeverity = "/fields/Microsoft.VSTS.Common.Severity"
fieldPathCompletedWork = "/fields/Microsoft.VSTS.Scheduling.CompletedWork"
fieldPathRemainingWork = "/fields/Microsoft.VSTS.Scheduling.RemainingWork"
fieldPathReason = "/fields/System.Reason"
```

**Update WorkItem struct** (around line 63):
```go
type WorkItem struct {
    WorkItemSummary
    Description        string
    AcceptanceCriteria string
    ReproSteps         string
    OriginalEstimate   float64
    CompletedWork      float64  `json:"completed_work,omitempty"`
    RemainingWork      float64  `json:"remaining_work,omitempty"`
    Size               string
    Severity           string   `json:"severity,omitempty"`
    Reason             string   `json:"reason,omitempty"`
    URL                string
}
```

**Update CreateOptions struct** (around line 75):
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

**Update UpdateOptions struct** (around line 84):
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

**Update GetWorkItem fields list** (around line 133):
```go
fields := []string{
    "System.Id", "System.Title", "System.State",
    "System.WorkItemType", "System.AssignedTo",
    "System.Description", "System.Tags", "System.Reason",  // Add Reason
    "System.AreaPath", "System.IterationPath", "System.Parent",
    "Microsoft.VSTS.Common.AcceptanceCriteria",
    "Microsoft.VSTS.Common.Priority",
    "Microsoft.VSTS.Common.Severity",  // Add Severity
    "Custom.Teeshirtsizing",
    "Microsoft.VSTS.Scheduling.StoryPoints",
    "Microsoft.VSTS.Scheduling.OriginalEstimate",
    "Microsoft.VSTS.Scheduling.CompletedWork",   // Add CompletedWork
    "Microsoft.VSTS.Scheduling.RemainingWork",   // Add RemainingWork
    "Microsoft.VSTS.TCM.ReproSteps",
}
```

**Update CreateWorkItem** (around line 181) - add JSON patch operations:
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

**Update buildUpdateOps** (around line 401) - add update operations:
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

**Update toWorkItem** (around line 308) - extract new fields:
```go
wi := &WorkItem{
    WorkItemSummary: WorkItemSummary{
        // ... existing fields ...
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

**Update ErrNoFieldsToUpdate** (around line 25):
```go
var ErrNoFieldsToUpdate = errors.New(
    "no fields to update: provide at least one of title, state, assigned_to, " +
    "description, acceptance_criteria, story_points, original_estimate, " +
    "completed_work, remaining_work, size, severity, or reason",
)
```

#### 2. internal/controller/controller.go

Update MCP tool definitions to expose new parameters:

**ado_create_work_item tool** - add input properties:
```go
"severity": {
    "type": "string",
    "description": "Severity level (Critical/High/Medium/Low) for Bug/Vulnerability"
}
"completed_work": {
    "type": "number",
    "description": "Completed work in hours (for Tasks)"
}
"remaining_work": {
    "type": "number",
    "description": "Remaining work in hours (for Tasks)"
}
```

**ado_update_work_item tool** - add input properties:
```go
"severity": {
    "type": "string",
    "description": "New severity level"
}
"completed_work": {
    "type": "number",
    "description": "New completed work in hours"
}
"remaining_work": {
    "type": "number",
    "description": "New remaining work in hours"
}
"reason": {
    "type": "string",
    "description": "Reason for state change"
}
```

#### 3. internal/tools/workitems.go

**Update formatWorkItem** (around line 104) - display new fields:
```go
if wi.Severity != "" {
    fmt.Fprintf(&b, " | Severity: %s", wi.Severity)
}

if wi.Reason != "" {
    fmt.Fprintf(&b, " | Reason: %s", wi.Reason)
}

// After description section:
if wi.CompletedWork > 0 || wi.RemainingWork > 0 {
    fmt.Fprintf(&b, "\nTime Tracking: Completed: %.1fh, Remaining: %.1fh\n", 
                wi.CompletedWork, wi.RemainingWork)
}
```

### Testing Strategy

#### Unit Tests

**internal/client/client_test.go:**
- Test creating work items with severity, completed_work, remaining_work
- Test updating work items with severity and reason
- Verify JSON patch operations include correct field paths and values
- Test toWorkItem() extracts new fields correctly
- Test that zero/empty values are omitted from patch operations

#### Integration Tests

Manual or automated integration testing:
- ✅ Create Vulnerability with severity="Medium"
- ✅ Create Bug with severity="High"
- ✅ Create Task with completed_work=5, remaining_work=3
- ✅ Update work item reason when changing state
- ✅ Verify fields are ignored gracefully on work item types that don't support them
- ✅ Verify existing work items without new fields still work

#### Edge Cases
- Empty/zero values omitted (existing pattern)
- Invalid severity values → ADO API error (pass through)
- Negative time values → ADO API error (pass through)

## Migration and Rollout

**Backward Compatibility:**
- ✅ All new fields are optional
- ✅ Existing API calls continue to work unchanged
- ✅ JSON responses include new fields but clients can ignore them
- ✅ No breaking changes to MCP tool signatures

**Rollout:**
1. Implement changes
2. Run full test suite (`go test ./...`)
3. Run linter (`golangci-lint run`)
4. Manual integration test with Vulnerability creation
5. Update README.md feature list
6. Release as minor version bump

## Documentation Updates

**README.md:**
- Update features list to mention new fields:
  ```
  - ✅ Full field support (severity, time tracking, story points, acceptance criteria, area/iteration paths, etc.)
  ```

**MCP Tool Schemas:**
- Automatically documented via controller.go tool definitions

**AGENTS.md:**
- No changes needed (patterns remain the same)

## Success Metrics

1. ✅ Can create Vulnerability work items with severity via MCP
2. ✅ Can create/update work items with time tracking fields
3. ✅ All existing tests pass
4. ✅ No linter errors
5. ✅ No breaking changes to existing clients

## Future Considerations

- Add WorkItemSummary.Severity for list views (currently only on full WorkItem)
- Consider adding Activity field for Bug work items
- Monitor for other commonly-requested fields based on user feedback
