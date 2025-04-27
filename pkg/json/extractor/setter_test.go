package extractor_test

import (
	"encoding/json"
	"testing"

	"github.com/raywall/aws-policy-engine-go/pkg/json/extractor"
)

func TestSetValueByPath(t *testing.T) {
	// Teste 1: Definir um valor simples
	t.Run("SetSimpleValue", func(t *testing.T) {
		data := map[string]interface{}{}
		
		err := extractor.SetValueByPath(data, "nome", "João")
		if err != nil {
			t.Fatalf("Erro ao definir valor: %v", err)
		}
		
		if data["nome"] != "João" {
			t.Errorf("Valor não definido corretamente. Esperado: João, Recebido: %v", data["nome"])
		}
	})
	
	// Teste 2: Definir um valor aninhado (criar estrutura)
	t.Run("SetNestedValueCreateStructure", func(t *testing.T) {
		data := map[string]interface{}{}
		
		err := extractor.SetValueByPath(data, "usuario.endereco.cidade", "São Paulo")
		if err != nil {
			t.Fatalf("Erro ao definir valor: %v", err)
		}
		
		usuario, ok := data["usuario"].(map[string]interface{})
		if !ok {
			t.Fatalf("Estrutura 'usuario' não criada corretamente")
		}
		
		endereco, ok := usuario["endereco"].(map[string]interface{})
		if !ok {
			t.Fatalf("Estrutura 'endereco' não criada corretamente")
		}
		
		if endereco["cidade"] != "São Paulo" {
			t.Errorf("Valor não definido corretamente. Esperado: São Paulo, Recebido: %v", endereco["cidade"])
		}
	})
	
	// Teste 3: Definir valor em um array
	t.Run("SetArrayValue", func(t *testing.T) {
		data := map[string]interface{}{
			"items": []interface{}{"item1", "item2", "item3"},
		}
		
		err := extractor.SetValueByPath(data, "items[1]", "novo-item")
		if err != nil {
			t.Fatalf("Erro ao definir valor: %v", err)
		}
		
		items, ok := data["items"].([]interface{})
		if !ok {
			t.Fatalf("Array 'items' não encontrado")
		}
		
		if items[1] != "novo-item" {
			t.Errorf("Valor não definido corretamente. Esperado: novo-item, Recebido: %v", items[1])
		}
	})
	
	// Teste 4: Definir valor em objeto dentro de array
	t.Run("SetObjectInArray", func(t *testing.T) {
		data := map[string]interface{}{
			"usuarios": []interface{}{
				map[string]interface{}{"id": 1, "nome": "João"},
				map[string]interface{}{"id": 2, "nome": "Maria"},
			},
		}
		
		err := extractor.SetValueByPath(data, "usuarios[1].nome", "Mariana")
		if err != nil {
			t.Fatalf("Erro ao definir valor: %v", err)
		}
		
		usuarios, ok := data["usuarios"].([]interface{})
		if !ok {
			t.Fatalf("Array 'usuarios' não encontrado")
		}
		
		usuario, ok := usuarios[1].(map[string]interface{})
		if !ok {
			t.Fatalf("Objeto usuário não encontrado no índice 1")
		}
		
		if usuario["nome"] != "Mariana" {
			t.Errorf("Valor não definido corretamente. Esperado: Mariana, Recebido: %v", usuario["nome"])
		}
	})
	
	// Teste 5: Criar novo objeto em um array existente
	t.Run("CreateObjectInExistingArray", func(t *testing.T) {
		data := map[string]interface{}{
			"transacoes": []interface{}{
				map[string]interface{}{"id": 1, "valor": 100.0},
			},
		}
		
		// Adicionar um novo objeto no índice 1
		err := extractor.SetValueByPath(data, "transacoes[1]", map[string]interface{}{
			"id": 2, 
			"valor": 200.0,
		})
		if err != nil {
			t.Fatalf("Erro ao definir valor: %v", err)
		}
		
		// Verificar se foi adicionado corretamente
		transacoes, ok := data["transacoes"].([]interface{})
		if !ok {
			t.Fatalf("Array 'transacoes' não encontrado")
		}
		
		if len(transacoes) != 2 {
			t.Fatalf("Tamanho incorreto do array. Esperado: 2, Recebido: %d", len(transacoes))
		}
		
		novaTransacao, ok := transacoes[1].(map[string]interface{})
		if !ok {
			t.Fatalf("Objeto transação não encontrado no índice 1")
		}
		
		if novaTransacao["valor"] != 200.0 {
			t.Errorf("Valor não definido corretamente. Esperado: 200.0, Recebido: %v", novaTransacao["valor"])
		}
	})
	
	// Teste 6: Testar com caminho completo do JSONPath (com $)
	t.Run("FullJSONPathWithDollarSign", func(t *testing.T) {
		data := map[string]interface{}{}
		
		err := extractor.SetValueByPath(data, "$.cliente.perfil.nivel", "premium")
		if err != nil {
			t.Fatalf("Erro ao definir valor: %v", err)
		}
		
		// Verificar o resultado
		jsonData, _ := json.Marshal(data)
		expected := `{"cliente":{"perfil":{"nivel":"premium"}}}`
		
		if string(jsonData) != expected {
			t.Errorf("Estrutura não criada corretamente.\nEsperado: %s\nRecebido: %s", expected, string(jsonData))
		}
	})
	
	// Teste 7: Definir um objeto complexo
	t.Run("SetComplexObject", func(t *testing.T) {
		data := map[string]interface{}{}
		
		endereco := map[string]interface{}{
			"rua":    "Av. Paulista",
			"numero": 1000,
			"cep":    "01310-100",
		}
		
		err := extractor.SetValueByPath(data, "empresa.filial.endereco", endereco)
		if err != nil {
			t.Fatalf("Erro ao definir valor: %v", err)
		}
		
		// Verificar se a estrutura complexa foi definida corretamente
		empresa, ok := data["empresa"].(map[string]interface{})
		if !ok {
			t.Fatalf("Estrutura 'empresa' não criada corretamente")
		}
		
		filial, ok := empresa["filial"].(map[string]interface{})
		if !ok {
			t.Fatalf("Estrutura 'filial' não criada corretamente")
		}
		
		enderecoSalvo, ok := filial["endereco"].(map[string]interface{})
		if !ok {
			t.Fatalf("Estrutura 'endereco' não criada corretamente")
		}
		
		if enderecoSalvo["rua"] != "Av. Paulista" || enderecoSalvo["numero"] != 1000 {
			t.Errorf("Objeto complexo não definido corretamente: %+v", enderecoSalvo)
		}
	})
	
	// Teste 8: Erro - caminho inválido
	t.Run("InvalidPath", func(t *testing.T) {
		data := map[string]interface{}{}
		
		err := extractor.SetValueByPath(data, "", "valor")
		if err == nil {
			t.Errorf("Deveria retornar erro para caminho vazio")
		}
	})
}