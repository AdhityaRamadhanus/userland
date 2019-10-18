package userland

import "io"

//ObjectMetaData is data about object
type ObjectMetaData struct {
	ContentType     string
	ContentEncoding string
	Size            int64
	CacheControl    string
	Path            string
}

//ObjectStorageService provide an interface to get objects
type ObjectStorageService interface {
	// should write with options
	Write(reader io.Reader, metadata ObjectMetaData) (string, error)
	Fetch(path string) ([]byte, ObjectMetaData, error)
}
