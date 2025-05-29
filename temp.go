package main

import (
    "encoding/json"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "gopkg.in/yaml.v3"
    "reflect"
)

// Policy representa uma política YAML com um nome e uma lista de regras.
type Policy struct {
    Name  string   `yaml:"name"`
    Rules []string `yaml:"rules"`
}

// getValue recupera um valor do mapa usando notação de ponto (ex.: $.idade).
func getValue(data map[string]interface{}, path string) (interface{}, error) {
    if strings.HasPrefix(path, "$.") {
        keys := strings.Split(strings.TrimPrefix(path, "$."), ".")
        current := data
        for i, key := range keys {
            if strings.Contains(key, "[") && strings.HasSuffix(key, "]") {
                arrayKey, index, subKey, err := parseArrayPath(key)
                if err != nil {
                    return nil, err
                }
                arr, ok := current[arrayKey].([]interface{})
                if !ok {
                    return nil, fmt.Errorf("path %s is not an array", arrayKey)
                }
                if index >= len(arr) {
                    return nil, fmt.Errorf("index %d out of bounds for %s", index, arrayKey)
                }
                current, ok = arr[index].(map[string]interface{})
                if !ok {
                    return nil, fmt.Errorf("array element at %s is not a map", path)
                }
                if subKey != "" {
                    if i == len(keys)-1 {
                        return current[subKey], nil
                    }
                    current, ok = current[subKey].(map[string]interface{})
                    if !ok {
                        return nil, fmt.Errorf("invalid subkey %s", subKey)
                    }
                } else if i == len(keys)-1 {
                    return arr[index], nil
                }
            } else {
                var ok bool
                current, ok = current[key].(map[string]interface{})
                if !ok {
                    if i == len(keys)-1 {
                        return data[key], nil
                    }
                    return nil, fmt.Errorf("invalid path: %s", path)
                }
            }
        }
        return current, nil
    }
    return parseLiteral(path)
}

// parseArrayPath analisa caminhos de array (ex.: teste[0].nome).
func parseArrayPath(key string) (string, int, string, error) {
    parts := strings.SplitN(key, "[", 2)
    arrayKey := parts[0]
    rest := strings.TrimSuffix(parts[1], "]")
    if strings.Contains(rest, ".") {
        subParts := strings.SplitN(rest, ".", 2)
        index, err := strconv.Atoi(subParts[0])
        if err != nil {
            return "", 0, "", fmt.Errorf("invalid array index: %s", subParts[0])
        }
        return arrayKey, index, subParts[1], nil
    }
    index, err := strconv.Atoi(rest)
    if err != nil {
        return "", 0, "", fmt.Errorf("invalid array index: %s", rest)
    }
    return arrayKey, index, "", nil
}

// setValue define um valor no mapa, com suporte a arrays.
func setValue(data map[string]interface{}, path string, value interface{}) error {
    if strings.HasPrefix(path, "$.") {
        keys := strings.Split(strings.TrimPrefix(path, "$."), ".")
        current := data
        for i, key := range keys[:len(keys)-1] {
            if strings.Contains(key, "[") && strings.HasSuffix(key, "]") {
                arrayKey, index, subKey, err := parseArrayPath(key)
                if err != nil {
                    return err
                }
                arr, ok := current[arrayKey].([]interface{})
                if !ok {
                    // Criar array se não existir
                    arr = make([]interface{}, index+1)
                    for j := range arr {
                        arr[j] = make(map[string]interface{})
                    }
                    current[arrayKey] = arr
                } else if index >= len(arr) {
                    // Expandir array se necessário
                    newArr := make([]interface{}, index+1)
                    copy(newArr, arr)
                    for j := len(arr); j <= index; j++ {
                        newArr[j] = make(map[string]interface{})
                    }
                    arr = newArr
                    current[arrayKey] = arr
                }
                current = arr[index].(map[string]interface{})
                if subKey != "" {
                    if i == len(keys)-2 {
                        current[subKey] = value
                    } else {
                        next, ok := current[subKey].(map[string]interface{})
                        if !ok {
                            next = make(map[string]interface{})
                            current[subKey] = next
                        }
                        current = next
                    }
                } else if i == len(keys)-2 {
                    arr[index] = value
                }
            } else {
                next, ok := current[key].(map[string]interface{})
                if !ok {
                    next = make(map[string]interface{})
                    current[key] = next
                }
                current = next
                if i == len(keys)-2 {
                    current[keys[len(keys)-1]] = value
                }
            }
        }
        return nil
    }
    return fmt.Errorf("invalid path: %s", path)
}

