package request

// Context contém informações de contexto da requisição
// - UserID: ID do usuário
// - Source: Origem da solicitação
type Context struct {
	UserID string `json:"userId,omitempty"`
	Source string `json:"source,omitempty"`
}

// Request representa a estrutura da requisição recebida
// - ID: Identificador único da solicitação
// - Timestamp: Data e hora da solicitação (opcional)
// - Data: Dados para aplicação das políticas (obrigatório)
// - Context: Contexto da solicitação (opcional)
// - Lista de políticas a serem aplicadas (obrigatório)
type Request struct {
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Context   *Context               `json:"context,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Policies  []string               `json:"policies"`
}
