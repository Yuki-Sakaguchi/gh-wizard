# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Actions workflow for automated cross-platform releases
- Comprehensive README with usage examples and feature descriptions
- CHANGELOG for tracking project changes

### Changed
- Improved documentation structure and clarity

### Technical
- Added `cli/gh-extension-precompile` action for automated binary releases
- Set up cross-platform binary generation for macOS, Linux, and Windows

## [0.1.0] - Development

### Added
- Initial gh-wizard GitHub CLI extension
- Interactive project creation wizard with create-next-app style UI
- Template repository discovery and selection
- Dynamic terminal width adaptation with CJK character support
- Cross-platform support (macOS, Linux, Windows)
- Command line options for non-interactive usage
- Project configuration with GitHub repository creation
- Smart template variable replacement
- Comprehensive test suite

### Features
- 🎨 **Beautiful create-next-app style UI** - Progressive disclosure interface
- 📦 **Template Repository Support** - Auto-discovery of template repositories
- 🖥️ **Terminal-Optimized Display** - Dynamic width adaptation
- ⚡ **Fast Installation** - Single command GitHub CLI installation
- 🌍 **Cross-Platform** - Works on all major platforms
- 🎯 **Zero Configuration** - Works with existing GitHub authentication

### UI Components
- Interactive template selection with descriptions
- Step-by-step project configuration
- Real-time terminal optimization
- CJK character width calculation
- Error handling with user-friendly messages

### Command Line Interface
- `gh wizard` - Interactive mode with beautiful UI
- `--classic-ui` - Traditional multi-question interface  
- `--name` - Project name specification
- `--template` - Template repository selection
- `--dry-run` - Preview mode without creation
- `--yes` - Skip confirmations for automation

### Technical Implementation
- Go 1.21+ with modern standard library usage
- GitHub CLI (`go-gh`) integration
- Survey library for interactive prompts
- runewidth for accurate character display width
- Comprehensive error handling and validation
- Extensive unit and integration test coverage