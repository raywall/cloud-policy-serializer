package service

import (
	"fmt"
	"net/http"

	"github.com/raywall/aws-policy-engine-go/pkg/service/schema"
)

// ServiceReceiver
type ServiceReceiver struct {
}

// ServiceEngineContext
type ServiceEngineContext struct {
	Receiver       ServiceReceiver
	ServiceSchemas *schema.ServiceSchemas
}

func (r *ServiceEngineContext) ListenAndServe(port int) error {
	fmt.Printf("Servidor iniciado em :%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
