package schema

import (
	"fmt"
	"net/http"

	"github.com/raywall/aws-policy-engine-go/pkg/service/router"
)

type ServiceSchemas struct {
	Schemas map[string]*ServiceSchema
}

type ServiceSchema struct {
	Paths        *SchemaPath
	ServiceRoute *router.ServiceRoute
}

type SchemaPath struct {
	Request  string
	Response string
	Policy   string
}

func NewServiceSchemas() *ServiceSchemas {
	return &ServiceSchemas{
		Schemas: make(map[string]*ServiceSchema, 0),
	}
}

func (s *ServiceSchemas) NewSchema(method, path string, schemaPath *SchemaPath) {
	key := fmt.Sprintf("%s %s", method, path)
	s.Schemas[key] = &ServiceSchema{
		Paths: schemaPath,
		ServiceRoute: &router.ServiceRoute{
			Method: method,
			Path:   path,
			Handler: func(res http.ResponseWriter, req *http.Request) {
				fmt.Println("Hello World!")
			},
		},
	}

	http.HandleFunc(path, s.Schemas[key].ServiceRoute.Handler)
}
