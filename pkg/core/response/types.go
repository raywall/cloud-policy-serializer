package response

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/raywall/aws-policy-engine-go/pkg/core/policy"
	"github.com/raywall/aws-policy-engine-go/pkg/core/request"
)

// Response representa a estrutura da resposta enviada
// - ID: Identificador único da solicitação
// - Timestamp: Data e hora da solicitação (opcional)
// - ElapsedTime: Tempo decorrido para execução das políticas (milessegundos)
// - Passed: Informa se todas as políticas foram aplicadas com sucesso
// - AppliedPolicies: Resultado das políticas aplicadas
// - Error: Erro ocorrido ao aplicar as políticas
type Response struct {
	ID              string                 `json:"id"`
	Timestamp       string                 `json:"timestamp"`
	Passed          bool                   `json:"passed"`
	Data            map[string]interface{} `json:"data,omitempty"`
	AppliedPolicies []policy.PolicyResult  `json:"appliedPolicies,omitempty"`
	Error           Error                  `json:"error,omitempty"`
	ElapsedTime     int64                  `json:"elapsedTime"`
	startedOn       time.Time              `json:"-"`
}

// Error representa informações de erro
// - Code: Código do erro ocorrido
// - Message: Descrição do erro ocorrido
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewResponse(req *request.Request) *Response {
	return &Response{
		ID:              req.ID,
		Timestamp:       req.Timestamp,
		Passed:          true,
		Data:            req.Data,
		AppliedPolicies: []policy.PolicyResult{},
		startedOn:       time.Now(),
	}
}

func (res *Response) BuildResponse(pe *policy.PolicyEngine, data map[string]interface{}, err *Error) {
	if err != nil {
		res.Error = *err
		return
	}

	res.Data = data

	for _, result := range pe.Results {
		if !result.Passed {
			res.Passed = false
		}

		res.AppliedPolicies = append(res.AppliedPolicies, result)
	}
}

func (res *Response) Send(w http.ResponseWriter) {
	res.send(w, 200)
}

func (res *Response) SendError(w http.ResponseWriter, code string, err error) {
	if err != nil {
		res.Error = Error{
			Code:    code,
			Message: err.Error(),
		}

		res.send(w, http.StatusInternalServerError)
	}

	res.Passed = false
	res.send(w, http.StatusBadRequest)
}

func (res *Response) send(w http.ResponseWriter, statusCode int) {
	res.ElapsedTime = time.Since(res.startedOn).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(*res); err != nil {
		log.Printf("Erro ao codificar resposta JSON: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
	}
}
