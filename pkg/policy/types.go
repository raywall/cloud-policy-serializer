package policy

import "encoding/json"

// Context contém informações de contexto da requisição
type Context struct {
	UserID string `json:"userId,omitempty"`
	Source string `json:"source,omitempty"`
}

// Request representa a estrutura da requisição recebida
type Request struct {
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Context   *Context               `json:"context,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Policies  []string               `json:"policies"`
}

// Response representa a estrutura da resposta enviada
type Response struct {
	ID              string                 `json:"id"`
	Timestamp       string                 `json:"timestamp"`
	Success         bool                   `json:"success"`
	Data            map[string]interface{} `json:"data,omitempty"`
	AppliedPolicies []PolicyResult         `json:"appliedPolicies,omitempty"`
	Error           *Error                 `json:"error,omitempty"`
}

// PolicyResult representa o resultado da aplicação de uma política
type PolicyResult struct {
	Name       string            `json:"name"`
	Success    bool              `json:"success"`
	Conditions []ConditionResult `json:"conditions,omitempty"`
}

// ConditionResult representa o resultado da avaliação de uma condição
type ConditionResult struct {
	Condition string `json:"condition"`
	Result    bool   `json:"result"`
}

// Error representa informações de erro
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ProcessResult representa o resultado do processamento das políticas
type ProcessResult struct {
	Success         bool
	Data            map[string]interface{}
	AppliedPolicies []PolicyResult
}

// DeepCopy realiza uma cópia profunda de um mapa usando json marshal/unmarshal
func DeepCopy(src map[string]interface{}) (map[string]interface{}, error) {
	if src == nil {
		return nil, nil
	}

	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	var dst map[string]interface{}
	err = json.Unmarshal(data, &dst)
	if err != nil {
		return nil, err
	}

	return dst, nil
}
