package generator

import (
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
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

type driver struct {
	packageName string
	rootPath    string
	out         io.Writer
}

func (d *driver) createFile(relPath string, rawTemplate string, args any) error {
	t, err := template.New(relPath).Parse(rawTemplate)

	newFileName := path.Join(d.rootPath, relPath)
	dirToCreate := path.Dir(newFileName)
	os.MkdirAll(dirToCreate, 0755)

	fileToWrite, err := os.Create(newFileName)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer fileToWrite.Close()
	err = t.Execute(fileToWrite, args)
	if err != nil {
		return fmt.Errorf("error executing template %s: %w", relPath, err)
	}

	fmt.Fprintf(d.out, "Created %s", relPath)

	return nil
}

func Generate(packageName string, packageRoot string, out io.Writer) error {
	driver := &driver{
		packageName: packageName,
		rootPath:    packageRoot,
		out:         out,
	}

	generateApp(driver)
	generateWeb(driver)

	return nil
}

func logCreate(out io.Writer, kind string, path string) {
	fmt.Fprintf(out, "create %s %s\n", kind, path)
}
