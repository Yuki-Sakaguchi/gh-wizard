package config

// GetDefault returns default settings
func GetDefault() *Config {
	return &Config{
		DefaultPrivate:   true,
		DefaultClone:     true,
		DefaultAddRemote: true,
		CacheTimeout:     30,
		Theme:            "default",
		RecentTemplates:  make([]string, 0),
	}
}

// GetConfigTemplate returns configuration file template YAML
func GetConfigTemplate() string {
	return `# gh-wizard configuration file
# Details: https://github.com/Yuki-Sakaguchi/gh-wizard

# Default settings
default_private: true        # Make repositories private by default
default_clone: true          # Clone locally after creation
default_add_readme: true     # Add README file

# Cache settings
cache_timeout_minutes: 30    # Template list cache timeout (minutes)

# UI settings
theme: "default"             # Theme: default, dark, light

# Recently used templates (auto-updated)
recent_templates: []
`
}
