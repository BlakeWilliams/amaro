package generator

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "template-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Replace with os.Stdout for debugging
	out := os.Stdout

	initFile := `package main
	import (
		"context"
		"github.com/blakewilliams/amaro"
	)
	
	func main() {
		runner := amaro.NewApplication("radical")
		runner.Execute(context.TODO())
	}
	`

	err = os.MkdirAll(tempDir+"/cmd/radical", 0755)
	require.NoError(t, err)
	f, err := os.Create(tempDir + "/cmd/radical/main.go")
	require.NoError(t, err)
	_, err = f.WriteString(initFile)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	initCmd := exec.Command("go", "mod", "init", "github.com/testing/testing")
	initCmd.Dir = tempDir
	err = initCmd.Run()
	require.NoError(t, err)

	cwd, err := os.Getwd()
	cwd = strings.TrimSuffix(cwd, "/generator")
	require.NoError(t, err)
	replaceCmd := exec.Command("go", "mod", "edit", "-replace", "github.com/blakewilliams/amaro="+cwd)
	replaceCmd.Dir = tempDir
	err = replaceCmd.Run()
	require.NoError(t, err)

	// cd to tempDir and run go test -v
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	tidyCmd.Stderr = out
	tidyCmd.Stdout = out
	err = tidyCmd.Run()
	require.NoError(t, err)

	// cd to tempDir and run go test -v
	generateCmd := exec.Command("go", "run", "cmd/radical/main.go", "generate")
	generateCmd.Dir = tempDir
	generateCmd.Stderr = out
	generateCmd.Stdout = out
	err = generateCmd.Run()
	require.NoError(t, err)

	// cd to tempDir and run go test -v
	tidyCmd = exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	tidyCmd.Stderr = out
	tidyCmd.Stdout = out
	err = tidyCmd.Run()
	require.NoError(t, err)

	require.FileExists(t, tempDir+"/internal/core/application.go")

	require.FileExists(t, tempDir+"/internal/web/server.go")

	testCmd := exec.Command("go", "test", "./...")
	testCmd.Dir = tempDir
	testCmd.Stderr = out
	testCmd.Stdout = out
	err = testCmd.Run()
	require.NoError(t, err)
}
