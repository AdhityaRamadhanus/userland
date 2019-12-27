package api

import (
	"github.com/gorilla/mux"
)

// TODO create a error handling for service errors

//Handler is a contract to every handler that want to be included in the apps
type Handler interface {
	RegisterRoutes(router *mux.Router)
}
