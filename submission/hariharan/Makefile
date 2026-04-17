build:
	docker build -t config-service:local .

kind-load:
	kind load docker-image config-service:local --name config-service

deploy:
	kubectl apply -f k8s/

tf-init:
	cd infra/terraform && terraform init

tf-apply:
	cd infra/terraform && terraform apply -auto-approve

restart:
	kubectl rollout restart deployment config-service -n config-service

port-forward:
	kubectl port-forward service/config-service 8080:80 -n config-service

logs:
	kubectl logs -f deployment/config-service -n config-service