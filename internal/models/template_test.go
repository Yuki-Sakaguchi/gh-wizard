package models

import (
	"testing"
)

func TestTemplate_GetDisplayNam(t *testing.T) {
	tests := []struct {
		name     string
		template Template
		want     string
	}{
		{
			name: "説明文ありの場合",
			template: Template{
				Name:        "test-template",
				Description: "テストテンプレート",
			},
			want: "test-template - テストテンプレート",
		},
		{
			name: "説明文なしの場合",
			template: Template{
				Name:        "test-template",
				Description: "",
			},
			want: "test-template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.template.GetDisplayName(); got != tt.want {
				t.Errorf("Template.GetDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}

}
