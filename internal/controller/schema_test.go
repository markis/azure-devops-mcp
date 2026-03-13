package controller

import (
	"slices"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestBuildCreateWorkItemSchema(t *testing.T) {
	schema := buildCreateWorkItemSchema()

	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	if schema.OneOf == nil || len(schema.OneOf) != 5 {
		t.Fatalf("expected 5 schemas in oneOf, got %d", len(schema.OneOf))
	}

	// Verify each schema has required type constraint
	expectedTypes := []string{"Bug", "Feature", "User Story", "Task", ""}

	for i, subSchema := range schema.OneOf {
		verifySchemaStructure(t, i, subSchema, expectedTypes)
	}
}

// verifySchemaStructure checks that a schema has proper structure and required fields.
func verifySchemaStructure(t *testing.T, index int, subSchema *jsonschema.Schema, expectedTypes []string) {
	t.Helper()

	if subSchema.Properties == nil {
		t.Errorf("schema %d: expected properties, got nil", index)
		return
	}

	verifyTypeProperty(t, index, subSchema, expectedTypes)
	verifyTitleRequired(t, index, subSchema)
}

// verifyTypeProperty checks that the type property has the correct const value.
func verifyTypeProperty(t *testing.T, index int, subSchema *jsonschema.Schema, expectedTypes []string) {
	t.Helper()

	typeSchema, ok := subSchema.Properties["type"]
	if !ok {
		t.Errorf("schema %d: missing 'type' property", index)
		return
	}

	// Check const value (empty string for "Other" schema)
	if index < 4 {
		if typeSchema.Const == nil {
			t.Errorf("schema %d: expected const=%q, got nil", index, expectedTypes[index])
		} else if constVal, ok := (*typeSchema.Const).(string); !ok || constVal != expectedTypes[index] {
			t.Errorf("schema %d: expected const=%q, got %v", index, expectedTypes[index], *typeSchema.Const)
		}
	}
}

// verifyTitleRequired checks that title is in the required fields list.
func verifyTitleRequired(t *testing.T, index int, subSchema *jsonschema.Schema) {
	t.Helper()

	if !slices.Contains(subSchema.Required, "title") {
		t.Errorf("schema %d: 'title' should be required", index)
	}
}

func TestBuildCreateWorkItemSchema_BugHasSpecificFields(t *testing.T) {
	schema := buildCreateWorkItemSchema()
	bugSchema := schema.OneOf[0]

	requiredFields := []string{"system_info", "blocked", "proposed_fix"}
	for _, field := range requiredFields {
		if _, ok := bugSchema.Properties[field]; !ok {
			t.Errorf("Bug schema missing field: %s", field)
		}
	}
}

func TestBuildCreateWorkItemSchema_FeatureHasSpecificFields(t *testing.T) {
	schema := buildCreateWorkItemSchema()
	featureSchema := schema.OneOf[1]

	requiredFields := []string{"at_risk", "documentation", "delivery_risk", "risk_reason", "mitigation_plan"}
	for _, field := range requiredFields {
		if _, ok := featureSchema.Properties[field]; !ok {
			t.Errorf("Feature schema missing field: %s", field)
		}
	}

	// Verify at_risk and documentation are required
	foundAtRisk := false
	foundDocumentation := false

	for _, req := range featureSchema.Required {
		if req == "at_risk" {
			foundAtRisk = true
		}

		if req == "documentation" {
			foundDocumentation = true
		}
	}

	if !foundAtRisk {
		t.Error("Feature schema: 'at_risk' should be required")
	}

	if !foundDocumentation {
		t.Error("Feature schema: 'documentation' should be required")
	}
}

func TestBuildCreateWorkItemSchema_UserStoryHasSpecificFields(t *testing.T) {
	schema := buildCreateWorkItemSchema()
	userStorySchema := schema.OneOf[2]

	requiredFields := []string{"dev_owner", "poker"}
	for _, field := range requiredFields {
		if _, ok := userStorySchema.Properties[field]; !ok {
			t.Errorf("User Story schema missing field: %s", field)
		}
	}
}

func TestBuildCreateWorkItemSchema_TaskHasCommonFields(t *testing.T) {
	schema := buildCreateWorkItemSchema()
	taskSchema := schema.OneOf[3]

	// Verify Task has effort tracking fields
	effortFields := []string{"original_estimate", "completed_work", "remaining_work"}
	for _, field := range effortFields {
		if _, ok := taskSchema.Properties[field]; !ok {
			t.Errorf("Task schema missing field: %s", field)
		}
	}
}

func TestBuildCreateWorkItemSchema_OtherHasAllFields(t *testing.T) {
	schema := buildCreateWorkItemSchema()
	otherSchema := schema.OneOf[4]

	// Verify Other has all type-specific fields
	allTypeSpecificFields := []string{
		"system_info", "blocked", "proposed_fix", // Bug
		"at_risk", "documentation", "mitigation_plan", // Feature
		"dev_owner", "poker", // User Story
	}

	for _, field := range allTypeSpecificFields {
		if _, ok := otherSchema.Properties[field]; !ok {
			t.Errorf("Other schema missing field: %s", field)
		}
	}
}
