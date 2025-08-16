# gh-wizard 最終要件定義書

## 1. プロジェクト概要

### 1.1 プロジェクト名
**`gh-wizard`** - GitHub CLI拡張機能

### 1.2 コンセプト
「シンプルで直感的なGitHubリポジトリ作成ウィザード」

### 1.3 プロジェクトの目的
テンプレートリポジトリを使った新しいプロジェクト作成を、create-next-appのようにシンプルで直感的にする。

### 1.4 ターゲットユーザー
- GitHub CLIを使う開発者（初心者〜上級者）
- テンプレートリポジトリを頻繁に使う開発者
- プロジェクト立ち上げを効率化したい開発者

### 1.5 既存ツールとの差別化
| 比較項目 | gh-wizard | gh-template | create-next-app |
|---------|-----------|-------------|-----------------|
| **用途** | 汎用テンプレート活用 | テンプレート加工 | Next.js特化 |
| **UI** | シンプルな対話UI | CLI引数のみ | シンプルな対話UI |
| **複雑さ** | 低い | 中程度 | 低い |
| **対象レベル** | 全レベル | 中級者以上 | 全レベル |

## 2. 機能要件

### 2.1 コア機能（MVP）

#### 2.1.1 ウィザード起動・ナビゲーション
- **FR-001**: `gh wizard` コマンドでウィザード起動
- **FR-002**: シンプルなキーボードナビゲーション（↑↓, Enter, Esc）
- **FR-003**: 順次質問形式のガイド

#### 2.1.2 テンプレート選択機能
- **FR-004**: ログイン中ユーザーのテンプレートリポジトリ自動取得
- **FR-005**: テンプレートのシンプルなリスト表示
- **FR-006**: 「テンプレートなし」オプション

#### 2.1.3 プロジェクト設定機能
- **FR-007**: プロジェクト名の入力・バリデーション
- **FR-008**: 説明文の入力（オプション）
- **FR-009**: GitHubリポジトリ作成可否の選択
- **FR-010**: 公開設定の選択（Public/Private）※GitHub作成時のみ

#### 2.1.4 実行機能
- **FR-011**: テンプレートからローカルディレクトリ作成
- **FR-012**: ローカルGitリポジトリ初期化
- **FR-013**: GitHub CLIによるリポジトリ作成（条件付き）
- **FR-014**: ローカルリポジトリのプッシュ（条件付き）
- **FR-015**: 成功・失敗フィードバック

## 3. ユーザーインタラクションフロー

### 3.1 質問フロー
```
$ gh wizard

🧙‍♂️ GitHub Repository Wizard

? テンプレートを選択してください:
  ▸ nextjs-starter (⭐ 15)
    react-component (⭐ 8)  
    go-cli-tool (⭐ 5)
    テンプレートなし

? プロジェクト名: my-awesome-project

? 説明 (オプション): My awesome Next.js project

? GitHubにリポジトリを作成しますか？ (y/n)

? プライベートリポジトリにしますか？ (y/n)  # GitHub作成時のみ表示

✓ テンプレートからディレクトリを作成中...
✓ GitHubリポジトリを作成中...           # GitHub作成時のみ
✓ ローカルリポジトリをプッシュ中...       # GitHub作成時のみ
✅ 完了！ プロジェクトの準備ができました
```

### 3.2 実行パターン

#### パターンA: GitHubリポジトリ作成あり
1. テンプレートからローカルディレクトリ作成
2. ローカルでGit初期化
3. GitHubリポジトリ作成
4. リモートリポジトリにプッシュ

#### パターンB: ローカルのみ
1. テンプレートからローカルディレクトリ作成
2. ローカルでGit初期化

## 4. 技術仕様

### 4.1 技術スタック

#### 4.1.1 言語・主要ライブラリ
```go
// 必須ライブラリ
github.com/AlecAivazis/survey/v2        // シンプルなCLI質問フレームワーク
github.com/cli/go-gh/v2                 // GitHub API クライアント

// 追加ライブラリ
github.com/spf13/cobra                  // CLI引数解析
gopkg.in/yaml.v3                        // 設定ファイル
```

### 4.2 アーキテクチャ設計

#### 4.2.1 プロジェクト構造
```
gh-wizard/
├── cmd/
│   ├── root.go              # CLI設定・引数処理
│   └── wizard.go            # メインコマンド
├── internal/
│   ├── config/
│   │   ├── config.go        # 設定管理
│   │   └── defaults.go      # デフォルト値
│   ├── github/
│   │   ├── client.go        # GitHub APIクライアント
│   │   ├── templates.go     # テンプレート操作
│   │   └── repository.go    # リポジトリ操作
│   ├── models/
│   │   ├── template.go      # テンプレート情報
│   │   └── project.go       # プロジェクト設定
│   ├── wizard/
│   │   ├── questions.go     # 質問定義
│   │   ├── executor.go      # 実行ロジック
│   │   └── validator.go     # バリデーション
│   └── utils/
│       └── git.go           # Git操作
├── main.go                  # エントリーポイント
├── go.mod
├── go.sum
└── README.md
```

