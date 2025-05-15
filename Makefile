.PHONY: all build run clean templ deps test

# Variáveis
APP_NAME = go-sensors-simulator
BUILD_DIR = build
CMD_DIR = cmd/server
CONFIG_DIR = configs
DATA_DIR = data

# Compilação
all: clean deps templ build

build:
	@echo "Compilando aplicação..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go

run:
	@echo "Executando aplicação..."
	go run $(CMD_DIR)/main.go

# Dependências
deps:
	@echo "Instalando dependências..."
	go mod tidy

# Templates
templ:
	@echo "Gerando templates Templ..."
	templ generate

# Testes
test:
	@echo "Executando testes..."
	go test ./... -v

# Limpeza
clean:
	@echo "Limpando arquivos gerados..."
	rm -rf $(BUILD_DIR)
	mkdir -p $(DATA_DIR)

# Criar diretórios
init:
	@echo "Criando diretórios do projeto..."
	mkdir -p $(CMD_DIR) $(CONFIG_DIR) $(DATA_DIR) web/static/js web/static/css

# Gerar chaves WireGuard
wireguard-keys:
	@echo "Gerando chaves WireGuard..."
	mkdir -p keys
	wg genkey > keys/server_private.key
	wg pubkey < keys/server_private.key > keys/server_public.key
	wg genkey > keys/client_private.key
	wg pubkey < keys/client_private.key > keys/client_public.key
	@echo "Chaves geradas em 'keys/'"

# Executar MQTT broker local (requer mosquitto)
mqtt-broker:
	@echo "Iniciando broker MQTT local..."
	mosquitto -v

# Ajuda
help:
	@echo "Comandos disponíveis:"
	@echo "  make build         - Compila a aplicação"
	@echo "  make run           - Executa a aplicação"
	@echo "  make deps          - Instala dependências"
	@echo "  make templ         - Gera templates Templ"
	@echo "  make test          - Executa testes"
	@echo "  make clean         - Remove arquivos gerados"
	@echo "  make init          - Cria diretórios do projeto"
	@echo "  make wireguard-keys- Gera chaves para WireGuard"
	@echo "  make mqtt-broker   - Inicia um broker MQTT local" 