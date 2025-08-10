package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// Client は GitHub API クライアントのインターフェース
type Client interface {
	// IsAuthenticated は GitHub CLI の認証状態を確認する
	IsAuthenticated() error

	// GetCurrentUser は現在のユーザー情報を取得する
	GetCurrentUser() (*User, error)

	// GetTemplateRepositories は認証ユーザーのテンプレートリポジトリを取得する
	GetTemplateRepositories(ctx context.Context) ([]models.Template, error)

	// CreateRepository はリポジトリを作成する（gh コマンド経由）
	CreateRepository(ctx context.Context, config models.RepositoryConfig, template *models.Template) error
}

// User は GitHub ユーザーの情報を表す
type User struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// client は GitHub API クライアントの実装
type client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// NewClient は GitHub API クライアントを作成する
func NewClient() *client {
	return &client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.github.com",
		userAgent: "gh-wizard/1.0",
	}
}

// IsAuthenticated は GitHub CLI の認証状態を確認する
func (c *client) IsAuthenticated() error {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("GitHub CLI の認証確認に失敗しました: %w\n出力: %s", err, string(output))
	}

	if !strings.Contains(string(output), "Logged in to github.com") {
		return fmt.Errorf("GitHub CLI にログインしていません。'gh auth login' を実行してください")
	}

	return nil
}

// GetCurrentUser は現在のユーザー情報を取得する
func (c *client) GetCurrentUser() (*User, error) {
	cmd := exec.Command("gh", "api", "user")
	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}

	var user User
	if err := json.Unmarshal(output, &user); err != nil {
		return nil, fmt.Errorf("ユーザー情報の解析に失敗しました: %w", err)
	}

	return &user, nil
}

// GetTemplateRepositories は認証ユーザーのテンプレートリポジトリを取得する
func (c *client) GetTemplateRepositories(ctx context.Context) ([]models.Template, error) {
	// まず現在のユーザーを取得
	_, err := c.GetCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗: %w", err)
	}

	// ユーザーのリポジトリ一覧を取得（テンプレートのみ）
	cmd := exec.Command("gh", "api", "user/repos", "--paginate", "-q", ".[] | select(.is_template == true)")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("テンプレートリポジトリの取得に失敗しました: %w", err)
	}

	// JSON の配列型式に変換（gh api の出力は改行区切り）
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		// テンプレートが見つからない場合
		return []models.Template{}, nil
	}

	var templates []models.Template
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var repo GitHubRepository
		if err := json.Unmarshal([]byte(line), &repo); err != nil {
			// エラーログを出力するが処理は継続
			fmt.Printf("リポジトリ情報の解析をスキップ: %v\n", err)
			continue
		}

		template := models.Template{
			ID:          fmt.Sprintf("%d", repo.ID),
			Name:        repo.Name,
			FullName:    repo.FullName,
			Owner:       repo.Owner.Login,
			Description: repo.Description,
			Stars:       repo.StargazersCount,
			Forks:       repo.ForksCount,
			Language:    repo.Language,
			UpdatedAt:   repo.UpdatedAt,
			IsTemplate:  repo.IsTemplate,
			Private:     repo.Private,
		}

		templates = append(templates, template)
	}

	return templates, nil
}

// CreateRepository はリポジトリを作成する
func (c *client) CreateRepository(ctx context.Context, config models.RepositoryConfig, template *models.Template) error {
	// gh repo create コマンドの引数を生成
	args := config.GetGHCommand(template)

	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("リポジトリ作成に失敗しました: %w\n出力: %s", err, string(output))
	}

	return nil
}

// GitHubRepository は GitHub API のリポジトリレスポンスを表す
type GitHubRepository struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	Private         bool      `json:"private"`
	IsTemplate      bool      `json:"is_template"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	Language        string    `json:"language"`
	UpdatedAt       time.Time `json:"updated_at"`
	Owner           struct {
		Login string `json:"login"`
	} `json:"owner"`
}
