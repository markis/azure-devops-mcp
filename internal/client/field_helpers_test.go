package client

import (
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
)

func TestAddHTMLField(t *testing.T) {
	add := webapi.OperationValues.Add
	path := "/fields/System.Description"

	tests := []struct {
		name      string
		value     string
		wantOps   int
		wantValue string
	}{
		{
			"empty string - no operation",
			"",
			0,
			"",
		},
		{
			"markdown text - converted and added",
			"**bold** text",
			1,
			"<p><strong>bold</strong> text</p>\n",
		},
		{
			"HTML input - sanitized and added",
			"<p>text</p>",
			1,
			"<p>text</p>",
		},
		{
			"dangerous HTML - sanitized",
			"<p>text</p><script>alert('xss')</script>",
			1,
			"<p>text</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ops []webapi.JsonPatchOperation
			addHTMLField(&ops, &add, &path, tt.value)

			if len(ops) != tt.wantOps {
				t.Errorf("got %d operations, want %d", len(ops), tt.wantOps)
				return
			}

			if tt.wantOps > 0 {
				got, ok := ops[0].Value.(string)
				if !ok {
					t.Errorf("expected string value, got %T", ops[0].Value)
					return
				}

				if got != tt.wantValue {
					t.Errorf("got value %q, want %q", got, tt.wantValue)
				}
			}
		})
	}
}
