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

//go:embed _base
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

func (g *Generator) CommandName() string {
	return "generate"
}

func (g *Generator) CommandDescription() string {
	return "Generates the base files for a new amaro project"
}

func Generate(packageName string, packageRoot string, out io.Writer) error {
	templateData := map[string]any{
		"PackageName": packageName,
	}

	err := fs.WalkDir(f, ".", func(rawPath string, d fs.DirEntry, err error) error {
		path := strings.TrimPrefix(rawPath, "_base/")
		root := strings.TrimSuffix(packageRoot, "/")

		if path == "." || path == "_base" {
			return nil
		}

		if d.IsDir() {
			dirName := root + "/" + path
			os.Mkdir(dirName, 0755)
			logCreate(out, "dir", dirName)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error walking the embedded file system: %w", err)
		}

		t, err := template.New(fspath.Base(path)).ParseFS(f, rawPath)
		if err != nil {
			fmt.Println("error parsing template", rawPath, err)
			return fmt.Errorf("error parsing template %s: %w", rawPath, err)
		}

		newFileName := packageRoot + "/" + path
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
