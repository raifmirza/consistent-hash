# Consistent Hashing Demo

A small HTTP load balancer backed by a consistent-hash ring, periodic health checks, and a backend discovery API.

## Project layout

- `cmd/loader`: load balancer and discovery API (`:8080`)
- `cmd/backend`: demo backend server (configurable port)
- `internal/hash`: consistent-hash ring
- `internal/balancer`: reverse-proxy load balancer
- `internal/health`: health probing and ring updates
- `internal/backend`: backend registry and reverse proxies
- `internal/discovery`: register and remove backends over HTTP
- `internal/node`: backend node model

## How routing works

The load balancer hashes a routing key and looks it up on the ring. Repeated requests with the same key go to the same backend while the backend set is stable.

The default key extractor uses the client IP:

1. `X-Forwarded-For` header, if present
2. otherwise the request's `RemoteAddr`

Because every request from localhost shares the same IP, use `X-Forwarded-For` in the demo to simulate different clients:

```bash
curl -H "X-Forwarded-For: user-42" http://127.0.0.1:8080/
curl -H "X-Forwarded-For: user-99" http://127.0.0.1:8080/
```

Each backend responds with JSON showing which server handled the request:

```json
{"server_id":"server-8082","path":"/","method":"GET","requests":1}
```

## Run the demo

Start three backends in separate terminals:

```bash
go run ./cmd/backend -port 8081
go run ./cmd/backend -port 8082
go run ./cmd/backend -port 8083
```

Start the load balancer:

```bash
go run ./cmd/loader
```

This starts:

- backend servers on `127.0.0.1:8081`, `127.0.0.1:8082`, and `127.0.0.1:8083`
- the load balancer on `127.0.0.1:8080`

Wait about 10 seconds after starting the loader so the health checker can mark backends healthy and add them to the ring.

## Try it

Send repeated requests with the same routing key:

```bash
curl -H "X-Forwarded-For: user-42" http://127.0.0.1:8080/
curl -H "X-Forwarded-For: user-42" http://127.0.0.1:8080/
curl -H "X-Forwarded-For: user-99" http://127.0.0.1:8080/
```

The same key should return the same `server_id`.

## Simulate an unhealthy backend

Stop one backend process (for example, the server on `:8082`).

The health checker probes each backend's `/health` endpoint every 5 seconds. After 3 consecutive failures, the backend is removed from the ring. After 2 consecutive successes, it is added back.

Restart the stopped backend to show recovery.

## Discovery API

Register a backend dynamically:

```bash
curl -X POST http://127.0.0.1:8080/backends \
  -H "Content-Type: application/json" \
  -d '{"id":"server-8084","address":"http://localhost:8084"}'
```

Remove a backend:

```bash
curl -X DELETE http://127.0.0.1:8080/backends/server-8084
```

## Test

```bash
go test ./...
```
