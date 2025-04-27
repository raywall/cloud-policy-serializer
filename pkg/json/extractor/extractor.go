package extractor

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Extract extrai valor usando uma expressão de caminho
func (je *JSONExtractor) Extract(path string) (interface{}, error) {
	// Verificar se o caminho começa com "$."
	if !strings.HasPrefix(path, "$.") {
		return nil, errors.New("o caminho deve começar com '$.'")
	}

	// Remover o prefixo "$."
	path = path[2:]

	// Dividir o caminho em segmentos
	segments := parsePath(path)

	// Começar com o dado raiz
	current := je.Data

	// Percorrer todos os segmentos do caminho
	for _, segment := range segments {
		// Verificar se é um acesso de array
		arrayAccess := regexp.MustCompile(`^(.+)\[(\d+)\]$`).FindStringSubmatch(segment)

		if arrayAccess != nil {
			// É um acesso de array: extrair o nome da propriedade e o índice
			property := arrayAccess[1]
			index, _ := strconv.Atoi(arrayAccess[2])

			// Se há uma propriedade antes do índice
			if property != "" {
				// Acessar a propriedade do objeto
				obj, ok := current.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("não é possível acessar a propriedade '%s' em um valor não-objeto", property)
				}

				value, exists := obj[property]
				if !exists {
					return nil, fmt.Errorf("propriedade '%s' não encontrada", property)
				}

				// Verificar se o valor é um array
				arr, ok := value.([]interface{})
				if !ok {
					return nil, fmt.Errorf("'%s' não é um array", property)
				}

				// Verificar se o índice está dentro dos limites
				if index < 0 || index >= len(arr) {
					return nil, fmt.Errorf("índice %d fora dos limites para o array '%s' (tamanho: %d)", index, property, len(arr))
				}

				// Atualizar o valor atual para o elemento do array
				current = arr[index]
			} else {
				// É um acesso direto de array sem propriedade
				arr, ok := current.([]interface{})
				if !ok {
					return nil, errors.New("não é possível acessar um índice em um valor não-array")
				}

				// Verificar se o índice está dentro dos limites
				if index < 0 || index >= len(arr) {
					return nil, fmt.Errorf("índice %d fora dos limites (tamanho: %d)", index, len(arr))
				}

				// Atualizar o valor atual para o elemento do array
				current = arr[index]
			}
		} else {
			// É um acesso de propriedade simples
			obj, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("não é possível acessar a propriedade '%s' em um valor não-objeto", segment)
			}

			value, exists := obj[segment]
			if !exists {
				return nil, fmt.Errorf("propriedade '%s' não encontrada", segment)
			}

			// Atualizar o valor atual
			current = value
		}
	}

	return current, nil
}

// ExtractString extrai um valor e o converte para string
func (je *JSONExtractor) ExtractString(path string) (string, error) {
	result, err := je.Extract(path)
	if err != nil {
		return "", err
	}

	strValue, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("o valor em '%s' não é uma string", path)
	}

	return strValue, nil
}

// ExtractInt extrai um valor e o converte para int
func (je *JSONExtractor) ExtractInt(path string) (int, error) {
	result, err := je.Extract(path)
	if err != nil {
		return 0, err
	}

	// Tentar converter diferentes tipos numéricos para int
	switch v := result.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case json.Number:
		intVal, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("o valor em '%s' não pode ser convertido para int: %w", path, err)
		}
		return int(intVal), nil
	default:
		return 0, fmt.Errorf("o valor em '%s' não é um número", path)
	}
}

// ExtractFloat extrai um valor e o converte para float64
func (je *JSONExtractor) ExtractFloat(path string) (float64, error) {
	result, err := je.Extract(path)
	if err != nil {
		return 0, err
	}

	// Tentar converter diferentes tipos numéricos para float64
	switch v := result.(type) {
	case int:
		return float64(v), nil
	case float64:
		return v, nil
	case json.Number:
		floatVal, err := v.Float64()
		if err != nil {
			return 0, fmt.Errorf("o valor em '%s' não pode ser convertido para float: %w", path, err)
		}
		return floatVal, nil
	default:
		return 0, fmt.Errorf("o valor em '%s' não é um número", path)
	}
}

// ExtractBool extrai um valor e o converte para bool
func (je *JSONExtractor) ExtractBool(path string) (bool, error) {
	result, err := je.Extract(path)
	if err != nil {
		return false, err
	}

	boolValue, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("o valor em '%s' não é um booleano", path)
	}

	return boolValue, nil
}

// ExtractArray extrai um valor e o converte para um array de interface{}
func (je *JSONExtractor) ExtractArray(path string) ([]interface{}, error) {
	result, err := je.Extract(path)
	if err != nil {
		return nil, err
	}

	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("o valor em '%s' não é um array", path)
	}

	return arr, nil
}

// ExtractMap extrai um valor e o converte para um mapa
func (je *JSONExtractor) ExtractMap(path string) (map[string]interface{}, error) {
	result, err := je.Extract(path)
	if err != nil {
		return nil, err
	}

	obj, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("o valor em '%s' não é um objeto", path)
	}

	return obj, nil
}

// ExtractInto extrai um valor e o decodifica em uma estrutura Go
func (je *JSONExtractor) ExtractInto(path string, v interface{}) error {
	result, err := je.Extract(path)
	if err != nil {
		return err
	}

	// Converter o resultado para JSON e depois decodificar na estrutura
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("erro ao converter para JSON: %w", err)
	}

	err = json.Unmarshal(jsonData, v)
	if err != nil {
		return fmt.Errorf("erro ao decodificar em estrutura: %w", err)
	}

	return nil
}
