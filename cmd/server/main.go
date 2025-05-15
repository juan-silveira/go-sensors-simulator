package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-sensors-simulator/configs"
	"go-sensors-simulator/pkg/data"
	"go-sensors-simulator/pkg/models"
	"go-sensors-simulator/pkg/mqtt"
	"go-sensors-simulator/pkg/opcua"
	"go-sensors-simulator/pkg/simulator"
	"go-sensors-simulator/pkg/vpn"
	"go-sensors-simulator/web"
)

func main() {
	// Definir flags
	configPath := flag.String("config", "configs/config.json", "Caminho para o arquivo de configuração")
	flag.Parse()

	// Carregar configuração
	config, err := configs.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Forçar modo de mapeamento para Prosys no modo de leitura
	config.OPCUA.MappingMode = "prosys-read"

	// Criar diretório de dados se não existir
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório de dados: %v", err)
	}

	// Configurar canal para capturar sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Context para controle de shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WaitGroup para aguardar todas as goroutines terminarem
	var wg sync.WaitGroup

	// Inicializar o simulador de sensores
	var mqttClient *mqtt.MQTTClient
	var opcuaClient *opcua.OPCUAClient
	var wireGuardManager *vpn.WireGuardManager
	var csvStorage *data.CSVStorage

	// Inicializar armazenamento CSV
	if config.EnableCSVStore {
		csvStorage, err = data.NewCSVStorage(config.DataDir)
		if err != nil {
			log.Fatalf("Erro ao inicializar armazenamento CSV: %v", err)
		}
		if err := csvStorage.Initialize(); err != nil {
			log.Fatalf("Erro ao inicializar arquivo CSV: %v", err)
		}
	}

	// Inicializar cliente MQTT
	if config.EnableMQTT {
		log.Printf("Conectando ao broker MQTT em: %s", config.MQTT.BrokerURL)
		mqttClient, err = mqtt.NewMQTTClient(config.MQTT)
		if err != nil {
			log.Fatalf("Erro ao criar cliente MQTT: %v", err)
		}

		if err := mqttClient.Connect(); err != nil {
			log.Printf("Aviso: não foi possível conectar ao broker MQTT: %v", err)
		} else {
			log.Println("Conectado ao broker MQTT:", config.MQTT.BrokerURL)
			defer mqttClient.Disconnect()
		}
	}

	// Inicializar cliente OPC-UA
	if config.EnableOPCUA {
		opcuaClient, err = opcua.NewOPCUAClient(config.OPCUA)
		if err != nil {
			log.Fatalf("Erro ao criar cliente OPC-UA: %v", err)
		}

		if err := opcuaClient.Connect(); err != nil {
			log.Printf("Aviso: não foi possível conectar ao servidor OPC-UA: %v", err)
		} else {
			log.Println("Conectado ao servidor OPC-UA:", config.OPCUA.Endpoint)
			defer opcuaClient.Disconnect()
		}
	}

	// Inicializar VPN, se configurado
	if config.EnableVPN {
		// Verificar instalação do WireGuard
		if err := vpn.CheckWireGuardInstallation(); err != nil {
			log.Printf("Aviso: WireGuard não está disponível: %v", err)
		} else {
			wireGuardManager = vpn.NewWireGuardManager(config.WireGuard)

			// Iniciar túnel VPN
			if err := wireGuardManager.Start(); err != nil {
				log.Printf("Aviso: não foi possível iniciar o túnel WireGuard: %v", err)
			} else {
				log.Println("Túnel WireGuard iniciado")
				defer wireGuardManager.Stop()
			}
		}
	}

	// Criar função de callback para processar leituras de sensores
	readingsHandler := func(readings []models.SensorReading) {
		// Armazenar em CSV a cada intervalo configurado
		if config.EnableCSVStore {
			if err := csvStorage.StoreReadings(readings); err != nil {
				log.Printf("Erro ao armazenar leituras em CSV: %v", err)
			}
		}

		// Publicar via MQTT
		if config.EnableMQTT && mqttClient != nil {
			if err := mqttClient.PublishReadings(readings); err != nil {
				log.Printf("Erro ao publicar leituras via MQTT: %v", err)
			}
		}

		// Publicar via OPC-UA
		if config.EnableOPCUA && opcuaClient != nil {
			if err := opcuaClient.WriteReadings(readings); err != nil {
				log.Printf("Erro ao escrever leituras via OPC-UA: %v", err)
			}
		}
	}

	// Criar simulador
	sim := simulator.NewSimulator(config.Sensors, readingsHandler)

	// Iniciar simulador em uma goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		sim.Start(config.SimulationRate)

		// Loop para manter o simulador rodando até receber sinal de parada
		<-ctx.Done()
		log.Println("Parando simulador...")
	}()

	// Inicializar servidor web
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ServerPort),
		Handler: web.NewRouter(sim, config),
	}

	// Iniciar servidor web em uma goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Iniciando servidor web na porta %d", config.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor web: %v", err)
		}
	}()

	// Aguardar sinal para encerrar
	<-sigChan
	log.Println("Recebido sinal de encerramento. Desligando servidor...")

	// Criar um contexto com timeout para shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Desligar o servidor HTTP graciosamente
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Erro ao desligar servidor: %v", err)
	}

	// Cancelar o contexto principal para sinalizar que as goroutines devem encerrar
	cancel()

	// Aguardar todas as goroutines terminarem
	wg.Wait()
	log.Println("Servidor encerrado com sucesso")
}
