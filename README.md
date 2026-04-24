# Log Ingestion & Query System

A high-performance distributed log ingestion and querying system built in **Go**, designed for scalable log processing using **Kafka** as a buffer and **ClickHouse** as the analytics database.

---

## 🚀 Overview

This project implements a full pipeline for log ingestion, processing, storage, and querying:

* Logs are received via HTTP API in Go
* Logs are pushed into **Kafka** as an intermediate buffer
* Go consumers read logs from Kafka in batches concurrently
* Batched logs are flushed into **ClickHouse** for fast querying
* A REST API and Web UI allow querying logs with filters and regex support

---

## 🧱 Architecture

```
Client → Go API → Kafka → Go Consumer (Batch Processor) → ClickHouse → Query API / Web UI
```

### Flow Breakdown:

1. **Ingestion Layer**

   * `/api/ingest` receives log entries via HTTP POST
   * Logs are published to Kafka topics

2. **Streaming Layer**

   * Kafka acts as a durable buffer and decouples ingestion from storage

3. **Processing Layer**

   * Go workers consume logs from Kafka
   * Logs are processed in concurrent batches
   * Batch flush writes to ClickHouse efficiently

4. **Storage Layer**

   * ClickHouse stores logs with indexed columns for fast querying

5. **Query Layer**

   * `/api/logs` supports advanced filtering and regex search
   * `/logs` provides a web-based UI for interactive querying

---

## ⚙️ Features

### 📥 Ingestion

* `POST /api/ingest`
* Accepts single log entries
* Immediately publishes to Kafka

### 📊 Query API

* `GET /api/logs`
* Supports:

  * Time range filtering
  * Log level filtering (info, error, debug, etc.)
  * Resource filtering
  * Regex matching on message field

### 🌐 Web UI

* `/logs`
* Built using Go `html/template`
* Provides:

  * Filter form
  * Live query results table
  * Regex search support

### ⚡ Performance

* Kafka-based buffering for resilience
* Concurrent batch processing in Go
* ClickHouse optimized columnar storage with indexing

---

## 🗄️ Storage (ClickHouse)

Logs are stored in ClickHouse with indexed fields such as:

* level
* resourceId
* traceId
* message

This enables fast analytical queries over large log volumes.

---

## 🧵 Kafka Usage

Kafka is used as a **durable intermediate queue**:

* Ensures no log loss during spikes
* Decouples ingestion from storage
* Enables horizontal scaling of consumers

---

## 🐳 Docker Setup

The system is fully containerized using Docker:

### Services:

* Go Application (API + Consumer)
* Kafka + KRaft
* ClickHouse

---

## ▶️ Running the Project

### 1. Clone the repository

```bash
git clone https://github.com/dagger021/log-ingestion-query-system.git
cd log-ingestion-query-system
```

### 2. Start all services

```bash
docker-compose up --build
```

---

## 📡 API Endpoints

### Ingest Log

```http
POST /api/ingest
```

**Request Body:**

```json
{
  "level":"error",
  "message":"Failed to connect to DB",
  "resourceId":"server-1234",
  "timestamp":"2023-09-15T08:00:00Z",
  "traceId":"abc-xyz-123",
  "spanId":"span-456",
  "commit":"5e5342f",
  "metadata":{
    "parentResourceId":"server-0987"
  }
}
```

---

### Query Logs

```http
GET /api/logs
```

**Query Parameters:**

* `level`
* `resourceId`
* `traceId`
* `message` & `token`
* `from` & `to` for periodic logs
* `limit` & `offset` for pagination

Example:

```
/api/logs?level=error&service=auth&message=timeout
```

Response:

```json
[
  {
    "id": "b733ef82-3e12-4e3b-a18d-83a678fd95b9",
    "level": "error",
    "message": "Failed to connect to DB",
    "resourceId": "server-1234",
    "timestamp": "2023-09-15T08:00:00Z",
    "traceId": "abc-xyz-123",
    "spanId": "span-456",
    "commit": "5e5342f",
    "metadata": {
      "parentResourceId": "server-0987"
    }
  }
]
```

---

## 🖥️ Web UI

Visit:

```
http://localhost:3000/logs
```

Features:

* Filter logs via form inputs
* Regex search on message field
* Paginated results

---

## 🧑‍💻 Tech Stack

* **Go** (Backend, API, Kafka consumer, templates)
* **Chi Router** (HTTP routing)
* **Kafka** (Message queue)
* **ClickHouse** (Log storage)
* **Docker / Docker Compose** (container orchestration)

---

## 📦 Design Goals

* High throughput log ingestion
* Fault-tolerant pipeline
* Horizontally scalable consumers
* Fast analytical queries on large datasets
* Simple operational deployment via Docker
