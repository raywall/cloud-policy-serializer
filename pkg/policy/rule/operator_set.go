package rule

import (
	"fmt"
	"strconv"
	"strings"
)

func (tr *TrimmedRule) Set(rule string, data map[string]interface{}) TrimmedRuleExecutionResult {
	trimmedRule := tr.String()

	if strings.HasPrefix(trimmedRule, "SET ") {
		parts := strings.SplitN(strings.TrimPrefix(trimmedRule, "SET "), "=", 2)
		if len(parts) != 2 {
			return TrimmedRuleExecutionResult{
				Executed: true,
				Passed:   false,
				Detail:   "",
				Err:      fmt.Errorf("regra SET inválida: %s", rule),
			}
		}
		targetPath := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])
		var valueToSet interface{}
		var evalDetails string

		if expr, isExpr := extractExpression(valueStr); isExpr {
			calculatedValue, details, err := evaluateMathExpression(expr, data)
			evalDetails = fmt.Sprintf("EXP(%s)", details)
			if err != nil {
				return TrimmedRuleExecutionResult{
					Executed: true,
					Passed:   false,
					Detail:   fmt.Sprintf("Erro ao avaliar EXP em SET para '%s': %s", targetPath, evalDetails),
					Err:      err,
				}
			}
			valueToSet = calculatedValue
		} else if strings.HasPrefix(valueStr, "$.") { // Atribuição direta de caminho
			val, err := getValue(data, valueStr)
			if err != nil {
				return TrimmedRuleExecutionResult{
					Executed: true,
					Passed:   false,
					Detail:   fmt.Sprintf("Falha ao obter valor para SET RHS path '%s': %v", valueStr, err),
					Err:      err,
				}
			}
			valueToSet = val
			evalDetails = fmt.Sprintf("path %s = %v", valueStr, val)
		} else { // Literal
			if fVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
				valueToSet = fVal
			} else if bVal, err := strconv.ParseBool(valueStr); err == nil {
				valueToSet = bVal
			} else { // String literal
				if (strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'")) ||
					(strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")) {
					valueToSet = valueStr[1 : len(valueStr)-1]
				} else {
					valueToSet = valueStr
				}
			}
			evalDetails = fmt.Sprintf("literal = %v", valueToSet)
		}

		err := setValue(data, targetPath, valueToSet)
		if err != nil {
			return TrimmedRuleExecutionResult{
				Executed: true,
				Passed:   false,
				Detail:   fmt.Sprintf("Falha ao SET valor para path '%s': %v. Detalhes da avaliação: %s", targetPath, err, evalDetails),
				Err:      err,
			}
		}
		return TrimmedRuleExecutionResult{
			Executed: true,
			Passed:   true,
			Detail:   fmt.Sprintf("SET %s = %v (Detalhes: %s)", targetPath, valueToSet, evalDetails),
			Err:      nil,
		}
	}

	return TrimmedRuleExecutionResult{
		Executed: false,
	}
}
