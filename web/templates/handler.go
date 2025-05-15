package templates

import (
	"context"
	"log"
	"net/http"

	"go-sensors-simulator/configs"
	"go-sensors-simulator/pkg/simulator"
)

// Handler gerencia os templates templ
type Handler struct {
	simulator *simulator.Simulator
	config    configs.AppConfig
}

// NewHandler cria um novo handler de templates
func NewHandler(sim *simulator.Simulator, config configs.AppConfig) *Handler {
	return &Handler{
		simulator: sim,
		config:    config,
	}
}

// HandleDashboard gerencia o endpoint do dashboard
func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	// Renderizar o template do dashboard
	component := Dashboard(h.config.Sensors)
	err := component.Render(context.Background(), w)
	if err != nil {
		log.Printf("Erro ao renderizar dashboard: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
	}
}
