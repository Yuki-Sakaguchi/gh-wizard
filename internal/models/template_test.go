package models

import (
	"testing"
)

func TestTemplate_GetDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		template    Template
		expected    string
	}{
		{
			name: "説明文ありの場合",
			template: Template{
				Name:        "test-template",
				Description: "テストテンプレート",
			},
			expected: "test-template - テストテンプレート",
		},
		{
			name: "説明文なしの場合",
			template: Template{
				Name:        "test-template",
				Description: "",
			},
			expected: "test-template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.template.GetDisplayName()
			if got != tt.expected {
				t.Errorf("GetDisplayName() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
