package models

import "fmt"

type ProjectConfig struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Template     *Template `json:"template,omitempty"`
	CreateGitHub bool      `json:"create_github"`
	IsPrivate    bool      `json:"is_private"`
	LocalPath    string    `json:"local_path"`
}

// Validate checks the validity of configuration values
func (pc *ProjectConfig) Validate() error {
	if pc.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if len(pc.Name) > 100 {
		return fmt.Errorf("project name must be at most 100 characters")
	}

	if len(pc.Description) > 500 {
		return fmt.Errorf("description must be at most 500 characters")
	}

	return nil
}

// GetGitHubCreateCommand generates arguments for gh repo create command
func (pc *ProjectConfig) GetGitHubCreateCommand() []string {
	args := []string{"repo", "create", pc.Name}

	if pc.Template != nil {
		args = append(args, "--template", pc.Template.FullName)
	}

	if pc.Description != "" {
		args = append(args, "--description", pc.Description)
	}

	if pc.IsPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	args = append(args, "--clone")

	return args
}

// GetLocalCreatePath returns the local creation path
func (pc *ProjectConfig) GetLocalCreatePath() string {
	if pc.LocalPath != "" {
		return pc.LocalPath
	}
	return "./" + pc.Name
}

// HasTemplate returns whether a template is configured
func (pc *ProjectConfig) HasTemplate() bool {
	return pc.Template != nil
}

// GetDisplaySummary returns a display summary of configuration
func (pc *ProjectConfig) GetDisplaySummary() []string {
	summary := []string{
		fmt.Sprintf("📦 Project name: %s", pc.Name),
	}

	if pc.Description != "" {
		summary = append(summary, fmt.Sprintf("📄 Description: %s", pc.Description))
	}

	if pc.Template != nil {
		summary = append(summary, fmt.Sprintf("📚 Template: %s", pc.Template.FullName))
	} else {
		summary = append(summary, "📚 Template: None")
	}

	if pc.CreateGitHub {
		visibility := "🌐 Public"
		if pc.IsPrivate {
			visibility = "🔒 Private"
		}
		summary = append(summary, fmt.Sprintf("🐙 GitHub: Create (%s)", visibility))
	} else {
		summary = append(summary, "🐙 GitHub: Do not create")
	}

	summary = append(summary, fmt.Sprintf("📁 Local path: %s", pc.GetLocalCreatePath()))

	return summary
}
