package policy

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/raywall/aws-policy-engine-go/pkg/json/extractor"
)

// executeSetOperation executa uma operação SET nos dados
func (pe *PolicyEngine) executeSetOperation(targetPath, expression string) error {
	// Limpa os caminhos JSONPath
	targetPath = strings.TrimSpace(targetPath)
	if targetPath[0] == '$' {
		targetPath = targetPath[1:]
	}

	// Avalia a expressão para obter o valor
	value, err := pe.evaluateExpression(expression)
	if err != nil {
		return fmt.Errorf("erro ao avaliar expressão '%s': %v", expression, err)
	}

	ex, err := extractor.NewFromStruct(pe.request.Data)
	if err != nil {
		return err
	}

	// Define o valor no caminho especificado
	err = ex.SetValueByPath(targetPath, value)
	if err != nil {
		return fmt.Errorf("erro ao definir valor em '%s': %v", targetPath, err)
	}

	pe.request.Data = ex.Data.(map[string]interface{})
	return nil
}

// executeAddOperation executa uma operação ADD nos dados
func (pe *PolicyEngine) executeAddOperation(targetPath, expression string) error {
	// Limpa os caminhos JSONPATH
	targetPath = strings.TrimSpace(targetPath)
	if targetPath[0] == '$' {
		targetPath = targetPath[1:]
	}

	// Encontra todas as referências de variáveis na string JSON
	variableRegex := regexp.MustCompile(`\$\.([a-zA-Z0-9._]+)`)
	evaluatedJSON := variableRegex.ReplaceAllStringFunc(expression, func(match string) string {
		path := match[2:] // Remove o "$.", deixando apenas o caminho
		resolvedValue, err := pe.evaluateExpression(path)
		if err != nil {
			fmt.Printf("Aviso: Não foi possível resolver a referência '%s': %v\n", match, err)
			return match // Mantém a referência original em caso de erro
		}

		// Formatar o valor resolvido como string para substituição
		return fmt.Sprintf("%v", resolvedValue)
	})

	// Realiza a conversão do JSONObject
	var jsonObject map[string]interface{}
	err := json.Unmarshal([]byte(evaluatedJSON), &jsonObject)
	if err != nil {
		return fmt.Errorf("Erro ao decodificar objeto JSON: %v", err)
	}

	ex, err := extractor.NewFromStruct(pe.request.Data)
	if err != nil {
		return err
	}

	// Define o valor no caminho especificado
	err = ex.SetValueByPath(targetPath, jsonObject)
	if err != nil {
		return fmt.Errorf("erro ao definir valor em '%s': %v", targetPath, err)
	}

	pe.request.Data = ex.Data.(map[string]interface{})
	return nil
}
