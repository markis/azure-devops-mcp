package client

import (
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

func TestFieldString_ValidValue(t *testing.T) {
	fields := map[string]any{
		"System.Title": "Test Work Item",
	}

	result := fieldString(&fields, "System.Title")
	if result != "Test Work Item" {
		t.Errorf("expected 'Test Work Item', got %q", result)
	}
}

func TestFieldString_MissingKey(t *testing.T) {
	fields := map[string]any{}

	result := fieldString(&fields, "System.Title")
	if result != "" {
		t.Errorf("expected empty string for missing key, got %q", result)
	}
}

func TestFieldString_NilValue(t *testing.T) {
	fields := map[string]any{
		"System.Title": nil,
	}

	result := fieldString(&fields, "System.Title")
	if result != "" {
		t.Errorf("expected empty string for nil value, got %q", result)
	}
}

func TestFieldInt_Float64Value(t *testing.T) {
	fields := map[string]any{
		"System.Id": float64(123),
	}

	result := fieldInt(&fields, "System.Id")
	if result != 123 {
		t.Errorf("expected 123, got %d", result)
	}
}

func TestFieldInt_IntValue(t *testing.T) {
	fields := map[string]any{
		"System.Id": 456,
	}

	result := fieldInt(&fields, "System.Id")
	if result != 456 {
		t.Errorf("expected 456, got %d", result)
	}
}

func TestFieldInt_MissingKey(t *testing.T) {
	fields := map[string]any{}

	result := fieldInt(&fields, "System.Id")
	if result != 0 {
		t.Errorf("expected 0 for missing key, got %d", result)
	}
}

func TestFieldInt_NilValue(t *testing.T) {
	fields := map[string]any{
		"System.Id": nil,
	}

	result := fieldInt(&fields, "System.Id")
	if result != 0 {
		t.Errorf("expected 0 for nil value, got %d", result)
	}
}

func TestFieldInt_InvalidType(t *testing.T) {
	fields := map[string]any{
		"System.Id": "not a number",
	}

	result := fieldInt(&fields, "System.Id")
	if result != 0 {
		t.Errorf("expected 0 for invalid type, got %d", result)
	}
}

func TestFieldFloat_ValidValue(t *testing.T) {
	fields := map[string]any{
		"StoryPoints": float64(5.5),
	}

	result := fieldFloat(&fields, "StoryPoints")
	if result != 5.5 {
		t.Errorf("expected 5.5, got %f", result)
	}
}

func TestFieldFloat_MissingKey(t *testing.T) {
	fields := map[string]any{}

	result := fieldFloat(&fields, "StoryPoints")
	if result != 0 {
		t.Errorf("expected 0 for missing key, got %f", result)
	}
}

func TestFieldFloat_NilValue(t *testing.T) {
	fields := map[string]any{
		"StoryPoints": nil,
	}

	result := fieldFloat(&fields, "StoryPoints")
	if result != 0 {
		t.Errorf("expected 0 for nil value, got %f", result)
	}
}

func TestExtractParentID_ValidParent(t *testing.T) {
	fields := map[string]any{
		"System.Parent": float64(789),
	}

	result := extractParentID(&fields)
	if result != 789 {
		t.Errorf("expected 789, got %d", result)
	}
}

func TestExtractParentID_NoParent(t *testing.T) {
	fields := map[string]any{}

	result := extractParentID(&fields)
	if result != 0 {
		t.Errorf("expected 0 for missing parent, got %d", result)
	}
}

func TestExtractAssignedTo_ValidIdentityRef(t *testing.T) {
	fields := map[string]any{
		"System.AssignedTo": map[string]any{
			"displayName": "John Doe",
			"uniqueName":  "john@example.com",
		},
	}

	result := extractAssignedTo(&fields)
	if result != "John Doe" {
		t.Errorf("expected 'John Doe', got %q", result)
	}
}

func TestExtractAssignedTo_MissingField(t *testing.T) {
	fields := map[string]any{}

	result := extractAssignedTo(&fields)
	if result != "" {
		t.Errorf("expected empty string for missing field, got %q", result)
	}
}

func TestExtractAssignedTo_NilValue(t *testing.T) {
	fields := map[string]any{
		"System.AssignedTo": nil,
	}

	result := extractAssignedTo(&fields)
	if result != "" {
		t.Errorf("expected empty string for nil value, got %q", result)
	}
}

func TestExtractAssignedTo_InvalidType(t *testing.T) {
	fields := map[string]any{
		"System.AssignedTo": "not a map",
	}

	result := extractAssignedTo(&fields)
	if result != "" {
		t.Errorf("expected empty string for invalid type, got %q", result)
	}
}

func TestConvertToMarkdown_EmptyString(t *testing.T) {
	result := convertToMarkdown("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestConvertToMarkdown_SimpleHTML(t *testing.T) {
	html := "<p>Hello <strong>World</strong></p>"
	result := convertToMarkdown(html)

	if result == "" {
		t.Errorf("expected non-empty result for valid HTML")
	}

	if len(result) < 5 {
		t.Errorf("expected meaningful conversion, got %q", result)
	}
}

func TestBuildUpdateOps_AllFields(t *testing.T) {
	opts := UpdateOptions{
		CommonFields: CommonFields{
			AssignedTo:       "user@example.com",
			Description:      "New description",
			StoryPoints:      5.0,
			OriginalEstimate: 8.0,
			Size:             "L",
		},
		Title:              "Updated Title",
		State:              "Active",
		AcceptanceCriteria: "AC updated",
	}

	ops := buildUpdateOps(opts)

	if len(ops) != 8 {
		t.Fatalf("expected 8 operations, got %d", len(ops))
	}

	for i, op := range ops {
		if op.Op == nil {
			t.Errorf("operation %d has nil Op", i)
			continue
		}

		if *op.Op != webapi.OperationValues.Replace {
			t.Errorf("operation %d should use Replace, got %s", i, *op.Op)
		}
	}
}

func TestBuildUpdateOps_EmptyOptions(t *testing.T) {
	opts := UpdateOptions{}
	ops := buildUpdateOps(opts)

	if len(ops) != 0 {
		t.Fatalf("expected 0 operations for empty options, got %d", len(ops))
	}
}

func TestBuildUpdateOps_PartialFields(t *testing.T) {
	opts := UpdateOptions{
		Title: "Only Title",
		State: "Done",
	}

	ops := buildUpdateOps(opts)

	if len(ops) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(ops))
	}
}

