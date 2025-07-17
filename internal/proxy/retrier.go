package proxy

import "net/http"

type RetrierHandler struct {
	Handler     Handler
	RetryPolicy RetryPolicy
	Retries     int
}

func (h *RetrierHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	var brw *BufferedResponseWriter
	maxTries := h.Retries + 1
	for try := 0; try < maxTries; try++ {
		brw = NewBufferedResponseWriter()
		h.Handler.ServeHTTP(brw, r)
		if !h.RetryPolicy.shouldRetry(brw.statusCode, brw.Header()) {
			break
		}
	}
	if brw == nil {
		panic("this should never happen")
	}

	w.WriteHeader(brw.statusCode)
	for name, values := range brw.Header() {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	_, err := w.Write(brw.buffer.Bytes())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}

type RetryPolicy interface {
	shouldRetry(
		statusCode int,
		header http.Header,
	) bool
}

type RetryOnNon2xxRetryPolicy struct{}

func (_ *RetryOnNon2xxRetryPolicy) shouldRetry(
	statusCode int,
	_ http.Header,
) bool {
	return !(statusCode >= 200 && statusCode < 300)
}
