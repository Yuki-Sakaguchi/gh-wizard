# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

**gh-wizard** は GitHub CLI 拡張機能として開発される対話型リポジトリ作成ウィザードです。Go 言語で実装され、Bubble Tea フレームワークを使って美しい TUI (Terminal User Interface) を提供します。

## コア技術スタック

- **言語**: Go 1.19+
- **TUI フレームワーク**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **スタイリング**: Lipgloss (github.com/charmbracelet/lipgloss)
- **CLI 構築**: Cobra (github.com/spf13/cobra)
- **GitHub API**: go-gh (github.com/cli/go-gh/v2)

## プロジェクト構造（計画）

```
gh-wizard/
├── cmd/
│   └── root.go              # CLI設定・引数処理
├── internal/
│   ├── config/              # 設定管理
│   ├── github/              # GitHub APIクライアント
│   ├── models/              # データモデル
│   ├── tui/                 # TUI画面実装
│   │   ├── app.go           # メインアプリケーション
│   │   ├── styles.go        # スタイル定義
│   │   ├── welcome.go       # ウェルカム画面
│   │   ├── template.go      # テンプレート選択画面
│   │   ├── settings.go      # リポジトリ設定画面
│   │   └── confirmation.go  # 確認画面
│   └── utils/               # ユーティリティ
├── pkg/
│   └── wizard/              # 公開API
└── main.go
```

## 開発コマンド

```bash
# プロジェクト初期化
go mod init github.com/Yuki-Sakaguchi/gh-wizard

# 依存関係インストール
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/cli/go-gh/v2
go get github.com/spf13/cobra

# 開発実行
go run main.go

# ビルド
go build -o gh-wizard

# テスト実行
go test ./...

# フォーマット
go fmt ./...
goimports -w .

# リンター
golangci-lint run
```

## Go 慣例・スタイル

- **命名**: 
  - Public は大文字開始 (User, GetName)
  - Private は小文字開始 (user, getName)
  - エラーは `Err` プレフィックス (ErrUserNotFound)
  - 定数は大文字スネークケース (MAX_RETRY_COUNT)

- **エラーハンドリング**: 
  - エラーは戻り値として返す
  - `fmt.Errorf("message: %w", err)` でエラーをラップ
  - 早期リターンパターンを使用

- **並行処理**:
  - goroutine + channel の組み合わせ
  - sync.WaitGroup でゴルーチンの完了待ち

## 実装方針

1. **MVP優先**: 基本的なウィザードフローから実装
2. **TUI重視**: 美しく直感的なユーザーインターフェース
3. **エラーハンドリング**: 分かりやすいエラーメッセージとガイダンス
4. **テスト**: 単体テスト・統合テストを含む
5. **パフォーマンス**: 起動2秒以内、レスポンス100ms以内

## GitHub CLI 拡張機能としての要件

- `gh wizard` コマンドで起動
- GitHub CLI 認証情報を使用
- `gh repo create` との統合

## 参考ドキュメント

詳細な要件や設計については以下を参照：
- `docs/requirement.md`: 完全な要件定義書
- `docs/go-study.md`: Go 言語学習ガイド (TypeScript経験者向け)