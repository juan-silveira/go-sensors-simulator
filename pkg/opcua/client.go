package opcua

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-sensors-simulator/pkg/models"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

// OPCUAConfig contém as configurações do cliente OPC-UA
type OPCUAConfig struct {
	Endpoint    string
	Policy      string
	Mode        string
	Certificate string
	PrivateKey  string
	Username    string
	Password    string
	Namespace   uint16
	// Adicionado modo de mapeamento para permitir compatibilidade com diferentes servidores
	MappingMode string // "auto", "prosys" ou "prosys-read"
}

// OPCUAClient gerencia a comunicação via OPC-UA
type OPCUAClient struct {
	client         *opcua.Client
	config         OPCUAConfig
	connected      bool
	nodeIDs        map[string]*ua.NodeID // Armazena apenas os NodeIDs
	useProsysNodes bool                  // Indica se deve usar os nós padrão do Prosys Simulation Server
	readOnlyMode   bool                  // Indica se está no modo apenas leitura (para servidores como Prosys)
}

// NewOPCUAClient cria um novo cliente OPC-UA
func NewOPCUAClient(config OPCUAConfig) (*OPCUAClient, error) {
	// Verificar se deve usar mapeamento para o Prosys Simulation Server
	useProsysNodes := config.MappingMode == "prosys" || config.MappingMode == "prosys-read"
	readOnlyMode := config.MappingMode == "prosys-read"

	log.Printf("Configuração OPC-UA: Endpoint=%s, MappingMode=%s, UseProsys=%v, ReadOnly=%v, Namespace=%d",
		config.Endpoint, config.MappingMode, useProsysNodes, readOnlyMode, config.Namespace)

	return &OPCUAClient{
		config:         config,
		nodeIDs:        make(map[string]*ua.NodeID),
		useProsysNodes: useProsysNodes,
		readOnlyMode:   readOnlyMode,
	}, nil
}

// Connect estabelece conexão com o servidor OPC-UA
func (c *OPCUAClient) Connect() error {
	// Configurar opções do cliente OPC-UA
	var securityMode ua.MessageSecurityMode

	// Converter string para MessageSecurityMode
	switch c.config.Mode {
	case "None":
		securityMode = ua.MessageSecurityModeNone
	case "Sign":
		securityMode = ua.MessageSecurityModeSign
	case "SignAndEncrypt":
		securityMode = ua.MessageSecurityModeSignAndEncrypt
	default:
		securityMode = ua.MessageSecurityModeNone
	}

	opts := []opcua.Option{
		opcua.SecurityPolicy(c.config.Policy),
		opcua.SecurityMode(securityMode),
		opcua.CertificateFile(c.config.Certificate),
		opcua.PrivateKeyFile(c.config.PrivateKey),
	}

	// Adicionar autenticação se fornecida
	if c.config.Username != "" {
		opts = append(opts, opcua.AuthUsername(c.config.Username, c.config.Password))
	}

	// Criar cliente
	client, err := opcua.NewClient(c.config.Endpoint, opts...)
	if err != nil {
		return fmt.Errorf("falha ao criar cliente OPC-UA: %w", err)
	}

	// Conectar ao servidor
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("falha ao conectar ao servidor OPC-UA: %w", err)
	}

	c.client = client
	c.connected = true
	log.Println("Conectado ao servidor OPC-UA:", c.config.Endpoint)

	// Pré-configurar mapeamento se estiver usando Prosys
	if c.useProsysNodes {
		c.setupProsysMapping()

		// Se estiver no modo de leitura apenas, iniciar uma goroutine para ler periodicamente
		if c.readOnlyMode {
			go c.startReadingLoop()
		}
	}

	return nil
}

// startReadingLoop inicia um loop de leitura periódica dos nós
func (c *OPCUAClient) startReadingLoop() {
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		if !c.connected {
			return
		}

		// Ler todos os nós mapeados
		for key, nodeID := range c.nodeIDs {
			value, err := c.readNodeValue(nodeID)
			if err != nil {
				log.Printf("Erro ao ler nó %s (%s): %v", key, nodeID, err)
				continue
			}

			log.Printf("Leitura OPC-UA: %s = %v", key, value)
		}
	}
}

// readNodeValue lê o valor de um nó do servidor OPC-UA
func (c *OPCUAClient) readNodeValue(nodeID *ua.NodeID) (interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("cliente OPC-UA não está conectado")
	}

	// Criar contexto
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ler o valor do nó
	req := &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{
			{
				NodeID:       nodeID,
				AttributeID:  ua.AttributeIDValue,
				DataEncoding: &ua.QualifiedName{},
			},
		},
	}

	resp, err := c.client.Read(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler valor: %w", err)
	}

	// Verificar se temos resultados
	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("nenhum resultado retornado")
	}

	// Verificar status
	if resp.Results[0].Status != ua.StatusOK {
		return nil, fmt.Errorf("falha ao ler valor, status: %v", resp.Results[0].Status)
	}

	// Retornar o valor
	return resp.Results[0].Value.Value(), nil
}

