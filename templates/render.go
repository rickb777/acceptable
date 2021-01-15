package templates

import (
	"html/template"
	"net/http"
	"time"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
)

const DefaultPage = "_index.html"

func productionProcessor(root *template.Template) acceptable.Processor {
	return func(rw http.ResponseWriter, req *http.Request, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		p := &internal.WriterProxy{W: w}

		d, err := data.GetContentAndApplyExtraHeaders(rw, req, match.Data, template, match.Language)
		if err != nil || d == nil {
			return err
		}

		if template == "" {
			template = DefaultPage
		}
		return root.ExecuteTemplate(p, template, d)
	}
}

//-------------------------------------------------------------------------------------------------

func debugProcessor(root *template.Template, rootDir, suffix string, files map[string]time.Time, funcMap template.FuncMap) acceptable.Processor {
	return func(rw http.ResponseWriter, req *http.Request, match acceptable.Match, template string) (err error) {
		path := rootDir + "/" + template
		if _, exists := files[path]; !exists {
			files = findTemplates(rootDir, suffix)
		}

		w := match.ApplyHeaders(rw)

		d, err := data.GetContentAndApplyExtraHeaders(rw, req, match.Data, template, match.Language)
		if err != nil || d == nil {
			return err
		}

		if template == "" {
			template = DefaultPage
		}

		p := &internal.WriterProxy{W: w}
		root = getCurrentTemplateTree(root, rootDir, suffix, files, funcMap)

		return root.ExecuteTemplate(p, template, d)
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
