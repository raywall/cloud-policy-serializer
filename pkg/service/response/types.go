package response

import (
	"encoding/json"
	"net/http"
)

// ServiceResponse
type ServiceResponse interface {
	Header() http.Header
	Write(data []byte) (int, error)
	WriteHeader(statusCode int)
	GetResponse() (map[string]interface{}, error)
}

type ServiceResponseWriter struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func NewResponseWriter() ServiceResponse {
	return &ServiceResponseWriter{
		Headers: make(http.Header),
	}
}

func (w *ServiceResponseWriter) Header() http.Header {
	return w.Headers
}

func (w *ServiceResponseWriter) Write(data []byte) (int, error) {
	w.Body = append(w.Body, data...)
	return len(data), nil
}

func (w *ServiceResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func (w *ServiceResponseWriter) GetResponse() (map[string]interface{}, error) {
	jsonData, err := json.Marshal(*w)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]interface{}, 0)
	if err = json.Unmarshal(jsonData, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
