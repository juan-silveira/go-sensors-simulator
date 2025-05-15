# Simulador de Sensores para Cultivo de Cannabis Medicinal

Um sistema de simulação de sensores para monitoramento de ambientes de cultivo de cannabis medicinal, implementado em Go com interface web usando Templ.

## Funcionalidades

- **Simulação de Sensores:**
  - Temperatura (°C)
  - Umidade (%)
  - Luminosidade (lux)
  - Pressão Atmosférica (hPa)

- **Armazenamento de Dados:**
  - Armazenamento em CSV a cada 5 segundos
  - Dashboard web para visualização em tempo real

- **Comunicação:**
  - Protocolo MQTT para IoT
  - Protocolo OPC-UA para integração industrial
  - Conexão VPN (WireGuard) para acesso remoto seguro

## Requisitos

- Go 1.16+
- WireGuard (para funcionalidade VPN)
- Navegador web moderno para o dashboard

## Estrutura do Projeto

```
go-sensors-simulator/
├── cmd/
│   └── server/         # Aplicação principal
├── configs/            # Configurações do aplicativo
├── pkg/
│   ├── data/           # Armazenamento em CSV
│   ├── models/         # Modelos de dados
│   ├── mqtt/           # Cliente MQTT
│   ├── opcua/          # Cliente OPC-UA
│   ├── simulator/      # Simulador de sensores
│   └── vpn/            # Cliente VPN
├── web/
│   ├── static/         # Arquivos estáticos
│   └── templates/      # Templates Templ
└── data/               # Diretório para armazenamento de dados
```

## Instalação

1. Clone o repositório:
   ```
   git clone https://github.com/seu-usuario/go-sensors-simulator.git
   cd go-sensors-simulator
   ```

2. Instale as dependências:
   ```
   go mod tidy
   ```

3. Compile os templates Templ:
   ```
   templ generate
   ```

## Configuração

A configuração do aplicativo está no arquivo `configs/config.json`. Se o arquivo não existir, será criado automaticamente com valores padrão.

### Parâmetros de configuração:

- `server_port`: Porta do servidor web
- `simulation_rate`: Taxa de atualização das leituras em segundos
- `storage_interval`: Intervalo para armazenamento em CSV
- `mqtt`: Configurações do MQTT broker
- `opcua`: Configurações do servidor OPC-UA
- `wireguard`: Configurações da VPN WireGuard

## Uso

1. Inicie o servidor:
   ```
   go run cmd/server/main.go
   ```

2. Acesse o dashboard em:
   ```
   http://localhost:8080
   ```

## Configuração da VPN WireGuard

Para configurar a VPN WireGuard para acesso remoto:

1. Instale o WireGuard:
   ```
   sudo apt install wireguard
   ```

2. Gere as chaves para o servidor e cliente:
   ```
   wg genkey > server_private.key
   wg pubkey < server_private.key > server_public.key
   wg genkey > client_private.key
   wg pubkey < client_private.key > client_public.key
   ```

3. Configure o arquivo `configs/config.json` com os detalhes da VPN.

## Dashboard

O dashboard web inclui:

- Visualização em tempo real dos valores dos sensores
- Gráficos de tendência para todos os sensores
- Indicadores de status para conexões MQTT, OPC-UA e VPN

## Licença

Este projeto é distribuído sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes. 