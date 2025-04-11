package templates_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/templates"
	"github.com/rickb777/expect"
	"github.com/spf13/afero"
)

func TestProductionInstance_using_files(t *testing.T) {
	templates.Fs = afero.NewOsFs() // real test files

	templates.ReloadOnTheFly = false

	render := templates.Templates("../example/templates/en", ".html", nil)

	data := datapkg.Of(Declaration{
		Proclamation: "A Title",
		Articles:     []Article{{N: 1, Text: "Text 1."}},
	})

	// request 1
	req := &http.Request{}
	w1 := httptest.NewRecorder()

	err := render(w1, req, data, "home.html", "en")
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w1.Body.String()).ToBe(t, "<html>\n<body>\n<h1>Home.</h1>\n<h4>A Title</h4>\n\n<h3>1</h3>\n<p>Text 1.</p>\n\n</body>\n</html>\n")

	// request 2
	w2 := httptest.NewRecorder()

	err = render(w2, req, data, "foo/bar.html", "en")
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w2.Body.String()).ToBe(t, "<html>\n<body>\n<h1>Bar.</h1>\n<h4>A Title</h4>\n\n<h3>1</h3>\n<p>Text 1.</p>\n\n</body>\n</html>\n")
}

func TestDebugInstance_using_fakes(t *testing.T) {
	rec := &recorder{fs: afero.NewMemMapFs()}
	templates.Fs = rec
	rec.fs.MkdirAll("foo/bar", 0755)
	afero.WriteFile(rec.fs, "synthetic/foo/home.html", []byte("<html>{{.Title}}-Home</html>"), 0644)
	afero.WriteFile(rec.fs, "synthetic/foo/bar/baz.html", []byte("<html>{{.Title}}-Baz</html>"), 0644)
	afero.WriteFile(rec.fs, "synthetic/foo/bar/util.js", []byte("func {{.Title}}() {}"), 0644)

	templates.ReloadOnTheFly = true

	render := templates.Templates("synthetic", ".html|.js", nil)

	data := datapkg.Of(map[string]string{"Title": "Hello"})

	//---------- request 1 ----------
	rec.opened = nil
	req := &http.Request{}
	w := httptest.NewRecorder()

	//t0 := time.Now()
	err := render(w, req, data, "foo/home.html", "en")
	//d1 := time.Now().Sub(t0)
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w.Body.String()).ToBe(t, "<html>Hello-Home</html>")
	expect.Slice(rec.opened).ToContainAll(t, "synthetic/foo/home.html", "synthetic/foo/bar/baz.html")

	//---------- request 2: javascript ----------
	rec.opened = nil
	w = httptest.NewRecorder()

	err = render(w, req, data, "foo/bar/util.js", "en")
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w.Body.String()).ToBe(t, "func Hello() {}")

	//---------- request 3: no change so no parsing ----------
	rec.opened = nil
	w = httptest.NewRecorder()

	//t2 := time.Now()
	err = render(w, req, data, "foo/home.html", "en")
	//d2 := time.Now().Sub(t2)
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w.Body.String()).ToBe(t, "<html>Hello-Home</html>")
	expect.Slice(rec.opened).ToBeEmpty(t)
	//expect.String(d2).To(BeNumerically("<", d1)) // it should be faster

	//---------- request 4: a different file ----------
	rec.opened = nil
	w = httptest.NewRecorder()

	err = render(w, req, data, "foo/bar/baz.html", "en")
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w.Body.String()).ToBe(t, "<html>Hello-Baz</html>")
	expect.Slice(rec.opened).ToBeEmpty(t)

	//---------- request 5: an altered file ----------
	rec.opened = nil
	w = httptest.NewRecorder()
	afero.WriteFile(rec.fs, "synthetic/foo/bar/baz.html", []byte("<html>{{.Title}}-Updated</html>"), 0644)

	err = render(w, req, data, "foo/bar/baz.html", "en")
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w.Body.String()).ToBe(t, "<html>Hello-Updated</html>")
	expect.Slice(rec.opened).ToContainAll(t, "synthetic/foo/home.html", "synthetic/foo/bar/baz.html")

	//---------- request 6: a new file ----------
	rec.opened = nil
	w = httptest.NewRecorder()
	afero.WriteFile(rec.fs, "synthetic/foo/bar/new.html", []byte("<html>{{.Title}}-New</html>"), 0644)

	err = render(w, req, data, "foo/bar/new.html", "en")
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w.Body.String()).ToBe(t, "<html>Hello-New</html>")
	expect.Slice(rec.opened).ToContainAll(t, "synthetic/foo/home.html", "synthetic/foo/bar/baz.html", "synthetic/foo/bar/new.html")

	//---------- request 7: ok after deleting an unrelated file ----------
	rec.opened = nil
	w = httptest.NewRecorder()
	rec.fs.Remove("synthetic/foo/bar/baz.html")

	err = render(w, req, data, "foo/bar/new.html", "en")
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.String(w.Body.String()).ToBe(t, "<html>Hello-New</html>")
	expect.Slice(rec.opened).ToBeEmpty(t)
}

//-------------------------------------------------------------------------------------------------

type recorder struct {
	fs     afero.Fs
	opened []string
}

func (r *recorder) Create(name string) (afero.File, error) {
	return r.fs.Create(name)
}

func (r *recorder) Mkdir(name string, perm os.FileMode) error {
	return r.fs.Mkdir(name, perm)
}

func (r *recorder) MkdirAll(path string, perm os.FileMode) error {
	return r.fs.MkdirAll(path, perm)
}

func (r *recorder) Open(name string) (afero.File, error) {
	r.opened = append(r.opened, name)
	return r.fs.Open(name)
}

func (r *recorder) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	//r.opened = append(r.opened, name)
	return r.fs.OpenFile(name, flag, perm)
}

func (r *recorder) Remove(name string) error {
	return r.fs.Remove(name)
}

func (r *recorder) RemoveAll(path string) error {
	return r.fs.RemoveAll(path)
}

func (r *recorder) Rename(oldname, newname string) error {
	return r.fs.Rename(oldname, newname)
}

func (r *recorder) Stat(name string) (os.FileInfo, error) {
	return r.fs.Stat(name)
}

func (r *recorder) Name() string {
	return r.fs.Name()
}

func (r *recorder) Chmod(name string, mode os.FileMode) error {
	return r.fs.Chmod(name, mode)
}

func (r *recorder) Chown(name string, uid, gid int) error {
	return r.fs.Chown(name, uid, gid)
}

func (r *recorder) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return r.fs.Chtimes(name, atime, mtime)
}

type Declaration struct {
	Proclamation string
	Articles     []Article
}

type Article struct {
	N    int
	Text string
}
