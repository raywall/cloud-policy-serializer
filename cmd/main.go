package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/raywall/aws-policy-engine-go/pkg/schema"
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

	fmt.Println("Servidor iniciado em :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var request interface{}

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

	// Enviar resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode("JSON válido segundo o schema!")
}

func handleError(w http.ResponseWriter, code, message interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(message)
}
