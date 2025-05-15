package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
)

func main() {
	var endpoint string
	var detailedExploration bool

	flag.StringVar(&endpoint, "endpoint", "opc.tcp://juan-FP750:53530/OPCUA/SimulationServer", "OPC-UA endpoint")
	flag.BoolVar(&detailedExploration, "detailed", true, "Realizar exploração detalhada dos nós de simulação")
	flag.Parse()

	// Criar cliente OPC-UA
	c, err := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err != nil {
		log.Fatalf("Erro ao criar cliente: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer c.Close(ctx)

	log.Printf("Conectado ao servidor: %s", endpoint)

	// Explorar o servidor a partir do nó raiz
	browseRoot(c, detailedExploration)
}

func browseRoot(c *opcua.Client, detailed bool) {
	nodeID := ua.NewTwoByteNodeID(id.RootFolder)
	log.Printf("Explorando a partir do nó raiz: %s", nodeID)

	// Criar contexto
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Realizar navegação
	req := &ua.BrowseRequest{
		NodesToBrowse: []*ua.BrowseDescription{
			{
				NodeID:          nodeID,
				BrowseDirection: ua.BrowseDirectionForward,
				ReferenceTypeID: ua.NewNumericNodeID(0, id.References),
				IncludeSubtypes: true,
				ResultMask:      uint32(ua.BrowseResultMaskAll),
			},
		},
	}

	resp, err := c.Browse(ctx, req)
	if err != nil {
		log.Fatalf("Erro ao navegar: %v", err)
	}

	if len(resp.Results) == 0 {
		log.Fatal("Nenhum resultado retornado")
	}

	// Exibir resultados
	refs := resp.Results[0].References
	log.Printf("Encontrados %d nós", len(refs))

	// Exibir os primeiros nós
	for _, ref := range refs {
		log.Printf("NodeID: %s, Nome: %s, Tipo: %s",
			ref.NodeID,
			ref.BrowseName.Name,
			ref.TypeDefinition)

		// Se for uma pasta, explorar mais profundamente
		if ref.BrowseName.Name == "Objects" {
			// Tentar extrair o ID de nó do nó expandido
			idString := ref.NodeID.String()
			log.Printf("Explorando Objects NodeID: %s", idString)

			// Explorar diretamente usando o ID original (sem conversão)
			exploreObjects(c, idString, detailed)
		}
	}
}

// Função alternativa que usa string de NodeID em vez de objeto NodeID
func exploreObjects(c *opcua.Client, nodeIDStr string, detailed bool) {
	log.Printf("Explorando nó por string: %s", nodeIDStr)

	// Parsear a string para NodeID
	nodeID, err := ua.ParseNodeID(nodeIDStr)
	if err != nil {
		log.Printf("Erro ao parsear NodeID %s: %v", nodeIDStr, err)
		return
	}

	// Criar contexto
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Realizar navegação
	req := &ua.BrowseRequest{
		NodesToBrowse: []*ua.BrowseDescription{
			{
				NodeID:          nodeID,
				BrowseDirection: ua.BrowseDirectionForward,
				ReferenceTypeID: ua.NewNumericNodeID(0, id.References),
				IncludeSubtypes: true,
				ResultMask:      uint32(ua.BrowseResultMaskAll),
			},
		},
	}

	resp, err := c.Browse(ctx, req)
	if err != nil {
		log.Printf("Erro ao navegar nó %s: %v", nodeIDStr, err)
		return
	}

	if len(resp.Results) == 0 {
		log.Printf("Nenhum resultado retornado para nó %s", nodeIDStr)
		return
	}

	// Exibir resultados
	refs := resp.Results[0].References
	log.Printf("Encontrados %d subnós em %s", len(refs), nodeIDStr)

	// Exibir os primeiros nós
	for _, ref := range refs {
		log.Printf("  * NodeID: %s, Nome: %s", ref.NodeID, ref.BrowseName.Name)

		// Explorar mais se o nome for "Simulation" ou contiver "Sensors"
		if ref.BrowseName.Name == "Simulation" {
			exploreObjects(c, ref.NodeID.String(), detailed)
		} else if detailed && (ref.BrowseName.Name == "Counter" ||
			ref.BrowseName.Name == "Random" ||
			ref.BrowseName.Name == "Sinusoid" ||
			ref.BrowseName.Name == "Sawtooth") {
			// Explorar mais detalhes dos nós de simulação
			exploreVariableNode(c, ref.NodeID.String())
		}
	}
}

// Explorar detalhes de um nó de variável
func exploreVariableNode(c *opcua.Client, nodeIDStr string) {
	// Parsear a string para NodeID
	nodeID, err := ua.ParseNodeID(nodeIDStr)
	if err != nil {
		log.Printf("Erro ao parsear NodeID %s: %v", nodeIDStr, err)
		return
	}

	// Criar contexto
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	resp, err := c.Read(ctx, req)
	if err != nil {
		log.Printf("Erro ao ler valor do nó %s: %v", nodeIDStr, err)
		return
	}

	if len(resp.Results) == 0 {
		log.Printf("Nenhum resultado retornado para nó %s", nodeIDStr)
		return
	}

	// Exibir o valor
	if resp.Results[0].Status == ua.StatusOK {
		log.Printf("    - Valor: %v, Tipo: %T", resp.Results[0].Value.Value(), resp.Results[0].Value.Value())
	} else {
		log.Printf("    - Erro ao ler valor: %v", resp.Results[0].Status)
	}

	// Também navegar pelas propriedades e variáveis
	exploreObjectProperties(c, nodeIDStr)
}

// Explorar propriedades de um objeto
func exploreObjectProperties(c *opcua.Client, nodeIDStr string) {
	// Parsear a string para NodeID
	nodeID, err := ua.ParseNodeID(nodeIDStr)
	if err != nil {
		log.Printf("Erro ao parsear NodeID %s: %v", nodeIDStr, err)
		return
	}

	// Criar contexto
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Realizar navegação
	req := &ua.BrowseRequest{
		NodesToBrowse: []*ua.BrowseDescription{
			{
				NodeID:          nodeID,
				BrowseDirection: ua.BrowseDirectionForward,
				ReferenceTypeID: ua.NewNumericNodeID(0, id.HasProperty),
				IncludeSubtypes: true,
				ResultMask:      uint32(ua.BrowseResultMaskAll),
			},
		},
	}

	resp, err := c.Browse(ctx, req)
	if err != nil {
		log.Printf("Erro ao navegar propriedades do nó %s: %v", nodeIDStr, err)
		return
	}

	if len(resp.Results) == 0 || len(resp.Results[0].References) == 0 {
		return
	}

	// Exibir resultados
	refs := resp.Results[0].References
	log.Printf("    - Encontradas %d propriedades em %s", len(refs), nodeIDStr)

	// Exibir as propriedades
	for _, ref := range refs {
		log.Printf("      + Propriedade: %s, NodeID: %s", ref.BrowseName.Name, ref.NodeID)
	}
}

// Função auxiliar para verificar se uma string contém outra (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	s, substr = toUpper(s), toUpper(substr)
	return containsString(s, substr)
}

// Converte uma string para maiúsculas
func toUpper(s string) string {
	result := make([]byte, len(s))
	for i, c := range []byte(s) {
		if 'a' <= c && c <= 'z' {
			result[i] = c - ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// Verifica se uma string contém outra
func containsString(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
