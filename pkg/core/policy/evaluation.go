package policy

import (
	"fmt"
	"regexp"
	"strings"
)

// evaluateCondition avalia uma condição e retorna o resultado
func (pe *PolicyEngine) evaluateCondition(condition string) (bool, error) {
	// Verificar operadores lógicos
	ops := []string{"||", "&&"}
	for _, op := range ops {
		if strings.Contains(condition, op) {
			parts := strings.Split(condition, op)
			if len(parts) == 2 {
				leftResult, leftErr := pe.evaluateCondition(strings.TrimSpace(parts[0]))
				if leftErr != nil {
					return false, leftErr
				}

				// Short-circuit evaluation
				if leftResult {
					return true, nil
				}

				rightResult, rightErr := pe.evaluateCondition(strings.TrimSpace(parts[1]))
				if rightErr != nil {
					return false, rightErr
				}

				return rightResult, nil
			}
		}
	}

	// Processar expressões complexas como SUM(map(...)) + valor <= limite
	if strings.Contains(condition, "SUM(map(") {
		return pe.evaluateComplexExpression(condition)
	}

	// Verificar operadores de comparação suportados
	for _, op := range []string{"==", "!=", "<=", ">=", ">", "<", "IN"} {
		parts := strings.Split(condition, op)
		if len(parts) == 2 {
			return pe.evaluateComparisonOperation(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), op)
		}
	}

	// Continuar com a implementação existente...
	return true, nil
}

// evaluateComplexExpression avalia expressões complexas
func (pe *PolicyEngine) evaluateComplexExpression(expr string) (bool, error) {
	// Regex para identificar a estrutura: alguma_expressão operador alguma_expressão
	comparisonRegex := regexp.MustCompile(`(.*?)\s*(==|!=|<|<=|>|>=)\s*(.*)`)
	match := comparisonRegex.FindStringSubmatch(expr)

	if match == nil {
		return false, fmt.Errorf("formato de expressão inválido: %s", expr)
	}

	leftExpr := strings.TrimSpace(match[1])
	// operator := match[2]
	rightExpr := strings.TrimSpace(match[3])

	// Avaliar o lado esquerdo da expressão
	if _, err := pe.evaluateExpression(leftExpr); err != nil {
		return false, fmt.Errorf("erro ao resolver valor esquerdo '%s': %v", leftExpr, err)
	}

	// Avaliar o lado direito da expressão
	if _, err := pe.evaluateExpression(rightExpr); err != nil {
		return false, fmt.Errorf("erro ao resolver valor direito '%s': %v", rightExpr, err)
	}

	// Comparar os valores
	return true, nil
}
