package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// --- Main (Exemplo de Uso) ---
func main() {
	// Definir Schemas (simplificado)
	reqSchemaPath := FilePath("request_schema.json")
	reqSchema, err := reqSchemaPath.GetSchema()
	if err != nil {
		panic(err)
	}

	respSchemaPath := FilePath("response_schema.json")
	respSchema, err := respSchemaPath.GetSchema()
	if err != nil {
		panic(err)
	}

	// Definir Políticas
	policiesPath := FilePath("policy.yaml")
	policies, err := policiesPath.GetPolicies()
	if err != nil {
		panic(err)
	}

	// Criar Contexto do Motor
	engine := NewEngineContext(reqSchema, respSchema, *policies, "Local")

	// Exemplo de Requisição
	requestBody, err := ioutil.ReadFile("request_data.json")
	if err != nil {
		panic(err)
	}

	fmt.Println("--- Processando Requisição (Válida) ---")
	response, err := engine.ProcessRequest(requestBody)
	if err != nil {
		fmt.Printf("Erro:\n%v\n", err)
	} else {
		respBytes, _ := json.MarshalIndent(response, "", "  ")
		fmt.Printf("Resposta:\n%s\n", string(respBytes))
	}
	fmt.Println("\n--- Fim Requisição ---")
}
