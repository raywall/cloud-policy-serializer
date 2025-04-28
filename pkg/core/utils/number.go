package utils

import (
	"reflect"
	"strconv"
)

type (
	Data struct {
		Value interface{}
	}

	DataChecker interface {
		IsNumeric() bool
		IsNumericString() bool
		ToFloat() float64
	}
)

func (d *Data) IsNumeric() bool {
	switch d.Value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	case complex64, complex128:
		return true
	default:
		_, err := strconv.ParseFloat(d.Value.(string), 64)
		if err == nil {
			return true
		}

		// fmt.Printf("Tipo desconhecido: %v (%s)\n", v, reflect.TypeOf(v).String())
		return false
	}
}

func (d *Data) IsNumericString() bool {
	_, err := strconv.ParseFloat(d.Value.(string), 64)
	return err == nil
}

func (d *Data) ToFloat() float64 {
	switch val := d.Value.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(val).Int())
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(val).Uint())
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}
