package templates

import (
	"html/template"
	"net/http"
	"time"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/internal"
)

func productionProcessor(root *template.Template) acceptable.Processor {
	return func(rw http.ResponseWriter, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		p := &internal.WriterProxy{W: w}

		if fn, isFunc := match.Data.(acceptable.Supplier); isFunc {
			match.Data, err = fn()
			if err != nil {
				return err
			}
		}

		if template == "" {
			return root.Execute(p, match.Data)
		}
		return root.ExecuteTemplate(p, template, match.Data)
	}
}

//-------------------------------------------------------------------------------------------------

func debugProcessor(root *template.Template, rootDir, suffix string, files map[string]time.Time, funcMap template.FuncMap) acceptable.Processor {
	//files := make(map[string]time.Time)

	return func(rw http.ResponseWriter, match acceptable.Match, template string) (err error) {
		path := rootDir + "/" + template
		if _, exists := files[path]; !exists {
			files = findTemplates(rootDir, suffix)
		}

		w := match.ApplyHeaders(rw)

		p := &internal.WriterProxy{W: w}

		if fn, isFunc := match.Data.(acceptable.Supplier); isFunc {
			match.Data, err = fn()
			if err != nil {
				return err
			}
		}

		root = getCurrentTemplateTree(root, rootDir, suffix, files, funcMap)
		if template == "" {
			return root.Execute(p, match.Data)
		}
		return root.ExecuteTemplate(p, template, match.Data)
	}
}

func getCurrentTemplateTree(root *template.Template, rootDir, suffix string, files map[string]time.Time, funcMap template.FuncMap) *template.Template {
	changed := checkForChanges(files)
	if changed {
		root = parseTemplates(rootDir, files, funcMap)
	}
	return root
}

func checkForChanges(files map[string]time.Time) bool {
	changed := false

	for path, modTime := range files {
		fi, err := Fs.Stat(path)
		if err == nil {
			if fi.ModTime().After(modTime) {
				files[path] = fi.ModTime()
				changed = true
			}
		} else {
			delete(files, path)
		}
	}

	return changed
}
