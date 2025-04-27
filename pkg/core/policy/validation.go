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
func (pe *PolicyEngine) evaluateExpression(expression string, data interface{}) (interface{}, error) {
	expression = strings.TrimSpace(expression)

	// Verifica se é uma operação matemática simples
	if mathMatch := regexp.MustCompile(`(.+?)\s*([*+\-/])\s*(.+)`).FindStringSubmatch(expression); len(mathMatch) > 0 {
		left := mathMatch[1]
		operator := mathMatch[2]
		right := mathMatch[3]

		leftValue, err := pe.resolveValue(left, data)
		if err != nil {
			return nil, err
		}

		rightValue, err := pe.resolveValue(right, data)
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
	return pe.resolveValue(expression, data)
}

// evaluateCondition avalia uma condição e retorna o resultado
func (pe *PolicyEngine) evaluateCondition(condition string, data interface{}) ConditionResult {
	condition = strings.TrimSpace(condition)

	result := ConditionResult{
		Condition: condition,
		Success:   false,
		Error:     "",
	}

	// Verificar operações COUNT
	if countMatch := regexp.MustCompile(`COUNT\((.+?)\)\s*([<>=!]+)\s*(.+)`).FindStringSubmatch(condition); len(countMatch) > 0 {
		path := countMatch[1]
		operator := countMatch[2]
		rightExpr := countMatch[3]

		// Extrair o array do path
		arrayValue, err := pe.resolveValue(path, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao resolver caminho '%s': %v", path, err)
			return result
		}

		// Verificar se é um array
		array, ok := pe.toArray(arrayValue)
		if !ok {
			result.Error = fmt.Sprintf("o valor em '%s' não é um array", path)
			return result
		}

		count := float64(len(array))

		// Resolver o valor à direita
		rightValue, err := pe.resolveValue(rightExpr, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao resolver valor direito '%s': %v", rightExpr, err)
			return result
		}

		rightNum, ok := pe.toNumber(rightValue)
		if !ok {
			result.Error = fmt.Sprintf("valor direito não é um número: %v", rightValue)
			return result
		}

		result.Success = pe.compareNumbers(count, operator, rightNum)
		return result
	}

	// Verificar operações SUM com map
	if sumMatch := regexp.MustCompile(`SUM\(map\((.+?),\s*"(.+?)"\)\)\s*([<>=!]+)\s*(.+)`).FindStringSubmatch(condition); len(sumMatch) > 0 {
		arrayPath := sumMatch[1]
		propertyName := sumMatch[2]
		operator := sumMatch[3]
		rightExpr := sumMatch[4]

		// Extrair o array
		arrayValue, err := pe.resolveValue(arrayPath, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao resolver caminho '%s': %v", arrayPath, err)
			return result
		}

		// Verificar se é um array
		array, ok := pe.toArray(arrayValue)
		if !ok {
			result.Error = fmt.Sprintf("o valor em '%s' não é um array", arrayPath)
			return result
		}

		// Calcular a soma dos valores da propriedade especificada
		sum := 0.0
		for _, item := range array {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if propValue, exists := itemMap[propertyName]; exists {
					if num, ok := pe.toNumber(propValue); ok {
						sum += num
					}
				}
			}
		}

		// Resolver o valor à direita
		rightValue, err := pe.resolveValue(rightExpr, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao resolver valor direito '%s': %v", rightExpr, err)
			return result
		}

		rightNum, ok := pe.toNumber(rightValue)
		if !ok {
			result.Error = fmt.Sprintf("valor direito não é um número: %v", rightValue)
			return result
		}

		result.Success = pe.compareNumbers(sum, operator, rightNum)
		return result
	}

	// Verificar operação IN
	if inMatch := regexp.MustCompile(`(.+?)\s+IN\s+\[(.*?)\]`).FindStringSubmatch(condition); len(inMatch) > 0 {
		leftExpr := inMatch[1]
		listValues := inMatch[2]

		// Resolver o valor à esquerda
		leftValue, err := pe.resolveValue(leftExpr, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao resolver valor esquerdo '%s': %v", leftExpr, err)
			return result
		}

		// Processar a lista de valores
		listItems := strings.Split(listValues, ",")
		for i, item := range listItems {
			listItems[i] = strings.Trim(item, " \"'")
		}

		// Converter leftValue para string para comparação
		leftStr := fmt.Sprintf("%v", leftValue)

		// Verificar se o valor está na lista
		for _, item := range listItems {
			if leftStr == item {
				result.Success = true
				return result
			}
		}

		return result
	}

	// Verificar operadores de comparação padrão (==, !=, >, <, >=, <=)
	for _, pattern := range []string{
		`(.+?)\s*(==)\s*(.+)`,
		`(.+?)\s*(!=)\s*(.+)`,
		`(.+?)\s*(>=)\s*(.+)`,
		`(.+?)\s*(<=)\s*(.+)`,
		`(.+?)\s*(>)\s*(.+)`,
		`(.+?)\s*(<)\s*(.+)`,
	} {
		if compMatch := regexp.MustCompile(pattern).FindStringSubmatch(condition); len(compMatch) > 0 {
			leftExpr := compMatch[1]
			operator := compMatch[2]
			rightExpr := compMatch[3]

			leftValue, err := pe.resolveValue(leftExpr, data)
			if err != nil {
				result.Error = fmt.Sprintf("erro ao resolver valor esquerdo '%s': %v", leftExpr, err)
				return result
			}

			rightValue, err := pe.resolveValue(rightExpr, data)
			if err != nil {
				result.Error = fmt.Sprintf("erro ao resolver valor direito '%s': %v", rightExpr, err)
				return result
			}

			// Verificar se ambos são números para comparação numérica
			leftNum, leftOk := pe.toNumber(leftValue)
			rightNum, rightOk := pe.toNumber(rightValue)

			if leftOk && rightOk {
				result.Success = pe.compareNumbers(leftNum, operator, rightNum)
				return result
			}

			// Caso contrário, comparar como strings
			leftStr := fmt.Sprintf("%v", leftValue)
			rightStr := fmt.Sprintf("%v", rightValue)

			switch operator {
			case "==":
				result.Success = leftStr == rightStr
			case "!=":
				result.Success = leftStr != rightStr
			case ">":
				result.Success = leftStr > rightStr
			case "<":
				result.Success = leftStr < rightStr
			case ">=":
				result.Success = leftStr >= rightStr
			case "<=":
				result.Success = leftStr <= rightStr
			}

			return result
		}
	}

	// Verificar operadores lógicos (|| e &&)
	if orMatch := regexp.MustCompile(`(.+?)\s*\|\|\s*(.+)`).FindStringSubmatch(condition); len(orMatch) > 0 {
		leftExpr := orMatch[1]
		rightExpr := orMatch[2]

		leftResult, err := pe.evaluateExpression(leftExpr, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao avaliar '%s': %v", leftExpr, err)
			return result
		}
		rightResult, err := pe.evaluateExpression(rightExpr, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao avaliar '%s': %v", rightExpr, err)
			return result
		}

		leftBool, ok := leftResult.(bool)
		if !ok {
			result.Error = fmt.Sprintf("'%s' não avaliou para um booleano", leftExpr)
			return result
		}
		rightBool, ok := rightResult.(bool)
		if !ok {
			result.Error = fmt.Sprintf("'%s' não avaliou para um booleano", rightExpr)
			return result
		}

		result.Success = leftBool || rightBool
		return result
	}

	if andMatch := regexp.MustCompile(`(.+?)\s*&&\s*(.+)`).FindStringSubmatch(condition); len(andMatch) > 0 {
		leftExpr := andMatch[1]
		rightExpr := andMatch[2]

		leftResult, err := pe.evaluateExpression(leftExpr, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao avaliar '%s': %v", leftExpr, err)
			return result
		}
		rightResult, err := pe.evaluateExpression(rightExpr, data)
		if err != nil {
			result.Error = fmt.Sprintf("erro ao avaliar '%s': %v", rightExpr, err)
			return result
		}

		leftBool, ok := leftResult.(bool)
		if !ok {
			result.Error = fmt.Sprintf("'%s' não avaliou para um booleano", leftExpr)
			return result
		}
		rightBool, ok := rightResult.(bool)
		if !ok {
			result.Error = fmt.Sprintf("'%s' não avaliou para um booleano", rightExpr)
			return result
		}

		result.Success = leftBool && rightBool
		return result
	}

	// Verificar verificações de nulidade
	if nullMatch := regexp.MustCompile(`(.+?)\s*(!=|==)\s*(null)`).FindStringSubmatch(condition); len(nullMatch) > 0 {
		path := nullMatch[1]
		operator := nullMatch[2]

		// Tentar obter o valor
		value, err := pe.resolveValue(path, data)
		if err != nil || value == nil {
			// Considerar erro como nil para comparação
			isNull := (err != nil || value == nil)
			result.Success = (operator == "==" && isNull) || (operator == "!=" && !isNull)
			return result
		}

		isNull := (value == nil)
		result.Success = (operator == "==" && isNull) || (operator == "!=" && !isNull)
		return result
	}

	result.Error = fmt.Sprintf("formato de condição não reconhecido: %s", condition)
	return result
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
func (pe *PolicyEngine) GetResults() map[string]PolicyResult {
	return pe.Results
}

// IsSuccessful verifica se todas as políticas foram executadas com sucesso
func (pe *PolicyEngine) IsSuccessful() bool {
	for _, result := range pe.Results {
		if !result.Passed {
			return false
		}
	}
	return true
}

// compareNumbers compara dois números com base no operador fornecido
func (pe *PolicyEngine) compareNumbers(left float64, operator string, right float64) bool {
	switch operator {
	case "==":
		return left == right
	case "!=":
		return left != right
	case ">":
		return left > right
	case "<":
		return left < right
	case ">=":
		return left >= right
	case "<=":
		return left <= right
	default:
		return false
	}
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

// toArray tenta converter um valor para um array
func (pe *PolicyEngine) toArray(value interface{}) ([]interface{}, bool) {
	if value == nil {
		return nil, false
	}

	switch v := value.(type) {
	case []interface{}:
		return v, true
	}

	// Tenta converter através de reflexão
	valueOfV := reflect.ValueOf(value)
	if valueOfV.Kind() == reflect.Slice || valueOfV.Kind() == reflect.Array {
		length := valueOfV.Len()
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			result[i] = valueOfV.Index(i).Interface()
		}
		return result, true
	}

	return nil, false
}