### 4.3 データモデル設計

#### 4.3.1 メインモデル
```go
// プロジェクト設定
type ProjectConfig struct {
    Name           string
    Description    string
    Template       *Template
    CreateGitHub   bool
    IsPrivate      bool
    LocalPath      string
}

// テンプレート情報
type Template struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    FullName    string    `json:"full_name"`
    Owner       string    `json:"owner"`
    Description string    `json:"description"`
    Stars       int       `json:"stargazers_count"`
    Language    string    `json:"language"`
    IsTemplate  bool      `json:"is_template"`
}

// 質問の回答
type Answers struct {
    Template    string `survey:"template"`
    ProjectName string `survey:"projectName"`
    Description string `survey:"description"`
    CreateGitHub bool  `survey:"createGitHub"`
    IsPrivate   bool   `survey:"isPrivate"`
}
```

## 5. 実装計画

### 5.1 Phase 1: 基盤構築 (1週間)
- [ ] プロジェクト構造の構築
- [ ] GitHub APIクライアントの実装
- [ ] 基本データモデルの定義
- [ ] Survey質問システムの実装

### 5.2 Phase 2: コア機能実装 (1週間)
- [ ] テンプレート取得・表示機能
- [ ] ローカルディレクトリ作成機能
- [ ] GitHubリポジトリ作成機能
- [ ] エラーハンドリング

### 5.3 Phase 3: 改善・リリース準備 (3日)
- [ ] テストの追加
- [ ] ドキュメントの作成
- [ ] リリースの準備

## 6. 非機能要件

### 6.1 パフォーマンス
- **NFR-001**: 起動時間1秒以内
- **NFR-002**: テンプレート一覧取得3秒以内
- **NFR-003**: メモリ使用量20MB以下

### 6.2 互換性・環境
- **NFR-004**: GitHub CLI v2.0+ 対応
- **NFR-005**: Go 1.19+ で動作
- **NFR-006**: macOS, Linux, Windows対応

### 6.3 ユーザビリティ・品質
- **NFR-007**: 直感的で分かりやすいCLI
- **NFR-008**: 分かりやすいエラーメッセージ
- **NFR-009**: 完全なキーボード操作

## 7. 成功指標

### 7.1 技術指標
- **コード量**: 500行以下（従来の1/6に削減）
- **保守性**: 1ファイル10分以内での理解可能
- **安定性**: クラッシュ率0.1%以下

### 7.2 ユーザー指標
- **使いやすさ**: 初回使用でのタスク完了率95%以上
- **効率性**: 従来方法と比較して70%の時間短縮
- **満足度**: create-next-app並みのシンプルさ

## 8. 実装例

### 8.1 メイン処理フロー
```go
func RunWizard() error {
    // 1. テンプレート取得
    templates, err := github.GetUserTemplates()
    if err != nil {
        return err
    }

    // 2. 質問実行
    answers := &Answers{}
    err = survey.Ask(createQuestions(templates), answers)
    if err != nil {
        return err
    }

    // 3. プロジェクト作成実行
    config := buildProjectConfig(answers, templates)
    return executor.CreateProject(config)
}
```

### 8.2 質問定義例
```go
func createQuestions(templates []Template) []*survey.Question {
    templateOptions := make([]string, len(templates)+1)
    for i, t := range templates {
        templateOptions[i] = fmt.Sprintf("%s (%d⭐)", t.Name, t.Stars)
    }
    templateOptions[len(templates)] = "テンプレートなし"

    questions := []*survey.Question{
        {
            Name: "template",
            Prompt: &survey.Select{
                Message: "テンプレートを選択してください:",
                Options: templateOptions,
            },
        },
        {
            Name: "projectName",
            Prompt: &survey.Input{
                Message: "プロジェクト名:",
            },
            Validate: validateProjectName,
        },
        {
            Name: "description",
            Prompt: &survey.Input{
                Message: "説明 (オプション):",
            },
        },
        {
            Name: "createGitHub",
            Prompt: &survey.Confirm{
                Message: "GitHubにリポジトリを作成しますか？",
                Default: true,
            },
        },
    }

    return questions
}
```

---

## 🚀 次のアクション

この要件定義書を基に、シンプルで効率的なgh-wizardの実装を開始できます：

1. **環境準備**: Survey v2ライブラリのセットアップ
2. **プロジェクト初期化**: 簡素化されたディレクトリ構造の作成
3. **Phase 1開始**: GitHub APIクライアントと質問システムから実装開始

準備が整いましたら、実装を開始いたします！ 🧙‍♂️✨