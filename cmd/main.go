package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/raywall/aws-policy-engine-go/pkg/core/policy"
	"github.com/raywall/aws-policy-engine-go/pkg/core/request"
	"github.com/raywall/aws-policy-engine-go/pkg/core/response"
	"github.com/raywall/aws-policy-engine-go/pkg/json/schema"
)

var (
	jsonSchema   *schema.Schema
	policyEngine *policy.PolicyEngine
)

func init() {
	sl, err := schema.NewLoader("examples/request_schema.json")
	if err != nil {
		panic(err)
	}

	// Exemplo de schema draft-07 (como no enunciado)
	jsonSchema, err = sl.Load()
	if err != nil {
		panic(err)
	}

	pl, err := policy.NewLoader("examples/policy.yaml")
	if err != nil {
		panic(err)
	}

	// Exemplo de arquivo de políticas em formato YAML
	policyEngine, err = pl.Load()
	if err != nil {
		panic(err)
	}

	// Exemplo de dados JSON a validar
	data, err := os.ReadFile("examples/request_data.json")
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
	var (
		req = &request.Request{}
		res *response.Response
	)

	// Decodificar o JSON da requisição
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(req)
	if err != nil {
		fmt.Println("invalid_json_decode", "Erro ao decodificar requisição")
		return
	}

	// Instancia o response da requisição
	res = response.NewResponse(req)

	// Validar o JSON Schema dos dados recebidos
	isValid, errs := jsonSchema.Validate(req.Data)
	if !isValid {
		data := []string{"Erros de validação:"}
		for _, e := range errs {
			data = append(data, fmt.Sprintf("- %s", e))
		}

		res.SendError(w, "invalid_json_schema", err)
		return
	}

	// Executar políticas nos dados recebidos
	if err = policyEngine.ExecutePolicies(req); err != nil {
		res.SendError(w, "invalid_exec_policies", err)
		return
	}

	res.BuildResponse(policyEngine, req.Data, nil)

	// Enviar resposta
	res.Send(w)
}
