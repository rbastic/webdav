package webdav

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"
)

func init() {
}

// FlushGlog exposes the glog object for you, so you can flush your log when you exit.
// E.g. Call webdav.FlushGlog() in whatever shutdown/finalizer code you have before your
// server terminates.
func FlushGlog() {
	glog.Flush()
}

// Handler configures the FileSystem object with the Server struct
func Handler(root FileSystem) http.Handler {
	return &Server{Fs: root}
}

// Server represents a given filesystem-server
type Server struct {
	// trimmed path prefix
	TrimPrefix string

	// files are readonly?
	ReadOnly bool

	// generate directory listings?
	Listings bool

	// access to a collection of named files
	Fs FileSystem
}

func generateToken() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%d-%d-00105A989226:%d",
		r.Int31(), r.Int31(), time.Now().UnixNano())
}

// NewServer allows us to create a new Server struct, for a given filesystem path
func NewServer(dir, prefix string, listDir bool) *Server {
	return &Server{
		Fs:         Dir(dir),
		TrimPrefix: prefix,
		Listings:   listDir,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// XXX disable this in production
	glog.Infoln("DAV:", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":
		s.doGet(w, r)
	case "HEAD":
		s.doHead(w, r)
	case "DELETE":
		s.doDelete(w, r)
	case "PUT":
		s.doPut(w, r)

	default:
		glog.Infoln("DAV:", "unknown method", r.Method)
		w.WriteHeader(StatusBadRequest)
	}
}

// convert request url to path
func (s *Server) url2path(u *url.URL) string {
	glog.Infoln("u.Path=", u.Path)
	if u.Path == "" {
		return "/"
	}

	if p := strings.TrimPrefix(u.Path, s.TrimPrefix); len(p) < len(u.Path) {
		glog.Infoln("strings.Trim,", p, " = ", strings.Trim(p, "/"))
		return strings.Trim(p, "/")
	}

	return "/"
}

// does path exists?
func (s *Server) pathExists(path string) bool {
	f, err := s.Fs.Open(path)
	if err != nil {
		// TODO: error logging?
		return false
	}
	defer f.Close()

	return true
}

// is path a directory?
func (s *Server) pathIsDirectory(path string) bool {
	f, err := s.Fs.Open(path)
	if err != nil {
		// TODO: error logging?
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		// TODO: error logging?
		return false
	}

	return fi.IsDir()
}

// http://www.webdav.org/specs/rfc4918.html#rfc.section.9.4
func (s *Server) doGet(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("DAV", "GET", r.RequestURI)
	s.serveResource(w, r, true)
}

// http://www.webdav.org/specs/rfc4918.html#rfc.section.9.4
func (s *Server) doHead(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("DAV", "HEAD", r.RequestURI)
	s.serveResource(w, r, false)
}

// TODO(rbastic): audit this code
func (s *Server) serveResource(w http.ResponseWriter, r *http.Request, serveContent bool) {
	path := s.url2path(r.URL)

	f, err := s.Fs.Open(path)
	if err != nil {
		glog.Infoln("DAV:", "404, File missing on disk:", r.RequestURI, "error", err)
		http.Error(w, r.RequestURI, StatusNotFound)
		return
	}
	defer f.Close()

	// TODO: what if path is collection?

	fi, err := f.Stat()
	if err != nil {
		// TODO: log locally also, configurably
		glog.Infoln("DAV:", "404, File missing on disk:", r.RequestURI, "error", err)
		http.Error(w, r.RequestURI, StatusNotFound)
		return
	}
	modTime := fi.ModTime()

	if serveContent {
		http.ServeContent(w, r, path, modTime, f)
	} else {
		// TODO: better way to send only head
		http.ServeContent(w, r, path, modTime, emptyFile{})
	}
}

// http://www.webdav.org/specs/rfc4918.html#METHOD_DELETE
func (s *Server) doDelete(w http.ResponseWriter, r *http.Request) {
	if s.ReadOnly {
		glog.Infoln("DAV:", "DELETE attempted, file read-only", r.URL)
		w.WriteHeader(StatusForbidden)
		return
	}

	if s.deleteResource(s.url2path(r.URL), w, r, true) {
		glog.Infoln("DAV:", "DELETE successful", r.URL)
	} else {
		glog.Infoln("DAV:", "DELETE unsuccessful", r.URL)
	}

}

// TODO(rbastic): audit this code
func (s *Server) deleteResource(path string, w http.ResponseWriter, r *http.Request, setStatus bool) bool {

	if !s.pathExists(path) {
		glog.Infoln("404", r.RequestURI)
		w.WriteHeader(StatusNotFound)
		return false
	}

	if !s.pathIsDirectory(path) {
		if err := s.Fs.Remove(path); err != nil {
			// TODO: log locally
			w.WriteHeader(StatusInternalServerError)
			return false
		}
	} else {
		// XXX: Deleting entire paths is completely disabled.
	}

	if setStatus {
		w.WriteHeader(StatusNoContent)
	}
	return true
}

func (s *Server) doPut(w http.ResponseWriter, r *http.Request) {
	if s.ReadOnly {
		w.WriteHeader(StatusForbidden)
		return
	}

	path := s.url2path(r.URL)

	if s.pathIsDirectory(path) {
		// use MKCOL instead
		glog.Infoln("use mkcol instead perhaps, path", path, "is already a directory")
		glog.Infoln("DAV:", "use mkcol instead perhaps, path", path)
		w.WriteHeader(StatusMethodNotAllowed)
		return
	}

	exists := s.pathExists(path)

	// TODO: content range / partial put ?, re-enable os.MkdirAll()

	/*
		err := os.MkdirAll(path, 0600)
		if err != nil {
			log.Printf("error %+v making directory %+v  ", err, path)
		}
	*/

	// truncate file if exists
	file, err := s.Fs.Create(path)
	if err != nil {
		// TODO: having stupid problems?
		glog.Infoln("DAV:", "error with create path", path, "error", err)
		w.WriteHeader(StatusConflict)
		return
	}

	// XXX: investigate how io.Copy() is implemented, is it thread-safe or do
	// we need to change this implementation to work more like how nginx's does,
	// using temporary filenames and then atomic rename's ?

	if _, err := io.Copy(file, r.Body); err != nil {
		glog.Infoln("DAV:", "error with ioCopy", file, "error", err)
		w.WriteHeader(StatusConflict)
		file.Close()
	} else {
		if exists {
			glog.Infoln("DAV:", "status no content", file, "error", err)
			w.WriteHeader(StatusNoContent)
		} else {
			w.WriteHeader(StatusCreated)
		}
	}
	file.Close()
}
