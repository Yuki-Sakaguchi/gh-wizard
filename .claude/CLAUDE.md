# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gh-wizard**は魔法のように簡単で直感的なGitHubリポジトリ作成ウィザードです。GitHub CLIの拡張機能として動作し、Bubble TeaフレームワークによるTUI（ターミナル・ユーザー・インターフェース）を提供します。

## 開発コマンド

### ビルド・実行
```bash
# アプリケーションのビルド
go build -o gh-wizard .

# GitHub CLI拡張として実行
gh wizard

# 依存関係の整理
go mod tidy
```

### テスト
```bash
# 全てのテストを実行
go test -v ./...

# 特定のパッケージのテスト
go test -v ./internal/models
```

### 開発環境
- Go 1.24.5+
- GitHub CLI v2.0+ (gh commandが必要)
- 必要な権限: GitHubへの認証済み環境

## アーキテクチャ概要

### プロジェクト構造
```
gh-wizard/
├── main.go              # エントリーポイント
├── cmd/
│   └── root.go          # Cobraを使用したCLI設定
├── internal/
│   ├── models/          # データモデルとビジネスロジック
│   ├── config/          # 設定管理（予定）
│   ├── github/          # GitHub API クライアント（予定）
│   ├── tui/            # TUIコンポーネント（予定）
│   └── utils/          # ユーティリティ関数（予定）
└── pkg/                # 公開API（予定）
```

### 技術スタック
- **CLI Framework**: Cobra
- **TUI Framework**: Bubble Tea（予定）
- **GitHub API**: github.com/cli/go-gh/v2（予定）
- **スタイリング**: Lipgloss（予定）

### コアモデル

#### WizardState
ウィザード全体の状態管理を行う中核モデル。以下のステップで進行：
1. `StepWelcome` - ようこそ画面
2. `StepTemplateSelection` - テンプレート選択
3. `StepRepositorySettings` - リポジトリ設定
4. `StepConfirmation` - 確認画面
5. `StepExecution` - 実行
6. `StepCompleted` - 完了

#### Template
GitHubテンプレートリポジトリの情報を格納。`GetDisplayName()`で表示用の名前を生成。

#### RepositoryConfig
リポジトリ作成設定。`Validate()`でバリデーション、`GetGHCommand()`で`gh repo create`コマンドを生成。

## 開発ガイドライン

### 実装優先順位
現在Phase 1（基盤構築）段階。以下の順序で開発中：
1. データモデル（✅ 実装済み）
2. GitHub APIクライアント
3. TUIコンポーネント
4. ウィザードフロー統合

### 既知の課題
- `models/wizard.go:52` - タイポ "SholdClone" → "ShouldClone"
- `models/repository.go:10,43` - 同様のタイポ
- テストファイルは`internal/models/template_test.go`のみ存在

### 日本語対応
- UIテキストは日本語
- エラーメッセージも日本語
- コメントとドキュメントは日本語優先

### テスト方針
- テーブル駆動テスト使用（既存のtemplate_test.goを参照）
- 各モデルに対応するテストファイル作成が必要
- カバレッジ目標: 80%以上

## 注意事項

- Bubble Teaライブラリが`go.mod`にコメントアウトされているため、TUI実装時に有効化が必要
- GitHub API連携は`github.com/cli/go-gh/v2`を使用予定
- 設定ファイルはYAML形式で`~/.config/gh-wizard/config.yml`に保存予定