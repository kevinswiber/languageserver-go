// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package span

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

const fileScheme = "file"

// URI represents the full URI for a file.
type URI string

// Filename returns the file path for the given URI.
// It is an error to call this on a URI that is not a valid filename.
func (uri URI) Filename() string {
	filename, err := filename(uri)
	if err != nil {
		panic(err)
	}
	return filepath.FromSlash(filename)
}

func filename(uri URI) (string, error) {
	u, err := url.ParseRequestURI(string(uri))
	if err != nil {
		return "", err
	}
	if u.Scheme != fileScheme {
		return "", fmt.Errorf("only file URIs are supported, got %v", u.Scheme)
	}
	if isWindowsDriveURI(u.Path) {
		u.Path = u.Path[1:]
	}
	return u.Path, nil
}

// NewURI returns a span URI for the string.
// It will attempt to detect if the string is a file path or uri.
func NewURI(s string) URI {
	if u, err := url.PathUnescape(s); err == nil {
		s = u
	}
	if strings.HasPrefix(s, fileScheme+"://") {
		return URI(s)
	}
	return FileURI(s)
}

func CompareURI(a, b URI) int {
	if a == b {
		return 0
	}
	// If we have the same URI basename, we may still have the same file URIs.
	fa := a.Filename()
	fb := b.Filename()
	if strings.EqualFold(filepath.Base(fa), filepath.Base(fb)) {
		// Stat the files to check if they are equal.
		if infoa, err := os.Stat(fa); err == nil {
			if infob, err := os.Stat(fb); err == nil {
				if os.SameFile(infoa, infob) {
					return 0
				}
			}
		}
	}
	return strings.Compare(fa, fb)
}

// FileURI returns a span URI for the supplied file path.
// It will always have the file scheme.
func FileURI(path string) URI {
	// Handle standard library paths that contain the literal "$GOROOT".
	// TODO(rstambler): The go/packages API should allow one to determine a user's $GOROOT.
	const prefix = "$GOROOT"
	if len(path) >= len(prefix) && strings.EqualFold(prefix, path[:len(prefix)]) {
		suffix := path[len(prefix):]
		path = runtime.GOROOT() + suffix
	}
	if !isWindowsDrivePath(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	// Check the file path again, in case it became absolute.
	if isWindowsDrivePath(path) {
		path = "/" + path
	}
	path = filepath.ToSlash(path)
	u := url.URL{
		Scheme: fileScheme,
		Path:   path,
	}
	return URI(u.String())
}

// isWindowsDrivePath returns true if the file path is of the form used by
// Windows. We check if the path begins with a drive letter, followed by a ":".
func isWindowsDrivePath(path string) bool {
	if len(path) < 4 {
		return false
	}
	return unicode.IsLetter(rune(path[0])) && path[1] == ':'
}

// isWindowsDriveURI returns true if the file URI is of the format used by
// Windows URIs. The url.Parse package does not specially handle Windows paths
// (see https://golang.org/issue/6027). We check if the URI path has
// a drive prefix (e.g. "/C:"). If so, we trim the leading "/".
func isWindowsDriveURI(uri string) bool {
	if len(uri) < 4 {
		return false
	}
	return uri[0] == '/' && unicode.IsLetter(rune(uri[1])) && uri[2] == ':'
}
