package proxy

import (
	"bytes"
	"net/http"
)

type BufferedResponseWriter struct {
	header     http.Header
	buffer     *bytes.Buffer
	statusCode int
}

func NewBufferedResponseWriter() *BufferedResponseWriter {
	return &BufferedResponseWriter{
		header:     make(http.Header),
		buffer:     new(bytes.Buffer),
		statusCode: http.StatusOK,
	}
}

func (w *BufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *BufferedResponseWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *BufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
