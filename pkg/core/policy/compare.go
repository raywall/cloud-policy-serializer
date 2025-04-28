package policy

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/raywall/aws-policy-engine-go/pkg/core/utils"
	"github.com/raywall/aws-policy-engine-go/pkg/json/extractor"
)

// evaluateComparisonOperation compara dois valores usando o operador especificado
func (pe *PolicyEngine) evaluateComparisonOperation(left, right interface{}, operator string) (bool, error) {
	left, err := pe.extractValue(left)
	if err != nil {
		return false, err
	}

	right, err = pe.extractValue(right)
	if err != nil {
		return false, err
	}

	// Converter para tipos compatíveis se necessário
	var leftValue, rightValue utils.DataChecker
	leftValue = &utils.Data{Value: left}
	rightValue = &utils.Data{Value: right}

	// Se ambos são numéricos, fazer comparação numérica
	if leftValue.IsNumeric() && rightValue.IsNumeric() {
		leftNumber, rightNumber := leftValue.ToFloat(), rightValue.ToFloat()

		switch operator {
		case "==":
			return leftNumber == rightNumber, nil
		case "!=":
			return leftNumber != rightNumber, nil
		case "<":
			return leftNumber < rightNumber, nil
		case "<=":
			return leftNumber <= rightNumber, nil
		case ">":
			return leftNumber > rightNumber, nil
		case ">=":
			return leftNumber >= rightNumber, nil
		default:
			return false, fmt.Errorf("operador não suportado para números: %s", operator)
		}
	}

	// Para comparações não numéricas
	switch operator {
	case "==":
		return reflect.DeepEqual(left, right), nil
	case "!=":
		return !reflect.DeepEqual(left, right), nil
	case "IN":
		// Verificar se right é um array e se left está contido nele
		rightArray, ok := right.([]interface{})
		if !ok {
			return false, fmt.Errorf("o operador IN requer um array no lado direito")
		}

		for _, item := range rightArray {
			if reflect.DeepEqual(left, item) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("operador não suportado para tipos não numéricos: %s", operator)
	}
}

// extractValue valida e extrai valores de atributos
func (pe *PolicyEngine) extractValue(value interface{}) (interface{}, error) {
	var checker utils.DataChecker = &utils.Data{
		Value: value,
	}

	if !checker.IsNumericString() && strings.HasPrefix(value.(string), "$.") {
		ex, err := extractor.NewFromStruct(pe.request.Data)
		if err != nil {
			return nil, err
		}
		value, err = ex.Extract(value.(string))
		if err != nil {
			return nil, err
		}
	}

	if _, ok := value.(string); ok {
		value = utils.RemoveOuterQuotes(value.(string))
	}

	return value, nil
}
