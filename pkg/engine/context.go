package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/raywall/aws-policy-engine-go/pkg/core/policy"
	"github.com/raywall/aws-policy-engine-go/pkg/core/request"
	"github.com/raywall/aws-policy-engine-go/pkg/core/response"
	"github.com/raywall/aws-policy-engine-go/pkg/json/schema"
)

type PolicyEngineContext interface {
	Load() *Error
	CreateRequest(r *http.Request) *Error
	CreateResponse(debug bool) *Error
	ExecutePolicies() *Error
	Answer(w http.ResponseWriter)
}

type paths struct {
	requestSchema  string `json:"requestSchema"`
	responseSchema string `json:"responseSchema"`
	policyRules    string `json:"policyRules"`
}

type Config struct {
	RequestSchema     *schema.Schema       `json:"requestSchema"`
	ResponseSchema    *schema.Schema       `json:"responseSchema"`
	PolicyEngineRules *policy.PolicyEngine `json:"policyEngineRules"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    error  `json:"data"`
}

type policyEngineContext struct {
	Config   Config             `json:"config"`
	Request  *request.Request   `json:"request"`
	Response *response.Response `json:"response"`
	paths    paths              `json:"pathConfig"`
}

func NewEngine(requestSchemaPath, responseSchemaPath, policyPath string) (PolicyEngineContext, error) {
	if requestSchemaPath == "" {
		return nil, errors.New("request schema path cannot be empty")
	}
	if responseSchemaPath == "" {
		return nil, errors.New("response schema path cannot be empty")
	}
	if policyPath == "" {
		return nil, errors.New("policy rules path cannot be empty")
	}

	ec := &policyEngineContext{
		Request: &request.Request{},
		paths: paths{
			requestSchema:  requestSchemaPath,
			responseSchema: responseSchemaPath,
			policyRules:    policyPath,
		},
	}
	return ec, nil
}

func (ec *policyEngineContext) Load() *Error {
	// request schema loader
	requestLoader, err := schema.NewLoader(ec.paths.requestSchema)
	if err != nil {
		return &Error{
			Code:    "",
			Message: "failed to build a new loader",
			Data:    err,
		}
	}
	ec.Config.RequestSchema, err = requestLoader.Load()
	if err != nil {
		return &Error{
			Code:    "",
			Message: "failed to load the request schema",
			Data:    err,
		}
	}

	// response schema loader
	responseLoader, err := schema.NewLoader(ec.paths.responseSchema)
	if err != nil {
		return &Error{
			Code:    "",
			Message: "failed to build a new loader",
			Data:    err,
		}
	}
	ec.Config.ResponseSchema, err = responseLoader.Load()
	if err != nil {
		return &Error{
			Code:    "",
			Message: "failed to load the response schema: %v",
			Data:    err,
		}
	}

	// policy rules loader
	policyLoader, err := policy.NewLoader(ec.paths.policyRules)
	if err != nil {
		return &Error{
			Code:    "",
			Message: "failed to build a new loader",
			Data:    err,
		}
	}
	ec.Config.PolicyEngineRules, err = policyLoader.Load()
	if err != nil {
		return &Error{
			Code:    "",
			Message: "failed to load the policy rules",
			Data:    err,
		}
	}

	return nil
}

func (ec *policyEngineContext) CreateRequest(r *http.Request) *Error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(ec.Request)
	if err != nil {
		return &Error{
			Code:    "invalid_json_decode",
			Message: "Erro ao decodificar requisição",
			Data:    err,
		}
	}
	defer r.Body.Close()

	// Validar o JSON Schema dos dados recebidos
	isValid, errs := ec.Config.RequestSchema.Validate(ec.Request.Data)
	if !isValid {
		data := []string{}
		for _, e := range errs {
			data = append(data, fmt.Sprintf("- %s", e))
		}

		return &Error{
			Code:    "invalid_json_schema",
			Message: "Erros de validação",
			Data:    errors.New(strings.Join(data, "\n")),
		}
	}

	return nil
}

func (ec *policyEngineContext) CreateResponse(debug bool) *Error {
	ec.Response = response.NewResponse(
		ec.Config.ResponseSchema,
		ec.Request,
		debug,
	)

	return nil
}

func (ec *policyEngineContext) ExecutePolicies() *Error {
	if err := ec.Config.PolicyEngineRules.ExecutePolicies(ec.Request); err != nil {
		return &Error{
			Code:    "invalid_exec_policies",
			Message: "failed to execute policies",
			Data:    err,
		}
	}

	return nil
}

func (ec *policyEngineContext) Answer(w http.ResponseWriter) {
	ec.Response.BuildResponse(ec.Config.PolicyEngineRules, ec.Request.Data, nil)
	ec.Response.Send(w)
}
