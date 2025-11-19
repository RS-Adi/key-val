# key-val
distributed key val storage with go..


# DKVS: Distributed Key-Value Store

A high-performance, distributed key-value storage system inspired by DynamoDB and Redis. Built to demonstrate core concepts of distributed systems including concurrency control, sharding, and replication.

## 🚀 Current Status: Phase 1 (Single Node Engine)
Currently, the system operates as a single-node, in-memory storage engine capable of handling concurrent requests via HTTP.

### Tech Stack
* **Language:** Go (Golang) 1.21+
* **Concurrency:** `sync.RWMutex` for thread-safe operations.
* **Communication:** REST API (Standard Library `net/http`).

### 🛠️ Installation & Run
```bash
git clone [https://github.com/yourusername/dkvs.git](https://github.com/yourusername/dkvs.git)
cd dkvs
go run main.go
