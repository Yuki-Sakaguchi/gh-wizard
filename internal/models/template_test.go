package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		template Template
		expected string
	}{
		{
			name: "complate information",
			template: Template{
				Name:     "nextjs-starter",
				Stars:    15,
				Language: "TypeScript",
			},
			expected: "nextjs-starter (⭐ 15) [TypeScript]",
		},
		{
			name: "no stars",
			template: Template{
				Name:     "simple-template",
				Stars:    0,
				Language: "JavaScript",
			},
			expected: "simple-template [JavaScript]",
		},
		{
			name: "no language",
			template: Template{
				Name:  "basic-template",
				Stars: 5,
			},
			expected: "basic-template (⭐ 5)",
		},
		{
			name: "minimal info",
			template: Template{
				Name: "minimal-template",
			},
			expected: "minimal-template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.template.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

