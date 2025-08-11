package models

import (
	"errors"
	"regexp"
	"strings"
)

// QuestionType は質問の種類を表す
type QuestionType int

const (
	QuestionTypeText QuestionType = iota
	QuestionTypeBool
	QuestionTypeSelect
)

// Question は連続入力形式での質問を表す
type Question struct {
	ID           string
	Type         QuestionType
	Title        string
	Description  string
	HelpText     string
	Required     bool
	Validation   ValidationFunc
	Options      []string // SelectTypeの選択肢
	DefaultValue string
}

// ValidationFunc は入力値の検証関数の型
type ValidationFunc func(value string) error

// Answer は質問に対する回答を表す
type Answer struct {
	QuestionID string
	Value      string
	IsValid    bool
	ErrorMsg   string
}

// QuestionFlow は質問フローを管理する
type QuestionFlow struct {
	Questions     []Question
	Answers       map[string]*Answer
	CurrentIndex  int
	IsCompleted   bool
	ValidationMsg string
}

// NewQuestionFlow は新しい質問フローを作成する
func NewQuestionFlow() *QuestionFlow {
	questions := []Question{
		{
			ID:          "repository_name",
			Type:        QuestionTypeText,
			Title:       "リポジトリ名を入力してください",
			Description: "GitHub上でのリポジトリ名を設定します",
			HelpText:    "英数字、ハイフン、アンダースコアが使用可能です（例: my-awesome-project）",
			Required:    true,
			Validation:  ValidateRepositoryName,
		},
		{
			ID:          "description",
			Type:        QuestionTypeText,
			Title:       "リポジトリの説明を入力してください（任意）",
			Description: "リポジトリの用途や内容を簡潔に説明します",
			HelpText:    "100文字以内で入力してください（空欄でもOK）",
			Required:    false,
			Validation:  ValidateDescription,
		},
		{
			ID:          "visibility",
			Type:        QuestionTypeSelect,
			Title:       "リポジトリの公開設定を選択してください",
			Description: "パブリック（公開）かプライベート（非公開）を選択します",
			HelpText:    "↑↓キーで選択、Enterで決定",
			Required:    true,
			Options:     []string{"private（非公開）", "public（公開）"},
			DefaultValue: "private（非公開）",
		},
		{
			ID:          "clone_after_create",
			Type:        QuestionTypeBool,
			Title:       "作成後にローカルにクローンしますか？",
			Description: "リポジトリ作成後、自動的にローカル環境にクローンします",
			HelpText:    "y/n または Yes/No で回答してください",
			Required:    true,
			DefaultValue: "yes",
			Validation:  ValidateBooleanInput,
		},
		{
			ID:          "add_readme",
			Type:        QuestionTypeBool,
			Title:       "READMEファイルを追加しますか？",
			Description: "初期のREADME.mdファイルを作成します（テンプレート使用時は無視されます）",
			HelpText:    "y/n または Yes/No で回答してください",
			Required:    true,
			DefaultValue: "yes",
			Validation:  ValidateBooleanInput,
		},
	}

	return &QuestionFlow{
		Questions:    questions,
		Answers:      make(map[string]*Answer),
		CurrentIndex: 0,
		IsCompleted:  false,
	}
}

// GetCurrentQuestion は現在の質問を取得する
func (qf *QuestionFlow) GetCurrentQuestion() *Question {
	if qf.CurrentIndex < 0 || qf.CurrentIndex >= len(qf.Questions) {
		return nil
	}
	return &qf.Questions[qf.CurrentIndex]
}

// GetProgress は進捗情報を取得する（現在の位置 / 総質問数）
func (qf *QuestionFlow) GetProgress() (current, total int) {
	return qf.CurrentIndex + 1, len(qf.Questions)
}

// GetProgressRatio は進捗の割合を取得する（0.0-1.0）
func (qf *QuestionFlow) GetProgressRatio() float64 {
	if len(qf.Questions) == 0 {
		return 0.0
	}
	return float64(qf.CurrentIndex+1) / float64(len(qf.Questions))
}

