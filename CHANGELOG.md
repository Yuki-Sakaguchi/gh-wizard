# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0](https://github.com/Yuki-Sakaguchi/gh-wizard/compare/v1.3.0...v1.4.0) (2025-08-29)


### 🚀 Features

* add spinner loading indicator for template repository fetching ([83f1feb](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/83f1febfed25e8afdfbf3831204dcd9ef6bf2fdd)), closes [#49](https://github.com/Yuki-Sakaguchi/gh-wizard/issues/49)
* コマンド実行時のローディング表示を追加 ([#49](https://github.com/Yuki-Sakaguchi/gh-wizard/issues/49)) ([b0b2619](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/b0b2619b9467e4841aebad05daa0f703f0f14607))


### 🐛 Bug Fixes

* remove unnecessary startup message ([67a10f1](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/67a10f1c19a9cc39bf400e0d8aef642f7ea50b6c))
* resolve Release Please workflow trigger limitations ([e43270a](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/e43270a96a242242d82d91278b28e99739e2b594))
* resolve release-please workflow token error ([524302c](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/524302c041f8b9bed63e6dfd68d9c2eb1d075491))

## [1.3.0](https://github.com/Yuki-Sakaguchi/gh-wizard/compare/v1.2.1...v1.3.0) (2025-08-24)


### 🚀 Features

* add comprehensive build and release workflow ([27703b5](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/27703b5f24fb653d0234a096577380d2ee86e091))


### 📚 Documentation

* highlight automatic conventional commits feature in README ([8011a2d](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/8011a2daec9b7d001eaadaffb1763f59c9599de9))

## [1.2.1](https://github.com/Yuki-Sakaguchi/gh-wizard/compare/v1.2.0...v1.2.1) (2025-08-24)


### 🐛 Bug Fixes

* enhance release workflow for reliable binary generation ([9ae1e35](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/9ae1e3562d1e47ad9436d2a23486f036b2bcae87))

## [1.2.0](https://github.com/Yuki-Sakaguchi/gh-wizard/compare/v1.1.0...v1.2.0) (2025-08-24)


### 🚀 Features

* add automatic setup scripts and Makefile ([a50e384](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/a50e38440f8749999f4b62bf07377dd769eacf56))

## [1.1.0](https://github.com/Yuki-Sakaguchi/gh-wizard/compare/v1.0.1...v1.1.0) (2025-08-24)


### 🚀 Features

* add automatic conventional commit prefix with lefthook ([65d8dfc](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/65d8dfc997aac4ed50e095aea3c35b2d1ddf8b34)), closes [#55](https://github.com/Yuki-Sakaguchi/gh-wizard/issues/55)
* implement automatic semantic versioning and release workflow ([04b5a89](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/04b5a89d8570c95c6d7a0224140f8b92899130c3)), closes [#53](https://github.com/Yuki-Sakaguchi/gh-wizard/issues/53)


### 🐛 Bug Fixes

* add workflow_dispatch trigger to release-please workflow ([d2e96b7](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/d2e96b77175fa7aec907e5ebb6046e46c40099ff))
* enhance workflow permissions for release-please ([2e544c8](https://github.com/Yuki-Sakaguchi/gh-wizard/commit/2e544c8f221023b44ba1aedf5c6db29dda758ecd))

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
