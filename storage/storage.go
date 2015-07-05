package storage

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DateFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
)

var (
	ErrNotImplemented error = errors.New("Not Implemented")
	ErrNotExists      error = errors.New("Not Exists")
	ErrNoExpires      error = errors.New("No Expires Field")
	ErrNoLastModified error = errors.New("No LastModified Field")
)

type Object interface {
	LastModified() (time.Time, error)
	ETag() string
	Expires() (time.Time, error)
	ContentMD5() string
	ContentLength() int64
	ContentType() string
	ContentEncoding() string
	Body() io.ReadCloser
	Response() (*http.Response, error)
}

type Store interface {
	URL() string
	DateFormat() string
	GetObject(object string, start, end int64) (Object, error)
	PutObject(object string, header http.Header, data io.ReadCloser) error
	CopyObject(destObject string, srcObject string) error
	HeadObject(object string) (http.Header, error)
	DeleteObject(object string) error
}

func OpenURI(uri string) (Store, error) {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid URI: %s", uri)
	}
	return Open(parts[0], parts[1])
}

func expandPick(sourceString string) string {
	matches, err := filepath.Glob(sourceString)
	if err != nil || len(matches) == 0 {
		return sourceString
	}

	sort.Strings(matches)

	return matches[len(matches)-1]
}

func Open(driver, sourceString string) (Store, error) {
	switch driver {
	case "file":
		return NewFileStore(expandPick(sourceString))
	case "zip":
		return NewZipStore(expandPick(sourceString))
	default:
		return nil, fmt.Errorf("Invaild Storage dirver: %#v", driver)
	}
}

func ReadJson(r io.Reader) ([]byte, error) {
	var b bytes.Buffer
	s := bufio.NewScanner(r)

	for s.Scan() {
		if !strings.HasPrefix(strings.TrimSpace(s.Text()), "//") {
			b.Write(s.Bytes())
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
