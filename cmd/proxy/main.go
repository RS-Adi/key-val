package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"net/http"
	"os"
)

var nodes = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	http.HandleFunc("/set", handleSet)
	http.HandleFunc("/get", handleGet)

	slog.Info("Proxy server starting", "port", 8000)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		slog.Error("Proxy failed", "error", err)
		os.Exit(1)
	}
}

func getNode(key string) string {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() % uint32(len(nodes))
	return nodes[idx]
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body to get key
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req SetRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	node := getNode(req.Key)
	slog.Info("Forwarding SET", "key", req.Key, "node", node)

	proxyReq, err := http.NewRequest(http.MethodPost, node+"/set", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key", http.StatusBadRequest)
		return
	}

	node := getNode(key)
	slog.Info("Forwarding GET", "key", key, "node", node)

	proxyReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/get?key=%s", node, key), nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
