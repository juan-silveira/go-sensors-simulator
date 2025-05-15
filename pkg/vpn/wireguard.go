package vpn

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WireGuardConfig contém as configurações do túnel WireGuard
type WireGuardConfig struct {
	InterfaceName string
	PrivateKey    string
	PublicKey     string
	Address       string
	ListenPort    int
	PeerPublicKey string
	PeerEndpoint  string
	AllowedIPs    string
	ConfigPath    string
}

// WireGuardManager gerencia o túnel WireGuard
type WireGuardManager struct {
	config WireGuardConfig
	active bool
}

// NewWireGuardManager cria um novo gerenciador de túnel WireGuard
func NewWireGuardManager(config WireGuardConfig) *WireGuardManager {
	return &WireGuardManager{
		config: config,
		active: false,
	}
}

// GenerateConfig gera um arquivo de configuração WireGuard
func (w *WireGuardManager) GenerateConfig() error {
	// Conteúdo da configuração WireGuard
	configContent := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
ListenPort = %d

[Peer]
PublicKey = %s
AllowedIPs = %s
Endpoint = %s
PersistentKeepalive = 25
`,
		w.config.PrivateKey,
		w.config.Address,
		w.config.ListenPort,
		w.config.PeerPublicKey,
		w.config.AllowedIPs,
		w.config.PeerEndpoint,
	)

	// Garantir que o diretório existe
	configDir := filepath.Dir(w.config.ConfigPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("falha ao criar diretório de configuração: %w", err)
	}

	// Escrever arquivo de configuração
	if err := ioutil.WriteFile(w.config.ConfigPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("falha ao escrever arquivo de configuração: %w", err)
	}

	log.Printf("Arquivo de configuração WireGuard criado em: %s\n", w.config.ConfigPath)
	return nil
}

// Start inicia o túnel WireGuard
func (w *WireGuardManager) Start() error {
	if w.active {
		return fmt.Errorf("túnel WireGuard já está ativo")
	}

	// Verificar se o arquivo de configuração existe
	if _, err := os.Stat(w.config.ConfigPath); os.IsNotExist(err) {
		if err := w.GenerateConfig(); err != nil {
			return err
		}
	}

	// Iniciar o túnel WireGuard usando wg-quick
	cmd := exec.Command("wg-quick", "up", w.config.ConfigPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("falha ao iniciar túnel WireGuard: %w, output: %s", err, string(output))
	}

	w.active = true
	log.Printf("Túnel WireGuard iniciado: %s\n", w.config.InterfaceName)
	return nil
}

// Stop para o túnel WireGuard
func (w *WireGuardManager) Stop() error {
	if !w.active {
		return nil
	}

	// Parar o túnel WireGuard usando wg-quick
	cmd := exec.Command("wg-quick", "down", w.config.ConfigPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("falha ao parar túnel WireGuard: %w, output: %s", err, string(output))
	}

	w.active = false
	log.Printf("Túnel WireGuard parado: %s\n", w.config.InterfaceName)
	return nil
}

// IsActive verifica se o túnel WireGuard está ativo
func (w *WireGuardManager) IsActive() bool {
	return w.active
}

// CheckWireGuardInstallation verifica se o WireGuard está instalado
func CheckWireGuardInstallation() error {
	cmd := exec.Command("which", "wg", "wg-quick")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("WireGuard não está instalado: %w", err)
	}

	if strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf("ferramentas WireGuard não encontradas, por favor instale o pacote 'wireguard-tools'")
	}

	return nil
}

// GenerateKeyPair gera um par de chaves para WireGuard
func GenerateKeyPair() (string, string, error) {
	// Gerar chave privada
	cmdPriv := exec.Command("wg", "genkey")
	privateKeyBytes, err := cmdPriv.Output()
	if err != nil {
		return "", "", fmt.Errorf("falha ao gerar chave privada: %w", err)
	}
	privateKey := strings.TrimSpace(string(privateKeyBytes))

	// Gerar chave pública a partir da chave privada
	cmdPub := exec.Command("wg", "pubkey")
	cmdPub.Stdin = strings.NewReader(privateKey)
	publicKeyBytes, err := cmdPub.Output()
	if err != nil {
		return "", "", fmt.Errorf("falha ao gerar chave pública: %w", err)
	}
	publicKey := strings.TrimSpace(string(publicKeyBytes))

	return privateKey, publicKey, nil
}
