package policy

import (
	"fmt"
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
