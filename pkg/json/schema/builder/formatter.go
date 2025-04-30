package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/raywall/aws-policy-engine-go/pkg/json/schema"
)

// NewSchema cria um novo objeto Schema a partir de um JSON
func NewSchemaFormatter(s *schema.Schema) (*SchemaFormatter, error) {
	return &SchemaFormatter{
		Schema: s,
	}, nil
}

// FormatResponse formata uma resposta conforme o schema
func (sf *SchemaFormatter) FormatResponse(response map[string]interface{}) (map[string]interface{}, error) {
	// Extrai o campo "data" se existir
	var data map[string]interface{}
	if dataValue, exists := response["data"]; exists {
		if dataMap, ok := dataValue.(map[string]interface{}); ok {
			data = dataMap
		} else {
			return nil, errors.New("campo 'data' não é um objeto válido")
		}
	} else {
		// Se não existe campo data, consideramos o próprio response como dados
		data = response
	}

	// Valida e formata os dados conforme o schema
	formatted, err := sf.validateAndFormat(data, *sf.Schema)
	if err != nil {
		return nil, err
	}

	// Retorna o resultado formatado
	return formatted.(map[string]interface{}), nil
}

// PrettyJSON formata um objeto como JSON com indentação
func PrettyJSON(data interface{}) string {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Erro ao formatar JSON: %v", err)
	}
	return string(bytes)
}

// validateAndFormat valida e formata um valor de acordo com um schema
func (sf *SchemaFormatter) validateAndFormat(value interface{}, s map[string]interface{}) (interface{}, error) {
	// Obter o tipo do valor esperado pelo schema
	typeValue, exists := s["type"]
	if !exists {
		return value, nil // Se não tem tipo definido, retorna o valor como está
	}

	schemaType, ok := typeValue.(string)
	if !ok {
		return nil, errors.New("tipo no schema não é uma string")
	}

	// Converter para SchemaType
	valueType := SchemaType(schemaType)

	// Validar conforme o tipo
	switch valueType {
	case TypeObject:
		return sf.formatObject(value, s)
	case TypeArray:
		return sf.formatArray(value, s)
	case TypeString:
		return sf.formatString(value, s)
	case TypeNumber:
		return sf.formatNumber(value, s)
	case TypeInteger:
		return sf.formatInteger(value, s)
	case TypeBoolean:
		return sf.formatBoolean(value, s)
	case TypeNull:
		if value == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("valor não é null: %v", value)
	default:
		return nil, fmt.Errorf("tipo de schema não suportado: %s", valueType)
	}
}

// formatObject formata um objeto de acordo com o schema
func (sf *SchemaFormatter) formatObject(value interface{}, s map[string]interface{}) (interface{}, error) {
	// Verificar se o valor é um objeto
	obj, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("valor não é um objeto: %v", value)
	}

	// Resultado formatado
	result := make(map[string]interface{})

	// Verificar propriedades definidas no schema
	properties, hasProps := s["properties"].(map[string]interface{})
	if !hasProps {
		return obj, nil // Sem propriedades definidas, retorna o objeto como está
	}

	// Verificar propriedades adicionais
	additionalProps := true
	if ap, exists := s["additionalProperties"]; exists {
		if apBool, ok := ap.(bool); ok {
			additionalProps = apBool
		}
	}

	// Verificar propriedades obrigatórias
	requiredProps := make(map[string]bool)
	if required, exists := s["required"].([]interface{}); exists {
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				requiredProps[reqStr] = true
			}
		}
	}

	// Processar cada propriedade definida no schema
	for propName, propSchema := range properties {
		propSchemaMap, ok := propSchema.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema inválido para propriedade '%s'", propName)
		}

		// Verificar se a propriedade existe no objeto
		if propValue, exists := obj[propName]; exists {
			// Validar e formatar o valor da propriedade
			formattedValue, err := sf.validateAndFormat(propValue, propSchemaMap)
			if err != nil {
				return nil, fmt.Errorf("erro na propriedade '%s': %w", propName, err)
			}
			result[propName] = formattedValue
		} else if requiredProps[propName] {
			// Propriedade obrigatória ausente
			return nil, fmt.Errorf("propriedade obrigatória '%s' ausente", propName)
		}
	}

	// Se não permite propriedades adicionais, verificar se existem propriedades não definidas
	if !additionalProps {
		for propName := range obj {
			if _, defined := properties[propName]; !defined {
				return nil, fmt.Errorf("propriedade adicional não permitida: '%s'", propName)
			}
		}
	} else {
		// Se propriedades adicionais são permitidas, copiar as que não estão no schema
		for propName, propValue := range obj {
			if _, defined := properties[propName]; !defined {
				result[propName] = propValue
			}
		}
	}

	return result, nil
}

