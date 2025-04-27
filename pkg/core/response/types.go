package response

import (
	"github.com/raywall/aws-policy-engine-go/pkg/json/core/policy"
)

// Response representa a estrutura da resposta enviada
type Response struct {
	ID              string                 `json:"id"`
	Timestamp       string                 `json:"timestamp"`
	Success         bool                   `json:"success"`
	Data            map[string]interface{} `json:"data,omitempty"`
	AppliedPolicies []policy.PolicyResult  `json:"appliedPolicies,omitempty"`
	Error           *Error                 `json:"error,omitempty"`
}

// Error representa informações de erro
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
