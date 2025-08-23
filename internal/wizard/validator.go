package wizard

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// ProjectNameValidator validates project names
type ProjectNameValidator struct {
	minLength       int
	maxLength       int
	validPattern    *regexp.Regexp
	invalidPatterns []*regexp.Regexp
}

// NewProjectNameValidator creates a new project name validator
func NewProjectNameValidator() *ProjectNameValidator {
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	invalidPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[^\w.-]`), // invalid characters
		regexp.MustCompile(`^\.`),     // starts with dot
		regexp.MustCompile(`\.$`),     // ends with dot
		regexp.MustCompile(`^-`),      // starts with hyphen
		regexp.MustCompile(`-$`),      // ends with hyphen
	}

	return &ProjectNameValidator{
		minLength:       1,
		maxLength:       100,
		validPattern:    validPattern,
		invalidPatterns: invalidPatterns,
	}
}

// Validate validates the project name
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

// validateBasicRules checks basic validation rules
func (v *ProjectNameValidator) validateBasicRules(name string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}

	if len(name) < v.minLength {
		return fmt.Errorf("project name must be at least %d characters long", v.minLength)
	}

	if len(name) > v.maxLength {
		return fmt.Errorf("project name must be at most %d characters long", v.maxLength)
	}

	// Check for whitespace only
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("project name is required")
	}

	// Invalid character check (only alphanumeric, hyphens, underscores, and dots allowed)
	if !isValidProjectName(name) {
		return fmt.Errorf("project name contains invalid characters. Only alphanumeric characters, hyphens, underscores, and dots are allowed")
	}

	return nil
}

// validateGitHubRules checks GitHub constraints
func (v *ProjectNameValidator) validateGitHubRules(name string) error {
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		if strings.HasPrefix(name, ".") {
			return fmt.Errorf("project names cannot start with a period")
		}
		if strings.HasSuffix(name, ".") {
			return fmt.Errorf("project names cannot end with a period")
		}
	}

	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("project names cannot start with a hyphen")
	}
	if strings.HasSuffix(name, "-") {
		return fmt.Errorf("project names cannot end with a hyphen")
	}
	if strings.HasPrefix(name, "_") {
		return fmt.Errorf("project names cannot start with an underscore")
	}
	if strings.HasSuffix(name, "_") {
		return fmt.Errorf("project names cannot end with an underscore")
	}

	// Check for consecutive periods or hyphens
	if strings.Contains(name, "..") {
		return fmt.Errorf("consecutive periods are not allowed")
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("consecutive hyphens are not allowed")
	}
	if strings.Contains(name, "__") {
		return fmt.Errorf("consecutive underscores are not allowed")
	}

	return nil
}

// validateReservedNames checks reserved names
func (v *ProjectNameValidator) validateReservedNames(name string) error {
	// System reserved names
	systemReserved := []string{".", "..", "CON", "PRN", "AUX", "NUL", "COM1", "LPT1"}
	upperName := strings.ToUpper(name)
	for _, reserved := range systemReserved {
		if upperName == reserved {
			return fmt.Errorf("'%s' is a reserved name and cannot be used", name)
		}
	}

	// Git-related reserved names
	gitReserved := []string{".git", ".github"}
	lowerNameForGit := strings.ToLower(name)
	for _, reserved := range gitReserved {
		if lowerNameForGit == reserved {
			return fmt.Errorf("'%s' is a reserved name and cannot be used", name)
		}
	}

	// General reserved names
	generalReserved := []string{"api", "www", "mail", "ftp", "admin", "root", "test", "debug"}
	lowerName := strings.ToLower(name)
	for _, reserved := range generalReserved {
		if lowerName == reserved {
			return fmt.Errorf("'%s' is a reserved name and cannot be used", name)
		}
	}

	return nil
}

// validateAdvancedRules checks advanced validation rules
func (v *ProjectNameValidator) validateAdvancedRules(name string) error {
	// Control character check
	if containsControlChars(name) {
		return fmt.Errorf("control characters are not allowed")
	}

	// Warning for all-digit names
	if isAllDigits(name) {
		return fmt.Errorf("all-numeric project names are not recommended")
	}

	// Too many special characters (40% or more are special characters)
	specialCount := countSpecialChars(name)
	if specialCount > 0 && float64(specialCount)/float64(len(name)) >= 0.4 {
		return fmt.Errorf("too many special characters")
	}

	return nil
}

// isValidProjectName checks if the project name is valid
func isValidProjectName(name string) bool {
	// Only alphanumeric characters, hyphens, underscores, and dots allowed
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return validPattern.MatchString(name)
}

// isAllDigits checks if the string contains only digits
func isAllDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// countSpecialChars counts the number of special characters
func countSpecialChars(s string) int {
	count := 0
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			count++
		}
	}
	return count
}

// containsControlChars checks if the string contains control characters
func containsControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// containsStrictControlChars checks for control characters in descriptions (allowing newline, tab, and CR)
func containsStrictControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return true
		}
	}
	return false
}

// ValidateDescription validates the description text
func ValidateDescription(description interface{}) error {
	// Type assertion
	desc, ok := description.(string)
	if !ok {
		return fmt.Errorf("invalid input type")
	}

	maxLength := 500
	maxLines := 5

	if len(desc) > maxLength {
		return fmt.Errorf("description must be at most %d characters long", maxLength)
	}

	// Line count check
	lines := strings.Split(desc, "\n")
	if len(lines) > maxLines {
		return fmt.Errorf("description must be at most %d lines", maxLines)
	}

	// Control character check (allowing newline, tab, and CR)
	if containsStrictControlChars(desc) {
		return fmt.Errorf("control characters are not allowed in description")
	}

	return nil
}

// TemplateValidator validates templates
type TemplateValidator struct {
	availableTemplates []models.Template
}

// NewTemplateValidator creates a new template validator
func NewTemplateValidator(templates ...[]models.Template) *TemplateValidator {
	validator := &TemplateValidator{}
	if len(templates) > 0 {
		validator.availableTemplates = templates[0]
	}
	return validator
}

// Validate validates the template
func (v *TemplateValidator) Validate(template *models.Template) error {
	if template == nil {
		return fmt.Errorf("template is nil")
	}

	if template.FullName == "" {
		return fmt.Errorf("template FullName is required")
	}

	if template.Name == "" {
		return fmt.Errorf("template Name is required")
	}

	if !template.IsTemplate {
		return fmt.Errorf("specified repository is not a template")
	}

	return nil
}

// ValidateTemplates validates multiple templates
func (v *TemplateValidator) ValidateTemplates(templates []models.Template) error {
	if len(templates) == 0 {
		return fmt.Errorf("no templates specified")
	}

	for i, template := range templates {
		if err := v.Validate(&template); err != nil {
			return fmt.Errorf("template[%d]: %w", i, err)
		}
	}

	return nil
}

// ValidateTemplateSelection validates template selection
func (v *TemplateValidator) ValidateTemplateSelection(selection string) error {
	if selection == "" {
		return fmt.Errorf("please select a template")
	}

	if selection == "No template" {
		return nil // Valid selection
	}

	// Search available templates (also search by display format)
	for _, template := range v.availableTemplates {
		// Compare by name, FullName, or display format
		displayName := template.GetDisplayName()
		if template.Name == selection || template.FullName == selection || displayName == selection {
			if !template.IsTemplate {
				return fmt.Errorf("'%s' is not configured as a template", selection)
			}
			return nil
		}
	}

	return fmt.Errorf("template '%s' is not available", selection)
}

// GetSurveyValidator returns a validation function for Survey
func (v *TemplateValidator) GetSurveyValidator() func(interface{}) error {
	return func(ans interface{}) error {
		selection, ok := ans.(string)
		if !ok {
			return fmt.Errorf("invalid input type")
		}
		return v.ValidateTemplateSelection(selection)
	}
}

// ConfigValidator validates configuration
type ConfigValidator struct {
	projectValidator  *ProjectNameValidator
	templateValidator *TemplateValidator
}

// NewConfigValidator creates a new configuration validator
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

// Validate validates ProjectConfig
func (v *ConfigValidator) Validate(config *models.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Project name validation
	if err := v.projectValidator.Validate(config.Name); err != nil {
		return fmt.Errorf("project name: %w", err)
	}

	// Description validation
	if err := ValidateDescription(config.Description); err != nil {
		return fmt.Errorf("description: %w", err)
	}

	// Template validation
	if config.Template != nil {
		if err := v.templateValidator.Validate(config.Template); err != nil {
			return fmt.Errorf("template: %w", err)
		}
	}

	// Local path validation
	if err := v.validateLocalPath(config.LocalPath); err != nil {
		return fmt.Errorf("local path: %w", err)
	}

	return nil
}

// validateLocalPath validates the local path
func (v *ConfigValidator) validateLocalPath(localPath string) error {
	if localPath == "" {
		return nil // Empty is OK (default value will be used)
	}

	// Relative path validation (../ is dangerous)
	if strings.Contains(localPath, "..") {
		return fmt.Errorf("relative path '..' is not allowed")
	}

	return nil
}

// ValidateConfigs validates multiple configurations
func (v *ConfigValidator) ValidateConfigs(configs []*models.ProjectConfig) error {
	if len(configs) == 0 {
		return fmt.Errorf("no configurations specified")
	}

	for i, config := range configs {
		if err := v.Validate(config); err != nil {
			return fmt.Errorf("configuration[%d]: %w", i, err)
		}
	}

	return nil
}

// ValidateProjectConfig validates ProjectConfig (alias)
func (v *ConfigValidator) ValidateProjectConfig(config *models.ProjectConfig) error {
	return v.Validate(config)
}

// validateGitHubConstraints validates GitHub-specific constraints
func (v *ConfigValidator) validateGitHubConstraints(config *models.ProjectConfig) error {
	if !config.CreateGitHub {
		return nil // Skip if not creating on GitHub
	}

	// Check GitHub-specific constraints
	if strings.Contains(config.Name, "..") {
		return fmt.Errorf("consecutive dots in project name are not allowed")
	}

	if len(config.Name) > 63 {
		return fmt.Errorf("GitHub repository name must be 63 characters or less")
	}

	return nil
}
