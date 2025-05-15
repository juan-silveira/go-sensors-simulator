package configs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"go-sensors-simulator/pkg/models"
	"go-sensors-simulator/pkg/mqtt"
	"go-sensors-simulator/pkg/opcua"
	"go-sensors-simulator/pkg/vpn"
)

// AppConfig é a configuração principal do aplicativo
type AppConfig struct {
	// Configurações gerais
	ServerPort      int           `json:"server_port"`
	DataDir         string        `json:"data_dir"`
	SimulationRate  time.Duration `json:"simulation_rate"`  // Intervalo entre leituras em segundos
	StorageInterval time.Duration `json:"storage_interval"` // Intervalo para salvar em CSV

	// Sensores
	Sensors []models.SensorConfig `json:"sensors"`

	// Configurações MQTT
	MQTT mqtt.MQTTConfig `json:"mqtt"`

	// Configurações OPC-UA
	OPCUA opcua.OPCUAConfig `json:"opcua"`

	// Configurações VPN
	WireGuard vpn.WireGuardConfig `json:"wireguard"`

	// Habilitar/Desabilitar componentes
	EnableMQTT     bool `json:"enable_mqtt"`
	EnableOPCUA    bool `json:"enable_opcua"`
	EnableVPN      bool `json:"enable_vpn"`
	EnableCSVStore bool `json:"enable_csv_store"`
}

// DefaultConfig retorna a configuração padrão
func DefaultConfig() AppConfig {
	return AppConfig{
		ServerPort:      8080,
		DataDir:         "./data",
		SimulationRate:  1 * time.Second,
		StorageInterval: 5 * time.Second,
		EnableMQTT:      true,
		EnableOPCUA:     true,
		EnableVPN:       false,
		EnableCSVStore:  true,
		Sensors: []models.SensorConfig{
			{
				ID:             "temp001",
				Type:           models.Temperature,
				MinValue:       18.0,
				MaxValue:       30.0,
				NoiseAmplitude: 0.3,
				Unit:           "°C",
			},
			{
				ID:             "hum001",
				Type:           models.Humidity,
				MinValue:       40.0,
				MaxValue:       75.0,
				NoiseAmplitude: 1.0,
				Unit:           "%",
			},
			{
				ID:             "light001",
				Type:           models.Light,
				MinValue:       0.0,
				MaxValue:       1000.0,
				NoiseAmplitude: 10.0,
				Unit:           "lux",
			},
			{
				ID:             "press001",
				Type:           models.Pressure,
				MinValue:       990.0,
				MaxValue:       1020.0,
				NoiseAmplitude: 0.5,
				Unit:           "hPa",
			},
		},
		MQTT: mqtt.MQTTConfig{
			BrokerURL:  "tcp://localhost:1883",
			ClientID:   "cannabis-sensor-sim",
			Username:   "",
			Password:   "",
			TopicBase:  "cannabis/sensors",
			QoS:        1,
			Retained:   false,
			CACertPath: "",
		},
		OPCUA: opcua.OPCUAConfig{
			Endpoint:    "opc.tcp://localhost:4840",
			Policy:      "None",
			Mode:        "None",
			Certificate: "",
			PrivateKey:  "",
			Username:    "",
			Password:    "",
			Namespace:   2,
			MappingMode: "prosys-read",
		},
		WireGuard: vpn.WireGuardConfig{
			InterfaceName: "wg0",
			Address:       "10.0.0.1/24",
			ListenPort:    51820,
			AllowedIPs:    "10.0.0.0/24",
			ConfigPath:    "/etc/wireguard/wg0.conf",
		},
	}
}

// LoadConfig carrega a configuração de um arquivo
func LoadConfig(filepath string) (AppConfig, error) {
	config := DefaultConfig()

	// Verificar se o arquivo existe
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		// Se o arquivo não existe, salva a configuração padrão
		SaveConfig(filepath, config)
		return config, nil
	}

	// Ler arquivo
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	// Desserializar
	err = json.Unmarshal(data, &config)
	return config, err
}

// SaveConfig salva a configuração em um arquivo
func SaveConfig(filepath string, config AppConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, data, 0644)
}