// evaluateExpression avalia uma expressão YAML (ex.: $.idade >= 18).
func evaluateExpression(data map[string]interface{}, expr string) (bool, error) {
    parts := strings.Split(expr, " ")
    if len(parts) < 3 {
        return false, fmt.Errorf("invalid expression: %s", expr)
    }

    left, op, right := parts[0], parts[1], strings.Join(parts[2:], " ")
    leftVal, err := getValue(data, left)
    if err != nil {
        return false, err
    }

    rightVal, err := parseValue(right, leftVal)
    if err != nil {
        return false, err
    }

    switch op {
    case "==":
        return reflect.DeepEqual(leftVal, rightVal), nil
    case ">=":
        return compareNumbers(leftVal, rightVal, ">=")
    case ">":
        return compareNumbers(leftVal, rightVal, ">")
    case "<=":
        return compareNumbers(leftVal, rightVal, "<=")
    case "<":
        return compareNumbers(leftVal, rightVal, "<")
    case "IN":
        return inArray(leftVal, rightVal)
    case "MATCHES":
        return matches(leftVal, rightVal)
    default:
        return false, fmt.Errorf("unsupported operator: %s", op)
    }
}

// arrayOperation executa operações em arrays (MAX, MIN, AVERAGE, SUM, COUNT).
func arrayOperation(data map[string]interface{}, op, path string) (interface{}, error) {
    val, err := getValue(data, path)
    if err != nil {
        return nil, err
    }
    arr, ok := val.([]interface{})
    if !ok {
        return nil, fmt.Errorf("path %s is not an array", path)
    }
    if len(arr) == 0 {
        return 0.0, nil
    }
    switch strings.ToUpper(op) {
    case "COUNT":
        return float64(len(arr)), nil
    case "SUM", "AVERAGE", "MAX", "MIN":
        sum := 0.0
        min := 0.0
        max := 0.0
        for i, item := range arr {
            val, ok := item.(float64)
            if !ok {
                return nil, fmt.Errorf("invalid number in array: %v", item)
            }
            sum += val
            if i == 0 {
                min, max = val, val
            } else {
                if val < min {
                    min = val
                }
                if val > max {
                    max = val
                }
            }
        }
        switch strings.ToUpper(op) {
        case "SUM":
            return sum, nil
        case "AVERAGE":
            return sum / float64(len(arr)), nil
        case "MAX":
            return max, nil
        case "MIN":
            return min, nil
        }
    }
    return nil, fmt.Errorf("unsupported array operation: %s", op)
}

// matches valida uma string contra uma expressão regular.
func matches(val, pattern interface{}) (bool, error) {
    str, ok := val.(string)
    if !ok {
        return false, fmt.Errorf("value %v is not a string", val)
    }
    pat, ok := pattern.(string)
    if !ok {
        return false, fmt.Errorf("pattern %v is not a string", pattern)
    }
    matched, err := regexp.MatchString(strings.Trim(pat, `"`), str)
    if err != nil {
        return false, fmt.Errorf("invalid regex pattern: %s", pat)
    }
    return matched, nil
}

