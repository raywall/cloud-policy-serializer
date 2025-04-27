package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/raywall/aws-policy-engine-go/pkg/json/extractor"
	"github.com/raywall/aws-policy-engine-go/pkg/json/schema"
)

var jsonSchema *schema.Schema

func init() {
	loader, err := schema.NewLoader("/Users/macmini/Documents/workspace/go/aws-policy-engine-go/examples/request_schema.json")
	if err != nil {
		panic(err)
	}

	// Exemplo de schema draft-07 (como no enunciado)
	jsonSchema, err = loader.Load()
	if err != nil {
		panic(err)
	}

	// Exemplo de dados JSON a validar
	data, err := os.ReadFile("/Users/macmini/Documents/workspace/go/aws-policy-engine-go/examples/request_data.json")
	if err != nil {
		panic(err)
	}

	var jsonData interface{}
	if err = json.Unmarshal(data, &jsonData); err != nil {
		panic(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/evaluate", handleEvaluate).Methods("POST")

	fmt.Println("Servidor iniciado em :9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var request, response interface{}

	// Decodificar o JSON da requisição
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		handleError(w, "invalid_request", "Erro ao decodificar requisição", http.StatusBadRequest)
		return
	}

	isValid, errs := jsonSchema.Validate(request)
	if !isValid {
		data := []string{"Erros de validação:"}
		for _, e := range errs {
			data = append(data, fmt.Sprintf("- %s", e))
		}

		handleError(w, "invalid_request", data, http.StatusBadRequest)
		return
	}

	ex, err := extractor.NewFromStruct(request)
	if err != nil {
		handleError(w, "invalid_request", err, http.StatusBadRequest)
	}

	response, err = ex.Extract("$.data")
	if err != nil {
		handleError(w, "invalid_request", err, http.StatusBadRequest)
	}

	// Enviar resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func handleError(w http.ResponseWriter, code, message interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(message)
}
