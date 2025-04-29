package response

import (
	"time"

	"github.com/raywall/aws-policy-engine-go/pkg/core/policy"
	"github.com/raywall/aws-policy-engine-go/pkg/core/request"
	"github.com/raywall/aws-policy-engine-go/pkg/json/schema"
	"github.com/raywall/aws-policy-engine-go/pkg/json/schema/builder"
)

func NewResponse(responseSchema *schema.Schema, req *request.Request, debug bool) *Response {
	return &Response{
		ID:              req.ID,
		Timestamp:       req.Timestamp,
		Passed:          true,
		Data:            req.Data,
		AppliedPolicies: []policy.PolicyResult{},
		responseSchema:  responseSchema,
		startedOn:       time.Now(),
		debugger:        debug,
	}
}

func (res *Response) BuildResponse(pe *policy.PolicyEngine, data map[string]interface{}, e *Error) {
	if e != nil {
		res.Error = *e
		return
	}

	res.Data = data

	for _, result := range pe.Results {
		if !result.Passed {
			res.Passed = false
		}

		res.AppliedPolicies = append(res.AppliedPolicies, result)
	}

	sf, err := builder.NewSchemaFormatter(res.responseSchema)
	if err != nil {
		res.Error = Error{
			Code:    "unexpected_formatter_error",
			Message: "Erro ao formatar resposta",
			Error:   err,
		}
		return
	}

	res.formattedResponse, err = sf.FormatResponse(data)
	if err != nil {
		res.Error = Error{
			Code:    "unexpected_response_error",
			Message: "Erro ao formatar resposta",
			Error:   err,
		}
		return
	}
}
