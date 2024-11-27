package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

var (
	allowedEncodingTypes = map[string]struct{}{
		"gzip": {},
	}

	allowedContentTypes = map[string]struct{}{
		"application/json": {},
		"text/html":        {},
	}
)

// compressWriter is a wrapper around `http.ResponseWriter` that compresses the
// response body using gzip. It sets the `Content-Encoding` header when the response
// status code is less than 300.
type compressWriter struct {
	w  http.ResponseWriter // The original ResponseWriter.
	zw *gzip.Writer        // The gzip writer for compression.
}

// newCompressWriter creates a new `compressWriter` that wraps an `http.ResponseWriter`
// for gzip compression.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header returns the header map for the writer, allowing modification before the
// response is written.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write compresses the provided bytes and writes them to the underlying writer.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader writes the HTTP status code to the underlying writer, setting the
// `Content-Encoding: gzip` header for successful responses.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip writer, flushing any remaining data to the underlying writer.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader is a wrapper around `io.ReadCloser` that decompresses the
// request body if it is gzip-encoded.
type compressReader struct {
	r  io.ReadCloser // The original request body.
	zr *gzip.Reader  // The gzip reader for decompression.
}

// newCompressReader creates a new `compressReader` that wraps an `io.ReadCloser`
// for gzip decompression. Returns an error if the reader cannot be initialized.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read reads decompressed data from the gzip reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the original request body and the gzip reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware compresses HTTP responses and decompresses HTTP requests if gzip
// is supported and enabled by the client.
//
//   - For responses: If the client indicates support for gzip via the `Accept-Encoding`
//     header, the response is compressed.
//   - For requests: If the client sends gzip-encoded content via the `Content-Encoding`
//     header, the middleware decompresses it.
//
// Parameters:
//   - h: The next `http.Handler` to wrap.
//
// Returns:
//   - A wrapped `http.Handler` with gzip compression and decompression logic.
func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if isContentTypeAllowed(r.Header.Get("Content-Type")) && isEncodingTypeAllowed(r.Header.Values("Accept-Encoding")) {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	})
}

func isEncodingTypeAllowed(acceptEncoding []string) bool {
	for _, value := range acceptEncoding {
		if _, ok := allowedEncodingTypes[value]; ok {
			return true
		}
	}
	return false
}

func isContentTypeAllowed(contentType string) bool {
	_, ok := allowedContentTypes[contentType]

	return ok
}
