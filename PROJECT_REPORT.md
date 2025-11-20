# Distributed Key-Value Store - Project Report

## 1. Project Overview
This project is a robust, distributed Key-Value Store built in Go. It is designed to demonstrate core distributed systems concepts including Sharding, Replication, High Availability, and Persistence.

## 2. Architecture
The system follows a **Coordinator-Worker** architecture:

*   **Proxy (Coordinator)**: The entry point for all clients. It does not store data itself but routes requests to the appropriate Data Nodes.
*   **Data Nodes (Workers)**: Responsible for storing data in memory and persisting it to disk.
*   **Client**: A CLI tool to interact with the system via HTTP/JSON.

## 3. Key Features & Mechanisms

### A. Sharding (Partitioning)
*   **Goal**: Distribute data evenly across multiple nodes to scale storage.
*   **Mechanism**: **Consistent Hashing** (Ring Topology).
*   **Implementation**: The Proxy maintains a "Ring" of nodes. Keys are hashed (CRC32) and mapped to the first node with a hash value greater than the key's hash. This minimizes data movement when nodes are added/removed.

### B. Replication & Fault Tolerance
*   **Goal**: Ensure data survives if a node crashes.
*   **Replication Factor**: 3. Every key is stored on the Primary Node + 2 Successor Nodes on the ring.
*   **Quorum Writes (W=2)**: A write is considered successful only if at least 2 out of 3 nodes acknowledge it. This guarantees strong consistency for the latest write.
*   **Read Failover**: If the Primary Node is down, the Proxy automatically tries the replicas.

### C. Persistence (Durability)
*   **Goal**: Prevent data loss on server restart.
*   **Mechanism**: **Write-Ahead Log (WAL)**.
*   **Implementation**: Before updating the in-memory map, every write operation is appended to a log file (`wal-{port}.log`) on disk. On startup, the node replays this log to restore its state.

### D. Containerization
*   **Docker**: The application is containerized using a multi-stage Dockerfile (Golang builder -> Alpine runner).
*   **Docker Compose**: Orchestrates the entire cluster (1 Proxy + 3 Data Nodes) with a single command.

## 4. Request Flow

### Write Path (SET key value)
1.  Client sends request to Proxy.
2.  Proxy hashes the key to find the Primary Node and its 2 Replicas.
3.  Proxy sends the write to all 3 nodes in parallel.
4.  Nodes append to WAL -> Update Memory -> Respond OK.
5.  Proxy waits for 2 "OK" responses (Quorum).
6.  Proxy responds "Success" to Client.

### Read Path (GET key)
1.  Client sends request to Proxy.
2.  Proxy calculates the Primary Node for the key.
3.  Proxy attempts to fetch from Primary.
4.  **Failover**: If Primary is down (timeout/error), Proxy tries the next Replica.
5.  Proxy returns the value to Client.

## 5. How to Run

The entire system can be started with Docker Compose:

```bash
# Start the cluster
docker compose up --build

# Run the client (in another terminal)
./bin/client -server http://localhost:8000 set mykey myvalue
./bin/client -server http://localhost:8000 get mykey
```

## 6. Technology Stack
*   **Language**: Go (Golang) 1.21+
*   **Communication**: HTTP / JSON
*   **Storage**: In-Memory Map + Append-Only File (WAL)
*   **Orchestration**: Docker & Docker Compose
