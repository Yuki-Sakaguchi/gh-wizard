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

func TestTemplate_GetShortDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "normal description",
			description: "A simple template",
			expected:    "A simple template",
		},
		{
			name:        "long description",
			description: "This is a very long description that exceeds the maximum character limit and should be truncated",
			expected:    "This is a very long description that exceeds the maximum character limit...",
		},
		{
			name:        "empty description",
			description: "",
			expected:    "説明なし",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := Template{Description: tt.description}
			result := template.GetShortDescription()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplate_GetRepoURL(t *testing.T) {
	template := Template{
		FullName: "user/awesome-repo",
	}

	expected := "https://github.com/user/awesome-repo"
	result := template.GetRepoURL()

	assert.Equal(t, expected, result)
}

func TestTemlate_IsPublic(t *testing.T) {
	tests := []struct {
		name     string
		private  bool
		expected bool
	}{
		{
			name:     "public repository",
			private:  false,
			expected: true,
		},
		{
			name:     "private repository",
			private:  true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := Template{Private: tt.private}
			result := template.GetIsPublic()
			assert.Equal(t, tt.expected, result)
		})
	}
}
