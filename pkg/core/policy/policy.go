package policy

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/raywall/aws-policy-engine-go/pkg/core/request"
	"gopkg.in/yaml.v3"
)

// NewPolicyEngine cria um novo motor de políticas
func NewPolicyEngine(policyContent []byte) (*PolicyEngine, error) {
	rules := make(map[string][]string)
	err := yaml.Unmarshal(policyContent, &rules)
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar arquivo de políticas: %v", err)
	}

	return &PolicyEngine{
		Results:     make(map[string]PolicyResult),
		policyRules: rules,
	}, nil
}

// ExecutePolicies executa as políticas definidas no request
func (pe *PolicyEngine) ExecutePolicies(req *request.Request) error {
	if req == nil {
		return errors.New("request não pode ser nulo")
	}

	if req.Policies == nil || len(req.Policies) == 0 {
		return errors.New("nenhuma política especificada para execução")
	}

	pe.request = req

	// Para cada política solicitada no request
	for _, policyName := range req.Policies {
		rules, exists := pe.policyRules[policyName]
		if !exists {
			return fmt.Errorf("política não encontrada: %s", policyName)
		}

		result := PolicyResult{
			PolicyName: policyName,
			Conditions: make([]ConditionResult, 0),
			Passed:     true,
		}

		// Para cada regra/condição da política
		for _, rule := range rules {
			condResult := pe.executeRule(rule)
			result.Conditions = append(result.Conditions, condResult)

			if !condResult.Success {
				result.Passed = false
			}
		}

		pe.Results[policyName] = result
	}

	return nil
}

// executeRule executa uma regra específica nos dados
func (pe *PolicyEngine) executeRule(rule string) ConditionResult {
	rule = strings.TrimSpace(rule)

	result := ConditionResult{
		Condition: rule,
		Success:   false,
		Error:     "",
	}

	// Verifica se é uma instrução SET
	if setMatch := regexp.MustCompile(`^SET\s+(.+?)\s*=\s*(.+)$`).FindStringSubmatch(rule); len(setMatch) > 0 {
		targetPath := setMatch[1]
		expression := setMatch[2]

		err := pe.executeSetOperation(targetPath, expression)
		if err != nil {
			result.Error = err.Error()
			return result
		}

		result.Success = true
		return result
	}

	// Verifica se é uma instrução IF-THEN
	if ifMatch := regexp.MustCompile(`^IF\s+(.+?)\s+THEN\s+(.+)$`).FindStringSubmatch(rule); len(ifMatch) > 0 {
		condition := ifMatch[1]
		thenAction := ifMatch[2]

		// Avaliar a condição
		if _, err := pe.evaluateCondition(condition); err != nil {
			// if !condResult.Success {
			// Se a condição falhar, a regra ainda é considerada bem-sucedida
			// porque é uma condição condicional
			result.Success = true
			return result
		}

		// Se a condição for verdadeira, execute a ação then
		thenResult := pe.executeRule(thenAction)
		if !thenResult.Success {
			result.Error = thenResult.Error
			return result
		}

		result.Success = true
		return result
	}

	// Caso contrário, é uma condição de validação padrão
	res, err := pe.evaluateCondition(rule)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
	}

	result.Success = res
	return result
}
