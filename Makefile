# Определяем переменные
GO := go
# SERVICES := api-gateway
SERVICES := api-gateway service-news service-comments service-censor

# Правило для сборки всех служб
.PHONY: all
all: build

# Правило для сборки каждой службы
.PHONY: build
build: $(SERVICES)

$(SERVICES):
	cd $@ && $(GO) build -o $$@ .

# Правило для запуска main.go в каждой службе
.PHONY: run
run:
	@for service in $(SERVICES); do \
		echo "Running main.go for $$service..."; \
		cd $$service && $(GO) run main.go; \
  		cd -; \
	done

# Правило для запуска main.go в каждой службе
.PHONY: mod
mod:
	@for service in $(SERVICES); do \
		echo "Go mod tidy for $$service..."; \
		cd $$service && $(GO) mod tidy; \
  		cd -; \
	done

# Правило для тестирования всех служб
.PHONY: test
test:
	@for service in $(SERVICES); do \
		echo "Running tests for $$service..."; \
		cd $$service && $(GO) test ./...; \
		cd -; \
	done

# Правило для запуска тестов с покрытием
.PHONY: test-cover
test-cover:
	@for service in $(SERVICES); do \
		echo "Running tests with coverage for $$service..."; \
		cd $$service && $(GO) test -cover ./...; \
		cd -; \
	done

# Правило для чистки
.PHONY: clean
clean:
	@for service in $(SERVICES); do \
		cd $$service && $(GO) clean; \
		cd -; \
	done

# Правило для получения зависимостей
.PHONY: get-deps
get-deps:
	@for service in $(SERVICES); do \
		cd $$service && $(GO) mod tidy; \
		cd -; \
	done

# Справка
.PHONY: help
help:
	@echo "Makefile для управления Go проектами"
	@echo "Доступные команды:"
	@echo "  all           - Сборка всех служб"
	@echo "  build         - Сборка служб"
	@echo "  run           - Запуск main.go для всех служб"
	@echo "  mod           - Запуск go mod tidy для всех служб"
	@echo "  test          - Запуск тестов"
	@echo "  test-cover    - Запуск тестов с покрытием"
	@echo "  clean         - Очистка сборки"
	@echo "  get-deps      - Получение зависимостей"
	@echo "  help          - Показать эту справку"