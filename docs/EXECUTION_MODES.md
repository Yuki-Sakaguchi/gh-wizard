# 実行モードについて

gh-wizardは環境変数により実行モードを制御できます。

## シミュレーションモード（デフォルト）

```bash
# デフォルト（何も設定しない場合）
./gh-wizard

# または明示的に指定
export GH_WIZARD_SIMULATION=true
./gh-wizard
```

**動作:**
- 実際のGitHub CLI / Git コマンドは実行されない
- リポジトリ作成やクローンなどをシミュレーション
- GitHub認証チェックをスキップ
- UI動作とフローを確認するのに最適

## 実際の実行モード

```bash
export GH_WIZARD_SIMULATION=false
./gh-wizard
```

**動作:**
- 実際のGitHub CLI コマンドを実行
- 本当にGitHubリポジトリが作成される
- GitHub CLI の認証が必要
- ローカルへのクローンも実際に実行される

**事前準備:**
1. GitHub CLI のインストール: `brew install gh`
2. GitHub CLI の認証: `gh auth login`

## 切り替え例

```bash
# 開発・テスト時
export GH_WIZARD_SIMULATION=true
./gh-wizard

# 実際にリポジトリを作成する時
export GH_WIZARD_SIMULATION=false
./gh-wizard
```

これにより、開発中はシミュレーションで動作確認し、実際にリポジトリを作成したい時だけ実モードに切り替えることができます。