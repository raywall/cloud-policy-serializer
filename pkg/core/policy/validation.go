package policy

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/raywall/aws-policy-engine-go/pkg/json/extractor"
)

// evaluateExpression avalia uma expressão e retorna seu valor
func (pe *PolicyEngine) evaluateExpression(expression string) (interface{}, error) {
	expression = strings.TrimSpace(expression)

	// Verifica se é uma operação matemática simples
	if mathMatch := regexp.MustCompile(`(.+?)\s*([*+\-/])\s*(.+)`).FindStringSubmatch(expression); len(mathMatch) > 0 {
		left := mathMatch[1]
		operator := mathMatch[2]
		right := mathMatch[3]

		leftValue, err := pe.resolveValue(left, pe.request.Data)
		if err != nil {
			return nil, err
		}

		rightValue, err := pe.resolveValue(right, pe.request.Data)
		if err != nil {
			return nil, err
		}

		// Converter para valores numéricos
		leftNum, leftOk := pe.toNumber(leftValue)
		rightNum, rightOk := pe.toNumber(rightValue)

		if !leftOk || !rightOk {
			return nil, fmt.Errorf("valores não numéricos em expressão matemática: %v %s %v", leftValue, operator, rightValue)
		}

		switch operator {
		case "*":
			return leftNum * rightNum, nil
		case "/":
			if rightNum == 0 {
				return nil, fmt.Errorf("divisão por zero")
			}
			return leftNum / rightNum, nil
		case "+":
			return leftNum + rightNum, nil
		case "-":
			return leftNum - rightNum, nil
		default:
			return nil, fmt.Errorf("operador não suportado: %s", operator)
		}
	}

	// Se não for uma operação, é um valor simples (literal ou referência)
	return pe.resolveValue(expression, pe.request.Data)
}

// resolveValue resolve um valor a partir de uma expressão
func (pe *PolicyEngine) resolveValue(expr string, data interface{}) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// Se for um literal numérico
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num, nil
	}

	// Se for uma string literal (entre aspas)
	if (strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"")) ||
		(strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) {
		return strings.Trim(expr, "\"'"), nil
	}

	// Se for null
	if expr == "null" {
		return nil, nil
	}

	// Se for boolean
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	// Se for uma referência JSONPath
	if strings.HasPrefix(expr, "$.") {
		ex, err := extractor.NewFromStruct(data)
		if err != nil {
			return nil, err
		}

		return ex.Extract(expr)
	}

	return nil, fmt.Errorf("não foi possível resolver o valor: %s", expr)
}

// GetResults retorna os resultados da execução das políticas
func (pe *PolicyEngine) getResults() map[string]PolicyResult {
	return pe.Results
}

// IsSuccessful verifica se todas as políticas foram executadas com sucesso
func (pe *PolicyEngine) isSuccessful() bool {
	for _, result := range pe.Results {
		if !result.Passed {
			return false
		}
	}
	return true
}

// toNumber tenta converter um valor para um número
func (pe *PolicyEngine) toNumber(value interface{}) (float64, bool) {
	if value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}

	// Tenta converter através de reflexão
	valueOfV := reflect.ValueOf(value)
	switch valueOfV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(valueOfV.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(valueOfV.Uint()), true
	case reflect.Float32, reflect.Float64:
		return valueOfV.Float(), true
	}

	return 0, false
}
