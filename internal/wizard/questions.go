package wizard

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// Answers stores Survey responses
type Answers struct {
	Template     string `survey:"template"`
	ProjectName  string `survey:"projectName"`
	Description  string `survey:"description"`
	CreateGitHub bool   `survey:"createGitHub"`
	IsPrivate    bool   `survey:"isPrivate"`
}

// SurveyExecutor interface for survey execution
type SurveyExecutor interface {
	Ask(questions []*survey.Question, response interface{}) error
}

// DefaultSurveyExecutor is the default survey executor
type DefaultSurveyExecutor struct{}

func (d *DefaultSurveyExecutor) Ask(questions []*survey.Question, response interface{}) error {
	return survey.Ask(questions, response)
}

// QuestionFlow manages question flow
type QuestionFlow struct {
	templates      []models.Template
	answers        *Answers
	surveyExecutor SurveyExecutor
}

// NewQuestionFlow creates a new question flow
func NewQuestionFlow(templates []models.Template) *QuestionFlow {
	return &QuestionFlow{
		templates:      templates,
		answers:        &Answers{},
		surveyExecutor: &DefaultSurveyExecutor{},
	}
}

// formatTemplateOption creates template option display format
func formatTemplateOption(template models.Template) string {
	stars := ""
	if template.Stars > 0 {
		stars = fmt.Sprintf(" (⭐ %d)", template.Stars)
	}

	language := ""
	if template.Language != "" {
		language = fmt.Sprintf(" [%s]", template.Language)
	}

	return fmt.Sprintf("%s%s%s", template.Name, stars, language)
}

// getTerminalWidth gets the current terminal width
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Fallback to 80 columns if we can't detect terminal size
		return 80
	}
	return width
}

// getStringDisplayWidth calculates the display width of a string considering multi-byte characters
func getStringDisplayWidth(s string) int {
	return runewidth.StringWidth(s)
}

// calculatePromptWidth calculates the actual width used by the survey prompt
func calculatePromptWidth() int {
	// "? Please select a template: " + some margin for selection indicators
	promptText := "? Please select a template: "
	baseWidth := getStringDisplayWidth(promptText)
	// Add margin for selection indicators and padding
	margin := 5
	return baseWidth + margin
}

// formatDescriptionForTerminal formats template description with dynamic terminal width consideration
func formatDescriptionForTerminal(description string) string {
	if description == "" {
		return "No description available"
	}
	
	termWidth := getTerminalWidth()
	// Calculate actual prompt width instead of using fixed 40
	promptWidth := calculatePromptWidth()
	
	// More aggressive optimization: reduce template option buffer based on actual content
	// Most template names are shorter than expected
	baseTemplateBuffer := 20  // Reduced from 35
	templateOptionBuffer := baseTemplateBuffer
	
	// For mixed language content, be even more generous with space allocation
	if containsCJKCharacters(description) {
		// Even more aggressive for CJK content - they need more space per character
		templateOptionBuffer = 15  // Further reduced for CJK
	}
	
	availableWidth := termWidth - promptWidth - templateOptionBuffer
	
	// Ensure we have reasonable minimum space for the description
	minWidth := 25
	if containsCJKCharacters(description) {
		minWidth = 35  // CJK characters need more minimum space
	}
	
	if availableWidth < minWidth {
		availableWidth = minWidth
	}
	
	currentWidth := getStringDisplayWidth(description)
	if currentWidth <= availableWidth {
		return description
	}
	
	// Truncate by runes, not bytes, considering display width
	ellipsis := "..."
	ellipsisWidth := getStringDisplayWidth(ellipsis)
	targetWidth := availableWidth - ellipsisWidth
	
	if targetWidth <= 0 {
		return ellipsis
	}
	
	runes := []rune(description)
	result := ""
	currentDisplayWidth := 0
	
	for _, r := range runes {
		runeWidth := runewidth.RuneWidth(r)
		if currentDisplayWidth+runeWidth > targetWidth {
			break
		}
		result += string(r)
		currentDisplayWidth += runeWidth
	}
	
	return result + ellipsis
}


