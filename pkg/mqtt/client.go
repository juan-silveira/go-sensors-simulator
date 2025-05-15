package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-sensors-simulator/pkg/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTConfig contém as configurações do cliente MQTT
type MQTTConfig struct {
	BrokerURL  string
	ClientID   string
	Username   string
	Password   string
	TopicBase  string
	QoS        byte
	Retained   bool
	CACertPath string
}

// MQTTClient gerencia a comunicação via MQTT
type MQTTClient struct {
	client    mqtt.Client
	config    MQTTConfig
	connected bool
}

// defaultHandler é a função de callback padrão para mensagens MQTT
var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("MQTT: Recebido: %s de %s\n", msg.Payload(), msg.Topic())
}

// NewMQTTClient cria um novo cliente MQTT
func NewMQTTClient(config MQTTConfig) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.BrokerURL)
	opts.SetClientID(config.ClientID)

	if config.Username != "" {
		opts.SetUsername(config.Username)
		opts.SetPassword(config.Password)
	}

	opts.SetDefaultPublishHandler(defaultHandler)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Minute)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)

	// Definir funções de callback
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("Conectado ao broker MQTT")
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Conexão perdida com o broker MQTT: %v\n", err)
	})

	client := mqtt.NewClient(opts)

	return &MQTTClient{
		client: client,
		config: config,
	}, nil
}

// Connect estabelece conexão com o broker MQTT
func (m *MQTTClient) Connect() error {
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("falha ao conectar ao broker MQTT: %w", token.Error())
	}
	m.connected = true
	return nil
}

// Disconnect desconecta do broker MQTT
func (m *MQTTClient) Disconnect() {
	if m.connected {
		m.client.Disconnect(250)
		m.connected = false
	}
}

// PublishReading publica uma leitura de sensor no tópico apropriado
func (m *MQTTClient) PublishReading(reading models.SensorReading) error {
	if !m.connected {
		return fmt.Errorf("cliente MQTT não está conectado")
	}

	// Criar tópico baseado no tipo de sensor e ID
	topic := fmt.Sprintf("%s/%s/%s", m.config.TopicBase, reading.SensorType, reading.SensorID)

	// Converter leitura para JSON
	payload, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("falha ao serializar leitura para JSON: %w", err)
	}

	// Publicar mensagem
	token := m.client.Publish(topic, m.config.QoS, m.config.Retained, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("falha ao publicar mensagem MQTT: %w", token.Error())
	}

	return nil
}

// PublishReadings publica várias leituras de sensores
func (m *MQTTClient) PublishReadings(readings []models.SensorReading) error {
	if !m.connected {
		return fmt.Errorf("cliente MQTT não está conectado")
	}

	for _, reading := range readings {
		if err := m.PublishReading(reading); err != nil {
			return err
		}
	}

	return nil
}
