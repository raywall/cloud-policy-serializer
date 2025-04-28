package policy

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MapFunction extrai uma propriedade específica de cada elemento em um array
func (pe *PolicyEngine) MapFunction(params ...interface{}) (interface{}, error) {
	if len(params) != 2 {
		return nil, errors.New("map requires exactly two arguments: array and property name")
	}

	// Primeiro parâmetro deve ser um array
	array, ok := params[0].([]interface{})
	if !ok {
		return nil, errors.New("first argument of map must be an array")
	}

	// Segundo parâmetro deve ser uma string (nome da propriedade)
	propName, ok := params[1].(string)
	if !ok {
		return nil, errors.New("second argument of map must be a string (property name)")
	}

	result := make([]interface{}, 0, len(array))
	for _, item := range array {
		// Verificar se o item é um mapa (objeto)
		obj, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("array item is not an object")
		}

		// Extrair a propriedade do objeto
		value, exists := obj[propName]
		if !exists {
			return nil, fmt.Errorf("property '%s' not found in object", propName)
		}

		result = append(result, value)
	}

	return result, nil
}

// SumFunction calcula a soma dos valores em um array
func (pe *PolicyEngine) SumFunction(params ...interface{}) (interface{}, error) {
	if len(params) != 1 {
		return nil, errors.New("sum requires exactly one argument: array of numbers")
	}

	var values []interface{}

	// O parâmetro pode ser um array direto ou o resultado de uma função Map
	switch v := params[0].(type) {
	case []interface{}:
		values = v
	default:
		return nil, errors.New("sum argument must be an array of numbers")
	}

	var sum float64
	for _, val := range values {
		// Converter para float64 conforme o tipo
		switch num := val.(type) {
		case float64:
			sum += num
		case float32:
			sum += float64(num)
		case int:
			sum += float64(num)
		case int32:
			sum += float64(num)
		case int64:
			sum += float64(num)
		case json.Number:
			f, err := num.Float64()
			if err != nil {
				return nil, fmt.Errorf("failed to convert json.Number to float64: %v", err)
			}
			sum += f
		default:
			return nil, fmt.Errorf("unsupported value type in array: %T", val)
		}
	}

	return sum, nil
}

// Função auxiliar para extrair valores de uma propriedade específica de um array
func extractArrayProperty(array []interface{}, property string) ([]interface{}, error) {
	result := make([]interface{}, 0, len(array))

	for _, item := range array {
		obj, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("array item is not an object")
		}

		value, exists := obj[property]
		if !exists {
			return nil, fmt.Errorf("property '%s' not found in object", property)
		}

		result = append(result, value)
	}

	return result, nil
}

// // Registrar as novas funções no engine
// func (engine *PolicyEngine) RegisterAdditionalFunctions() {
// 	engine.RegisterFunction("map", MapFunction)
// 	engine.RegisterFunction("SUM", SumFunction)
// }
//
// // Ajuste na função ParseExpression para processar corretamente a sintaxe "map($.transacoes, "valor")"
// func ParseExpression(expr string, ctx *validation.Context) (interface{}, error) {
// 	// Detectar padrão de função com mapas e arrays
// 	mapRegex := regexp.MustCompile(`(?i)map\(([^,]+),\s*"([^"]+)"\)`)
// 	match := mapRegex.FindStringSubmatch(expr)
//
// 	if match != nil {
// 		// Encontrou uma expressão map
// 		arrayPath := match[1]
// 		property := match[2]
//
// 		// Resolver o array pelo caminho
// 		array, err := ctx.ResolveValue(arrayPath)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to resolve array path '%s': %v", arrayPath, err)
// 		}
//
// 		arrayValue, ok := array.([]interface{})
// 		if !ok {
// 			return nil, fmt.Errorf("'%s' is not an array", arrayPath)
// 		}
//
// 		// Aplicar a função map
// 		result, err := ExtractArrayProperty(arrayValue, property)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		// Se a expressão completa for algo como SUM(map(...)), continuar processando
// 		if strings.HasPrefix(expr, "SUM(") && strings.HasSuffix(expr, ")") {
// 			return SumFunction(ctx, result)
// 		}
//
// 		return result, nil
// 	}
//
// 	// Outros padrões para reconhecer (adicionar conforme necessário)
// 	sumArrayRegex := regexp.MustCompile(`(?i)SUM\(([^)]+)\)`)
// 	sumMatch := sumArrayRegex.FindStringSubmatch(expr)
//
// 	if sumMatch != nil {
// 		// É uma expressão SUM
// 		innerExpr := sumMatch[1]
//
// 		// Verificar se é um caminho simples
// 		if strings.HasPrefix(innerExpr, "$.") {
// 			// Resolver o valor
// 			value, err := ctx.ResolveValue(innerExpr)
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to resolve value '%s': %v", innerExpr, err)
// 			}
//
// 			// Aplicar SUM
// 			return SumFunction(ctx, value)
// 		}
//
// 		// Se não for um caminho simples, pode ser algo mais complexo
// 		// como SUM(map(...)), que já foi tratado acima
// 		return nil, fmt.Errorf("unsupported expression within SUM: %s", innerExpr)
// 	}
//
// 	// Implementação existente do ParseExpression
// 	// ...
//
// 	return nil, fmt.Errorf("unsupported expression: %s", expr)
// }
