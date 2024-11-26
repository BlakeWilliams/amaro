package envy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

var validKey = regexp.MustCompile(`\A[A-Z_][A-Z0-9_]*\z`)

// Load reads the .env file in the current directory and sets the environment
// variables. If `env` is passed, it will load the equivalent
// `fmt.Sprintf(".env.%s", env)` file before the `.env` file, taking precedence
// over the .env values.
//
// Load will look up the directory tree for an .env file, stopping when .env or
// .git is found.
func Load(env string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		if _, err := os.Stat(filepath.Join(root, ".env")); err == nil {
			break
		}

		if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
			break
		}

		root = path.Dir(root)
		if root == "/" || root == "." {
			return fmt.Errorf("could not find a .env file")
		}
	}

	fs := os.DirFS(root)
	return LoadFS(fs, env)
}

// LoadFS reads the .env file in the provided file system and sets
// the environment variables.
//
// If `env` is passed, it will load the equivalent `.env.{env}` file first,
// taking precedence over the `.env` values.
func LoadFS(fs fs.FS, env string) error {
	if env != "" {
		if err := loadEnvFile(fs, ".env."+env); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	err := loadEnvFile(fs, ".env")
	if err != nil {
		return err
	}

	return nil
}

// LoadString reads the provided string as an .env file and sets the
// defined environment variables.
func LoadString(b string) error {
	p := &parser{
		mapping: make(map[string]string),
		content: []rune(string(b)),
	}

	err := p.parse()
	if err != nil {
		return fmt.Errorf("failed to load env: %w", err)
	}

	for k, v := range p.mapping {
		if _, ok := os.LookupEnv(k); !ok {
			os.Setenv(k, v)
		}
	}

	return nil
}

func loadEnvFile(fs fs.FS, fileName string) error {
	f, err := fs.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	str, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return LoadString(string(str))
}
