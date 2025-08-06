# Имя проекта
PROJECT_NAME := archiver_service
BINARY_NAME := archiver

# Версия (можно переопределить через командную строке)
VERSION ?= dev

# Директории
BUILD_DIR := bin
DIST_DIR := dist

# Поддерживаемые платформы
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
## windows/amd64  # раскомментируйте для поддержки Windows

# Цвета для вывода
RED    := \033[31m
GREEN  := \033[32m
YELLOW := \033[33m
BLUE   := \033[34m
RESET  := \033[0m

# --- Основные цели ---
.PHONY: help
help: ## Показать помощь
	@echo "$(BLUE)=== $(PROJECT_NAME) Makefile ===$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "} {printf "$(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'

.PHONY: build
build: ## Собрать для текущей платформы
	@echo "$(GREEN)Building for current platform...$(RESET)"
	go build -o $(BUILD_DIR)/$(BINARY_NAME) \
		-ldflags "-X main.version=$(VERSION)" \
		./cmd/$(PROJECT_NAME)

.PHONY: clean
clean: ## Очистить сборки
	@echo "$(YELLOW)Cleaning build directories...$(RESET)"
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	mkdir -p $(BUILD_DIR) $(DIST_DIR)

.PHONY: build-all
build-all: clean ## Собрать для всех платформ
	@echo "$(GREEN)Building for all platforms...$(RESET)"
	@$(foreach PLATFORM,$(PLATFORMS), \
		OS_ARCH=$(PLATFORM); \
		OS=$${OS_ARCH%/*}; \
		ARCH=$${OS_ARCH#*/}; \
		echo "$(BLUE)Building for $$OS/$$ARCH...$(RESET)"; \
		GOOS=$$OS GOARCH=$$ARCH \
		go build -o $(DIST_DIR)/$(BINARY_NAME)_$$OS\_$$ARCH \
			-ldflags "-X main.version=$(VERSION)" \
			./cmd/$(PROJECT_NAME); \
	)

.PHONY: release
release: build-all ## Создать релиз (требует VERSION)
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "$(RED)Error: VERSION must be set for release (например: make release VERSION=v1.0.0)$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating release $(VERSION)...$(RESET)"
	# Создать директорию для релиза
	@mkdir -p $(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)
	# Переместить бинарники
	@mv $(DIST_DIR)/$(BINARY_NAME)_linux_amd64   $(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)/
	@mv $(DIST_DIR)/$(BINARY_NAME)_linux_arm64   $(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)/
	@mv $(DIST_DIR)/$(BINARY_NAME)_darwin_amd64 $(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)/
	@mv $(DIST_DIR)/$(BINARY_NAME)_darwin_arm64 $(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)/
	# Переименовать бинарники с версией
	@cd $(DIST_DIR)/$(PROJECT_NAME)_$(VERSION) && \
		mv $(BINARY_NAME)_linux_amd64   $(PROJECT_NAME)_linux_amd64_$(VERSION) && \
		mv $(BINARY_NAME)_linux_arm64   $(PROJECT_NAME)_linux_arm64_$(VERSION) && \
		mv $(BINARY_NAME)_darwin_amd64 $(PROJECT_NAME)_darwin_amd64_$(VERSION) && \
		mv $(BINARY_NAME)_darwin_arm64 $(PROJECT_NAME)_darwin_arm64_$(VERSION)
	# Создать tar.gz архивы
	@cd $(DIST_DIR) && \
		tar -czf $(PROJECT_NAME)_$(VERSION)_linux_amd64.tar.gz   $(PROJECT_NAME)_$(VERSION)/$(PROJECT_NAME)_linux_amd64_$(VERSION) && \
		tar -czf $(PROJECT_NAME)_$(VERSION)_linux_arm64.tar.gz   $(PROJECT_NAME)_$(VERSION)/$(PROJECT_NAME)_linux_arm64_$(VERSION) && \
		tar -czf $(PROJECT_NAME)_$(VERSION)_darwin_amd64.tar.gz $(PROJECT_NAME)_$(VERSION)/$(PROJECT_NAME)_darwin_amd64_$(VERSION) && \
		tar -czf $(PROJECT_NAME)_$(VERSION)_darwin_arm64.tar.gz $(PROJECT_NAME)_$(VERSION)/$(PROJECT_NAME)_darwin_arm64_$(VERSION)
	# Удалить временную директорию
	@rm -rf $(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)
	@echo "$(GREEN)Release $(VERSION) created successfully!$(RESET)"
	@echo "$(BLUE)Release files:$(RESET)"
	@ls -la $(DIST_DIR)/$(PROJECT_NAME)_$(VERSION)_*

.PHONY: install
install: ## Установить бинарник в $$GOPATH/bin
	@echo "$(GREEN)Installing to GOPATH...$(RESET)"
	go install ./cmd/$(PROJECT_NAME)

.PHONY: run
run: ## Запустить приложение
	@echo "$(GREEN)Running application...$(RESET)"
	go run ./cmd/$(PROJECT_NAME)

.PHONY: test
test: ## Запустить тесты
	@echo "$(GREEN)Running tests...$(RESET)"
	go test -v ./...

.PHONY: test-cover
test-cover: ## Запустить тесты с покрытием
	@echo "$(GREEN)Running tests with coverage...$(RESET)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(BLUE)Coverage report saved to coverage.html$(RESET)"

.PHONY: lint
lint: ## Проверить код линтером
	@echo "$(GREEN)Running linter...$(RESET)"
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "$(YELLOW)Installing golangci-lint...$(RESET)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

.PHONY: fmt
fmt: ## Форматировать код
	@echo "$(GREEN)Formatting code...$(RESET)"
	go fmt ./...

.PHONY: vet
vet: ## Проверить код vet
	@echo "$(GREEN)Running go vet...$(RESET)"
	go vet ./...

.PHONY: check
check: vet lint test ## Проверить код полностью

.PHONY: list-platforms
list-platforms: ## Показать поддерживаемые платформы
	@echo "$(BLUE)Supported platforms:$(RESET)"
	@$(foreach PLATFORM,$(PLATFORMS), echo "  - $(PLATFORM)")

# Показать помощь по умолчанию
.DEFAULT_GOAL := help