// formatDescriptionForTerminalWithTemplates formats description with actual template display widths
func formatDescriptionForTerminalWithTemplates(description string, templates []models.Template) string {
	if description == "" {
		return "No description available"
	}
	
	termWidth := getTerminalWidth()
	promptWidth := calculatePromptWidth()
	
	// Calculate the actual maximum width of formatted template display
	// Template is displayed as: "nextjs-starter - Description here"
	maxTemplateDisplayWidth := 0
	for _, template := range templates {
		formattedOption := formatTemplateOption(template)
		displayWidth := getStringDisplayWidth(formattedOption)
		if displayWidth > maxTemplateDisplayWidth {
			maxTemplateDisplayWidth = displayWidth
		}
	}
	
	// Add buffer for the separator " - " and some spacing
	templateDisplayBuffer := maxTemplateDisplayWidth + 3  // 3 for " - "
	
	availableWidth := termWidth - promptWidth - templateDisplayBuffer
	
	// Ensure we have reasonable minimum space for the description
	minWidth := 20
	if containsCJKCharacters(description) {
		minWidth = 25
	}
	
	if availableWidth < minWidth {
		availableWidth = minWidth
	}
	
	// Debug output
	if os.Getenv("DEBUG_DESCRIPTION") == "1" {
		currentWidth := getStringDisplayWidth(description)
		isCJK := containsCJKCharacters(description)
		
		fmt.Printf("\n[DEBUG] Terminal: %d, Prompt: %d, MaxTemplateDisplay: %d, Buffer: %d, Available: %d, Description: %d, CJK: %t\n", 
			termWidth, promptWidth, maxTemplateDisplayWidth, templateDisplayBuffer, availableWidth, currentWidth, isCJK)
		fmt.Printf("[DEBUG] Description: %s\n", description)
	}
	
	currentWidth := getStringDisplayWidth(description)
	if currentWidth <= availableWidth {
		return description
	}
	
	// Truncate by runes, not bytes, considering display width
	ellipsis := "..."
	ellipsisWidth := getStringDisplayWidth(ellipsis)
	targetWidth := availableWidth - ellipsisWidth
	
	if targetWidth <= 0 {
		return ellipsis
	}
	
	runes := []rune(description)
	result := ""
	currentDisplayWidth := 0
	
	for _, r := range runes {
		runeWidth := runewidth.RuneWidth(r)
		if currentDisplayWidth+runeWidth > targetWidth {
			break
		}
		result += string(r)
		currentDisplayWidth += runeWidth
	}
	
	return result + ellipsis
}

