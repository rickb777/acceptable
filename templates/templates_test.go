package templates_test

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/templates"
	"github.com/spf13/afero"
)

func xTestProductionInstance_using_files(t *testing.T) {
	g := NewGomegaWithT(t)
	templates.Fs = afero.NewOsFs() // real test files

	render := templates.Templates("test-data", ".html", nil, false)

	match := acceptable.Match{
		Type:     "text",
		Subtype:  "html",
		Language: "en",
		Charset:  "utf-8",
		Data:     map[string]string{"Title": "Hello"},
	}

	// request 1
	w1 := httptest.NewRecorder()

	err := render(w1, match, "foo/home.html")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w1.Body.String()).To(Equal("<html>\n<body>\n<h1>Hello</h1>\n<p>Home.</p>\n</body>\n</html>"))

	// request 2
	w2 := httptest.NewRecorder()

	err = render(w2, match, "foo/bar/baz.html")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w2.Body.String()).To(Equal("<html>\n<body>\n<h1>Hello</h1>\n<p>Baz.</p>\n</body>\n</html>"))
}

func TestDebugInstance_using_fakes(t *testing.T) {
	g := NewGomegaWithT(t)
	rec := &recorder{fs: afero.NewMemMapFs()}
	templates.Fs = rec
	rec.fs.MkdirAll("foo/bar", 0755)
	afero.WriteFile(rec.fs, "synthetic/foo/home.html", []byte("<html>{{.Title}}-Home</html>"), 0644)
	afero.WriteFile(rec.fs, "synthetic/foo/bar/baz.html", []byte("<html>{{.Title}}-Baz</html>"), 0644)

	render := templates.Templates("synthetic", ".html", nil, true)

	match := acceptable.Match{
		Type:     "text",
		Subtype:  "html",
		Language: "en",
		Charset:  "utf-8",
		Data:     map[string]string{"Title": "Hello"},
	}

	//---------- request 1 ----------
	rec.opened = nil
	w := httptest.NewRecorder()

	//t0 := time.Now()
	err := render(w, match, "foo/home.html")
	//d1 := time.Now().Sub(t0)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w.Body.String()).To(Equal("<html>Hello-Home</html>"))
	g.Expect(rec.opened).To(ContainElements("synthetic/foo/home.html", "synthetic/foo/bar/baz.html"))

	//---------- request 2: no change so no parsing ----------
	rec.opened = nil
	w = httptest.NewRecorder()

	//t2 := time.Now()
	err = render(w, match, "foo/home.html")
	//d2 := time.Now().Sub(t2)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w.Body.String()).To(Equal("<html>Hello-Home</html>"))
	g.Expect(rec.opened).To(BeEmpty())
	//g.Expect(d2).To(BeNumerically("<", d1)) // it should be faster

	//---------- request 3: a different file ----------
	rec.opened = nil
	w = httptest.NewRecorder()

	err = render(w, acceptable.Match{
		Type:     "application",
		Subtype:  "xhtml+xml",
		Language: "en",
		Charset:  "utf-8",
		Data:     map[string]string{"Title": "Hello"},
	}, "foo/bar/baz.html")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w.Body.String()).To(Equal("<html>Hello-Baz</html>"))
	g.Expect(rec.opened).To(BeEmpty())

	//---------- request 4: an altered file ----------
	rec.opened = nil
	w = httptest.NewRecorder()
	afero.WriteFile(rec.fs, "synthetic/foo/bar/baz.html", []byte("<html>{{.Title}}-Updated</html>"), 0644)

	err = render(w, acceptable.Match{
		Type:     "application",
		Subtype:  "xhtml+xml",
		Language: "en",
		Charset:  "utf-8",
		Data:     map[string]string{"Title": "Hello"},
	}, "foo/bar/baz.html")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w.Body.String()).To(Equal("<html>Hello-Updated</html>"))
	g.Expect(rec.opened).To(ContainElements("synthetic/foo/home.html", "synthetic/foo/bar/baz.html"))

	//---------- request 5: a new file ----------
	rec.opened = nil
	w = httptest.NewRecorder()
	afero.WriteFile(rec.fs, "synthetic/foo/bar/new.html", []byte("<html>{{.Title}}-New</html>"), 0644)

	err = render(w, match, "foo/bar/new.html")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w.Body.String()).To(Equal("<html>Hello-New</html>"))
	g.Expect(rec.opened).To(ContainElements("synthetic/foo/home.html", "synthetic/foo/bar/baz.html", "synthetic/foo/bar/new.html"))

	//---------- request 5: ok after deleting an unrelated file ----------
	rec.opened = nil
	w = httptest.NewRecorder()
	rec.fs.Remove("synthetic/foo/bar/baz.html")

	err = render(w, match, "foo/bar/new.html")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w.Body.String()).To(Equal("<html>Hello-New</html>"))
	g.Expect(rec.opened).To(BeEmpty())
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