// SetAnswer は現在の質問に対する回答を設定する
func (qf *QuestionFlow) SetAnswer(value string) error {
	question := qf.GetCurrentQuestion()
	if question == nil {
		return errors.New("無効な質問です")
	}

	// 空の入力のチェック
	if strings.TrimSpace(value) == "" {
		if question.Required {
			return errors.New("この項目は必須です")
		}
		// 任意項目で空の場合はデフォルト値を使用
		value = question.DefaultValue
	}

	// バリデーション実行
	var validationError error
	if question.Validation != nil {
		validationError = question.Validation(value)
	}

	// 回答を保存
	answer := &Answer{
		QuestionID: question.ID,
		Value:      value,
		IsValid:    validationError == nil,
	}
	
	if validationError != nil {
		answer.ErrorMsg = validationError.Error()
	}

	qf.Answers[question.ID] = answer

	return validationError
}

// GoToNext は次の質問に進む
func (qf *QuestionFlow) GoToNext() bool {
	if qf.CurrentIndex < len(qf.Questions)-1 {
		qf.CurrentIndex++
		return true
	}
	
	// 全ての質問が完了
	qf.IsCompleted = true
	return false
}

// GoToPrevious は前の質問に戻る
func (qf *QuestionFlow) GoToPrevious() bool {
	if qf.CurrentIndex > 0 {
		qf.CurrentIndex--
		return true
	}
	return false
}

// GetAnswer は指定されたIDの回答を取得する
func (qf *QuestionFlow) GetAnswer(id string) *Answer {
	return qf.Answers[id]
}

// ToRepositoryConfig は回答からRepositoryConfigを生成する
func (qf *QuestionFlow) ToRepositoryConfig() *RepositoryConfig {
	config := &RepositoryConfig{}

	if answer := qf.GetAnswer("repository_name"); answer != nil {
		config.Name = answer.Value
	}

	if answer := qf.GetAnswer("description"); answer != nil {
		config.Description = answer.Value
	}

	if answer := qf.GetAnswer("visibility"); answer != nil {
		config.IsPrivate = strings.Contains(answer.Value, "private")
	}

	if answer := qf.GetAnswer("clone_after_create"); answer != nil {
		config.SholdClone = parseBooleanAnswer(answer.Value)
	}

	if answer := qf.GetAnswer("add_readme"); answer != nil {
		config.AddReadme = parseBooleanAnswer(answer.Value)
	}

	return config
}

// バリデーション関数

// ValidateRepositoryName はリポジトリ名を検証する
func ValidateRepositoryName(name string) error {
	if name == "" {
		return errors.New("リポジトリ名は必須です")
	}
	
	if len(name) > 100 {
		return errors.New("リポジトリ名は100文字以内で入力してください")
	}

	// GitHub リポジトリ名の規則に準拠
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, name)
	if !matched {
		return errors.New("リポジトリ名は英数字、ピリオド、ハイフン、アンダースコアのみ使用可能です")
	}

	// 先頭と末尾のピリオドとハイフンは不可
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") ||
		strings.HasSuffix(name, ".") || strings.HasSuffix(name, "-") {
		return errors.New("リポジトリ名の先頭・末尾にピリオドやハイフンは使用できません")
	}

	return nil
}

// ValidateDescription は説明を検証する
func ValidateDescription(desc string) error {
	if len(desc) > 100 {
		return errors.New("説明は100文字以内で入力してください")
	}
	return nil
}

// ValidateBooleanInput はブール値入力を検証する
func ValidateBooleanInput(value string) error {
	lower := strings.ToLower(strings.TrimSpace(value))
	validValues := []string{"y", "n", "yes", "no", "true", "false", "1", "0"}
	
	for _, valid := range validValues {
		if lower == valid {
			return nil
		}
	}
	
	return errors.New("y/n、yes/no、true/false、または1/0で入力してください")
}

// parseBooleanAnswer はブール値回答を解析する
func parseBooleanAnswer(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return lower == "y" || lower == "yes" || lower == "true" || lower == "1"
}