// containsCJKCharacters checks if the string contains Chinese, Japanese, or Korean characters
func containsCJKCharacters(s string) bool {
	for _, r := range s {
		// Check for CJK Unified Ideographs, Hiragana, Katakana, and other CJK ranges
		if (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		   (r >= 0x3040 && r <= 0x309F) || // Hiragana
		   (r >= 0x30A0 && r <= 0x30FF) || // Katakana
		   (r >= 0xAC00 && r <= 0xD7AF) || // Hangul (Korean)
		   (r >= 0x3100 && r <= 0x312F) || // Bopomofo
		   (r >= 0x31A0 && r <= 0x31BF) {  // Bopomofo Extended
			return true
		}
	}
	return false
}

// clearPreviousLines clears the previous lines from terminal
func clearPreviousLines(lineCount int) {
	for i := 0; i < lineCount; i++ {
		fmt.Print("\033[1A") // Move up one line
		fmt.Print("\033[2K") // Clear entire line
	}
}


// formatDescription formats template description with truncation (legacy function for tests)
func formatDescription(description string, maxLength int) string {
	if description == "" {
		return "No description available"
	}
	
	currentWidth := getStringDisplayWidth(description)
	if currentWidth <= maxLength {
		return description
	}
	
	ellipsis := "..."
	ellipsisWidth := getStringDisplayWidth(ellipsis)
	targetWidth := maxLength - ellipsisWidth
	
	if targetWidth <= 0 {
		return ellipsis
	}
	
	runes := []rune(description)
	result := ""
	currentDisplayWidth := 0
	
	for _, r := range runes {
		runeWidth := runewidth.RuneWidth(r)
		if currentDisplayWidth+runeWidth > targetWidth {
			break
		}
		result += string(r)
		currentDisplayWidth += runeWidth
	}
	
	return result + ellipsis
}

// findSelectedTemplate retrieves the selected template
func (qf *QuestionFlow) findSelectedTemplate() *models.Template {
	// Return nil if no template is selected
	if qf.answers.Template == "" {
		return nil
	}

	for _, template := range qf.templates {
		if formatTemplateOption(template) == qf.answers.Template {
			return &template
		}
	}
	return nil
}

// GetProjectConfig generates ProjectConfig from answers
func (qf *QuestionFlow) GetProjectConfig() *models.ProjectConfig {
	template := qf.findSelectedTemplate()

	return &models.ProjectConfig{
		Name:         qf.answers.ProjectName,
		Description:  qf.answers.Description,
		Template:     template,
		CreateGitHub: qf.answers.CreateGitHub,
		IsPrivate:    qf.answers.IsPrivate,
		LocalPath:    "./" + qf.answers.ProjectName,
	}
}

// CreateQuestions generates questions based on template information
func (qf *QuestionFlow) CreateQuestions() []*survey.Question {
	// Generate template options (exclude no-template option)
	if len(qf.templates) == 0 {
		// Skip if no templates available
		return []*survey.Question{}
	}

	templateOptions := make([]string, len(qf.templates))
	for i, t := range qf.templates {
		templateOptions[i] = formatTemplateOption(t)
	}

	questions := []*survey.Question{
		{
			Name: "template",
			Prompt: &survey.Select{
				Message: "Please select a template:",
				Options: templateOptions,
				Description: func(value string, index int) string {
					return formatDescriptionForTerminalWithTemplates(qf.templates[index].Description, qf.templates)
				},
			},
			Validate: survey.Required,
		},
	}

	return questions
}

// CreateConditionalQuestions generates conditional questions
func (qf *QuestionFlow) CreateConditionalQuestions() []*survey.Question {
	var questions []*survey.Question

	// Questions displayed only when creating GitHub repository
	if qf.answers.CreateGitHub {
		questions = append(questions, &survey.Question{
			Name: "isPrivate",
			Prompt: &survey.Confirm{
				Message: "Create as private repository?",
				Default: true,
				Help:    "Private: Only you can access / Public: Anyone can access",
			},
		})
	}

	return questions
}

// CreateBasicQuestions creates questions about project basic information
func (qf *QuestionFlow) CreateBasicQuestions() []*survey.Question {
	return []*survey.Question{
		{
			Name: "projectName",
			Prompt: &survey.Input{
				Message: "Enter project name:",
				Help:    "Alphanumeric characters, hyphens, and underscores are allowed",
			},
			Validate: survey.Required,
		},
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "Enter project description (optional):",
				Help:    "Brief description of the project",
			},
		},
		{
			Name: "createGitHub",
			Prompt: &survey.Confirm{
				Message: "Create repository on GitHub?",
				Default: false,
				Help:    "If No, project will be created locally only",
			},
		},
	}
}

