package wizard

import (
	"context"
	"fmt"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/utils"
)

// PushService はGitHubへのプッシュを管理する
type PushService struct {
	repoService *github.RepositoryService
	gitService  *utils.GitService
}

// NewPushService は新しいプッシュサービスを作成する
func NewPushService(projectPath string) (*PushService, error) {
	repoService, err := github.NewRepositoryService()
	if err != nil {
		return nil, err
	}

	gitService := utils.NewGitService(projectPath)

	return &PushService{
		repoService: repoService,
		gitService:  gitService,
	}, nil
}

// PushToGitHub はローカルリポジトリをGitHubにプッシュする
func (ps *PushService) PushToGitHub(ctx context.Context, config *models.ProjectConfig) error {
	if !config.CreateGitHub {
		return nil // GitHubプッシュが不要
	}

	// GitHubリポジトリを作成
	repoInfo, err := ps.repoService.CreateRepository(ctx, config)
	if err != nil {
		return err
	}

	if repoInfo == nil {
		return nil // リポジトリ作成がスキップされた
	}

	// Git初期化とコミット
	if err := ps.initializeLocalRepository(ctx, config); err != nil {
		return err
	}

	// リモートリポジトリの設定とプッシュ
	if err := ps.pushToRemoteRepository(ctx, repoInfo); err != nil {
		return err
	}

	// 成功メッセージ
	fmt.Printf("✅ GitHubリポジトリが作成されました: %s\n", repoInfo.HTMLURL)

	return nil
}

// initializeLocalRepository はローカルリポジトリを初期化する
func (ps *PushService) initializeLocalRepository(ctx context.Context, config *models.ProjectConfig) error {
	// Gitリポジトリ初期化
	if err := ps.gitService.InitializeRepository(ctx); err != nil {
		return err
	}

	// ファイルを追加
	if err := ps.gitService.AddAllFiles(ctx); err != nil {
		return err
	}

	// 初期コミット
	commitMessage := fmt.Sprintf("Initial commit for %s", config.Name)
	if config.HasTemplate() {
		commitMessage = fmt.Sprintf("Initial commit from template %s", config.Template.FullName)
	}

	if err := ps.gitService.CreateInitialCommit(ctx, commitMessage); err != nil {
		return err
	}

	return nil
}

// pushToRemoteRepository はリモートリポジトリにプッシュする
func (ps *PushService) pushToRemoteRepository(ctx context.Context, repoInfo *github.RepositoryInfo) error {
	// リモートリポジトリを追加
	if err := ps.gitService.AddRemote(ctx, "origin", repoInfo.CloneURL); err != nil {
		return err
	}

	// 現在のブランチを取得
	branch, err := ps.gitService.GetCurrentBranch(ctx)
	if err != nil {
		branch = "main" // デフォルト
	}

	// プッシュ実行
	if err := ps.gitService.PushToRemote(ctx, "origin", branch); err != nil {
		return models.NewGitHubError(
			fmt.Sprintf("GitHubへのプッシュに失敗しました。リポジトリURL: %s", repoInfo.HTMLURL),
			err,
		)
	}

	return nil
}

// SetupGitConfiguration はGit設定をセットアップする
func (ps *PushService) SetupGitConfiguration(ctx context.Context, name, email string) error {
	return ps.gitService.ConfigureUserInfo(ctx, name, email)
}
