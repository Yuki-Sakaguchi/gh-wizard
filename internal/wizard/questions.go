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

// formatDescriptionForTerminal formats template description with dynamic terminal width consideration
func formatDescriptionForTerminal(description string) string {
	if description == "" {
		return "No description available"
	}
	
	termWidth := getTerminalWidth()
	// Reserve space for: "? Please select a template: " prompt text and some padding
	// Survey typically uses about 30-40 characters for the prompt part
	reservedSpace := 40
	availableWidth := termWidth - reservedSpace
	
	// Ensure we have at least 20 characters for the description
	if availableWidth < 20 {
		availableWidth = 20
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
					return formatDescriptionForTerminal(qf.templates[index].Description)
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
