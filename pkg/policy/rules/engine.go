package rules

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// --- Motor de Políticas ---
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

// evaluateMathExpression avalia uma expressão matemática simples (op1 operator op2).
func evaluateMathExpression(expressionStr string, data map[string]interface{}) (float64, string, error) {
	expressionStr = strings.TrimSpace(expressionStr)
	var op1Str, op2Str, operator string

	// Tenta encontrar operadores. A ordem pode importar se permitirmos expressões mais complexas no futuro.
	// Por agora, encontra o primeiro operador que divide em duas partes.
	operators := []string{"+", "-", "*", "/"} // Ordem de busca pode ser relevante para parsing simples
	foundOperator := false

	for _, op := range operators {
		// Garante que o operador não está no início (ex: número negativo) ou protegido
		// Para simplificar, vamos assumir que operadores são infixos e não unários no início da expressão.
		// Uma regex mais sofisticada seria melhor aqui.
		// Ex: "$.val * -5" não é suportado, mas "$.val * EXP(-5)" seria (se EXP aninhado fosse suportado).
		// Por agora, "$.val * -5" seria interpretado como op2Str = "-5".
		if strings.Contains(expressionStr, op) && !isOperatorProtected(expressionStr, op) {
			parts := strings.SplitN(expressionStr, op, 2)
			if len(parts) == 2 {
				// Verifica se o operador é realmente o principal ou parte de um número negativo
				// Ex: "$.campo -5" vs "$.campo-negativo"
				// Esta é uma heurística simples.
				potentialOp1 := strings.TrimSpace(parts[0])
				potentialOp2 := strings.TrimSpace(parts[1])

				// Evita falsos positivos como "campo-valor" sendo split em "-"
				// Se op1 ou op2 estiver vazio após o split, provavelmente não é o operador correto.
				if potentialOp1 == "" || potentialOp2 == "" {
					continue
				}
				// Se o operador for '-', verificar se é um menos unário no segundo operando
				// Ex: "$.a * -$.b" -> op1="$.a", op="*", op2="-$.b" (parseOperand lida com "-$.b")
				// Mas se for "$.a - $.b", op1="$.a", op="-", op2="$.b"

				op1Str = potentialOp1
				op2Str = potentialOp2
				operator = op
				foundOperator = true
				break
			}
		}
	}

	if !foundOperator {
		// Pode ser um único operando (um número literal ou um caminho $)
		num, err := parseOperand(expressionStr, data)
		if err != nil {
			return 0, fmt.Sprintf("Expressão '%s' não é um número nem uma expressão válida: %v", expressionStr, err), err
		}
		return num, fmt.Sprintf("%f", num), nil // Retorna o número como está
	}

	op1Num, err := parseOperand(op1Str, data)
	if err != nil {
		return 0, fmt.Sprintf("Erro no operando esquerdo ('%s') da expressão '%s': %v", op1Str, expressionStr, err), err
	}
	op2Num, err := parseOperand(op2Str, data)
	if err != nil {
		return 0, fmt.Sprintf("Erro no operando direito ('%s') da expressão '%s': %v", op2Str, expressionStr, err), err
	}

	var result float64
	switch operator {
	case "+":
		result = op1Num + op2Num
	case "-":
		result = op1Num - op2Num
	case "*":
		result = op1Num * op2Num
	case "/":
		if op2Num == 0 {
			err := errors.New("divisão por zero")
			return 0, fmt.Sprintf("%.2f %s %.2f -> ERRO: %v", op1Num, operator, op2Num, err), err
		}
		result = op1Num / op2Num
	default:
		err := fmt.Errorf("operador matemático desconhecido '%s' na expressão '%s'", operator, expressionStr)
		return 0, err.Error(), err
	}
	return result, fmt.Sprintf("%.2f %s %.2f = %.2f", op1Num, operator, op2Num, result), nil
}

func extractExpression(valueStr string) (string, bool) {
	valueStr = strings.TrimSpace(valueStr)
	if strings.HasPrefix(valueStr, "EXP(") && strings.HasSuffix(valueStr, ")") {
		return valueStr[4 : len(valueStr)-1], true
	}
	return "", false
}

// isOperatorProtected verifica se um operador (como " OR " ou "*") está dentro de aspas,
// o que significaria que faz parte de um valor literal de string e não um operador real.
// Esta é uma verificação simples e pode não cobrir todos os casos de escape.
func isOperatorProtected(rulePart string, operator string) bool {
	opIndex := strings.Index(rulePart, operator)
	if opIndex == -1 {
		return false
	}
	inSingleQuote, inDoubleQuote := false, false
	for i := 0; i < opIndex; i++ {
		char := rulePart[i]
		if char == '\'' {
			inSingleQuote = !inSingleQuote
		}
		if char == '"' {
			inDoubleQuote = !inDoubleQuote
		}
	}
	return inSingleQuote || inDoubleQuote
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

func compareEquals(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	aNum, aIsNum := convertToFloat64(a)
	bNum, bIsNum := convertToFloat64(b)
	if aIsNum && bIsNum {
		return aNum == bNum
	}

	// Se não são ambos números, compara os tipos e depois os valores
	// Isso é importante para "string" == "string" e bool == bool
	// mas também para evitar que 1 (int) seja igual a "1" (string) sem conversão explícita
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		return reflect.DeepEqual(a, b) // DeepEqual lida com strings, bools, etc.
	}

	// Se os tipos são diferentes e não são ambos convertíveis para float64, considera-os diferentes.
	// Ex: 10 (número) != "10" (string)
	return false
}