func TestToWorkItem_BasicFields(t *testing.T) {
	id := 123
	fields := map[string]any{
		"System.Title":        "Test Item",
		"System.State":        "Active",
		"System.WorkItemType": "Bug",
		"System.Id":           float64(123),
	}

	item := &workitemtracking.WorkItem{
		Id:     &id,
		Fields: &fields,
	}

	wi := toWorkItem(item)

	if wi.ID != 123 {
		t.Errorf("expected ID 123, got %d", wi.ID)
	}

	if wi.Title != "Test Item" {
		t.Errorf("expected title 'Test Item', got %q", wi.Title)
	}

	if wi.State != "Active" {
		t.Errorf("expected state 'Active', got %q", wi.State)
	}

	if wi.Type != "Bug" {
		t.Errorf("expected type 'Bug', got %q", wi.Type)
	}
}

func TestToWorkItem_AllFields(t *testing.T) {
	id := 456
	fields := map[string]any{
		"System.Title":                               "Full Item",
		"System.State":                               "Active",
		"System.WorkItemType":                        "User Story",
		"System.Description":                         "<p>HTML Description</p>",
		"Microsoft.VSTS.Common.AcceptanceCriteria":   "<p>HTML AC</p>",
		"System.Tags":                                "tag1;tag2",
		"System.AreaPath":                            "Area\\Path",
		"System.IterationPath":                       "Iteration\\Path",
		"Microsoft.VSTS.Common.Priority":             float64(1),
		"Microsoft.VSTS.TCM.ReproSteps":              "<p>Repro steps</p>",
		"System.Parent":                              float64(789),
		"Microsoft.VSTS.Scheduling.StoryPoints":      float64(5.0),
		"Microsoft.VSTS.Scheduling.OriginalEstimate": float64(8.0),
		"Custom.Teeshirtsizing":                      "M",
	}

	item := &workitemtracking.WorkItem{
		Id:     &id,
		Fields: &fields,
	}

	wi := toWorkItem(item)

	if wi.ID != 456 {
		t.Errorf("expected ID 456, got %d", wi.ID)
	}

	if wi.Priority != 1 {
		t.Errorf("expected Priority 1, got %d", wi.Priority)
	}

	if wi.ParentID != 789 {
		t.Errorf("expected ParentID 789, got %d", wi.ParentID)
	}

	if wi.StoryPoints != 5.0 {
		t.Errorf("expected StoryPoints 5.0, got %f", wi.StoryPoints)
	}

	if wi.OriginalEstimate != 8.0 {
		t.Errorf("expected OriginalEstimate 8.0, got %f", wi.OriginalEstimate)
	}

	if wi.Size != "M" {
		t.Errorf("expected Size 'M', got %q", wi.Size)
	}
}

const (
	testPathTitle       = "/fields/System.Title"
	testPathStoryPoints = "/fields/Microsoft.VSTS.Scheduling.StoryPoints"
)

