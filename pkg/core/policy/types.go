package policy

import "github.com/raywall/aws-policy-engine-go/pkg/core/request"

// PolicyEngine representa o motor de execução de políticas
type PolicyEngine struct {
	Results     map[string]PolicyResult `json:"results"`
	policyRules map[string][]string     `json:"policyRules"`
	request     *request.Request        `json:"request"`
}

// Policy representa uma definição de política
type Policy struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Validate    map[string][]Validation `yaml:"validate,omitempty"`
	Compare     []Comparison            `yaml:"compare,omitempty"`
	Set         map[string]interface{}  `yaml:"set,omitempty"`
}

// ConditionResult armazena o resultado da execução de uma condição
type ConditionResult struct {
	Condition string `json:"condition"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}

// Validation representa uma regra de validação para um campo
type Validation struct {
	Type    string        `yaml:"type"`
	Values  []interface{} `yaml:"values,omitempty"`
	Pattern string        `yaml:"pattern,omitempty"`
	Min     int           `yaml:"min,omitempty"`
	Max     int           `yaml:"max,omitempty"`
}

// Comparison representa uma operação de comparação entre dois valores
type Comparison struct {
	Left     string `yaml:"left"`
	Operator string `yaml:"operator"`
	Right    string `yaml:"right"`
}

// PolicyResult armazena o resultado da execução de uma política
type PolicyResult struct {
	PolicyName string                 `json:"policyName"`
	Conditions []ConditionResult      `json:"conditionsResult"`
	Passed     bool                   `json:"passed"`
	Errors     []string               `json:"errors,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
}
