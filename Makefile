BINARY_NAME := vless-server
BUILD_DIR := build
DOCKER_IMAGE := vless-server

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/vless-server

.PHONY: run
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-debug
run-debug: build
	$(BUILD_DIR)/$(BINARY_NAME) -debug

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run:
	docker run -d -p 443:443 -p 443:443/udp \
		-v $(PWD)/data:/app/data \
		--name $(DOCKER_IMAGE) $(DOCKER_IMAGE)

.PHONY: docker-stop
docker-stop:
	docker stop $(DOCKER_IMAGE) || true
	docker rm $(DOCKER_IMAGE) || true

.PHONY: install
install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo mkdir -p /etc/vless-server
	sudo cp configs/env.properties /etc/vless-server/.env
	sudo mkdir -p /var/lib/vless-server
	@echo "Creating systemd service..."
	sudo cp systemd/vless-server.service /etc/systemd/system/
	sudo systemctl daemon-reload
	@echo "Installation complete"
	@echo "Run: sudo systemctl enable vless-server"
	@echo "     sudo systemctl start vless-server"

.PHONY: uninstall
uninstall:
	sudo systemctl stop vless-server || true
	sudo systemctl disable vless-server || true
	sudo rm -f /etc/systemd/system/vless-server.service
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -rf /etc/vless-server
	sudo systemctl daemon-reload

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: help
help:
	@echo "Commands:"
	@echo "  make build          - Build binary"
	@echo "  make run            - Build and run"
	@echo "  make run-debug      - Build and run with debug"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make docker-stop    - Stop Docker container"
	@echo "  make install        - Install to system"
	@echo "  make uninstall      - Remove from system"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deps           - Download dependencies"

.DEFAULT_GOAL := help
