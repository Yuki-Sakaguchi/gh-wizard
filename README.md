# gh-wizard 🔮
A magical GitHub CLI extension that creates repositories using template repositories with an intuitive, create-next-app style interface.

[![Release](https://img.shields.io/github/v/release/Yuki-Sakaguchi/gh-wizard)](https://github.com/Yuki-Sakaguchi/gh-wizard/releases)

![result](https://github.com/user-attachments/assets/fb5e10b0-8390-42f7-995d-db69b61b9373)

## ✨ Features

- 🎨 **Beautiful create-next-app style UI** - Interactive, step-by-step project creation
- 📦 **Template Repository Support** - Automatically discovers your template repositories
- 🖥️ **Terminal-Optimized Display** - Dynamically adapts to your terminal width with CJK character support
- ⚡ **Fast Installation** - Single command installation via GitHub CLI
- 🌍 **Cross-Platform** - Works on macOS, Linux, and Windows
- 🎯 **Zero Configuration** - Works out of the box with your GitHub account

## 🚀 Quick Start

### Installation

#### Option 1: Simple Installation
```bash
gh extension install Yuki-Sakaguchi/gh-wizard
```

#### Option 2: Complete Setup (Recommended)
Auto-installs extension + optional git hooks for automatic conventional commits:

```bash
# Clone and run setup script
git clone https://github.com/Yuki-Sakaguchi/gh-wizard.git
cd gh-wizard
./scripts/setup.sh
```

Or using Make:
```bash
make setup  # Install extension + git hooks
make dev    # Development setup with tests
```

## 🎯 Usage Examples

### Interactive Mode (Default)

The default mode provides a beautiful, create-next-app inspired interface:

```bash
$ gh wizard

🔍 Fetching your template repositories...
✅ Found 3 template repositories

? Please select a template: [Use arrows to move, type to filter]
> nextjs-starter - Next.js project starter kit with TypeScript
  node-ts - TypeScript Node.js development environment
  react-component - Reusable React component template

✓ Please select a template: … nextjs-starter
✓ What is your project named? … my-awesome-app
✓ Enter project description (optional): … My awesome new project
✓ Create repository on GitHub? … Yes
✓ Create as private repository? … No

📝 Configuration Review
✓ Project Name: my-awesome-app
✓ Description:  My awesome new project
✓ Template:     nextjs-starter (15⭐)
✓ Local Path:   ./my-awesome-app
✓ Private:      False

? Create project with this configuration? (y/N) 
```

## 🔧 How It Works

1. **Template Discovery**: Automatically finds repositories marked as "Template repository" in your GitHub account
2. **Interactive Selection**: Choose from your templates with rich descriptions and metadata
3. **Project Configuration**: Set up project name, description, and GitHub repository options
4. **Smart Creation**: 
   - Clones template repository
   - Creates local project directory
   - Initializes new Git repository
   - Optionally creates GitHub repository
   - Handles file templating and variable replacement

## 📋 Prerequisites

- [GitHub CLI](https://cli.github.com/) installed and authenticated
- Git installed
- Go 1.21+ (for development)

## 🛠️ Development

### Local Development

```bash
# Clone the repository
git clone https://github.com/Yuki-Sakaguchi/gh-wizard.git
cd gh-wizard

# Install dependencies
go mod tidy

# Build the project
go build

# Run tests
go test ./...

# Install locally for testing
gh extension install .
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/wizard/...
```

## 🪝 Git Hooks (Optional)

This project supports automatic conventional commit prefix generation using [lefthook](https://github.com/evilmartians/lefthook).

### What is Lefthook?

Lefthook is a fast, cross-platform Git hooks manager that helps automate development workflows. In gh-wizard, it automatically adds conventional commit prefixes based on your branch name.

### Quick Setup

1. **Install lefthook**:
   ```bash
   # macOS (Homebrew)
   brew install lefthook
   
   # Linux/macOS (Go install) 
   go install github.com/evilmartians/lefthook@latest
   
   # Windows (Scoop)
   scoop install lefthook
   ```

2. **Install hooks**:
   ```bash
   lefthook install
   ```

3. **Start using automatic prefixes**:
   ```bash
   git checkout -b feature/awesome-feature
   git commit -m "add awesome feature"  # Becomes: "feat: add awesome feature"
   ```

### Branch Name Mapping

| Branch Pattern | Prefix | Example |
|----------------|--------|---------|
| `feature/*`, `feat/*` | `feat:` | `feature/user-auth` → `feat: your message` |
| `fix/*`, `bugfix/*` | `fix:` | `fix/login-bug` → `fix: your message` |
| `docs/*` | `docs:` | `docs/update-api` → `docs: your message` |
| `refactor/*` | `refactor:` | `refactor/clean-code` → `refactor: your message` |
| `test/*` | `test:` | `test/add-validation` → `test: your message` |
| `chore/*` | `chore:` | `chore/update-deps` → `chore: your message` |
| `perf/*` | `perf:` | `perf/optimize-query` → `perf: your message` |
| `ci/*` | `ci:` | `ci/update-workflow` → `ci: your message` |
| `build/*` | `build:` | `build/webpack-config` → `build: your message` |
| `style/*` | `style:` | `style/format-code` → `style: your message` |

### Smart Behavior

- **Already prefixed commits**: No changes made
- **Main branches** (`main`, `master`, `develop`): No prefix added
- **Merge/revert commits**: Automatically skipped
- **Unknown branch patterns**: No prefix added (manual control)

### Debugging

Enable debug mode to see what's happening:

```bash
DEBUG_LEFTHOOK=1 git commit -m "test message"
```

### Configuration

The configuration is stored in `.lefthook.yml` and shared across the team. The setup includes:

- **commit-msg hook**: Automatic prefix generation
- **pre-commit hooks**: Go formatting, tests, and module tidying
- **Cross-platform compatibility**: Works on macOS, Linux, and Windows

### Manual Override

You can always use manual prefixes - lefthook won't modify them:

```bash
git commit -m "feat: my custom message"  # No changes made
git commit -m "fix(auth): specific fix"  # No changes made
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🐛 Issues & Support

- 🐞 [Report bugs](https://github.com/Yuki-Sakaguchi/gh-wizard/issues/new?template=bug_report.md)
- 💡 [Request features](https://github.com/Yuki-Sakaguchi/gh-wizard/issues/new?template=feature_request.md)
- 💬 [Ask questions](https://github.com/Yuki-Sakaguchi/gh-wizard/discussions)

---

<p align="center">
  Made with ❤️ for the GitHub community
</p>
