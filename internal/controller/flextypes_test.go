package controller_test

import (
	"encoding/json"
	"testing"

	"github.com/markis/azure-devops-mcp/internal/controller"
)

func TestFlexID_UnmarshalJSON_Number(t *testing.T) {
	var id controller.FlexID

	err := json.Unmarshal([]byte(`123`), &id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if int(id) != 123 {
		t.Errorf("expected 123, got %d", id)
	}
}

func TestFlexID_UnmarshalJSON_StringNumber(t *testing.T) {
	var id controller.FlexID

	err := json.Unmarshal([]byte(`"456"`), &id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if int(id) != 456 {
		t.Errorf("expected 456, got %d", id)
	}
}

func TestFlexID_UnmarshalJSON_WorkItemReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"simple reference", `"AB#123"`, 123},
		{"long prefix", `"PROJECT#456"`, 456},
		{"single letter", `"A#789"`, 789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id controller.FlexID

			err := json.Unmarshal([]byte(tt.input), &id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if int(id) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, id)
			}
		})
	}
}

func TestFlexID_UnmarshalJSON_InvalidReference(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase prefix", `"ab#123"`},
		{"no hash", `"AB123"`},
		{"no prefix", `"#123"`},
		{"no number", `"AB#"`},
		{"invalid format", `"AB-123"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id controller.FlexID

			err := json.Unmarshal([]byte(tt.input), &id)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestFlexID_UnmarshalJSON_Null(t *testing.T) {
	var id controller.FlexID

	err := json.Unmarshal([]byte(`null`), &id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if int(id) != 0 {
		t.Errorf("expected 0, got %d", id)
	}
}

func TestFlexID_MarshalJSON(t *testing.T) {
	id := controller.FlexID(123)

	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `123`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestFlexFloat_UnmarshalJSON_Number(t *testing.T) {
	var f controller.FlexFloat

	err := json.Unmarshal([]byte(`3.14`), &f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if float64(f) != 3.14 {
		t.Errorf("expected 3.14, got %f", f)
	}
}

func TestFlexFloat_UnmarshalJSON_String(t *testing.T) {
	var f controller.FlexFloat

	err := json.Unmarshal([]byte(`"5.5"`), &f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if float64(f) != 5.5 {
		t.Errorf("expected 5.5, got %f", f)
	}
}

func TestFlexFloat_UnmarshalJSON_EmptyString(t *testing.T) {
	var f controller.FlexFloat

	err := json.Unmarshal([]byte(`""`), &f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if float64(f) != 0 {
		t.Errorf("expected 0, got %f", f)
	}
}

func TestFlexInt_UnmarshalJSON_Number(t *testing.T) {
	var i controller.FlexInt

	err := json.Unmarshal([]byte(`42`), &i)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if int(i) != 42 {
		t.Errorf("expected 42, got %d", i)
	}
}

func TestFlexInt_UnmarshalJSON_Float64(t *testing.T) {
	var i controller.FlexInt

	err := json.Unmarshal([]byte(`3.0`), &i)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if int(i) != 3 {
		t.Errorf("expected 3, got %d", i)
	}
}

func TestFlexInt_UnmarshalJSON_String(t *testing.T) {
	var i controller.FlexInt

	err := json.Unmarshal([]byte(`"99"`), &i)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if int(i) != 99 {
		t.Errorf("expected 99, got %d", i)
	}
}

func TestFlexInt_UnmarshalJSON_Null(t *testing.T) {
	var i controller.FlexInt

	err := json.Unmarshal([]byte(`null`), &i)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if int(i) != 0 {
		t.Errorf("expected 0, got %d", i)
	}
}

func TestFlexInt_UnmarshalJSON_InvalidString(t *testing.T) {
	var i controller.FlexInt

	err := json.Unmarshal([]byte(`"not a number"`), &i)
	if err == nil {
		t.Error("expected error for invalid string, got nil")
	}
}

func TestFlexInt_MarshalJSON(t *testing.T) {
	i := controller.FlexInt(123)

	data, err := json.Marshal(i)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `123`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}
