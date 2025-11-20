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
	"strings"
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

func (ch *ConsistentHash) GetNodes(key string, count int) []string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.circle) == 0 {
		return nil
	}

	hash := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(ch.sortedHashes), func(i int) bool {
		return ch.sortedHashes[i] >= hash
	})

	if idx == len(ch.sortedHashes) {
		idx = 0
	}

	nodes := make([]string, 0, count)
	seen := make(map[string]bool)

	for len(nodes) < count && len(nodes) < len(ch.circle) {
		node := ch.circle[ch.sortedHashes[idx]]
		if !seen[node] {
			nodes = append(nodes, node)
			seen[node] = true
		}
		idx = (idx + 1) % len(ch.sortedHashes)
	}

	return nodes
}

var (
	ring *ConsistentHash
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ring = NewConsistentHash()

	// Check if "NODES" env var is set (format: "node1:8080,node2:8080")
	nodeList := os.Getenv("NODES")
	if nodeList == "" {
		// Default for local testing
		nodeList = "http://localhost:8081,http://localhost:8082,http://localhost:8083"
	}

	nodes := strings.Split(nodeList, ",")
	for _, node := range nodes {
		ring.AddNode(node)
	}

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

	nodes := ring.GetNodes(req.Key, 3)
	if len(nodes) == 0 {
		http.Error(w, "No nodes available", http.StatusServiceUnavailable)
		return
	}
	slog.Info("Replicating SET", "key", req.Key, "nodes", nodes)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for _, node := range nodes {
		wg.Add(1)
		go func(nodeAddr string) {
			defer wg.Done()
			proxyReq, err := http.NewRequest(http.MethodPost, nodeAddr+"/set", bytes.NewBuffer(body))
			if err != nil {
				slog.Error("Failed to create request", "node", nodeAddr, "error", err)
				return
			}
			proxyReq.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(proxyReq)
			if err != nil {
				slog.Error("Failed to forward request", "node", nodeAddr, "error", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(node)
	}

	wg.Wait()

	if successCount >= 2 { // Quorum W=2
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		http.Error(w, "Failed to achieve write quorum", http.StatusBadGateway)
	}
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

	nodes := ring.GetNodes(key, 3)
	if len(nodes) == 0 {
		http.Error(w, "No nodes available", http.StatusServiceUnavailable)
		return
	}
	slog.Info("Reading GET", "key", key, "nodes", nodes)

	for _, node := range nodes {
		proxyReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/get?key=%s", node, key), nil)
		if err != nil {
			slog.Error("Failed to create request", "node", node, "error", err)
			continue
		}

		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			slog.Error("Failed to forward request", "node", node, "error", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			w.WriteHeader(http.StatusOK)
			io.Copy(w, resp.Body)
			return
		}
	}

	http.Error(w, "Key not found or all nodes failed", http.StatusNotFound)
}
