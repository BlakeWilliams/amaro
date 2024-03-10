package template

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	fspath "path"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/mod/modfile"
)

//go:embed base
var f embed.FS
var packageNameRegex = regexp.MustCompile(`^[a-z]+$`)

type Generator struct {
	path string
}

func (g *Generator) RunCommand(ctx context.Context, w io.Writer) error {
	if g.path == "" {
		cwd, _ := os.Getwd()
		g.path = cwd
	}
	modFile := g.path + "/go.mod"
	f, err := os.Open(modFile)
	if err != nil {
		return fmt.Errorf("error opening go.mod file: %w", err)
	}
	contents, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading go.mod file: %w", err)
	}
	mf, err := modfile.Parse("go.mod", contents, nil)
	if err != nil {
		return fmt.Errorf("error parsing go.mod file: %w", err)
	}

	err = Generate(mf.Module.Mod.Path, g.path, w)
	if err != nil {
		return fmt.Errorf("error generating files: %w", err)
	}

	return nil
}

func Generate(packageName string, packageRoot string, out io.Writer) error {
	templateData := map[string]any{
		"packageName": packageName,
	}

	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			dirName := strings.TrimSuffix(packageRoot, "/") + "/" + strings.TrimPrefix(path, "base/")
			os.Mkdir(dirName, 0755)
			logCreate(out, "dir", dirName)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error walking the embedded file system: %w", err)
		}

		t, err := template.New(fspath.Base(path)).ParseFS(f, path)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %w", path, err)
		}

		newFileName := strings.TrimSuffix(packageRoot, "/") + "/" + strings.TrimPrefix(path, "base/")
		newFileName = strings.TrimSuffix(newFileName, ".tmpl")
		fileToWrite, err := os.Create(newFileName)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}

		defer func() {
			err := fileToWrite.Close()
			if err != nil {
				logCreate(out, "file", newFileName)
			}
		}()

		err = t.Execute(fileToWrite, templateData)
		if err != nil {
			return fmt.Errorf("error executing template %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the embedded file system: %w", err)
	}

	return nil
}

func logCreate(out io.Writer, kind string, path string) {
	fmt.Fprintf(out, "create %s %s\n", kind, path)
}
