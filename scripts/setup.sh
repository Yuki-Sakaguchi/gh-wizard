#!/bin/bash
# gh-wizard setup script
# Auto-detects OS and installs dependencies

set -e

echo "🔮 Setting up gh-wizard..."

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# Detect OS
OS="unknown"
case "$(uname -s)" in
    Darwin*)  OS="macos" ;;
    Linux*)   OS="linux" ;;
    CYGWIN*)  OS="windows" ;;
    MINGW*)   OS="windows" ;;
    MSYS*)    OS="windows" ;;
esac

log_info "Detected OS: $OS"

# Check if gh CLI is installed
if ! command -v gh >/dev/null 2>&1; then
    log_error "GitHub CLI (gh) is required but not installed."
    log_info "Install it from: https://cli.github.com/"
    exit 1
fi
log_success "GitHub CLI found: $(gh --version | head -1)"

# Check if user is authenticated
if ! gh auth status >/dev/null 2>&1; then
    log_warning "Not logged in to GitHub CLI"
    log_info "Please run: gh auth login"
    read -p "Continue setup anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    log_success "GitHub CLI authenticated"
fi

# Install gh-wizard extension
log_info "Installing gh-wizard as GitHub CLI extension..."
if gh extension list | grep -q "gh-wizard"; then
    log_warning "gh-wizard extension already installed"
    read -p "Reinstall? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        gh extension remove gh-wizard
        gh extension install .
        log_success "gh-wizard extension reinstalled"
    fi
else
    gh extension install .
    log_success "gh-wizard extension installed"
fi

# Install lefthook (optional)
log_info "Setting up git hooks with lefthook (optional)..."

if command -v lefthook >/dev/null 2>&1; then
    log_success "lefthook already installed: $(lefthook version)"
else
    log_warning "lefthook not found"
    read -p "Install lefthook for automatic conventional commits? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        case $OS in
            "macos")
                if command -v brew >/dev/null 2>&1; then
                    log_info "Installing lefthook via Homebrew..."
                    brew install lefthook
                else
                    log_warning "Homebrew not found. Installing via Go..."
                    go install github.com/evilmartians/lefthook@latest
                fi
                ;;
            "linux")
                if command -v go >/dev/null 2>&1; then
                    log_info "Installing lefthook via Go..."
                    go install github.com/evilmartians/lefthook@latest
                else
                    log_warning "Go not found. Please install lefthook manually."
                    log_info "Visit: https://github.com/evilmartians/lefthook#installation"
                fi
                ;;
            "windows")
                if command -v scoop >/dev/null 2>&1; then
                    log_info "Installing lefthook via Scoop..."
                    scoop install lefthook
                else
                    log_warning "Scoop not found. Please install lefthook manually."
                    log_info "Visit: https://github.com/evilmartians/lefthook#installation"
                fi
                ;;
            *)
                log_warning "Unknown OS. Please install lefthook manually."
                log_info "Visit: https://github.com/evilmartians/lefthook#installation"
                ;;
        esac
    else
        log_info "Skipping lefthook installation"
    fi
fi

# Setup git hooks if lefthook is available
if command -v lefthook >/dev/null 2>&1; then
    log_info "Installing git hooks..."
    lefthook install
    log_success "Git hooks installed"
    
    # Test the setup
    log_info "Testing git hooks setup..."
    echo "test commit message" > /tmp/test-commit-msg
    if bash scripts/add-conventional-prefix.sh /tmp/test-commit-msg; then
        result=$(cat /tmp/test-commit-msg)
        log_success "Git hooks working correctly: '$result'"
        rm -f /tmp/test-commit-msg
    else
        log_warning "Git hooks test failed, but installation completed"
    fi
else
    log_warning "lefthook not available. Git hooks not installed."
    log_info "You can install lefthook later and run: lefthook install"
fi

# Run tests
log_info "Running tests..."
if go test ./...; then
    log_success "All tests passed"
else
    log_warning "Some tests failed, but setup completed"
fi

# Final success message
echo ""
log_success "🎉 Setup complete!"
echo ""
echo "📋 What was installed:"
echo "  ✅ gh-wizard GitHub CLI extension"
if command -v lefthook >/dev/null 2>&1; then
    echo "  ✅ lefthook git hooks (automatic conventional commits)"
else
    echo "  ⏳ lefthook git hooks (skipped - install manually if needed)"
fi
echo ""
echo "🚀 Try it out:"
echo "  gh wizard                    # Run the wizard"
echo "  gh wizard --help            # Show help"
echo ""
if command -v lefthook >/dev/null 2>&1; then
    echo "🪝 Git hooks enabled:"
    echo "  git checkout -b feature/test"
    echo "  git commit -m 'add feature'  # Becomes: 'feat: add feature'"
    echo ""
fi
echo "📚 More info: https://github.com/Yuki-Sakaguchi/gh-wizard"