package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/raywall/aws-policy-engine-go/pkg/engine"
)

var ec engine.PolicyEngineContext

func init() {
	var err error

	ec, err = engine.NewEngine(
		"examples/request_schema.json",
		"examples/response_schema.json",
		"examples/policy.yaml",
	)
	if err != nil {
		panic(err)
	}

	if err := ec.Load(); err != nil {
		panic(err)
	}
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	// Verifica se o método é POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodifica o JSON recebido no body da requisição
	if err := ec.CreateRequest(r); err != nil {
		responseHandler(w, nil, err)
		return
	}

	// Verifica se a solicitação está em modo debug
	responseDebug := false
	queryValues, _ := url.ParseQuery(r.URL.RawQuery)
	if queryValues.Get("debug") == "true" {
		responseDebug = true
	}

	// Instancia o response da requisição
	_ = ec.CreateResponse(responseDebug)

	// Executar políticas nos dados recebidos
	if err := ec.ExecutePolicies(); err != nil {
		responseHandler(w, nil, err)
		return
	}

	// Enviar resposta
	ec.Answer(w)
}

func main() {
	// Configurar rotas
	http.HandleFunc("/evaluate", handleEvaluate)

	// Iniciar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	fmt.Printf("Iniciando servidor na porta %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(map[string]interface{}{
			"success": false,
			"error": map[string]interface{}{
				"code":    "unexpected_server_error",
				"message": "Erro ao iniciar servidor",
				"data":    fmt.Sprintf("%w", err),
			},
		})
	}
}

// responseHandler formata e envia a resposta de uma requisição
func responseHandler(w http.ResponseWriter, formattedResponse interface{}, err *engine.Error) {
	// Se houver erro na formatação
	if err != nil {
		errorResponse := map[string]interface{}{
			"success": false,
			"error": map[string]interface{}{
				"code":    err.Code,
				"message": err.Message,
				"data":    err.Data,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Preparar resposta de sucesso
	successResponse := map[string]interface{}{
		"success": true,
		"data":    formattedResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(successResponse)
}
