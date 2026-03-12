package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidType is returned when a JSON value has an unexpected type.
	ErrInvalidType = errors.New("invalid type")

	// ErrInvalidWorkItemID is returned when a work item ID cannot be parsed.
	ErrInvalidWorkItemID = errors.New("must be a valid work item ID or reference")

	// ErrInvalidDateFormat is returned when a date string cannot be parsed.
	ErrInvalidDateFormat = errors.New("invalid date format")

	// ErrInvalidBooleanString is returned when a boolean string cannot be parsed.
	ErrInvalidBooleanString = errors.New("invalid boolean string")

	// workItemRefRegex matches Azure DevOps work item references like "AB#123".
	workItemRefRegex = regexp.MustCompile(`^[A-Z]+#(\d+)$`)
)

// parseIntFromAny attempts to parse an integer from various JSON value types.
func parseIntFromAny(v any) (int, error) {
	switch val := v.(type) {
	case float64:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("%w: expected number or string, got %T", ErrInvalidType, v)
	}
}

// parseWorkItemID parses a work item ID from a string, supporting both
// plain numbers ("123") and Azure DevOps references ("AB#123").
func parseWorkItemID(s string) (int, error) {
	// Try plain number first
	if i, err := strconv.Atoi(s); err == nil {
		return i, nil
	}

	// Try work item reference (e.g., "AB#123")
	if matches := workItemRefRegex.FindStringSubmatch(s); len(matches) > 1 {
		return strconv.Atoi(matches[1])
	}

	return 0, ErrInvalidWorkItemID
}

// FlexID is a work item ID that can be unmarshaled from a number, string number,
// or Azure DevOps work item reference (e.g., "AB#123").
type FlexID int

// UnmarshalJSON implements json.Unmarshaler for FlexID.
func (f *FlexID) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	var (
		i   int
		err error
	)

	if s, ok := v.(string); ok {
		i, err = parseWorkItemID(s)
	} else {
		i, err = parseIntFromAny(v)
	}

	if err != nil {
		return err
	}

	*f = FlexID(i)

	return nil
}

// MarshalJSON implements json.Marshaler for FlexID.
func (f FlexID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// JSONSchemaExtend customizes the JSON Schema to accept both integer and string types.
func (*FlexID) JSONSchemaExtend(schema *map[string]any) {
	(*schema)["oneOf"] = []map[string]any{
		{"type": "integer"},
		{"type": "string"},
	}
	delete(*schema, "type")
}

// FlexFloat is a float that can be unmarshaled from either a number or string.
type FlexFloat float64

// UnmarshalJSON implements json.Unmarshaler for FlexFloat.
func (f *FlexFloat) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch val := v.(type) {
	case float64:
		*f = FlexFloat(val)
	case string:
		if val == "" {
			*f = 0
			return nil
		}

		fl, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("must be a valid number: %w", err)
		}

		*f = FlexFloat(fl)
	case nil:
		*f = 0
	default:
		return fmt.Errorf("%w: must be a number or string, got %T", ErrInvalidType, v)
	}

	return nil
}

// MarshalJSON implements json.Marshaler for FlexFloat.
func (f FlexFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(f))
}

// FlexInt is an int that can be unmarshaled from either a number or string.
type FlexInt int

// UnmarshalJSON implements json.Unmarshaler for FlexInt.
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	i, err := parseIntFromAny(v)
	if err != nil {
		return err
	}

	*f = FlexInt(i)

	return nil
}

// MarshalJSON implements json.Marshaler for FlexInt.
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// FlexDateTime is a time.Time that can be unmarshaled from various date/time formats.
type FlexDateTime time.Time

// UnmarshalJSON implements json.Unmarshaler for FlexDateTime.
func (f *FlexDateTime) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch val := v.(type) {
	case string:
		if val == "" {
			*f = FlexDateTime(time.Time{})
			return nil
		}

		// Try multiple formats
		formats := []string{
			time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
			time.RFC3339Nano,      // "2006-01-02T15:04:05.999999999Z07:00"
			"2006-01-02T15:04:05", // Without timezone
			"2006-01-02",          // Date only
		}

		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				*f = FlexDateTime(t)
				return nil
			}
		}

		return fmt.Errorf("%w: %s (expected ISO8601 or YYYY-MM-DD)", ErrInvalidDateFormat, val)
	case nil:
		*f = FlexDateTime(time.Time{})
	default:
		return fmt.Errorf("%w: must be a date string, got %T", ErrInvalidType, v)
	}

	return nil
}

// MarshalJSON implements json.Marshaler for FlexDateTime.
func (f FlexDateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(f)
	if t.IsZero() {
		return json.Marshal(nil)
	}

	return json.Marshal(t.Format(time.RFC3339))
}

// JSONSchemaExtend customizes the JSON Schema for FlexDateTime.
func (*FlexDateTime) JSONSchemaExtend(schema *map[string]any) {
	(*schema)["type"] = "string"
	(*schema)["format"] = "date-time"
	(*schema)["description"] = "ISO 8601 date-time or date (YYYY-MM-DD)"
}

// FlexBool is a bool that can be unmarshaled from various formats.
type FlexBool bool

// UnmarshalJSON implements json.Unmarshaler for FlexBool.
func (f *FlexBool) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch val := v.(type) {
	case bool:
		*f = FlexBool(val)
	case string:
		lower := strings.ToLower(val)
		switch lower {
		case "true", "yes", "1":
			*f = FlexBool(true)
		case "false", "no", "0", "":
			*f = FlexBool(false)
		default:
			return fmt.Errorf("%w: %s (expected true/false, yes/no, or 1/0)", ErrInvalidBooleanString, val)
		}
	case float64:
		*f = FlexBool(val != 0)
	case nil:
		*f = FlexBool(false)
	default:
		return fmt.Errorf("%w: must be a boolean, got %T", ErrInvalidType, v)
	}

	return nil
}

// MarshalJSON implements json.Marshaler for FlexBool.
func (f FlexBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(f))
}
