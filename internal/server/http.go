package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"distributed-kv/internal/store"
)

type HTTPServer struct {
	store store.Store
	port  int
}

func NewHTTPServer(port int, store store.Store) *HTTPServer {
	return &HTTPServer{
		store: store,
		port:  port,
	}
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", s.handleGet)
	mux.HandleFunc("/set", s.handleSet)
	mux.HandleFunc("/delete", s.handleDelete)

	addr := fmt.Sprintf(":%d", s.port)
	slog.Info("Starting HTTP server", "address", addr)
	return http.ListenAndServe(addr, mux)
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetResponse struct {
	Value string `json:"value"`
}

func (s *HTTPServer) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.store.Set(req.Key, req.Value); err != nil {
		slog.Error("Failed to set key", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *HTTPServer) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	val, err := s.store.Get(key)
	if err != nil {
		if err == store.ErrKeyNotFound {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
		slog.Error("Failed to get key", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := GetResponse{Value: val}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *HTTPServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	if err := s.store.Delete(key); err != nil {
		slog.Error("Failed to delete key", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
