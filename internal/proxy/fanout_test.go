package proxy

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFanOutHandler_ServeHTTP(t *testing.T) {
	t.Run("successful fanout request", func(t *testing.T) {
		capturingHandler := CapturingHandler{}

		handler := FanOutHandler{
			Handlers: []Handler{
				&capturingHandler,
				&capturingHandler,
			},
			ResponseStrategy: &FirstSuccessfulResponseStrategy{},
		}

		responseRecorder := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/capture", nil)
		if err != nil {
			t.Fatal(err)
		}

		handler.ServeHTTP(responseRecorder, req)

		assert.Equal(t, 2, capturingHandler.Invocations, "Expected handler to be invoked twice")
	})

}

type CapturingHandler struct {
	Invocations int
}

func (h *CapturingHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.Invocations++
	w.WriteHeader(http.StatusOK)
}
