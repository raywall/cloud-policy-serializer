package main

import (
	"log"

	"github.com/raywall/aws-policy-engine-go/pkg/service"
	"github.com/raywall/aws-policy-engine-go/pkg/service/schema"
)

func main() {
	ctx := service.ServiceEngineContext{
		Receiver:       service.ServiceReceiver{},
		ServiceSchemas: schema.NewServiceSchemas(),
	}

	ctx.ServiceSchemas.NewSchema("POST", "/evaluate",
		&schema.SchemaPath{
			Request:  "examples/request_schema.json",
			Response: "examples/",
			Policy:   "examples/",
		})

	if err := ctx.ListenAndServe(9000); err != nil {
		log.Fatal(err)
	}

}
