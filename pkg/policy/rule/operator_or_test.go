package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	orConditionPayload = map[string]interface{}{
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

func TestPolicyOrCondition(t *testing.T) {
	t.Run("", func(t *testing.T) {
		tr := NewTrimmedRule(`$.moeda == "BRL" OR $.moeda == "USD"`)

		executedExpected := true
		passedExpected := true

		value := tr.Or("", orConditionPayload)

		assert.Equal(t, value.Executed, executedExpected, "O atributo 'Executado' deveria ser verdadeiro")
		assert.Equal(t, value.Passed, passedExpected, "O atributo 'Passed' deveria ser verdadeiro")
	})

	t.Run("", func(t *testing.T) {
		tr := NewTrimmedRule(`$.moeda == "EUR" OR $.idade == 21`)

		executedExpected := true
		passedExpected := true

		value := tr.Or("", orConditionPayload)

		assert.Equal(t, value.Executed, executedExpected, "O atributo 'Executado' deveria ser verdadeiro")
		assert.Equal(t, value.Passed, passedExpected, "O atributo 'Passed' deveria ser verdadeiro")
	})

	t.Run("", func(t *testing.T) {
		tr := NewTrimmedRule(`$.moeda != "EUR" OR $.idade == 25`)

		executedExpected := true
		passedExpected := true

		value := tr.Or("", orConditionPayload)

		assert.Equal(t, value.Executed, executedExpected, "O atributo 'Executado' deveria ser verdadeiro")
		assert.Equal(t, value.Passed, passedExpected, "O atributo 'Passed' deveria ser verdadeiro")
	})

	t.Run("", func(t *testing.T) {
		tr := NewTrimmedRule(`$.moeda == "EUR" OR $.idade == 25`)

		executedExpected := true
		passedExpected := false

		value := tr.Or("", orConditionPayload)

		assert.Equal(t, value.Executed, executedExpected, "O atributo 'Executado' deveria ser verdadeiro")
		assert.Equal(t, value.Passed, passedExpected, "O atributo 'Passed' deveria ser verdadeiro")
	})
}
