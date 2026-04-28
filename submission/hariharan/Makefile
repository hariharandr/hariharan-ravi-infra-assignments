# ============================================================
# Config Service — Local Development Makefile
# ============================================================
# Fresh setup order:
#   make env-check
#   make build
#   make cluster-up
#   make load
#   make infra
#   make deploy
#   make port-forward
#   make smoke-test
# ============================================================

CLUSTER_NAME=config-service
NAMESPACE=config-service
IMAGE=config-service:local

## Verify all required tools are installed
env-check:
	@echo "Checking dependencies..."
	@docker version --format 'Docker: {{.Client.Version}}' 2>/dev/null || echo "MISSING: docker"
	@kind version 2>/dev/null || echo "MISSING: kind"
	@kubectl version --client 2>/dev/null | head -1 || echo "MISSING: kubectl"
	@terraform version 2>/dev/null | head -1 || echo "MISSING: terraform"
	@helm version --short 2>/dev/null || echo "MISSING: helm"
	@go version 2>/dev/null || echo "MISSING: go"
	@echo "All checks done."

## Build the Docker image
build:
	docker build -t $(IMAGE) .

## Create the local Kind cluster
cluster-up:
	kind create cluster --name $(CLUSTER_NAME)
	kubectl config use-context kind-$(CLUSTER_NAME)

## Delete the local Kind cluster (full teardown)
cluster-down:
	kind delete cluster --name $(CLUSTER_NAME)

## Load local Docker image into Kind cluster
load:
	kind load docker-image $(IMAGE) --name $(CLUSTER_NAME)

## Apply Terraform — creates namespace, secret, and PostgreSQL
infra:
	cd infra/terraform && terraform init && terraform apply -auto-approve

## Deploy app into Kubernetes
deploy:
	kubectl apply -f k8s/

## Watch pods in config-service namespace
status:
	kubectl get pods -n $(NAMESPACE)

## Restart app pods (use after rebuild + load)
restart:
	kubectl config use-context kind-$(CLUSTER_NAME)
	kubectl rollout restart deployment config-service -n $(NAMESPACE)
	kubectl rollout status deployment/config-service -n $(NAMESPACE)

## Stream app logs
logs:
	kubectl logs -f deployment/config-service -n $(NAMESPACE)

## Forward local port 8080 to the service (keep terminal open)
port-forward:
	kubectl port-forward service/config-service 8080:80 -n $(NAMESPACE)

## Run Go unit tests
test:
	go test ./... -v -count=1

## Run smoke test against live service (port-forward must be running)
smoke-test:
	@echo "--- health check ---"
	@curl -sf http://localhost:8080/ping && echo " OK"
	@echo "--- create config ---"
	@curl -sf -X POST http://localhost:8080/configs \
		-H "Content-Type: application/json" \
		-d '{"id":"cfg_smoke","host":"localhost","port":8080,"app_name":"smoke-test","log_level":"INFO"}' && echo " OK"
	@echo "--- get config ---"
	@curl -sf http://localhost:8080/configs/cfg_smoke && echo " OK"
	@echo "--- all smoke tests passed ---"

## Show liveness and readiness probe config
probe-check:
	kubectl describe pod -l app=config-service -n $(NAMESPACE) | grep -A 6 "Liveness\|Readiness"

## Show recent Kubernetes events
events:
	kubectl get events -n $(NAMESPACE) --sort-by='.lastTimestamp'

.PHONY: env-check build cluster-up cluster-down load infra deploy status restart logs port-forward test smoke-test probe-check events
