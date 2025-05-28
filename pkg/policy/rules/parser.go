package rules

import (
	"fmt"
	"strconv"
	"strings"
)

// parseOperand converte uma string de operando (literal ou caminho $) em float64.
func parseOperand(operandStr string, data map[string]interface{}) (float64, error) {
	operandStr = strings.TrimSpace(operandStr)
	if strings.HasPrefix(operandStr, "$.") {
		val, err := getValue(data, operandStr)
		if err != nil {
			return 0, fmt.Errorf("falha ao obter valor do caminho do operando '%s': %v", operandStr, err)
		}
		num, ok := convertToFloat64(val)
		if !ok {
			return 0, fmt.Errorf("operando do caminho '%s' (valor: %v, tipo: %T) não é um número válido", operandStr, val, val)
		}
		return num, nil
	}
	// É um literal
	num, ok := convertToFloat64(operandStr)
	if !ok {
		return 0, fmt.Errorf("operando literal '%s' não é um número válido", operandStr)
	}
	return num, nil
}

func convertToFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false // Não pode converter nil para float64
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f, true
		}
	}
	return 0, false
}
