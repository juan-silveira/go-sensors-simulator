package simulator

import (
	"math"
	"math/rand"
	"time"

	"go-sensors-simulator/pkg/models"
)

// Inicializar o gerador de números aleatórios com uma semente baseada no tempo atual
func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Simulator representa o simulador de sensores
type Simulator struct {
	configs        []models.SensorConfig
	readings       []models.SensorReading
	lastValues     map[string]float64
	changeCallback func([]models.SensorReading)
	rng            *rand.Rand         // Gerador de números aleatórios dedicado
	driftFactors   map[string]float64 // Fatores de drift para cada sensor
}

// NewSimulator cria um novo simulador de sensores
func NewSimulator(configs []models.SensorConfig, callback func([]models.SensorReading)) *Simulator {
	lastValues := make(map[string]float64)
	driftFactors := make(map[string]float64)

	// Criar um gerador de números aleatórios dedicado
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Inicializa os valores com médias realistas e fatores de drift aleatórios
	for _, config := range configs {
		baseValue := (config.MaxValue + config.MinValue) / 2
		// Adicionar um pouco de aleatoriedade aos valores iniciais
		baseValue += (rng.Float64()*2 - 1) * config.NoiseAmplitude * 2
		lastValues[config.ID] = baseValue

		// Inicializar fatores de drift aleatórios para cada sensor
		driftFactors[config.ID] = (rng.Float64()*2 - 1) * 0.02 // ±2% de drift por ciclo
	}

	return &Simulator{
		configs:        configs,
		readings:       []models.SensorReading{},
		lastValues:     lastValues,
		changeCallback: callback,
		rng:            rng,
		driftFactors:   driftFactors,
	}
}

// Start inicia o simulador
func (s *Simulator) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			s.simulateReadings()
		}
	}()
}

// simulateReadings gera novas leituras simuladas para todos os sensores
func (s *Simulator) simulateReadings() {
	readings := make([]models.SensorReading, 0, len(s.configs))

	// Obter o timestamp atual para todas as leituras
	now := time.Now()
	hourOfDay := float64(now.Hour()) + float64(now.Minute())/60.0 // Hora do dia com fração para transições mais suaves

	for _, config := range s.configs {
		// Calcular um novo valor com base no último valor e adicionando alguma variação
		lastValue := s.lastValues[config.ID]

		// Gerar uma variação aleatória dentro da amplitude de ruído configurada
		noise := (s.rng.Float64()*2 - 1) * config.NoiseAmplitude

		// Adicionar um drift lento ao longo do tempo
		drift := s.driftFactors[config.ID] * lastValue

		// Ocasionalmente, mudar a direção do drift
		if s.rng.Float64() < 0.05 { // 5% de chance de mudar a direção
			s.driftFactors[config.ID] = -s.driftFactors[config.ID] + (s.rng.Float64()*0.01 - 0.005) // Mudar e adicionar pequena variação
		}

		// Calcular um fator sazonal/cíclico para simular padrões reais
		seasonalFactor := 0.0

		// Diferentes padrões para diferentes tipos de sensores
		switch config.Type {
		case models.Temperature:
			// Temperatura mais alta durante o dia, mais baixa à noite
			dailyCycle := math.Sin((hourOfDay / 24) * 2 * math.Pi)
			// Adicionar pequena variação de minuto em minuto
			minuteCycle := math.Sin((float64(now.Minute())/60)*2*math.Pi) * 0.3
			seasonalFactor = dailyCycle*2 + minuteCycle
		case models.Humidity:
			// Umidade inversamente relacionada à temperatura
			dailyCycle := -math.Sin((hourOfDay / 24) * 2 * math.Pi)
			// Adicionar pico de umidade aleatório ocasional (simulando chuva/irrigação)
			if s.rng.Float64() < 0.01 { // 1% de chance
				seasonalFactor = dailyCycle*3 + s.rng.Float64()*10
			} else {
				// Variação normal
				minuteCycle := math.Sin((float64(now.Minute())/30)*2*math.Pi) * 0.8
				seasonalFactor = dailyCycle*3 + minuteCycle
			}
		case models.Light:
			// Luz alta durante o dia, baixa à noite com transições suaves
			if hourOfDay >= 5 && hourOfDay <= 19 {
				// Dia (forma de sino)
				midday := 12.0
				hourFactor := math.Pow((hourOfDay-midday)/7, 2)
				baseFactor := 400 * math.Exp(-hourFactor)
				// Adicionar variação para simular passagem de nuvens
				cloudFactor := 1.0
				if s.rng.Float64() < 0.1 { // 10% de chance de nuvens
					cloudFactor = 0.5 + s.rng.Float64()*0.3
				}
				seasonalFactor = baseFactor * cloudFactor
			} else {
				// Noite - pequena luz para simular luz ambiente/artificial
				seasonalFactor = s.rng.Float64() * 5
			}
		case models.Pressure:
			// Pressão com variação menor e mais lenta + tendências aleatórias
			dailyCycle := math.Sin((hourOfDay/48)*2*math.Pi) * 0.5
			weekCycle := math.Sin((float64(now.Weekday())/7)*2*math.Pi) * 1.0
			seasonalFactor = dailyCycle + weekCycle
		}

		// Calcular novo valor
		newValue := lastValue + noise + seasonalFactor + drift

		// Garantir que o valor está dentro dos limites
		if newValue < config.MinValue {
			newValue = config.MinValue + s.rng.Float64()*config.NoiseAmplitude
		}
		if newValue > config.MaxValue {
			newValue = config.MaxValue - s.rng.Float64()*config.NoiseAmplitude
		}

		// Armazenar o novo valor
		s.lastValues[config.ID] = newValue

		// Criar a leitura do sensor
		reading := models.NewSensorReading(config, newValue)
		readings = append(readings, reading)
	}

	// Atualizar leituras e notificar callback
	s.readings = readings
	if s.changeCallback != nil {
		s.changeCallback(readings)
	}
}

// GetReadings retorna as leituras mais recentes
func (s *Simulator) GetReadings() []models.SensorReading {
	return s.readings
}

// ResetSimulation reinicia a simulação com novos valores aleatórios
func (s *Simulator) ResetSimulation() {
	// Criar um novo gerador com nova semente
	s.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Reinicializar valores e fatores de drift
	for _, config := range s.configs {
		baseValue := (config.MaxValue + config.MinValue) / 2
		baseValue += (s.rng.Float64()*2 - 1) * config.NoiseAmplitude * 3
		s.lastValues[config.ID] = baseValue
		s.driftFactors[config.ID] = (s.rng.Float64()*2 - 1) * 0.02
	}
}