func TestAddStringField_WithValue(t *testing.T) {
	add := webapi.OperationValues.Add
	path := testPathTitle

	var ops []webapi.JsonPatchOperation

	addStringField(&ops, &add, &path, "Test Value")

	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}

	if *ops[0].Op != add {
		t.Errorf("expected Add operation, got %v", *ops[0].Op)
	}

	if *ops[0].Path != path {
		t.Errorf("expected path %s, got %s", path, *ops[0].Path)
	}

	if ops[0].Value != "Test Value" {
		t.Errorf("expected value 'Test Value', got %v", ops[0].Value)
	}
}

func TestAddStringField_EmptyValue(t *testing.T) {
	add := webapi.OperationValues.Add
	path := testPathTitle

	var ops []webapi.JsonPatchOperation

	addStringField(&ops, &add, &path, "")

	if len(ops) != 0 {
		t.Errorf("expected 0 operations for empty string, got %d", len(ops))
	}
}

func TestAddFloatField_WithNonZeroValue(t *testing.T) {
	add := webapi.OperationValues.Add
	path := testPathStoryPoints

	var ops []webapi.JsonPatchOperation

	addFloatField(&ops, &add, &path, 5.5)

	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}

	if *ops[0].Op != add {
		t.Errorf("expected Add operation, got %v", *ops[0].Op)
	}

	if *ops[0].Path != path {
		t.Errorf("expected path %s, got %s", path, *ops[0].Path)
	}

	if ops[0].Value != 5.5 {
		t.Errorf("expected value 5.5, got %v", ops[0].Value)
	}
}

func TestAddFloatField_ZeroValue(t *testing.T) {
	add := webapi.OperationValues.Add
	path := testPathStoryPoints

	var ops []webapi.JsonPatchOperation

	addFloatField(&ops, &add, &path, 0)

	if len(ops) != 0 {
		t.Errorf("expected 0 operations for zero value, got %d", len(ops))
	}
}

func TestAddStringField_MultipleFields(t *testing.T) {
	add := webapi.OperationValues.Add
	pathTitle := testPathTitle
	pathDesc := "/fields/System.Description"
	pathTags := "/fields/System.Tags"

	var ops []webapi.JsonPatchOperation

	addStringField(&ops, &add, &pathTitle, "Title")
	addStringField(&ops, &add, &pathDesc, "")
	addStringField(&ops, &add, &pathTags, "tag1;tag2")

	if len(ops) != 2 {
		t.Fatalf("expected 2 operations (empty desc skipped), got %d", len(ops))
	}

	if ops[0].Value != "Title" {
		t.Errorf("expected first value 'Title', got %v", ops[0].Value)
	}

	if ops[1].Value != "tag1;tag2" {
		t.Errorf("expected second value 'tag1;tag2', got %v", ops[1].Value)
	}
}

func TestAddFloatField_MultipleFields(t *testing.T) {
	replace := webapi.OperationValues.Replace
	pathStory := testPathStoryPoints
	pathEst := "/fields/Microsoft.VSTS.Scheduling.OriginalEstimate"
	pathComp := "/fields/Microsoft.VSTS.Scheduling.CompletedWork"

	var ops []webapi.JsonPatchOperation

	addFloatField(&ops, &replace, &pathStory, 3.0)
	addFloatField(&ops, &replace, &pathEst, 0)
	addFloatField(&ops, &replace, &pathComp, 1.5)

	if len(ops) != 2 {
		t.Fatalf("expected 2 operations (zero estimate skipped), got %d", len(ops))
	}

	if ops[0].Value != 3.0 {
		t.Errorf("expected first value 3.0, got %v", ops[0].Value)
	}

	if ops[1].Value != 1.5 {
		t.Errorf("expected second value 1.5, got %v", ops[1].Value)
	}
}

func TestAddIntField_WithNonZeroValue(t *testing.T) {
	add := webapi.OperationValues.Add
	pathPriority := "/fields/Microsoft.VSTS.Common.Priority"

	var ops []webapi.JsonPatchOperation

	addIntField(&ops, &add, &pathPriority, 2)

	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}

	if *ops[0].Op != add {
		t.Errorf("expected Add operation, got %v", *ops[0].Op)
	}

	if *ops[0].Path != pathPriority {
		t.Errorf("expected path %s, got %s", pathPriority, *ops[0].Path)
	}

	if ops[0].Value != 2 {
		t.Errorf("expected value 2, got %v", ops[0].Value)
	}
}

func TestAddIntField_ZeroValue(t *testing.T) {
	add := webapi.OperationValues.Add
	pathPriority := "/fields/Microsoft.VSTS.Common.Priority"

	var ops []webapi.JsonPatchOperation

	addIntField(&ops, &add, &pathPriority, 0)

	if len(ops) != 0 {
		t.Errorf("expected 0 operations for zero value, got %d", len(ops))
	}
}
