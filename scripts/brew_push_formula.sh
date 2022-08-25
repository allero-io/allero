#!/bin/bash
set -e

if [ $# -lt 1 ]
  then
    echo "Missing version"
    exit
fi

VERSION=$1

BREW_REPO_URL="https://${GITHUB_TOKEN}@github.com/allero-io/homebrew-allero.git"

git clone $BREW_REPO_URL
bash ./scripts/brew_formula_generator.sh $VERSION
cd homebrew-allero
git add -A
git commit -m "Brew formula update for allero version $VERSION"
git push
cd ..
