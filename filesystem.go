package webdav

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// A FileSystem implements access to a collection of named files.
// The elements in a file path are separated by slash ('/', U+002F)
// characters, regardless of host operating system convention.
type FileSystem interface {
	Open(name string) (File, error)
	Create(name string) (File, error)
	Mkdir(path string) error
	Remove(name string) error
}

// A File is returned by a FileSystem's Open and Create method and can
// be served by the FileServer implementation.
type File interface {
	Stat() (os.FileInfo, error)
	Readdir(count int) ([]os.FileInfo, error)

	Read([]byte) (int, error)
	Write(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Close() error

	/* TODO: needed?
		Chdir() error
	    Chmod(mode FileMode) error
	    Chown(uid, gid int) error
	*/
}

// A Dir implements webdav.FileSystem using the native file
// system restricted to a specific directory tree.
//
// An empty Dir is treated as ".".
type Dir string

func (d Dir) sanitizePath(name string) (string, error) {
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return "", ErrInvalidCharPath
	}

	dir := string(d)
	if dir == "" {
		dir = "."
	}

	return filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name))), nil
}

// Open calls sanitizePath() and attempts to os.Open()
func (d Dir) Open(name string) (File, error) {
	p, err := d.sanitizePath(name)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Create calls sanitizePath() and attempts to os.Create()
func (d Dir) Create(name string) (File, error) {
	p, err := d.sanitizePath(name)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Mkdir creates a new directory with the specified name
func (d Dir) Mkdir(name string) error {
	p, err := d.sanitizePath(name)
	if err != nil {
		return err
	}

	return os.MkdirAll(p, os.ModePerm)
}

// Remove calls os.Remove() with a sanitized path
func (d Dir) Remove(name string) error {
	p, err := d.sanitizePath(name)
	if err != nil {
		return err
	}

	return os.Remove(p)
}

// mockup zero content file aka only header
type emptyFile struct{}

func (e emptyFile) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (e emptyFile) Seek(offset int64, whence int) (ret int64, err error) {
	return 0, nil
}