// evaluatePolicy avalia uma política YAML.
func evaluatePolicy(data map[string]interface{}, policy Policy) (bool, error) {
    for _, rule := range policy.Rules {
        if strings.HasPrefix(rule, "SET ") {
            parts := strings.SplitN(rule, "=", 2)
            if len(parts) != 2 {
                return false, fmt.Errorf("invalid SET rule: %s", rule)
            }
            path := strings.TrimSpace(strings.Split(parts[0], " ")[1])
            expr := strings.TrimSpace(parts[1])
            if strings.HasPrefix(expr, "EXP(") {
                expr = strings.TrimSuffix(strings.TrimPrefix(expr, "EXP("), ")")
                parts := strings.Split(expr, "*")
                if len(parts) != 2 {
                    return false, fmt.Errorf("invalid EXP expression: %s", expr)
                }
                val, err := getValue(data, strings.TrimSpace(parts[0]))
                if err != nil {
                    return false, err
                }
                numVal, ok := val.(float64)
                if !ok {
                    return false, fmt.Errorf("invalid number in EXP: %v", val)
                }
                multiplier, err := parseFloat(strings.TrimSpace(parts[1]))
                if err != nil {
                    return false, err
                }
                if err := setValue(data, path, numVal*multiplier); err != nil {
                    return false, err
                }
            } else if strings.HasPrefix(expr, "MAX(") || strings.HasPrefix(expr, "MIN(") ||
                strings.HasPrefix(expr, "AVERAGE(") || strings.HasPrefix(expr, "SUM(") ||
                strings.HasPrefix(expr, "COUNT(") {
                op := strings.Split(expr, "(")[0]
                pathExpr := strings.TrimSuffix(strings.Split(expr, "(")[1], ")")
                result, err := arrayOperation(data, op, pathExpr)
                if err != nil {
                    return false, err
                }
                if err := setValue(data, path, result); err != nil {
                    return false, err
                }
            } else {
                val, err := parseValue(expr, nil)
                if err != nil {
                    return false, err
                }
                if err := setValue(data, path, val); err != nil {
                    return false, err
                }
            }
        } else if strings.HasPrefix(rule, "ADD ") {
            parts := strings.SplitN(rule, " TO ", 2)
            if len(parts) != 2 {
                return false, fmt.Errorf("invalid ADD rule: %s", rule)
            }
            item := strings.TrimSpace(strings.Split(parts[0], " ")[1])
            path := strings.TrimSpace(parts[1])
            var newItem interface{}
            if err := json.Unmarshal([]byte(item), &newItem); err != nil {
                return false, fmt.Errorf("invalid ADD item: %s", item)
            }
            current, err := getValue(data, path)
            if err != nil {
                // Criar array se não existir
                if err := setValue(data, path, []interface{}{newItem}); err != nil {
                    return false, err
                }
            } else {
                arr, ok := current.([]interface{})
                if !ok {
                    return false, fmt.Errorf("path %s is not an array", path)
                }
                arr = append(arr, newItem)
                if err := setValue(data, path, arr); err != nil {
                    return false, err
                }
            }
        } else if strings.HasPrefix(rule, "COUNT(") || strings.HasPrefix(rule, "SUM(") ||
            strings.HasPrefix(rule, "MAX(") || strings.HasPrefix(rule, "MIN(") ||
            strings.HasPrefix(rule, "AVERAGE(") {
            parts := strings.SplitN(rule, ")", 2)
            if len(parts) != 2 {
                return false, fmt.Errorf("invalid array operation rule: %s", rule)
            }
            op := strings.Split(parts[0], "(")[0]
            path := strings.TrimPrefix(parts[0], op+"(")
            opParts := strings.Split(parts[1], " ")
            if len(opParts) < 2 {
                return false, fmt.Errorf("invalid comparison: %s", rule)
            }
            opComp, right := opParts[0], strings.Join(opParts[1:], " ")
            result, err := arrayOperation(data, op, path)
            if err != nil {
                return false, err
            }
            rightVal, err := parseFloat(right)
            if err != nil {
                return false, err
            }
            compResult, err := compareNumbers(result.(float64), rightVal, opComp)
            if err != nil || !compResult {
                return false, err
            }
        } else {
            result, err := evaluateExpression(data, rule)
            if err != nil || !result {
                return false, err
            }
        }
    }
    return true, nil
}