// setupProsysMapping configura o mapeamento de sensores para os nós do Prosys Simulation Server
func (c *OPCUAClient) setupProsysMapping() {
	log.Println("Configurando mapeamento para nós do Prosys Simulation Server")

	// Mapeamento básico entre tipos de sensores e nós de simulação do Prosys
	// temperature -> Sinusoid (oscilação de temperatura)
	tempNodeID, err := ua.ParseNodeID("ns=3;i=1004") // Sinusoid
	if err != nil {
		log.Printf("Erro ao parsear NodeID da temperatura: %v", err)
	} else {
		c.nodeIDs["temperature-temp001"] = tempNodeID
		log.Printf("Mapeado: temperature-temp001 -> %s (Sinusoid)", tempNodeID)
	}

	// humidity -> Random (variação aleatória de umidade)
	humNodeID, err := ua.ParseNodeID("ns=3;i=1002") // Random
	if err != nil {
		log.Printf("Erro ao parsear NodeID da umidade: %v", err)
	} else {
		c.nodeIDs["humidity-hum001"] = humNodeID
		log.Printf("Mapeado: humidity-hum001 -> %s (Random)", humNodeID)
	}

	// light -> Sawtooth (padrão de luz com variação diurna)
	lightNodeID, err := ua.ParseNodeID("ns=3;i=1003") // Sawtooth
	if err != nil {
		log.Printf("Erro ao parsear NodeID da luminosidade: %v", err)
	} else {
		c.nodeIDs["light-light001"] = lightNodeID
		log.Printf("Mapeado: light-light001 -> %s (Sawtooth)", lightNodeID)
	}

	// pressure -> Constant (pressão mais estável)
	pressNodeID, err := ua.ParseNodeID("ns=3;i=1007") // Constant
	if err != nil {
		log.Printf("Erro ao parsear NodeID da pressão: %v", err)
	} else {
		c.nodeIDs["pressure-press001"] = pressNodeID
		log.Printf("Mapeado: pressure-press001 -> %s (Constant)", pressNodeID)
	}

	log.Println("Mapeamento para nós do Prosys Simulation Server configurado com sucesso")
}

// Disconnect desconecta do servidor OPC-UA
func (c *OPCUAClient) Disconnect() {
	if c.connected && c.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.client.Close(ctx); err != nil {
			log.Printf("Erro ao desconectar do servidor OPC-UA: %v\n", err)
		} else {
			log.Println("Desconectado do servidor OPC-UA")
		}
		c.connected = false
	}
}

// getOrCreateNodeID obtém ou cria um NodeID para um sensor específico
func (c *OPCUAClient) getOrCreateNodeID(reading models.SensorReading) (*ua.NodeID, error) {
	if !c.connected {
		return nil, fmt.Errorf("cliente OPC-UA não está conectado")
	}

	// Criar chave única para o sensor
	key := fmt.Sprintf("%s-%s", reading.SensorType, reading.SensorID)

	// Verificar se o NodeID já existe
	if nodeID, exists := c.nodeIDs[key]; exists {
		log.Printf("Usando nó existente para %s: %s", key, nodeID)
		return nodeID, nil
	}

	// Se estiver usando Prosys mas o nó não foi encontrado no mapeamento pré-configurado
	if c.useProsysNodes {
		log.Printf("ERRO: Sensor %s não encontrado no mapeamento Prosys", key)
		return nil, fmt.Errorf("sensor %s não mapeado para nenhum nó do Prosys Simulation Server", key)
	}

	// Caso contrário, criar ID do nó baseado no namespace e no sensor (modo automático)
	// Exemplo: ns=1;s=Sensors.Temperature.Temp001
	nodeIDString := fmt.Sprintf("ns=%d;s=Sensors.%s.%s", c.config.Namespace, reading.SensorType, reading.SensorID)
	log.Printf("Criando novo nó automático para %s: %s", key, nodeIDString)
	nodeID, err := ua.ParseNodeID(nodeIDString)
	if err != nil {
		return nil, fmt.Errorf("falha ao analisar NodeID: %w", err)
	}

	// Armazenar NodeID
	c.nodeIDs[key] = nodeID

	return nodeID, nil
}

// WriteReading escreve uma leitura de sensor no servidor OPC-UA
func (c *OPCUAClient) WriteReading(reading models.SensorReading) error {
	if !c.connected {
		return fmt.Errorf("cliente OPC-UA não está conectado")
	}

	// Se estiver no modo apenas leitura, pular escrita
	if c.readOnlyMode {
		return nil
	}

	// Obter ou criar NodeID para o sensor
	nodeID, err := c.getOrCreateNodeID(reading)
	if err != nil {
		return err
	}

	// Criar valor para escrita
	v, err := ua.NewVariant(reading.Value)
	if err != nil {
		return fmt.Errorf("falha ao criar variante para valor: %w", err)
	}

	// Escrever valor no nó
	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      nodeID,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask:    ua.DataValueValue | ua.DataValueSourceTimestamp,
					Value:           v,
					SourceTimestamp: time.Now(),
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.Write(ctx, req)
	if err != nil {
		return fmt.Errorf("falha ao escrever valor: %w", err)
	}

	// Verificar resultado da operação
	if resp.Results[0] != ua.StatusOK {
		return fmt.Errorf("falha ao escrever valor, status: %v", resp.Results[0])
	}

	return nil
}

// WriteReadings escreve várias leituras de sensores no servidor OPC-UA
func (c *OPCUAClient) WriteReadings(readings []models.SensorReading) error {
	if !c.connected {
		return fmt.Errorf("cliente OPC-UA não está conectado")
	}

	// Se estiver no modo apenas leitura, pular escrita
	if c.readOnlyMode {
		return nil
	}

	for _, reading := range readings {
		if err := c.WriteReading(reading); err != nil {
			return err
		}
	}

	return nil
}
