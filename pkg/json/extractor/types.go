package extractor

import (
	"encoding/json"
	"fmt"
)

// JSONExtractor representa a estrutura principal para extração de dados JSON
type JSONExtractor struct {
	Data interface{}
}

// NewFromBytes cria um novo extrator a partir de bytes JSON
func NewFromBytes(jsonData []byte) (*JSONExtractor, error) {
	var data interface{}
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do JSON: %w", err)
	}
	return &JSONExtractor{Data: data}, nil
}

// NewFromString cria um novo extrator a partir de uma string JSON
func NewFromString(jsonStr string) (*JSONExtractor, error) {
	return NewFromBytes([]byte(jsonStr))
}

// NewFromMap cria um novo extrator a partir de um mapa já decodificado
func NewFromMap(data map[string]interface{}) *JSONExtractor {
	return &JSONExtractor{Data: data}
}

// NewFromStruct cria um novo extrator a partir de uma estrutura Go
func NewFromStruct(v interface{}) (*JSONExtractor, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter estrutura para JSON: %w", err)
	}
	return NewFromBytes(data)
}
