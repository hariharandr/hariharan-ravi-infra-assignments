# Config Service

A Go HTTP service that stores and retrieves configuration records in PostgreSQL.
Deployed on local Kubernetes using Kind, provisioned with Terraform and Helm.

---

## Architecture

curl (your laptop)
→ port-forward tunnel
→ Kubernetes Service (config-service)
→ Go App Pod

- reads DATABASE_URL from Kubernetes Secret
- reads PORT from Kubernetes ConfigMap
- runs schema migration on startup automatically
- connects to PostgreSQL via cluster DNS
  → PostgreSQL Pod
- provisioned by Terraform + Helm
- runs in same namespace (config-service)

---

## Prerequisites

Install these before running anything:

| Tool      | Version | Install                                |
| --------- | ------- | -------------------------------------- |
| Docker    | 20+     | https://docs.docker.com/get-docker/    |
| Kind      | 0.20+   | `brew install kind`                    |
| kubectl   | 1.28+   | `brew install kubectl`                 |
| Terraform | 1.8+    | `brew install hashicorp/tap/terraform` |
| Helm      | 3.14+   | `brew install helm`                    |
| Go        | 1.26+   | https://go.dev/dl/                     |

Verify all are installed:

```bash
cd submission/hariharan/
make env-check
```

---

## Setup — Run in this exact order

### 1. Build Docker image

```bash
make build
```

### 2. Create Kubernetes cluster

```bash
make cluster-up
```

### 3. Load image into cluster

```bash
make load
```

> Kind cannot access your local Docker images automatically. This copies the image in.

### 4. Provision infrastructure

```bash
make infra
```

> Terraform creates: namespace, DB credentials secret, PostgreSQL via Helm.
> This takes 1–2 minutes.

### 5. Wait for PostgreSQL to be ready

```bash
make status
```

> Re-run until `postgres-postgresql-0` shows `1/1 Running`.

### 6. Deploy the app

```bash
make deploy
```

> The app runs schema migration automatically on startup. No manual SQL needed.

### 7. Wait for app to be ready

```bash
make status
```

> Re-run until `config-service-xxx` shows `1/1 Running`.

### 8. Expose the service

```bash
make port-forward
```

> Keep this terminal open. Open a new terminal for all curl commands.

---

## Validate the API

```bash
# Health check
curl http://localhost:8080/ping

# Create a config
curl -X POST http://localhost:8080/configs \
  -H "Content-Type: application/json" \
  -d '{"id":"cfg_1","host":"localhost","port":8080,"app_name":"myapp","log_level":"INFO"}'

# Retrieve the config
curl http://localhost:8080/configs/cfg_1

# Run all smoke tests at once
make smoke-test
```

---

## Unit Tests

```bash
make test
```

Tests run without any infrastructure. Uses in-memory fake service — no database needed.

---

## Teardown

```bash
# Ctrl+C to stop port-forward first
make cluster-down
```

Deletes the entire cluster and everything inside it.

---

## Rebuild after code change

```bash
make build
make load
make restart
```

---

## Observability — Liveness and Readiness Probes

The app is health-checked by Kubernetes automatically via `/ping`.

### Show probe configuration

```bash
make probe-check
```

Expected output:
Liveness: http-get http://:8080/ping delay=10s timeout=1s period=10s #success=1 #failure=3
Readiness: http-get http://:8080/ping delay=10s timeout=1s period=5s #success=1 #failure=3

- **Liveness** — if `/ping` fails 3 times → Kubernetes restarts the container
- **Readiness** — if `/ping` fails → Kubernetes stops sending traffic (but does not restart)

### Show live probe events

```bash
make events
```

### Live readiness demo

Terminal 1 — watch pods:

```bash
kubectl get pods -n config-service -w
```

Terminal 2 — simulate failure and recovery:

```bash
# Remove pods
kubectl scale deployment config-service --replicas=0 -n config-service

# Traffic fails — no ready pods
curl http://localhost:8080/ping

# Bring back
kubectl scale deployment config-service --replicas=1 -n config-service
```

Watch Terminal 1: pod goes `0/1` (readiness failing, no traffic) → `1/1` (ready, traffic allowed).

---

## Dev Container (pre-installed environment)

Use this if you need a clean environment with all tools pre-installed.

```bash
# Build dev image once (takes ~3 minutes)
docker build -f Dockerfile.dev -t config-service-dev:latest .

# Start clean container (no project files mounted)
docker run -it --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --name dev-env \
  config-service-dev:latest

# Inside container — clone and run
git clone https://github.com/hariharandr/hariharan-ravi-infra-assignments.git
cd hariharan-ravi-infra-assignments/submission/hariharan
make env-check
make build
make cluster-up
make load
make infra
make status
make deploy
make status
make port-forward
```

> The container uses your laptop's Docker daemon via socket mount.
> Kind clusters appear on your host Docker.

---

## API Reference

### GET /ping

Returns `pong`. Used for liveness and readiness probes.

### POST /configs

Creates or updates a config record (upsert).

Request:

```json
{
  "id": "cfg_1",
  "host": "localhost",
  "port": 8080,
  "app_name": "my-service",
  "log_level": "INFO"
}
```

Response: saved config with timestamps.

Error responses:

- `400` — missing or invalid fields
- `500` — internal error

### GET /configs/:id

Returns config by ID.

Error responses:

- `404` — config not found

---

## Design Decisions

**Migration on startup** — App runs `CREATE TABLE IF NOT EXISTS` automatically.
Safe to run multiple times. No separate migration step or manual SQL needed.

**Secrets via Terraform** — DB credentials secret is provisioned by Terraform.
Single source of truth. App manifest references that secret by name.

**DB retry on startup** — App retries DB connection 10 times with 2s delay.
Handles the Kubernetes race where app container starts before PostgreSQL is ready.

**Layered architecture** — handler → service → repository. Each layer has one responsibility.
Handler never writes SQL. Repository never writes HTTP status codes.

**Persistence disabled** — PostgreSQL runs without a PersistentVolumeClaim for local dev.
Data is lost on pod restart. In production, enable persistent storage.

**Port-forward over Ingress** — Simpler for local development. In production, use Ingress
or a LoadBalancer service.

---

## Known Limitations and Production Improvements

- Enable PostgreSQL persistent storage (PersistentVolumeClaim)
- Use AWS Secrets Manager or Vault instead of Kubernetes Secrets
- Add proper database migration tooling (golang-migrate)
- Add resource limits and requests on pods
- Add Prometheus metrics endpoint
- Set up CI/CD pipeline (GitHub Actions → build → push → deploy)
- Readiness probe should check DB connectivity, not just HTTP liveness

---

## Responsible AI Usage

Built with guidance from Claude and ChatGPT to learn Kubernetes and Terraform.
All commands were run and debugged personally. All design decisions were
understood and chosen deliberately.
