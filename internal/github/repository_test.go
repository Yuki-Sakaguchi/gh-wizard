package github

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

func TestRepositoryService_CreateRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}

	// 実際のGitHub APIを使用するテストは慎重に実行
	t.Skip("実際のGitHub APIテストはCI環境でのみ実行")

	service, err := NewRepositoryService()
	if err != nil {
		t.Fatalf("サービス作成エラー: %v", err)
	}

	config := &models.ProjectConfig{
		Name:         "test-repo-" + fmt.Sprintf("%d", time.Now().Unix()),
		Description:  "テスト用リポジトリ",
		CreateGitHub: true,
		IsPrivate:    true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repoInfo, err := service.CreateRepository(ctx, config)
	if err != nil {
		t.Fatalf("リポジトリ作成エラー: %v", err)
	}

	if repoInfo == nil {
		t.Fatal("リポジトリ情報がnilです")
	}

	if repoInfo.Name != config.Name {
		t.Errorf("リポジトリ名が一致しません: got %s, want %s",
			repoInfo.Name, config.Name)
	}

	// クリーンアップ（実際のテストでは手動で削除）
	fmt.Printf("作成されたテストリポジトリ: %s\n", repoInfo.HTMLURL)
	fmt.Println("手動で削除してください")
}

func TestRepositoryService_checkRepositoryExists(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}

	service, err := NewRepositoryService()
	if err != nil {
		t.Fatalf("サービス作成エラー: %v", err)
	}

	ctx := context.Background()

	// 存在しないリポジトリ（エラーなし）
	err = service.checkRepositoryExists(ctx, "nonexistent-user", "nonexistent-repo")
	if err != nil {
		t.Errorf("存在しないリポジトリでエラーが返されました: %v", err)
	}

	// 存在するリポジトリ（エラーあり）
	err = service.checkRepositoryExists(ctx, "octocat", "Hello-World")
	if err == nil {
		t.Error("存在するリポジトリでエラーが期待されましたが、エラーが返されませんでした")
	}
}
