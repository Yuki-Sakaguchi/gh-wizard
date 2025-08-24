# Contributing Guide

Thank you for your interest in contributing to gh-wizard! 🎉

This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Conventional Commits](#conventional-commits)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## 🤝 Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## 🚀 Getting Started

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [GitHub CLI](https://cli.github.com/) installed and authenticated
- Git installed

### Setup Development Environment

1. **Fork and Clone**
   ```bash
   gh repo fork Yuki-Sakaguchi/gh-wizard --clone
   cd gh-wizard
   ```

2. **Install Dependencies**
   ```bash
   go mod tidy
   ```

3. **Build and Test**
   ```bash
   go build
   go test ./...
   ```

4. **Install Locally**
   ```bash
   gh extension install .
   ```

5. **Verify Installation**
   ```bash
   gh wizard --help
   ```

## 🔄 Development Workflow

### Branch Strategy

1. Create a feature branch from `main`:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the guidelines below

3. Test your changes thoroughly

4. Create a Pull Request

### Branch Naming Convention

- `feature/` - New features
- `fix/` - Bug fixes  
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements
- `chore/` - Maintenance tasks

**Examples:**
- `feature/template-filtering`
- `fix/terminal-width-calculation`
- `docs/update-readme`

## 📝 Conventional Commits

**This project uses [Conventional Commits](https://www.conventionalcommits.org/) for automated versioning and changelog generation.**

### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | New feature | Minor (1.0.0 → 1.1.0) |
| `fix` | Bug fix | Patch (1.0.0 → 1.0.1) |
| `docs` | Documentation changes | Patch |
| `style` | Code style changes | Patch |
| `refactor` | Code refactoring | Patch |
| `perf` | Performance improvements | Patch |
| `test` | Test additions/updates | Patch |
| `build` | Build system changes | Patch |
| `ci` | CI/CD changes | Patch |
| `chore` | Maintenance tasks | No bump |
| `revert` | Revert changes | Patch |

### Breaking Changes

For breaking changes, add `!` after the type or include `BREAKING CHANGE:` in the footer:

```bash
feat!: change command structure for direct execution
# or
feat: change command structure

BREAKING CHANGE: Users now run `gh wizard` instead of `gh wizard wizard`
```

This will trigger a **Major** version bump (1.0.0 → 2.0.0).

### Examples

```bash
# New feature
feat: add template filtering by language

# Bug fix
fix: resolve terminal width calculation for CJK characters

# Documentation
docs: update installation instructions

# Breaking change
feat!: redesign CLI interface

# With scope
feat(ui): implement create-next-app style interface
fix(auth): handle GitHub authentication timeout
```

### Commit Message Guidelines

- Use the imperative mood ("add" not "added" or "adds")
- Keep the first line under 50 characters
- Separate subject from body with a blank line
- Use the body to explain what and why vs. how
- Reference issues and PRs when appropriate

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/wizard/...

# Run tests in verbose mode
go test -v ./...
```

### Testing Guidelines

- Write tests for all new functionality
- Maintain test coverage above 80%
- Use table-driven tests for multiple scenarios
- Mock external dependencies (GitHub API, filesystem operations)
- Test both success and error cases

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 🔍 Pull Request Process

### Before Creating a PR

- [ ] Code follows the project style guidelines
- [ ] All tests pass locally
- [ ] New code has appropriate test coverage
- [ ] Documentation is updated if needed
- [ ] Commit messages follow Conventional Commits format

### PR Template

When creating a PR, include:

1. **Description**: What does this PR do?
2. **Related Issue**: Link to related issue(s)
3. **Testing**: How was this tested?
4. **Breaking Changes**: Any breaking changes?
5. **Screenshots**: For UI changes

### PR Title Format

Use the same format as commit messages:

```
feat: add template filtering functionality
fix: resolve terminal width calculation bug
docs: update contributing guidelines
```

### Review Process

1. **Automated Checks**: Ensure all CI checks pass
2. **Code Review**: At least one maintainer review required
3. **Testing**: Manual testing if needed
4. **Documentation**: Verify documentation updates
5. **Merge**: Squash and merge with conventional commit message

## 🚀 Release Process

This project uses **automated releases** powered by [Release Please](https://github.com/googleapis/release-please).

### How Releases Work

1. **Commit to Main**: When PRs are merged to main with conventional commits
2. **Release PR**: Release Please automatically creates a release PR with:
   - Updated version numbers
   - Updated CHANGELOG.md
   - Release notes
3. **Manual Review**: Maintainers review and merge the release PR
4. **Automated Release**: Upon merge:
   - Git tag is created
   - GitHub release is published  
   - Cross-platform binaries are built and attached
   - Extension is available for installation/upgrade

### Version Calculation

Based on conventional commits since the last release:

- **Major** (1.0.0 → 2.0.0): Breaking changes (`feat!`, `BREAKING CHANGE:`)
- **Minor** (1.0.0 → 1.1.0): New features (`feat:`)
- **Patch** (1.0.0 → 1.0.1): Bug fixes, docs, etc. (`fix:`, `docs:`, etc.)

### Manual Releases

For emergency releases or special cases:

```bash
git tag v1.2.3
git push origin v1.2.3
```

This will trigger the existing release workflow directly.

## 🎯 Project Structure

```
gh-wizard/
├── cmd/                    # Command-line interface
├── internal/
│   ├── config/            # Configuration management
│   ├── github/            # GitHub API client
│   ├── models/            # Data models
│   ├── utils/             # Utility functions
│   └── wizard/            # Core wizard logic
├── .github/
│   └── workflows/         # CI/CD workflows
├── docs/                  # Additional documentation
└── README.md
```

## 💡 Tips for Contributors

### Development Tips

- Use `go run main.go` for quick testing during development
- Use `--dry-run` flag to test without creating actual repositories
- Test with different terminal sizes for UI responsiveness
- Test with both English and CJK characters

### Debugging

- Use `fmt.Printf` for temporary debugging (remove before committing)
- Use `go test -v` for verbose test output
- Check GitHub CLI authentication with `gh auth status`

### Common Issues

- **Import path**: Always use `github.com/Yuki-Sakaguchi/gh-wizard/internal/...`
- **Cross-platform**: Test file paths work on Windows, macOS, and Linux
- **Character width**: Consider CJK character display width in UI calculations

## 📞 Getting Help

- 💬 [Discussions](https://github.com/Yuki-Sakaguchi/gh-wizard/discussions) - Ask questions
- 🐞 [Issues](https://github.com/Yuki-Sakaguchi/gh-wizard/issues) - Report bugs or request features
- 📧 Contact maintainers directly for sensitive issues

## 🙏 Recognition

Contributors will be recognized in:

- CHANGELOG.md for their contributions
- GitHub releases attribution
- README contributors section (coming soon)

Thank you for contributing to gh-wizard! 🎉