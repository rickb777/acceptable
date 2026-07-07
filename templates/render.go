package templates

import (
	tmplpkg "html/template"
	"io"
	"net/http"
	"time"

	dpkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
	"github.com/rickb777/acceptable/offer"
)

// DefaultPage is the template name when a blank string is supplied.
// Alter this during startup if required.
var DefaultPage = "_index.html"

func productionProcessor(root *tmplpkg.Template) offer.Processor {
	return func(w io.Writer, req *http.Request, data dpkg.Data, params dpkg.Chosen) (err error) {
		p := internal.EnsureNewline(w)

		d, _, err := data.Content(params)
		if err != nil {
			return err
		}

		if params.Template == "" {
			params.Template = DefaultPage
		}
		return root.ExecuteTemplate(p, params.Template, d)
	}
}

//-------------------------------------------------------------------------------------------------

func debugProcessor(root *tmplpkg.Template, rootDir, suffix string, files map[string]time.Time, funcMap tmplpkg.FuncMap) offer.Processor {
	return func(w io.Writer, req *http.Request, data dpkg.Data, params dpkg.Chosen) (err error) {
		if params.Template == "" {
			params.Template = DefaultPage
		}

		path := rootDir + "/" + params.Template
		if _, exists := files[path]; !exists {
			files = findTemplates(rootDir, suffix)
		}

		d, _, err := data.Content(params)
		if err != nil {
			return err
		}

		p := internal.EnsureNewline(w)
		root = getCurrentTemplateTree(root, rootDir, suffix, files, funcMap)

		return root.ExecuteTemplate(p, params.Template, d)
	}
}

func getCurrentTemplateTree(root *tmplpkg.Template, rootDir, suffix string, files map[string]time.Time, funcMap tmplpkg.FuncMap) *tmplpkg.Template {
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
