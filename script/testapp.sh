#!/bin/bash

set -eo pipefail

# Create a temporary directory
temp_dir=$(mktemp -d -t template-test)

# Initialize variables
init_file='package main
import (
	"context"
	"github.com/blakewilliams/amaro"
)

type tempApp struct {}
func (t *tempApp) AppName() string {
	return "mytestapp"
}

func main() {
	runner := amaro.NewApplication(&tempApp{})
	runner.Execute(context.Background())
}
'

originalWd=$(pwd)

# Create directories and files
mkdir -p "$temp_dir/cmd/radical"
echo "$init_file" > "$temp_dir/cmd/radical/main.go"

# Initialize Go module
cd "$temp_dir"
go mod init github.com/testing/testing

# Get the current working directory, adjusted as needed
cwd=$(pwd | sed 's:/generator$::')

# Replace the module dependency
go mod edit -replace "github.com/blakewilliams/amaro=$originalWd"

# Run `go mod tidy`
go mod tidy

# Run the generate command
go run cmd/radical/main.go generate

# Run `go mod tidy` again
go mod tidy

echo "Created new project in $temp_dir"
