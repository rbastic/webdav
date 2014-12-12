package webdav

import (
	"errors"
	"net/http"
)

// status codes
const (
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusAccepted            = http.StatusAccepted
	StatusNoContent           = http.StatusNoContent
	StatusMovedPermanently    = http.StatusMovedPermanently
	StatusMovedTemporarily    = 302 // TODO: duplicate of http.StatusFound ?
	StatusNotModified         = http.StatusNotModified
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusInternalServerError = http.StatusInternalServerError
	StatusNotImplemented      = http.StatusNotImplemented
	StatusMethodNotAllowed    = http.StatusMethodNotAllowed
	StatusConflict            = http.StatusConflict
	StatusPreconditionFailed  = http.StatusPreconditionFailed
)

// extended status codes, http://www.webdav.org/specs/rfc4918.html#status.code.extensions.to.http11
const (
	StatusMulti               = 207
	StatusLocked              = 423
	StatusFailedDependency    = 424
	StatusInsufficientStorage = 507
)

var statusText = map[int]string{
	StatusMovedTemporarily:    "Moved Temporarily",
	StatusMulti:               "Multi-Status",
	StatusLocked:              "Locked",
	StatusFailedDependency:    "Failed Dependency",
	StatusInsufficientStorage: "Insufficient Storage",
}

// StatusText returns a text for the HTTP status code. It returns the empty string if the code is unknown.
func StatusText(code int) string {
	if t, ok := statusText[code]; ok {
		return t
	}

	return http.StatusText(code)
}

// internal error variables
var (
	ErrInvalidCharPath = errors.New("invalid character in file path")
	ErrNotImplemented  = errors.New("feature not yet implemented")
)
