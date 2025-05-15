package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"go-sensors-simulator/pkg/models"
)

// CSVStorage gerencia o armazenamento de dados em arquivo CSV
type CSVStorage struct {
	filePath    string
	mu          sync.Mutex
	initialized bool
}

// NewCSVStorage cria uma nova instância de armazenamento CSV
func NewCSVStorage(dataDir string) (*CSVStorage, error) {
	// Verificar se o diretório existe, se não, criar
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório de dados: %w", err)
	}

	// Gerar nome de arquivo baseado na data e hora atual
	timestamp := time.Now().Format("2006-01-02")
	filePath := filepath.Join(dataDir, fmt.Sprintf("sensor_data_%s.csv", timestamp))

	storage := &CSVStorage{
		filePath: filePath,
	}

	return storage, nil
}

// Initialize inicializa o arquivo CSV com o cabeçalho
func (s *CSVStorage) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar se o arquivo já existe
	_, err := os.Stat(s.filePath)
	fileExists := !os.IsNotExist(err)

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("falha ao abrir arquivo CSV: %w", err)
	}
	defer file.Close()

	// Se o arquivo não existir, escrever o cabeçalho
	if !fileExists {
		writer := csv.NewWriter(file)
		defer writer.Flush()

		header := []string{"timestamp", "sensor_id", "sensor_type", "value", "unit"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("falha ao escrever cabeçalho CSV: %w", err)
		}
	}

	s.initialized = true
	return nil
}

// StoreReadings armazena leituras de sensores no arquivo CSV
func (s *CSVStorage) StoreReadings(readings []models.SensorReading) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar se o armazenamento foi inicializado
	if !s.initialized {
		if err := s.Initialize(); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("falha ao abrir arquivo CSV: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, reading := range readings {
		record := []string{
			reading.Timestamp.Format(time.RFC3339),
			reading.SensorID,
			string(reading.SensorType),
			strconv.FormatFloat(reading.Value, 'f', 2, 64),
			reading.Unit,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("falha ao escrever leitura no CSV: %w", err)
		}
	}

	return nil
}
