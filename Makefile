.PHONY: help install run dev build clean docker-up docker-down test fmt vet

# Variables
BINARY_NAME=gokeki
MAIN_PATH=./main.go

# Ayuda por defecto
help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Instalación de dependencias
install: ## Instalar dependencias de Go
	go mod download
	go mod tidy

# Ejecutar la aplicación
run: ## Ejecutar la aplicación
	go run $(MAIN_PATH)

# Modo desarrollo con hot reload (requiere air)
dev: ## Ejecutar en modo desarrollo (requiere air: go install github.com/cosmtrek/air@latest)
	air

# Compilar la aplicación
build: ## Compilar la aplicación
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Limpiar archivos compilados
clean: ## Limpiar archivos compilados
	rm -rf bin/
	go clean

# Formatear código
fmt: ## Formatear código Go
	go fmt ./...

# Verificar código
vet: ## Verificar código Go con go vet
	go vet ./...

# Ejecutar tests
test: ## Ejecutar tests
	go test ./... -v

# Docker - Levantar servicios
docker-up: ## Levantar servicios de Docker (Redis)
	docker-compose up -d

# Docker - Detener servicios
docker-down: ## Detener servicios de Docker
	docker-compose down

# Docker - Ver logs
docker-logs: ## Ver logs de Redis
	docker-compose logs -f redis

# Docker - Ver logs de la aplicación
docker-logs-app: ## Ver logs de la aplicación
	docker-compose logs -f app

# Docker - Build de la imagen
docker-build: ## Construir imagen Docker
	./scripts/build.sh latest

# Docker - Build optimizado con argumentos
docker-build-release: ## Construir imagen Docker para release
	./scripts/build.sh $(VERSION) --push

# Docker - Build multi-platform 
docker-build-multi: ## Construir imagen Docker multi-platform
	docker buildx build --platform linux/amd64,linux/arm64 -t gokeki:latest --push .

# Docker - Ejecutar stack completa
docker-run-full: ## Ejecutar aplicación y Redis con Docker
	docker-compose up -d

# Docker - Ejecutar solo aplicación (asume Redis corriendo)
docker-run-app: ## Ejecutar solo la aplicación con Docker
	docker-compose up -d app

# Verificar que Redis esté funcionando
check-redis: ## Verificar conexión a Redis
	@echo "Verificando conexión a Redis..."
	@docker exec gokeki-redis redis-cli ping || echo "Redis no está disponible"

# Setup completo
setup: install docker-up ## Setup completo del proyecto
	@echo "Esperando que Redis esté listo..."
	@sleep 5
	@make check-redis
	@echo ""
	@echo "✅ Setup completado!"
	@echo "Para ejecutar la aplicación:"
	@echo "  make run"
	@echo ""
	@echo "Para ver todos los comandos disponibles:"
	@echo "  make help"
