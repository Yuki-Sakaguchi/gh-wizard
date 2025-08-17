package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/cli/go-gh/v2/pkg/api"
)

// RepositoryService はGitHubリポジトリ操作を提供する
type RepositoryService struct {
	client *api.RESTClient
}

// NewRepositoryService は新しいリポジトリサービスを作成する
func NewRepositoryService() (*RepositoryService, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, models.NewGitHubError(
			"GitHub CLI の初期化に失敗しました",
			err,
		)
	}

	return &RepositoryService{client: client}, nil
}

// CreateRepository はGitHubリポジトリを作成する
func (rs *RepositoryService) CreateRepository(ctx context.Context, config *models.ProjectConfig) (*RepositoryInfo, error) {
	if !config.CreateGitHub {
		return nil, nil // GitHubリポジトリ作成が不要
	}

	// 現在のユーザー情報を取得
	user, err := rs.getCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// リポジトリの重複チェック
	if err := rs.checkRepositoryExists(ctx, user.Login, config.Name); err != nil {
		return nil, err
	}

	// リポジトリ作成
	repoInfo, err := rs.createRepositoryViaAPI(ctx, config)
	if err != nil {
		return nil, err
	}

	return repoInfo, nil
}

// getCurrentUser は現在のGitHubユーザー情報を取得する
func (rs *RepositoryService) getCurrentUser(ctx context.Context) (*GitHubUser, error) {
	var user GitHubUser
	err := rs.client.Get("user", &user)
	if err != nil {
		return nil, models.NewGitHubError(
			"ユーザー情報の取得に失敗しました",
			err,
		)
	}
	return &user, nil
}

// checkRepositoryExists はリポジトリの重複をチェックする
func (rs *RepositoryService) checkRepositoryExists(ctx context.Context, owner, name string) error {
	var repo RepositoryInfo
	err := rs.client.Get(fmt.Sprintf("repos/%s/%s", owner, name), &repo)

	if err == nil {
		// リポジトリが存在する
		return models.NewGitHubError(
			fmt.Sprintf("リポジトリ '%s/%s' は既に存在します", owner, name),
			nil,
		)
	}

	// 404エラーの場合は正常（リポジトリが存在しない）
	if strings.Contains(err.Error(), "404") {
		return nil
	}

	// その他のエラー
	return models.NewGitHubError(
		"リポジトリの存在確認に失敗しました",
		err,
	)
}

// createRepositoryViaAPI はAPI経由でリポジトリを作成する
func (rs *RepositoryService) createRepositoryViaAPI(ctx context.Context, config *models.ProjectConfig) (*RepositoryInfo, error) {
	// リクエストボディの作成
	createReq := CreateRepositoryRequest{
		Name:        config.Name,
		Description: config.Description,
		Private:     config.IsPrivate,
		AutoInit:    false, // テンプレートまたはローカルで初期化済み
	}

	// テンプレートリポジトリの場合
	if config.HasTemplate() {
		createReq.TemplateOwner = config.Template.Owner
		createReq.TemplateRepo = config.Template.Name
	}

	// リクエストボディをJSONに変換
	jsonData, err := json.Marshal(createReq)
	if err != nil {
		return nil, models.NewGitHubError(
			"リクエストデータの作成に失敗しました",
			err,
		)
	}

	var repoInfo RepositoryInfo
	err = rs.client.Post("user/repos", bytes.NewReader(jsonData), &repoInfo)
	if err != nil {
		return nil, models.NewGitHubError(
			fmt.Sprintf("リポジトリの作成に失敗しました: %v", err),
			err,
		)
	}

	return &repoInfo, nil
}

// データ構造

// GitHubUser はGitHubユーザー情報を表す
type GitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Email string `json:"email"`
}

// RepositoryInfo はリポジトリ情報を表す
type RepositoryInfo struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	FullName  string     `json:"full_name"`
	Owner     GitHubUser `json:"owner"`
	Private   bool       `json:"private"`
	HTMLURL   string     `json:"html_url"`
	CloneURL  string     `json:"clone_url"`
	SSHURL    string     `json:"ssh_url"`
	GitURL    string     `json:"git_url"`
	CreatedAt string     `json:"created_at"`
}

// CreateRepositoryRequest はリポジトリ作成リクエストを表す
type CreateRepositoryRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Private       bool   `json:"private"`
	AutoInit      bool   `json:"auto_init"`
	TemplateOwner string `json:"template_owner,omitempty"`
	TemplateRepo  string `json:"template_repo,omitempty"`
}
