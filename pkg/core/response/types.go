package response

import (
	"time"

	"github.com/raywall/aws-policy-engine-go/pkg/core/policy"
	"github.com/raywall/aws-policy-engine-go/pkg/json/schema"
)

// Response representa a estrutura da resposta enviada
// - ID: Identificador único da solicitação
// - Timestamp: Data e hora da solicitação (opcional)
// - ElapsedTime: Tempo decorrido para execução das políticas (milessegundos)
// - Passed: Informa se todas as políticas foram aplicadas com sucesso
// - AppliedPolicies: Resultado das políticas aplicadas
// - Error: Erro ocorrido ao aplicar as políticas
type Response struct {
	ID                string                 `json:"id"`
	Timestamp         string                 `json:"timestamp"`
	Passed            bool                   `json:"passed"`
	Data              map[string]interface{} `json:"data,omitempty"`
	AppliedPolicies   []policy.PolicyResult  `json:"appliedPolicies,omitempty"`
	Error             Error                  `json:"error,omitempty"`
	ElapsedTime       int64                  `json:"elapsedTime"`
	startedOn         time.Time              `json:"-"`
	responseSchema    *schema.Schema         `json:"-"`
	formattedResponse interface{}            `json:"-"`
	debugger          bool                   `json:"-"`
}

// Error representa informações de erro
// - Code: Código do erro ocorrido
// - Message: Descrição do erro ocorrido
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   error  `json:"error,omitempty"`
}
