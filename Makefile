COMPOSE := docker compose -f deploy/dev/docker-compose.yaml
DEV_DS := deploy/dev/provisioning/datasources/wazuh.yaml
DEV_DS_EXAMPLE := deploy/dev/provisioning/datasources/wazuh.yaml.example
LAB_DS_EXAMPLE := deploy/wazuh-lab/examples/datasource.yaml.example

.PHONY: build backend dev dev-config lab-dev-config dev-down k8s-forward lab-up lab-down lab-reset lab-connect

build:
	npm run build
	go run github.com/magefile/mage@latest -v build:linux

backend:
	go run github.com/magefile/mage@latest -v build:linux

dev-config:
	@if [ ! -f "$(DEV_DS)" ]; then \
		cp "$(DEV_DS_EXAMPLE)" "$(DEV_DS)"; \
		echo ""; \
		echo "Created $(DEV_DS) for Kubernetes port-forward dev (Path B)."; \
		echo "For local Wazuh lab (Path A), run: make lab-dev-config"; \
		echo "Edit passwords (from kubectl secrets) then run: make dev"; \
		echo ""; \
	fi

# Provisioning for local Wazuh Docker lab — use with make lab-up && make dev && make lab-connect
lab-dev-config:
	cp "$(LAB_DS_EXAMPLE)" "$(DEV_DS)"
	@echo "Created $(DEV_DS) for local Wazuh lab (wazuh.manager / wazuh.indexer)."

dev: build dev-config
	-$(COMPOSE) down --remove-orphans 2>/dev/null
	$(COMPOSE) up --build

dev-down:
	$(COMPOSE) down

# Requires kubectl access to your Wazuh namespace. Keep running while using make dev.
k8s-forward:
	chmod +x deploy/dev/k8s-forward.sh
	./deploy/dev/k8s-forward.sh

lab-up:
	chmod +x deploy/wazuh-lab/setup.sh deploy/wazuh-lab/connect-grafana.sh
	./deploy/wazuh-lab/setup.sh

lab-down:
	cd deploy/wazuh-lab/lab/wazuh-docker/single-node && \
		(docker compose -f docker-compose.podman.yml down 2>/dev/null || docker compose down)

lab-reset:
	./deploy/wazuh-lab/setup.sh reset

lab-connect:
	./deploy/wazuh-lab/connect-grafana.sh
