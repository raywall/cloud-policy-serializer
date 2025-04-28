package policy

import (
	"encoding/json"
	"fmt"
)

// convertToFloat64 converte diferentes tipos para float64
func convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	default:
		return 0, fmt.Errorf("tipo não suportado para conversão para float64: %T", value)
	}
}
