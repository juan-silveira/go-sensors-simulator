package web

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"

	"go-sensors-simulator/configs"
	"go-sensors-simulator/pkg/simulator"
	"go-sensors-simulator/web/templates"
)

// Router gerencia as rotas HTTP
type Router struct {
	simulator       *simulator.Simulator
	config          configs.AppConfig
	templateHandler *templates.Handler
}

// NewRouter cria um novo roteador HTTP
func NewRouter(sim *simulator.Simulator, config configs.AppConfig) *Router {
	return &Router{
		simulator:       sim,
		config:          config,
		templateHandler: templates.NewHandler(sim, config),
	}
}

// ServeHTTP implementa a interface http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Configurar headers básicos para todas as respostas
	w.Header().Set("Server", "Cannabis Sensor Simulator")

	// Roteamento baseado no caminho
	switch req.URL.Path {
	case "/", "/index.html", "/dashboard":
		r.templateHandler.HandleDashboard(w, req)
	case "/api/sensors":
		r.handleAPIGetSensors(w, req)
	case "/api/readings":
		r.handleAPIGetReadings(w, req)
	case "/api/reset-simulation":
		r.handleAPIResetSimulation(w, req)
	default:
		// Verificar se está tentando acessar um recurso estático
		if req.URL.Path == "/static/" || filepath.HasPrefix(req.URL.Path, "/static/") {
			// Servir arquivos estáticos
			fs := http.StripPrefix("/static/", http.FileServer(http.Dir("web/static")))
			fs.ServeHTTP(w, req)
			return
		}

		// Verificar se é um arquivo estático
		if filepath.Ext(req.URL.Path) != "" {
			// Caminho com extensão, tentar servir como arquivo estático
			http.ServeFile(w, req, filepath.Join("web", req.URL.Path))
			return
		}
		// Se não for um arquivo estático, retornar 404
		http.NotFound(w, req)
	}
}

// handleAPIGetSensors retorna a lista de sensores em formato JSON
func (r *Router) handleAPIGetSensors(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Serializar sensores como JSON
	if err := json.NewEncoder(w).Encode(r.config.Sensors); err != nil {
		log.Printf("Erro ao serializar sensores para JSON: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
	}
}

// handleAPIGetReadings retorna as leituras mais recentes em formato JSON
func (r *Router) handleAPIGetReadings(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Obter leituras do simulador
	readings := r.simulator.GetReadings()

	// Serializar leituras como JSON
	if err := json.NewEncoder(w).Encode(readings); err != nil {
		log.Printf("Erro ao serializar leituras para JSON: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
	}
}

// handleAPIResetSimulation reseta a simulação com novos valores aleatórios
func (r *Router) handleAPIResetSimulation(w http.ResponseWriter, req *http.Request) {
	// Verificar se é uma requisição POST
	if req.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Resetar a simulação
	r.simulator.ResetSimulation()

	// Retornar resposta de sucesso
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "success", "message": "Simulação resetada com sucesso"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Erro ao serializar resposta para JSON: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
	}
}
