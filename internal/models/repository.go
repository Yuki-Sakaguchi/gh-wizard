package models

import "errors"

// Repository represents GitHub repository creation settings
type RepositoryConfig struct {
	Name        string
	Description string
	IsPrivate   bool
	SholdClone  bool
	AddReadme   bool
}

// Validate checks the validity of configuration values
func (rc RepositoryConfig) Validate() error {
	if rc.Name == "" {
		return errors.New("repository name is required")
	}
	if len(rc.Name) > 100 {
		return errors.New("repository name must be at most 100 characters")
	}
	return nil
}

// GetGHCommand generates gh repo create command
func (rc RepositoryConfig) GetGHCommand(template *Template) []string {
	args := []string{"repo", "create", rc.Name}

	if template != nil {
		args = append(args, "--template", template.FullName)
	}

	if rc.Description != "" {
		args = append(args, "--description", rc.Description)
	}

	if rc.IsPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	if rc.SholdClone {
		args = append(args, "--clone")
	}

	if rc.AddReadme && template == nil {
		args = append(args, "--add-readme")
	}

	return args
}
