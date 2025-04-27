package policy_test

package policy_test

import (
	"encoding/json"
	"testing"

	"github.com/raywall/aws-policy-engine-go/pkg/core/policy"
	"github.com/raywall/aws-policy-engine-go/pkg/core/request"
)

func TestPolicyEngine(t *testing.T) {
	// Conteúdo do arquivo policy.yaml
	policyYaml := []byte(`
ValidarIdade:
- $.idade >= 18
- $.tipo == "adulto"

ValidarValorTransacao:
- $.valor > 0
- $.valor <= $.limiteMaximo
- $.moeda == "BRL" || $.moeda == "USD"

CalcularDesconto:
- $.valor > 100
- SET $.desconto = $.valor * 0.1
- IF $.cliente.tipo == "premium" THEN SET $.desconto = $.valor * 0.15

AplicarImpostos:
- SET $.impostos = {}
- SET $.impostos.iss = $.valor * 0.05
- IF $.tipo == "servico" THEN SET $.impostos.pis = $.valor * 0.0165

ValidarEndereco:
- $.endereco.cep != null
- $.endereco.cidade != null
- $.endereco.estado IN ["SP", "RJ", "MG", "RS"]

VerificarLimites:
- COUNT($.transacoes) < $.limites.maxTransacoes
- SUM(map($.transacoes, "valor")) + $.valor <= $.limites.valorTotal
`)

	// Inicializar o motor de políticas
	engine, err := policy.NewPolicyEngine(policyYaml)
	if err != nil {
		t.Fatalf("Erro ao criar o motor de políticas: %v", err)
	}

	// Teste 1: ValidarIdade (sucesso)
	t.Run("ValidarIdade_Sucesso", func(t *testing.T) {
		data := map[string]interface{}{
			"idade": 25,
			"tipo":  "adulto",
		}
		req := &request.Request{
			Policies: []string{"ValidarIdade"},
			Data:     data,
		}

		err := engine.ExecutePolicies(req)
		if err != nil {
			t.Fatalf("Erro ao executar políticas: %v", err)
		}

		results := engine.GetResults()
		if !results["ValidarIdade"].Success {
			t.Errorf("ValidarIdade deveria ter sucesso. Resultados: %+v", results["ValidarIdade"])
		}
	})

	// Teste 2: ValidarIdade (falha)
	t.Run("ValidarIdade_Falha", func(t *testing.T) {
		data := map[string]interface{}{
			"idade": 15,
			"tipo":  "adulto",
		}
		req := &request.Request{
			Policies: []string{"ValidarIdade"},
			Data:     data,
		}

		err := engine.ExecutePolicies(req)
		if err != nil {
			t.Fatalf("Erro ao executar políticas: %v", err)
		}

		results := engine.GetResults()
		if results["ValidarIdade"].Success {
			t.Errorf("ValidarIdade deveria falhar. Resultados: %+v", results["ValidarIdade"])
		}
	})

	// Teste 3: CalcularDesconto (cliente regular)
	t.Run("CalcularDesconto_Regular", func(t *testing.T) {
		data := map[string]interface{}{
			"valor": 200.0,
			"cliente": map[string]interface{}{
				"tipo": "regular",
			},
		}
		req := &request.Request{
			Policies: []string{"CalcularDesconto"},
			Data:     data,
		}

		err := engine.ExecutePolicies(req)
		if err != nil {
			t.Fatalf("Erro ao executar políticas: %v", err)
		}

		results := engine.GetResults()
		if !results["CalcularDesconto"].Success {
			t.Errorf("CalcularDesconto deveria ter sucesso. Resultados: %+v", results["CalcularDesconto"])
		}

		// Verificar se o desconto foi calculado corretamente (10%)
		desconto, ok := data["desconto"].(float64)
		if !ok {
			t.Errorf("Desconto não foi definido ou não é um número")
		} else if desconto != 20.0 {
			t.Errorf("Desconto calculado incorretamente. Esperado: 20.0, Recebido: %f", desconto)
		}
	})

	// Teste 4: CalcularDesconto (cliente premium)
	t.Run("CalcularDesconto_Premium", func(t *testing.T) {
		data := map[string]interface{}{
			"valor": 200.0,
			"cliente": map[string]interface{}{
				"tipo": "premium",
			},
		}
		req := &request.Request{
			Policies: []string{"CalcularDesconto"},
			Data:     data,
		}

		err := engine.ExecutePolicies(req)
		if err != nil {
			t.Fatalf("Erro ao executar políticas: %v", err)
		}

		results := engine.GetResults()
		if !results["CalcularDesconto"].Success {
			t.Errorf("CalcularDesconto deveria ter sucesso. Resultados: %+v", results["CalcularDesconto"])
		}

		// Verificar se o desconto foi calculado corretamente (15%)
		desconto, ok := data["desconto"].(float64)
		if !ok {
			t.Errorf("Desconto não foi definido ou não é um número")
		} else if desconto != 30.0 {
			t.Errorf("Desconto calculado incorretamente. Esperado: 30.0, Recebido: %f", desconto)
		}
	})

	// Teste 5: AplicarImpostos (para serviço)
	t.Run("AplicarImpostos_Servico", func(t *testing.T) {
		data := map[string]interface{}{
			"valor": 100.0,
			"tipo":  "servico",
		}
		req := &request.Request{
			Policies: []string{"AplicarImpostos"},
			Data:     data,
		}

		err := engine.ExecutePolicies(req)
		if err != nil {
			t.Fatalf("Erro ao executar políticas: %v", err)
		}

		results := engine.GetResults()
		if !results["AplicarImpostos"].Success {
			t.Errorf("AplicarImpostos deveria ter sucesso. Resultados: %+v", results["AplicarImpostos"])
		}

		// Verificar se os impostos foram calculados corretamente
		impostos, ok := data["impostos"].(map[string]interface{})
		if !ok {
			t.Errorf("Impostos não foram definidos ou não é um objeto")
			return
		}

		iss, okIss := impostos["iss"].(float64)
		pis, okPis := impostos["pis"].(float64)

		if !okIss || iss != 5.0 {
			t.Errorf("ISS calculado incorretamente. Esperado: 5.0, Recebido: %v", impostos["iss"])
		}

		if !okPis || pis != 1.65 {
			t.Errorf("PIS calculado incorretamente. Esperado: 1.65, Recebido: %v", impostos["pis"])
		}
	})

	// Teste 6: ValidarEndereco
	t.Run("ValidarEndereco_Sucesso", func(t *testing.T) {
		data := map[string]interface{}{
			"endereco": map[string]interface{}{
				"cep":    "01234-567",
				"cidade": "São Paulo",
				"estado": "SP",
			},
		}
		req := &request.Request{
			Policies: []string{"ValidarEndereco"},
			Data:     data,
		}

		err := engine.ExecutePolicies(req)
		if err != nil {
			t.Fatalf("Erro ao executar políticas: %v", err)
		}

		results := engine.GetResults()
		if !results["ValidarEndereco"].Success {
			t.Errorf("ValidarEndereco deveria ter sucesso. Resultados: %+v", results["ValidarEndereco"])
		}
	})

	// Teste 7: ValidarEndereco (falha - estado não na lista)
	t.Run("ValidarEndereco_Falha", func(t *testing.T) {
		data := map[string]interface{}{
			"endereco": map[string]interface{}{
				"cep":    "01234-567",
				"cidade": "Curitiba",
				"estado": "PR",
			},
		}
		req := &request.Request{
			Policies: []string{"ValidarEndereco"},
			Data:     data,
		}

		err := engine.ExecutePolicies(req)
		if err != nil {
			t.Fatalf("Erro ao executar políticas: %v", err)
		}

		results := engine.GetResults()
		if results["ValidarEndereco"].Success {
			t.Errorf("ValidarEndereco deveria falhar. Resultados: %+v", results["ValidarEndereco"])
		}
	})
}