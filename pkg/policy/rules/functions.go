package rules

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// --- Funções Utilitárias para map[string]interface{} (sem alterações) ---
// getValue recupera um valor de um map[string]interface{} aninhado usando um caminho separado por pontos.
// Exemplo de caminho: "user.address.zipcode" ou "items[0].name"
func getValue(data map[string]interface{}, path string) (interface{}, error) {
	path = strings.TrimPrefix(path, "$.")
	parts := strings.Split(path, ".")
	current := interface{}(data)
	for _, part := range parts {
		arrayMatch := regexp.MustCompile(`(\w+)\[(\d+)\]`).FindStringSubmatch(part)
		if len(arrayMatch) == 3 {
			key := arrayMatch[1]
			index, err := strconv.Atoi(arrayMatch[2])
			if err != nil {
				return nil, fmt.Errorf("índice de array inválido no caminho '%s': %s", part, err)
			}
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("não é possível acessar a chave '%s' em um tipo não-objeto (%T) para a parte do caminho '%s'", key, current, part)
			}
			arrInterface, exists := m[key]
			if !exists {
				return nil, fmt.Errorf("chave de array '%s' não encontrada para a parte do caminho '%s'", key, part)
			}
			arr, ok := arrInterface.([]interface{})
			if !ok {
				return nil, fmt.Errorf("chave '%s' não é um array para a parte do caminho '%s'", key, part)
			}
			if index < 0 || index >= len(arr) {
				return nil, fmt.Errorf("índice %d fora dos limites para o array '%s' (tamanho %d)", index, key, len(arr))
			}
			current = arr[index]
		} else {
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("não é possível acessar a chave '%s' em um tipo não-objeto (%T)", part, current)
			}
			value, exists := m[part]
			if !exists {
				return nil, nil
			}
			current = value
		}
	}
	return current, nil
}

// setValue define um valor em um map[string]interface{} aninhado usando um caminho separado por pontos.
// Cria mapas intermediários se eles não existirem.
func setValue(data map[string]interface{}, path string, value interface{}) error {
	path = strings.TrimPrefix(path, "$.")
	parts := strings.Split(path, ".")
	currentMap := data
	for i, part := range parts {
		arrayMatch := regexp.MustCompile(`(\w+)\[(\d+)\]`).FindStringSubmatch(part)
		if len(arrayMatch) == 3 {
			key := arrayMatch[1]
			index, err := strconv.Atoi(arrayMatch[2])
			if err != nil {
				return fmt.Errorf("índice de array inválido no caminho '%s': %s", part, err)
			}
			if _, exists := currentMap[key]; !exists {
				currentMap[key] = make([]interface{}, index+1)
			}
			arrInterface, ok := currentMap[key].([]interface{})
			if !ok {
				return fmt.Errorf("chave '%s' não é um array para operação SET, mas sim %T", key, currentMap[key])
			}
			if index >= len(arrInterface) {
				newArr := make([]interface{}, index+1)
				copy(newArr, arrInterface)
				arrInterface = newArr
				currentMap[key] = arrInterface
			}
			if i == len(parts)-1 {
				arrInterface[index] = value
				return nil
			}
			if arrInterface[index] == nil {
				arrInterface[index] = make(map[string]interface{})
			}
			if subMap, ok := arrInterface[index].(map[string]interface{}); ok {
				currentMap = subMap
			} else {
				return fmt.Errorf("não é possível SET caminho: elemento em '%s[%d]' não é um objeto, mas sim %T", key, index, arrInterface[index])
			}
		} else {
			if i == len(parts)-1 {
				currentMap[part] = value
				return nil
			}
			if _, exists := currentMap[part]; !exists {
				currentMap[part] = make(map[string]interface{})
			}
			nextMap, ok := currentMap[part].(map[string]interface{})
			if !ok {
				if currentMap[part] != nil {
					return fmt.Errorf("não é possível definir valor, segmento do caminho '%s' não é um mapa (é %T)", part, currentMap[part])
				}
				currentMap[part] = make(map[string]interface{})
				nextMap = currentMap[part].(map[string]interface{})
			}
			currentMap = nextMap
		}
	}
	return errors.New("caminho muito curto para operação set")
}
