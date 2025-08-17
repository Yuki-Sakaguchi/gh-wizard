package wizard

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// ProjectNameValidator はプロジェクト名のバリデーター
type ProjectNameValidator struct {
	minLength      int
	maxLength      int
	validPattern   *regexp.Regexp
	invalidPatterns []*regexp.Regexp
}

// NewProjectNameValidator は新しいプロジェクト名バリデーターを作成する
func NewProjectNameValidator() *ProjectNameValidator {
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	invalidPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[^\w.-]`), // 無効文字
		regexp.MustCompile(`^\.`),     // ドットで開始
		regexp.MustCompile(`\.$`),     // ドットで終了
		regexp.MustCompile(`^-`),      // ハイフンで開始
		regexp.MustCompile(`-$`),      // ハイフンで終了
	}

	return &ProjectNameValidator{
		minLength:       1,
		maxLength:       100,
		validPattern:    validPattern,
		invalidPatterns: invalidPatterns,
	}
}

// Validate はプロジェクト名を検証する
func (v *ProjectNameValidator) Validate(name string) error {
	if err := v.validateBasicRules(name); err != nil {
		return err
	}
	if err := v.validateGitHubRules(name); err != nil {
		return err
	}
	if err := v.validateReservedNames(name); err != nil {
		return err
	}
	if err := v.validateAdvancedRules(name); err != nil {
		return err
	}
	return nil
}

// validateBasicRules は基本的なバリデーションルールをチェックする
func (v *ProjectNameValidator) validateBasicRules(name string) error {
	if name == "" {
		return fmt.Errorf("プロジェクト名は必須です")
	}

	if len(name) < v.minLength {
		return fmt.Errorf("プロジェクト名は%d文字以上である必要があります", v.minLength)
	}

	if len(name) > v.maxLength {
		return fmt.Errorf("プロジェクト名は%d文字以内で入力してください", v.maxLength)
	}

	// スペースのみのチェック
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("プロジェクト名は必須です")
	}

	// 無効な文字チェック（英数字、ハイフン、アンダースコア、ドットのみ）
	if !isValidProjectName(name) {
		return fmt.Errorf("プロジェクト名に無効な文字が含まれています。英数字、ハイフン、アンダースコア、ドットのみ使用できます")
	}

	return nil
}

// validateGitHubRules はGitHubの制約をチェックする
func (v *ProjectNameValidator) validateGitHubRules(name string) error {
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		if strings.HasPrefix(name, ".") {
		return fmt.Errorf("ピリオドで始まるプロジェクト名は使用できません")
	}
	if strings.HasSuffix(name, ".") {
		return fmt.Errorf("ピリオドで終わるプロジェクト名は使用できません")
	}
	}

	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("ハイフンで始まるプロジェクト名は使用できません")
	}
	if strings.HasSuffix(name, "-") {
		return fmt.Errorf("ハイフンで終わるプロジェクト名は使用できません")
	}
	if strings.HasPrefix(name, "_") {
		return fmt.Errorf("アンダースコアで始まるプロジェクト名は使用できません")
	}
	if strings.HasSuffix(name, "_") {
		return fmt.Errorf("アンダースコアで終わるプロジェクト名は使用できません")
	}

	// 連続ピリオドや連続ハイフンのチェック
	if strings.Contains(name, "..") {
		return fmt.Errorf("連続するピリオドは使用できません")
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("連続するハイフンは使用できません")
	}
	if strings.Contains(name, "__") {
		return fmt.Errorf("連続するアンダースコアは使用できません")
	}

	return nil
}

// validateReservedNames は予約語をチェックする
func (v *ProjectNameValidator) validateReservedNames(name string) error {
	// システム予約語
	systemReserved := []string{".", "..", "CON", "PRN", "AUX", "NUL", "COM1", "LPT1"}
	upperName := strings.ToUpper(name)
	for _, reserved := range systemReserved {
		if upperName == reserved {
			return fmt.Errorf("'%s'は予約名のため使用できません", name)
		}
	}

	// Git関連の予約語
	gitReserved := []string{".git", ".github"}
	lowerNameForGit := strings.ToLower(name)
	for _, reserved := range gitReserved {
		if lowerNameForGit == reserved {
			return fmt.Errorf("'%s'は予約名のため使用できません", name)
		}
	}

	// 一般的な予約語
	generalReserved := []string{"api", "www", "mail", "ftp", "admin", "root", "test", "debug"}
	lowerName := strings.ToLower(name)
	for _, reserved := range generalReserved {
		if lowerName == reserved {
			return fmt.Errorf("'%s'は予約名のため使用できません", name)
		}
	}
	
	return nil
}

// validateAdvancedRules は高度なバリデーションルールをチェックする
func (v *ProjectNameValidator) validateAdvancedRules(name string) error {
	// 制御文字チェック
	if containsControlChars(name) {
		return fmt.Errorf("制御文字は使用できません")
	}

	// 全て数字の場合は警告
	if isAllDigits(name) {
		return fmt.Errorf("数字のみのプロジェクト名は推奨されません")
	}

	// 特殊文字が多すぎる場合 (40%以上が特殊文字)
	specialCount := countSpecialChars(name)
	if specialCount > 0 && float64(specialCount)/float64(len(name)) >= 0.4 {
		return fmt.Errorf("特殊文字が多すぎます")
	}

	return nil
}

// isValidProjectName はプロジェクト名が有効かチェックする
func isValidProjectName(name string) bool {
	// 英数字、ハイフン、アンダースコア、ドットのみ許可
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return validPattern.MatchString(name)
}

// isAllDigits は文字列が全て数字かどうかをチェックする
func isAllDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// countSpecialChars は特殊文字の数をカウントする
func countSpecialChars(s string) int {
	count := 0
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			count++
		}
	}
	return count
}

// containsControlChars は制御文字が含まれているかチェックする
func containsControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// containsStrictControlChars は説明文用の制御文字チェック（改行、タブ、CRは許可）
func containsStrictControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return true
		}
	}
	return false
}

// ValidateDescription は説明文を検証する
func ValidateDescription(description interface{}) error {
	// 型アサーション
	desc, ok := description.(string)
	if !ok {
		return fmt.Errorf("無効な入力タイプです")
	}

	maxLength := 500
	maxLines := 5

	if len(desc) > maxLength {
		return fmt.Errorf("説明は%d文字以内である必要があります", maxLength)
	}

	// 行数チェック
	lines := strings.Split(desc, "\n")
	if len(lines) > maxLines {
		return fmt.Errorf("説明は%d行以内である必要があります", maxLines)
	}

	// 制御文字チェック（改行、タブ、CRは許可）
	if containsStrictControlChars(desc) {
		return fmt.Errorf("説明に制御文字は使用できません")
	}

	return nil
}

// TemplateValidator はテンプレートのバリデーター
type TemplateValidator struct {
	availableTemplates []models.Template
}

// NewTemplateValidator は新しいテンプレートバリデーターを作成する
func NewTemplateValidator(templates ...[]models.Template) *TemplateValidator {
	validator := &TemplateValidator{}
	if len(templates) > 0 {
		validator.availableTemplates = templates[0]
	}
	return validator
}

// Validate はテンプレートを検証する
func (v *TemplateValidator) Validate(template *models.Template) error {
	if template == nil {
		return fmt.Errorf("テンプレートがnilです")
	}

	if template.FullName == "" {
		return fmt.Errorf("テンプレートのFullNameは必須です")
	}

	if template.Name == "" {
		return fmt.Errorf("テンプレートのNameは必須です")
	}

	if !template.IsTemplate {
		return fmt.Errorf("指定されたリポジトリはテンプレートではありません")
	}

	return nil
}

// ValidateTemplates は複数のテンプレートを検証する
func (v *TemplateValidator) ValidateTemplates(templates []models.Template) error {
	if len(templates) == 0 {
		return fmt.Errorf("テンプレートが指定されていません")
	}

	for i, template := range templates {
		if err := v.Validate(&template); err != nil {
			return fmt.Errorf("テンプレート[%d]: %w", i, err)
		}
	}

	return nil
}

// ValidateTemplateSelection はテンプレート選択を検証する
func (v *TemplateValidator) ValidateTemplateSelection(selection string) error {
	if selection == "" {
		return fmt.Errorf("テンプレートを選択してください")
	}

	if selection == "テンプレートなし" {
		return nil // 有効な選択
	}

	// 利用可能なテンプレートから検索（表示形式でも検索）
	for _, template := range v.availableTemplates {
		// 名前、FullName、または表示形式で比較
		displayName := template.GetDisplayName()
		if template.Name == selection || template.FullName == selection || displayName == selection {
			if !template.IsTemplate {
				return fmt.Errorf("'%s'はテンプレートとして設定されていません", selection)
			}
			return nil
		}
	}

	return fmt.Errorf("テンプレート'%s'は利用できません", selection)
}

// GetSurveyValidator はSurvey用のバリデーション関数を返す
func (v *TemplateValidator) GetSurveyValidator() func(interface{}) error {
	return func(ans interface{}) error {
		selection, ok := ans.(string)
		if !ok {
			return fmt.Errorf("無効な入力タイプです")
		}
		return v.ValidateTemplateSelection(selection)
	}
}

// ConfigValidator は設定のバリデーター
type ConfigValidator struct {
	projectValidator *ProjectNameValidator
	templateValidator *TemplateValidator
}

// NewConfigValidator は新しい設定バリデーターを作成する
func NewConfigValidator(templates ...[]models.Template) *ConfigValidator {
	validator := &ConfigValidator{
		projectValidator: NewProjectNameValidator(),
	}
	if len(templates) > 0 {
		validator.templateValidator = NewTemplateValidator(templates[0])
	} else {
		validator.templateValidator = NewTemplateValidator()
	}
	return validator
}

// Validate は ProjectConfig を検証する
func (v *ConfigValidator) Validate(config *models.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("設定がnilです")
	}

	// プロジェクト名検証
	if err := v.projectValidator.Validate(config.Name); err != nil {
		return fmt.Errorf("プロジェクト名: %w", err)
	}

	// 説明検証
	if err := ValidateDescription(config.Description); err != nil {
		return fmt.Errorf("説明: %w", err)
	}

	// テンプレート検証
	if config.Template != nil {
		if err := v.templateValidator.Validate(config.Template); err != nil {
			return fmt.Errorf("テンプレート: %w", err)
		}
	}

	// ローカルパス検証
	if err := v.validateLocalPath(config.LocalPath); err != nil {
		return fmt.Errorf("ローカルパス: %w", err)
	}

	return nil
}

// validateLocalPath はローカルパスを検証する
func (v *ConfigValidator) validateLocalPath(localPath string) error {
	if localPath == "" {
		return nil // 空の場合はOK（デフォルト値が使用される）
	}

	// 相対パスの検証（../ は危険）
	if strings.Contains(localPath, "..") {
		return fmt.Errorf("相対パス「..」は使用できません")
	}

	return nil
}

// ValidateConfigs は複数の設定を検証する
func (v *ConfigValidator) ValidateConfigs(configs []*models.ProjectConfig) error {
	if len(configs) == 0 {
		return fmt.Errorf("設定が指定されていません")
	}

	for i, config := range configs {
		if err := v.Validate(config); err != nil {
			return fmt.Errorf("設定[%d]: %w", i, err)
		}
	}

	return nil
}

// ValidateProjectConfig は ProjectConfig を検証する（エイリアス）
func (v *ConfigValidator) ValidateProjectConfig(config *models.ProjectConfig) error {
	return v.Validate(config)
}

// validateGitHubConstraints はGitHub特有の制約を検証する
func (v *ConfigValidator) validateGitHubConstraints(config *models.ProjectConfig) error {
	if !config.CreateGitHub {
		return nil // GitHubに作成しない場合はスキップ
	}

	// GitHub特有の制約をチェック
	if strings.Contains(config.Name, "..") {
		return fmt.Errorf("プロジェクト名に連続するドットは使用できません")
	}

	if len(config.Name) > 63 {
		return fmt.Errorf("GitHubリポジトリ名は63文字以下である必要があります")
	}

	return nil
}