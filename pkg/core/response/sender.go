package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (res *Response) Send(w http.ResponseWriter) {
	res.send(w, 200)
}

func (res *Response) SendError(w http.ResponseWriter, code string, err error) {
	if err != nil {
		res.Error = Error{
			Code:    code,
			Message: err.Error(),
		}

		res.send(w, http.StatusInternalServerError)
	}

	res.Passed = false
	res.send(w, http.StatusBadRequest)
}

func (res *Response) send(w http.ResponseWriter, statusCode int) {
	res.ElapsedTime = time.Since(res.startedOn).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var result interface{}
	switch res.debugger {
	case true:
		result = res
	default:
		result = res.formattedResponse
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, fmt.Sprintf("Erro ao codifocar resposta JSON: %v", err), http.StatusInternalServerError)
	}
}
