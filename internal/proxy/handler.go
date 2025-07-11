package proxy

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Handler interface {
	ServeHTTP(
		writer http.ResponseWriter,
		req *http.Request,
	)
}

type StaticHandler struct {
	message string
}

func (h *StaticHandler) ServeHTTP(
	w http.ResponseWriter,
	_ *http.Request,
) {
	_, err := w.Write([]byte(h.message))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	return
}

type DebugHandler struct{}

func (h *DebugHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	body, err := readBody(r)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return
	}
	response := debugResponse{
		Method:  r.Method,
		URL:     r.URL.String(),
		Headers: r.Header,
		Body:    string(body),
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	err = encoder.Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	return
}

func readBody(r *http.Request) ([]byte, error) {
	if (r.Body == nil) || (r.ContentLength == 0) {
		return []byte{}, nil
	}
	return io.ReadAll(r.Body)
}

type debugResponse struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type EchoHandler struct{}

func (h *EchoHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := r.Write(w)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	return
}

type NotFoundHandler struct{}

func (f *NotFoundHandler) ServeHTTP(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.WriteHeader(http.StatusNotFound)
	_, err := w.Write([]byte("Not found"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	return
}

type ChaosHandler struct {
	Handler       Handler
	FailureChance float64
	rand          *rand.Rand
}

var randGenerator = rand.New(rand.NewSource(time.Now().UnixMilli()))

func NewChaosHandler(h Handler, failureChance float64) *ChaosHandler {
	return &ChaosHandler{
		Handler:       h,
		FailureChance: failureChance,
		rand:          randGenerator,
	}
}

func (h *ChaosHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if h.rand.Float64() <= h.FailureChance {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Chaos"))
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
		return
	}

	h.Handler.ServeHTTP(w, r)
}
