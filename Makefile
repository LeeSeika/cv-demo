# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help build-int 2fa-start image-storage-demo-start image-storage-demo-stop image-storage-demo-status oauth2-demo-start oauth2-demo-stop oauth2-demo-status

.DEFAULT_GOAL := help

OAUTH2_DEMO_DIR := tmp/oauth2-demo
AUTH_SERVER_PID := $(OAUTH2_DEMO_DIR)/auth-server.pid
APP_SERVER_PID := $(OAUTH2_DEMO_DIR)/app-server.pid
RESOURCE_SERVER_PID := $(OAUTH2_DEMO_DIR)/resource-server.pid
IMAGE_STORAGE_DEMO_DIR := tmp/image-storage-demo
IMAGE_STORAGE_INTEGRATION_PID := $(IMAGE_STORAGE_DEMO_DIR)/integration.pid
IMAGE_STORAGE_APP_PID := $(IMAGE_STORAGE_DEMO_DIR)/app.pid

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build-int: ## build integration backend
	go build -o int-bin integration/main.go

2fa-start: ## start 2fa example server on :9100
	go run ./biz/examples/2fa

image-storage-demo-start: ## start image storage demo and required integration services
	@mkdir -p $(IMAGE_STORAGE_DEMO_DIR)
	@if [ -f $(IMAGE_STORAGE_INTEGRATION_PID) ] && kill -0 $$(cat $(IMAGE_STORAGE_INTEGRATION_PID)) 2>/dev/null; then echo "integration is already running (pid=$$(cat $(IMAGE_STORAGE_INTEGRATION_PID)))"; else \
		echo "starting integration services on :4000 and :4001"; \
		nohup go run ./integration > $(IMAGE_STORAGE_DEMO_DIR)/integration.log 2>&1 & \
		echo $$! > $(IMAGE_STORAGE_INTEGRATION_PID); \
	fi
	@if [ -f $(IMAGE_STORAGE_APP_PID) ] && kill -0 $$(cat $(IMAGE_STORAGE_APP_PID)) 2>/dev/null; then echo "image storage demo is already running (pid=$$(cat $(IMAGE_STORAGE_APP_PID)))"; else \
		echo "starting image storage demo on :2020"; \
		IMAGE_STORAGE_DB_DSN=sqlite://$(IMAGE_STORAGE_DEMO_DIR)/image-storage.db \
			nohup go run ./biz/design/image/storage > $(IMAGE_STORAGE_DEMO_DIR)/app.log 2>&1 & \
		echo $$! > $(IMAGE_STORAGE_APP_PID); \
	fi
	@echo "image storage demo started"
	@echo "integration s3:  http://localhost:4000"
	@echo "integration gcs: http://localhost:4001"
	@echo "demo app:        http://localhost:2020"
	@echo "logs:            $(IMAGE_STORAGE_DEMO_DIR)/*.log"

image-storage-demo-stop: ## stop image storage demo and required integration services
	@mkdir -p $(IMAGE_STORAGE_DEMO_DIR)
	@for file in $(IMAGE_STORAGE_APP_PID) $(IMAGE_STORAGE_INTEGRATION_PID); do \
		if [ -f $$file ]; then \
			pid=$$(cat $$file); \
			if kill -0 $$pid 2>/dev/null; then \
				echo "stopping process $$pid"; \
				kill $$pid; \
			else \
				echo "process $$pid is not running"; \
			fi; \
			rm -f $$file; \
		fi; \
	done

image-storage-demo-status: ## show image storage demo and integration status
	@for item in "integration $(IMAGE_STORAGE_INTEGRATION_PID) http://localhost:4000,http://localhost:4001" "image-storage-demo $(IMAGE_STORAGE_APP_PID) http://localhost:2020"; do \
		set -- $$item; \
		name=$$1; pidfile=$$2; url=$$3; \
		if [ -f $$pidfile ] && kill -0 $$(cat $$pidfile) 2>/dev/null; then \
			echo "$$name: running (pid=$$(cat $$pidfile)) $$url"; \
		else \
			echo "$$name: stopped $$url"; \
		fi; \
	done

oauth2-demo-start: ## start oauth2 demo servers: auth-server, app-server, resource-server
	@mkdir -p $(OAUTH2_DEMO_DIR)
	@if [ -f $(AUTH_SERVER_PID) ] && kill -0 $$(cat $(AUTH_SERVER_PID)) 2>/dev/null; then echo "auth-server is already running (pid=$$(cat $(AUTH_SERVER_PID)))"; else \
		echo "starting auth-server on :3360"; \
		AUTH_SERVER_DB_DSN=sqlite://$(OAUTH2_DEMO_DIR)/auth-server.db PORT=3360 \
			nohup go run ./biz/design/oauth2/auth-server > $(OAUTH2_DEMO_DIR)/auth-server.log 2>&1 & \
		echo $$! > $(AUTH_SERVER_PID); \
	fi
	@if [ -f $(RESOURCE_SERVER_PID) ] && kill -0 $$(cat $(RESOURCE_SERVER_PID)) 2>/dev/null; then echo "resource-server is already running (pid=$$(cat $(RESOURCE_SERVER_PID)))"; else \
		echo "starting resource-server on :3370"; \
		RESOURCE_SERVER_DB_DSN=sqlite://$(OAUTH2_DEMO_DIR)/resource-server.db RESOURCE_SERVER_PORT=3370 \
			nohup go run ./biz/design/oauth2/resource-server > $(OAUTH2_DEMO_DIR)/resource-server.log 2>&1 & \
		echo $$! > $(RESOURCE_SERVER_PID); \
	fi
	@if [ -f $(APP_SERVER_PID) ] && kill -0 $$(cat $(APP_SERVER_PID)) 2>/dev/null; then echo "app-server is already running (pid=$$(cat $(APP_SERVER_PID)))"; else \
		echo "starting app-server on :9094"; \
		nohup go run ./biz/design/oauth2/app-server > $(OAUTH2_DEMO_DIR)/app-server.log 2>&1 & \
		echo $$! > $(APP_SERVER_PID); \
	fi
	@echo "oauth2 demo started"
	@echo "auth-server:     http://localhost:3360"
	@echo "resource-server: http://localhost:3370"
	@echo "app-server:      http://localhost:9094"
	@echo "logs:            $(OAUTH2_DEMO_DIR)/*.log"

oauth2-demo-stop: ## stop oauth2 demo servers
	@mkdir -p $(OAUTH2_DEMO_DIR)
	@for file in $(APP_SERVER_PID) $(RESOURCE_SERVER_PID) $(AUTH_SERVER_PID); do \
		if [ -f $$file ]; then \
			pid=$$(cat $$file); \
			if kill -0 $$pid 2>/dev/null; then \
				echo "stopping process $$pid"; \
				kill $$pid; \
			else \
				echo "process $$pid is not running"; \
			fi; \
			rm -f $$file; \
		fi; \
	done

oauth2-demo-status: ## show oauth2 demo server status
	@for item in "auth-server $(AUTH_SERVER_PID) http://localhost:3360" "resource-server $(RESOURCE_SERVER_PID) http://localhost:3370" "app-server $(APP_SERVER_PID) http://localhost:9094"; do \
		set -- $$item; \
		name=$$1; pidfile=$$2; url=$$3; \
		if [ -f $$pidfile ] && kill -0 $$(cat $$pidfile) 2>/dev/null; then \
			echo "$$name: running (pid=$$(cat $$pidfile)) $$url"; \
		else \
			echo "$$name: stopped $$url"; \
		fi; \
	done
