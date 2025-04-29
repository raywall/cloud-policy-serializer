package builder

import (
	"os"
)

// SchemaType representa os tipos suportados no schema JSON
type SchemaType string

const (
	TypeObject  SchemaType = "object"
	TypeArray   SchemaType = "array"
	TypeString  SchemaType = "string"
	TypeNumber  SchemaType = "number"
	TypeInteger SchemaType = "integer"
	TypeBoolean SchemaType = "boolean"
	TypeNull    SchemaType = "null"
)

// SchemaFormatter encapsula a funcionalidade de formatação baseada em Schema
type SchemaFormatter struct {
	Schema interface{}
}

// LoadSchemaFromFile carrega um schema a partir de um arquivo
func LoadSchemaFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
