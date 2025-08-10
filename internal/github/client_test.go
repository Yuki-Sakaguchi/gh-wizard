package github

import (
	"context"
	"testing"
)

func TestMockClient_GetTemplateRepositories(t *testing.T) {
	client := &MockClient{}
	ctx := context.Background()

	templates, err := client.GetTemplateRepositories(ctx)
	if err != nil {
		t.Fatalf("GetTemplateRepositories() error = %v", err)
	}

	if len(templates) == 0 {
		t.Error("テンプレートが取得できませんでした")
	}

	// 最初のテンプレートチェック
	template := templates[0]
	if template.Name == "" {
		t.Error("テンプレート名が空です")
	}

	if template.FullName == "" {
		t.Error("フル名がからです")
	}

	if !template.IsTemplate {
		t.Error("IsTemplate が false です")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Error("NewClient() が nil を返しました")
	}
}
