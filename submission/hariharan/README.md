# Config Service — Infrastructure Assignment

## Overview
A Go-based configuration management service deployed on Kubernetes with PostgreSQL. The system is fully containerized and infrastructure is provisioned using Terraform and Helm.

---

## Architecture

Client → Config Service (Kubernetes Deployment)  
       → PostgreSQL (Helm via Terraform, same namespace)

---

## Tech Stack

- Go (HTTP API)
- Docker (multi-stage build)
- Kubernetes (Kind cluster)
- Terraform + Helm (OCI registry)
- PostgreSQL

---

## Features

- Create and fetch configuration records
- Layered architecture (handler → service → repository)
- Database retry logic for startup reliability
- Health check endpoint (/ping)
- Kubernetes liveness and readiness probes
- Secrets managed via Kubernetes Secret
- Configuration via ConfigMap
- Fully reproducible local setup

---

## Setup Instructions

\`\`\`bash
cd submission/hariharan

make build
make kind-load
make deploy
make tf-init
make tf-apply
make port-forward
\`\`\`

---

## API Endpoints

- GET /ping
- POST /configs
- GET /configs/{id}

---

## Example Usage

\`\`\`bash
curl -X POST http://localhost:8080/configs \
-H "Content-Type: application/json" \
-d '{
  "id": "cfg_1",
  "host": "localhost",
  "port": 8080,
  "app_name": "config-service",
  "log_level": "INFO"
}'
\`\`\`

\`\`\`bash
curl http://localhost:8080/configs/cfg_1
\`\`\`

---

## Design Decisions

- Multi-stage Docker build for minimal image size
- Clean layered architecture for maintainability
- Kubernetes namespace isolation (config-service)
- PostgreSQL deployed inside cluster using Helm via Terraform
- OCI-based Helm registry for consistent chart resolution
- Separation of concerns using ConfigMap and Secret

---

## Tradeoffs

- Persistence disabled for faster local testing
- Port-forward used instead of ingress for simplicity
- Minimal test coverage (only handler test included)

---

## Testing

\`\`\`bash
go test ./...
\`\`\`

---

## Notes

- Terraform state and provider binaries are excluded via .gitignore
- Local Kubernetes cluster created using Kind
- Helm charts installed via OCI registry for stability
