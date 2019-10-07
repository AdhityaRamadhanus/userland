package userland

import "io"

type ObjectMetaData struct {
	ContentType     string
	ContentEncoding string
	Size            int64
	CacheControl    string
	Path            string
}

type ObjectStorageService interface {
	// should write with options
	Write(reader io.Reader, metadata ObjectMetaData) (string, error)
	Fetch(path string) ([]byte, ObjectMetaData, error)
}
