#!/bin/bash

set -eo pipefail

if [[ -z "$1" ]]; then
  echo "Error: A package name (go style, e.g. github.com/login/project) argument is required."
  exit 1
fi

NAME="$1"

if [[ ! "$NAME" =~ ^[a-zA-Z_-]+$ ]]; then
  echo "Error: The name must contain only letters, dashes, and underscores."
  exit 1
fi

TMPDIR=$(mktemp -d)
git clone --depth 1 git@github.com:blakewilliams/amaro.git --branch template --single-branch "$TMPDIR"

cp -r "$TMPDIR/_template" "$NAME"
cd "$TMPDIR/_template"
