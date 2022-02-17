package templates

import (
	"html/template"
	"io"
	"net/http"
	"time"

	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
	"github.com/rickb777/acceptable/offer"
)

const DefaultPage = "_index.html"

func productionProcessor(root *template.Template) offer.Processor {
	return func(w io.Writer, req *http.Request, data datapkg.Data, template, language string) (err error) {
		p := &internal.WriterProxy{W: w}

		d, _, err := data.Content(template, language)
		if err != nil {
			return err
		}

		if template == "" {
			template = DefaultPage
		}
		return root.ExecuteTemplate(p, template, d)
	}
}

//-------------------------------------------------------------------------------------------------

func debugProcessor(root *template.Template, rootDir, suffix string, files map[string]time.Time, funcMap template.FuncMap) offer.Processor {
	return func(w io.Writer, req *http.Request, data datapkg.Data, template, language string) (err error) {
		path := rootDir + "/" + template
		if _, exists := files[path]; !exists {
			files = findTemplates(rootDir, suffix)
		}

		d, _, err := data.Content(template, language)
		if err != nil {
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