// ExecuteCreateNextAppStyle runs the question flow with create-next-app style UI
func (qf *QuestionFlow) ExecuteCreateNextAppStyle() (*models.ProjectConfig, error) {
	fmt.Println()

	// 1. Template selection (if templates are available)
	if len(qf.templates) > 0 {
		templateQuestion := qf.CreateQuestions()[0]
		
		var templateAnswer struct {
			Template string `survey:"template"`
		}
		
		err := survey.AskOne(templateQuestion.Prompt, &templateAnswer.Template, survey.WithValidator(templateQuestion.Validate))
		if err != nil {
			return nil, fmt.Errorf("failed to execute template selection: %w", err)
		}
		qf.answers.Template = templateAnswer.Template
		
		// Clear the question lines (typically 2-3 lines for select prompt)
		clearPreviousLines(3)
		
		// Show completed template selection
		selectedTemplate := qf.findSelectedTemplate()
		templateName := "None"
		if selectedTemplate != nil {
			templateName = selectedTemplate.Name
		}
		fmt.Printf("✓ Which template would you like to use? … %s\n", templateName)
	}

	// 2. Project name
	projectNamePrompt := &survey.Input{
		Message: "What is your project named?",
		Help:    "Alphanumeric characters, hyphens, and underscores are allowed",
	}
	
	err := survey.AskOne(projectNamePrompt, &qf.answers.ProjectName, survey.WithValidator(survey.Required))
	if err != nil {
		return nil, fmt.Errorf("failed to get project name: %w", err)
	}
	
	// Clear the input question line
	clearPreviousLines(1)
	fmt.Printf("✓ What is your project named? … %s\n", qf.answers.ProjectName)

	// 3. Description (optional)
	descPrompt := &survey.Input{
		Message: "Enter project description (optional):",
		Help:    "Brief description of the project",
	}
	
	err = survey.AskOne(descPrompt, &qf.answers.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to get project description: %w", err)
	}
	
	// Clear the input question line
	clearPreviousLines(1)
	if qf.answers.Description != "" {
		fmt.Printf("✓ Enter project description (optional): … %s\n", qf.answers.Description)
	} else {
		fmt.Printf("✓ Enter project description (optional): … (skipped)\n")
	}

	// 4. GitHub repository creation
	githubPrompt := &survey.Confirm{
		Message: "Create repository on GitHub?",
		Default: false,
		Help:    "If No, project will be created locally only",
	}
	
	err = survey.AskOne(githubPrompt, &qf.answers.CreateGitHub)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub preference: %w", err)
	}
	
	// Clear the confirmation question line
	clearPreviousLines(1)
	githubAnswer := "No"
	if qf.answers.CreateGitHub {
		githubAnswer = "Yes"
	}
	fmt.Printf("✓ Create repository on GitHub? … %s\n", githubAnswer)

	// 5. Private repository (if creating GitHub repo)
	if qf.answers.CreateGitHub {
		privatePrompt := &survey.Confirm{
			Message: "Create as private repository?",
			Default: true,
			Help:    "Private: Only you can access / Public: Anyone can access",
		}
		
		err = survey.AskOne(privatePrompt, &qf.answers.IsPrivate)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository privacy preference: %w", err)
		}
		
		// Clear the confirmation question line
		clearPreviousLines(1)
		privateAnswer := "Public"
		if qf.answers.IsPrivate {
			privateAnswer = "Private"
		}
		fmt.Printf("✓ Create as private repository? … %s\n", privateAnswer)
	}

	fmt.Println()

	// Convert answers to ProjectConfig
	config := &models.ProjectConfig{
		Name:         qf.answers.ProjectName,
		Description:  qf.answers.Description,
		CreateGitHub: qf.answers.CreateGitHub,
		IsPrivate:    qf.answers.IsPrivate,
	}

	// Search for template
	for _, template := range qf.templates {
		if formatTemplateOption(template) == qf.answers.Template {
			config.Template = &template
			break
		}
	}

	return config, nil
}

// Execute runs the question flow and returns ProjectConfig
func (qf *QuestionFlow) Execute() (*models.ProjectConfig, error) {
	// Execute template selection questions (only if templates are available)
	questions := qf.CreateQuestions()
	if len(questions) > 0 {
		err := qf.surveyExecutor.Ask(questions, qf.answers)
		if err != nil {
			return nil, fmt.Errorf("failed to execute template selection: %w", err)
		}
	}

	// Project basic information questions
	basicQuestions := qf.CreateBasicQuestions()
	err := qf.surveyExecutor.Ask(basicQuestions, qf.answers)
	if err != nil {
		return nil, fmt.Errorf("failed to execute basic questions: %w", err)
	}

	// Execute conditional questions
	conditionalQuestions := qf.CreateConditionalQuestions()
	if len(conditionalQuestions) > 0 {
		err = qf.surveyExecutor.Ask(conditionalQuestions, qf.answers)
		if err != nil {
			return nil, fmt.Errorf("failed to execute conditional questions: %w", err)
		}
	}

	// Convert answers to ProjectConfig
	config := &models.ProjectConfig{
		Name:         qf.answers.ProjectName,
		Description:  qf.answers.Description,
		CreateGitHub: qf.answers.CreateGitHub,
		IsPrivate:    qf.answers.IsPrivate,
	}

	// Search for template
	for _, template := range qf.templates {
		if formatTemplateOption(template) == qf.answers.Template {
			config.Template = &template
			break
		}
	}

	return config, nil
}
