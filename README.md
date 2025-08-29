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
