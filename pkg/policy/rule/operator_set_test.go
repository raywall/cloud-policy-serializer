package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	setConditionPayload = map[string]interface{}{
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
)

func TestPolicySetCondition(t *testing.T) {
	t.Run("", func(t *testing.T) {
		tr := NewTrimmedRule(`SET $.desconto = 15.0`)

		discountExpected := 15.0
		executedExpected := true
		passedExpected := true

		value := tr.Set("", setConditionPayload)

		assert.Equal(t, setConditionPayload["desconto"], discountExpected, "O valor do desconto deveria ser de 15.0")
		assert.Equal(t, value.Executed, executedExpected, "O atributo 'Executado' deveria ser verdadeiro")
		assert.Equal(t, value.Passed, passedExpected, "O atributo 'Passed' deveria ser verdadeiro")
	})

	t.Run("", func(t *testing.T) {
		tr := NewTrimmedRule(`SET $.idade = 22`)

		ageExpected := 22.0
		executedExpected := true
		passedExpected := true

		value := tr.Set("", setConditionPayload)

		assert.Equal(t, setConditionPayload["idade"], ageExpected, "O valor da 'Idade' deveria ser de 22")
		assert.Equal(t, value.Executed, executedExpected, "O atributo 'Executado' deveria ser verdadeiro")
		assert.Equal(t, value.Passed, passedExpected, "O atributo 'Passed' deveria ser verdadeiro")
	})
}
