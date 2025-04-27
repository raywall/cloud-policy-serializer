package policy

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/seu-usuario/policy-engine/models"
	"gopkg.in/yaml.v3"
)

// PolicyMap armazena todas as políticas carregadas
var PolicyMap = make(map[string][]string)

// Funções customizadas para o avaliador de expressões
var customFunctions = map[string]govaluate.ExpressionFunction{
	"SUM": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("SUM espera exatamente 1 argumento")
		}

		slice, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("SUM espera um array como argumento")
		}

		var sum float64
		for _, v := range slice {
			if num, ok := v.(float64); ok {
				sum += num
			} else if numStr, ok := v.(string); ok {
				if parsed, err := strconv.ParseFloat(numStr, 64); err == nil {
					sum += parsed
				}
			} else if numMap, ok := v.(map[string]interface{}); ok {
				// Tenta extrair "valor" de um objeto (caso comum em transações)
				if val, exists := numMap["valor"]; exists {
					if numVal, ok := val.(float64); ok {
						sum += numVal
					}
				}
			}
		}
		return sum, nil
	},
	"COUNT": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("COUNT espera exatamente 1 argumento")
		}

		slice, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("COUNT espera um array como argumento")
		}

		return float64(len(slice)), nil
	},
	"IN": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("IN espera exatamente 2 argumentos")
		}

		item := args[0]
		slice, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("IN espera um array como segundo argumento")
		}

		for _, v := range slice {
			if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", item) {
				return true, nil
			}
		}
		return false, nil
	},
	"map": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("map espera exatamente 2 argumentos (array e caminho)")
		}

		slice, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("map espera um array como primeiro argumento")
		}

		path, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("map espera uma string como segundo argumento")
		}

		result := make([]interface{}, 0, len(slice))
		for _, item := range slice {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// Separar o caminho por pontos
				parts := strings.Split(path, ".")
				value := getNestedValue(itemMap, parts)
				result = append(result, value)
			}
		}

		return result, nil
	},
}

// LoadPoliciesFromFile carrega políticas de um arquivo YAML
func LoadPoliciesFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &PolicyMap)
	if err != nil {
		return err
	}

	return nil
}

// ProcessPolicies processa as políticas solicitadas na requisição
func ProcessPolicies(request models.Request) (models.ProcessResult, error) {
	result := models.ProcessResult{
		Success:         true,
		AppliedPolicies: []models.PolicyResult{},
	}

	// Clonar os dados para não modificar o original
	data, err := models.DeepCopy(request.Data)
	if err != nil {
		return result, err
	}
	result.Data = data

	// Para cada política solicitada
	for _, policyName := range request.Policies {
		policyResult := models.PolicyResult{
			Name:       policyName,
			Success:    true,
			Conditions: []models.ConditionResult{},
		}

		// Verificar se a política existe
		conditions, exists := PolicyMap[policyName]
		if !exists {
			policyResult.Success = false
			result.Success = false
			policyResult.Conditions = append(policyResult.Conditions, models.ConditionResult{
				Condition: fmt.Sprintf("Policy %s not found", policyName),
				Result:    false,
			})
			result.AppliedPolicies = append(result.AppliedPolicies, policyResult)
			continue
		}

		// Aplicar cada condição da política
		for _, conditionStr := range conditions {
			conditionResult := models.ConditionResult{
				Condition: conditionStr,
				Result:    true,
			}

			var conditionErr error

			// Verificar se é uma condição SET
			if strings.HasPrefix(conditionStr, "SET ") {
				conditionErr = processSetStatement(conditionStr[4:], data)
			} else if strings.HasPrefix(conditionStr, "IF ") {
				// Processar declaração IF
				conditionErr = processIfStatement(conditionStr, data)
			} else {
				// Processar expressão booleana normal
				expr, err := govaluate.NewEvaluableExpressionWithFunctions(conditionStr, customFunctions)
				if err != nil {
					conditionErr = err
				} else {
					// Criar parâmetros para avaliação
					parameters := map[string]interface{}{
						"data": data,
					}

					// Avaliar a expressão
					exprResult, err := expr.Evaluate(parameters)
					if err != nil {
						conditionErr = err
					} else {
						// Verificar se o resultado é booleano e true
						boolResult, ok := exprResult.(bool)
						if !ok {
							conditionErr = fmt.Errorf("a expressão não resultou em um valor booleano: %v", exprResult)
						} else if !boolResult {
							conditionErr = fmt.Errorf("condição avaliada como falsa")
						}
					}
				}
			}

			// Atualizar o resultado da condição
			if conditionErr != nil {
				conditionResult.Result = false
				policyResult.Success = false
				result.Success = false
				// Adicionar erro ao log ou a um campo de detalhes
				fmt.Printf("Erro na condição '%s': %v\n", conditionStr, conditionErr)
			}

			policyResult.Conditions = append(policyResult.Conditions, conditionResult)
		}

		result.AppliedPolicies = append(result.AppliedPolicies, policyResult)
	}

	return result, nil
}

