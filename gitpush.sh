#!/usr/bin/env bash

set -e  # exit on error

REPO_SSH="git@github.com:derickschaefer/portabletext.git"
BRANCH="main"

# Initialize repo if needed
if [ ! -d ".git" ]; then
  echo "Initializing git repository..."
  git init
fi

# Ensure remote exists
if ! git remote | grep -q origin; then
  echo "Adding origin remote..."
  git remote add origin "$REPO_SSH"
fi

# Show status
git status

# Add all changes
git add .

# Commit (allow empty commit message override)
if git diff --cached --quiet; then
  echo "No changes to commit."
else
  git commit -m "Update portabletext AST"
fi

# Ensure branch exists and is checked out
git branch -M "$BRANCH"

# Push
echo "Pushing to GitHub..."
git push -u origin "$BRANCH"

echo "Done."
