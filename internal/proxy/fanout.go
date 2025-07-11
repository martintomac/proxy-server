package proxy

import (
	"bytes"
	"log"
	"net/http"
	"sync"
)

// FanOutHandler executes multiple handlers concurrently
type FanOutHandler struct {
	Handlers         []Handler
	ResponseStrategy ResponseStrategy
}

func (h *FanOutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(h.Handlers) == 0 {
		http.Error(w, "No handlers configured", http.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	responses := make([]bufferedResponse, len(h.Handlers))

	for i, handler := range h.Handlers {
		wg.Add(1)
		go func(index int, h Handler) {
			defer wg.Done()

			brw := newBufferedResponseWriter()
			h.ServeHTTP(brw, r)
			responses[index] = bufferedResponse{
				statusCode: brw.statusCode,
				header:     brw.header,
				body:       brw.buffer.Bytes(),
			}
		}(i, handler)
	}

	wg.Wait()

	h.ResponseStrategy.write(w, responses)
}

type ResponseStrategy interface {
	write(
		w http.ResponseWriter,
		responses []bufferedResponse,
	)
}

type bufferedResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

type bufferedResponseWriter struct {
	header     http.Header
	buffer     *bytes.Buffer
	statusCode int
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		header:     make(http.Header),
		buffer:     new(bytes.Buffer),
		statusCode: http.StatusOK,
	}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

type FirstSuccessfulResponseStrategy struct{}

func (s *FirstSuccessfulResponseStrategy) write(
	w http.ResponseWriter,
	responses []bufferedResponse,
) {
	// Write the first successful response
	for _, result := range responses {
		if result.statusCode >= 200 && result.statusCode < 300 {
			// Copy headers
			for name, values := range result.header {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}
			w.WriteHeader(result.statusCode)
			if _, err := w.Write(result.body); err != nil {
				log.Printf("Error writing response: %v", err)
			}
			return
		}
	}

	// If no successful response, return the first one
	if len(responses) > 0 {
		result := responses[0]
		for name, values := range result.header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(result.statusCode)
		if _, err := w.Write(result.body); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}
}