// processSetStatement processa uma declaração SET (atribuição)
func processSetStatement(statement string, data map[string]interface{}) error {
	parts := strings.Split(statement, "=")
	if len(parts) != 2 {
		return fmt.Errorf("declaração SET inválida: %s", statement)
	}

	// Extrair o caminho do valor e a expressão
	path := strings.TrimSpace(parts[0])
	expr := strings.TrimSpace(parts[1])

	// Avaliar o valor da expressão
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(expr, customFunctions)
	if err != nil {
		return err
	}

	// Criar parâmetros para avaliação
	parameters := map[string]interface{}{
		"data": data,
	}

	// Avaliar a expressão
	result, err := expression.Evaluate(parameters)
	if err != nil {
		return err
	}

	// Definir o valor no caminho especificado
	err = setValueByPath(data, path, result)
	if err != nil {
		return err
	}

	return nil
}

// processIfStatement processa uma declaração IF-THEN
func processIfStatement(statement string, data map[string]interface{}) error {
	// Remover o "IF " do início
	statement = strings.TrimPrefix(statement, "IF ")

	// Dividir em condição e ação
	parts := strings.Split(statement, " THEN ")
	if len(parts) != 2 {
		return fmt.Errorf("declaração IF inválida: %s", statement)
	}

	condition := parts[0]
	action := parts[1]

	// Avaliar a condição
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(condition, customFunctions)
	if err != nil {
		return err
	}

	// Criar parâmetros para avaliação
	parameters := map[string]interface{}{
		"data": data,
	}

	// Avaliar a expressão
	result, err := expr.Evaluate(parameters)
	if err != nil {
		return err
	}

	// Verificar se o resultado é booleano e true
	boolResult, ok := result.(bool)
	if !ok {
		return fmt.Errorf("a condição não resultou em um valor booleano")
	}

	// Se a condição for verdadeira, executar a ação
	if boolResult {
		if strings.HasPrefix(action, "SET ") {
			return processSetStatement(action[4:], data)
		}
		return fmt.Errorf("ação não suportada: %s", action)
	}

	return nil
}

// getNestedValue obtém um valor aninhado de um mapa
func getNestedValue(data map[string]interface{}, parts []string) interface{} {
	if len(parts) == 0 {
		return nil
	}

	part := parts[0]
	if len(parts) == 1 {
		return data[part]
	}

	if next, ok := data[part].(map[string]interface{}); ok {
		return getNestedValue(next, parts[1:])
	}

	return nil
}

// setValueByPath define um valor em um mapa usando um caminho com notação de pontos
func setValueByPath(data map[string]interface{}, path string, value interface{}) error {
	parts := strings.Split(path, ".")

	current := data
	for i, part := range parts {
		if i == len(parts)-1 {
			// Último elemento, definir o valor
			current[part] = value
			return nil
		}

		// Verificar se o caminho intermediário existe
		if _, exists := current[part]; !exists {
			// Criar um novo mapa se não existir
			current[part] = make(map[string]interface{})
		}

		// Navegar para o próximo nível
		next, ok := current[part].(map[string]interface{})
		if !ok {
			return fmt.Errorf("não é possível navegar pelo caminho: %s não é um objeto", part)
		}
		current = next
	}

	return nil
}
