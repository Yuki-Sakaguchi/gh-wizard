#!/bin/bash
# Automatic conventional commit prefix based on branch name
# Part of gh-wizard lefthook configuration
#
# This script automatically adds conventional commit prefixes based on branch naming patterns:
# - feature/* or feat/* → feat:
# - fix/* or bugfix/* → fix:
# - docs/* → docs:
# - refactor/* → refactor:
# - test/* → test:
# - chore/* → chore:
# - perf/* → perf:
# - ci/* → ci:
# - build/* → build:
# - style/* → style:
#
# Usage: Called automatically by lefthook commit-msg hook
# Manual usage: ./scripts/add-conventional-prefix.sh <commit-msg-file>

set -e

# Check if commit message file is provided
if [ -z "$1" ]; then
    echo "Error: Commit message file path required"
    echo "Usage: $0 <commit-msg-file>"
    exit 1
fi

commit_msg_file="$1"

# Check if commit message file exists
if [ ! -f "$commit_msg_file" ]; then
    echo "Error: Commit message file not found: $commit_msg_file"
    exit 1
fi

# Read commit message
commit_msg=$(cat "$commit_msg_file")

# Skip empty commit messages
if [ -z "$commit_msg" ]; then
    exit 0
fi

# Get current branch name
current_branch=$(git symbolic-ref --short HEAD 2>/dev/null || echo "HEAD")

# Debug output (only if DEBUG_LEFTHOOK is set)
if [ -n "$DEBUG_LEFTHOOK" ]; then
    echo "Debug: Current branch: $current_branch"
    echo "Debug: Original commit message: $commit_msg"
fi

# Skip if already has conventional commit prefix
if echo "$commit_msg" | grep -qE "^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?:" ; then
    if [ -n "$DEBUG_LEFTHOOK" ]; then
        echo "Debug: Commit message already has conventional prefix, skipping"
    fi
    exit 0
fi

# Skip for merge commits, revert commits, etc.
if echo "$commit_msg" | grep -qE "^(Merge|Revert|fixup!|squash!)" ; then
    if [ -n "$DEBUG_LEFTHOOK" ]; then
        echo "Debug: Special commit type detected, skipping"
    fi
    exit 0
fi

# Skip for commit messages that start with '#' (commented out)
if echo "$commit_msg" | grep -qE "^#" ; then
    if [ -n "$DEBUG_LEFTHOOK" ]; then
        echo "Debug: Commit message starts with #, skipping"
    fi
    exit 0
fi

# Branch name to prefix mapping
case "$current_branch" in
    feature/*|feat/*)
        prefix="feat"
        ;;
    fix/*|bugfix/*)
        prefix="fix"
        ;;
    docs/*)
        prefix="docs"
        ;;
    refactor/*)
        prefix="refactor"
        ;;
    test/*)
        prefix="test"
        ;;
    chore/*)
        prefix="chore"
        ;;
    perf/*)
        prefix="perf"
        ;;
    ci/*)
        prefix="ci"
        ;;
    build/*)
        prefix="build"
        ;;
    style/*)
        prefix="style"
        ;;
    main|master|develop)
        # Skip prefix for main branches
        if [ -n "$DEBUG_LEFTHOOK" ]; then
            echo "Debug: Main branch detected, skipping prefix"
        fi
        exit 0
        ;;
    HEAD)
        # Detached HEAD state, skip
        if [ -n "$DEBUG_LEFTHOOK" ]; then
            echo "Debug: Detached HEAD state, skipping prefix"
        fi
        exit 0
        ;;
    *)
        # Skip prefix for unknown patterns
        if [ -n "$DEBUG_LEFTHOOK" ]; then
            echo "Debug: Unknown branch pattern '$current_branch', skipping prefix"
        fi
        exit 0
        ;;
esac

# Add prefix to commit message
prefixed_message="${prefix}: ${commit_msg}"

# Write prefixed message back to file
echo "$prefixed_message" > "$commit_msg_file"

# Debug output
if [ -n "$DEBUG_LEFTHOOK" ]; then
    echo "Debug: Added prefix '$prefix' to commit message"
    echo "Debug: New commit message: $prefixed_message"
fi

# Success
exit 0