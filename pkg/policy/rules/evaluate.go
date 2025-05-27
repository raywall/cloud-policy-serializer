package rules

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// evaluateRule avalia uma única string de regra de política contra os dados.
// Retorna: bool (passou), string (detalhes), error (erro de avaliação)
func EvaluateRule(rule string, data map[string]interface{}) (bool, string, error) {
	tr := NewTrimmedRule(rule)

	// Lógica OR (sem alterações)
	if res := tr.Or(rule, data); res.Executed {
		return res.Passed, res.Detail, res.Err
	}

	// Lógica SET
	if res := tr.Set(rule, data); res.Executed {
		return res.Passed, res.Detail, res.Err
	}

	// Lógica IF THEN (sem alterações na estrutura, mas usará EXP se presente na ação)
	if res := tr.Condition(rule, data); res.Executed {
		return res.Passed, res.Detail, res.Err
	}

	trimmedRule := tr.String()

	// Lógica de Condição (LHS op RHS)
	re := regexp.MustCompile(`(.*?)\s*(>=|<=|==|!=|>|<|IN|NOT IN)\s*(.*)`)
	matches := re.FindStringSubmatch(trimmedRule)
	if len(matches) != 4 {
		// Checagem de null
		nullRe := regexp.MustCompile(`(.*?)\s*(==|!=)\s*null`)
		nullMatches := nullRe.FindStringSubmatch(trimmedRule)
		if len(nullMatches) == 3 {
			lhsPath, op := strings.TrimSpace(nullMatches[1]), strings.TrimSpace(nullMatches[2])
			lhsValue, errLhs := getValue(data, lhsPath) // Não checa EXP para null check
			if errLhs != nil {
				return false, fmt.Sprintf("Erro LHS '%s' em null check: %v", lhsPath, errLhs), errLhs
			}
			result := (op == "==" && lhsValue == nil) || (op == "!=" && lhsValue != nil)
			return result, fmt.Sprintf("%s (%v) %s null -> %t", lhsPath, lhsValue, op, result), nil
		}
		// Funções não implementadas
		funcRe := regexp.MustCompile(`(COUNT|SUM)\((.*?)\)`).FindStringSubmatch(trimmedRule)
		if len(funcRe) > 0 {
			err := fmt.Errorf("função não implementada: %s", funcRe[0])
			return false, err.Error(), err
		}
		return false, "", fmt.Errorf("regra de condição inválida: '%s'", rule)
	}

	lhsStr := strings.TrimSpace(matches[1])
	op := strings.TrimSpace(matches[2])
	rhsStr := strings.TrimSpace(matches[3])

	var lhsValue, rhsValue interface{}
	var lhsDetails, rhsDetails string = "literal ou path", "literal ou path"
	var err error

	// Avaliar LHS
	if expr, isExpr := extractExpression(lhsStr); isExpr {
		lhsValue, lhsDetails, err = evaluateMathExpression(expr, data)
		lhsDetails = fmt.Sprintf("EXP(%s)", lhsDetails)
	} else {
		lhsValue, err = getValue(data, lhsStr)             // Assume que é um path se não for EXP()
		if err == nil && strings.HasPrefix(lhsStr, "$.") { // Se foi path e deu certo
			lhsDetails = fmt.Sprintf("path %s = %v", lhsStr, lhsValue)
		} else if err != nil { // Se getValue falhou, pode ser um literal
			// Tenta converter como literal numérico ou booleano
			if fVal, eConv := strconv.ParseFloat(lhsStr, 64); eConv == nil {
				lhsValue = fVal
				err = nil
				lhsDetails = fmt.Sprintf("literal %f", fVal)
			} else if bVal, eConv := strconv.ParseBool(lhsStr); eConv == nil {
				lhsValue = bVal
				err = nil
				lhsDetails = fmt.Sprintf("literal %t", bVal)
			} else { // É uma string literal ou um path inválido
				if !strings.HasPrefix(lhsStr, "$.") { // Não é um path, então é string literal
					lhsValue = strings.Trim(lhsStr, "\"'") // Remove aspas
					err = nil
					lhsDetails = fmt.Sprintf("literal string '%s'", lhsValue)
				} else {
					// É um path que falhou em getValue, mantém o erro
				}
			}
		} else { // getValue retornou nil, nil (path válido, valor ausente)
			lhsDetails = fmt.Sprintf("path %s = nil (ausente)", lhsStr)
		}
	}
	if err != nil {
		return false, fmt.Sprintf("Erro ao avaliar LHS '%s': %s. Detalhes: %v", lhsStr, lhsDetails, err), err
	}

	// Avaliar RHS (exceto para IN/NOT IN que têm tratamento especial de lista)
	if op != "IN" && op != "NOT IN" {
		if expr, isExpr := extractExpression(rhsStr); isExpr {
			rhsValue, rhsDetails, err = evaluateMathExpression(expr, data)
			rhsDetails = fmt.Sprintf("EXP(%s)", rhsDetails)
		} else if strings.HasPrefix(rhsStr, "$.") {
			rhsValue, err = getValue(data, rhsStr)
			if err == nil {
				rhsDetails = fmt.Sprintf("path %s = %v", rhsStr, rhsValue)
			}
		} else { // Literal
			if fVal, eConv := strconv.ParseFloat(rhsStr, 64); eConv == nil {
				rhsValue = fVal
				rhsDetails = fmt.Sprintf("literal %f", fVal)
			} else if bVal, eConv := strconv.ParseBool(rhsStr); eConv == nil {
				rhsValue = bVal
				rhsDetails = fmt.Sprintf("literal %t", bVal)
			} else {
				rhsValue = strings.Trim(rhsStr, "\"'")
				rhsDetails = fmt.Sprintf("literal string '%s'", rhsValue)
			}
		}
		if err != nil {
			return false, fmt.Sprintf("Erro ao avaliar RHS '%s': %s. Detalhes: %v", rhsStr, rhsDetails, err), err
		}
	} else { // Para IN / NOT IN, RHS é uma lista de strings
		tempRhsStr := strings.TrimSpace(rhsStr)
		if !(strings.HasPrefix(tempRhsStr, "[") && strings.HasSuffix(tempRhsStr, "]")) {
			return false, "", fmt.Errorf("lista para %s deve ser [...]: %s", op, tempRhsStr)
		}
		listContent := strings.Trim(tempRhsStr, "[] ")
		if listContent == "" {
			rhsValue = []string{}
		} else {
			elements := regexp.MustCompile(`\s*,\s*`).Split(listContent, -1)
			strList := make([]string, len(elements))
			for i, el := range elements {
				strList[i] = strings.Trim(strings.TrimSpace(el), "\"'")
			}
			rhsValue = strList
		}
		rhsDetails = fmt.Sprintf("lista %v", rhsValue)
	}

	result := false
	var finalDetails string
	switch op {
	case "==":
		result = compareEquals(lhsValue, rhsValue)
	case "!=":
		result = !compareEquals(lhsValue, rhsValue)
	case ">", ">=", "<", "<=":
		lhsNum, okLhs := convertToFloat64(lhsValue)
		rhsNum, okRhs := convertToFloat64(rhsValue)
		if !okLhs || !okRhs {
			errText := fmt.Sprintf("não numérico: LHS (%v %T, num:%t), RHS (%v %T, num:%t)", lhsValue, lhsValue, okLhs, rhsValue, rhsValue, okRhs)
			return false, fmt.Sprintf("%s %s %s -> ERRO: %s", lhsDetails, op, rhsDetails, errText), errors.New(errText)
		}
		switch op {
		case ">":
			result = lhsNum > rhsNum
		case ">=":
			result = lhsNum >= rhsNum
		case "<":
			result = lhsNum < rhsNum
		case "<=":
			result = lhsNum <= rhsNum
		}
	case "IN":
		rhsList, _ := rhsValue.([]string) // Já validado
		lhsStrVal, ok := lhsValue.(string)
		if !ok {
			result = false
		} else {
			for _, item := range rhsList {
				if lhsStrVal == item {
					result = true
					break
				}
			}
		}
	case "NOT IN":
		rhsList, _ := rhsValue.([]string) // Já validado
		lhsStrVal, ok := lhsValue.(string)
		if !ok {
			result = true
		} else { // Se LHS não é string, ou é nil, não está na lista
			found := false
			for _, item := range rhsList {
				if lhsStrVal == item {
					found = true
					break
				}
			}
			result = !found
		}
	default:
		return false, "", fmt.Errorf("operador não suportado: %s", op)
	}
	finalDetails = fmt.Sprintf("%s %s %s -> %t", lhsDetails, op, rhsDetails, result)
	return result, finalDetails, nil
}
