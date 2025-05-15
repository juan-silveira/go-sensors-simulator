package models

import (
	"time"
)

// SensorType define o tipo de sensor
type SensorType string

const (
	Light       SensorType = "light"       // Luminosidade (lux)
	Humidity    SensorType = "humidity"    // Umidade (%)
	Temperature SensorType = "temperature" // Temperatura (°C)
	Pressure    SensorType = "pressure"    // Pressão (hPa)
)

// SensorReading representa uma leitura de um sensor
type SensorReading struct {
	SensorID   string     `json:"sensor_id"`
	SensorType SensorType `json:"sensor_type"`
	Value      float64    `json:"value"`
	Unit       string     `json:"unit"`
	Timestamp  time.Time  `json:"timestamp"`
}

// SensorConfig contém as configurações de um sensor
type SensorConfig struct {
	ID             string     `json:"id"`
	Type           SensorType `json:"type"`
	MinValue       float64    `json:"min_value"`
	MaxValue       float64    `json:"max_value"`
	NoiseAmplitude float64    `json:"noise_amplitude"` // Amplitude do ruído para simulação
	Unit           string     `json:"unit"`
}

// NewSensorReading cria uma nova leitura de sensor
func NewSensorReading(config SensorConfig, value float64) SensorReading {
	return SensorReading{
		SensorID:   config.ID,
		SensorType: config.Type,
		Value:      value,
		Unit:       config.Unit,
		Timestamp:  time.Now(),
	}
}
