package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	respWrtr http.ResponseWriter
	gzipWrtr *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		respWrtr: w,
		gzipWrtr: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.respWrtr.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.gzipWrtr.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.respWrtr.Header().Set("Content-Encoding", "gzip")
	}
	c.respWrtr.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.gzipWrtr.Close()
}

type compressReader struct {
	readCloser io.ReadCloser
	gzipRdr    *gzip.Reader
}

func newCompressReader(readCloser io.ReadCloser) (*compressReader, error) {
	gzipRdr, err := gzip.NewReader(readCloser)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		readCloser: readCloser,
		gzipRdr:    gzipRdr,
	}, nil
}

func (compRdr compressReader) Read(p []byte) (n int, err error) {
	return compRdr.gzipRdr.Read(p)
}

func (compRdr *compressReader) Close() error {
	if err := compRdr.readCloser.Close(); err != nil {
		return err
	}
	return compRdr.gzipRdr.Close()
}

func gzipMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(respWrtr http.ResponseWriter, req *http.Request) {
		origRespWrtr := respWrtr

		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			respWrtr.Header().Set("Content-Encoding", "gzip")
			compWrtr := newCompressWriter(respWrtr)
			origRespWrtr = compWrtr
			defer compWrtr.Close()
		}

		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			compRdr, err := newCompressReader(req.Body)
			if err != nil {
				respWrtr.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = compRdr
			defer compRdr.Close()
		}

		next.ServeHTTP(origRespWrtr, req)
	}
}
