package templates

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rickb777/acceptable/offer"
	"github.com/spf13/afero"
)

// Fs is used to obtain file information and content. It can be stubbed for testing.
var Fs = afero.NewOsFs()

// ReloadOnTheFly enables a development mode that reloads template files whenever they
// change, without restarting the server. This reduces performance and should be off
// (false) for production.
var ReloadOnTheFly = false

// Templates finds all the templates in the directory dir and its subdirectories
// that have names ending with the given suffix (usually ".html").
//
// The function map can be nil if not required.
//
// A processor is returned that handles requests using the templates available.
func Templates(dir, suffix string, funcMap template.FuncMap) offer.Processor {
	if funcMap == nil {
		funcMap = template.FuncMap{}
	}

	rootDir := filepath.Clean(dir)

	files := findTemplates(rootDir, suffix)

	if len(files) == 0 {
		panic("No HTML files were found in " + rootDir)
	}

	root := parseTemplates(rootDir, files, funcMap)

	if ReloadOnTheFly {
		return debugProcessor(root, rootDir, suffix, files, funcMap)
	}

	return productionProcessor(root)
}

//-------------------------------------------------------------------------------------------------

func findTemplates(rootDir, suffix string) map[string]time.Time {
	cleanRoot := filepath.Clean(rootDir)
	files := make(map[string]time.Time)

	err := afero.Walk(Fs, cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if e1 != nil {
			panic(fmt.Sprintf("Cannot load templates from: %s: %v\n", rootDir, e1))
		}

		if !info.IsDir() && strings.HasSuffix(path, suffix) {
			files[path] = time.Time{}
		}

		return nil
	})

	if err != nil {
		panic(fmt.Sprintf("Cannot load templates from: %s: %v\n", rootDir, err))
	}

	return files
}

func parseTemplates(rootDir string, files map[string]time.Time, funcMap template.FuncMap) *template.Template {
	pfx := len(rootDir) + 1
	root := template.New("")

	for path := range files {
		b, e2 := afero.ReadFile(Fs, path)
		if e2 != nil {
			panic(fmt.Sprintf("Read template error: %s: %v\n", path, e2))
		}

		name := path[pfx:]
		t := root.New(name).Funcs(funcMap)
		t, e2 = t.Parse(string(b))
		if e2 != nil {
			panic(fmt.Sprintf("Parse template error: %s: %v\n", path, e2))
		}
	}

	return root
}
