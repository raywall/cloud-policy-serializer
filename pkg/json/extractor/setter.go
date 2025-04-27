package extractor

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// SetValueByPath define um valor em um objeto usando um caminho JSONPath
// Exemplo: SetValueByPath(data, ".user.address.city", "Nova York")
func (je *JSONExtractor) SetValueByPath(path string, value interface{}) error {
	if je.Data == nil {
		return errors.New("dados de origem não podem ser nulos")
	}

	// Normaliza o caminho
	// Remove o $ inicial se existir
	if strings.HasPrefix(path, "$") {
		path = path[1:]
	}

	// Remove o ponto inicial se existir
	if strings.HasPrefix(path, ".") {
		path = path[1:]
	}

	// Se o caminho estiver vazio, tente definir o valor diretamente
	if path == "" {
		return errors.New("caminho não pode ser vazio")
	}

	// Divide o caminho em segmentos
	segments := parsePath(path)
	if len(segments) == 0 {
		return errors.New("caminho inválido")
	}

	// Usa reflexão para navegar na estrutura e definir o valor
	return setValue(je.Data, segments, value)
}

// setValue define recursivamente um valor no caminho especificado
func setValue(current interface{}, segments []pathSegment, value interface{}) error {
	if len(segments) == 0 {
		return errors.New("caminho inválido: segmentos vazios")
	}

	segment := segments[0]
	remainingSegments := segments[1:]

	switch current := current.(type) {
	case map[string]interface{}:
		switch seg := segment.(type) {
		case propertySegment:
			if len(remainingSegments) == 0 {
				current[seg.name] = value
				return nil
			}

			next, exists := current[seg.name]
			if !exists {
				// Cria um novo objeto (mapa) por padrão para propriedades
				current[seg.name] = map[string]interface{}{}
				next = current[seg.name]
			}

			return setValue(next, remainingSegments, value)

		case arrayIndexSegment:
			// Verifica se current é um array (obtido através da função auxiliar)
			arr, ok := getArray(current)
			if !ok {
				return fmt.Errorf("propriedade não é um array para o índice %d", seg.index)
			}

			if seg.index < 0 || seg.index >= len(arr) {
				return fmt.Errorf("índice de array fora dos limites: %d", seg.index)
			}

			if len(remainingSegments) == 0 {
				arr[seg.index] = value
				return nil
			}

			return setValue(arr[seg.index], remainingSegments, value)
		}

	case []interface{}:
		if seg, ok := segment.(arrayIndexSegment); ok {
			if seg.index < 0 || seg.index > len(current) { // Permite expandir o slice
				newSlice := make([]interface{}, seg.index+1)
				copy(newSlice, current)
				current = newSlice
			}

			if len(remainingSegments) == 0 {
				current[seg.index] = value
				return nil
			}

			if current[seg.index] == nil {
				// Cria um novo objeto (mapa) por padrão para elementos de array
				if len(remainingSegments) > 0 {
					if _, isArray := remainingSegments[0].(arrayIndexSegment); isArray {
						current[seg.index] = []interface{}{}
					} else {
						current[seg.index] = map[string]interface{}{}
					}
				} else {
					current[seg.index] = value
					return nil
				}
			}

			return setValue(current[seg.index], remainingSegments, value)
		} else {
			return fmt.Errorf("não é possível acessar um índice de array em um array usando '%v'", segment)
		}

	default:
		v := reflect.ValueOf(current)

		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return errors.New("ponteiro nulo encontrado no caminho")
			}
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Map:
			switch seg := segment.(type) {
			case propertySegment:
				key := reflect.ValueOf(seg.name)

				if len(remainingSegments) == 0 {
					v.SetMapIndex(key, reflect.ValueOf(value))
					return nil
				}

				item := v.MapIndex(key)
				if !item.IsValid() {
					// Cria um novo objeto (mapa) por padrão
					item = reflect.MakeMap(reflect.TypeOf(map[string]interface{}{}))
					v.SetMapIndex(key, item)
				}

				return setValue(item.Interface(), remainingSegments, value)
			}

		case reflect.Slice, reflect.Array:
			if seg, ok := segment.(arrayIndexSegment); ok {
				if seg.index < 0 || seg.index >= v.Len() {
					return fmt.Errorf("índice de array fora dos limites: %d", seg.index)
				}

				if len(remainingSegments) == 0 {
					item := v.Index(seg.index)
					if item.CanSet() {
						item.Set(reflect.ValueOf(value))
						return nil
					}
					return fmt.Errorf("não é possível definir valor no índice %d", seg.index)
				}

				return setValue(v.Index(seg.index).Interface(), remainingSegments, value)
			}

		case reflect.Struct:
			if seg, ok := segment.(propertySegment); ok {
				field := v.FieldByName(strings.Title(seg.name))
				if !field.IsValid() {
					return fmt.Errorf("campo '%s' não encontrado na estrutura", seg.name)
				}
				if !field.CanSet() {
					return fmt.Errorf("campo '%s' não pode ser modificado", seg.name)
				}

				if len(remainingSegments) == 0 {
					if field.Type().AssignableTo(reflect.TypeOf(value)) {
						field.Set(reflect.ValueOf(value))
						return nil
					}
					return fmt.Errorf("tipo incompatível para atribuição ao campo '%s'", seg.name)
				}

				return setValue(field.Interface(), remainingSegments, value)
			}
		}
	}

	return fmt.Errorf("não foi possível definir o valor no caminho: tipo não suportado para %T", current)
}

// getArray obtém um array de um mapa usando a chave especial de array
func getArray(m map[string]interface{}) ([]interface{}, bool) {
	// Procura uma chave que represente um array (para fins de compatibilidade)
	for _, v := range m {
		if arr, ok := v.([]interface{}); ok {
			return arr, true
		}
	}
	return nil, false
}

func (a arrayIndexSegment) String() string {
	return fmt.Sprintf("[%d]", a.index)
}
