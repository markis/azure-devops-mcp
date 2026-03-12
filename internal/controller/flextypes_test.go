package controller_test

import (
	"encoding/json"
	"testing"
	"time"

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

func TestFlexDateTime_UnmarshalJSON_RFC3339(t *testing.T) {
	var dt controller.FlexDateTime

	err := json.Unmarshal([]byte(`"2024-03-15T10:30:00Z"`), &dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)

	actual := time.Time(dt)
	if !actual.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestFlexDateTime_UnmarshalJSON_RFC3339Nano(t *testing.T) {
	var dt controller.FlexDateTime

	err := json.Unmarshal([]byte(`"2024-03-15T10:30:00.123456789Z"`), &dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2024, 3, 15, 10, 30, 0, 123456789, time.UTC)

	actual := time.Time(dt)
	if !actual.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestFlexDateTime_UnmarshalJSON_WithoutTimezone(t *testing.T) {
	var dt controller.FlexDateTime

	err := json.Unmarshal([]byte(`"2024-03-15T10:30:00"`), &dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)

	actual := time.Time(dt)
	if !actual.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestFlexDateTime_UnmarshalJSON_DateOnly(t *testing.T) {
	var dt controller.FlexDateTime

	err := json.Unmarshal([]byte(`"2024-03-15"`), &dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	actual := time.Time(dt)
	if !actual.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestFlexDateTime_UnmarshalJSON_EmptyString(t *testing.T) {
	var dt controller.FlexDateTime

	err := json.Unmarshal([]byte(`""`), &dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actual := time.Time(dt)
	if !actual.IsZero() {
		t.Errorf("expected zero time, got %v", actual)
	}
}

func TestFlexDateTime_UnmarshalJSON_Null(t *testing.T) {
	var dt controller.FlexDateTime

	err := json.Unmarshal([]byte(`null`), &dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actual := time.Time(dt)
	if !actual.IsZero() {
		t.Errorf("expected zero time, got %v", actual)
	}
}

func TestFlexDateTime_UnmarshalJSON_InvalidFormat(t *testing.T) {
	var dt controller.FlexDateTime

	err := json.Unmarshal([]byte(`"not a date"`), &dt)
	if err == nil {
		t.Error("expected error for invalid date format, got nil")
	}
}

func TestFlexDateTime_UnmarshalJSON_InvalidType(t *testing.T) {
	var dt controller.FlexDateTime

	err := json.Unmarshal([]byte(`123`), &dt)
	if err == nil {
		t.Error("expected error for non-string type, got nil")
	}
}

func TestFlexDateTime_MarshalJSON(t *testing.T) {
	dt := controller.FlexDateTime(time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC))

	data, err := json.Marshal(dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"2024-03-15T10:30:00Z"`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestFlexDateTime_MarshalJSON_ZeroTime(t *testing.T) {
	dt := controller.FlexDateTime(time.Time{})

	data, err := json.Marshal(dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `null`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestFlexBool_UnmarshalJSON_Bool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true", `true`, true},
		{"false", `false`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b controller.FlexBool

			err := json.Unmarshal([]byte(tt.input), &b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bool(b) != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b)
			}
		})
	}
}

func TestFlexBool_UnmarshalJSON_String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"string true", `"true"`, true},
		{"string false", `"false"`, false},
		{"string yes", `"yes"`, true},
		{"string no", `"no"`, false},
		{"string 1", `"1"`, true},
		{"string 0", `"0"`, false},
		{"string Yes (caps)", `"Yes"`, true},
		{"string No (caps)", `"No"`, false},
		{"string TRUE (upper)", `"TRUE"`, true},
		{"string FALSE (upper)", `"FALSE"`, false},
		{"empty string", `""`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b controller.FlexBool

			err := json.Unmarshal([]byte(tt.input), &b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bool(b) != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b)
			}
		})
	}
}

func TestFlexBool_UnmarshalJSON_Number(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"number 0", `0`, false},
		{"number 1", `1`, true},
		{"number 2", `2`, true},
		{"number -1", `-1`, true},
		{"float 0.0", `0.0`, false},
		{"float 1.5", `1.5`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b controller.FlexBool

			err := json.Unmarshal([]byte(tt.input), &b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bool(b) != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b)
			}
		})
	}
}

func TestFlexBool_UnmarshalJSON_Null(t *testing.T) {
	var b controller.FlexBool

	err := json.Unmarshal([]byte(`null`), &b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bool(b) != false {
		t.Errorf("expected false, got %v", b)
	}
}

func TestFlexBool_UnmarshalJSON_InvalidString(t *testing.T) {
	var b controller.FlexBool

	err := json.Unmarshal([]byte(`"invalid"`), &b)
	if err == nil {
		t.Error("expected error for invalid string, got nil")
	}
}

func TestFlexBool_UnmarshalJSON_InvalidType(t *testing.T) {
	var b controller.FlexBool

	err := json.Unmarshal([]byte(`[]`), &b)
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}

func TestFlexBool_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{"true", true, `true`},
		{"false", false, `false`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := controller.FlexBool(tt.value)

			data, err := json.Marshal(b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}
