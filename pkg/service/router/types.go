package router

import (
	"fmt"
	"net/http"
)

type ServiceRoutes struct {
	Routes map[string]*ServiceRoute
}

type ServiceRoute struct {
	Method  string
	Path    string
	Handler func(http.ResponseWriter, *http.Request)
}

func NewServiceRoutes() *ServiceRoutes {
	return &ServiceRoutes{
		Routes: make(map[string]*ServiceRoute, 0),
	}
}

func (r *ServiceRoutes) HandleFunc(method, path string, handler func(http.ResponseWriter, *http.Request)) {
	key := fmt.Sprintf("%s %s", method, path)
	r.Routes[key] = &ServiceRoute{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
	http.HandleFunc(path, r.Routes[key].Handler)
}

func (r *ServiceRoutes) Get(method, path string) *ServiceRoute {
	return r.Routes[fmt.Sprintf("%s %s", method, path)]
}