// Funções auxiliares
func parseValue(val string, reference interface{}) (interface{}, error) {
    if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
        var arr []interface{}
        if err := json.Unmarshal([]byte(val), &arr); err != nil {
            return nil, fmt.Errorf("invalid array: %s", val)
        }
        return arr, nil
    }
    switch reference.(type) {
    case float64:
        return parseFloat(val)
    case string:
        return strings.Trim(val, `"'`), nil
    default:
        return parseLiteral(val)
    }
}

func parseFloat(val string) (float64, error) {
    return strconv.ParseFloat(val, 64)
}

func compareNumbers(left, right interface{}, op string) (bool, error) {
    l, lok := left.(float64)
    r, rok := right.(float64)
    if !lok || !rok {
        return false, fmt.Errorf("invalid number comparison: %v %s %v", left, op, right)
    }
    switch op {
    case ">=":
        return l >= r, nil
    case ">":
        return l > r, nil
    case "<=":
        return l <= r, nil
    case "<":
        return l < r, nil
    default:
        return false, fmt.Errorf("unsupported number operator: %s", op)
    }
}

func inArray(val, arr interface{}) (bool, error) {
    arrVal, ok := arr.([]interface{})
    if !ok {
        return false, fmt.Errorf("invalid array for IN: %v", arr)
    }
    for _, item := range arrVal {
        if reflect.DeepEqual(val, item) {
            return true, nil
        }
    }
    return false, nil
}

func parseLiteral(val string) (interface{}, error) {
    if val == "null" {
        return nil, nil
    }
    if i, err := strconv.Atoi(val); err == nil {
        return float64(i), nil
    }
    if f, err := strconv.ParseFloat(val, 64); err == nil {
        return f, nil
    }
    if b, err := strconv.ParseBool(val); err == nil {
        return b, nil
    }
    return strings.Trim(val, `"'`), nil
}

// Exemplo de uso
func main() {
    conditionPayload := map[string]interface{}{
        "valor":        150.00,
        "limiteMaximo": 500.00,
        "moeda":        "BRL",
        "idade":        21,
        "tipo":         "servico",
        "cliente": map[string]interface{}{
            "tipo": "premium",
        },
        "endereco": map[string]interface{}{
            "cep":    "01234-567",
            "cidade": "São Paulo",
            "estado": "SP",
        },
        "transacoes": []interface{}{
            map[string]interface{}{"id": "t1", "valor": 50.00},
            map[string]interface{}{"id": "t2", "valor": 75.00},
        },
        "limites": map[string]interface{}{
            "maxTransacoes": 10,
            "valorTotal":    1000.00,
        },
    }

    policyYAML := `
- name: Regra
  rules:
    - $.idade >= 18
    - $.tipo == "adulto"
    - SET $.desconto = EXP($.valor * 0.1)
    - SET $.teste[0].nome = "ray"
    - SET $.maxTransacao = MAX($.transacoes[*].valor)
    - SET $.minTransacao = MIN($.transacoes[*].valor)
    - SET $.mediaTransacao = AVERAGE($.transacoes[*].valor)
    - $.endereco.cep MATCHES "^\\d{5}-\\d{3}$"
    - ADD [{"id":"t3","valor":100.00}] TO $.transacoes
`

    var policies []Policy
    if err := yaml.Unmarshal([]byte(policyYAML), &policies); err != nil {
        fmt.Printf("Error parsing YAML: %v\n", err)
        return
    }

    for _, policy := range policies {
        result, err := evaluatePolicy(conditionPayload, policy)
        if err != nil {
            fmt.Printf("Error evaluating policy %s: %v\n", policy.Name, err)
            continue
        }
        fmt.Printf("Policy %s: %v\n", policy.Name, result)
    }

    output, _ := json.MarshalIndent(conditionPayload, "", "  ")
    fmt.Println("Updated Data:", string(output))
}