// formatString formata e valida uma string de acordo com o schema
func (sf *SchemaFormatter) formatString(value interface{}, s map[string]interface{}) (interface{}, error) {
	// Converter para string
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case json.Number:
		strValue = string(v)
	default:
		return nil, fmt.Errorf("valor não pode ser convertido para string: %v", value)
	}

	// Validar minLength
	if minLength, exists := s["minLength"].(float64); exists {
		if float64(len(strValue)) < minLength {
			return nil, fmt.Errorf("string com tamanho menor que o mínimo permitido (%d): %s", int(minLength), strValue)
		}
	}

	// Validar maxLength
	if maxLength, exists := s["maxLength"].(float64); exists {
		if float64(len(strValue)) > maxLength {
			return nil, fmt.Errorf("string com tamanho maior que o máximo permitido (%d): %s", int(maxLength), strValue)
		}
	}

	// Validar pattern (expressão regular)
	if pattern, exists := s["pattern"].(string); exists {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("expressão regular inválida no schema: %s", pattern)
		}
		if !re.MatchString(strValue) {
			return nil, fmt.Errorf("string não corresponde ao padrão '%s': %s", pattern, strValue)
		}
	}

	// Validar enum (valores permitidos)
	if enum, exists := s["enum"].([]interface{}); exists {
		found := false
		for _, allowedValue := range enum {
			if allowedStr, ok := allowedValue.(string); ok && allowedStr == strValue {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("valor não está entre os permitidos: %s", strValue)
		}
	}

	return strValue, nil
}

// formatNumber formata e valida um número de acordo com o schema
func (sf *SchemaFormatter) formatNumber(value interface{}, s map[string]interface{}) (interface{}, error) {
	var numValue float64

	// Converter para número
	switch v := value.(type) {
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case json.Number:
		var err error
		numValue, err = v.Float64()
		if err != nil {
			return nil, fmt.Errorf("erro ao converter para número: %w", err)
		}
	case string:
		var err error
		numValue, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("erro ao converter string para número: %w", err)
		}
	default:
		return nil, fmt.Errorf("valor não pode ser convertido para número: %v", value)
	}

	// Validar minimum
	if minimum, exists := s["minimum"].(float64); exists {
		if numValue < minimum {
			return nil, fmt.Errorf("número menor que o mínimo permitido (%f): %f", minimum, numValue)
		}
	}

	// Validar maximum
	if maximum, exists := s["maximum"].(float64); exists {
		if numValue > maximum {
			return nil, fmt.Errorf("número maior que o máximo permitido (%f): %f", maximum, numValue)
		}
	}

	// Validar multipleOf
	if multipleOf, exists := s["multipleOf"].(float64); exists {
		if math.Mod(numValue, multipleOf) != 0 {
			return nil, fmt.Errorf("número não é múltiplo de %f: %f", multipleOf, numValue)
		}
	}

	return numValue, nil
}

// formatInteger formata e valida um inteiro de acordo com o schema
func (sf *SchemaFormatter) formatInteger(value interface{}, s map[string]interface{}) (interface{}, error) {
	var intValue int64

	// Converter para inteiro
	switch v := value.(type) {
	case int:
		intValue = int64(v)
	case int64:
		intValue = v
	case float64:
		// Verificar se o float é realmente um inteiro
		if v != float64(int64(v)) {
			return nil, fmt.Errorf("valor não é um inteiro: %f", v)
		}
		intValue = int64(v)
	case json.Number:
		var err error
		// Tentar converter para int64
		intValue, err = v.Int64()
		if err != nil {
			// Se falhar, tentar via float
			f, err := v.Float64()
			if err != nil || f != float64(int64(f)) {
				return nil, fmt.Errorf("erro ao converter para inteiro: %w", err)
			}
			intValue = int64(f)
		}
	case string:
		var err error
		intValue, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("erro ao converter string para inteiro: %w", err)
		}
	default:
		return nil, fmt.Errorf("valor não pode ser convertido para inteiro: %v", value)
	}

	// Validar minimum
	if minimum, exists := s["minimum"].(float64); exists {
		if float64(intValue) < minimum {
			return nil, fmt.Errorf("inteiro menor que o mínimo permitido (%d): %d", int(minimum), intValue)
		}
	}

	// Validar maximum
	if maximum, exists := s["maximum"].(float64); exists {
		if float64(intValue) > maximum {
			return nil, fmt.Errorf("inteiro maior que o máximo permitido (%d): %d", int(maximum), intValue)
		}
	}

	return intValue, nil
}

// formatBoolean formata e valida um booleano de acordo com o schema
func (sf *SchemaFormatter) formatBoolean(value interface{}, s map[string]interface{}) (interface{}, error) {
	// Verificar e converter para booleano
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		if v == "true" {
			return true, nil
		} else if v == "false" {
			return false, nil
		}
		return nil, fmt.Errorf("string não pode ser convertida para booleano: %s", v)
	default:
		return nil, fmt.Errorf("valor não pode ser convertido para booleano: %v", value)
	}
}

// formatArray formata um array de acordo com o schema
func (sf *SchemaFormatter) formatArray(value interface{}, s map[string]interface{}) (interface{}, error) {
	// Verificar se o valor é um array
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("valor não é um array: %v", value)
	}

	// Verificar se há definição de items no schema
	itemsSchema, hasItems := s["items"].(map[string]interface{})
	if !hasItems {
		return arr, nil // Sem definição de items, retorna o array como está
	}

	// Formatar cada item do array
	result := make([]interface{}, len(arr))
	for i, item := range arr {
		formattedItem, err := sf.validateAndFormat(item, itemsSchema)
		if err != nil {
			return nil, fmt.Errorf("erro no item %d do array: %w", i, err)
		}
		result[i] = formattedItem
	}

	return result, nil
}
