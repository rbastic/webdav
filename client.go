package webdav

// FileSystemCloser is an interface for webdav to wrap basic filesystem ops. See filesystem.go
type FileSystemCloser interface {
	FileSystem
	Close() error
}
