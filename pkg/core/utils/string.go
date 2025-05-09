package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func RemoveOuterQuotes(s string) string {
	re := regexp.MustCompile(`^"(.*)"$`)
	return re.ReplaceAllString(s, `$1`)
}

// StringToArray converte uma string formatada como array JSON para um slice de strings.
func StringToArray(arrayStr string) ([]string, error) {
	arrayStr = strings.TrimSpace(arrayStr)
	if len(arrayStr) < 2 || arrayStr[0] != '[' || arrayStr[len(arrayStr)-1] != ']' {
		return nil, fmt.Errorf("string inválida para array: %s", arrayStr)
	}

	var strSlice []string
	err := json.Unmarshal([]byte(arrayStr), &strSlice)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar string JSON: %w", err)
	}
	return strSlice, nil
}

// IsArray verifica se uma string tem a formatação de um array JSON.
func IsArray(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return false
	}

	var temp interface{}
	return json.Unmarshal([]byte(s), &temp) == nil && reflect.TypeOf(temp).Kind() == reflect.Slice
}
