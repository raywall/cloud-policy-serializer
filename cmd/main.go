package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/raywall/aws-policy-engine-go/pkg/core/policy"
	"github.com/raywall/aws-policy-engine-go/pkg/core/request"
	"github.com/raywall/aws-policy-engine-go/pkg/core/response"
	"github.com/raywall/aws-policy-engine-go/pkg/json/schema"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    error  `json:"data"`
}

var (
	requestSchema  *schema.Schema
	responseSchema *schema.Schema
	policyEngine   *policy.PolicyEngine
)

func init() {
	// Carrega o JSON schema draft-07 do request
	sl, err := schema.NewLoader("examples/request_schema.json")
	if err != nil {
		panic(err)
	}

	requestSchema, err = sl.Load()
	if err != nil {
		panic(err)
	}

	// Carrega o JSON Schema draft-07 do response
	rl, err := schema.NewLoader("examples/response_schema.json")
	if err != nil {
		panic(err)
	}

	responseSchema, err = rl.Load()
	if err != nil {
		panic(err)
	}

	// Carrega o arquivo de políticas em formato YAML
	pl, err := policy.NewLoader("examples/policy.yaml")
	if err != nil {
		panic(err)
	}

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

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var (
		req = &request.Request{}
		res *response.Response
	)

	// Verifica se o método é POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Lê o corpo da requisição
	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("Erro ao ler corpo da requisição: %v", err), http.StatusBadRequest)
	// 	return
	// }

	// Decodifica o JSON recebido no body da requisição
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(req)
	if err != nil {
		responseHandler(w, nil, &Error{
			Code:    "invalid_json_decode",
			Message: "Erro ao decodificar requisição",
			Data:    err,
		})
		return
	}
	defer r.Body.Close()

	// Instancia o response da requisição
	res = response.NewResponse(responseSchema, req, false)

	// Validar o JSON Schema dos dados recebidos
	isValid, errs := requestSchema.Validate(req.Data)
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

	// Formata e envia a resposta de acordo com o Schema
	res.BuildResponse(policyEngine, req.Data, nil)

	// Enviar resposta
	res.Send(w)
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
func responseHandler(w http.ResponseWriter, formattedResponse interface{}, err *Error) {
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
