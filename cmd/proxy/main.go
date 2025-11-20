package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"sync"
)

type ConsistentHash struct {
	sortedHashes []uint32
	circle       map[uint32]string
	mu           sync.RWMutex
}

func NewConsistentHash() *ConsistentHash {
	return &ConsistentHash{
		circle: make(map[uint32]string),
	}
}

func (ch *ConsistentHash) AddNode(node string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	hash := crc32.ChecksumIEEE([]byte(node))
	ch.circle[hash] = node
	ch.sortedHashes = append(ch.sortedHashes, hash)
	sort.Slice(ch.sortedHashes, func(i, j int) bool {
		return ch.sortedHashes[i] < ch.sortedHashes[j]
	})
	slog.Info("Node added to ring", "node", node, "hash", hash)
}

func (ch *ConsistentHash) GetNode(key string) string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.circle) == 0 {
		return ""
	}

	hash := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(ch.sortedHashes), func(i int) bool {
		return ch.sortedHashes[i] >= hash
	})

	// Wrap around
	if idx == len(ch.sortedHashes) {
		idx = 0
	}

	return ch.circle[ch.sortedHashes[idx]]
}

var (
	ring *ConsistentHash
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ring = NewConsistentHash()
	ring.AddNode("http://localhost:8081")
	ring.AddNode("http://localhost:8082")
	ring.AddNode("http://localhost:8083")

	http.HandleFunc("/set", handleSet)
	http.HandleFunc("/get", handleGet)

	slog.Info("Proxy server starting", "port", 8000)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		slog.Error("Proxy failed", "error", err)
		os.Exit(1)
	}
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

	node := ring.GetNode(req.Key)
	if node == "" {
		http.Error(w, "No nodes available", http.StatusServiceUnavailable)
		return
	}
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

	node := ring.GetNode(key)
	if node == "" {
		http.Error(w, "No nodes available", http.StatusServiceUnavailable)
		return
	}
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
